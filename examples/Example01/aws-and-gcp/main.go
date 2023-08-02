package main

import (
	"bytes"
	"cloud.google.com/go/storage"
	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/polly"
	"github.com/aws/aws-sdk-go-v2/service/polly/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"strconv"
	"strings"
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

	provider := "GCP"

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		return "", err
	}
	fmt.Println("VoiceId:", ev.VoiceId)
	fmt.Println("VoiceId conv:", types.VoiceId(ev.VoiceId))
	for i := 0; i < 20; i++ {
		if strings.EqualFold(provider, "AWS") {
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
		} else if strings.EqualFold(provider, "GCP") {
			t2sClient, err1 := texttospeech.NewClient(context.Background())
			fmt.Println("Text:", ev.Text)
			if err1 != nil {
				fmt.Println(err1)
				return "", err1
			}
			input := &texttospeechpb.SynthesisInput{InputSource: &texttospeechpb.SynthesisInput_Text{Text: ev.Text}}
			req := texttospeechpb.SynthesizeSpeechRequest{
				Input: input,
				Voice: &texttospeechpb.VoiceSelectionParams{
					LanguageCode: ev.VoiceId[0:5],
					Name:         ev.VoiceId,
				},
				AudioConfig: &texttospeechpb.AudioConfig{AudioEncoding: texttospeechpb.AudioEncoding_MP3},
			}
			result, err2 := t2sClient.SynthesizeSpeech(context.Background(), &req)
			if err2 != nil {
				fmt.Println(err2)
				return "", err2
			}
			audioStream := bytes.NewReader(result.GetAudioContent())
			storageClient, err3 := storage.NewClient(context.Background())
			if err3 != nil {
				fmt.Println(err3)
				return "", err3
			}
			cloudObj := storageClient.Bucket(ev.TargetBucket).Object(ev.TargetKey + strconv.Itoa(i))
			wc := cloudObj.NewWriter(context.Background())
			if _, err4 := io.Copy(wc, audioStream); err4 != nil {
				return "", fmt.Errorf("io.Copy: %w", err4)
			}
			if err5 := wc.Close(); err5 != nil {
				return "", fmt.Errorf("Writer.Close: %w", err5)
			}
			defer storageClient.Close()
			defer t2sClient.Close()
		}
	}
	return fmt.Sprintf("Upload to %s/%s done!", ev.TargetBucket, ev.TargetKey), nil
}

func main() {
	lambda.Start(HandleRequest)

}
