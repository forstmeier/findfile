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

type mockEvtClient struct {
	mockAddBucketListenersError    error
	mockRemoveBucketListenersError error
}

func (m *mockEvtClient) AddBucketListeners(ctx context.Context, buckets []string) error {
	return m.mockAddBucketListenersError
}

func (m *mockEvtClient) RemoveBucketListeners(ctx context.Context, buckets []string) error {
	return m.mockRemoveBucketListenersError
}

type mockFSClient struct {
	mockListFilesOutput []string
	mockListFilesError  error
}

func (m *mockFSClient) ListFiles(ctx context.Context, bucket string) ([]string, error) {
	return m.mockListFilesOutput, m.mockListFilesError
}

type mockParsClient struct {
	mockParseOutput *pars.Document
	mockParseError  error
}

func (m *mockParsClient) Parse(ctx context.Context, fileBucket, fileKey string) (*pars.Document, error) {
	return m.mockParseOutput, m.mockParseError
}

type mockDBClient struct {
	mockUpsertDocumentsError          error
	mockDeleteDocumentsByBucketsError error
}

func (m *mockDBClient) SetupDatabase(ctx context.Context) error {
	return nil
}

func (m *mockDBClient) UpsertDocuments(ctx context.Context, documents []pars.Document) error {
	return m.mockUpsertDocumentsError
}

func (m *mockDBClient) DeleteDocumentsByIDs(ctx context.Context, documentIDs []string) error {
	return nil
}

func (m *mockDBClient) DeleteDocumentsByBuckets(ctx context.Context, buckets []string) error {
	return m.mockDeleteDocumentsByBucketsError
}

func (m *mockDBClient) QueryDocuments(ctx context.Context, query db.Query) ([]pars.Document, error) {
	return nil, nil
}

