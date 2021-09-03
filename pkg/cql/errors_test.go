package cql

import (
	"errors"
	"testing"
)

func TestErrorParseCQL(t *testing.T) {
	err := &ErrorParseCQL{err: errors.New("mock parse cql error")}

	recieved := err.Error()
	expected := "[cql] [convert cql]: mock parse cql error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}
