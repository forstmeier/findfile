package infra

import (
	"errors"
	"testing"
)

func TestErrorCreateFilesystem(t *testing.T) {
	err := &ErrorCreateFilesystem{err: errors.New("mock create filesystem error")}

	recieved := err.Error()
	expected := "infra: create filesystem: mock create filesystem error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorDeleteFilesystem(t *testing.T) {
	err := &ErrorDeleteFilesystem{err: errors.New("mock delete filesystem error")}

	recieved := err.Error()
	expected := "infra: delete filesystem: mock delete filesystem error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorCreateDatabase(t *testing.T) {
	err := &ErrorCreateDatabase{err: errors.New("mock create database error")}

	recieved := err.Error()
	expected := "infra: create database: mock create database error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorDeleteDatabase(t *testing.T) {
	err := &ErrorDeleteDatabase{err: errors.New("mock delete database error")}

	recieved := err.Error()
	expected := "infra: delete database: mock delete database error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}
