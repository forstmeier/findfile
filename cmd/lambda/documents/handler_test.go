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

type mockDBClient struct {
	mockQueryDocumentsOutput []pars.Document
	mockQueryDocumentsError  error
}

func (m *mockDBClient) SetupDatabase(ctx context.Context) error {
	return nil
}

func (m *mockDBClient) UpsertDocuments(ctx context.Context, documents []pars.Document) error {
	return nil
}

func (m *mockDBClient) DeleteDocumentsByIDs(ctx context.Context, documentIDs []string) error {
	return nil
}

func (m *mockDBClient) DeleteDocumentsByBuckets(ctx context.Context, documentIDs []string) error {
	return nil
}

func (m *mockDBClient) QueryDocuments(ctx context.Context, query db.Query) ([]pars.Document, error) {
	return m.mockQueryDocumentsOutput, m.mockQueryDocumentsError
}

func Test_handler(t *testing.T) {
	tests := []struct {
		description              string
		request                  events.APIGatewayProxyRequest
		mockQueryDocumentsOutput []pars.Document
		mockQueryDocumentsError  error
		statusCode               int
		body                     string
	}{

		{
			description:              "no security header received",
			request:                  events.APIGatewayProxyRequest{},
			mockQueryDocumentsOutput: nil,
			mockQueryDocumentsError:  nil,
			statusCode:               400,
			body:                     `{"error": "security key header not provided"}`,
		},
		{
			description: "incorrect security header received",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"http-security-header": "incorrect-value",
				},
			},
			mockQueryDocumentsOutput: nil,
			mockQueryDocumentsError:  nil,
			statusCode:               400,
			body:                     `{"error": "security key "incorrect-value" incorrect"}`,
		},
		{
			description: "unmarshal request body error",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"http-security-header": "http-security-header-value",
				},
				Body: "invalid-json",
			},
			mockQueryDocumentsOutput: nil,
			mockQueryDocumentsError:  nil,
			statusCode:               400,
			body:                     `{"error": "invalid character 'i' looking for beginning of value"}`,
		},
		{
			description: "query documents error",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"http-security-header": "http-security-header-value",
				},
				Body: `{"text": "lookup text"}`,
			},
			mockQueryDocumentsOutput: nil,
			mockQueryDocumentsError:  errors.New("mock query documents error"),
			statusCode:               500,
			body:                     `{"error": "mock query documents error"}`,
		},
		{
			description: "successful invocation",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"http-security-header": "http-security-header-value",
				},
				Body: `{"text": "lookup text"}`,
			},
			mockQueryDocumentsOutput: []pars.Document{
				{
					FileBucket: "bucket",
					FileKey:    "key.jpeg",
				},
			},
			mockQueryDocumentsError: nil,
			statusCode:              200,
			body:                    `{"message":"success","file_paths":["bucket/key.jpeg"]}`,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			dbClient := &mockDBClient{
				mockQueryDocumentsOutput: test.mockQueryDocumentsOutput,
				mockQueryDocumentsError:  test.mockQueryDocumentsError,
			}

			handlerFunc := handler(dbClient, "http-security-header", "http-security-header-value")

			response, _ := handlerFunc(context.Background(), test.request)

			if response.StatusCode != test.statusCode {
				t.Errorf("incorrect status code, received: %d, expected: %d", response.StatusCode, test.statusCode)
			}

			if response.Body != test.body {
				t.Errorf("incorrect body, received: %q, expected: %q", test.body, response.Body)
			}
		})
	}
}
