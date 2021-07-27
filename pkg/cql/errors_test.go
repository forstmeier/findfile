package cql

import (
	"errors"
	"testing"
)

func TestErrorConvertCQL(t *testing.T) {
	err := &ErrorConvertCQL{err: errors.New("mock convert cql error")}

	recieved := err.Error()
	expected := "cql: convert cql: mock convert cql error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}
