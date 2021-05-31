package csql

import (
	"errors"
	"testing"
)

func TestErrorParseCSQLJSON(t *testing.T) {
	err := &ErrorParseCSQLJSON{err: errors.New("mock parse csql json error")}

	recieved := err.Error()
	expected := "csql: mock parse csql json error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}
