package main

import (
	"fmt"
	. "github.com/FaaSTools/GoText2Speech/GoText2Speech"
	providers "github.com/FaaSTools/GoText2Speech/GoText2Speech/providers"
	. "github.com/FaaSTools/GoText2Speech/GoText2Speech/shared"
	"os"
)

// main shows how T2S might be executed.
func main() {
	fmt.Println("Starting speech synthesis...")

	t2sClient := CreateGoT2SClient(nil, "us-east-1")

	options := GetDefaultTextToSpeechOptions()
	options.Provider = providers.ProviderAWS
	//options.VoiceConfig.VoiceIdConfig = VoiceIdConfig{VoiceId: "en-US-News-N"}
	options.VoiceConfig.VoiceParamsConfig = VoiceParamsConfig{
		LanguageCode: "en-US",
		Gender:       VoiceGenderFemale,
		Engine:       "neural",
	}

	bucket := "YOUR_BUCKET_HERE"

	var err error = nil
	t2sClient, err = t2sClient.T2SDirect("<speak><prosody volume=\"10.000dB\">Hello World, how are you today? Lovely day, isn't it?</prosody></speak>", "s3://"+bucket+"/testfile.mp3", *options)
	t2sClient, err = t2sClient.T2SDirect("Test", "s3://"+bucket+"/testfile_02.mp3", *options)
	t2sClient, err = t2sClient.T2S("https://www.davemeyer.io/GoSpeechLess/T2S_Test_file_01.txt", "s3://"+bucket+"/testfile_03.mp3", *options)

	t2sClient.SetTempBucket(providers.ProviderAWS, bucket)
	t2sClient, err = t2sClient.T2S("https://"+bucket+".s3.amazonaws.com/T2S_Test_file_01.txt", "D:\\testfile_04.mp3", *options)

	err = t2sClient.CloseAllProviderClients()

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Speech successfully synthesized!")
}
