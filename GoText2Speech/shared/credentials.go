package shared

import "github.com/FaaSTools/GoStorage/gostorage"

/*
type CredentialsHolder struct {
	AwsCredentials *session.Options
	//GoogleCredentials *google.Credentials
	GoogleCredentials any // TODO when implementing GCP
}
*/

type CredentialsHolder = gostorage.CredentialsHolder
