package db

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/findfiledev/api/pkg/pars"
)

type mockHelper struct {
	query                             []byte
	uploadKeyToError                  string
	mockUploadObjectError             error
	mockDeleteDocumentsByKeysError    error
	mockExecuteQueryExecutionID       *string
	mockExecuteQueryError             error
	mockGetQueryResultDocumentsOutput []pars.Document
	mockGetQueryResultDocumentsError  error
	mockGetQueryResultKeysOutput      []string
	mockGetQueryResultKeysError       error
	mockAddFolderError                error
}

func (m *mockHelper) uploadObject(ctx context.Context, body interface{}, key string) error {
	if m.uploadKeyToError == key {
		return m.mockUploadObjectError
	}

	return nil
}

func (m *mockHelper) deleteDocumentsByKeys(ctx context.Context, keys []string) error {
	return m.mockDeleteDocumentsByKeysError
}

func (m *mockHelper) executeQuery(ctx context.Context, query []byte) (*string, error) {
	m.query = query

	return m.mockExecuteQueryExecutionID, m.mockExecuteQueryError
}

func (m *mockHelper) getQueryResultDocuments(ctx context.Context, executionID string) ([]pars.Document, error) {
	return m.mockGetQueryResultDocumentsOutput, m.mockGetQueryResultDocumentsError
}

func (m *mockHelper) getQueryResultKeys(ctx context.Context, executionID string) ([]string, error) {
	return m.mockGetQueryResultKeysOutput, m.mockGetQueryResultKeysError
}

func (m *mockHelper) addFolder(ctx context.Context, folder string) error {
	return m.mockAddFolderError
}

func TestNew(t *testing.T) {
	client := New(session.New(), "database", "bucket")
	if client == nil {
		t.Error("error creating database client")
	}
}

func TestSetupDatabase(t *testing.T) {
	tests := []struct {
		description        string
		mockAddFolderError error
		error              error
	}{
		{
			description:        "error adding folders to database",
			mockAddFolderError: errors.New("mock add folder error"),
			error:              &ErrorAddFolder{},
		},
		{
			description:        "successful invocation",
			mockAddFolderError: nil,
			error:              nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				helper: &mockHelper{
					mockAddFolderError: test.mockAddFolderError,
				},
			}

			err := client.SetupDatabase(context.Background())

			if err != nil {
				switch e := test.error.(type) {
				case *ErrorAddFolder:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
					}
				default:
					t.Fatalf("unexpected error type: %v", err)
				}
			} else {
				if test.error != nil {
					t.Errorf("incorrect error, received: nil, expected: %s", test.error.Error())
				}
			}
		})
	}
}

