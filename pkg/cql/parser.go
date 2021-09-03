package cql

import (
	"context"
	"errors"
)

var (
	errorTooManyAttributes    = errors.New("search object contains too many attributes")
	errorKeyNotSupported      = errors.New("search object must be under \"search\" key")
	errorTypeIncorrect        = errors.New("search object must be search object type")
	errorMissingText          = errors.New("search object must contain text")
	errorPageNumberZero       = errors.New("search object page number must not be \"0\"")
	errorCoordinatesZero      = errors.New("search object bottom coordinates cannot include zero")
	errorCoordinatesMisplaced = errors.New("search object bottom coordinates cannot be greater than or equal to top coordinates")
)

type search struct {
	Text        string        `json:"text"`
	PageNumber  int64         `json:"page_number"`
	Coordinates [2][2]float64 `json:"coordinates"`
}

func parseCQL(ctx context.Context, accountID string, cqlQuery map[string]interface{}) ([]byte, error) {
	if len(cqlQuery) > 1 {
		return nil, errorTooManyAttributes
	}

	searchValue, searchOK := cqlQuery["search"]
	if !searchOK {
		return nil, errorKeyNotSupported
	}

	searchJSON, typeOK := searchValue.(search)
	if !typeOK {
		return nil, errorTypeIncorrect
	}

	if err := validateSearchJSON(searchJSON); err != nil {
		return nil, err
	}

	query := []byte("TEMP")

	return query, nil
}

func validateSearchJSON(input search) error {
	if input.Text == "" {
		return errorMissingText
	}

	if input.PageNumber == 0 {
		return errorPageNumberZero
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
