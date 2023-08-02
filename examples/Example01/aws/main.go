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
)

type MyEvent struct {
	Text         string `json:"text"`
	VoiceId      string `json:"voiceId"`
	TargetBucket string `json:"targetBucket"`
	TargetKey    string `json:"targetKey"`
}

func HandleRequest(ctx context.Context, ev0 MyEvent) (string, error) {
	ev := MyEvent{
		Text:         "Hello World",
		VoiceId:      "Joey",
		TargetBucket: "test",
		TargetKey:    "example01/example01-aws.mp3",
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		return "", err
	}
	fmt.Println("VoiceId:", ev.VoiceId)
	fmt.Println("VoiceId conv:", types.VoiceId(ev.VoiceId))
	for i := 0; i < 20; i++ {
		speechInput := &polly.SynthesizeSpeechInput{
			OutputFormat: types.OutputFormatMp3,
			Text:         aws.String(ev.Text),
			VoiceId:      types.VoiceId(ev.VoiceId),
			Engine:       types.EngineStandard,
		}
		client := polly.NewFromConfig(cfg)
		out, err1 := client.SynthesizeSpeech(context.Background(), speechInput)
		if err1 != nil {
			return "", err1
		}
		buf := new(bytes.Buffer)
		_, err2 := buf.ReadFrom(out.AudioStream)
		if err2 != nil {
			return "", err2
		}
		uploader := s3.NewFromConfig(cfg)
		_, err3 := uploader.PutObject(context.Background(), &s3.PutObjectInput{
			Bucket:        &ev.TargetBucket,
			Key:           &ev.TargetKey,
			Body:          buf,
			ContentLength: int64(buf.Len()),
		})
		if err3 != nil {
			return "", err3
		}
	}
	return fmt.Sprintf("Upload to %s/%s done!", ev.TargetBucket, ev.TargetKey), nil
}

func main() {
	lambda.Start(HandleRequest)

}
