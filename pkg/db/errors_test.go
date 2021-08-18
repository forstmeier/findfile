package db

import (
	"errors"
	"testing"
)

func TestErrorUploadObject(t *testing.T) {
	err := &ErrorUploadObject{
		err:      errors.New("mock upload object error"),
		function: "function",
	}

	recieved := err.Error()
	expected := "[db] [function] [upload object]: mock upload object error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}
