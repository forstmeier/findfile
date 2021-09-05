package db

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/cheesesteakio/api/pkg/pars"
)

var _ Databaser = &Client{}

// Client implements the db.Databaser methods using AWS Athena.
type Client struct {
	bucketName string
	helper     helper
}

type athenaClient interface {
	StartQueryExecution(input *athena.StartQueryExecutionInput) (*athena.StartQueryExecutionOutput, error)
	GetQueryExecution(input *athena.GetQueryExecutionInput) (*athena.GetQueryExecutionOutput, error)
	GetQueryResults(input *athena.GetQueryResultsInput) (*athena.GetQueryResultsOutput, error)
}

type s3Client interface {
	ListObjectsV2(input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error)
	PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error)
	DeleteObjects(input *s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error)
}

// New generates a db.Client pointer instance with an AWS Athena client.
func New(newSession *session.Session, databaseName, bucketName string) *Client {
	return &Client{
		bucketName: bucketName,
		helper: &help{
			databaseName: databaseName,
			bucketName:   bucketName,
			athenaClient: athena.New(newSession),
			s3Client:     s3.New(newSession),
		},
	}
}

const (
	documentsPath   = "documents/%s/%s.json"
	pagesPath       = "pages/%s/%s.json"
	linesPath       = "lines/%s/%s.json"
	coordinatesPath = "coordinates/%s/%s.json"
)

// UpsertDocuments implements the db.Databaser.UpsertDocuments method.
//
// Note that JSON objects stored in S3 must be represented in a single line
// in their respective files in order for Athena to be able to query correctly.
func (c *Client) UpsertDocuments(ctx context.Context, documents []pars.Document) error {
	for _, document := range documents {
		accountID := document.AccountID
		documentID := document.ID

		documentJSON := struct {
			ID        string `json:"id"`
			AccountID string `json:"account_id"`
			Filename  string `json:"filename"`
			Filepath  string `json:"filepath"`
		}{
			ID:        documentID,
			AccountID: accountID,
			Filename:  document.Filename,
			Filepath:  document.Filepath,
		}

		documentKey := fmt.Sprintf(documentsPath, accountID, documentID)
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
				DocumentID string `json:"document_id"`
				PageNumber int64  `json:"page_number"`
			}{
				ID:         pageID,
				DocumentID: documentID,
				PageNumber: page.PageNumber,
			}

			pageKey := fmt.Sprintf(pagesPath, accountID, pageID)
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
					PageID string `json:"page_id"`
					Text   string `json:"text"`
				}{
					ID:     lineID,
					PageID: pageID,
					Text:   line.Text,
				}

				lineKey := fmt.Sprintf(linesPath, accountID, lineID)
				if err := c.helper.uploadObject(ctx, lineJSON, lineKey); err != nil {
					return &ErrorUploadObject{
						err:      err,
						function: "upsert documents",
						entity:   "line",
					}
				}

				coordinatesJSON := struct {
					ID           string  `json:"id"`
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

				coordinatesKey := fmt.Sprintf(coordinatesPath, accountID, coordinatesID)
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

// DeleteDocuments implements the db.Databaser.DeleteDocuments method.
func (c *Client) DeleteDocuments(ctx context.Context, documentsInfo []DocumentInfo) error {
	for _, documentInfo := range documentsInfo {
		deleteKeys := []string{}

		paths := []string{"documents", "pages", "lines"}
		for _, path := range paths {
			pathDeleteKeys, err := c.helper.listDocumentKeys(ctx, c.bucketName, fmt.Sprintf("%s/%s", path, documentInfo.AccountID))
			if err != nil {
				return &ErrorListDocumentKeys{
					err: err,
				}
			}
			deleteKeys = append(deleteKeys, pathDeleteKeys...)
		}

		chunkSize := 1000 // S3 max delete objects count
		for i := 0; i < len(deleteKeys); i += chunkSize {
			end := i + chunkSize
			if end > len(deleteKeys) {
				end = len(deleteKeys)
			}

			deleteKeysSubset := deleteKeys[i:end]
			if err := c.helper.deleteDocumentsByKeys(ctx, deleteKeysSubset); err != nil {
				return &ErrorDeleteDocumentsByKeys{
					err: err,
				}
			}
		}
	}

	return nil
}

// QueryDocuments implements the db.Databaser.QueryDocuments method.
//
// This implementation only returns the account id as well as the file
// name and path in the docpars.Document objects slice.
func (c *Client) QueryDocuments(ctx context.Context, query []byte) ([]pars.Document, error) {
	executionID, state, err := c.helper.executeQuery(ctx, query)
	if err != nil {
		return nil, &ErrorExecuteQuery{
			err:      err,
			function: "query documents",
		}
	}

	documents, err := c.helper.getQueryResultDocuments(*state, *executionID)
	if err != nil {
		return nil, &ErrorGetQueryResults{
			err:         err,
			function:    "query documents",
			subfunction: "get query result documents",
		}
	}

	return documents, nil
}