func Test_handler(t *testing.T) {
	tests := []struct {
		description                       string
		request                           events.APIGatewayProxyRequest
		mockAddBucketListenersError       error
		mockRemoveBucketListenersError    error
		mockListFilesOutput               []string
		mockListFilesError                error
		mockParseOutput                   *pars.Document
		mockParseError                    error
		mockUpsertDocumentsError          error
		mockDeleteDocumentsByBucketsError error
		statusCode                        int
		body                              string
	}{
		{
			description:                       "no security header received",
			request:                           events.APIGatewayProxyRequest{},
			mockAddBucketListenersError:       nil,
			mockRemoveBucketListenersError:    nil,
			mockListFilesOutput:               nil,
			mockListFilesError:                nil,
			mockParseOutput:                   nil,
			mockParseError:                    nil,
			mockUpsertDocumentsError:          nil,
			mockDeleteDocumentsByBucketsError: nil,
			statusCode:                        400,
			body:                              `{"error": "security key header not provided"}`,
		},
		{
			description: "incorrect security header received",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"http-security-header": "incorrect-value",
				},
			},
			mockAddBucketListenersError:       nil,
			mockRemoveBucketListenersError:    nil,
			mockListFilesOutput:               nil,
			mockListFilesError:                nil,
			mockParseOutput:                   nil,
			mockParseError:                    nil,
			mockUpsertDocumentsError:          nil,
			mockDeleteDocumentsByBucketsError: nil,
			statusCode:                        400,
			body:                              `{"error": "security key "incorrect-value" incorrect"}`,
		},
		{
			description: "unmarshal request body error",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"http-security-header": "http-security-header-value",
				},
				Body: "invalid-json",
			},
			mockAddBucketListenersError:       nil,
			mockRemoveBucketListenersError:    nil,
			mockListFilesOutput:               nil,
			mockListFilesError:                nil,
			mockParseOutput:                   nil,
			mockParseError:                    nil,
			mockUpsertDocumentsError:          nil,
			mockDeleteDocumentsByBucketsError: nil,
			statusCode:                        400,
			body:                              `{"error": "invalid character 'i' looking for beginning of value"}`,
		},
		{
			description: "add bucket listeners error",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"http-security-header": "http-security-header-value",
				},
				Body: `{"add": ["bucket"]}`,
			},
			mockAddBucketListenersError:       errors.New("mock add bucket listeners error"),
			mockRemoveBucketListenersError:    nil,
			mockListFilesOutput:               nil,
			mockListFilesError:                nil,
			mockParseOutput:                   nil,
			mockParseError:                    nil,
			mockUpsertDocumentsError:          nil,
			mockDeleteDocumentsByBucketsError: nil,
			statusCode:                        500,
			body:                              `{"error": "mock add bucket listeners error"}`,
		},
		{
			description: "list files error",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"http-security-header": "http-security-header-value",
				},
				Body: `{"add": ["bucket"]}`,
			},
			mockAddBucketListenersError:       nil,
			mockRemoveBucketListenersError:    nil,
			mockListFilesOutput:               nil,
			mockListFilesError:                errors.New("mock list files error"),
			mockParseOutput:                   nil,
			mockParseError:                    nil,
			mockUpsertDocumentsError:          nil,
			mockDeleteDocumentsByBucketsError: nil,
			statusCode:                        500,
			body:                              `{"error": "mock list files error"}`,
		},
		{
			description: "parse error",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"http-security-header": "http-security-header-value",
				},
				Body: `{"add": ["bucket"]}`,
			},
			mockAddBucketListenersError:       nil,
			mockRemoveBucketListenersError:    nil,
			mockListFilesOutput:               []string{"key.jpeg"},
			mockListFilesError:                nil,
			mockParseOutput:                   nil,
			mockParseError:                    errors.New("mock parse error"),
			mockUpsertDocumentsError:          nil,
			mockDeleteDocumentsByBucketsError: nil,
			statusCode:                        500,
			body:                              `{"error": "mock parse error"}`,
		},
		{
			description: "upsert documents error",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"http-security-header": "http-security-header-value",
				},
				Body: `{"add": ["bucket"]}`,
			},
			mockAddBucketListenersError:       nil,
			mockRemoveBucketListenersError:    nil,
			mockListFilesOutput:               []string{"key.jpeg"},
			mockListFilesError:                nil,
			mockParseOutput:                   &pars.Document{},
			mockParseError:                    nil,
			mockUpsertDocumentsError:          errors.New("mock upsert documents error"),
			mockDeleteDocumentsByBucketsError: nil,
			statusCode:                        500,
			body:                              `{"error": "mock upsert documents error"}`,
		},
		{
			description: "remove bucket listeners error",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"http-security-header": "http-security-header-value",
				},
				Body: `{"remove": ["bucket"]}`,
			},
			mockAddBucketListenersError:       nil,
			mockRemoveBucketListenersError:    errors.New("mock remove bucket listeners error"),
			mockListFilesOutput:               nil,
			mockListFilesError:                nil,
			mockParseOutput:                   nil,
			mockParseError:                    nil,
			mockUpsertDocumentsError:          nil,
			mockDeleteDocumentsByBucketsError: nil,
			statusCode:                        500,
			body:                              `{"error": "mock remove bucket listeners error"}`,
		},
		{
			description: "delete documents by buckets error",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"http-security-header": "http-security-header-value",
				},
				Body: `{"remove": ["bucket"]}`,
			},
			mockAddBucketListenersError:       nil,
			mockRemoveBucketListenersError:    nil,
			mockListFilesOutput:               nil,
			mockListFilesError:                nil,
			mockParseOutput:                   nil,
			mockParseError:                    nil,
			mockUpsertDocumentsError:          nil,
			mockDeleteDocumentsByBucketsError: errors.New("mock delete documents by buckets error"),
			statusCode:                        500,
			body:                              `{"error": "mock delete documents by buckets error"}`,
		},
		{
			description: "successful invocation",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"http-security-header": "http-security-header-value",
				},
				Body: `{"add":["add_bucket"], "remove": ["remove_bucket"]}`,
			},
			mockAddBucketListenersError:       nil,
			mockRemoveBucketListenersError:    nil,
			mockListFilesOutput:               nil,
			mockListFilesError:                nil,
			mockParseOutput:                   nil,
			mockParseError:                    nil,
			mockUpsertDocumentsError:          nil,
			mockDeleteDocumentsByBucketsError: nil,
			statusCode:                        200,
			body:                              `{"message": "success", "buckets_added": 1, "buckets_removed": 1}`,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			evtClient := &mockEvtClient{
				mockAddBucketListenersError:    test.mockAddBucketListenersError,
				mockRemoveBucketListenersError: test.mockRemoveBucketListenersError,
			}

			fsClient := &mockFSClient{
				mockListFilesOutput: test.mockListFilesOutput,
				mockListFilesError:  test.mockListFilesError,
			}

			parsClient := &mockParsClient{
				mockParseOutput: test.mockParseOutput,
				mockParseError:  test.mockParseError,
			}

			dbClient := &mockDBClient{
				mockUpsertDocumentsError:          test.mockUpsertDocumentsError,
				mockDeleteDocumentsByBucketsError: test.mockDeleteDocumentsByBucketsError,
			}

			handlerFunc := handler(
				evtClient,
				fsClient,
				parsClient,
				dbClient,
				"http-security-header",
				"http-security-header-value",
			)

			response, _ := handlerFunc(context.Background(), test.request)

			if response.StatusCode != test.statusCode {
				t.Errorf("incorrect status code, received: %d, expected: %d", response.StatusCode, test.statusCode)
			}

			if response.Body != test.body {
				t.Errorf("incorrect body, received: %q, expected: %q", response.Body, test.body)
			}
		})
	}
}
