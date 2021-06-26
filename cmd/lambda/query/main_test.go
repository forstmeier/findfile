package main

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"

	"github.com/cheesesteakio/api/pkg/acct"
	"github.com/cheesesteakio/api/pkg/db"
	"github.com/cheesesteakio/api/pkg/docpars"
	"github.com/cheesesteakio/api/pkg/fs"
)

type mockAcctClient struct {
	mockReadAccountOutput *acct.Account
	mockReadAccountError  error
}

func (m *mockAcctClient) CreateAccount(ctx context.Context, accountID string) error {
	return nil
}

func (m *mockAcctClient) ReadAccount(ctx context.Context, accountID string) (*acct.Account, error) {
	return m.mockReadAccountOutput, m.mockReadAccountError
}

func (m *mockAcctClient) UpdateAccount(ctx context.Context, accountID string, values map[string]string) error {
	return nil
}

func (m *mockAcctClient) DeleteAccount(ctx context.Context, accountID string) error {
	return nil
}

type mockCSQLClient struct {
	mockConvertCSQLOutput []byte
	mockConvertCSQLError  error
}

func (m *mockCSQLClient) ConvertCSQL(ctx context.Context, accountID string, csqlQuery map[string]interface{}) ([]byte, error) {
	return m.mockConvertCSQLOutput, m.mockConvertCSQLError
}

type mockDBClient struct {
	mockQueryDocumentsOutput []docpars.Document
	mockQueryDocumentsError  error
}

func (m *mockDBClient) CreateOrUpdateDocuments(ctx context.Context, documents []docpars.Document) error {
	return nil
}

func (m *mockDBClient) DeleteDocuments(ctx context.Context, documentsInfo []db.DocumentInfo) error {
	return nil
}

func (m *mockDBClient) QueryDocuments(ctx context.Context, query []byte) ([]docpars.Document, error) {
	return m.mockQueryDocumentsOutput, m.mockQueryDocumentsError
}

type mockFSClient struct {
	mockGeneratePresignedURLOutput string
	mockGeneratePresignedURLError  error
}

func (m *mockFSClient) GenerateUploadURL(ctx context.Context, accountID string, fileInfo fs.FileInfo) (string, error) {
	return "", nil
}

func (m *mockFSClient) GenerateDownloadURL(ctx context.Context, accountID string, fileInfo fs.FileInfo) (string, error) {
	return m.mockGeneratePresignedURLOutput, m.mockGeneratePresignedURLError
}

func (m *mockFSClient) DeleteFiles(ctx context.Context, accountID string, filesInfo []fs.FileInfo) error {
	return nil
}

