package csql

import (
	"encoding/json"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	errorKeyNotSupported      = errors.New("search object key received not supported")
	errorTooManyAttributes    = errors.New("search object contains too many attributes")
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

func parseCSQL(accountID string, csqlQuery map[string]interface{}) ([]byte, error) {
	var output []byte

	if len(csqlQuery) > 1 {
		return nil, errorTooManyAttributes
	}
	if searchValue, searchOK := csqlQuery["search"]; searchOK {
		searchJSON, typeOK := searchValue.(search)
		if !typeOK {
			return nil, errorTypeIncorrect
		}

		if err := validateSearchJSON(searchJSON); err != nil {
			return nil, err
		}

		bsonQuery := newBSONQuery(
			accountID,
			searchJSON.PageNumber,
			searchJSON.Text,
			searchJSON.Coordinates,
		)

		var err error
		output, err = json.Marshal(bsonQuery)
		if err != nil {
			return nil, err
		}

	} else {
		return nil, errorKeyNotSupported
	}

	return output, nil
}

func newBSONQuery(accountID string, pageNumber int64, text string, coordinates [2][2]float64) bson.D {
	return bson.D{
		primitive.E{
			Key:   "account_id",
			Value: accountID,
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
