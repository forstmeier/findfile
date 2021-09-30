package db

import (
	"errors"
	"testing"
)

func TestAddFolderError(t *testing.T) {
	err := &AddFolderError{
		err: errors.New("mock add folder error"),
	}

	recieved := err.Error()
	expected := "[db] [setup database] [add folder]: mock add folder error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestUploadObjectError(t *testing.T) {
	err := &UploadObjectError{
		err:      errors.New("mock upload object error"),
		function: "function",
		entity:   "entity",
	}

	recieved := err.Error()
	expected := "[db] [function] [upload object] [entity: entity]: mock upload object error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestExecuteQueryError(t *testing.T) {
	err := &ExecuteQueryError{
		err:      errors.New("mock execute query error"),
		function: "function",
	}

	recieved := err.Error()
	expected := "[db] [function] [execute query]: mock execute query error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestGetQueryResultsError(t *testing.T) {
	err := &GetQueryResultsError{
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

func TestDeleteDocumentsByKeysError(t *testing.T) {
	err := &DeleteDocumentsByKeysError{
		err: errors.New("mock delete documents by keys error"),
	}

	recieved := err.Error()
	expected := "[db] [delete documents] [delete documents by keys]: mock delete documents by keys error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}
