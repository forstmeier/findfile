package fql

import (
	"context"
	"errors"
	"fmt"
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

const queryString = `
select account_id, filename, filepath
from documents
where id in (
	select document_id
	from pages
	where id in (
		select page_id
		from (
			select id, page_id
			from lines
			where text = '%s'
		) as lines
		inner join (
			select line_id
			from coordinates
			where %f <= top_left_y
			and %f >= bottom_right_y
			and %f <= top_left_x
			and %f >= bottom_right_x
			and partition_0 = '%s'
		) as filtered_coordinates
		on lines.id = filtered_coordinates.line_id
	)
	and page_number = %d
);
`

func parseFQL(ctx context.Context, accountID string, fqlQuery map[string]interface{}) ([]byte, error) {
	if len(fqlQuery) > 1 {
		return nil, errorTooManyAttributes
	}

	searchValue, searchOK := fqlQuery["search"]
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

	query := fmt.Sprintf(
		queryString,
		searchJSON.Text,
		searchJSON.Coordinates[0][1],
		searchJSON.Coordinates[1][1],
		searchJSON.Coordinates[0][0],
		searchJSON.Coordinates[1][0],
		accountID,
		searchJSON.PageNumber,
	)

	return []byte(query), nil
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
