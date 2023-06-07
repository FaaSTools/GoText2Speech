package shared

type T2SProvider interface {
	// TransformOptions Transforms the given options object such that it can be used for the chosen provider.
	TransformOptions(text string, options TextToSpeechOptions) (string, TextToSpeechOptions, error)
	// FindVoice finds a voice that is available on the provider based on the given parameters (language, gender and optionally engine).
	FindVoice(options TextToSpeechOptions) (*VoiceIdConfig, error)
	// CreateServiceClient creates t2s client for the chosen provider and stores it in the struct.
	CreateServiceClient(credentials CredentialsHolder, region string) (T2SProvider, error)
	ExecuteT2SDirect(text string, destination string, options TextToSpeechOptions) error
	ExecuteT2S(source string, destination string, options TextToSpeechOptions) error
	// IsURLonOwnStorage checks if the given URL references a file that is hosted on the provider's own storage service
	// (i.e. S3 on AWS or Cloud Storage on GCP).
	IsURLonOwnStorage(url string) bool
	// CreateTempDestination creates a URL for the provider's own storage service (i.e. S3 on AWS or Cloud Storage on GCP)
	// on which a temporary file can be stored.
	CreateTempDestination(tempBucket string, fileName string) string
	GetSupportedAudioFormats() []AudioFormat
	CloseServiceClient() error
}
