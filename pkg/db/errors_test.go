package db

import (
	"errors"
	"testing"
)

func TestErrorAddFolder(t *testing.T) {
	err := &ErrorAddFolder{
		err: errors.New("mock add folder error"),
	}

	recieved := err.Error()
	expected := "[db] [setup database] [add folder]: mock add folder error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorStartCrawler(t *testing.T) {
	err := &ErrorStartCrawler{
		err: errors.New("mock start crawler error"),
	}

	recieved := err.Error()
	expected := "[db] [add partition]: mock start crawler error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorUploadObject(t *testing.T) {
	err := &ErrorUploadObject{
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
