package fql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

var (
	errorMissingText          = errors.New("search object must contain text")
	errorPageNumberZero       = errors.New("search object page number must not be \"0\"")
	errorCoordinatesZero      = errors.New("search object bottom coordinates cannot include zero")
	errorCoordinatesMisplaced = errors.New("search object bottom coordinates cannot be greater than or equal to top coordinates")
)

type body struct {
	Search search `json:"search"`
}

type search struct {
	Text        string        `json:"text"`
	PageNumber  int64         `json:"page_number"`
	Coordinates [2][2]float64 `json:"coordinates"`
}

const queryString = `
select id, file_key, file_bucket
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
		) as filtered_coordinates
		on lines.id = filtered_coordinates.line_id
	)
	and page_number = %d
);
`

func parseFQL(ctx context.Context, fqlQuery []byte) ([]byte, error) {
	bodyJSON := body{}
	if err := json.Unmarshal(fqlQuery, &bodyJSON); err != nil {
		return nil, err
	}

	searchJSON := bodyJSON.Search
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
