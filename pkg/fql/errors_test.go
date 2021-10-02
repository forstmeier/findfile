package fql

import (
	"errors"
	"testing"
)

func TestParseFQLError(t *testing.T) {
	err := &ParseFQLError{err: errors.New("mock parse fql error")}

	recieved := err.Error()
	expected := "package fql: mock parse fql error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}
