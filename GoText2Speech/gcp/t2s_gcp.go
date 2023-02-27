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
