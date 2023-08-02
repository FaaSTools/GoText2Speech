package helloworld

import (
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"context"
	"fmt"
	goT2S "github.com/FaaSTools/GoText2Speech/GoText2Speech"
	"github.com/FaaSTools/GoText2Speech/GoText2Speech/providers"
	goT2SShared "github.com/FaaSTools/GoText2Speech/GoText2Speech/shared"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"golang.org/x/oauth2/google"
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

func execT2S(r *http.Request) error {
	var MyEvent struct { // don't count this struct
		Text         string                         `json:"Text"`
		LanguageCode string                         `json:"language"`
		Gender       texttospeechpb.SsmlVoiceGender `json:"gender"`
		TargetBucket string                         `json:"TargetBucket"`
		TargetKey    string                         `json:"TargetKey"`
	}
	MyEvent.TargetKey = "example03/example03-got2s.mp3"
	MyEvent.TargetBucket = "test"
	MyEvent.LanguageCode = "en-US"
	MyEvent.Gender = texttospeechpb.SsmlVoiceGender_MALE
	MyEvent.Text = "Hello World"

	fmt.Println("Text:", MyEvent.Text)
	region := "us-east-1"
	googleCredentials, err := google.CredentialsFromJSON(
		context.Background(),
		[]byte(""),
		"https://www.googleapis.com/auth/devstorage.full_control",
		"https://www.googleapis.com/auth/cloud-platform",
	)
	fmt.Println("err while reading credentials:", err)

	cred := &goT2SShared.CredentialsHolder{
		GoogleCredentials: googleCredentials,
	}
	for i := 0; i < 20; i++ {
		t2sClient := goT2S.CreateGoT2SClient(cred, region)
		options := goT2SShared.GetDefaultTextToSpeechOptions()
		options.VoiceConfig.VoiceParamsConfig.Gender = goT2SShared.VoiceGender(MyEvent.Gender)
		options.VoiceConfig.VoiceParamsConfig.LanguageCode = MyEvent.LanguageCode
		options.Provider = providers.ProviderGCP
		options.AddFileExtension = false
		var err2 error = nil
		t2sClient, err2 = t2sClient.T2SDirect(MyEvent.Text, "gs://"+MyEvent.TargetBucket+"/"+MyEvent.TargetKey+strconv.Itoa(i), *options)
		if err2 != nil {
			return err2
		}
	}
	return nil
}

func init() {
	// Register an HTTP function with the Functions Framework
	functions.HTTP("MyHTTPFunction", MyHTTPFunction)
}

func main() {} // needs to be here, otherwise it can't be built
