package shared

import "testing"

func TestGetDefaultVoiceParamsConfig(t *testing.T) {
	options := GetDefaultVoiceParamsConfig()
	if (options == VoiceParamsConfig{}) {
		t.Error("Default voice parameter config was empty")
		return
	}
	if options.LanguageCode == "" {
		t.Error("Default voice language was empty")
	}
	if options.Gender == VoiceGenderUnspecified {
		t.Error("Default voice gender was empty")
	}
}

func TestVoiceIdConfig_IsEmpty(t *testing.T) {
	config1 := VoiceIdConfig{}
	if !config1.IsEmpty() {
		t.Error("VoiceIdConfig was not seen as empty, even though it was empty")
	}

	config2 := VoiceIdConfig{VoiceId: ""}
	if !config2.IsEmpty() {
		t.Error("VoiceIdConfig was not seen as empty, even though the VoiceId was undefined")
	}

	config3 := VoiceIdConfig{VoiceId: "Test"}
	if config3.IsEmpty() {
		t.Error("VoiceIdConfig was seen as empty, even though it had a defined VoiceId")
	}
}

func TestVoiceParamsConfig_IsEmpty(t *testing.T) {
	config1 := VoiceParamsConfig{}
	if !config1.IsEmpty() {
		t.Error("VoiceParamsConfig was not seen as empty, even though it was empty")
	}

	config2 := VoiceParamsConfig{LanguageCode: "", Gender: VoiceGenderUnspecified}
	if !config2.IsEmpty() {
		t.Error("VoiceParamsConfig was not seen as empty, even though LanguageCode and Gender were undefined")
	}

	config3 := VoiceParamsConfig{LanguageCode: "en-US", Gender: VoiceGenderUnspecified}
	if config3.IsEmpty() {
		t.Error("VoiceParamsConfig was seen as empty, even though LanguageCode was defined")
	}

	config4 := VoiceParamsConfig{LanguageCode: "", Gender: VoiceGenderMale}
	if config4.IsEmpty() {
		t.Error("VoiceParamsConfig was seen as empty, even though Gender was defined")
	}

	config5 := VoiceParamsConfig{LanguageCode: "en-US", Gender: VoiceGenderMale}
	if config5.IsEmpty() {
		t.Error("VoiceParamsConfig was seen as empty, even though LanguageCode and Gender were defined")
	}
}