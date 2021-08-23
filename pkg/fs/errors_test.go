package fs

import (
	"errors"
	"testing"
)

func TestErrorAddNotification(t *testing.T) {
	err := &ErrorAddNotification{
		err: errors.New("mock add notification error"),
	}

	recieved := err.Error()
	expected := "[fs] [create file watcher] [add notification]: mock add notification error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorRemoveNotification(t *testing.T) {
	err := &ErrorRemoveNotification{
		err: errors.New("mock remove notification error"),
	}

	recieved := err.Error()
	expected := "[fs] [delete file watcher] [remove notification]: mock remove notification error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}
