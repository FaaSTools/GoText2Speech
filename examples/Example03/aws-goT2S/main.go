package main

import (
	"context"
	"fmt"
	goT2S "github.com/FaaSTools/GoText2Speech/GoText2Speech"
	"github.com/FaaSTools/GoText2Speech/GoText2Speech/providers"
	goT2SShared "github.com/FaaSTools/GoText2Speech/GoText2Speech/shared"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
)

type MyEvent struct {
	LanguageCode string `json:"language"`
	Gender       int    `json:"gender"`
	Text         string `json:"text"`
	TargetBucket string `json:"targetBucket"`
	TargetKey    string `json:"targetKey"`
}

func HandleRequest(ctx context.Context, ev MyEvent) (string, error) {
	TargetKey := "example03/example03-got2s.mp3"
	TargetBucket := "test"
	LanguageCode := "en-US"
	Gender := 1
	Text := "Hello World"

	region := "us-east-1"
	cfg, err0 := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err0 != nil {
		return "", err0
	}
	awsCred, err1 := cfg.Credentials.Retrieve(context.Background())
	if err1 != nil {
		return "", err1
	}
	cred := &goT2SShared.CredentialsHolder{AwsCredentials: &awsCred}
	for i := 0; i < 20; i++ {
		t2sClient := goT2S.CreateGoT2SClient(cred, region)
		options := goT2SShared.GetDefaultTextToSpeechOptions()
		options.VoiceConfig.VoiceParamsConfig.Gender = goT2SShared.VoiceGender(Gender)
		options.VoiceConfig.VoiceParamsConfig.LanguageCode = LanguageCode
		options.VoiceConfig.VoiceParamsConfig.Engine = "standard"
		options.Provider = providers.ProviderAWS
		var err2 error = nil
		t2sClient, err2 = t2sClient.T2SDirect(Text, "s3://"+TargetBucket+"/"+TargetKey, *options)
		if err2 != nil {
			return "", err2
		}
	}
	return fmt.Sprintf("Upload to %s/%s done!", TargetBucket, TargetKey), nil
}

func main() {
	lambda.Start(HandleRequest)
}
