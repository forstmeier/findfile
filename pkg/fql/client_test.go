package fql

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

func TestConvertFQL(t *testing.T) {
	tests := []struct {
		description string
		input       map[string]interface{}
		parseOutput []byte
		parseError  error
		error       error
	}{
		{
			description: "error from parse fql helper function",
			input:       map[string]interface{}{},
			parseOutput: nil,
			parseError:  errors.New("mock parse fql error"),
			error:       &ErrorParseFQL{},
		},
		{
			description: "successful convert fql invocation",
			input:       map[string]interface{}{},
			parseOutput: []byte("test byte output"),
			error:       nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			c := &Client{
				parseFQL: func(ctx context.Context, accountID string, fqlQuery map[string]interface{}) ([]byte, error) {
					return test.parseOutput, test.parseError
				},
			}

			received, err := c.ConvertFQL(context.Background(), "account_id", test.input)

			if err != nil {
				switch e := test.error.(type) {
				case *ErrorParseFQL:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
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
