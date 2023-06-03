package shared

import "goTest/GoText2Speech"

type T2SProvider interface {
	// TransformOptions Transforms the given options object such that it can be used for the chosen provider.
	TransformOptions(text string, options TextToSpeechOptions) (string, TextToSpeechOptions, error)
	// ChooseVoice chooses a voice that is available on the provider based on the given parameters (language and gender).
	ChooseVoice(options TextToSpeechOptions) (TextToSpeechOptions, error)
	// CreateServiceClient creates t2s client for the chosen provider and stores it in the struct.
	CreateServiceClient(credentials CredentialsHolder, region string) T2SProvider
	ExecuteT2SDirect(text string, destination string, options TextToSpeechOptions) error
	ExecuteT2S(source string, destination string, options TextToSpeechOptions) error
	// IsURLonOwnStorage checks if the given URL references a file that is hosted on the provider's own storage service
	// (i.e. S3 on AWS or Cloud Storage on GCP).
	IsURLonOwnStorage(url string) bool
	// CreateTempDestination creates a URL for the provider's own storage service (i.e. S3 on AWS or Cloud Storage on GCP)
	// on which a temporary file can be stored.
	CreateTempDestination(goT2SClient GoText2Speech.GoT2SClient, fileName string) string
}
