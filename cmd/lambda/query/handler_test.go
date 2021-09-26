package main

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"

	"github.com/findfiledev/api/pkg/pars"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

type mockFQLClient struct {
	mockConvertFQLOutput []byte
	mockConvertFQLError  error
}

func (m *mockFQLClient) ConvertFQL(ctx context.Context, fqlQuery []byte) ([]byte, error) {
	return m.mockConvertFQLOutput, m.mockConvertFQLError
}

type mockDBClient struct {
	mockDeleteDocumentsError      error
	mockQueryDocumentsByFQLOutput []pars.Document
	mockQueryDocumentsByFQLError  error
}

func (m *mockDBClient) SetupDatabase(ctx context.Context) error {
	return nil
}

func (m *mockDBClient) UpsertDocuments(ctx context.Context, documents []pars.Document) error {
	return nil
}

func (m *mockDBClient) DeleteDocuments(ctx context.Context, documentIDs []string) error {
	return m.mockDeleteDocumentsError
}

func (m *mockDBClient) QueryDocumentsByFQL(ctx context.Context, query []byte) ([]pars.Document, error) {
	return m.mockQueryDocumentsByFQLOutput, m.mockQueryDocumentsByFQLError
}

func (m *mockDBClient) QueryDocumentKeysByFileInfo(ctx context.Context, query []byte) ([]string, error) {
	return nil, nil
}

func Test_handler(t *testing.T) {
	tests := []struct {
		description                   string
		request                       events.APIGatewayProxyRequest
		mockConvertFQLOutput          []byte
		mockConvertFQLError           error
		mockQueryDocumentsByFQLOutput []pars.Document
		mockQueryDocumentsByFQLError  error
		statusCode                    int
		body                          string
	}{
		{
			description:                   "no security key header in request",
			request:                       events.APIGatewayProxyRequest{},
			mockConvertFQLOutput:          nil,
			mockConvertFQLError:           nil,
			mockQueryDocumentsByFQLOutput: nil,
			mockQueryDocumentsByFQLError:  nil,
			statusCode:                    http.StatusBadRequest,
			body:                          `{"error": "security key header not provided"}`,
		},
		{
			description: "incorrect security key header value in request",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"security_header": "not_key",
				},
			},
			mockConvertFQLOutput:          nil,
			mockConvertFQLError:           nil,
			mockQueryDocumentsByFQLOutput: nil,
			mockQueryDocumentsByFQLError:  nil,
			statusCode:                    http.StatusBadRequest,
			body:                          `{"error": "security key value incorrect"}`,
		},
		{
			description: "unsupported http method",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"security_header": "security_key",
				},
				HTTPMethod: http.MethodGet,
			},
			mockConvertFQLOutput:          nil,
			mockConvertFQLError:           nil,
			mockQueryDocumentsByFQLOutput: nil,
			mockQueryDocumentsByFQLError:  nil,
			statusCode:                    http.StatusBadRequest,
			body:                          `{"error": "http method [GET] not supported"}`,
		},
		{
			description: "fql client error converting fql query",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"security_header": "security_key",
				},
				HTTPMethod: http.MethodPost,
				Body:       `{"test": "query"}`,
			},
			mockConvertFQLOutput:          nil,
			mockConvertFQLError:           errors.New("mock convert fql error"),
			mockQueryDocumentsByFQLOutput: nil,
			mockQueryDocumentsByFQLError:  nil,
			statusCode:                    http.StatusInternalServerError,
			body:                          `{"error": "error converting fql to query"}`,
		},
		{
			description: "error querying documents in database",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"security_header": "security_key",
				},
				HTTPMethod: http.MethodPost,
				Body:       `{"test": "query"}`,
			},
			mockConvertFQLOutput:          []byte("test_query"),
			mockConvertFQLError:           nil,
			mockQueryDocumentsByFQLOutput: nil,
			mockQueryDocumentsByFQLError:  errors.New("mock query documents error"),
			statusCode:                    http.StatusInternalServerError,
			body:                          `{"error": "error running query"}`,
		},
		{
			description: "successful handler invocation",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"security_header": "security_key",
				},
				HTTPMethod: http.MethodPost,
				Body:       `{"test": "query"}`,
			},
			mockConvertFQLOutput: []byte("test_query"),
			mockConvertFQLError:  nil,
			mockQueryDocumentsByFQLOutput: []pars.Document{
				{
					FileKey:    "key.jpg",
					FileBucket: "bucket",
				},
			},
			mockQueryDocumentsByFQLError: nil,
			statusCode:                   http.StatusOK,
			body:                         `{"message":"success","data":{"bucket":["key.jpg"]}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			fqlClient := &mockFQLClient{
				mockConvertFQLOutput: test.mockConvertFQLOutput,
				mockConvertFQLError:  test.mockConvertFQLError,
			}

			dbClient := &mockDBClient{
				mockQueryDocumentsByFQLOutput: test.mockQueryDocumentsByFQLOutput,
				mockQueryDocumentsByFQLError:  test.mockQueryDocumentsByFQLError,
			}

			handlerFunc := handler(fqlClient, dbClient, "security_header", "security_key")

			response, _ := handlerFunc(context.Background(), test.request)

			if response.StatusCode != test.statusCode {
				t.Errorf("incorrect status code, received: %d, expected: %d", response.StatusCode, test.statusCode)
			}

			if response.Body != test.body {
				t.Errorf("incorrect body, received: %s, expected: %s", response.Body, test.body)
			}
		})
	}
}
