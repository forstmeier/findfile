package db

import (
	"errors"
	"testing"
)

func TestErrorListingObjects(t *testing.T) {
	err := &ErrorListingObjects{
		err:    errors.New("mock list objects error"),
		action: "action",
	}

	recieved := err.Error()
	expected := "db: action documents: mock list objects error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorMarshalData(t *testing.T) {
	err := &ErrorMarshalData{
		err:    errors.New("mock marshal data error"),
		entity: "entity",
	}

	recieved := err.Error()
	expected := "db: create or update documents: [entity] mock marshal data error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorPutObject(t *testing.T) {
	err := &ErrorPutObject{
		err:    errors.New("mock put objects error"),
		entity: "entity",
	}

	recieved := err.Error()
	expected := "db: create or update documents: [entity] mock put objects error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}
