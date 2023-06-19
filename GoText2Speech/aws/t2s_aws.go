package aws

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	. "github.com/FaaSTools/GoText2Speech/GoText2Speech/shared"
	"github.com/aws/aws-sdk-go-v2/service/polly/types"
	//"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/polly"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"strconv"
	"strings"
	"time"
)

type T2SAmazonWebServices struct {
	credentials CredentialsHolder
	t2sClient   *polly.Client
	region      string
	//sess        client.ConfigProvider
}

// AudioFormatToAWSValue Converts the given AudioFormat into a valid format that can be used on AWS.
// If AudioFormat is unspecified, mp3 will be used.
// If AudioFormat is not supported on AWS, an error is thrown.
// Available audio formats on AWS can be seen here: https://docs.aws.amazon.com/polly/latest/dg/API_SynthesizeSpeech.html#polly-SynthesizeSpeech-request-OutputFormat
func AudioFormatToAWSValue(format AudioFormat) (string, error) {
	switch format {
	case AudioFormatUnspecified:
		fallthrough
	case AudioFormatMp3:
		return "mp3", nil
	case AudioFormatOgg:
		return "ogg_vorbis", nil
	case AudioFormatJson:
		return "json", nil
	case AudioFormatPcm:
		return "pcm", nil
	default:
		return "", errors.New("the specified audio format " + string(format) + " is not available on AWS. Either choose a different audio format, choose a different provider or use the TextToSpeechOptions.OutputFormatRaw property to  format check.")
	}
}

// AWSValueToAudioFormat Reverse of AudioFormatToAWSValue function.
// It gets a rawFormat value and returns the corresponding AudioFormat enum value.
// If enum value couldn't be found or if the specified rawFormat is undefined/empty, an error is returned.
func AWSValueToAudioFormat(rawFormat string) (AudioFormat, error) {
	if strings.EqualFold(rawFormat, "") {
		return "", errors.New("the specified rawFormat was empty")
	}
	for _, audioFormat := range GetAllAudioFormats() {
		a, _ := AudioFormatToAWSValue(audioFormat)
		if strings.EqualFold(a, rawFormat) {
			return audioFormat, nil
		}
	}
	return "", errors.New("the specified rawFormat " + rawFormat + " has no defined AudioFormat value.")
}

var awsSupportedAudioFormats = []AudioFormat{
	AudioFormatMp3,
	AudioFormatOgg,
	AudioFormatPcm,
	AudioFormatJson,
}

func (a T2SAmazonWebServices) GetSupportedAudioFormats() []AudioFormat {
	return awsSupportedAudioFormats
}

func (a T2SAmazonWebServices) IsURLonOwnStorage(url string) bool {
	return IsAWSUrl(url)
}

func (a T2SAmazonWebServices) CreateTempDestination(tempBucket string, fileName string) string {
	now := time.Now()
	return "https://" + tempBucket + ".s3.amazonaws.com/" + fileName + strconv.FormatInt(now.UnixNano(), 10)
}

// TransformOptions
// This function assumes that the basic options check was already executed,
// i.e. that options.TextType cannot be TextTypeAuto and that if the text is SSML text, it contains correctly formed <speak>-tags.
func (a T2SAmazonWebServices) TransformOptions(text string, options TextToSpeechOptions) (string, TextToSpeechOptions, error) {
	// if these modifiers are defined, the text needs to be wrapped in SSML <speak>...</speak> tags to add those modifiers
	if SSMLModifiersDefined(options) {
		if options.TextType == TextTypeText {
			text = TransformTextIntoSSML(text, options)
			options.TextType = TextTypeSsml
		} else {
			// integrate parameters into a new <prosody> tag
			openingTag := GetOpeningTagOfSSMLText(text)
			settingsTag := CreateProsodyTag(options)
			text = openingTag + settingsTag + RemoveClosingSpeakTagOfSSMLText(RemoveOpeningTagOfSSMLText(text)) + "</prosody></speak>"
		}
	}

	if options.OutputFormatRaw == nil {
		outputFormatRaw, audioFormatError := AudioFormatToAWSValue(options.OutputFormat)
		if audioFormatError != nil {
			return text, options, audioFormatError
		}
		options.OutputFormatRaw = outputFormatRaw
	}

	return text, options, nil
}

