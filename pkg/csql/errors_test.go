package csql

import (
	"errors"
	"testing"
)

func TestErrorConvertCSQL(t *testing.T) {
	err := &ErrorConvertCSQL{err: errors.New("mock convert csql error")}

	recieved := err.Error()
	expected := "csql: convert csql: mock convert csql error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}
