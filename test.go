package main

import (
	"fmt"
	"os"
)

import (
	. "goTest/GoText2Speech"
	. "goTest/GoText2Speech/shared"
)

// main shows how T2S might be executed.
func main() {

	fmt.Println("Synthesizing speech...")
	s := "Hello World, how are you today? Lovely day, isn't it?"
	destination := "testfile.mp3"

	options := GetDefaultTextToSpeechOptions()
	options.VoiceConfig.VoiceIdConfig = VoiceIdConfig{VoiceId: "Joanna"}
	// Alternative voice config
	//options.VoiceConfig.VoiceParamsConfig = VoiceParamsConfig{LanguageCode: "en-US", Gender: VoiceGenderFemale}
	options.SpeakingRate = 1.1

	err := T2SDirect(s, destination, *options)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Speech successfully synthesized!")
}
