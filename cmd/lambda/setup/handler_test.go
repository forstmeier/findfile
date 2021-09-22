package main

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/findfiledev/api/pkg/pars"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

type mockDBClient struct {
	mockSetupDatabaseError error
}

func (m *mockDBClient) SetupDatabase(ctx context.Context) error {
	return m.mockSetupDatabaseError
}

func (m *mockDBClient) UpsertDocuments(ctx context.Context, documents []pars.Document) error {
	return nil
}

func (m *mockDBClient) DeleteDocuments(ctx context.Context, documentIDs []string) error {
	return nil
}

func (m *mockDBClient) QueryDocumentsByFQL(ctx context.Context, query []byte) ([]pars.Document, error) {
	return nil, nil
}

func (m *mockDBClient) QueryDocumentKeysByFileInfo(ctx context.Context, query []byte) ([]string, error) {
	return nil, nil
}

func Test_handler(t *testing.T) {
	tests := []struct {
		description            string
		mockSetupDatabaseError error
		mockSendResponseError  error
		event                  cfn.Event
		responseReason         string
	}{
		{
			description:            "received event non-create",
			mockSetupDatabaseError: nil,
			mockSendResponseError:  nil,
			event: cfn.Event{
				RequestType: cfn.RequestDelete,
			},
			responseReason: "received non-create event type [Delete]",
		},
		{
			description:            "error setting up database",
			mockSetupDatabaseError: errors.New("mock setup database error"),
			mockSendResponseError:  nil,
			event: cfn.Event{
				RequestType: cfn.RequestCreate,
			},
			responseReason: "setup database error [mock setup database error]",
		},
		{
			description:            "error sending custom resource response",
			mockSetupDatabaseError: nil,
			mockSendResponseError:  errors.New("mock send response error"),
			event: cfn.Event{
				RequestType: cfn.RequestCreate,
			},
			responseReason: "send response error [mock send response error]",
		},
		{
			description:            "successful handler invocation",
			mockSetupDatabaseError: nil,
			mockSendResponseError:  nil,
			event: cfn.Event{
				RequestType: cfn.RequestCreate,
			},
			responseReason: "successful invocation",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			dbClient := &mockDBClient{
				mockSetupDatabaseError: test.mockSetupDatabaseError,
			}

			mockResponse := &cfn.Response{}

			mockSendResponse := func(response *cfn.Response) error {
				mockResponse = response
				return test.mockSendResponseError
			}

			handlerFunc := handler(dbClient, mockSendResponse)

			// suppress error since returned value is always nil
			handlerFunc(context.Background(), test.event)

			if mockResponse.Reason != test.responseReason {
				t.Errorf("incorrect response reason, received: %s, expected: %s", mockResponse.Reason, test.responseReason)
			}
		})
	}
}
