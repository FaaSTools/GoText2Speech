package aws

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/polly"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	. "goTest/GoText2Speech/shared"
	"io"
	"strings"
)

type T2SAmazonWebServices struct {
	t2sClient *polly.Polly
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
			// integrate parameters into root speak element
			openingTag := GetOpeningTagOfSSMLText(text)
			openingTag = IntegrateVolumeAttributeValueIntoTag(openingTag, options.Volume)
			openingTag = IntegrateSpeakingRateAttributeValueIntoTag(openingTag, options.SpeakingRate*100)
			openingTag = IntegratePitchAttributeValueIntoTag(openingTag, options.Pitch*100)
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

// ChooseVoice chooses a voice for AWS Polly based on the given parameters (language and gender)
func (a T2SAmazonWebServices) ChooseVoice(options TextToSpeechOptions) (TextToSpeechOptions, error) {

	// Get list of available voices for the chosen language and pick the first one with the correct gender
	input := &polly.DescribeVoicesInput{LanguageCode: aws.String(options.VoiceConfig.VoiceParamsConfig.LanguageCode)}
	resp, err := a.t2sClient.DescribeVoices(input)

	if err != nil {
		return options, errors.New("Error while describing voices: " + err.Error())
	}

	voiceFound := false
	for _, v := range resp.Voices {
		if strings.EqualFold(*v.Gender, options.VoiceConfig.VoiceParamsConfig.Gender.String()) {
			options.VoiceConfig.VoiceIdConfig = VoiceIdConfig{VoiceId: *v.Name}
			voiceFound = true
			fmt.Printf("Found voice with language %s and gender %s: %s\n",
				options.VoiceConfig.VoiceParamsConfig.LanguageCode,
				options.VoiceConfig.VoiceParamsConfig.Gender.String(),
				*v.Name)
		}
	}

	if !voiceFound {
		errText := fmt.Sprintf("error: No voice found for language %s and gender %s\n",
			options.VoiceConfig.VoiceParamsConfig.LanguageCode,
			options.VoiceConfig.VoiceParamsConfig.Gender.String())
		return options, errors.New(errText)
	}

	return options, nil
}

func (a T2SAmazonWebServices) CreateClient() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	a.t2sClient = polly.New(sess)
}

// ExecuteT2SDirect executes Text-to-Speech using AWS Polly service. The given text is transformed into speech
// using the given options. The created audio file is uploaded to AWS S3 on the given destination.
// The destination string can either be a AWS S3 URI (starting with "s3://") or AWS S3 Object URL (starting with "https://").
func (a T2SAmazonWebServices) ExecuteT2SDirect(text string, destination string, options TextToSpeechOptions) error {

	outputFormatRaw, outputFormatAssertedCorrectly := options.OutputFormatRaw.(string)

	if !outputFormatAssertedCorrectly {
		outputFormatError := errors.New("the raw output format was not a string, but AWS can only use strings as output format")
		return outputFormatError
	}

	speechInput := polly.SynthesizeSpeechInput{
		OutputFormat: aws.String(outputFormatRaw),
		Text:         aws.String(text),
		VoiceId:      aws.String(options.VoiceConfig.VoiceIdConfig.VoiceId),
		TextType:     aws.String(options.TextType.String())}

	if options.SampleRate != 0 {
		speechInput.SetSampleRate(fmt.Sprintf("%d", options.SampleRate))
	}

	output, err := a.t2sClient.SynthesizeSpeech(&speechInput)

	if err != nil {
		errNew := errors.New("Error while synthesizing speech on AWS: " + err.Error())
		fmt.Printf(errNew.Error())
		return errNew
	}

	// TODO test
	if options.AddFileExtension {
		audioFormat, err := AWSValueToAudioFormat(outputFormatRaw)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			fmt.Printf("No file extension found for the specified raw audio format %s. No file extension is added to file name.\n", outputFormatRaw)
		} else {
			audioFormatStr := AudioFormatToFileExtension(audioFormat)
			if !strings.HasSuffix(destination, audioFormatStr) {
				destination += audioFormatStr
			}
		}
	}

	bucket, key, destinationFormatErr := GetBucketAndKeyFromAWSDestination(destination)
	if destinationFormatErr != nil {
		return destinationFormatErr
	}

	err = uploadFileToS3(output.AudioStream, bucket, key)
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

// UploadFileToS3 takes a file stream and uploads it to S3 using the given S3 bucket and key.
// Code adapted from AWS Docs (https://docs.aws.amazon.com/sdk-for-go/api/service/s3/#hdr-Upload_Managers)
func uploadFileToS3(fileContents io.Reader, bucket string, key string) error {
	// The session the S3 Uploader will use
	sess := session.Must(session.NewSession())

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   fileContents,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}
	fmt.Printf("file uploaded to, %s\n", result.Location)
	return nil
}

func (a T2SAmazonWebServices) ExecuteT2S(source string, destination string, options TextToSpeechOptions) error {
	// TODO
	return nil
}
