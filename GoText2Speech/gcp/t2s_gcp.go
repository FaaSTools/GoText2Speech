package gcp

import (
	"errors"
	"fmt"
	. "goTest/GoText2Speech/shared"
)

type T2SGoogleCloudPlatform struct {
	credentials CredentialsHolder
	t2sClient   any // TODO set correct type
}

// AudioFormatToGCPValue Converts the given AudioFormat into a valid format that can be used on GCP.
// If AudioFormat is unspecified, mp3 will be used.
// If AudioFormat is not supported on GCP, an error is thrown.
// Available audio formats on GCP can be seen here: https://pkg.go.dev/cloud.google.com/go/texttospeech@v1.6.0/apiv1/texttospeechpb#AudioEncoding
func AudioFormatToGCPValue(format AudioFormat) (int, error) { // TODO use AudioEncoding type
	switch format {
	case AudioFormatUnspecified:
		fallthrough
	case AudioFormatMp3:
		return 2, nil
	case AudioFormatOgg:
		return 3, nil
	case AudioFormatLinear16:
		return 1, nil
	case AudioFormatMulaw:
		return 5, nil
	case AudioFormatAlaw:
		return 6, nil
	default:
		return 0, errors.New("the specified audio format " + string(format) + " is not available on GCP. Either choose a different audio format, choose a different provider or use the TextToSpeechOptions.OutputFormatRaw property to bypass format check.")
	}
}

// GCPValueToAudioFormat Reverse of AudioFormatToGCPValue function.
// It gets a rawFormat value and returns the corresponding AudioFormat enum value.
// If enum value couldn't be found or if the specified rawFormat is undefined/empty, an error is returned.
func GCPValueToAudioFormat(rawFormat int) (AudioFormat, error) {
	if rawFormat < 1 {
		return "", errors.New("the specified rawFormat was undefined")
	}
	for _, audioFormat := range GetAllAudioFormats() {
		a, _ := AudioFormatToGCPValue(audioFormat)
		if a == rawFormat {
			return audioFormat, nil
		}
	}
	return "", errors.New(fmt.Sprintf("the specified rawFormat %d has no defined AudioFormat value.", rawFormat))
}

var gcpSupportedAudioFormats = []AudioFormat{
	AudioFormatMp3,
	AudioFormatOgg,
	AudioFormatLinear16,
	AudioFormatMulaw,
	AudioFormatAlaw,
}

func (a T2SGoogleCloudPlatform) GetSupportedAudioFormats() []AudioFormat {
	return gcpSupportedAudioFormats
}

func (a T2SGoogleCloudPlatform) TransformOptions(text string, options TextToSpeechOptions) (string, TextToSpeechOptions, error) {
	fmt.Println("Not yet (fully) implemented")

	// TODO implement

	if options.OutputFormatRaw == nil {
		outputFormatRaw, audioFormatError := AudioFormatToGCPValue(options.OutputFormat)
		if audioFormatError != nil {
			return text, options, audioFormatError
		}
		options.OutputFormatRaw = outputFormatRaw
	}

	return text, options, nil
}

func (a T2SGoogleCloudPlatform) IsURLonOwnStorage(url string) bool {
	return IsGoogleUrl(url)
}

func (a T2SGoogleCloudPlatform) CreateTempDestination(tempBucket string, fileName string) string {
	fmt.Println("Not yet implemented")
	// TODO
	return ""
}

func (a T2SGoogleCloudPlatform) FindVoice(options TextToSpeechOptions) (*VoiceIdConfig, error) {
	fmt.Println("Not yet implemented")
	// TODO implement
	return nil, nil
}

func (a T2SGoogleCloudPlatform) CreateServiceClient(credentials CredentialsHolder, region string) T2SProvider {
	fmt.Println("Not yet implemented")
	// TODO implement
	return a
}

func (a T2SGoogleCloudPlatform) ExecuteT2SDirect(text string, destination string, options TextToSpeechOptions) error {
	fmt.Println("Not yet implemented")
	// TODO implement
	return nil
}

func (a T2SGoogleCloudPlatform) ExecuteT2S(source string, destination string, options TextToSpeechOptions) error {
	fmt.Println("Not yet implemented")
	// TODO implement
	return nil
}
