package db

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/findfiledev/api/pkg/pars"
)

var paths = []string{
	"documents",
	"pages",
	"lines",
	"coordinates",
}

var _ Databaser = &Client{}

// Client implements the db.Databaser methods using AWS Athena
// and AWS S3.
type Client struct {
	helper helper
}

// New generates a db.Client pointer instance with AWS Athena,
// AWS S3, and AWS Glue clients.
func New(newSession *session.Session, databaseName, databaseBucket string) *Client {
	return &Client{
		helper: &help{
			databaseName:   databaseName,
			databaseBucket: databaseBucket,
			athenaClient:   athena.New(newSession),
			s3Client:       s3.New(newSession),
		},
	}
}

// SetupDatabase implements the db.Databaser.SetupDatabase method
// using AWS S3 and AWS Glue.
func (c *Client) SetupDatabase(ctx context.Context) error {
	for _, path := range paths {
		if err := c.helper.addFolder(ctx, path+"/"); err != nil {
			return &ErrorAddFolder{
				err: err,
			}
		}
	}

	return nil
}

// UpsertDocuments implements the db.Databaser.UpsertDocuments method
// using AWS S3.
//
// Note that JSON objects stored in S3 must be represented in a single line
// in their respective files in order for Athena to be able to query correctly.
func (c *Client) UpsertDocuments(ctx context.Context, documents []pars.Document) error {
	for _, document := range documents {
		documentID := document.ID

		documentJSON := struct {
			ID         string `json:"id"`
			Entity     string `json:"entity"`
			FileKey    string `json:"file_key"`
			FileBucket string `json:"file_bucket"`
		}{
			ID:         documentID,
			Entity:     "document",
			FileKey:    document.FileKey,
			FileBucket: document.FileBucket,
		}

		documentKey := fmt.Sprintf("%s/%s.json", paths[0], documentID)
		if err := c.helper.uploadObject(ctx, documentJSON, documentKey); err != nil {
			return &ErrorUploadObject{
				err:      err,
				function: "upsert documents",
				entity:   "document",
			}
		}

		for _, page := range document.Pages {
			pageID := page.ID

			pageJSON := struct {
				ID         string `json:"id"`
				Entity     string `json:"entity"`
				DocumentID string `json:"document_id"`
				PageNumber int64  `json:"page_number"`
			}{
				ID:         pageID,
				Entity:     "page",
				DocumentID: documentID,
				PageNumber: page.PageNumber,
			}

			pageKey := fmt.Sprintf("%s/%s.json", paths[1], pageID)
			if err := c.helper.uploadObject(ctx, pageJSON, pageKey); err != nil {
				return &ErrorUploadObject{
					err:      err,
					function: "upsert documents",
					entity:   "page",
				}
			}

			for _, line := range page.Lines {
				lineID := line.ID
				coordinates := line.Coordinates
				coordinatesID := coordinates.ID

				lineJSON := struct {
					ID     string `json:"id"`
					Entity string `json:"entity"`
					PageID string `json:"page_id"`
					Text   string `json:"text"`
				}{
					ID:     lineID,
					Entity: "line",
					PageID: pageID,
					Text:   line.Text,
				}

				lineKey := fmt.Sprintf("%s/%s.json", paths[2], lineID)
				if err := c.helper.uploadObject(ctx, lineJSON, lineKey); err != nil {
					return &ErrorUploadObject{
						err:      err,
						function: "upsert documents",
						entity:   "line",
					}
				}

				coordinatesJSON := struct {
					ID           string  `json:"id"`
					Entity       string  `json:"entity"`
					LineID       string  `json:"line_id"`
					TopLeftX     float64 `json:"top_left_x"`
					TopLeftY     float64 `json:"top_left_y"`
					TopRightX    float64 `json:"top_right_x"`
					TopRightY    float64 `json:"top_right_y"`
					BottomLeftX  float64 `json:"bottom_left_x"`
					BottomLeftY  float64 `json:"bottom_left_y"`
					BottomRightX float64 `json:"bottom_right_x"`
					BottomRightY float64 `json:"bottom_right_y"`
				}{
					ID:           coordinatesID,
					Entity:       "coordinates",
					LineID:       lineID,
					TopLeftX:     coordinates.TopLeft.X,
					TopLeftY:     coordinates.TopLeft.Y,
					TopRightX:    coordinates.TopRight.X,
					TopRightY:    coordinates.TopRight.Y,
					BottomLeftX:  coordinates.BottomLeft.X,
					BottomLeftY:  coordinates.BottomLeft.Y,
					BottomRightX: coordinates.BottomRight.X,
					BottomRightY: coordinates.BottomRight.Y,
				}

				coordinatesKey := fmt.Sprintf("%s/%s.json", paths[3], coordinatesID)
				if err := c.helper.uploadObject(ctx, coordinatesJSON, coordinatesKey); err != nil {
					return &ErrorUploadObject{
						err:      err,
						function: "upsert documents",
						entity:   "coordinates",
					}
				}
			}
		}
	}

	return nil
}

// DeleteDocuments implements the db.Databaser.DeleteDocuments method
// using AWS S3.
func (c *Client) DeleteDocuments(ctx context.Context, documentKeys []string) error {
	chunkSize := 1000 // S3 max delete objects count
	for i := 0; i < len(documentKeys); i += chunkSize {
		end := i + chunkSize
		if end > len(documentKeys) {
			end = len(documentKeys)
		}

		documentKeysSubset := documentKeys[i:end]
		if err := c.helper.deleteDocumentsByKeys(ctx, documentKeysSubset); err != nil {
			return &ErrorDeleteDocumentsByKeys{
				err: err,
			}
		}
	}

	return nil
}

// QueryDocumentsByFQL implements the db.Databaser.QueryDocumentsByFQL method
// using AWS Athena.
func (c *Client) QueryDocumentsByFQL(ctx context.Context, query []byte) ([]pars.Document, error) {
	executionID, err := c.helper.executeQuery(ctx, query)
	if err != nil {
		return nil, &ErrorExecuteQuery{
			err:      err,
			function: "query documents",
		}
	}

	documents, err := c.helper.getQueryResultDocuments(ctx, *executionID)
	if err != nil {
		return nil, &ErrorGetQueryResults{
			err:         err,
			function:    "query documents",
			subfunction: "get query result documents",
		}
	}

	return documents, nil
}

// QueryDocumentKeysByFileInfo implements the db.Databaser.QueryDocumentKeysByFileInfo
// method using AWS Athena.
func (c *Client) QueryDocumentKeysByFileInfo(ctx context.Context, query []byte) ([]string, error) {
	executionID, err := c.helper.executeQuery(ctx, query)
	if err != nil {
		return nil, &ErrorExecuteQuery{
			err:      err,
			function: "query documents",
		}
	}

	keys, err := c.helper.getQueryResultKeys(ctx, *executionID)
	if err != nil {
		return nil, &ErrorGetQueryResults{
			err:         err,
			function:    "query documents",
			subfunction: "get query result keys",
		}
	}

	return keys, nil
}
