package aws

import (
	"goTest/GoText2Speech/shared"
	"strings"
	"testing"
)

func TestGetBucketAndKeyFromAWSDestination(t *testing.T) {

	type TestData struct {
		input      string
		wantBucket string
		wantKey    string
		error      bool
	}

	testData := []TestData{
		{
			input:      "s3://testbucket/file1.test",
			wantBucket: "testbucket",
			wantKey:    "file1.test",
			error:      false,
		},
		{
			input:      "s3://testbucket/dir1/file1.test",
			wantBucket: "testbucket",
			wantKey:    "dir1/file1.test",
			error:      false,
		},
		{
			input:      "s3://testbucket/dir1/dir2/file1.test",
			wantBucket: "testbucket",
			wantKey:    "dir1/dir2/file1.test",
			error:      false,
		},
		{
			input:      "testbucket/dir1/file1.test",
			wantBucket: "",
			wantKey:    "",
			error:      true,
		},
		{
			input:      "s3:/testbucket/dir1/file1.test",
			wantBucket: "",
			wantKey:    "",
			error:      true,
		},
		{
			input:      "http://testbucket/dir1/file1.test",
			wantBucket: "",
			wantKey:    "",
			error:      true,
		},
		{
			input:      "https://testbucket.s3.eu-central-1.amazonaws.com/file1",
			wantBucket: "testbucket",
			wantKey:    "file1",
			error:      false,
		},
		{
			input:      "https://testbucket.s3.eu-central-1.amazonaws.com/file1.test",
			wantBucket: "testbucket",
			wantKey:    "file1.test",
			error:      false,
		},
		{
			input:      "https://testbucket.s3.eu-central-1.amazonaws.com/dir1/file1.test",
			wantBucket: "testbucket",
			wantKey:    "dir1/file1.test",
			error:      false,
		},
		{
			input:      "https://testbucket.s3.eu-central-1.amazonaws.com/dir1/dir2/file1.test",
			wantBucket: "testbucket",
			wantKey:    "dir1/dir2/file1.test",
			error:      false,
		},
	}

	for _, test := range testData {
		resultBucket, resultKey, resultError := GetBucketAndKeyFromAWSDestination(test.input)

		if (resultError == nil) && test.error {
			t.Errorf("Error was expected, but no error was thrown for input '%s'.", test.input)
			continue
		}
		if resultError != nil {
			if !test.error {
				t.Errorf("No error was expected, but an error was thrown for input '%s'.", test.input)
				continue
			}
			continue // error was expected and error was thrown -> success
		}

		if !strings.EqualFold(resultBucket, test.wantBucket) {
			t.Errorf("Expected bucket to be '%s', but was '%s'.", test.wantBucket, resultBucket)
		}
		if !strings.EqualFold(resultKey, test.wantKey) {
			t.Errorf("Expected key to be '%s', but was '%s'.", test.wantKey, resultKey)
		}
	}
}

func TestAddFileExtensionToDestinationIfNeededDeactivated(t *testing.T) {
	options := shared.TextToSpeechOptions{AddFileExtension: false}
	outputFormatRaw, err1 := AudioFormatToAWSValue(shared.AudioFormatMp3)
	destination, err2 := AddFileExtensionToDestinationIfNeeded(options, outputFormatRaw, "test1")

	if err1 != nil {
		t.Errorf("AudioFormatToAWSValue returned error: %s\n", err1.Error())
		return
	}
	if err2 != nil {
		t.Errorf("AddFileExtensionToDestinationIfNeeded returned error: %s\n", err2.Error())
		return
	}

	if !strings.EqualFold(destination, "test1") {
		t.Errorf("File extension was added, even though 'AddFileExtension' option was false.")
	}
}

func TestAddFileExtensionToDestinationIfNeeded(t *testing.T) {
	options := shared.TextToSpeechOptions{AddFileExtension: true}
	outputFormatRaw, err1 := AudioFormatToAWSValue(shared.AudioFormatMp3)
	destination, err2 := AddFileExtensionToDestinationIfNeeded(options, outputFormatRaw, "test1")

	if err1 != nil {
		t.Errorf("AudioFormatToAWSValue returned error: %s\n", err1.Error())
		return
	}
	if err2 != nil {
		t.Errorf("AddFileExtensionToDestinationIfNeeded returned error: %s\n", err2.Error())
		return
	}

	if strings.EqualFold(destination, "test1") {
		t.Errorf("File extension was not added, even though 'AddFileExtension' option was true.")
		return
	}
	if !strings.EqualFold(destination, "test1.mp3") {
		t.Errorf("Wrong file extension was added. Actual value: %s\n", destination)
	}
}

func TestAddFileExtensionToDestinationIfNeededNotNeeded(t *testing.T) {
	options := shared.TextToSpeechOptions{AddFileExtension: true}
	outputFormatRaw, err1 := AudioFormatToAWSValue(shared.AudioFormatMp3)
	destination, err2 := AddFileExtensionToDestinationIfNeeded(options, outputFormatRaw, "test1.mp3")

	if err1 != nil {
		t.Errorf("AudioFormatToAWSValue returned error: %s\n", err1.Error())
		return
	}
	if err2 != nil {
		t.Errorf("AddFileExtensionToDestinationIfNeeded returned error: %s\n", err2.Error())
		return
	}

	if !strings.EqualFold(destination, "test1.mp3") {
		t.Errorf("File extension was incorrectly added. Actual value: %s\n", destination)
	}
}

func TestAddFileExtensionToDestinationIfNeededOtherExtension(t *testing.T) {
	options := shared.TextToSpeechOptions{AddFileExtension: true}
	outputFormatRaw, err1 := AudioFormatToAWSValue(shared.AudioFormatMp3)
	destination, err2 := AddFileExtensionToDestinationIfNeeded(options, outputFormatRaw, "test1.wav")

	if err1 != nil {
		t.Errorf("AudioFormatToAWSValue returned error: %s\n", err1.Error())
		return
	}
	if err2 != nil {
		t.Errorf("AddFileExtensionToDestinationIfNeeded returned error: %s\n", err2.Error())
		return
	}

	if !strings.EqualFold(destination, "test1.wav.mp3") {
		t.Errorf("File extension was incorrectly added. Actual value: %s\n", destination)
	}
}
