package db

import (
	"errors"
	"testing"
)

func TestErrorReadPEMFile(t *testing.T) {
	err := &ErrorReadPEMFile{err: errors.New("mock ioutil read file error")}

	recieved := err.Error()
	expected := "db: get tls config: mock ioutil read file error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorParsePEMFile(t *testing.T) {
	err := &ErrorParsePEMFile{}

	recieved := err.Error()
	expected := "db: get tls config: error parsing pem file"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorNewClient(t *testing.T) {
	err := &ErrorNewClient{err: errors.New("mock new client error")}

	recieved := err.Error()
	expected := "db: new client: mock new client error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorUpdateDocument(t *testing.T) {
	err := &ErrorUpdateDocument{err: errors.New("mock create/update documents error")}

	recieved := err.Error()
	expected := "db: create or update: mock create/update documents error"

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
