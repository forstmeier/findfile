package fql

import (
	"errors"
	"testing"
)

func TestErrorParseFQL(t *testing.T) {
	err := &ErrorParseFQL{err: errors.New("mock parse fql error")}

	recieved := err.Error()
	expected := "[fql] [convert fql]: mock parse fql error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}
