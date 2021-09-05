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
	"github.com/cheesesteakio/api/pkg/db"
	"github.com/cheesesteakio/api/pkg/pars"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

type mockAcctClient struct {
	mockGetAccountByIDOutput *acct.Account
	mockGetAccountByIDError  error
}

func (m *mockAcctClient) CreateAccount(ctx context.Context, accountID, bucketName string) error {
	return nil
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
	return nil
}

type mockCQLClient struct {
	mockConvertCQLOutput []byte
	mockConvertCQLError  error
}

func (m *mockCQLClient) ConvertCQL(ctx context.Context, accountID string, cqlQuery map[string]interface{}) ([]byte, error) {
	return m.mockConvertCQLOutput, m.mockConvertCQLError
}

type mockDBClient struct {
	mockQueryDocumentsOutput []pars.Document
	mockQueryDocumentsError  error
}

func (m *mockDBClient) UpsertDocuments(ctx context.Context, documents []pars.Document) error {
	return nil
}

func (m *mockDBClient) DeleteDocuments(ctx context.Context, documentsInfo []db.DocumentInfo) error {
	return nil
}

func (m *mockDBClient) QueryDocuments(ctx context.Context, query []byte) ([]pars.Document, error) {
	return m.mockQueryDocumentsOutput, m.mockQueryDocumentsError
}

func Test_handler(t *testing.T) {
	tests := []struct {
		description              string
		request                  events.APIGatewayProxyRequest
		mockGetAccountByIDOutput *acct.Account
		mockGetAccountByIDError  error
		mockConvertCQLOutput     []byte
		mockConvertCQLError      error
		mockQueryDocumentsOutput []pars.Document
		mockQueryDocumentsError  error
		statusCode               int
		body                     string
	}{
		{
			description:              "no account id in request",
			request:                  events.APIGatewayProxyRequest{},
			mockGetAccountByIDOutput: nil,
			mockGetAccountByIDError:  nil,
			mockConvertCQLOutput:     nil,
			mockConvertCQLError:      nil,
			mockQueryDocumentsOutput: nil,
			mockQueryDocumentsError:  nil,
			statusCode:               http.StatusBadRequest,
			body:                     `{"error": "account id not provided"}`,
		},
		{
			description: "unsupported http method",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
				HTTPMethod: http.MethodGet,
			},
			mockGetAccountByIDOutput: nil,
			mockGetAccountByIDError:  nil,
			mockConvertCQLOutput:     nil,
			mockConvertCQLError:      nil,
			mockQueryDocumentsOutput: nil,
			mockQueryDocumentsError:  nil,
			statusCode:               http.StatusBadRequest,
			body:                     `{"error": "http method [GET] not supported"}`,
		},
		{
			description: "dynamodb client error reading account from database",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
				HTTPMethod: http.MethodPost,
			},
			mockGetAccountByIDOutput: nil,
			mockGetAccountByIDError:  errors.New("mock read account error"),
			mockConvertCQLOutput:     nil,
			mockConvertCQLError:      nil,
			mockQueryDocumentsOutput: nil,
			mockQueryDocumentsError:  nil,
			statusCode:               http.StatusInternalServerError,
			body:                     `{"error": "error getting account values"}`,
		},
		{
			description: "dynamodb client account not found in database",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
				HTTPMethod: http.MethodPost,
			},
			mockGetAccountByIDOutput: nil,
			mockGetAccountByIDError:  nil,
			mockConvertCQLOutput:     nil,
			mockConvertCQLError:      nil,
			mockQueryDocumentsOutput: nil,
			mockQueryDocumentsError:  nil,
			statusCode:               http.StatusInternalServerError,
			body:                     `{"error": "account [account_id] not found}`,
		},
		{
			description: "error unmarshalling request body cql query",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
				HTTPMethod: http.MethodPost,
				Body:       "---------",
			},
			mockGetAccountByIDOutput: &acct.Account{},
			mockGetAccountByIDError:  nil,
			mockConvertCQLOutput:     nil,
			mockConvertCQLError:      nil,
			mockQueryDocumentsOutput: nil,
			mockQueryDocumentsError:  nil,
			statusCode:               http.StatusInternalServerError,
			body:                     `{"error": "error unmarshalling query"}`,
		},
		{
			description: "cql client error converting cql query",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					accountIDHeader: "account_id",
				},
				HTTPMethod: http.MethodPost,
				Body:       `{"test": "query"}`,
			},
			mockGetAccountByIDOutput: &acct.Account{},
			mockGetAccountByIDError:  nil,
			mockConvertCQLOutput:     nil,
			mockConvertCQLError:      errors.New("mock convert cql error"),
			mockQueryDocumentsOutput: nil,
			mockQueryDocumentsError:  nil,
			statusCode:               http.StatusInternalServerError,
			body:                     `{"error": "error converting cql to query"}`,
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
			mockGetAccountByIDOutput: &acct.Account{},
			mockGetAccountByIDError:  nil,
			mockConvertCQLOutput:     []byte("test_query"),
			mockConvertCQLError:      nil,
			mockQueryDocumentsOutput: nil,
			mockQueryDocumentsError:  errors.New("mock query documents error"),
			statusCode:               http.StatusInternalServerError,
			body:                     `{"error": "error running query"}`,
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
			mockGetAccountByIDOutput: &acct.Account{},
			mockGetAccountByIDError:  nil,
			mockConvertCQLOutput:     []byte("test_query"),
			mockConvertCQLError:      nil,
			mockQueryDocumentsOutput: []pars.Document{
				{
					Filename: "filename.jpg",
				},
			},
			mockQueryDocumentsError: nil,
			statusCode:              http.StatusOK,
			body:                    `{"message":"success","filenames":["filename.jpg"]}`,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			acctClient := &mockAcctClient{
				mockGetAccountByIDOutput: test.mockGetAccountByIDOutput,
				mockGetAccountByIDError:  test.mockGetAccountByIDError,
			}

			cqlClient := &mockCQLClient{
				mockConvertCQLOutput: test.mockConvertCQLOutput,
				mockConvertCQLError:  test.mockConvertCQLError,
			}

			dbClient := &mockDBClient{
				mockQueryDocumentsOutput: test.mockQueryDocumentsOutput,
				mockQueryDocumentsError:  test.mockQueryDocumentsError,
			}

			handlerFunc := handler(acctClient, cqlClient, dbClient)

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