// FindVoice finds a voice for AWS Polly based on the given parameters (language, gender and optionally engine).
// If voice was found, returns VoiceIdConfig object on which VoiceId and Engine is set (and nil as error).
// If a voice with the needed parameters is not found, returns nil and error.
func (a T2SAmazonWebServices) FindVoice(options TextToSpeechOptions) (*VoiceIdConfig, error) {

	// Get list of available voices for the chosen language and pick the first one with the correct gender and engine
	input := &polly.DescribeVoicesInput{LanguageCode: types.LanguageCode(options.VoiceConfig.VoiceParamsConfig.LanguageCode)}
	resp, err := a.t2sClient.DescribeVoices(context.Background(), input)

	if err != nil {
		return nil, errors.New("Error while describing voices: " + err.Error())
	}

	targetEngine := options.VoiceConfig.VoiceParamsConfig.Engine
	var voiceConfig *VoiceIdConfig = nil
	for _, v := range resp.Voices {
		if strings.EqualFold(string(v.Gender), options.VoiceConfig.VoiceParamsConfig.Gender.String()) {

			// if engine param is specified, check if engine is supported for this voice
			if !strings.EqualFold("", targetEngine) {
				voiceWithEngineFound := false
				for _, e := range v.SupportedEngines {
					if strings.EqualFold(string(e), targetEngine) {
						voiceWithEngineFound = true
						break
					}
				}
				if !voiceWithEngineFound {
					continue
				}
			}

			voiceConfig = &VoiceIdConfig{
				VoiceId: *v.Name,
				Engine:  targetEngine,
			}
			fmt.Printf("Found voice with language %s, gender %s and engine %s: %s\n",
				options.VoiceConfig.VoiceParamsConfig.LanguageCode,
				options.VoiceConfig.VoiceParamsConfig.Gender.String(),
				targetEngine,
				*v.Name)
		}
	}

	if voiceConfig != nil {
		errText := fmt.Sprintf("error: No voice found for language %s and gender %s\n",
			options.VoiceConfig.VoiceParamsConfig.LanguageCode,
			options.VoiceConfig.VoiceParamsConfig.Gender.String())
		return nil, errors.New(errText)
	}

	return voiceConfig, nil
}

type CredentialsProvider struct {
	credentials aws.Credentials
}

func (b CredentialsProvider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	return b.credentials, nil
}

func (a T2SAmazonWebServices) CreateServiceClient(cred CredentialsHolder, region string) (T2SProvider, error) {
	credProv := CredentialsProvider{
		credentials: *cred.AwsCredentials,
	}
	a.credentials = cred
	a.region = region
	a.t2sClient = polly.New(polly.Options{
		Credentials: credProv,
		Region:      region,
	})
	return a, nil
}

func (a T2SAmazonWebServices) AddFileExtensionToDestinationIfNeeded(options TextToSpeechOptions, outputFormatRaw any, destination string) (string, error) {
	if options.AddFileExtension {
		audioFormat, err := AWSValueToAudioFormat(outputFormatRaw.(string))
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			errNew := errors.New(fmt.Sprintf("No file extension found for the specified raw audio format %s. No file extension is added to file name.\n", outputFormatRaw.(string)))
			return destination, errors.Join(err, errNew)
		} else {
			audioFormatStr := AudioFormatToFileExtension(audioFormat)
			if !strings.HasSuffix(destination, audioFormatStr) {
				destination += audioFormatStr
			}
		}
	}
	return destination, nil
}

