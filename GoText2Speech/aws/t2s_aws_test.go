package aws

import (
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
