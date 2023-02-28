package providers

type Provider string

const (
	ProviderAWS         Provider = "AWS"
	ProviderGCP         Provider = "GCP"
	ProviderUnspecified Provider = ""
)
