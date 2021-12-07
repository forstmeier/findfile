package db

import (
	"errors"
	"testing"
)

func TestNewClientError(t *testing.T) {
	err := &NewClientError{
		err: errors.New("mock new client error"),
	}

	recieved := err.Error()
	expected := "package db: mock new client error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestExecuteCreateError(t *testing.T) {
	err := &ExecuteCreateError{
		err: errors.New("mock execute create error"),
	}

	recieved := err.Error()
	expected := "package db: mock execute create error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestMarshalDocumentError(t *testing.T) {
	err := &MarshalDocumentError{
		err: errors.New("mock marshal document error"),
	}

	recieved := err.Error()
	expected := "package db: mock marshal document error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestExecuteBulkError(t *testing.T) {
	err := &ExecuteBulkError{
		err: errors.New("mock execute bulk error"),
	}

	recieved := err.Error()
	expected := "package db: mock execute bulk error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestExecuteDeleteError(t *testing.T) {
	err := &ExecuteDeleteError{
		err: errors.New("mock execute delete error"),
	}

	recieved := err.Error()
	expected := "package db: mock execute delete error"

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

func TestReadQueryResponseBodyError(t *testing.T) {
	err := &ReadQueryResponseBodyError{
		err: errors.New("mock read query response body error"),
	}

	recieved := err.Error()
	expected := "package db: mock read query response body error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestUnmarshalQueryResponseBodyError(t *testing.T) {
	err := &UnmarshalQueryResponseBodyError{
		err: errors.New("mock unmarshal query response body error"),
	}

	recieved := err.Error()
	expected := "package db: mock unmarshal query response body error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}
