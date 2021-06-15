package main

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

type mockInfraClient struct {
	createFilesystemError error
	deleteFilesystemError error
	createDatabaseError   error
	deleteDatabaseError   error
}

func (m *mockInfraClient) CreateFilesystem(ctx context.Context, accountID string) error {
	return m.createFilesystemError
}

func (m *mockInfraClient) DeleteFilesystem(ctx context.Context, accountID string) error {
	return m.deleteFilesystemError
}

func (m *mockInfraClient) CreateDatabase(ctx context.Context, accountID string) error {
	return m.createDatabaseError
}

func (m *mockInfraClient) DeleteDatabase(ctx context.Context, accountID string) error {
	return m.deleteDatabaseError
}

func Test_handler(t *testing.T) {
	tests := []struct {
		description           string
		createFilesystemError error
		deleteFilesystemError error
		createDatabaseError   error
		deleteDatabaseError   error
		requestParameters     map[string]string
		requestMethod         string
		statusCode            int
		body                  string
	}{
		{
			description:           "http method not supported",
			createFilesystemError: nil,
			deleteFilesystemError: nil,
			createDatabaseError:   nil,
			deleteDatabaseError:   nil,
			requestParameters:     nil,
			requestMethod:         http.MethodGet,
			statusCode:            http.StatusBadRequest,
			body:                  `{"error": "method not supported [GET]"}`,
		},
		{
			description:           "error validating request query parameters",
			createFilesystemError: nil,
			deleteFilesystemError: nil,
			createDatabaseError:   nil,
			deleteDatabaseError:   nil,
			requestParameters:     nil,
			requestMethod:         http.MethodPost,
			statusCode:            http.StatusBadRequest,
			body:                  `{"error": "missing parameters [account id,create filesystem,create database]"}`,
		},
		{
			description:           "error creating filesystem",
			createFilesystemError: errors.New("mock create filesystem error"),
			deleteFilesystemError: nil,
			createDatabaseError:   nil,
			deleteDatabaseError:   nil,
			requestParameters: map[string]string{
				accountIDParameter:  "account_id",
				filesystemParameter: "true",
				databaseParameter:   "true",
			},
			requestMethod: http.MethodPost,
			statusCode:    http.StatusInternalServerError,
			body:          `{"error": "error creating filesystem"}`,
		},
		{
			description:           "error creating database",
			createFilesystemError: nil,
			deleteFilesystemError: nil,
			createDatabaseError:   errors.New("mock create database error"),
			deleteDatabaseError:   nil,
			requestParameters: map[string]string{
				accountIDParameter:  "account_id",
				filesystemParameter: "true",
				databaseParameter:   "true",
			},
			requestMethod: http.MethodPost,
			statusCode:    http.StatusInternalServerError,
			body:          `{"error": "error creating database"}`,
		},
		{
			description:           "error deleting filesystem",
			createFilesystemError: nil,
			deleteFilesystemError: errors.New("mock delete filesystem error"),
			createDatabaseError:   nil,
			deleteDatabaseError:   nil,
			requestParameters: map[string]string{
				accountIDParameter:  "account_id",
				filesystemParameter: "true",
				databaseParameter:   "true",
			},
			requestMethod: http.MethodDelete,
			statusCode:    http.StatusInternalServerError,
			body:          `{"error": "error deleting filesystem"}`,
		},
		{
			description:           "error deleting database",
			createFilesystemError: nil,
			deleteFilesystemError: nil,
			createDatabaseError:   nil,
			deleteDatabaseError:   errors.New("mock delete database error"),
			requestParameters: map[string]string{
				accountIDParameter:  "account_id",
				filesystemParameter: "true",
				databaseParameter:   "true",
			},
			requestMethod: http.MethodDelete,
			statusCode:    http.StatusInternalServerError,
			body:          `{"error": "error deleting database"}`,
		},
		{
			description:           "successful create invocation",
			createFilesystemError: nil,
			deleteFilesystemError: nil,
			createDatabaseError:   nil,
			deleteDatabaseError:   nil,
			requestParameters: map[string]string{
				accountIDParameter:  "account_id",
				filesystemParameter: "true",
				databaseParameter:   "true",
			},
			requestMethod: http.MethodPost,
			statusCode:    http.StatusOK,
			body:          `{"message": "[POST] success"}`,
		},
		{
			description:           "successful delete invocation",
			createFilesystemError: nil,
			deleteFilesystemError: nil,
			createDatabaseError:   nil,
			deleteDatabaseError:   nil,
			requestParameters: map[string]string{
				accountIDParameter:  "account_id",
				filesystemParameter: "true",
				databaseParameter:   "true",
			},
			requestMethod: http.MethodDelete,
			statusCode:    http.StatusOK,
			body:          `{"message": "[DELETE] success"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			infraClient := &mockInfraClient{
				createFilesystemError: test.createFilesystemError,
				deleteFilesystemError: test.deleteFilesystemError,
				createDatabaseError:   test.createDatabaseError,
				deleteDatabaseError:   test.deleteDatabaseError,
			}

			handlerFunc := handler(infraClient)

			response, _ := handlerFunc(context.Background(), events.APIGatewayProxyRequest{
				QueryStringParameters: test.requestParameters,
				HTTPMethod:            test.requestMethod,
			})

			if response.StatusCode != test.statusCode {
				t.Errorf("incorrect status code, received: %d, expected: %d", response.StatusCode, test.statusCode)
			}

			if response.Body != test.body {
				t.Errorf("incorrect body, received: %s, expected: %s", response.Body, test.body)
			}
		})
	}
}

func Test_validateCreateParameters(t *testing.T) {
	tests := []struct {
		description string
		parameters  map[string]string
		output      string
		check       bool
	}{
		{
			description: "missing all parameter values",
			parameters:  map[string]string{},
			output:      "account id,create filesystem,create database",
			check:       false,
		},
		{
			description: "missing create filesystem parameter",
			parameters: map[string]string{
				accountIDParameter: "account_id",
				databaseParameter:  "true",
			},
			output: "create filesystem",
			check:  false,
		},
		{
			description: "all parameter values received",
			parameters: map[string]string{
				accountIDParameter:  "account_id",
				filesystemParameter: "true",
				databaseParameter:   "true",
			},
			output: "",
			check:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			output, check := validateCreateParameters(test.parameters)

			if output != test.output {
				t.Errorf("incorrect output, received: %s, expected: %s", output, test.output)
			}

			if check != test.check {
				t.Errorf("incorrect check, received: %t, expected: %t", check, test.check)
			}
		})
	}
}
