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
	mockCreateAccountError error
	mockUpdateAccountError error
	mockReadAccountOutput  *acct.Account
	mockReadAccountError   error
	mockDeleteAccountError error
}

func (m *mockAcctClient) CreateAccount(ctx context.Context, accountID, bucketName string) error {
	return m.mockCreateAccountError
}

func (m *mockAcctClient) ReadAccount(ctx context.Context, accountID string) (*acct.Account, error) {
	return m.mockReadAccountOutput, m.mockReadAccountError
}

func (m *mockAcctClient) UpdateAccount(ctx context.Context, accountID string, values map[string]string) error {
	return m.mockUpdateAccountError
}

func (m *mockAcctClient) DeleteAccount(ctx context.Context, accountID string) error {
	return m.mockDeleteAccountError
}

func Test_handler(t *testing.T) {
	tests := []struct {
		description            string
		request                events.APIGatewayProxyRequest
		mockCreateAccountError error
		mockUpdateAccountError error
		mockReadAccountOutput  *acct.Account
		mockReadAccountError   error
		mockDeleteAccountError error
		statusCode             int
		body                   string
	}{
		{
			description: "unsupported http method recieved",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodGet,
			},
			mockCreateAccountError: nil,
			mockUpdateAccountError: nil,
			mockReadAccountOutput:  nil,
			mockReadAccountError:   nil,
			mockDeleteAccountError: nil,
			statusCode:             http.StatusBadRequest,
			body:                   `{"error": "http method \[GET\] not supported"}`,
		},
		{
			description: "error unmarshalling user information",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodPost,
				Body:       "---------",
			},
			mockCreateAccountError: nil,
			mockUpdateAccountError: nil,
			mockReadAccountOutput:  nil,
			mockReadAccountError:   nil,
			mockDeleteAccountError: nil,
			statusCode:             http.StatusInternalServerError,
			body:                   `{"error": "error unmarshalling request"}`,
		},
		{
			description: "error creating user account",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodPost,
				Body:       `{"bucket_name": "bucket"}`,
			},
			mockCreateAccountError: errors.New("mock create account error"),
			mockUpdateAccountError: nil,
			mockReadAccountOutput:  nil,
			mockReadAccountError:   nil,
			mockDeleteAccountError: nil,
			statusCode:             http.StatusInternalServerError,
			body:                   `{"error": "error creating user account"}`,
		},
		{
			description: "successful handler create invocation",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodPost,
				Body:       `{"bucket_name": "bucket"}`,
			},
			mockCreateAccountError: nil,
			mockUpdateAccountError: nil,
			mockReadAccountOutput:  nil,
			mockReadAccountError:   nil,
			mockDeleteAccountError: nil,
			statusCode:             http.StatusOK,
			body:                   `{"message": "success", "account_id": ".*"}`,
		},
		{
			description: "account id not provided in request header",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodDelete,
			},
			mockCreateAccountError: nil,
			mockUpdateAccountError: nil,
			mockReadAccountOutput:  nil,
			mockReadAccountError:   nil,
			mockDeleteAccountError: nil,
			statusCode:             http.StatusBadRequest,
			body:                   `{"error": "account id not provided"}`,
		},
		{
			description: "dynamodb client error removing account from database",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodDelete,
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
			},
			mockCreateAccountError: nil,
			mockUpdateAccountError: nil,
			mockReadAccountOutput:  &acct.Account{},
			mockReadAccountError:   nil,
			mockDeleteAccountError: errors.New("mock delete account error"),
			statusCode:             http.StatusInternalServerError,
			body:                   `{"error": "error removing user account"}`,
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
			mockUpdateAccountError: nil,
			mockReadAccountOutput:  &acct.Account{},
			mockReadAccountError:   nil,
			mockDeleteAccountError: nil,
			statusCode:             http.StatusOK,
			body:                   `{"message": "success", "account_id": "account_id"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			acctClient := &mockAcctClient{
				mockCreateAccountError: test.mockCreateAccountError,
				mockUpdateAccountError: test.mockUpdateAccountError,
				mockReadAccountOutput:  test.mockReadAccountOutput,
				mockReadAccountError:   test.mockReadAccountError,
				mockDeleteAccountError: test.mockDeleteAccountError,
			}

			handlerFunc := handler(acctClient)

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
