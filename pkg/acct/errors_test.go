package acct

import (
	"errors"
	"testing"
)

func TestErrorPutItem(t *testing.T) {
	err := &ErrorPutItem{err: errors.New("mock put item error")}

	recieved := err.Error()
	expected := "acct: create account: mock put item error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorGetItem(t *testing.T) {
	err := &ErrorGetItem{err: errors.New("mock get item error")}

	recieved := err.Error()
	expected := "acct: read account: mock get item error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorIncorrectValue(t *testing.T) {
	err := &ErrorIncorrectValue{key: "test_key"}

	recieved := err.Error()
	expected := "acct: update account: incorrect key [test_key] in values"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorUpdateItem(t *testing.T) {
	err := &ErrorUpdateItem{err: errors.New("mock update item error")}

	recieved := err.Error()
	expected := "acct: update account: mock update item error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}

func TestErrorDeleteItem(t *testing.T) {
	err := &ErrorDeleteItem{err: errors.New("mock delete item error")}

	recieved := err.Error()
	expected := "acct: delete account: mock delete item error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}
