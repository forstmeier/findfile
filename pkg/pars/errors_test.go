package pars

import (
	"errors"
	"testing"
)

func TestParseDocumentError(t *testing.T) {
	err := &ParseDocumentError{err: errors.New("mock parse doc error")}

	recieved := err.Error()
	expected := "package pars: mock parse doc error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}
