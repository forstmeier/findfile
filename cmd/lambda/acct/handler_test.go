package main

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-lambda-go/events"

	"github.com/cheesesteakio/api/pkg/acct"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

type mockAcctClient struct {
	mockCreateAccountError   error
	mockGetAccountByIDOutput *acct.Account
	mockGetAccountByIDError  error
	mockDeleteAccountError   error
}

func (m *mockAcctClient) CreateAccount(ctx context.Context, accountID, bucketName string) error {
	return m.mockCreateAccountError
}

func (m *mockAcctClient) GetAccountByID(ctx context.Context, accountID string) (*acct.Account, error) {
	return m.mockGetAccountByIDOutput, m.mockGetAccountByIDError
}

func (m *mockAcctClient) GetAccountBySecondaryID(ctx context.Context, secondaryID string) (*acct.Account, error) {
	return nil, nil
}

func (m *mockAcctClient) UpdateAccount(ctx context.Context, accountID string, values map[string]string) error {
	return nil
}

func (m *mockAcctClient) DeleteAccount(ctx context.Context, accountID string) error {
	return m.mockDeleteAccountError
}

type mockPartitionerClient struct {
	mockAddPartitionError    error
	mockRemovePartitionError error
}

func (m *mockPartitionerClient) AddPartition(ctx context.Context, accountID string) error {
	return m.mockAddPartitionError
}

func (m *mockPartitionerClient) RemovePartition(ctx context.Context, accountID string) error {
	return m.mockRemovePartitionError
}

type mockFSClient struct {
	mockCreateFileWatcherError error
	mockDeleteFileWatcherError error
}

func (m *mockFSClient) CreateFileWatcher(ctx context.Context, path string) error {
	return m.mockCreateFileWatcherError
}

func (m *mockFSClient) DeleteFileWatcher(ctx context.Context, path string) error {
	return m.mockDeleteFileWatcherError
}

