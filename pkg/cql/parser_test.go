package cql

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Test_parseCQL(t *testing.T) {
	tests := []struct {
		description string
		input       map[string]interface{}
		error       error
	}{
		{
			description: "too many attributes received",
			input: map[string]interface{}{
				"search":  search{},
				"another": search{},
			},
			error: errorTooManyAttributes,
		},
		{
			description: "incorrect search key received",
			input: map[string]interface{}{
				"unsupported": search{},
			},
			error: errorKeyNotSupported,
		},
		{
			description: "incorrect search type received",
			input: map[string]interface{}{
				"search": "not_search_type",
			},
			error: errorTypeIncorrect,
		},
		{
			description: "search object validation error",
			input: map[string]interface{}{
				"search": search{},
			},
			error: errorMissingText,
		},
		{
			description: "successfull parse invocation",
			input: map[string]interface{}{
				"search": search{
					Text:       "lookup text",
					PageNumber: 1,
					Coordinates: [2][2]float64{
						{0.1, 0.1},
						{0.5, 0.5},
					},
				},
			},
			error: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			received, err := parseCQL("account_id", test.input)

			if err != nil {
				if err != test.error {
					t.Errorf("incorrect error, received: %s, expected: %s", err.Error(), test.error.Error())
				}
			} else {
				coordinates := [2][2]float64{
					{0.1, 0.1},
					{0.5, 0.5},
				}

				bsonQuery := newBSONQuery("account_id", 1, "lookup text", coordinates)
				expected, err := json.Marshal(bsonQuery)
				if err != nil {
					t.Fatalf("error creating expected bson query: %s", err.Error())
				}

				if bytes.Compare(received, expected) != 0 {
					t.Errorf("incorrect byptes, received: %v, expected: %v", received, expected)
				}
			}
		})
	}
}

func Test_newBSONQuery(t *testing.T) {
	pageNumber := int64(1)
	text := "lookup text"
	coordinates := [2][2]float64{
		{0.1, 0.1},
		{0.5, 0.5},
	}

	received := newBSONQuery("account_id", pageNumber, text, coordinates)

	expected := bson.D{
		primitive.E{
			Key:   "account_id",
			Value: "account_id",
		},
		primitive.E{
			Key: "pages",
			Value: bson.D{
				primitive.E{
					Key:   "page_number",
					Value: pageNumber,
				},
				primitive.E{
					Key: "lines",
					Value: bson.D{
						primitive.E{
							Key: "$text",
							Value: bson.D{
								primitive.E{
									Key:   "$search",
									Value: text,
								},
							},
						},
						primitive.E{
							Key: "coordinates.top_left.x",
							Value: bson.D{
								primitive.E{
									Key:   "$lte",
									Value: coordinates[1][0],
								},
							},
						},
						primitive.E{
							Key: "coordinates.bottom_right.x",
							Value: bson.D{
								primitive.E{
									Key:   "$gte",
									Value: coordinates[0][0],
								},
							},
						},
						primitive.E{
							Key: "coordinates.top_right.y",
							Value: bson.D{
								primitive.E{
									Key:   "$gte",
									Value: coordinates[1][1],
								},
							},
						},
						primitive.E{
							Key: "coordinates.bottom_left.y",
							Value: bson.D{
								primitive.E{
									Key:   "$lte",
									Value: coordinates[0][1],
								},
							},
						},
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(received, expected) {
		t.Errorf("incorrect query, received: %+v, expected: %+v", received, expected)
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
			description: "successful invocation with correct cql query",
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
