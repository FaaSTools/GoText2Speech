// This files defines provider-independent option types for Text-to-Speech operations

package shared

import (
	"github.com/FaaSTools/GoText2Speech/GoText2Speech/providers"
)

type VoiceGender int16

const (
	// VoiceGenderUnspecified unspecified gender means that there is no preference for gender
	// and that the default gender will be selected.
	VoiceGenderUnspecified VoiceGender = iota
	// VoiceGenderMale adult male voice. Available on AWS and GCP.
	VoiceGenderMale
	// VoiceGenderFemale adult female voice. Available on AWS and GCP.
	VoiceGenderFemale
	// VoiceGenderNeutral gender-neutral voice. Defined only by GCP, but not yet supported.
	VoiceGenderNeutral
)

func (voiceGender VoiceGender) String() string {
	switch voiceGender {
	case VoiceGenderFemale:
		return "Female"
	case VoiceGenderMale:
		return "Male"
	case VoiceGenderUnspecified:
		return "Unspecified"
	case VoiceGenderNeutral:
		return "Neutral"
	//case VoiceGenderMaleChild:
	//	return "Male_Child"
	//case VoiceGenderFemaleChild:
	//	return "Female_Child"
	default:
		return ""
	}
}

// VoiceIdConfig Defines the voice ID and Engine that should be used for speech synthesis.
// A voice ID indirectly specifies the gender and language of a voice.
// For example, the voice ID "Joanna" is a female en-US voice for AWS.
// Generally, voice IDs are different for different providers
// (i.e. the voice ID "Joanna" exists for AWS, but not GCP).
// If the Engine parameter is left empty, the default engine of the chosen provider will be used.
type VoiceIdConfig struct {
	VoiceId string
	Engine  string
}

// VoiceParamsConfig Defines parameters of a voice that should be used for speech synthesis.
type VoiceParamsConfig struct {
	// LanguageCode The language identification tag (ISO 639 code for the language name-ISO 3166
	// country code) for filtering the list of voices returned.
	LanguageCode string
	Gender       VoiceGender
	Engine       string
}

type TextType string

const (
	// TextTypeSsml is a TextType enum value
	TextTypeSsml TextType = "ssml"
	// TextTypeText is a TextType enum value
	TextTypeText TextType = "text"
	// TextTypeAuto is a TextType enum value
	TextTypeAuto TextType = "auto"
)

func (t TextType) String() string {
	return string(t)
}

// VoiceConfig Either specify VoiceIdConfig or VoiceParamsConfig.
// When VoiceIdConfig is specified with its VoiceIdConfig.VoiceId value, the voice with the specified VoiceId is used
// and VoiceParamsConfig gets ignored. In that case, the optional VoiceIdConfig.Engine specifies the voice engine to use.
// If no engine is specified (i.e. empty string), the default engine for the provider is used.
// The T2S function doesn't check if the VoiceId actually exists. So if the VoiceId doesn't exist for the specified
// provider, the provider's error is thrown.
//
// When VoiceIdConfig is undefined or empty (see VoiceIdConfig.IsEmpty function), then a voice based on the
// VoiceParamsConfig is selected.
// VoiceParamsConfig specifies the language, gender and engine of the voice that should be selected. The T2S function automatically
// chooses the first voice id with the specified language, gender and engine parameters.
//
// If VoiceParamsConfig is undefined as well, the default value from GetDefaultVoiceParamsConfig is used.
// If one of the properties of VoiceParamsConfig is undefined (empty string for LanguageCode and VoiceGenderUnspecified
// for Gender), the corresponding value from GetDefaultVoiceParamsConfig is used.
// If the Engine parameter is undefined (empty string), engine will be ignored for choosing voice.
type VoiceConfig struct {
	_                 struct{}
	VoiceIdConfig     VoiceIdConfig
	VoiceParamsConfig VoiceParamsConfig
}

// AudioFormat See which output formats are available on each provider in the respective documentation:
// AWS Doc: https://docs.aws.amazon.com/polly/latest/dg/API_SynthesizeSpeech.html#polly-SynthesizeSpeech-request-OutputFormat
// GCP Doc: https://pkg.go.dev/cloud.google.com/go/texttospeech@v1.6.0/apiv1/texttospeechpb#AudioEncoding
type AudioFormat string

// The defined strings for AudioFormat are just for text logs, they shouldn't be used as the raw output format for
// the chosen provider.
const (
	// AudioFormatUnspecified If audio format is unspecified, the default audio format will be used.
	// The default audio format is defined per provider, for example in the `AudioFormatToAWSValue` function for AWS
	// or the `AudioFormatToGCPValue` function for GCP.
	AudioFormatUnspecified AudioFormat = ""
	// AudioFormatMp3 Available on both AWS and GCP
	AudioFormatMp3 AudioFormat = "mp3"
	// AudioFormatOgg Available on both AWS (as OGG vorbis) and GCP (as OGG opus)
	AudioFormatOgg AudioFormat = "ogg"
	// AudioFormatPcm Available only on AWS
	AudioFormatPcm AudioFormat = "pcm"
	// AudioFormatJson Available only on AWS
	AudioFormatJson AudioFormat = "json"
	// AudioFormatLinear16 Available only on GCP
	AudioFormatLinear16 AudioFormat = "linear16"
	// AudioFormatMulaw Available only on GCP
	AudioFormatMulaw AudioFormat = "mulaw"
	// AudioFormatAlaw Available only on GCP
	AudioFormatAlaw AudioFormat = "alaw"
)

