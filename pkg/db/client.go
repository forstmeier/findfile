package db

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/cheesesteakio/api/pkg/docpars"
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
	documentsPath = "%s/documents/%s/%s.json"
	pagesPath     = "%s/documents/%s/pages/%s/%s.json"
	linesPath     = "%s/documents/%s/pages/%s/lines/%s/%s.json"
)

// UpsertDocuments implements the db.Databaser.UpsertDocuments method.
func (c *Client) UpsertDocuments(ctx context.Context, documents []docpars.Document) error {
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

		key := fmt.Sprintf(documentsPath, accountID, documentID, documentID)
		if err := c.helper.uploadObject(ctx, documentJSON, key); err != nil {
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
				PageNumber int64  `json:"page_number"`
			}{
				ID:         pageID,
				PageNumber: page.PageNumber,
			}

			key := fmt.Sprintf(pagesPath, accountID, documentID, pageID, pageID)
			if err := c.helper.uploadObject(ctx, pageJSON, key); err != nil {
				return &ErrorUploadObject{
					err:      err,
					function: "upsert documents",
					entity:   "page",
				}
			}

			for _, line := range page.Lines {
				lineID := line.ID

				lineJSON := struct {
					ID          string              `json:"id"`
					Text        string              `json:"text"`
					Coordinates docpars.Coordinates `json:"coordinates"`
				}{
					ID:          lineID,
					Text:        line.Text,
					Coordinates: line.Coordinates,
				}

				key := fmt.Sprintf(linesPath, accountID, documentID, pageID, lineID, lineID)
				if err := c.helper.uploadObject(ctx, lineJSON, key); err != nil {
					return &ErrorUploadObject{
						err:      err,
						function: "upsert documents",
						entity:   "line",
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

		_ = documentInfo // NOTE: use in query construction

		query := []byte{}

		executionID, state, err := c.helper.executeQuery(ctx, query)
		if err != nil {
			return &ErrorExecuteQuery{
				err:      err,
				function: "delete documents",
			}
		}

		accountID, documentID, err := c.helper.getQueryResultIDs(*state, *executionID)
		if err != nil {
			return &ErrorGetQueryResults{
				err:         err,
				function:    "delete documents",
				subfunction: "get query result ids",
			}
		}

		deleteKeys, err := c.helper.listDocumentKeys(ctx, c.bucketName, fmt.Sprintf("%s/documents/%s", *accountID, *documentID))
		if err != nil {
			return &ErrorListDocumentKeys{
				err: err,
			}
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
func (c *Client) QueryDocuments(ctx context.Context, query []byte) ([]docpars.Document, error) {
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
