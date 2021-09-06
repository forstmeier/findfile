package pars

import (
	"errors"
	"testing"
)

func TestErrorAnalyzeDocument(t *testing.T) {
	err := &ErrorAnalyzeDocument{err: errors.New("mock analyze doc error")}

	recieved := err.Error()
	expected := "[pars] [analyze document]: mock analyze doc error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}
