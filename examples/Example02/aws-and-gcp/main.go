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

	provider := "GCP"

	cfg, err0 := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err0 != nil {
		return "", err0
	}
	for i := 0; i < 20; i++ {
		if strings.EqualFold(provider, "AWS") {
			VoiceId := types.VoiceId("Joey")

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
		} else if strings.EqualFold(provider, "GCP") {
			VoiceId := "Joey"
			storageClient, err3 := storage.NewClient(context.Background())
			if err3 != nil {
				return "", err3
			}
			sourceObj := storageClient.Bucket(SourceBucket).Object(SourceKey)
			sourceReader, err6 := sourceObj.NewReader(context.Background())
			if err6 != nil {
				return "", err6
			}
			sourceBuf := new(bytes.Buffer)
			_, err7 := sourceBuf.ReadFrom(sourceReader)
			if err7 != nil {
				return "", err7
			}
			text := sourceBuf.String()
			t2sClient, err1 := texttospeech.NewClient(context.Background())
			if err1 != nil {
				return "", err1
			}
			input := &texttospeechpb.SynthesisInput{InputSource: &texttospeechpb.SynthesisInput_Text{Text: text}}
			if strings.HasPrefix(text, "<speak>") {
				input.InputSource = &texttospeechpb.SynthesisInput_Ssml{Ssml: text}
			}
			req := texttospeechpb.SynthesizeSpeechRequest{
				Input: input,
				Voice: &texttospeechpb.VoiceSelectionParams{
					LanguageCode: VoiceId[0:5],
					Name:         VoiceId,
				},
				AudioConfig: &texttospeechpb.AudioConfig{AudioEncoding: texttospeechpb.AudioEncoding_MP3},
			}
			result, err2 := t2sClient.SynthesizeSpeech(context.Background(), &req)
			if err2 != nil {
				return "", err2
			}
			audioStream := bytes.NewReader(result.GetAudioContent())
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
