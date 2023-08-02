package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/polly"
	"github.com/aws/aws-sdk-go-v2/service/polly/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"strings"
)

type MyEvent struct {
	VoiceId      string `json:"voiceId"`
	SourceBucket string `json:"sourceBucket"`
	SourceKey    string `json:"sourceKey"`
	TargetBucket string `json:"targetBucket"`
	TargetKey    string `json:"targetKey"`
}

func HandleRequest(ctx context.Context, ev MyEvent) (string, error) {
	SourceKey := "example02/example02.txt"
	SourceBucket := "test"
	TargetKey := "example02/example02-aws.mp3"
	TargetBucket := "test"
	VoiceId := types.VoiceId("Joey")
	fmt.Println("VoiceId:", VoiceId)

	cfg, err0 := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err0 != nil {
		return "", err0
	}
	for i := 0; i < 20; i++ {
		s3Client := s3.NewFromConfig(cfg)
		getObjectInput := &s3.GetObjectInput{
			Bucket: &SourceBucket,
			Key:    &SourceKey,
		}
		fmt.Println("getObjectInput:", *getObjectInput.Bucket, *getObjectInput.Key)
		getObjectOutput, err1 := s3Client.GetObject(context.Background(), getObjectInput)
		if err1 != nil {
			return "", err1
		}
		inputFileBuf := new(bytes.Buffer)
		_, err2 := inputFileBuf.ReadFrom(getObjectOutput.Body)
		if err2 != nil {
			return "", err2
		}
		text := inputFileBuf.String()
		textType := types.TextTypeText
		if strings.HasPrefix(text, "<speak>") {
			textType = types.TextTypeSsml
		}
		client := polly.NewFromConfig(cfg)
		speechInput := &polly.SynthesizeSpeechInput{
			OutputFormat: types.OutputFormatMp3,
			Text:         aws.String(text),
			VoiceId:      VoiceId,
			Engine:       types.EngineStandard,
			TextType:     textType,
		}
		err3 := getObjectOutput.Body.Close()
		if err3 != nil {
			return "", err3
		}
		audioOut, err4 := client.SynthesizeSpeech(context.Background(), speechInput)
		if err4 != nil {
			return "", err4
		}
		audioBuf := new(bytes.Buffer)
		_, err5 := audioBuf.ReadFrom(audioOut.AudioStream)
		if err5 != nil {
			return "", err5
		}
		_, err6 := s3Client.PutObject(context.Background(), &s3.PutObjectInput{
			Bucket:        &TargetBucket,
			Key:           &TargetKey,
			Body:          audioBuf,
			ContentLength: int64(audioBuf.Len()),
		})
		if err6 != nil {
			return "", err6
		}
	}
	return fmt.Sprintf("Upload to %s/%s done!", TargetBucket, TargetKey), nil
}

func main() {
	lambda.Start(HandleRequest)

}
