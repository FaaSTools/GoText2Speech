package gcp

import (
	"bytes"
	"cloud.google.com/go/storage"
	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"context"
	"errors"
	"fmt"
	. "github.com/FaaSTools/GoText2Speech/GoText2Speech/shared"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
)

type T2SGoogleCloudPlatform struct {
	credentials CredentialsHolder
	t2sClient   *texttospeech.Client
}

// AudioFormatToGCPValue Converts the given AudioFormat into a valid format that can be used on GCP.
// If AudioFormat is unspecified, mp3 will be used.
// If AudioFormat is not supported on GCP, an error is thrown.
// Available audio formats on GCP can be seen here: https://pkg.go.dev/cloud.google.com/go/texttospeech@v1.6.0/apiv1/texttospeechpb#AudioEncoding
func AudioFormatToGCPValue(format AudioFormat) (int16, error) {
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

func VoiceGenderToGCPGender(gender VoiceGender) texttospeechpb.SsmlVoiceGender {
	switch gender {
	case VoiceGenderFemale:
		return texttospeechpb.SsmlVoiceGender_FEMALE
	case VoiceGenderMale:
		return texttospeechpb.SsmlVoiceGender_MALE
	case VoiceGenderNeutral:
		return texttospeechpb.SsmlVoiceGender_NEUTRAL
	case VoiceGenderUnspecified:
		fallthrough
	default:
		return texttospeechpb.SsmlVoiceGender_SSML_VOICE_GENDER_UNSPECIFIED
	}
}

// GCPValueToAudioFormat Reverse of AudioFormatToGCPValue function.
// It gets a rawFormat value and returns the corresponding AudioFormat enum value.
// If enum value couldn't be found or if the specified rawFormat is undefined/empty, an error is returned.
func GCPValueToAudioFormat(rawFormat int16) (AudioFormat, error) {
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
	// on GCP, the pitch value is in range [-20.0, 20.0]. GoTextToSpeech pitch value is in [-1.0, 1.0].
	options.Pitch = math.Min(math.Max(options.Pitch, -1), 1) * 20.0

	if options.OutputFormatRaw == nil {
		fmt.Printf("Setting OutputFormatRaw\n")
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
	now := time.Now()
	return "https://storage.cloud.google.com/" + tempBucket + "/" + fileName + strconv.FormatInt(now.UnixNano(), 10)
}

func (a T2SGoogleCloudPlatform) FindVoice(options TextToSpeechOptions) (*VoiceIdConfig, error) {
	req := &texttospeechpb.ListVoicesRequest{
		LanguageCode: options.VoiceConfig.VoiceParamsConfig.LanguageCode,
	}
	resp, err := a.t2sClient.ListVoices(context.Background(), req)
	if err != nil {
		return nil, err // TODO wrap
	}

	gcpGender := VoiceGenderToGCPGender(options.VoiceConfig.VoiceParamsConfig.Gender)
	var gcpVoice *texttospeechpb.Voice = nil
	for _, voice := range resp.GetVoices() {
		if voice.GetSsmlGender() == gcpGender { // TODO engine?
			gcpVoice = voice
		}
	}

	if gcpVoice == nil {
		errText := fmt.Sprintf("error: No voice found for language %s and gender %s\n",
			options.VoiceConfig.VoiceParamsConfig.LanguageCode,
			options.VoiceConfig.VoiceParamsConfig.Gender.String())
		return nil, errors.New(errText)
	}

	fmt.Printf("Voice found for language %s and gender %s: %s\n",
		options.VoiceConfig.VoiceParamsConfig.LanguageCode,
		options.VoiceConfig.VoiceParamsConfig.Gender.String(),
		gcpVoice.GetName())

	voiceConfig := &VoiceIdConfig{
		VoiceId: gcpVoice.GetName(),
	}
	return voiceConfig, nil
}

func (a T2SGoogleCloudPlatform) CreateServiceClient(credentials CredentialsHolder, region string) (T2SProvider, error) {
	ctx := context.Background()
	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		return a, err
	}
	a.t2sClient = client
	return a, nil
}

func (a T2SGoogleCloudPlatform) AddFileExtensionToDestinationIfNeeded(options TextToSpeechOptions, outputFormatRaw any, destination string) (string, error) {
	if options.AddFileExtension {
		//audioFormat, err := GCPValueToAudioFormat(texttospeechpb.AudioEncoding(outputFormatRaw.(int)))
		audioFormat, err := GCPValueToAudioFormat(outputFormatRaw.(int16))
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			errNew := errors.New(fmt.Sprintf("No file extension found for the specified raw audio format %d. No file extension is added to file name.\n", outputFormatRaw.(int)))
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

func GetBucketAndKeyFromCLoudStorageDestination(destination string) (string, string, error) {
	// TODO streamline; use GCP parser directly
	storageObj := ParseUrlToGoStorageObject(destination)
	return storageObj.Bucket, storageObj.Key, nil
}

func (a T2SGoogleCloudPlatform) ExecuteT2SDirect(text string, destination string, options TextToSpeechOptions) (io.Reader, error) {
	var input *texttospeechpb.SynthesisInput = nil
	if options.TextType == TextTypeSsml {
		inputSource := &texttospeechpb.SynthesisInput_Ssml{
			Ssml: text,
		}
		input = &texttospeechpb.SynthesisInput{
			InputSource: inputSource,
		}
	} else {
		inputSource := &texttospeechpb.SynthesisInput_Text{
			Text: text,
		}
		input = &texttospeechpb.SynthesisInput{
			InputSource: inputSource,
		}
	}

	req := texttospeechpb.SynthesizeSpeechRequest{
		Input: input,
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: options.VoiceConfig.VoiceIdConfig.VoiceId[0:5], // TODO assuming all voices start with language code
			Name:         options.VoiceConfig.VoiceIdConfig.VoiceId,
			//SsmlGender:   0,
			//CustomVoice:  nil,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding(options.OutputFormatRaw.(int16)),
			//AudioEncoding:    options.OutputFormatRaw,
			SpeakingRate:     options.SpeakingRate,
			Pitch:            options.Pitch,
			VolumeGainDb:     options.Volume,
			SampleRateHertz:  options.SampleRate,
			EffectsProfileId: options.AudioEffects,
		},
	}

	result, err := a.t2sClient.SynthesizeSpeech(context.Background(), &req)
	if err != nil {
		return nil, err
	}
	fmt.Printf("GCP speech synthesis successful!\n")

	stream := bytes.NewReader(result.GetAudioContent())
	return stream, nil

	/*
		bucket, key, destinationFormatErr := GetBucketAndKeyFromCLoudStorageDestination(destination)
		if destinationFormatErr != nil {
			return destinationFormatErr
		}

		uploadErr := a.uploadFileToCS(bytes.NewReader(result.GetAudioContent()), bucket, key)
		if uploadErr != nil {
			return uploadErr // TODO wrap
		}

		return nil
	*/
}

func (a T2SGoogleCloudPlatform) UploadFile(fileContents io.Reader, destination string) error {
	bucket, key, destinationFormatErr := GetBucketAndKeyFromCLoudStorageDestination(destination)
	if destinationFormatErr != nil {
		return destinationFormatErr
	}
	return a.uploadFileToCS(fileContents, bucket, key)
}

// uploadFileToCS takes a file stream and uploads it to Google Cloud Storage using the given CS bucket and key.
// Inspired by Google Cloud Storage examples (https://cloud.google.com/storage/docs/uploading-objects#permissions-client-libraries).
func (a T2SGoogleCloudPlatform) uploadFileToCS(fileContents io.Reader, bucket string, key string) error {

	fmt.Printf("Uploading file...\n")

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return errors.Join(errors.New(fmt.Sprintf("Error while uploading file '%s' on bucket '%s' to Google Cloud Storage.", key, bucket)), err)
	}

	defer client.Close()

	cloudObj := client.Bucket(bucket).Object(key)

	wc := cloudObj.NewWriter(ctx)
	if _, err = io.Copy(wc, fileContents); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %w", err)
	}
	fmt.Printf("file uploaded to, %s/%s\n", bucket, key)
	return nil

}

func (a T2SGoogleCloudPlatform) CloseServiceClient() error {
	if a.t2sClient == nil {
		fmt.Println("Warning: Couldn't close GCP service client, because client doesn't exist.")
		return nil
	}
	return a.t2sClient.Close()
}
