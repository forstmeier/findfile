package csql

import (
	"errors"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	client := New()
	if client == nil {
		t.Error("error creating parser client")
	}
}

func TestCSQLToES(t *testing.T) {
	tests := []struct {
		description     string
		parseJSONOutput interface{}
		parseJSONError  error
		output          map[string]interface{}
		error           error
	}{
		{
			description:     "error parsing csql json",
			parseJSONOutput: nil,
			parseJSONError:  errors.New("mock parsing error"),
			output:          nil,
			error:           &ErrorParseCSQLJSON{err: errors.New("mock parsing error")},
		},
		{
			description: "successful invocation for csql to es",
			parseJSONOutput: map[string]interface{}{
				"search": map[string]interface{}{
					"text": "lookup",
					"coordinates": [2][2]float64{
						{0.1, 0.2},
						{0.3, 0.4},
					},
				},
			},
			parseJSONError: nil,
			output: map[string]interface{}{
				"search": map[string]interface{}{
					"text": "lookup",
					"coordinates": [2][2]float64{
						{0.1, 0.2},
						{0.3, 0.4},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			mockParseJSON := func(input interface{}) (interface{}, error) {
				return test.parseJSONOutput, test.parseJSONError
			}

			client := Client{
				parseJSON: mockParseJSON,
			}

			output, err := client.CSQLToES(map[string]interface{}{})

			if err != nil {
				var testError *ErrorParseCSQLJSON
				if !errors.As(err, &testError) {
					t.Errorf("incorrect error, received: %v, expected: %v", err, test.error)
				}
			} else {
				if !reflect.DeepEqual(output, test.output) {
					t.Errorf("incorrect output, received: %v, expected: %v", output, test.output)
				}
			}
		})
	}
}
