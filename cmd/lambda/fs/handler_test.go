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

	"github.com/cheesesteakio/api/pkg/acct"
	"github.com/cheesesteakio/api/pkg/fs"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

type mockFSClient struct {
	mockGeneratePresignedURLOutput string
	mockGeneratePresignedURLError  error
	mockDeleteFilesError           error
}

func (m *mockFSClient) GenerateUploadURL(ctx context.Context, accountID string, fileInfo fs.FileInfo) (string, error) {
	return m.mockGeneratePresignedURLOutput, m.mockGeneratePresignedURLError
}

func (m *mockFSClient) GenerateDownloadURL(ctx context.Context, accountID string, fileInfo fs.FileInfo) (string, error) {
	return "", nil
}

func (m *mockFSClient) ListFilesByAccountID(ctx context.Context, filepath, accountID string) ([]fs.FileInfo, error) {
	return nil, nil
}

func (m *mockFSClient) DeleteFiles(ctx context.Context, accountID string, filesInfo []fs.FileInfo) error {
	return m.mockDeleteFilesError
}

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

func Test_handler(t *testing.T) {
	tests := []struct {
		description                    string
		request                        events.APIGatewayProxyRequest
		mockGeneratePresignedURLOutput string
		mockGeneratePresignedURLError  error
		mockDeleteFilesError           error
		mockReadAccountOutput          *acct.Account
		mockReadAccountError           error
		statusCode                     int
		body                           string
	}{
		{
			description:                    "no account id in request",
			request:                        events.APIGatewayProxyRequest{},
			mockGeneratePresignedURLOutput: "",
			mockGeneratePresignedURLError:  nil,
			mockDeleteFilesError:           nil,
			mockReadAccountOutput:          nil,
			mockReadAccountError:           nil,
			statusCode:                     http.StatusBadRequest,
			body:                           `{"error": "account id not provided"}`,
		},
		{
			description: "dynamodb client error reading account from database",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
				HTTPMethod: http.MethodPost,
			},
			mockGeneratePresignedURLOutput: "",
			mockGeneratePresignedURLError:  nil,
			mockDeleteFilesError:           nil,
			mockReadAccountOutput:          nil,
			mockReadAccountError:           errors.New("mock read account error"),
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
			mockGeneratePresignedURLOutput: "",
			mockGeneratePresignedURLError:  nil,
			mockDeleteFilesError:           nil,
			mockReadAccountOutput:          nil,
			mockReadAccountError:           nil,
			statusCode:                     http.StatusInternalServerError,
			body:                           `{"error": "account [account_id] not found}`,
		},
		{
			description: "unmarshal received file names",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
				HTTPMethod: http.MethodPost,
				Body:       "---------",
			},
			mockGeneratePresignedURLOutput: "",
			mockGeneratePresignedURLError:  nil,
			mockDeleteFilesError:           nil,
			mockReadAccountOutput:          &acct.Account{},
			mockReadAccountError:           nil,
			statusCode:                     http.StatusInternalServerError,
			body:                           `{"error": "error unmarshalling filenames array"}`,
		},
		{
			description: "unsupported http method",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
				HTTPMethod: http.MethodPatch,
				Body:       `["filename.jpg"]`,
			},
			mockGeneratePresignedURLOutput: "",
			mockGeneratePresignedURLError:  nil,
			mockDeleteFilesError:           nil,
			mockReadAccountOutput:          &acct.Account{},
			mockReadAccountError:           nil,
			statusCode:                     http.StatusBadRequest,
			body:                           `{"error": "http method [PATCH] not supported"}`,
		},
		{
			description: "s3 client error presigning upload urls",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
				HTTPMethod: http.MethodPost,
				Body:       `["filename.jpg"]`,
			},
			mockGeneratePresignedURLOutput: "",
			mockGeneratePresignedURLError:  errors.New("mock generate presigned url error"),
			mockDeleteFilesError:           nil,
			mockReadAccountOutput:          &acct.Account{},
			mockReadAccountError:           nil,
			statusCode:                     http.StatusInternalServerError,
			body:                           `{"error": "error generating [filename.jpg] presigned url"}`,
		},
		{
			description: "successful post handler invocation",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
				HTTPMethod: http.MethodPost,
				Body:       `["filename.jpg"]`,
			},
			mockGeneratePresignedURLOutput: "https://s3.amazonaws.com/bucket/account_id/filename.jpg",
			mockGeneratePresignedURLError:  nil,
			mockDeleteFilesError:           nil,
			mockReadAccountOutput:          &acct.Account{},
			mockReadAccountError:           nil,
			statusCode:                     http.StatusOK,
			body:                           `{"message": "success", "urls": ["https://s3.amazonaws.com/bucket/account_id/filename.jpg"]}`,
		},
		{
			description: "s3 client delete objects error",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
				HTTPMethod: http.MethodDelete,
				Body:       `["filename.jpg"]`,
			},
			mockGeneratePresignedURLOutput: "",
			mockGeneratePresignedURLError:  nil,
			mockDeleteFilesError:           errors.New("mock delete object error"),
			mockReadAccountOutput:          &acct.Account{},
			mockReadAccountError:           nil,
			statusCode:                     http.StatusInternalServerError,
			body:                           `{"error": "error deleting files"}`,
		},
		{
			description: "successful delete handler invocation",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
				HTTPMethod: http.MethodDelete,
				Body:       `["filename.jpg"]`,
			},
			mockGeneratePresignedURLOutput: "",
			mockGeneratePresignedURLError:  nil,
			mockDeleteFilesError:           nil,
			mockReadAccountOutput:          &acct.Account{},
			mockReadAccountError:           nil,
			statusCode:                     http.StatusOK,
			body:                           `{"message": "success"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			acctClient := &mockAcctClient{
				mockReadAccountOutput: test.mockReadAccountOutput,
				mockReadAccountError:  test.mockReadAccountError,
			}

			fsClient := &mockFSClient{
				mockGeneratePresignedURLOutput: test.mockGeneratePresignedURLOutput,
				mockGeneratePresignedURLError:  test.mockGeneratePresignedURLError,
				mockDeleteFilesError:           test.mockDeleteFilesError,
			}

			handlerFunc := handler(acctClient, fsClient)

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