// ExecuteT2SDirect executes Text-to-Speech using AWS Polly service. The given text is transformed into speech
// using the given options. The created audio file is uploaded to AWS S3 on the given destination.
// The destination string can either be an AWS S3 URI (starting with "s3://") or AWS S3 Object URL (starting with "https://").
func (a T2SAmazonWebServices) ExecuteT2SDirect(text string, destination string, options TextToSpeechOptions) (io.Reader, error) {

	outputFormatRaw, outputFormatAssertedCorrectly := options.OutputFormatRaw.(string)

	if !outputFormatAssertedCorrectly {
		outputFormatError := errors.New("the raw output format was not a string, but AWS can only use strings as output format")
		return nil, outputFormatError
	}

	speechInput := &polly.SynthesizeSpeechInput{
		OutputFormat: types.OutputFormat(outputFormatRaw),
		Text:         aws.String(text),
		VoiceId:      types.VoiceId(options.VoiceConfig.VoiceIdConfig.VoiceId),
		TextType:     types.TextType(options.TextType.String()),
	}
	if !strings.EqualFold("", options.VoiceConfig.VoiceIdConfig.Engine) {
		speechInput.Engine = types.Engine(options.VoiceConfig.VoiceIdConfig.Engine)
		//speechInput.SetEngine(options.VoiceConfig.VoiceIdConfig.Engine)
	}
	if options.SampleRate != 0 {
		s := fmt.Sprintf("%d", options.SampleRate)
		speechInput.SampleRate = &s
	}

	fmt.Printf("Synthesizing...\n")
	output, err := a.t2sClient.SynthesizeSpeech(context.Background(), speechInput)

	if err != nil {
		errNew := errors.New("Error while synthesizing speech on AWS: " + err.Error() + "\n")
		fmt.Printf(errNew.Error())
		return nil, errNew
	}
	fmt.Printf("Synthesizing done!\n")
	return output.AudioStream, nil

	/*
		destination, err = AddFileExtensionToDestinationIfNeeded(options, outputFormatRaw, destination)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			// not a fatal error -> just resume code
		}

		bucket, key, destinationFormatErr := GetBucketAndKeyFromAWSDestination(destination)
		if destinationFormatErr != nil {
			return destinationFormatErr
		}

		err = a.uploadFileToS3(output.AudioStream, bucket, key)
		errClose := output.AudioStream.Close()
		if err != nil {
			errNew := errors.New("Error while uploading file on AWS S3: " + err.Error())
			fmt.Printf(errNew.Error())
			return errNew
		}
		if errClose != nil {
			errNew := errors.New("Error while closing speech synthesis audio stream: " + errClose.Error())
			fmt.Printf(errNew.Error())
			return errNew
		}

		return nil
	*/
}

// GetBucketAndKeyFromAWSDestination receives either an AWS S3 URI (starting with "s3://") or
// AWS S3 Object URL (starting with "https://") and returns the bucket and key (without preceding slash) of the file.
// If the given destination is not valid, then two empty strings and an error is returned.
func GetBucketAndKeyFromAWSDestination(destination string) (string, string, error) {
	if strings.HasPrefix(destination, "s3://") {
		withoutPrefix, _ := strings.CutPrefix(destination, "s3://")
		bucket := strings.Split(withoutPrefix, "/")[0]
		key, _ := strings.CutPrefix(withoutPrefix, bucket+"/")
		return bucket, key, nil
	} else if strings.HasPrefix(destination, "https://") && strings.Contains(destination, "s3") {
		withoutPrefix, _ := strings.CutPrefix(destination, "https://")
		dotSplits := strings.SplitN(withoutPrefix, ".", 3)
		bucket := dotSplits[0]
		key := strings.SplitN(dotSplits[2], "/", 2)[1]
		return bucket, key, nil
	} else {
		return "", "", errors.New(fmt.Sprintf("The given destination '%s' is not a valid S3 URI or S3 Object URL.", destination))
	}
}

func (a T2SAmazonWebServices) UploadFile(fileContents io.Reader, destination string) error {
	bucket, key, destinationFormatErr := GetBucketAndKeyFromAWSDestination(destination)
	if destinationFormatErr != nil {
		return destinationFormatErr
	}
	return a.uploadFileToS3(fileContents, bucket, key)
}

// UploadFileToS3 takes a file stream and uploads it to S3 using the given S3 bucket and key.
// Code adapted from AWS Docs (https://docs.aws.amazon.com/sdk-for-go/api/service/s3/#hdr-Upload_Managers)
func (a T2SAmazonWebServices) uploadFileToS3(fileContents io.Reader, bucket string, key string) error {

	fmt.Printf("Uploading file...\n")

	// Create an uploader with the session and default options
	uploader := s3.New(s3.Options{
		Credentials: CredentialsProvider{
			credentials: *a.credentials.AwsCredentials,
		},
		Region: a.region,
	})

	buf := new(bytes.Buffer)
	_, err1 := buf.ReadFrom(fileContents)
	if err1 != nil {
		fmt.Println(err1.Error())
	}

	// Upload the file to S3.
	_, err := uploader.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          buf,
		ContentLength: int64(buf.Len()),
	})

	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}
	fmt.Printf("file uploaded to, %s/%s\n", bucket, key)
	return nil
}

func (a T2SAmazonWebServices) CloseServiceClient() error {
	// Doesn't do anything, because AWS Polly Client cannot be closed
	return nil
}
