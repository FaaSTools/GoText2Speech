package GoText2Speech

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/FaaSTools/GoStorage/gostorage"
	ts2_aws "github.com/FaaSTools/GoText2Speech/GoText2Speech/aws"
	ts2_gcp "github.com/FaaSTools/GoText2Speech/GoText2Speech/gcp"
	"github.com/FaaSTools/GoText2Speech/GoText2Speech/providers"
	. "github.com/FaaSTools/GoText2Speech/GoText2Speech/shared"
	"github.com/aws/aws-sdk-go-v2/aws"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

type GoT2SClient struct {
	providerInstances map[providers.Provider]*T2SProvider
	region            string
	credentials       *CredentialsHolder
	tempBuckets       map[providers.Provider]string
	DeleteTempFile    bool
	gostorageClient   *gostorage.GoStorage
}

func CreateGoT2SClient(credentials *CredentialsHolder, region string) GoT2SClient {
	if credentials == nil {
		awsCred, gcpCred := gostorage.LoadCredentialsFromDefaultLocation()
		awsCred = &aws.Credentials{
			AccessKeyID:     awsCred.AccessKeyID,
			SecretAccessKey: awsCred.SecretAccessKey,
		}
		credentials = &CredentialsHolder{
			AwsCredentials:    awsCred,
			GoogleCredentials: gcpCred,
		}
	}
	return GoT2SClient{
		providerInstances: make(map[providers.Provider]*T2SProvider),
		tempBuckets:       make(map[providers.Provider]string),
		credentials:       credentials,
		region:            region,
		DeleteTempFile:    true,
	}
}

func (a GoT2SClient) getProviderInstance(provider providers.Provider) T2SProvider {
	if a.providerInstances[provider] == nil {
		prov := CreateProviderInstance(provider)
		var err error = nil
		prov, err = prov.CreateServiceClient(*a.credentials, a.region)
		if err != nil {
			fmt.Printf("Error while creating service client: %s\n", err)
		}
		a.providerInstances[provider] = &prov
	}
	return *a.providerInstances[provider]
}

func (a GoT2SClient) CloseProviderClient(provider providers.Provider) error {
	return a.getProviderInstance(provider).CloseServiceClient()
}

func (a GoT2SClient) CloseAllProviderClients() error {
	var allErrors error = nil
	for _, instance := range a.providerInstances {
		err := (*instance).CloseServiceClient()
		if err != nil {
			if allErrors == nil {
				allErrors = err
			} else {
				allErrors = errors.Join(allErrors, err)
			}
		}
	}
	return allErrors
}

func (a GoT2SClient) SetTempBucket(provider providers.Provider, tempBucket string) {
	a.tempBuckets[provider] = tempBucket
}

