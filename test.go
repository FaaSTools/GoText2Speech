package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/polly"
	"regexp"
)

import (
	"github.com/aws/aws-sdk-go/aws/session"
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
	T2SDirect(s, TextToSpeechOptions{
		TextType:    TextTypeText,
		VoiceConfig: VoiceConfig{VoiceIdConfig: VoiceIdConfig{VoiceId: "Joanna"}},
		// Alternative voice config
		//VoiceConfig: VoiceConfig{VoiceParamsConfig{LanguageCode: "en-US", Gender: VoiceGender_FEMALE},
		SpeakingRate: 1.1,
		Pitch:        0.0,
		Volume:       0.0,
	})

	/*
		output, err := svc.SynthesizeSpeech(&input2)

		if err != nil {
			fmt.Printf("Error:" + err.Error())
			os.Exit(1)
		}

		outFile, err := os.Create("test.mp3")

		if err != nil {
			fmt.Println("Error creating file: " + err.Error())
			os.Exit(1)
		}

		defer outFile.Close()
		_, err = io.Copy(outFile, output.AudioStream)
		if err != nil {
			fmt.Println("Error writing mp3 file: " + err.Error())
			os.Exit(1)
		}

		fmt.Println("Speech synthesized and mp3 file written")
	*/
}

func CreatePollyClient() *polly.Polly {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := polly.New(sess)
	return svc
}

// T2SDirect creates text to speech input (AWS specific)
func T2SDirect(text string, options TextToSpeechOptions) polly.SynthesizeSpeechInput {

	// error check: If the given text is supposed to be a SSML text and does not contain <speak>-tags, it is invalid.
	if (options.TextType == TextTypeSsml) && !HasSpeakTag(text) {
		// TODO throw error
		return polly.SynthesizeSpeechInput{}
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

		newOptions, err := provider.ChooseVoice(options)
		if err != nil {
			fmt.Print(err.Error())
			return polly.SynthesizeSpeechInput{}
			// TODO throw error
		}
		options = newOptions
	}

	// adjust parameters for Google/AWS
	text, options = provider.TransformOptions(text, options)

	fmt.Println("Final Text: " + text)

	input2 := polly.SynthesizeSpeechInput{
		OutputFormat: aws.String("mp3"),
		Text:         aws.String(text),
		VoiceId:      aws.String(options.VoiceConfig.VoiceIdConfig.VoiceId),
		TextType:     aws.String(options.TextType.String())}

	if options.SampleRate != 0 {
		input2.SetSampleRate(fmt.Sprintf("%d", options.SampleRate))
	}

	// TODO execute T2S

	return input2
}

// GetProviderInstance returns an instance of a provider struct
func GetProviderInstance() T2SProvider {
	// TODO via options and heuristics
	return ts2_aws.T2SAmazonWebServices{}
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
