package cql

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	client := New()

	if client == nil {
		t.Error("error creating parser client")
	}
}

func TestConvertCQL(t *testing.T) {
	tests := []struct {
		description string
		input       map[string]interface{}
		parseOutput []byte
		parseError  error
		error       error
	}{
		{
			description: "error from parse cql helper function",
			input:       map[string]interface{}{},
			parseOutput: nil,
			parseError:  errors.New("mock parse cql error"),
			error:       &ErrorConvertCQL{},
		},
		{
			description: "successful convert cql invocation",
			input:       map[string]interface{}{},
			parseOutput: []byte("test byte output"),
			error:       nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			c := &Client{
				parseCQL: func(accountID string, cqlQuery map[string]interface{}) ([]byte, error) {
					return test.parseOutput, test.parseError
				},
			}

			received, err := c.ConvertCQL(context.Background(), "account_id", test.input)

			if err != nil {
				switch test.error.(type) {
				case *ErrorConvertCQL:
					var testError *ErrorConvertCQL
					if !errors.As(err, &testError) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, testError)
					}
				default:
					t.Fatalf("unexpected error type: %v", err)
				}

			} else {
				expected := []byte("test byte output")
				if bytes.Compare(received, expected) != 0 {
					t.Errorf("incorrect byptes, received: %v, expected: %v", received, expected)
				}
			}
		})
	}
}
