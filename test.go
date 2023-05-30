package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"os"
)

import (
	. "goTest/GoText2Speech"
	. "goTest/GoText2Speech/shared"
)

// main shows how T2S might be executed.
func main() {

	fmt.Println("Synthesizing speech...")
	s := "<speak><prosody>Hello World, how are you today? Lovely day, isn't it?</prosody></speak>"
	destination := "s3://test/testfile.mp3"

	s2 := "<speak><prosody>Hello World</prosody></speak>"
	destination2 := "s3://test/testfile2.mp3"

	options := GetDefaultTextToSpeechOptions()
	options.VoiceConfig.VoiceIdConfig = VoiceIdConfig{VoiceId: "Salli", Engine: "neural"}
	// Alternative voice config
	//options.VoiceConfig.VoiceParamsConfig = VoiceParamsConfig{LanguageCode: "en-US", Gender: VoiceGenderFemale}
	//options.SpeakingRate = 1.1

	t2sClient := CreateGoT2SClient(CredentialsHolder{
		AwsCredentials: &session.Options{
			SharedConfigState: session.SharedConfigEnable,
			Profile:           "test",
		},
	}, "us-east-1")
	var err error = nil
	t2sClient, err = t2sClient.T2SDirect(s, destination, *options)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	t2sClient, err = t2sClient.T2SDirect(s2, destination2, *options)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Speech successfully synthesized!")
}
