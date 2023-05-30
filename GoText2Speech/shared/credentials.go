package shared

import (
	"github.com/aws/aws-sdk-go/aws/session"
)

type CredentialsHolder struct {
	AwsCredentials *session.Options
	//GoogleCredentials *google.Credentials
	GoogleCredentials any // TODO when implementing GCP
}
