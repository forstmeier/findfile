package db

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/cheesesteakio/api/pkg/docpars"
)

type mockHelper struct {
	uploadKeyToError      string
	mockUploadObjectError error
}

func (m *mockHelper) uploadObject(ctx context.Context, body interface{}, key string) error {
	if m.uploadKeyToError == key {
		return m.mockUploadObjectError
	}

	return nil
}

func (m *mockHelper) listDocumentKeys(ctx context.Context, bucket, prefix string) ([]string, error) {
	return nil, nil // TEMP
}

func (m *mockHelper) deleteDocumentsByKeys(ctx context.Context, keys []string) error {
	return nil // TEMP
}

func (m *mockHelper) executeQuery(ctx context.Context, query []byte) (*string, *string, error) {
	return nil, nil, nil // TEMP
}

func (m *mockHelper) getQueryResultIDs(state, executionID string) (*string, *string, error) {
	return nil, nil, nil // TEMP
}

func (m *mockHelper) getQueryResultDocuments(state, executionID string) ([]docpars.Document, error) {
	return nil, nil // TEMP
}

func TestNew(t *testing.T) {
	client := New(session.New(), "database", "bucket")
	if client == nil {
		t.Error("error creating database client")
	}
}

func TestUpsertDocuments(t *testing.T) {
	accountID := "account_id"
	documentID := "document_id"
	pageID := "page_id"
	lineID := "line_id"

	tests := []struct {
		description           string
		documents             []docpars.Document
		uploadKeyToError      string
		mockUploadObjectError error
		entity                string
		error                 error
	}{
		{
			description: "error uploading document entity",
			documents: []docpars.Document{
				{
					ID:        documentID,
					AccountID: accountID,
					Filename:  "filename.jpg",
					Filepath:  "filepath",
				},
			},
			uploadKeyToError:      fmt.Sprintf("%s/documents/%s/%s.json", accountID, documentID, documentID),
			mockUploadObjectError: errors.New("mock upload object error"),
			entity:                "document",
			error:                 &ErrorUploadObject{},
		},
		{
			description: "error uploading page entity",
			documents: []docpars.Document{
				{
					ID:        documentID,
					AccountID: accountID,
					Filename:  "filename.jpg",
					Filepath:  "filepath",
					Pages: []docpars.Page{
						{
							ID:         pageID,
							PageNumber: 1,
						},
					},
				},
			},
			uploadKeyToError:      fmt.Sprintf("%s/documents/%s/pages/%s/%s.json", accountID, documentID, pageID, pageID),
			mockUploadObjectError: errors.New("mock upload object error"),
			entity:                "page",
			error:                 &ErrorUploadObject{},
		},
		{
			description: "error uploading line entity",
			documents: []docpars.Document{
				{
					ID:        documentID,
					AccountID: accountID,
					Filename:  "filename.jpg",
					Filepath:  "filepath",
					Pages: []docpars.Page{
						{
							ID:         pageID,
							PageNumber: 1,
							Lines: []docpars.Line{
								{
									ID:          lineID,
									Text:        "text",
									Coordinates: docpars.Coordinates{},
								},
							},
						},
					},
				},
			},
			uploadKeyToError:      fmt.Sprintf("%s/documents/%s/pages/%s/lines/%s/%s.json", accountID, documentID, pageID, lineID, lineID),
			mockUploadObjectError: errors.New("mock upload object error"),
			entity:                "line",
			error:                 &ErrorUploadObject{},
		},
		{
			description: "successful upsert documents invocation",
			documents: []docpars.Document{
				{
					ID:        documentID,
					AccountID: accountID,
					Filename:  "filename.jpg",
					Filepath:  "filepath",
					Pages: []docpars.Page{
						{
							ID:         pageID,
							PageNumber: 1,
							Lines: []docpars.Line{
								{
									ID:          lineID,
									Text:        "text",
									Coordinates: docpars.Coordinates{},
								},
							},
						},
					},
				},
			},
			uploadKeyToError:      "",
			mockUploadObjectError: nil,
			entity:                "",
			error:                 nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				helper: &mockHelper{
					uploadKeyToError:      test.uploadKeyToError,
					mockUploadObjectError: test.mockUploadObjectError,
				},
			}

			err := client.UpsertDocuments(context.Background(), test.documents)

			if err != nil {
				switch e := test.error.(type) {
				case *ErrorUploadObject:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
					}

					if err.(*ErrorUploadObject).entity != test.entity {
						t.Errorf("incorrect entity, received: %s, expected: %s", err.(*ErrorUploadObject).entity, test.entity)
					}
				default:
					t.Fatalf("unexpected error type: %v", err)
				}
			} else {
				if err != test.error {
					t.Errorf("incorrect error, received: %v, expected: %v", err, test.error)
				}
			}
		})
	}
}

func TestDeleteDocuments(t *testing.T) {
	tests := []struct {
		description string
	}{}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {

		})
	}
}

func TestQueryDocuments(t *testing.T) {
	tests := []struct {
		description string
	}{}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {

		})
	}
}
