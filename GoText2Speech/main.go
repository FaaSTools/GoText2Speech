package GoText2Speech

import (
	"errors"
	"fmt"
	ts2_aws "goTest/GoText2Speech/aws"
	ts2_gcp "goTest/GoText2Speech/gcp"
	"goTest/GoText2Speech/providers"
	. "goTest/GoText2Speech/shared"
)

// T2SDirect Transforms the given text into speech and stores the file in destination.
// If the given options specify a provider, this provider will be used.
// If the given options don't specify a provider, a provider will be chosen based on heuristics.
func T2SDirect(text string, destination string, options TextToSpeechOptions) error {

	// error check: If the given text is supposed to be a SSML text and does not contain <speak>-tags, it is invalid.
	if (options.TextType == TextTypeSsml) && !HasSpeakTag(text) {
		return errors.New("invalid text. The text type was SSML, but the given text didn't contain <speak>-tags")
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

	provider := CreateProviderInstance(options.Provider)
	provider.CreateClient()

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

		newOptions, chooseVoiceErr := provider.ChooseVoice(options)
		if chooseVoiceErr != nil {
			return chooseVoiceErr
		}
		options = newOptions
	}

	// adjust parameters for Google/AWS
	var transformOptionsError error
	text, options, transformOptionsError = provider.TransformOptions(text, options)

	if transformOptionsError != nil {
		return transformOptionsError
		// TODO throw error
	}

	fmt.Println("Final Text: " + text)

	t2sErr := provider.ExecuteT2SDirect(text, destination, options)
	return t2sErr
}

// T2S Transforms the text in the source file into speech and stores the file in destination.
// If the given options specify a provider, this provider will be used.
// If the given options don't specify a provider, a provider will be chosen based on heuristics.
func T2S(source string, destination string, options TextToSpeechOptions) error {
	// TODO implement
	return nil
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