// T2SDirect Transforms the given text into speech and stores the file in destination.
// If the given options specify a provider, this provider will be used.
// If the given options don't specify a provider, a provider will be chosen based on heuristics.
func (a GoT2SClient) T2SDirect(text string, destination string, options TextToSpeechOptions) (GoT2SClient, error) {

	// error check: If the given text is supposed to be a SSML text and does not contain <speak>-tags, it is invalid.
	if (options.TextType == TextTypeSsml) && !HasSpeakTag(text) {
		return a, errors.New("invalid text. The text type was SSML, but the given text didn't contain <speak>-tags")
	}

	// if text type is auto, text type needs to be inferred
	if options.TextType == TextTypeAuto {
		// SSML text needs to be wrapped in a "speak" root node (i.e. <speak>...</speak>)
		if HasSpeakTag(text) {
			options.TextType = TextTypeSsml
		} else {
			options.TextType = TextTypeText
		}
	}

	if options.Provider == providers.ProviderUnspecified {
		if !options.VoiceConfig.VoiceIdConfig.IsEmpty() {
			fmt.Printf("Cloud provider was unspecified, but voiceId was specified. In most cases, the voiceId is " +
				"only available on a single provider. This means that the provider that will be chosen automatically " +
				"might not support the specified voiceId. For best results, either specify the cloud provider " +
				"alongslide the voiceId, or remove voiceId and specify voice parameters (gender & language).\n")
		}

		var err error
		options, err = a.determineProvider(options, destination)
		if err != nil {
			return a, err
		}
	}

	provider := a.getProviderInstance(options.Provider)

	if options.VoiceConfig.VoiceIdConfig.IsEmpty() {
		// if both VoiceParamsConfig is undefined -> use default object
		if options.VoiceConfig.VoiceParamsConfig == (VoiceParamsConfig{}) {
			options.VoiceConfig.VoiceParamsConfig = GetDefaultVoiceParamsConfig()
		} else {
			// if either of the properties are unset, set default values
			if options.VoiceConfig.VoiceParamsConfig.Gender == VoiceGenderUnspecified {
				options.VoiceConfig.VoiceParamsConfig.Gender = GetDefaultVoiceParamsConfig().Gender
			}
			if options.VoiceConfig.VoiceParamsConfig.LanguageCode == "" {
				options.VoiceConfig.VoiceParamsConfig.LanguageCode = GetDefaultVoiceParamsConfig().LanguageCode
			}
		}

		fmt.Printf("Trying to find voice\n")
		voiceIdConfig, chooseVoiceErr := provider.FindVoice(options)
		if chooseVoiceErr != nil {
			return a, chooseVoiceErr
		}
		options.VoiceConfig.VoiceIdConfig = *voiceIdConfig
	}

	// adjust parameters for the chosen provider
	var transformOptionsError error
	text, options, transformOptionsError = provider.TransformOptions(text, options)

	if transformOptionsError != nil {
		return a, transformOptionsError
	}

	fmt.Println("Final Text: " + text)

	// adjust provider-specific settings and execute T2S on selected provider
	audioData, t2sErr := provider.ExecuteT2SDirect(text, destination, options)
	if t2sErr != nil {
		return a, t2sErr
	}

	var fileExtErr error = nil
	destination, fileExtErr = provider.AddFileExtensionToDestinationIfNeeded(options, options.OutputFormatRaw, destination)
	if fileExtErr != nil { // not a fatal error
		fmt.Printf(fileExtErr.Error())
	}

	if provider.IsURLonOwnStorage(destination) { // own storage -> upload directly
		err := provider.UploadFile(audioData, destination)
		if err != nil {
			return a, errors.Join(errors.New(fmt.Sprintf("error while uploading audio file to %s", destination)), err)
		}
	} else if a.IsProviderStorageUrl(destination) { // other cloud storage -> upload via GoStorage
		tmpFile, err := os.CreateTemp("", "sample")
		if err != nil {
			return a, errors.Join(errors.New("error while creating file for temporarily storing audio file before upload"), err)
		}

		err = StoreAudioToLocalFile(audioData, tmpFile)
		if err != nil {
			if a.IsProviderStorageUrl(destination) {
				return a, errors.Join(errors.New("error while writing audio to temporary file"), err)
			}
			return a, errors.Join(errors.New("error while writing audio to local file"), err)
		}

		target := ParseUrlToGoStorageObject(destination)
		a.gostorageClient.UploadFile(gostorage.GoStorageObject{
			Bucket:        target.Bucket,
			Key:           target.Key,
			Region:        target.Region,
			IsLocal:       true,
			LocalFilePath: tmpFile.Name(),
			ProviderType:  target.ProviderType,
		})

		closeErr := tmpFile.Close()
		if closeErr != nil {
			return a, errors.Join(errors.New("error while closing tmp file"), closeErr)
		}

		if a.DeleteTempFile {
			removeErr := os.Remove(tmpFile.Name())
			if removeErr != nil {
				return a, errors.Join(errors.New("error while removing temporarily stored audio file"), removeErr)
			}
		}
	} else { // local file -> store locally
		file, err := os.Open(destination)
		if err != nil {
			return a, errors.Join(errors.New(fmt.Sprintf("error while opening file at destination %s", destination)), err)
		}

		err = StoreAudioToLocalFile(audioData, file)
		if err != nil {
			return a, errors.Join(errors.New("error while writing audio to local file"), err)
		}
		closeErr := file.Close()
		if closeErr != nil {
			return a, errors.Join(errors.New("error while closing local file"), closeErr)
		}
	}

	// move file to actual destination, if needed
	/*
		if !strings.EqualFold(providerDestination, destination) {

			tempStorageObj := ParseUrlToGoStorageObject(providerDestination)
			if a.IsProviderStorageUrl(destination) {
				actualStorageObj := ParseUrlToGoStorageObject(destination)
				a.gostorageClient.Copy(tempStorageObj, actualStorageObj)
			} else { // local file
				a.gostorageClient.DownloadFile(tempStorageObj, destination)
			}

			if a.DeleteTempFile {
				a.gostorageClient.DeleteFile(tempStorageObj)
			}
		}
	*/

	return a, t2sErr
}

