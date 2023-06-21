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

	/*
		var MyEvent struct { // don't count this struct
			Text         string `json:"Text"`
			VoiceId      string `json:"VoiceId"`
			TargetBucket string `json:"TargetBucket"`
			TargetKey    string `json:"TargetKey"`
		}

		data := "{\"Text\":\"Hello World!\",\"VoiceId\":\"en-US-Polyglot-1\",\"TargetBucket\":\"davemeyer-test\",\"TargetKey\":\"example01/example01-gcp.mp3\"}"
		fmt.Println("data:", data)
		err := json.NewDecoder(strings.NewReader(data)).Decode(&MyEvent)
		if err != nil {
			fmt.Println("err", err.Error())
			return
		}
		fmt.Println("MyEvent.Text:", MyEvent.Text)
	*/

	t2sClient := CreateGoT2SClient(nil, "us-east-1")

	options := GetDefaultTextToSpeechOptions()
	options.Provider = providers.ProviderAWS
	options.VoiceConfig.VoiceIdConfig = VoiceIdConfig{VoiceId: "en-US-News-N"}

	var err error = nil
	/*
		t2sClient, err = t2sClient.T2SDirect("<speak><prosody volume=\"10.000dB\">Hello World, how are you today? Lovely day, isn't it?</prosody></speak>", "s3://davemeyer-test/testfile.mp3", *options)
		t2sClient, err = t2sClient.T2SDirect("Test", "s3://davemeyer-test/testfile_02.mp3", *options)
		t2sClient, err = t2sClient.T2S("https://www.davemeyer.io/GoSpeechLess/T2S_Test_file_01.txt", "s3://davemeyer-test/testfile_03.mp3", *options)

		t2sClient.SetTempBucket(providers.ProviderAWS, "davemeyer-test")
	*/
	//t2sClient, err = t2sClient.T2S("https://davemeyer-test.s3.amazonaws.com/T2S_Test_file_01.txt", "D:\\testfile_04.mp3", *options)

	options.Provider = providers.ProviderGCP
	//options.VoiceConfig.VoiceIdConfig = VoiceIdConfig{VoiceId: "Salli", Engine: "neural"}
	/*
		options.VoiceConfig.VoiceParamsConfig = VoiceParamsConfig{
			LanguageCode: "en-US",
			Gender:       VoiceGenderMale,
			//Engine:       "",
		}
	*/
	t2sClient, err = t2sClient.T2SDirect("<speak><prosody volume=\"10.000dB\">Hello World, how are you today? Lovely day, isn't it?</prosody></speak>", "gs://davemeyer-test/testfile.mp3", *options)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	err = t2sClient.CloseAllProviderClients()

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Speech successfully synthesized!")
}