func Test_handler(t *testing.T) {
	tests := []struct {
		description                    string
		request                        events.APIGatewayProxyRequest
		mockReadAccountOutput          *acct.Account
		mockReadAccountError           error
		mockConvertCSQLOutput          []byte
		mockConvertCSQLError           error
		mockQueryDocumentsOutput       []docpars.Document
		mockQueryDocumentsError        error
		mockGeneratePresignedURLOutput string
		mockGeneratePresignedURLError  error
		statusCode                     int
		body                           string
	}{
		{
			description:                    "no account id in request",
			request:                        events.APIGatewayProxyRequest{},
			mockReadAccountOutput:          nil,
			mockReadAccountError:           nil,
			mockConvertCSQLOutput:          nil,
			mockConvertCSQLError:           nil,
			mockQueryDocumentsOutput:       nil,
			mockQueryDocumentsError:        nil,
			mockGeneratePresignedURLOutput: "",
			mockGeneratePresignedURLError:  nil,
			statusCode:                     http.StatusBadRequest,
			body:                           `{"error": "account id not provided"}`,
		},
		{
			description: "unsupported http method",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
				HTTPMethod: http.MethodGet,
			},
			mockReadAccountOutput:          nil,
			mockReadAccountError:           nil,
			mockConvertCSQLOutput:          nil,
			mockConvertCSQLError:           nil,
			mockQueryDocumentsOutput:       nil,
			mockQueryDocumentsError:        nil,
			mockGeneratePresignedURLOutput: "",
			mockGeneratePresignedURLError:  nil,
			statusCode:                     http.StatusBadRequest,
			body:                           `{"error": "http method [GET] not supported"}`,
		},
		{
			description: "dynamodb client error reading account from database",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
				HTTPMethod: http.MethodPost,
			},
			mockReadAccountOutput:          nil,
			mockReadAccountError:           errors.New("mock read account error"),
			mockConvertCSQLOutput:          nil,
			mockConvertCSQLError:           nil,
			mockQueryDocumentsOutput:       nil,
			mockQueryDocumentsError:        nil,
			mockGeneratePresignedURLOutput: "",
			mockGeneratePresignedURLError:  nil,
			statusCode:                     http.StatusInternalServerError,
			body:                           `{"error": "error getting account values"}`,
		},
		{
			description: "dynamodb client account not found in database",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
				HTTPMethod: http.MethodPost,
			},
			mockReadAccountOutput:          nil,
			mockReadAccountError:           nil,
			mockConvertCSQLOutput:          nil,
			mockConvertCSQLError:           nil,
			mockQueryDocumentsOutput:       nil,
			mockQueryDocumentsError:        nil,
			mockGeneratePresignedURLOutput: "",
			mockGeneratePresignedURLError:  nil,
			statusCode:                     http.StatusInternalServerError,
			body:                           `{"error": "account [account_id] not found}`,
		},
		{
			description: "error unmarshalling request body csql query",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
				HTTPMethod: http.MethodPost,
				Body:       "---------",
			},
			mockReadAccountOutput:          &acct.Account{},
			mockReadAccountError:           nil,
			mockConvertCSQLOutput:          nil,
			mockConvertCSQLError:           nil,
			mockQueryDocumentsOutput:       nil,
			mockQueryDocumentsError:        nil,
			mockGeneratePresignedURLOutput: "",
			mockGeneratePresignedURLError:  nil,
			statusCode:                     http.StatusInternalServerError,
			body:                           `{"error": "error unmarshalling query"}`,
		},
		{
			description: "csql client error converting csql query",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
				HTTPMethod: http.MethodPost,
				Body:       `{"test": "query"}`,
			},
			mockReadAccountOutput:          &acct.Account{},
			mockReadAccountError:           nil,
			mockConvertCSQLOutput:          nil,
			mockConvertCSQLError:           errors.New("mock convert csql error"),
			mockQueryDocumentsOutput:       nil,
			mockQueryDocumentsError:        nil,
			mockGeneratePresignedURLOutput: "",
			mockGeneratePresignedURLError:  nil,
			statusCode:                     http.StatusInternalServerError,
			body:                           `{"error": "error converting query to csql"}`,
		},
		{
			description: "documentdb client error querying documents in database",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
				HTTPMethod: http.MethodPost,
				Body:       `{"test": "query"}`,
			},
			mockReadAccountOutput:          &acct.Account{},
			mockReadAccountError:           nil,
			mockConvertCSQLOutput:          []byte("test_query"),
			mockConvertCSQLError:           nil,
			mockQueryDocumentsOutput:       nil,
			mockQueryDocumentsError:        errors.New("mock query documents error"),
			mockGeneratePresignedURLOutput: "",
			mockGeneratePresignedURLError:  nil,
			statusCode:                     http.StatusInternalServerError,
			body:                           `{"error": "error runing [test_query] query"}`,
		},
		{
			description: "s3 client error presigning download urls",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
				HTTPMethod: http.MethodPost,
				Body:       `{"test": "query"}`,
			},
			mockReadAccountOutput: &acct.Account{},
			mockReadAccountError:  nil,
			mockConvertCSQLOutput: []byte("test_query"),
			mockConvertCSQLError:  nil,
			mockQueryDocumentsOutput: []docpars.Document{
				{
					Filename: "filename.jpg",
				},
			},
			mockQueryDocumentsError:        nil,
			mockGeneratePresignedURLOutput: "",
			mockGeneratePresignedURLError:  errors.New("mock generate presigned url error"),
			statusCode:                     http.StatusInternalServerError,
			body:                           `{"error": "error generating [filename.jpg] presigned url"}`,
		},
		{
			description: "successful handler invocation",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
				HTTPMethod: http.MethodPost,
				Body:       `{"test": "query"}`,
			},
			mockReadAccountOutput: &acct.Account{},
			mockReadAccountError:  nil,
			mockConvertCSQLOutput: []byte("test_query"),
			mockConvertCSQLError:  nil,
			mockQueryDocumentsOutput: []docpars.Document{
				{
					Filename: "filename.jpg",
				},
			},
			mockQueryDocumentsError:        nil,
			mockGeneratePresignedURLOutput: "https://s3.amazonaws.com/bucket/account_id/filename.jpg",
			mockGeneratePresignedURLError:  nil,
			statusCode:                     http.StatusOK,
			body:                           `{"message":"success","filenames":["filename.jpg"],"presigned_urls":["https://s3.amazonaws.com/bucket/account_id/filename.jpg"]}`,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			acctClient := &mockAcctClient{
				mockReadAccountOutput: test.mockReadAccountOutput,
				mockReadAccountError:  test.mockReadAccountError,
			}

			csqlClient := &mockCSQLClient{
				mockConvertCSQLOutput: test.mockConvertCSQLOutput,
				mockConvertCSQLError:  test.mockConvertCSQLError,
			}

			dbClient := &mockDBClient{
				mockQueryDocumentsOutput: test.mockQueryDocumentsOutput,
				mockQueryDocumentsError:  test.mockQueryDocumentsError,
			}

			fsClient := &mockFSClient{
				mockGeneratePresignedURLOutput: test.mockGeneratePresignedURLOutput,
				mockGeneratePresignedURLError:  test.mockGeneratePresignedURLError,
			}

			handlerFunc := handler(acctClient, csqlClient, dbClient, fsClient)

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
