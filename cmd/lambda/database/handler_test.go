package main

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"

	"github.com/findfiledev/api/pkg/pars"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

type mockParsClient struct {
	mockParseFileKey    string
	mockParseFileBucket string
	mockParseOutput     *pars.Document
	mockParseError      error
}

func (m *mockParsClient) Parse(ctx context.Context, fileKey, fileBucket string) (*pars.Document, error) {
	m.mockParseFileKey = fileKey
	m.mockParseFileBucket = fileBucket

	return m.mockParseOutput, m.mockParseError
}

type mockDBClient struct {
	mockUpsertDocumentsError              error
	mockDeleteDocumentsError              error
	mockQueryDocumentKeysByFileInfoOutput []string
	mockQueryDocumentKeysByFileInfoError  error
}

func (m *mockDBClient) SetupDatabase(ctx context.Context) error {
	return nil
}

func (m *mockDBClient) UpsertDocuments(ctx context.Context, documents []pars.Document) error {
	return m.mockUpsertDocumentsError
}

func (m *mockDBClient) DeleteDocuments(ctx context.Context, documentsInfo []string) error {
	return m.mockDeleteDocumentsError
}

func (m *mockDBClient) QueryDocumentsByFQL(ctx context.Context, query []byte) ([]pars.Document, error) {
	return nil, nil
}

func (m *mockDBClient) QueryDocumentKeysByFileInfo(ctx context.Context, query []byte) ([]string, error) {
	return m.mockQueryDocumentKeysByFileInfoOutput, m.mockQueryDocumentKeysByFileInfoError
}

