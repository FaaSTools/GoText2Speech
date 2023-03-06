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

	err := T2SDirect(s, destination, TextToSpeechOptions{
		TextType:    TextTypeText,
		VoiceConfig: VoiceConfig{VoiceIdConfig: VoiceIdConfig{VoiceId: "Joanna"}},
		// Alternative voice config
		//VoiceConfig: VoiceConfig{VoiceParamsConfig{LanguageCode: "en-US", Gender: VoiceGender_FEMALE},
		SpeakingRate: 1.1,
		Pitch:        0.0,
		Volume:       0.0,
	})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Speech successfully synthesized!")
}
