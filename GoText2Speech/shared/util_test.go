package shared

import "testing"

func TestIncludesAudioFormat(t *testing.T) {

	formats1 := []AudioFormat{
		AudioFormatMp3,
		AudioFormatOgg,
	}

	format1 := AudioFormatMp3
	format2 := AudioFormatAlaw

	if !IncludesAudioFormat(formats1, format1) {
		t.Error("IncludesAudioFormat returned false, even though given audio format was in audio formats array.")
	}
	if IncludesAudioFormat(formats1, format2) {
		t.Error("IncludesAudioFormat returned true, even though given audio format was not in audio formats array.")
	}
}
