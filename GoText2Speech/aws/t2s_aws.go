package aws

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/polly"
	. "goTest/GoText2Speech/shared"
	"strings"
)

type T2SAmazonWebServices struct {
	t2sClient *polly.Polly
}

// TransformOptions
// This function assumes that the basic options check was already executed,
// i.e. that options.TextType cannot be TextTypeAuto and that if the text is SSML text, it contains correctly formed <speak>-tags.
func (a T2SAmazonWebServices) TransformOptions(text string, options TextToSpeechOptions) (string, TextToSpeechOptions) {
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
	return text, options
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
		errText := fmt.Sprintf("Error: No voice found for language %s and gender %s\n",
			options.VoiceConfig.VoiceParamsConfig.LanguageCode,
			options.VoiceConfig.VoiceParamsConfig.Gender.String())
		return options, errors.New(errText)
	}

	return options, nil
}
