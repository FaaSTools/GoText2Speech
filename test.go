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
	fmt.Println("Starting speech synthesis...")

	t2sClient := CreateGoT2SClient(CredentialsHolder{
		AwsCredentials: &session.Options{
			SharedConfigState: session.SharedConfigEnable,
			Profile:           "davidmeyer",
		},
	}, "us-east-1")

	options := GetDefaultTextToSpeechOptions()
	options.VoiceConfig.VoiceIdConfig = VoiceIdConfig{VoiceId: "Salli", Engine: "neural"}

	var err error = nil
	t2sClient, err = t2sClient.T2SDirect("<speak><prosody>Hello World, how are you today? Lovely day, isn't it?</prosody></speak>", "s3://davemeyer-test/testfile.mp3", *options)
	t2sClient, err = t2sClient.T2SDirect("Test", "s3://davemeyer-test/testfile_02.mp3", *options)
	t2sClient, err = t2sClient.T2S("https://www.davemeyer.io/GoSpeechLess/T2S_Test_file_01.txt", "s3://davemeyer-test/testfile_03.mp3", *options)

	t2sClient.AwsTempBucket = "davemeyer-test"
	t2sClient, err = t2sClient.T2S("https://davemeyer-test.s3.amazonaws.com/T2S_Test_file_01.txt", "D:\\testfile_04.mp3", *options)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Speech successfully synthesized!")
}
