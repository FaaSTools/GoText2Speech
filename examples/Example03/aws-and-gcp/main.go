package main

import (
	"bytes"
	"cloud.google.com/go/storage"
	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"context"
	"errors"
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
	LanguageCode string `json:"language"`
	Gender       string `json:"gender"`
	Text         string `json:"text"`
	TargetBucket string `json:"targetBucket"`
	TargetKey    string `json:"targetKey"`
}

func HandleRequest(ctx context.Context, ev MyEvent) (string, error) {
	TargetKey := "example03/example03-aws.mp3"
	TargetBucket := "test"
	LanguageCode := "en-US"
	Text := "Hello World"

	provider := "GCP"

	cfg, err0 := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err0 != nil {
		return "", err0
	}
	for i := 0; i < 20; i++ {
		if strings.EqualFold(provider, "AWS") {
			Gender := "Male"
			client := polly.NewFromConfig(cfg)
			describeVoicesInput := &polly.DescribeVoicesInput{
				Engine:       types.EngineStandard,
				LanguageCode: types.LanguageCode(LanguageCode),
			}
			fmt.Println("Language code:", LanguageCode)
			describeVoicesOutput, err1 := client.DescribeVoices(context.Background(), describeVoicesInput)
			if err1 != nil {
				return "", err1
			}
			var voiceId *types.VoiceId = nil
			for _, voice := range describeVoicesOutput.Voices {
				fmt.Println("Compare voice ", voice.Gender, " with ", Gender)
				if strings.EqualFold(string(voice.Gender), Gender) {
					voiceId = &voice.Id
					break
				}
			}
			if voiceId == nil {
				return "", errors.New("no suitable voice found")
			}
			speechInput := &polly.SynthesizeSpeechInput{
				OutputFormat: types.OutputFormatMp3,
				Text:         aws.String(Text),
				VoiceId:      *voiceId,
				Engine:       types.EngineStandard,
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
			s3Client := s3.NewFromConfig(cfg)
			_, err6 := s3Client.PutObject(context.Background(), &s3.PutObjectInput{
				Bucket:        &TargetBucket,
				Key:           &TargetKey,
				Body:          audioBuf,
				ContentLength: int64(audioBuf.Len()),
			})
			if err6 != nil {
				return "", err6
			}
		} else if strings.EqualFold(provider, "GCP") {
			Gender := texttospeechpb.SsmlVoiceGender_MALE
			t2sClient, err0 := texttospeech.NewClient(context.Background())
			if err0 != nil {
				return "", err0
			}
			input := &texttospeechpb.SynthesisInput{InputSource: &texttospeechpb.SynthesisInput_Text{Text: Text}}
			req := texttospeechpb.SynthesizeSpeechRequest{
				Input: input,
				Voice: &texttospeechpb.VoiceSelectionParams{
					LanguageCode: LanguageCode,
					SsmlGender:   Gender,
				},
				AudioConfig: &texttospeechpb.AudioConfig{AudioEncoding: texttospeechpb.AudioEncoding_MP3},
			}
			result, err1 := t2sClient.SynthesizeSpeech(context.Background(), &req)
			if err1 != nil {
				return "", err1
			}
			audioStream := bytes.NewReader(result.GetAudioContent())
			storageClient, err2 := storage.NewClient(context.Background())
			if err2 != nil {
				return "", err2
			}
			cloudObj := storageClient.Bucket(TargetBucket).Object(TargetKey + strconv.Itoa(i))
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
	return fmt.Sprintf("Upload to %s/%s done!", TargetBucket, TargetKey), nil
}

func main() {
	lambda.Start(HandleRequest)
}
