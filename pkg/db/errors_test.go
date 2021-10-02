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
	expected := "package db: mock add folder error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestUploadObjectError(t *testing.T) {
	err := &UploadObjectError{
		err: errors.New("mock upload object error"),
	}

	recieved := err.Error()
	expected := "package db: mock upload object error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestExecuteQueryError(t *testing.T) {
	err := &ExecuteQueryError{
		err: errors.New("mock execute query error"),
	}

	recieved := err.Error()
	expected := "package db: mock execute query error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestGetQueryResultsError(t *testing.T) {
	err := &GetQueryResultsError{
		err: errors.New("mock get query results error"),
	}

	recieved := err.Error()
	expected := "package db: mock get query results error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestDeleteDocumentsByKeysError(t *testing.T) {
	err := &DeleteDocumentsByKeysError{
		err: errors.New("mock delete documents by keys error"),
	}

	recieved := err.Error()
	expected := "package db: mock delete documents by keys error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}
