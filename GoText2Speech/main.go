package GoText2Speech

import (
	"errors"
	"fmt"
	"github.com/FaaSTools/GoStorage/gostorage"
	//"github.com/dave-meyer/GoStorage/gostorage"
	ts2_aws "goTest/GoText2Speech/aws"
	ts2_gcp "goTest/GoText2Speech/gcp"
	"goTest/GoText2Speech/providers"
	. "goTest/GoText2Speech/shared"
	"io"
	"net/http"
	"os"
	"strings"
)

type GoT2SClient struct {
	providerInstances map[providers.Provider]*T2SProvider
	region            string
	credentials       CredentialsHolder
	tempBuckets       map[providers.Provider]string
	DeleteTempFile    bool
	gostorageClient   gostorage.GoStorage
}

func CreateGoT2SClient(credentials CredentialsHolder, region string) GoT2SClient {
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
		prov = prov.CreateServiceClient(a.credentials, a.region)
		a.providerInstances[provider] = &prov
	}
	return *a.providerInstances[provider]
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
		// TODO choose provider based on heuristics
		options.Provider = providers.ProviderAWS
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

		voiceIdConfig, chooseVoiceErr := provider.FindVoice(options)
		if chooseVoiceErr != nil {
			return a, chooseVoiceErr
		}
		options.VoiceConfig.VoiceIdConfig = *voiceIdConfig
	}

	// adjust parameters for Google/AWS
	var transformOptionsError error
	text, options, transformOptionsError = provider.TransformOptions(text, options)

	if transformOptionsError != nil {
		return a, transformOptionsError
	}

	fmt.Println("Final Text: " + text)

	// if destination of file is not on the storage service of the selected provider:
	// create temporary location, execute T2S, and move file to actual destination.
	providerDestination := destination
	if !provider.IsURLonOwnStorage(destination) {
		splits := strings.Split(destination, "/")
		fileName := splits[len(splits)-1]
		providerDestination = provider.CreateTempDestination(a.tempBuckets[options.Provider], fileName)
	}

	// adjust provider-specific settings and execute T2S on selected provider
	t2sErr := provider.ExecuteT2SDirect(text, providerDestination, options)

	// move file to actual destination, if needed
	if !strings.EqualFold(providerDestination, destination) {

		tempStorageObj := ParseUrlToGoStorageObject(providerDestination)
		if IsProviderStorageUrl(destination) {
			actualStorageObj := ParseUrlToGoStorageObject(destination)
			a.gostorageClient.Copy(tempStorageObj, actualStorageObj)
		} else { // local file
			a.gostorageClient.DownloadFile(tempStorageObj, destination)
		}

		if a.DeleteTempFile {
			a.gostorageClient.DeleteFile(tempStorageObj)
		}
	}

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
	if IsProviderStorageUrl(source) { // file on cloud provider
		f, err := os.CreateTemp("", "sample")
		if err != nil {
			return a, errors.Join(errors.New(fmt.Sprintf("Couldn't download the source file '%s' because creation of temporary file failed.", source)), err)
		}
		// TODO parse S3 URIs as well
		storageObj := ParseUrlToGoStorageObject(source)
		a.initializeGoStorage()
		a.gostorageClient.DownloadFile(storageObj, f.Name())
		localFilePath = f.Name()
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
	// TODO check if exists
	a.gostorageClient = gostorage.GoStorage{} // TODO
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

// IsProviderStorageUrl checks if the given string is a valid file URL for a storage service of one of the
// supported storage providers.
// Currently, this function returns true if the given URL is an S3 or Google Cloud Storage URL/URI.
// This function should be extended when adding new providers.
func IsProviderStorageUrl(url string) bool {

	// TODO dynamically
	/*
		for _, provider := range providers.GetAllProviders() {
		}
	*/

	return IsAWSUrl(url) || IsGoogleUrl(url)
}
