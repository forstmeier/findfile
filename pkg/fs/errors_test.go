package fs

import (
	"errors"
	"testing"
)

func TestErrorNewClient(t *testing.T) {
	err := &ErrorNewClient{err: errors.New("mock new client error")}

	recieved := err.Error()
	expected := "fs: new client: mock new client error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorPresignURL(t *testing.T) {
	err := &ErrorPresignURL{err: errors.New("mock presign url error")}

	recieved := err.Error()
	expected := "fs: presign: mock presign url error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorDeleteObjects(t *testing.T) {
	err := &ErrorDeleteObjects{err: errors.New("mock delete object error")}

	recieved := err.Error()
	expected := "fs: delete files: mock delete object error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}
