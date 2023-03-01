package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
)

import (
	ts2_aws "goTest/GoText2Speech/aws"
	ts2_gcp "goTest/GoText2Speech/gcp"
	"goTest/GoText2Speech/providers"
	. "goTest/GoText2Speech/shared"
)

func main() {
	fmt.Printf("Hello World\n")

	r, _ := regexp.Compile("volume=\"(?P<volume>.*)db\"")
	str := r.FindStringSubmatch("<speak volume=\"10.00db\" >")
	fmt.Printf("%#v: ", str)
	fmt.Println()
	fmt.Printf("%#v: ", r.SubexpNames())
	fmt.Println()

	fmt.Printf("openingTag: %s\n", IntegrateVolumeAttributeValueIntoTag("<speak volume=\"5db\">", 10))
	fmt.Printf("openingTag: %s\n", IntegrateVolumeAttributeValueIntoTag("<speak volume=\"loud\">", 10))
	fmt.Printf("openingTag: %s\n", IntegrateVolumeAttributeValueIntoTag("<speak rate=\"5%\" volume=\"5db\">", 10))
	fmt.Printf("openingTag: %s\n", IntegrateVolumeAttributeValueIntoTag("<speak rate=\"5%\" volume=\"loud\">", 10))
	fmt.Printf("openingTag: %s\n", IntegrateVolumeAttributeValueIntoTag("<speak rate=\"5%\">", 10))
	fmt.Printf("openingTag: %s\n", IntegrateVolumeAttributeValueIntoTag("<speak>", 10))

	/*
		svc := CreatePollyClient()
		input := &polly.DescribeVoicesInput{LanguageCode: aws.String("en-US")}
		resp, err := svc.DescribeVoices(input)

		if err != nil {
			fmt.Printf("Error:" + err.Error())
			os.Exit(1)
		}

		for _, v := range resp.Voices {
			fmt.Printf("Name:\t" + *v.Name + "\n")
			fmt.Printf("Gender:\t" + *v.Gender + "\n\n")
		}
	*/

	fmt.Println("Synthesizing speech...")
	s := "Hello World, how are you today? Lovely day, isn't it?"

	err := T2SDirect(s, "testfile.mp3", TextToSpeechOptions{
		TextType:    TextTypeText,
		VoiceConfig: VoiceConfig{VoiceIdConfig: VoiceIdConfig{VoiceId: "Joanna"}},
		// Alternative voice config
		//VoiceConfig: VoiceConfig{VoiceParamsConfig{LanguageCode: "en-US", Gender: VoiceGender_FEMALE},
		SpeakingRate: 1.1,
		Pitch:        0.0,
		Volume:       0.0,
	})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Speech successfully synthesized!")
}

// T2SDirect creates text to speech input (AWS specific)
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
