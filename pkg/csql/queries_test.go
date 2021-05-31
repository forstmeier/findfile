package csql

import (
	"reflect"
	"testing"
)

func Test_newSearchQuery(t *testing.T) {
	searchText := "lookup"
	topLeft := 0.1
	topRight := 0.2
	bottomLeft := 0.3
	bottomRight := 0.4

	searchJSON := searchObject{
		Text: searchText,
		Coordinates: [2][2]float64{
			{
				topLeft,
				topRight,
			},
			{
				bottomLeft,
				bottomRight,
			},
		},
	}

	received := newSearchQuery(searchJSON)

	expected := map[string]interface{}{
		"bool": map[string]interface{}{
			"must": map[string]interface{}{
				"match": map[string]interface{}{
					"text": searchText,
				},
			},
			"filter": map[string]interface{}{
				"geo_shape": map[string]interface{}{
					"coordinates": map[string]interface{}{
						"shape": map[string]interface{}{
							"type": "envelope",
							"coordinates": [2][2]float64{
								{
									bottomLeft,
									bottomRight,
								},
								{
									topLeft,
									topRight,
								},
							},
						},
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(received, expected) {
		t.Errorf("incorrect search query, received: %v, expected: %v", received, expected)
	}
}

func Test_newOperatorQuery(t *testing.T) {
	value := []interface{}{
		map[string]interface{}{
			"search": map[string]interface{}{
				"text": "lookup",
				"coordinates": [2][2]float64{
					{
						0.1,
						0.2,
					},
					{
						0.3,
						0.4,
					},
				},
			},
		},
	}
	name := "and"

	received := newOperatorQuery(value, name)

	expected := map[string]interface{}{
		"bool": map[string]interface{}{
			name: value,
		},
	}

	if !reflect.DeepEqual(received, expected) {
		t.Errorf("incorrect operator query, received: %v, expected: %v", received, expected)
	}
}
