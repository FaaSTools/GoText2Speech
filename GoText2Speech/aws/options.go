package aws

import (
	. "goTest/GoText2Speech/shared"
)

// SSMLModifiersDefined returns true if the options that AWS only supports using SSML are defined.
// So, if either SpeakingRate, Pitch or Volume have a non-default value, it returns true.
func SSMLModifiersDefined(options TextToSpeechOptions) bool {
	return (options.SpeakingRate != 1.0) || (options.Pitch != 0.0) || (options.Volume != 0.0)
}
