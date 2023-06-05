package providers

type Provider string

const (
	ProviderAWS         Provider = "AWS"
	ProviderGCP         Provider = "GCP"
	ProviderUnspecified Provider = ""
)

var allProviders = []Provider{ProviderAWS, ProviderGCP}

func GetAllProviders() []Provider {
	return allProviders
}
