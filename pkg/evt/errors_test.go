package evt

import (
	"errors"
	"testing"
)

func TestGetEventValuesError(t *testing.T) {
	err := &GetEventValuesError{
		err: errors.New("mock get event values error"),
	}

	recieved := err.Error()
	expected := "package evt: mock get event values error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestPutEventValuesError(t *testing.T) {
	err := &PutEventValuesError{
		err: errors.New("mock put event values error"),
	}

	recieved := err.Error()
	expected := "package evt: mock put event values error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}
