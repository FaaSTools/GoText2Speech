package gcp

import (
	. "goTest/GoText2Speech/shared"
)

type T2SGoogleCloudPlatform struct {
}

func (a T2SGoogleCloudPlatform) TransformOptions(text string, options TextToSpeechOptions) (string, TextToSpeechOptions) {
	// TODO implement
	return text, options
}

func (a T2SGoogleCloudPlatform) ChooseVoice(options TextToSpeechOptions) (TextToSpeechOptions, error) {
	// TODO implement
	return options, nil
}
