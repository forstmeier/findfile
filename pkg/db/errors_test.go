package db

import (
	"errors"
	"testing"
)

func TestErrorNewClient(t *testing.T) {
	err := &ErrorNewClient{err: errors.New("mock new client error")}

	recieved := err.Error()
	expected := "db: new client: mock new client error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorCreateDocuments(t *testing.T) {
	err := &ErrorCreateDocuments{err: errors.New("mock create documents error")}

	recieved := err.Error()
	expected := "db: create: mock create documents error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorUpdateDocuments(t *testing.T) {
	err := &ErrorUpdateDocuments{err: errors.New("mock update documents error")}

	recieved := err.Error()
	expected := "db: update: mock update documents error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorDeleteDocuments(t *testing.T) {
	err := &ErrorDeleteDocuments{err: errors.New("mock delete documents error")}

	recieved := err.Error()
	expected := "db: delete: mock delete documents error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorQueryDocuments(t *testing.T) {
	err := &ErrorQueryDocuments{err: errors.New("mock query documents error")}

	recieved := err.Error()
	expected := "db: query: mock query documents error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorParseQueryResults(t *testing.T) {
	err := &ErrorParseQueryResults{err: errors.New("mock parse query results error")}

	recieved := err.Error()
	expected := "db: query: mock parse query results error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}
