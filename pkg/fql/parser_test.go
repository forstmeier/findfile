package fql

import (
	"bytes"
	"context"
	"fmt"
	"testing"
)

func Test_parseFQL(t *testing.T) {
	tests := []struct {
		description string
		input       string
		error       error
	}{
		{
			description: "invalid json body received",
			input:       `{"search": {}}`,
			error:       errorMissingText,
		},
		{
			description: "successfull parse fql invocation",
			input:       `{"search": {"text": "lookup text", "page_number": 1, "coordinates": [[0.1,0.1],[0.5,0.5]]}}`,
			error:       nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			received, err := parseFQL(context.Background(), []byte(test.input))

			if err != nil {
				if err != test.error {
					t.Errorf("incorrect error, received: %s, expected: %s", err.Error(), test.error.Error())
				}
			} else {
				expected := fmt.Sprintf(
					queryString,
					"lookup text",
					0.1,
					0.5,
					0.1,
					0.5,
					1,
				)

				if bytes.Compare(received, []byte(expected)) != 0 {
					t.Errorf("incorrect byptes, received: %s, expected: %s", received, expected)
				}
			}
		})
	}
}

func Test_validateSearchJSON(t *testing.T) {
	tests := []struct {
		description string
		input       search
		error       error
	}{
		{
			description: "empty text field",
			input: search{
				Text: "",
			},
			error: errorMissingText,
		},
		{
			description: "empty page number field",
			input: search{
				Text:       "search value",
				PageNumber: 0,
			},
			error: errorPageNumberZero,
		},
		{
			description: "bottom right coordinates equal zero",
			input: search{
				Text:       "search value",
				PageNumber: 1,
				Coordinates: [2][2]float64{
					{
						float64(0),
						float64(0),
					},
					{
						float64(0),
						float64(0),
					},
				},
			},
			error: errorCoordinatesZero,
		},
		{
			description: "top left values equal bottom right values",
			input: search{
				Text:       "search value",
				PageNumber: 1,
				Coordinates: [2][2]float64{
					{
						float64(0.3),
						float64(0.4),
					},
					{
						float64(0.3),
						float64(0.4),
					},
				},
			},
			error: errorCoordinatesMisplaced,
		},
		{
			description: "top left values greater than bottom right values",
			input: search{
				Text:       "search value",
				PageNumber: 1,
				Coordinates: [2][2]float64{
					{
						float64(0.3),
						float64(0.4),
					},
					{
						float64(0.1),
						float64(0.2),
					},
				},
			},
			error: errorCoordinatesMisplaced,
		},
		{
			description: "successful invocation with correct fql query",
			input: search{
				Text:       "search value",
				PageNumber: 1,
				Coordinates: [2][2]float64{
					{
						float64(0.1),
						float64(0.2),
					},
					{
						float64(0.3),
						float64(0.4),
					},
				},
			},
			error: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			if err := validateSearchJSON(test.input); err != test.error {
				t.Errorf("incorrect error, received: %v, expected: %v", err, test.error)
			}
		})
	}
}
