package helloworld

import (
	"bytes"
	"cloud.google.com/go/storage"
	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"context"
	"fmt"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"io"
	"net/http"
	"strconv"
)

// MyHTTPFunction Function is an HTTP handler
func MyHTTPFunction(w http.ResponseWriter, r *http.Request) {
	err := execT2S(r)
	if err != nil {
		fmt.Println("Error:", err.Error())
		_, err1 := io.WriteString(w, "Error: "+err.Error())
		if err1 != nil {
			fmt.Println("Error while writing error to output: ", err1)
		}
	} else {
		_, err1 := io.WriteString(w, "Done successfully!")
		if err1 != nil {
			fmt.Println("Error while writing success message to output: ", err1)
		}
	}
}

// Count lines of code in this function
func execT2S(r *http.Request) error {
	var MyEvent struct { // don't count this struct
		Text         string                         `json:"Text"`
		LanguageCode string                         `json:"language"`
		Gender       texttospeechpb.SsmlVoiceGender `json:"gender"`
		TargetBucket string                         `json:"TargetBucket"`
		TargetKey    string                         `json:"TargetKey"`
	}
	MyEvent.TargetKey = "example03/example03-gcp.mp3"
	MyEvent.TargetBucket = "test"
	MyEvent.LanguageCode = "en-US"
	MyEvent.Gender = texttospeechpb.SsmlVoiceGender_MALE
	MyEvent.Text = "Hello World"

	//region := "us-east-1"
	for i := 0; i < 20; i++ {
		t2sClient, err0 := texttospeech.NewClient(context.Background())
		if err0 != nil {
			return err0
		}
		input := &texttospeechpb.SynthesisInput{InputSource: &texttospeechpb.SynthesisInput_Text{Text: MyEvent.Text}}
		req := texttospeechpb.SynthesizeSpeechRequest{
			Input: input,
			Voice: &texttospeechpb.VoiceSelectionParams{
				LanguageCode: MyEvent.LanguageCode,
				SsmlGender:   MyEvent.Gender,
			},
			AudioConfig: &texttospeechpb.AudioConfig{AudioEncoding: texttospeechpb.AudioEncoding_MP3},
		}
		result, err1 := t2sClient.SynthesizeSpeech(context.Background(), &req)
		if err1 != nil {
			return err1
		}
		audioStream := bytes.NewReader(result.GetAudioContent())
		storageClient, err2 := storage.NewClient(context.Background())
		if err2 != nil {
			return err2
		}
		cloudObj := storageClient.Bucket(MyEvent.TargetBucket).Object(MyEvent.TargetKey + strconv.Itoa(i))
		wc := cloudObj.NewWriter(context.Background())
		if _, err4 := io.Copy(wc, audioStream); err4 != nil {
			return fmt.Errorf("io.Copy: %w", err4)
		}
		if err5 := wc.Close(); err5 != nil {
			return fmt.Errorf("Writer.Close: %w", err5)
		}
		defer storageClient.Close()
		defer t2sClient.Close()
	}
	return nil
}

func init() {
	// Register an HTTP function with the Functions Framework
	functions.HTTP("MyHTTPFunction", MyHTTPFunction)
}

func main() {} // needs to be here, otherwise it can't be built
