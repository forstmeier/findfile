package pars

import (
	"errors"
	"testing"
)

func TestErrorParseDocument(t *testing.T) {
	err := &ErrorParseDocument{err: errors.New("mock parse doc error")}

	recieved := err.Error()
	expected := "[pars] [parse document]: mock parse doc error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}
