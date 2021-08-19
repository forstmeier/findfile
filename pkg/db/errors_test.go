package db

import (
	"errors"
	"testing"
)

func TestErrorUploadObject(t *testing.T) {
	err := &ErrorUploadObject{
		err:      errors.New("mock upload object error"),
		function: "function",
	}

	recieved := err.Error()
	expected := "[db] [function] [upload object]: mock upload object error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorExecuteQuery(t *testing.T) {
	err := &ErrorExecuteQuery{
		err:      errors.New("mock execute query error"),
		function: "function",
	}

	recieved := err.Error()
	expected := "[db] [function] [execute query]: mock execute query error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorGetQueryResults(t *testing.T) {
	err := &ErrorGetQueryResults{
		err:         errors.New("mock get query results error"),
		function:    "function",
		subfunction: "subfunction",
	}

	recieved := err.Error()
	expected := "[db] [function] [subfunction]: mock get query results error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorListDocumentKeys(t *testing.T) {
	err := &ErrorListDocumentKeys{
		err: errors.New("mock list document keys error"),
	}

	recieved := err.Error()
	expected := "[db] [delete documents] [list document keys]: mock list document keys error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorDeleteDocumentsByKeys(t *testing.T) {
	err := &ErrorDeleteDocumentsByKeys{
		err: errors.New("mock delete documents by keys error"),
	}

	recieved := err.Error()
	expected := "[db] [delete documents] [delete documents by keys]: mock delete documents by keys error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}
