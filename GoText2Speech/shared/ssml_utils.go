package shared

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// HasSpeakTag checks if the given text is surrounded by <speak>-tags, i.e. <speak ...>...</speak>.
// If the given text is not surrounded by <speak>-tags, it is not a valid SSML text.
func HasSpeakTag(text string) bool {
	trimmedText := strings.TrimSpace(text)
	return strings.HasPrefix(trimmedText, "<speak") && strings.HasSuffix(trimmedText, "</speak>")
}

// GetOpeningTagOfSSMLText Returns the opening tag (i.e. opening of root node) from SSML text
// Example 1: <speak>...</speak> -> <speak>
// Example 2: <speak volume="10db">...</speak> -> <speak volume="10db">
// TODO what if <speak> tag wasn't found at all?
func GetOpeningTagOfSSMLText(text string) string {
	return strings.SplitAfter(text, ">")[0]
}

func VolumeToSSMLAttribute(volume float64) string {
	return fmt.Sprintf("volume=\"%fdb\"", volume)
}

func PitchToSSMLAttribute(pitch float64) string {
	return fmt.Sprintf("pitch=\"%f%%\"", pitch*100.0)
}

func RateToSSMLAttribute(rate float64) string {
	return fmt.Sprintf("rate=\"%f%%\"", rate*100.0)
}

// TransformTextIntoSSML escapes special characters and wraps text into speak-tag with volume, pitch and rate parameters.
// Example: Hello World -> <speak volume="0db" rate="0%" pitch="0%">Hello World</speak>
func TransformTextIntoSSML(text string, options TextToSpeechOptions) string {
	text = EscapeTextForSSML(text)
	openingTag := fmt.Sprintf("<speak %s %s %s>",
		VolumeToSSMLAttribute(options.Volume),
		PitchToSSMLAttribute(options.Pitch),
		RateToSSMLAttribute(options.SpeakingRate))

	return openingTag + text + "</speak>"
}

// EscapeTextForSSML In order to use text in SSML, it needs to be escaped.
// These reserved characters are the same for GCP and AWS.
func EscapeTextForSSML(text string) string {
	replacements := [][]string{
		{"&", "&amp;"},
		{"\"", "&quot;"},
		{"'", "&apos;"},
		{"<", "&lt;"},
		{">", "&gt;"},
	}
	for _, replacement := range replacements {
		text = strings.Replace(text, replacement[0], replacement[1], -1)
	}

	return text
}

// IntegrateVolumeAttributeValueIntoTag integrates a value for the volume attribute into an existing <speak>-tag.
// Returns new opening speak tag.
// Following examples with volumeValue of 10:
// Example 1: <speak> -> <speak volume="10db">
// Example 2: <speak rate="10%"> -> <speak rate="10%" volume="10db">
// Example 3: <speak volume="5db"> -> <speak volume="15db">
// Example 4: <speak volume="loud"> -> <speak volume="10db">
func IntegrateVolumeAttributeValueIntoTag(openingTag string, volumeValue float64) string {
	return integrateAttributeValueIntoTag(openingTag, volumeValue, "volume", "db")
}

// IntegrateSpeakingRateAttributeValueIntoTag integrates a value for the speaking rate attribute into an existing <speak>-tag.
// Returns new opening speak tag.
// Following examples with speakingRateValue of 10:
// Example 1: <speak> -> <speak rate="10%">
// Example 2: <speak volume="5db"> -> <speak volume="5db" rate="10%">
// Example 3: <speak rate="5%"> -> <speak rate="15%">
// Example 4: <speak rate="fast"> -> <speak rate="10%">
func IntegrateSpeakingRateAttributeValueIntoTag(openingTag string, speakingRateValue float64) string {
	return integrateAttributeValueIntoTag(openingTag, speakingRateValue, "rate", "%")
}

// IntegratePitchAttributeValueIntoTag integrates a value for the pitch attribute into an existing <speak>-tag.
// Returns new opening speak tag.
// Following examples with pitchValue of 10:
// Example 1: <speak> -> <speak pitch="10%">
// Example 2: <speak volume="5db"> -> <speak volume="5db" pitch="10%">
// Example 3: <speak pitch="5%"> -> <speak pitch="15%">
// Example 4: <speak pitch="high"> -> <speak pitch="10%">
func IntegratePitchAttributeValueIntoTag(openingTag string, pitchValue float64) string {
	return integrateAttributeValueIntoTag(openingTag, pitchValue, "pitch", "%")
}

// integrateAttributeValueIntoTag integrates a value for a given attribute into an existing <speak>-tag.
func integrateAttributeValueIntoTag(openingTag string, value float64, attributeName string, attributeUnit string) string {
	regexWithUnit, _ := regexp.Compile(fmt.Sprintf("%s=\"(?P<%s>.*)%s\"", attributeName, attributeName, attributeUnit))
	submatchesWithUnit := regexWithUnit.FindStringSubmatch(openingTag)
	if submatchesWithUnit != nil {
		predefinedValueStr := submatchesWithUnit[1]
		predefinedValue, _ := strconv.ParseFloat(predefinedValueStr, 64) // TODO catch err?
		value += predefinedValue
		return regexWithUnit.ReplaceAllString(openingTag, fmt.Sprintf("%s=\"%f%s\"", attributeName, value, attributeUnit))
	} else { // check if attribute exists without unit -> overwrite
		regexWithoutUnit, _ := regexp.Compile(fmt.Sprintf("%s=\"(?P<%s>.*)\"", attributeName, attributeName))
		submatchesWithoutUnit := regexWithoutUnit.FindStringSubmatch(openingTag)
		if submatchesWithoutUnit != nil {
			return regexWithoutUnit.ReplaceAllString(openingTag, fmt.Sprintf("%s=\"%f%s\"", attributeName, value, attributeUnit))
		} else {
			// attribute doesn't exist -> add it
			openingTag = openingTag[:len(openingTag)-1] // remove last >
			openingTag += fmt.Sprintf(" %s=\"%f%s\">", attributeName, value, attributeUnit)
			return openingTag
		}
	}
}