func GetAllAudioFormats() []AudioFormat {
	return []AudioFormat{
		AudioFormatMp3,
		AudioFormatOgg,
		AudioFormatPcm,
		AudioFormatJson,
		AudioFormatLinear16,
		AudioFormatMulaw,
		AudioFormatAlaw,
	}
}

func AudioFormatToFileExtension(audioFormat AudioFormat) string {
	switch audioFormat {
	case AudioFormatMp3:
		return ".mp3"
	case AudioFormatOgg:
		return ".ogg"
	case AudioFormatJson:
		return ".json"
	case AudioFormatPcm:
		fallthrough
	case AudioFormatAlaw:
		fallthrough
	case AudioFormatMulaw:
		fallthrough
	case AudioFormatLinear16:
		return ".wav"
	case AudioFormatUnspecified:
		fallthrough
	default:
		return ""
	}
}

type TextToSpeechOptions struct {
	_           struct{}
	Provider    providers.Provider
	TextType    TextType
	VoiceConfig VoiceConfig
	// SpeakingRate 1.0 is normal speed, 0.5 is half speed, 2.0 is double speed
	SpeakingRate float64
	// Pitch 0.0 is normal pitch, 0.05 is a little higher pitch, -0.05 a little lower pitch. Recommended range: [-1.0, 1.0]
	Pitch float64
	// Volume increase in dB. 0.0 is normal, 6.0 is approximately double the normal volume, -6.0 is half the normal volume.
	// Recommended range: [-96.0, 16.0]
	Volume float64
	// AudioEffects only available on Google Cloud Platform.
	// See documentation for more information: https://cloud.google.com/text-to-speech/docs/audio-profiles
	AudioEffects []string
	// SampleRate in Hz. Not all values are supported for all audio encodings.
	// For more information, see documentation of cloud provider.
	// AWS Doc: https://docs.aws.amazon.com/polly/latest/dg/API_SynthesisTask.html
	// GCP Doc: https://pkg.go.dev/cloud.google.com/go/texttospeech@v1.6.0/apiv1/texttospeechpb#AudioConfig
	// If SampleRate is 0, the default value for the provider is selected
	SampleRate int32
	// OutputFormat Each provider allows different output formats.
	// AWS Doc: https://docs.aws.amazon.com/polly/latest/dg/API_SynthesizeSpeech.html#polly-SynthesizeSpeech-request-OutputFormat
	// GCP Doc: https://pkg.go.dev/cloud.google.com/go/texttospeech@v1.6.0/apiv1/texttospeechpb#AudioEncoding
	OutputFormat AudioFormat
	// OutputFormatRaw The raw output format that is directly given to the t2s function of the chosen provider.
	// It can be used to overwrite the OutputFormat value and thereby bypass the enum type check.
	// If this is specified, OutputFormat is ignored.
	// OutputFormatRaw is not used to determine what provider to choose. The t2s functions don't check if the value
	// of OutputFormatRaw is allowed for the chosen provider. So, only use this property if you know what you are doing.
	OutputFormatRaw any
	// AddFileExtension If true, the appropriate file extension for the chosen OutputFormat is automatically appended
	// to the file name (only if that exact file extension is not already the suffix of the filename).
	AddFileExtension bool
}

func GetDefaultTextToSpeechOptions() *TextToSpeechOptions {
	return &TextToSpeechOptions{
		Provider: providers.ProviderUnspecified,
		TextType: TextTypeAuto,
		VoiceConfig: VoiceConfig{
			VoiceParamsConfig: GetDefaultVoiceParamsConfig(),
		},
		SpeakingRate:     1.0,
		Pitch:            0,
		Volume:           0,
		AudioEffects:     nil,
		SampleRate:       0,
		OutputFormat:     AudioFormatUnspecified,
		OutputFormatRaw:  nil,
		AddFileExtension: true,
	}
}

// GetDefaultVoiceParamsConfig The default value for VoiceParamsConfig
func GetDefaultVoiceParamsConfig() VoiceParamsConfig {
	return VoiceParamsConfig{
		LanguageCode: "en-US",
		Gender:       VoiceGenderMale,
		Engine:       "",
	}
}

// IsEmpty checks if VoiceIdConfig.VoiceId is defined.
// Since the VoiceIdConfig.Engine parameter is optional, its presence doesn't make the VoiceIdConfig not-empty.
func (config VoiceIdConfig) IsEmpty() bool {
	return (config == (VoiceIdConfig{})) || (config.VoiceId == "")
}

func (config VoiceParamsConfig) IsEmpty() bool {
	return (config == (VoiceParamsConfig{})) || ((config.LanguageCode == "") && (config.Gender == VoiceGenderUnspecified))
}
