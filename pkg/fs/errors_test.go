package fs

import (
	"errors"
	"testing"
)

func TestListObjectsError(t *testing.T) {
	err := &ListObjectsError{
		err: errors.New("mock list objects error"),
	}

	recieved := err.Error()
	expected := "package fs: mock list objects error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}