// T2S Transforms the text in the source file into speech and stores the file in destination.
// The given source parameter specifies the location of the file. The file can have one of the following locations:
// * AWS S3
// * Google Cloud Storage
// * Other publicly accessible URL (beginning with 'http' or 'https')
// * Local file
// If the given options specify a provider, this provider will be used.
// If the given options don't specify a provider, a provider will be chosen based on heuristics.
func (a GoT2SClient) T2S(source string, destination string, options TextToSpeechOptions) (GoT2SClient, error) {

	localFilePath := ""
	text := ""
	fileOnCloudProvider := false
	if a.IsProviderStorageUrl(source) { // file on supported cloud provider
		storageObj := ParseUrlToGoStorageObject(source)
		a = a.initializeGoStorage()
		fileReader := a.gostorageClient.DownloadFileAsReader(storageObj)
		fileBuf := new(bytes.Buffer)
		_, bufErr := fileBuf.ReadFrom(fileReader)
		if bufErr != nil {
			return a, errors.Join(errors.New("error occurred while reading input file from file reader"), bufErr)
		}
		text = fileBuf.String()
		readerCloseErr := (fileReader.(io.ReadCloser)).Close()
		if readerCloseErr != nil {
			fmt.Printf("non fatal-error while closing input file reader: %s\n", readerCloseErr.Error())
		}
		fileOnCloudProvider = true
	} else if strings.HasPrefix(source, "http") { // file somewhere else online
		response, err := http.Get(source)
		if err != nil {
			return a, errors.Join(errors.New(fmt.Sprintf("Couldn't download the source file '%s'.", source)), err)
		}

		// close body after function call ended
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Printf(errors.Join(errors.New(fmt.Sprintf("A non-fatal error occurred while closing the HTTP response for the source file '%s'.", source)), err).Error())
			}
		}(response.Body)

		textBytes, err2 := io.ReadAll(response.Body)
		if err2 != nil {
			return a, errors.Join(errors.New(fmt.Sprintf("Couldn't download the source file '%s'. An error occurred while reading body.", source)), err)
		}
		text = string(textBytes)
	} else { // local file
		localFilePath = source
	}

	if !strings.EqualFold("", localFilePath) {
		dat, err := os.ReadFile(localFilePath)
		if err != nil {
			helperText := ""
			if fileOnCloudProvider {
				helperText = "temporarily stored "
			}
			return a, errors.Join(errors.New(fmt.Sprintf("Couldn't read the %stext file on '%s'.", helperText, localFilePath)), err)
		}
		text = string(dat)
	}

	fmt.Printf("Read the following text from file: %s\n", text)
	return a.T2SDirect(text, destination, options)
}

