package shared

type T2SProvider interface {
	TransformOptions(text string, options TextToSpeechOptions) (string, TextToSpeechOptions)
}
