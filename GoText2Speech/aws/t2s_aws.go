package aws

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/polly"
	. "goTest/GoText2Speech/shared"
	"io"
	"os"
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
			openingTag = IntegrateSpeakingRateAttributeValueIntoTag(openingTag, options.SpeakingRate)
			openingTag = IntegratePitchAttributeValueIntoTag(openingTag, options.Pitch)
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
			destination += AudioFormatToFileExtension(audioFormat)
		}
	}

	// TODO create file on AWS S3
	outFile, err := os.Create(destination)

	if err != nil {
		errNew := errors.New("Error creating file: " + err.Error())
		fmt.Printf(errNew.Error())
		return errNew
	}

	defer outFile.Close()
	_, err = io.Copy(outFile, output.AudioStream)
	if err != nil {
		errNew := errors.New("Error writing mp3 file: " + err.Error())
		fmt.Printf(errNew.Error())
		return errNew
	}

	return nil
}

func (a T2SAmazonWebServices) ExecuteT2S(source string, destination string, options TextToSpeechOptions) error {
	// TODO
	return nil
}