func (a GoT2SClient) initializeGoStorage() GoT2SClient {
	if a.gostorageClient == nil {
		a.gostorageClient = &gostorage.GoStorage{
			Credentials: *a.credentials,
		}
	}
	return a
}

func CreateProviderInstance(provider providers.Provider) T2SProvider {
	switch provider {
	case providers.ProviderAWS:
		return ts2_aws.T2SAmazonWebServices{}
	case providers.ProviderGCP:
		return ts2_gcp.T2SGoogleCloudPlatform{}
	default:
		return nil
	}
}

func (a GoT2SClient) determineProvider(options TextToSpeechOptions, destination string) (TextToSpeechOptions, error) {

	// First heuristic: Choose provider that offers voice parameters (gender, language)
	var wg sync.WaitGroup
	var mut sync.Mutex

	voicePerProvider := make(map[providers.Provider]*VoiceIdConfig)
	for _, provider := range providers.GetAllProviders() {

		if !options.VoiceConfig.VoiceIdConfig.IsEmpty() {
			voicePerProvider[provider] = &options.VoiceConfig.VoiceIdConfig
			continue
		}

		wg.Add(1)
		go func(prov providers.Provider) {
			defer wg.Done()
			defer mut.Unlock()
			voiceId, err := a.getProviderInstance(prov).FindVoice(options)
			mut.Lock()
			if err != nil {
				fmt.Printf("Error while trying to find voice for provider %s: %s", prov, err.Error())
				voicePerProvider[prov] = nil
			} else {
				voicePerProvider[prov] = voiceId
			}
		}(provider)
	}

	wg.Wait()

	for prov, voice := range voicePerProvider {
		if voice == nil {
			delete(voicePerProvider, prov)
		}
	}

	if len(voicePerProvider) < 2 {
		if len(voicePerProvider) < 1 {
			return options, errors.New(fmt.Sprintf(
				"Error while trying to find voice. No voice found with the given language '%s' and gender '%s' on any provider.",
				options.VoiceConfig.VoiceParamsConfig.LanguageCode, options.VoiceConfig.VoiceParamsConfig.Gender))
		}

		// Only one provider offers this voice -> use this provider
		for prov, voice := range voicePerProvider {
			options.Provider = prov
			options.VoiceConfig.VoiceIdConfig = *voice
			return options, nil
		}
	}

	// Multiple providers offer this voice -> next heuristic
	// Second heuristic: Choose provider on which the destination file should be stored

	for prov, voice := range voicePerProvider {
		if a.getProviderInstance(prov).IsURLonOwnStorage(destination) {
			options.Provider = prov
			options.VoiceConfig.VoiceIdConfig = *voice
			return options, nil
		}
	}

	// Third/Fourth heuristic: Choose provider that offers the chosen output format
	if options.OutputFormat != AudioFormatUnspecified {
		for prov, _ := range voicePerProvider {
			audioFormats := a.getProviderInstance(prov).GetSupportedAudioFormats()
			// make sure at least one provider is still available in the end
			if !IncludesAudioFormat(audioFormats, options.OutputFormat) && (len(voicePerProvider) > 1) {
				delete(voicePerProvider, prov)
			}
		}
	}

	// The following code is executed in one of these cases:
	// * Output format was unspecified
	// * Multiple providers support the output format
	// * Only one provider is left, which supports the output format
	// * Only one provider is left, which doesn't support the output format
	// Use first provider in map (i.e. random provider that is still left)
	for prov, voice := range voicePerProvider {
		options.Provider = prov
		options.VoiceConfig.VoiceIdConfig = *voice
		return options, nil
	}

	return options, errors.New("error while choosing provider for text-to-speech: Undefined error. This error should not have happened")
}

// IsProviderStorageUrl checks if the given string is a valid file URL for a storage service of one of the
// supported storage providers.
func (a GoT2SClient) IsProviderStorageUrl(url string) bool {
	return IsAWSUrl(url) || IsGoogleUrl(url)
}
