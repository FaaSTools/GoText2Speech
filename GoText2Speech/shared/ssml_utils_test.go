package shared

import (
	"fmt"
	"testing"
)

func TestEscapeTextForSSMLUnchanged(t *testing.T) {
	t1 := "Hello World!"
	if t1 != EscapeTextForSSML(t1) {
		t.Error("Text shouldn't change when it doesn't include an escapable character")
	}
}

func TestEscapeTextForSSMLEscapes(t *testing.T) {
	input := "He said: \"10 < 9 & 11 > 12\". That's not true!"
	want := "He said: &quot;10 &lt; 9 &amp; 11 &gt; 12&quot;. That&apos;s not true!"
	result := EscapeTextForSSML(input)
	if want != result {
		t.Errorf("Escape characters weren't properly escaped.\nWanted:\t%s\nGot:\t%s", want, result)
	}
}

func TestGetOpeningTagOfSSMLTextNormal(t *testing.T) {
	input := "<speak>Hello World</speak>"
	want := "<speak>"
	result := GetOpeningTagOfSSMLText(input)
	if want != result {
		t.Errorf("Opening tag was not correctly retrieved.\nWatend:\t%s\nGot:\t%s", want, result)
	}
}

func TestGetOpeningTagOfSSMLTextWithAttributes(t *testing.T) {
	input := "<speak attr1=\"test\" attr2=\"123\">Hello World</speak>"
	want := "<speak attr1=\"test\" attr2=\"123\">"
	result := GetOpeningTagOfSSMLText(input)
	if want != result {
		t.Errorf("Opening tag was not correctly retrieved.\nWatend:\t%s\nGot:\t%s", want, result)
	}
}

// TestTransformTextIntoSSMLNormal test if text is escaped properly, wrapped in <speak>-tags and
// if attributes are added to opening speak tag
func TestTransformTextIntoSSMLNormal(t *testing.T) {
	options := TextToSpeechOptions{Volume: 10.0, SpeakingRate: 2.0, Pitch: 0.05}
	input := "Hello World! Lovely day, isn't it?"
	want := "<speak volume=\"10.000000db\" pitch=\"5.000000%\" rate=\"200.000000%\">Hello World! Lovely day, isn&apos;t it?</speak>"
	result := TransformTextIntoSSML(input, options)
	if want != result {
		t.Errorf("Text was not properly transformed into SSML.\nWanted:\t%s\nGot:\t%s", want, result)
	}
}

type testcase struct {
	name  string
	input string
	want  string
}

func TestIntegrateVolumeAttributeValueIntoTag(t *testing.T) {
	volumeValueDb := 10.0
	var inputs = []testcase{
		{"Normal", "<speak>", fmt.Sprintf("<speak volume=\"%fdb\">", volumeValueDb)},
		{"Additional Attribute", "<speak rate=\"10%\">", fmt.Sprintf("<speak rate=\"10%%\" volume=\"%fdb\">", volumeValueDb)},
		{"Integrate and add", "<speak volume=\"5db\">", fmt.Sprintf("<speak volume=\"%fdb\">", volumeValueDb+5)},
		{"Integrate and overwrite", "<speak volume=\"loud\">", fmt.Sprintf("<speak volume=\"%fdb\">", volumeValueDb)},
	}

	for _, testcase := range inputs {
		t.Run(testcase.name, func(t *testing.T) {
			result := IntegrateVolumeAttributeValueIntoTag(testcase.input, volumeValueDb)
			if testcase.want != result {
				t.Errorf("Test failed.\nWanted:\t%s\nGot:\t%s", testcase.want, result)
			}
		})
	}
}

func TestIntegratePitchAttributeValueIntoTag(t *testing.T) {
	pitchValue := 10.0
	var inputs = []testcase{
		{"Normal", "<speak>", fmt.Sprintf("<speak pitch=\"%f%%\">", pitchValue)},
		{"Additional Attribute", "<speak volume=\"5db\">", fmt.Sprintf("<speak volume=\"5db\" pitch=\"%f%%\">", pitchValue)},
		{"Integrate and add", "<speak pitch=\"5%\">", fmt.Sprintf("<speak pitch=\"%f%%\">", pitchValue+5)},
		{"Integrate and overwrite", "<speak pitch=\"high\">", fmt.Sprintf("<speak pitch=\"%f%%\">", pitchValue)},
	}

	for _, testcase := range inputs {
		t.Run(testcase.name, func(t *testing.T) {
			result := IntegratePitchAttributeValueIntoTag(testcase.input, pitchValue)
			if testcase.want != result {
				t.Errorf("Test failed.\nWanted:\t%s\nGot:\t%s", testcase.want, result)
			}
		})
	}
}

func TestIntegrateSpeakingRateAttributeValueIntoTag(t *testing.T) {
	speakingRateValue := 10.0
	var inputs = []testcase{
		{"Normal", "<speak>", fmt.Sprintf("<speak rate=\"%f%%\">", speakingRateValue)},
		{"Additional Attribute", "<speak volume=\"5db\">", fmt.Sprintf("<speak volume=\"5db\" rate=\"%f%%\">", speakingRateValue)},
		{"Integrate and add", "<speak rate=\"5%\">", fmt.Sprintf("<speak rate=\"%f%%\">", speakingRateValue+5)},
		{"Integrate and overwrite", "<speak rate=\"fast\">", fmt.Sprintf("<speak rate=\"%f%%\">", speakingRateValue)},
	}

	for _, testcase := range inputs {
		t.Run(testcase.name, func(t *testing.T) {
			result := IntegrateSpeakingRateAttributeValueIntoTag(testcase.input, speakingRateValue)
			if testcase.want != result {
				t.Errorf("Test failed.\nWanted:\t%s\nGot:\t%s", testcase.want, result)
			}
		})
	}
}