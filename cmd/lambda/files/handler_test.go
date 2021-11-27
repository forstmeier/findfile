package main

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"

	"github.com/forstmeier/findfile/pkg/db"
	"github.com/forstmeier/findfile/pkg/pars"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

type mockParsClient struct {
	mockParseOutput *pars.Document
	mockParseError  error
}

func (m *mockParsClient) Parse(ctx context.Context, fileBucket, fileKey string) (*pars.Document, error) {
	return m.mockParseOutput, m.mockParseError
}

type mockDBClient struct {
	mockUpsertDocumentsError error
	mockDeleteDocumentsError error
}

func (m *mockDBClient) SetupDatabase(ctx context.Context) error {
	return nil
}

func (m *mockDBClient) UpsertDocuments(ctx context.Context, documents []pars.Document) error {
	return m.mockUpsertDocumentsError
}

func (m *mockDBClient) DeleteDocumentsByIDs(ctx context.Context, documentIDs []string) error {
	return m.mockDeleteDocumentsError
}

func (m *mockDBClient) DeleteDocumentsByBuckets(ctx context.Context, documentIDs []string) error {
	return m.mockDeleteDocumentsError
}

func (m *mockDBClient) QueryDocuments(ctx context.Context, query db.Query) ([]pars.Document, error) {
	return nil, nil
}

func Test_handler(t *testing.T) {
	parseError := errors.New("mock parse error")
	upsertError := errors.New("mock upsert error")
	deleteError := errors.New("mock delete error")

	tests := []struct {
		description              string
		event                    events.CloudWatchEvent
		mockParseOutput          *pars.Document
		mockParseError           error
		mockUpsertDocumentsError error
		mockDeleteDocumentsError error
		error                    error
	}{
		{
			description: "parse file error",
			event: events.CloudWatchEvent{
				Detail: []byte(`{ "eventName": "PutObject", "requestParameters": { "bucketName": "bucket", "key": "key.jpeg" } }`),
			},
			mockParseOutput:          nil,
			mockParseError:           parseError,
			mockUpsertDocumentsError: nil,
			mockDeleteDocumentsError: nil,
			error:                    parseError,
		},
		{
			description: "upsert document error",
			event: events.CloudWatchEvent{
				Detail: []byte(`{ "eventName": "PutObject", "requestParameters": { "bucketName": "bucket", "key": "key.jpeg" } }`),
			},
			mockParseOutput:          &pars.Document{},
			mockParseError:           nil,
			mockUpsertDocumentsError: upsertError,
			mockDeleteDocumentsError: nil,
			error:                    upsertError,
		},
		{
			description: "delete document error",
			event: events.CloudWatchEvent{
				Detail: []byte(`{ "eventName": "DeleteObjects", "requestParameters": { "bucketName": "bucket", "key": "key.jpeg" } }`),
			},
			mockParseOutput:          nil,
			mockParseError:           nil,
			mockUpsertDocumentsError: nil,
			mockDeleteDocumentsError: deleteError,
			error:                    deleteError,
		},
		{
			description: "successful invocation",
			event: events.CloudWatchEvent{
				Detail: []byte(`{ "eventName": "DeleteObjects", "requestParameters": { "bucketName": "bucket", "key": "key.jpeg" } }`),
			},
			mockParseOutput:          nil,
			mockParseError:           nil,
			mockUpsertDocumentsError: nil,
			mockDeleteDocumentsError: nil,
			error:                    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			parsClient := &mockParsClient{
				mockParseOutput: test.mockParseOutput,
				mockParseError:  test.mockParseError,
			}

			dbClient := &mockDBClient{
				mockUpsertDocumentsError: test.mockUpsertDocumentsError,
				mockDeleteDocumentsError: test.mockDeleteDocumentsError,
			}

			handlerFunc := handler(parsClient, dbClient)

			err := handlerFunc(context.Background(), test.event)

			if err != nil {
				if !errors.As(err, &test.error) {
					t.Errorf("incorrect error, received: %v, expected: %v", err, test.error)
				}
			}
		})
	}
}
