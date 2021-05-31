package csql

import (
	"encoding/json"
	"errors"
	"fmt"
)

var (
	errorTypeNotSupported     = errors.New("search object received not supported")
	errorTooManyAttributes    = errors.New("search object contains too many attributes")
	errorMissingText          = errors.New("search object must contain text")
	errorCoordinatesZero      = errors.New("search object bottom coordinates cannot include zero")
	errorCoordinatesMisplaced = errors.New("search object bottom coordinates cannot be greater than or equal to top coordinates")
)

type searchObject struct {
	Text        string        `json:"text"`
	Coordinates [2][2]float64 `json:"coordinates"`
}

func parseJSON(input interface{}) (interface{}, error) {
	switch inputJSON := input.(type) {
	case map[string]interface{}:
		if len(inputJSON) != 1 {
			return nil, errorTooManyAttributes
		}

		if searchValue, ok := inputJSON["search"]; ok {
			searchBytes, err := json.Marshal(searchValue)
			if err != nil {
				return nil, err
			}

			searchJSON := searchObject{}
			if err := json.Unmarshal(searchBytes, &searchJSON); err != nil {
				return nil, err
			}

			if err := validateSearchJSON(searchJSON); err != nil {
				return nil, err
			}

			return newSearchQuery(searchJSON), nil
		} else if andValue, ok := inputJSON["and"]; ok {
			andJSON, err := parseJSON(andValue)
			if err != nil {
				return nil, err
			}

			return newOperatorQuery(andJSON, "must"), nil
		} else if orValue, ok := inputJSON["or"]; ok {
			orJSON, err := parseJSON(orValue)
			if err != nil {
				return nil, err
			}

			return newOperatorQuery(orJSON, "should"), nil
		} else if notValue, ok := inputJSON["not"]; ok {
			notJSON, err := parseJSON(notValue)
			if err != nil {
				return nil, err
			}

			return newOperatorQuery(notJSON, "must_not"), nil
		} else {
			key := ""
			for k := range inputJSON {
				key = k
			}

			return nil, fmt.Errorf("key \"%s\" not supported", key)
		}
	case []interface{}:
		objectArray := make([]interface{}, len(inputJSON))

		for i, jsonObject := range inputJSON {
			subObject, err := parseJSON(jsonObject)
			if err != nil {
				return nil, err
			}

			objectArray[i] = subObject
		}

		return objectArray, nil
	}

	return nil, errorTypeNotSupported
}

func validateSearchJSON(input searchObject) error {
	if input.Text == "" {
		return errorMissingText
	}

	topLeft := input.Coordinates[0]
	bottomRight := input.Coordinates[1]

	if bottomRight[0] == 0 || bottomRight[1] == 0 {
		return errorCoordinatesZero
	}

	if topLeft[0] >= bottomRight[0] || topLeft[1] >= bottomRight[1] {
		return errorCoordinatesMisplaced
	}

	return nil
}
