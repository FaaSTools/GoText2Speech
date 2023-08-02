package main

import (
	"context"
	"fmt"
	goT2S "github.com/FaaSTools/GoText2Speech/GoText2Speech"
	goT2SProvider "github.com/FaaSTools/GoText2Speech/GoText2Speech/providers"
	goT2SShared "github.com/FaaSTools/GoText2Speech/GoText2Speech/shared"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
)

type MyEvent struct {
	Text         string `json:"text"`
	VoiceId      string `json:"voiceId"`
	TargetBucket string `json:"targetBucket"`
	TargetKey    string `json:"targetKey"`
}

func HandleRequest(ctx context.Context, ev MyEvent) (string, error) {
	ev = MyEvent{
		Text:         "Hello World",
		VoiceId:      "Joey",
		TargetBucket: "test",
		TargetKey:    "example01/example01-got2s.mp3",
	}

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
		options.VoiceConfig.VoiceIdConfig = goT2SShared.VoiceIdConfig{
			VoiceId: ev.VoiceId,
			Engine:  "standard",
		}
		options.Provider = goT2SProvider.ProviderAWS
		var err2 error = nil
		t2sClient, err2 = t2sClient.T2SDirect(ev.Text, "s3://"+ev.TargetBucket+"/"+ev.TargetKey, *options)
		if err2 != nil {
			return "", err2
		}
	}
	return fmt.Sprintf("Upload to %s/%s done!", ev.TargetBucket, ev.TargetKey), nil
}

func main() {
	lambda.Start(HandleRequest)

}
