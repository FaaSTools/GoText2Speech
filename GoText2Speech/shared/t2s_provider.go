package shared

type T2SProvider interface {
	// TransformOptions Transforms the given options object such that it can be used for the chosen provider.
	TransformOptions(text string, options TextToSpeechOptions) (string, TextToSpeechOptions)
	// ChooseVoice chooses a voice that is available on the provider based on the given parameters (language and gender).
	ChooseVoice(options TextToSpeechOptions) (TextToSpeechOptions, error)
	// CreateClient creates t2s client for the chosen provider and stores it in the struct.
	CreateClient() // TODO parameters?
	// TODO add function for executing t2s
}
