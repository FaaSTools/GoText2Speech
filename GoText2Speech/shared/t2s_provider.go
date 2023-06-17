package shared

import "io"

type T2SProvider interface {
	// TransformOptions Transforms the given options object such that it can be used for the chosen provider.
	TransformOptions(text string, options TextToSpeechOptions) (string, TextToSpeechOptions, error)
	// FindVoice finds a voice that is available on the provider based on the given parameters (language, gender and optionally engine).
	FindVoice(options TextToSpeechOptions) (*VoiceIdConfig, error)
	// CreateServiceClient creates t2s client for the chosen provider and stores it in the struct.
	CreateServiceClient(credentials CredentialsHolder, region string) (T2SProvider, error)
	ExecuteT2SDirect(text string, destination string, options TextToSpeechOptions) (io.Reader, error)
	UploadFile(file io.Reader, destination string) error
	// IsURLonOwnStorage checks if the given URL references a file that is hosted on the provider's own storage service
	// (i.e. S3 on AWS or Cloud Storage on GCP).
	IsURLonOwnStorage(url string) bool
	// CreateTempDestination creates a URL for the provider's own storage service (i.e. S3 on AWS or Cloud Storage on GCP)
	// on which a temporary file can be stored.
	CreateTempDestination(tempBucket string, fileName string) string
	// GetSupportedAudioFormats returns an array of all audio formats that are supported as output format by the t2s service of this provider.
	GetSupportedAudioFormats() []AudioFormat
	// CloseServiceClient closes the connection of the t2s client in the struct (if such an operation is available on the provider).
	CloseServiceClient() error
	AddFileExtensionToDestinationIfNeeded(options TextToSpeechOptions, outputFormatRaw any, destination string) (string, error)
}