func TestUpsertDocuments(t *testing.T) {
	documentID := "document_id"
	pageID := "page_id"
	lineID := "line_id"

	tests := []struct {
		description           string
		documents             []pars.Document
		uploadKeyToError      string
		mockUploadObjectError error
		entity                string
		error                 error
	}{
		{
			description: "error uploading document entity",
			documents: []pars.Document{
				{
					ID:         documentID,
					FileKey:    "key.jpg",
					FileBucket: "bucket",
				},
			},
			uploadKeyToError:      fmt.Sprintf("documents/%s.json", documentID),
			mockUploadObjectError: errors.New("mock upload object error"),
			entity:                "document",
			error:                 &ErrorUploadObject{},
		},
		{
			description: "error uploading page entity",
			documents: []pars.Document{
				{
					ID:         documentID,
					FileKey:    "key.jpg",
					FileBucket: "bucket",
					Pages: []pars.Page{
						{
							ID:         pageID,
							PageNumber: 1,
						},
					},
				},
			},
			uploadKeyToError:      fmt.Sprintf("pages/%s.json", pageID),
			mockUploadObjectError: errors.New("mock upload object error"),
			entity:                "page",
			error:                 &ErrorUploadObject{},
		},
		{
			description: "error uploading line entity",
			documents: []pars.Document{
				{
					ID:         documentID,
					FileKey:    "key.jpg",
					FileBucket: "bucket",
					Pages: []pars.Page{
						{
							ID:         pageID,
							PageNumber: 1,
							Lines: []pars.Line{
								{
									ID:          lineID,
									Text:        "text",
									Coordinates: pars.Coordinates{},
								},
							},
						},
					},
				},
			},
			uploadKeyToError:      fmt.Sprintf("lines/%s.json", lineID),
			mockUploadObjectError: errors.New("mock upload object error"),
			entity:                "line",
			error:                 &ErrorUploadObject{},
		},
		{
			description: "successful upsert documents invocation",
			documents: []pars.Document{
				{
					ID:         documentID,
					FileKey:    "key.jpg",
					FileBucket: "bucket",
					Pages: []pars.Page{
						{
							ID:         pageID,
							PageNumber: 1,
							Lines: []pars.Line{
								{
									ID:          lineID,
									Text:        "text",
									Coordinates: pars.Coordinates{},
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
		description                    string
		mockDeleteDocumentsByKeysError error
		error                          error
	}{
		{
			description:                    "error deleting documents by keys",
			mockDeleteDocumentsByKeysError: errors.New("mock delete documents by keys error"),
			error:                          &ErrorDeleteDocumentsByKeys{},
		},
		{
			description:                    "successful delete documents invocation",
			mockDeleteDocumentsByKeysError: nil,
			error:                          nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				helper: &mockHelper{
					mockDeleteDocumentsByKeysError: test.mockDeleteDocumentsByKeysError,
				},
			}

			err := client.DeleteDocuments(context.Background(), []string{"document_id"})

			if err != nil {
				switch e := test.error.(type) {
				case *ErrorDeleteDocumentsByKeys:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
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

func TestQueryDocumentsByFQL(t *testing.T) {
	tests := []struct {
		description                       string
		mockExecuteQueryExecutionID       *string
		mockExecuteQueryError             error
		mockGetQueryResultDocumentsOutput []pars.Document
		mockGetQueryResultDocumentsError  error
		documents                         []pars.Document
		error                             error
	}{
		{
			description:                       "error executing query",
			mockExecuteQueryExecutionID:       nil,
			mockExecuteQueryError:             errors.New("mock execute query error"),
			mockGetQueryResultDocumentsOutput: nil,
			mockGetQueryResultDocumentsError:  nil,
			documents:                         nil,
			error:                             &ErrorExecuteQuery{},
		},
		{
			description:                       "error getting query result documents",
			mockExecuteQueryExecutionID:       aws.String("execution_id"),
			mockExecuteQueryError:             nil,
			mockGetQueryResultDocumentsOutput: nil,
			mockGetQueryResultDocumentsError:  errors.New("mock get query result documents error"),
			documents:                         nil,
			error:                             &ErrorGetQueryResults{},
		},
		{
			description:                 "successful query documents invocation",
			mockExecuteQueryExecutionID: aws.String("execution_id"),
			mockExecuteQueryError:       nil,
			mockGetQueryResultDocumentsOutput: []pars.Document{
				{
					Entity: "document",
				},
			},
			mockGetQueryResultDocumentsError: nil,
			documents: []pars.Document{
				{
					Entity: "document",
				},
			},
			error: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				helper: &mockHelper{
					mockExecuteQueryExecutionID:       test.mockExecuteQueryExecutionID,
					mockExecuteQueryError:             test.mockExecuteQueryError,
					mockGetQueryResultDocumentsOutput: test.mockGetQueryResultDocumentsOutput,
					mockGetQueryResultDocumentsError:  test.mockGetQueryResultDocumentsError,
				},
			}

			documents, err := client.QueryDocumentsByFQL(context.Background(), []byte("query"))

			if err != nil {
				switch e := test.error.(type) {
				case *ErrorExecuteQuery:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
					}
				case *ErrorGetQueryResults:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
					}
				default:
					t.Fatalf("unexpected error type: %v", err)
				}
			} else {
				receivedEntity := documents[0].Entity
				expectedEntity := test.documents[0].Entity
				if receivedEntity != expectedEntity {
					t.Errorf("incorrect entity, received: %s, expected: %s", receivedEntity, expectedEntity)
				}
			}
		})
	}
}

func TestQueryDocumentKeysByFileInfo(t *testing.T) {
	tests := []struct {
		description                  string
		mockExecuteQueryExecutionID  *string
		mockExecuteQueryError        error
		mockGetQueryResultKeysOutput []string
		mockGetQueryResultKeysError  error
		keys                         []string
		error                        error
	}{
		{
			description:                  "error executing query",
			mockExecuteQueryExecutionID:  nil,
			mockExecuteQueryError:        errors.New("mock execute query error"),
			mockGetQueryResultKeysOutput: nil,
			mockGetQueryResultKeysError:  nil,
			keys:                         nil,
			error:                        &ErrorExecuteQuery{},
		},
		{
			description:                  "error getting query result keys",
			mockExecuteQueryExecutionID:  aws.String("execution_id"),
			mockExecuteQueryError:        nil,
			mockGetQueryResultKeysOutput: nil,
			mockGetQueryResultKeysError:  errors.New("mock get query result keys error"),
			keys:                         nil,
			error:                        &ErrorGetQueryResults{},
		},
		{
			description:                 "successful query keys invocation",
			mockExecuteQueryExecutionID: aws.String("execution_id"),
			mockExecuteQueryError:       nil,
			mockGetQueryResultKeysOutput: []string{
				"documents/document_id.json",
				"pages/page_id.json",
				"lines/line_id.json",
				"coordinates/coordinates_id.json",
			},
			mockGetQueryResultKeysError: nil,
			keys: []string{
				"documents/document_id.json",
				"pages/page_id.json",
				"lines/line_id.json",
				"coordinates/coordinates_id.json",
			},
			error: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				helper: &mockHelper{
					mockExecuteQueryExecutionID:  test.mockExecuteQueryExecutionID,
					mockExecuteQueryError:        test.mockExecuteQueryError,
					mockGetQueryResultKeysOutput: test.mockGetQueryResultKeysOutput,
					mockGetQueryResultKeysError:  test.mockGetQueryResultKeysError,
				},
			}

			keys, err := client.QueryDocumentKeysByFileInfo(context.Background(), []byte("query"))

			if err != nil {
				switch e := test.error.(type) {
				case *ErrorExecuteQuery:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
					}
				case *ErrorGetQueryResults:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
					}
				default:
					t.Fatalf("unexpected error type: %v", err)
				}
			} else {
				if len(keys) != len(test.keys) {
					t.Errorf("incorrect keys count, received: %d, expected: %d", len(keys), len(test.keys))
				}

				for i, key := range keys {
					if key != test.keys[i] {
						t.Errorf("incorrect key, received: %s, expected: %s", key, test.keys[i])
					}
				}
			}
		})
	}
}
