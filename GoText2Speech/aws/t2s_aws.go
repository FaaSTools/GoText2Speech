package aws

import (
	. "goTest/GoText2Speech/shared"
)

type T2SAmazonWebServices struct {
}

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