func Test_handler(t *testing.T) {
	tests := []struct {
		description                string
		request                    events.APIGatewayProxyRequest
		mockCreateAccountError     error
		mockGetAccountByIDOutput   *acct.Account
		mockGetAccountByIDError    error
		mockDeleteAccountError     error
		mockAddPartitionError      error
		mockRemovePartitionError   error
		mockCreateFileWatcherError error
		mockDeleteFileWatcherError error
		statusCode                 int
		body                       string
	}{
		{
			description: "unsupported http method recieved",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodGet,
			},
			mockCreateAccountError:     nil,
			mockGetAccountByIDOutput:   nil,
			mockGetAccountByIDError:    nil,
			mockDeleteAccountError:     nil,
			mockAddPartitionError:      nil,
			mockRemovePartitionError:   nil,
			mockCreateFileWatcherError: nil,
			mockDeleteFileWatcherError: nil,
			statusCode:                 http.StatusBadRequest,
			body:                       `{"error": "http method \[GET\] not supported"}`,
		},
		{
			description: "error unmarshalling user information",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodPost,
				Body:       "---------",
			},
			mockCreateAccountError:     nil,
			mockGetAccountByIDOutput:   nil,
			mockGetAccountByIDError:    nil,
			mockDeleteAccountError:     nil,
			mockAddPartitionError:      nil,
			mockRemovePartitionError:   nil,
			mockCreateFileWatcherError: nil,
			mockDeleteFileWatcherError: nil,
			statusCode:                 http.StatusInternalServerError,
			body:                       `{"error": "error unmarshalling request"}`,
		},
		{
			description: "error creating user account",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodPost,
				Body:       `{"bucket_name": "bucket"}`,
			},
			mockCreateAccountError:     errors.New("mock create account error"),
			mockGetAccountByIDOutput:   nil,
			mockGetAccountByIDError:    nil,
			mockDeleteAccountError:     nil,
			mockAddPartitionError:      nil,
			mockRemovePartitionError:   nil,
			mockCreateFileWatcherError: nil,
			mockDeleteFileWatcherError: nil,
			statusCode:                 http.StatusInternalServerError,
			body:                       `{"error": "error creating user account"}`,
		},
		{
			description: "error creating account partition",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodPost,
				Body:       `{"bucket_name": "bucket"}`,
			},
			mockCreateAccountError:     nil,
			mockGetAccountByIDOutput:   nil,
			mockGetAccountByIDError:    nil,
			mockDeleteAccountError:     nil,
			mockAddPartitionError:      errors.New("mock add partition error"),
			mockRemovePartitionError:   nil,
			mockCreateFileWatcherError: nil,
			mockDeleteFileWatcherError: nil,
			statusCode:                 http.StatusInternalServerError,
			body:                       `{"error": "error creating account partition"}`,
		},
		{
			description: "error creating file watcher",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodPost,
				Body:       `{"bucket_name": "bucket"}`,
			},
			mockCreateAccountError:     nil,
			mockGetAccountByIDOutput:   nil,
			mockGetAccountByIDError:    nil,
			mockDeleteAccountError:     nil,
			mockAddPartitionError:      nil,
			mockRemovePartitionError:   nil,
			mockCreateFileWatcherError: errors.New("mock create file watcher error"),
			mockDeleteFileWatcherError: nil,
			statusCode:                 http.StatusInternalServerError,
			body:                       `{"error": "error creating bucket notification"}`,
		},
		{
			description: "successful handler create invocation",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodPost,
				Body:       `{"bucket_name": "bucket"}`,
			},
			mockCreateAccountError:     nil,
			mockGetAccountByIDOutput:   nil,
			mockGetAccountByIDError:    nil,
			mockDeleteAccountError:     nil,
			mockAddPartitionError:      nil,
			mockRemovePartitionError:   nil,
			mockCreateFileWatcherError: nil,
			mockDeleteFileWatcherError: nil,
			statusCode:                 http.StatusOK,
			body:                       `{"message": "success", "account_id": ".*"}`,
		},
		{
			description: "account id not provided in request header",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodDelete,
			},
			mockCreateAccountError:     nil,
			mockGetAccountByIDOutput:   nil,
			mockGetAccountByIDError:    nil,
			mockDeleteAccountError:     nil,
			mockAddPartitionError:      nil,
			mockRemovePartitionError:   nil,
			mockCreateFileWatcherError: nil,
			mockDeleteFileWatcherError: nil,
			statusCode:                 http.StatusBadRequest,
			body:                       `{"error": "account id not provided"}`,
		},
		{
			description: "error getting user account",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodDelete,
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
			},
			mockCreateAccountError:     nil,
			mockGetAccountByIDOutput:   nil,
			mockGetAccountByIDError:    errors.New("mock get account by id error"),
			mockDeleteAccountError:     nil,
			mockAddPartitionError:      nil,
			mockRemovePartitionError:   nil,
			mockCreateFileWatcherError: nil,
			mockDeleteFileWatcherError: nil,
			statusCode:                 http.StatusInternalServerError,
			body:                       `{"error": "error getting user account"}`,
		},
		{
			description: "dynamodb client error removing account from database",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodDelete,
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
			},
			mockCreateAccountError:     nil,
			mockGetAccountByIDOutput:   &acct.Account{},
			mockGetAccountByIDError:    nil,
			mockDeleteAccountError:     errors.New("mock delete account error"),
			mockAddPartitionError:      nil,
			mockRemovePartitionError:   nil,
			mockCreateFileWatcherError: nil,
			mockDeleteFileWatcherError: nil,
			statusCode:                 http.StatusInternalServerError,
			body:                       `{"error": "error removing user account"}`,
		},
		{
			description: "error removing account partition",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodDelete,
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
			},
			mockCreateAccountError:     nil,
			mockGetAccountByIDOutput:   &acct.Account{},
			mockGetAccountByIDError:    nil,
			mockDeleteAccountError:     nil,
			mockAddPartitionError:      nil,
			mockRemovePartitionError:   errors.New("mock remove partition error"),
			mockCreateFileWatcherError: nil,
			mockDeleteFileWatcherError: nil,
			statusCode:                 http.StatusInternalServerError,
			body:                       `{"error": "error removing account partition"}`,
		},
		{
			description: "error deleting file watcher",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodDelete,
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
			},
			mockCreateAccountError: nil,
			mockGetAccountByIDOutput: &acct.Account{
				BucketName: "bucket",
			},
			mockGetAccountByIDError:    nil,
			mockDeleteAccountError:     nil,
			mockAddPartitionError:      nil,
			mockRemovePartitionError:   nil,
			mockCreateFileWatcherError: nil,
			mockDeleteFileWatcherError: errors.New("mock delete file watcher error"),
			statusCode:                 http.StatusInternalServerError,
			body:                       `{"error": "error deleting bucket notification"}`,
		},
		{
			description: "successful handler delete invocation",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodDelete,
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
			},
			mockCreateAccountError: nil,
			mockGetAccountByIDOutput: &acct.Account{
				BucketName: "bucket",
			},
			mockGetAccountByIDError:    nil,
			mockDeleteAccountError:     nil,
			mockAddPartitionError:      nil,
			mockRemovePartitionError:   nil,
			mockCreateFileWatcherError: nil,
			mockDeleteFileWatcherError: nil,
			statusCode:                 http.StatusOK,
			body:                       `{"message": "success", "account_id": "account_id"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			acctClient := &mockAcctClient{
				mockCreateAccountError:   test.mockCreateAccountError,
				mockGetAccountByIDOutput: test.mockGetAccountByIDOutput,
				mockGetAccountByIDError:  test.mockGetAccountByIDError,
				mockDeleteAccountError:   test.mockDeleteAccountError,
			}

			partitionerClient := &mockPartitionerClient{
				mockAddPartitionError:    test.mockAddPartitionError,
				mockRemovePartitionError: test.mockRemovePartitionError,
			}

			fsClient := &mockFSClient{
				mockCreateFileWatcherError: test.mockCreateFileWatcherError,
				mockDeleteFileWatcherError: test.mockDeleteFileWatcherError,
			}

			handlerFunc := handler(acctClient, partitionerClient, fsClient)

			response, _ := handlerFunc(context.Background(), test.request)

			if response.StatusCode != test.statusCode {
				t.Errorf("incorrect status code, received: %d, expected: %d", response.StatusCode, test.statusCode)
			}

			matched, err := regexp.MatchString(test.body, response.Body)
			if err != nil {
				t.Fatalf("error matching body regexp: %s", err.Error())
			}

			if !matched {
				t.Errorf("incorrect body, received: %s", response.Body)
			}
		})
	}
}
