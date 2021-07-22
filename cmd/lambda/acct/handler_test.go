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
	"github.com/cheesesteakio/api/pkg/fs"
	"github.com/cheesesteakio/api/pkg/subscr"
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

func (m *mockAcctClient) CreateAccount(ctx context.Context, accountID string) error {
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

type mockSubscrClient struct {
	mockCreateSubscriptionOutput *subscr.Subscription
	mockCreateSubscriptionError  error
	mockRemoveSubscriptionError  error
}

func (m *mockSubscrClient) CreateSubscription(ctx context.Context, accountID string, info subscr.SubscriberInfo) (*subscr.Subscription, error) {
	return m.mockCreateSubscriptionOutput, m.mockCreateSubscriptionError
}

func (m *mockSubscrClient) RemoveSubscription(ctx context.Context, subscription subscr.Subscription) error {
	return m.mockRemoveSubscriptionError
}

func (m *mockSubscrClient) AddUsage(ctx context.Context, itemID string, itemQuantity int64) error {
	return nil
}

type mockFSClient struct {
	mockDeleteFilesError           error
	mockListFilesByAccountIDOutput []fs.FileInfo
	mockListFilesByAccountIDError  error
}

func (m *mockFSClient) GenerateUploadURL(ctx context.Context, accountID string, fileInfo fs.FileInfo) (string, error) {
	return "", nil
}

func (m *mockFSClient) GenerateDownloadURL(ctx context.Context, accountID string, fileInfo fs.FileInfo) (string, error) {
	return "", nil
}

func (m *mockFSClient) ListFilesByAccountID(ctx context.Context, filepath, accountID string) ([]fs.FileInfo, error) {
	return m.mockListFilesByAccountIDOutput, m.mockListFilesByAccountIDError
}

func (m *mockFSClient) DeleteFiles(ctx context.Context, accountID string, filesInfo []fs.FileInfo) error {
	return m.mockDeleteFilesError
}

func Test_handler(t *testing.T) {
	tests := []struct {
		description                    string
		request                        events.APIGatewayProxyRequest
		mockCreateAccountError         error
		mockUpdateAccountError         error
		mockReadAccountOutput          *acct.Account
		mockReadAccountError           error
		mockDeleteAccountError         error
		mockCreateSubscriptionOutput   *subscr.Subscription
		mockCreateSubscriptionError    error
		mockRemoveSubscriptionError    error
		mockListFilesByAccountIDOutput []fs.FileInfo
		mockListFilesByAccountIDError  error
		mockDeleteFilesError           error
		statusCode                     int
		body                           string
	}{
		{
			description: "unsupported http method recieved",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodGet,
			},
			mockCreateAccountError:         nil,
			mockUpdateAccountError:         nil,
			mockReadAccountOutput:          nil,
			mockReadAccountError:           nil,
			mockDeleteAccountError:         nil,
			mockCreateSubscriptionOutput:   nil,
			mockCreateSubscriptionError:    nil,
			mockRemoveSubscriptionError:    nil,
			mockListFilesByAccountIDOutput: nil,
			mockListFilesByAccountIDError:  nil,
			mockDeleteFilesError:           nil,
			statusCode:                     http.StatusBadRequest,
			body:                           `{"error": "http method \[GET\] not supported"}`,
		},
		{
			description: "error unmarshalling user information",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodPost,
				Body:       "---------",
			},
			mockCreateAccountError:         nil,
			mockUpdateAccountError:         nil,
			mockReadAccountOutput:          nil,
			mockReadAccountError:           nil,
			mockDeleteAccountError:         nil,
			mockCreateSubscriptionOutput:   nil,
			mockCreateSubscriptionError:    nil,
			mockRemoveSubscriptionError:    nil,
			mockListFilesByAccountIDOutput: nil,
			mockListFilesByAccountIDError:  nil,
			mockDeleteFilesError:           nil,
			statusCode:                     http.StatusInternalServerError,
			body:                           `{"error": "error unmarshalling subscriber info"}`,
		},
		{
			description: "error creating user account",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodPost,
				Body:       `{"email": "test@email.com"}`,
			},
			mockCreateAccountError:         errors.New("mock create account error"),
			mockUpdateAccountError:         nil,
			mockReadAccountOutput:          nil,
			mockReadAccountError:           nil,
			mockDeleteAccountError:         nil,
			mockCreateSubscriptionOutput:   nil,
			mockCreateSubscriptionError:    nil,
			mockRemoveSubscriptionError:    nil,
			mockListFilesByAccountIDOutput: nil,
			mockListFilesByAccountIDError:  nil,
			mockDeleteFilesError:           nil,
			statusCode:                     http.StatusInternalServerError,
			body:                           `{"error": "error creating user account"}`,
		},
		{
			description: "error creating user subscription",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodPost,
				Body:       `{"email": "test@email.com"}`,
			},
			mockCreateAccountError:         nil,
			mockUpdateAccountError:         nil,
			mockReadAccountOutput:          nil,
			mockReadAccountError:           nil,
			mockDeleteAccountError:         nil,
			mockCreateSubscriptionOutput:   nil,
			mockCreateSubscriptionError:    errors.New("mock create subscription error"),
			mockRemoveSubscriptionError:    nil,
			mockListFilesByAccountIDOutput: nil,
			mockListFilesByAccountIDError:  nil,
			mockDeleteFilesError:           nil,
			statusCode:                     http.StatusInternalServerError,
			body:                           `{"error": "error creating user subscription"}`,
		},
		{
			description: "error update user account with subscription",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodPost,
				Body:       `{"email": "test@email.com"}`,
			},
			mockCreateAccountError: nil,
			mockUpdateAccountError: errors.New("mock update account error"),
			mockReadAccountOutput:  nil,
			mockReadAccountError:   nil,
			mockDeleteAccountError: nil,
			mockCreateSubscriptionOutput: &subscr.Subscription{
				ID:                    "test_subscription_id",
				StripePaymentMethodID: "test_stripe_payment_method_id",
				StripeCustomerID:      "test_stripe_customer_id",
				StripeSubscriptionID:  "test_stripe_subscription_id",
			},
			mockCreateSubscriptionError:    nil,
			mockRemoveSubscriptionError:    nil,
			mockListFilesByAccountIDOutput: nil,
			mockListFilesByAccountIDError:  nil,
			mockDeleteFilesError:           nil,
			statusCode:                     http.StatusInternalServerError,
			body:                           `{"error": "error adding subscription to user account"}`,
		},
		{
			description: "successful handler create invocation",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodPost,
				Body:       `{"email": "test@email.com"}`,
			},
			mockCreateAccountError: nil,
			mockUpdateAccountError: nil,
			mockReadAccountOutput:  nil,
			mockReadAccountError:   nil,
			mockDeleteAccountError: nil,
			mockCreateSubscriptionOutput: &subscr.Subscription{
				ID:                    "test_subscription_id",
				StripePaymentMethodID: "test_stripe_payment_method_id",
				StripeCustomerID:      "test_stripe_customer_id",
				StripeSubscriptionID:  "test_stripe_subscription_id",
			},
			mockCreateSubscriptionError:    nil,
			mockRemoveSubscriptionError:    nil,
			mockListFilesByAccountIDOutput: nil,
			mockListFilesByAccountIDError:  nil,
			mockDeleteFilesError:           nil,
			statusCode:                     http.StatusOK,
			body:                           `{"message": "success", "account_id": ".*"}`,
		},
		{
			description: "account id not provided in request header",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodDelete,
			},
			mockCreateAccountError:         nil,
			mockUpdateAccountError:         nil,
			mockReadAccountOutput:          nil,
			mockReadAccountError:           nil,
			mockDeleteAccountError:         nil,
			mockCreateSubscriptionOutput:   nil,
			mockCreateSubscriptionError:    nil,
			mockRemoveSubscriptionError:    nil,
			mockListFilesByAccountIDOutput: nil,
			mockListFilesByAccountIDError:  nil,
			mockDeleteFilesError:           nil,
			statusCode:                     http.StatusBadRequest,
			body:                           `{"error": "account id not provided"}`,
		},
		{
			description: "dynamodb client error reading account from database",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodDelete,
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
			},
			mockCreateAccountError:         nil,
			mockUpdateAccountError:         nil,
			mockReadAccountOutput:          nil,
			mockReadAccountError:           errors.New("mock read account error"),
			mockDeleteAccountError:         nil,
			mockCreateSubscriptionOutput:   nil,
			mockCreateSubscriptionError:    nil,
			mockRemoveSubscriptionError:    nil,
			mockListFilesByAccountIDOutput: nil,
			mockListFilesByAccountIDError:  nil,
			mockDeleteFilesError:           nil,
			statusCode:                     http.StatusInternalServerError,
			body:                           `{"error": "error getting account values"}`,
		},
		{
			description: "stripe client error deleting user subscription",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodDelete,
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
			},
			mockCreateAccountError: nil,
			mockUpdateAccountError: nil,
			mockReadAccountOutput: &acct.Account{
				StripeCustomerID: "stripe_customer_id",
			},
			mockReadAccountError:           nil,
			mockDeleteAccountError:         nil,
			mockCreateSubscriptionOutput:   nil,
			mockCreateSubscriptionError:    nil,
			mockRemoveSubscriptionError:    errors.New("mock delete subscription error"),
			mockListFilesByAccountIDOutput: nil,
			mockListFilesByAccountIDError:  nil,
			mockDeleteFilesError:           nil,
			statusCode:                     http.StatusInternalServerError,
			body:                           `{"error": "error removing user subscription"}`,
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
			mockReadAccountOutput: &acct.Account{
				StripeCustomerID: "stripe_customer_id",
			},
			mockReadAccountError:           nil,
			mockDeleteAccountError:         errors.New("mock delete account error"),
			mockCreateSubscriptionOutput:   nil,
			mockCreateSubscriptionError:    nil,
			mockRemoveSubscriptionError:    nil,
			mockListFilesByAccountIDOutput: nil,
			mockListFilesByAccountIDError:  nil,
			mockDeleteFilesError:           nil,
			statusCode:                     http.StatusInternalServerError,
			body:                           `{"error": "error removing user account"}`,
		},
		{
			description: "s3 client error listing files in filesystem",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodDelete,
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
			},
			mockCreateAccountError: nil,
			mockUpdateAccountError: nil,
			mockReadAccountOutput: &acct.Account{
				StripeCustomerID: "stripe_customer_id",
			},
			mockReadAccountError:           nil,
			mockDeleteAccountError:         nil,
			mockCreateSubscriptionOutput:   nil,
			mockCreateSubscriptionError:    nil,
			mockRemoveSubscriptionError:    nil,
			mockListFilesByAccountIDOutput: nil,
			mockListFilesByAccountIDError:  errors.New("mock list files error"),
			mockDeleteFilesError:           nil,
			statusCode:                     http.StatusInternalServerError,
			body:                           `{"error": "error listing user files"}`,
		},
		{
			description: "s3 client error removing files from filesystem",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodDelete,
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
			},
			mockCreateAccountError: nil,
			mockUpdateAccountError: nil,
			mockReadAccountOutput: &acct.Account{
				StripeCustomerID: "stripe_customer_id",
			},
			mockReadAccountError:           nil,
			mockDeleteAccountError:         nil,
			mockCreateSubscriptionOutput:   nil,
			mockCreateSubscriptionError:    nil,
			mockRemoveSubscriptionError:    nil,
			mockListFilesByAccountIDOutput: []fs.FileInfo{},
			mockListFilesByAccountIDError:  nil,
			mockDeleteFilesError:           errors.New("mock delete files error"),
			statusCode:                     http.StatusInternalServerError,
			body:                           `{"error": "error removing user files"}`,
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
			mockReadAccountOutput: &acct.Account{
				StripeCustomerID: "stripe_customer_id",
			},
			mockReadAccountError:           nil,
			mockDeleteAccountError:         nil,
			mockCreateSubscriptionOutput:   nil,
			mockCreateSubscriptionError:    nil,
			mockRemoveSubscriptionError:    nil,
			mockListFilesByAccountIDOutput: []fs.FileInfo{},
			mockListFilesByAccountIDError:  nil,
			mockDeleteFilesError:           nil,
			statusCode:                     http.StatusOK,
			body:                           `{"message": "success", "account_id": "account_id"}`,
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

			subscrClient := &mockSubscrClient{
				mockCreateSubscriptionOutput: test.mockCreateSubscriptionOutput,
				mockCreateSubscriptionError:  test.mockCreateSubscriptionError,
				mockRemoveSubscriptionError:  test.mockRemoveSubscriptionError,
			}

			fsClient := &mockFSClient{
				mockListFilesByAccountIDOutput: test.mockListFilesByAccountIDOutput,
				mockListFilesByAccountIDError:  test.mockListFilesByAccountIDError,
				mockDeleteFilesError:           test.mockDeleteFilesError,
			}

			handlerFunc := handler(acctClient, subscrClient, fsClient)

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
