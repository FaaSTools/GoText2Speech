package aws

import (
	. "goTest/GoText2Speech/shared"
)

type T2SAmazonWebServices struct {
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