func Test_handler(t *testing.T) {
	tests := []struct {
		description                           string
		s3Event                               events.S3Event
		mockParseOutput                       *pars.Document
		mockParseError                        error
		mockQueryDocumentKeysByFileInfoOutput []string
		mockQueryDocumentKeysByFileInfoError  error
		mockUpsertDocumentsError              error
		mockDeleteDocumentsError              error
		parseFileKey                          string
		parseFileBucket                       string
		error                                 error
	}{
		{
			description: "unsupported event type error",
			s3Event: events.S3Event{
				Records: []events.S3EventRecord{
					{
						EventName: "not_supported",
						S3: events.S3Entity{
							Object: events.S3Object{
								Key: "key.jpg",
							},
						},
					},
				},
			},
			mockParseOutput:                       nil,
			mockParseError:                        nil,
			mockQueryDocumentKeysByFileInfoOutput: nil,
			mockQueryDocumentKeysByFileInfoError:  nil,
			mockUpsertDocumentsError:              nil,
			mockDeleteDocumentsError:              nil,
			parseFileKey:                          "",
			parseFileBucket:                       "",
			error:                                 errorUnsupportedEvent,
		},
		{
			description: "error parsing request file",
			s3Event: events.S3Event{
				Records: []events.S3EventRecord{
					{
						EventName: "ObjectCreated:Put",
						S3: events.S3Entity{
							Bucket: events.S3Bucket{
								Name: "bucket",
							},
							Object: events.S3Object{
								Key: "key.jpg",
							},
						},
					},
				},
			},
			mockParseOutput:                       nil,
			mockParseError:                        errors.New("mock parse error"),
			mockQueryDocumentKeysByFileInfoOutput: nil,
			mockQueryDocumentKeysByFileInfoError:  nil,
			mockUpsertDocumentsError:              nil,
			mockDeleteDocumentsError:              nil,
			parseFileKey:                          "key.jpg",
			parseFileBucket:                       "bucket",
			error:                                 errorParseFile,
		},
		{
			description: "error query document keys by file info",
			s3Event: events.S3Event{
				Records: []events.S3EventRecord{
					{
						EventName: "ObjectRemoved:Delete",
						S3: events.S3Entity{
							Bucket: events.S3Bucket{
								Name: "bucket",
							},
							Object: events.S3Object{
								Key: "key.jpg",
							},
						},
					},
				},
			},
			mockParseOutput:                       nil,
			mockParseError:                        nil,
			mockQueryDocumentKeysByFileInfoOutput: nil,
			mockQueryDocumentKeysByFileInfoError:  errors.New("mock query documents by file info error"),
			mockUpsertDocumentsError:              nil,
			mockDeleteDocumentsError:              nil,
			parseFileKey:                          "",
			parseFileBucket:                       "",
			error:                                 errorQueryDocuments,
		},
		{
			description: "upsert documents method error",
			s3Event: events.S3Event{
				Records: []events.S3EventRecord{
					{
						EventName: "ObjectCreated:Put",
						S3: events.S3Entity{
							Bucket: events.S3Bucket{
								Name: "bucket",
							},
							Object: events.S3Object{
								Key: "key.jpg",
							},
						},
					},
				},
			},
			mockParseOutput:                       &pars.Document{},
			mockParseError:                        nil,
			mockQueryDocumentKeysByFileInfoOutput: nil,
			mockQueryDocumentKeysByFileInfoError:  nil,
			mockUpsertDocumentsError:              errors.New("upsert documents mock error"),
			mockDeleteDocumentsError:              nil,
			parseFileKey:                          "key.jpg",
			parseFileBucket:                       "bucket",
			error:                                 errorUpsertDocuments,
		},
		{
			description: "delete documents method error",
			s3Event: events.S3Event{
				Records: []events.S3EventRecord{
					{
						EventName: "ObjectRemoved:Delete",
						S3: events.S3Entity{
							Bucket: events.S3Bucket{
								Name: "bucket",
							},
							Object: events.S3Object{
								Key: "key.jpg",
							},
						},
					},
				},
			},
			mockParseOutput:                       &pars.Document{},
			mockParseError:                        nil,
			mockQueryDocumentKeysByFileInfoOutput: nil,
			mockQueryDocumentKeysByFileInfoError:  nil,
			mockUpsertDocumentsError:              nil,
			mockDeleteDocumentsError:              errors.New("delete documents mock error"),
			parseFileKey:                          "",
			parseFileBucket:                       "",
			error:                                 errorDeleteDocuments,
		},
		{
			description: "successful database handler invocation",
			s3Event: events.S3Event{
				Records: []events.S3EventRecord{
					{
						EventName: "ObjectCreated:Put",
						S3: events.S3Entity{
							Bucket: events.S3Bucket{
								Name: "bucket",
							},
							Object: events.S3Object{
								Key: "key1.jpg",
							},
						},
					},
					{
						EventName: "ObjectRemoved:Delete",
						S3: events.S3Entity{
							Bucket: events.S3Bucket{
								Name: "bucket",
							},
							Object: events.S3Object{
								Key: "key2.jpg",
							},
						},
					},
				},
			},
			mockParseOutput:                       &pars.Document{},
			mockParseError:                        nil,
			mockQueryDocumentKeysByFileInfoOutput: nil,
			mockQueryDocumentKeysByFileInfoError:  nil,
			mockUpsertDocumentsError:              nil,
			mockDeleteDocumentsError:              nil,
			parseFileKey:                          "key1.jpg",
			parseFileBucket:                       "bucket",
			error:                                 nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			parsClient := &mockParsClient{
				mockParseOutput: test.mockParseOutput,
				mockParseError:  test.mockParseError,
			}

			dbClient := &mockDBClient{
				mockQueryDocumentKeysByFileInfoOutput: test.mockQueryDocumentKeysByFileInfoOutput,
				mockQueryDocumentKeysByFileInfoError:  test.mockQueryDocumentKeysByFileInfoError,
				mockUpsertDocumentsError:              test.mockUpsertDocumentsError,
				mockDeleteDocumentsError:              test.mockDeleteDocumentsError,
			}

			handlerFunc := handler(parsClient, dbClient)

			err := handlerFunc(context.Background(), test.s3Event)

			if err != test.error {
				t.Errorf("incorrect error, received: %v, expected: %v", err, test.error)
			}

			if parsClient.mockParseFileKey != test.parseFileKey {
				t.Errorf("incorrect parse file key, received: %s, expected: %s", parsClient.mockParseFileKey, test.parseFileKey)
			}

			if parsClient.mockParseFileBucket != test.parseFileBucket {
				t.Errorf("incorrect parse file bucket, received: %s, expected: %s", parsClient.mockParseFileBucket, test.parseFileBucket)
			}
		})
	}
}
