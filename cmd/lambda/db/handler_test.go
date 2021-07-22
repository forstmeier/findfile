package main

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"

	"github.com/cheesesteakio/api/pkg/db"
	"github.com/cheesesteakio/api/pkg/docpars"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

type mockDocParsClient struct {
	parseAccountID string
	parseFilename  string
	parseFilepath  string
	parseOutput    *docpars.Document
	parseError     error
}

func (m *mockDocParsClient) Parse(ctx context.Context, accountID, filename, filepath string, doc []byte) (*docpars.Document, error) {
	m.parseAccountID = accountID
	m.parseFilename = filename
	m.parseFilepath = filepath

	return m.parseOutput, m.parseError
}

type mockDBClient struct {
	createOrUpdateDocumentsError error
	deleteDocumentsError         error
}

func (m *mockDBClient) CreateOrUpdateDocuments(ctx context.Context, documents []docpars.Document) error {
	return m.createOrUpdateDocumentsError
}

func (m *mockDBClient) DeleteDocuments(ctx context.Context, documentsInfo []db.DocumentInfo) error {
	return m.deleteDocumentsError
}

func (m *mockDBClient) QueryDocuments(ctx context.Context, query []byte) ([]docpars.Document, error) {
	return nil, nil
}

func Test_handler(t *testing.T) {
	tests := []struct {
		description                  string
		event                        events.S3Event
		parseOutput                  *docpars.Document
		parseError                   error
		createOrUpdateDocumentsError error
		deleteDocumentsError         error
		parseAccountID               string
		parseFilename                string
		parseFilepath                string
		error                        error
	}{
		{
			description: "unsupported event type error",
			event: events.S3Event{
				Records: []events.S3EventRecord{
					{
						EventName: "not_supported",
					},
				},
			},
			parseOutput:                  nil,
			parseError:                   nil,
			createOrUpdateDocumentsError: nil,
			deleteDocumentsError:         nil,
			parseAccountID:               "",
			parseFilename:                "",
			parseFilepath:                "",
			error:                        errorUnsupportedEvent,
		},
		{
			description: "error parsing request file",
			event: events.S3Event{
				Records: []events.S3EventRecord{
					{
						EventName: "s3:ObjectCreated:Put",
						S3: events.S3Entity{
							Bucket: events.S3Bucket{
								Name: "test_bucket",
							},
							Object: events.S3Object{
								Key: "account_id/test_file.jpg",
							},
						},
					},
				},
			},
			parseOutput:                  nil,
			parseError:                   errors.New("mock parse error"),
			createOrUpdateDocumentsError: nil,
			deleteDocumentsError:         nil,
			parseAccountID:               "account_id",
			parseFilename:                "account_id/test_file.jpg",
			parseFilepath:                "test_bucket",
			error:                        errorParseFile,
		},
		{
			description: "create or update documents method error",
			event: events.S3Event{
				Records: []events.S3EventRecord{
					{
						EventName: "s3:ObjectCreated:Put",
						S3: events.S3Entity{
							Bucket: events.S3Bucket{
								Name: "test_bucket",
							},
							Object: events.S3Object{
								Key: "account_id/test_file.jpg",
							},
						},
					},
				},
			},
			parseOutput:                  &docpars.Document{},
			parseError:                   nil,
			createOrUpdateDocumentsError: errors.New("create or update documents mock error"),
			deleteDocumentsError:         nil,
			parseAccountID:               "account_id",
			parseFilename:                "account_id/test_file.jpg",
			parseFilepath:                "test_bucket",
			error:                        errorCreateOrUpdateDocuments,
		},
		{
			description: "delete documents method error",
			event: events.S3Event{
				Records: []events.S3EventRecord{
					{
						EventName: "s3:ObjectRemoved:Delete",
						S3: events.S3Entity{
							Bucket: events.S3Bucket{
								Name: "test_bucket",
							},
							Object: events.S3Object{
								Key: "account_id/test_file.jpg",
							},
						},
					},
				},
			},
			parseOutput:                  &docpars.Document{},
			parseError:                   nil,
			createOrUpdateDocumentsError: nil,
			deleteDocumentsError:         errors.New("delete documents mock error"),
			parseAccountID:               "",
			parseFilename:                "",
			parseFilepath:                "",
			error:                        errorDeleteDocuments,
		},
		{
			description: "successful database handler invocation",
			event: events.S3Event{
				Records: []events.S3EventRecord{
					{
						EventName: "s3:ObjectCreated:Put",
						S3: events.S3Entity{
							Bucket: events.S3Bucket{
								Name: "test_bucket",
							},
							Object: events.S3Object{
								Key: "account_id/test_file_1.jpg",
							},
						},
					},
					{
						EventName: "s3:ObjectRemoved:Delete",
						S3: events.S3Entity{
							Bucket: events.S3Bucket{
								Name: "test_bucket",
							},
							Object: events.S3Object{
								Key: "account_id/test_file_2.jpg",
							},
						},
					},
				},
			},
			parseOutput:                  &docpars.Document{},
			parseError:                   nil,
			createOrUpdateDocumentsError: nil,
			deleteDocumentsError:         nil,
			parseAccountID:               "account_id",
			parseFilename:                "account_id/test_file_1.jpg",
			parseFilepath:                "test_bucket",
			error:                        nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			docparseClient := &mockDocParsClient{
				parseOutput: test.parseOutput,
				parseError:  test.parseError,
			}

			dbClient := &mockDBClient{
				createOrUpdateDocumentsError: test.createOrUpdateDocumentsError,
				deleteDocumentsError:         test.deleteDocumentsError,
			}

			handlerFunc := handler(docparseClient, dbClient)

			err := handlerFunc(context.Background(), test.event)

			if err != test.error {
				t.Errorf("incorrect error, received: %v, expected: %v", err, test.error)
			}

			if docparseClient.parseAccountID != test.parseAccountID {
				t.Errorf("incorrect parse account id, received: %s, expected: %s", docparseClient.parseAccountID, test.parseAccountID)
			}

			if docparseClient.parseFilename != test.parseFilename {
				t.Errorf("incorrect parse filename, received: %s, expected: %s", docparseClient.parseFilename, test.parseFilename)
			}

			if docparseClient.parseFilepath != test.parseFilepath {
				t.Errorf("incorrect parse filepath, received: %s, expected: %s", docparseClient.parseFilepath, test.parseFilepath)
			}
		})
	}
}
