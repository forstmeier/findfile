package db

import (
	"context"
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/cheesesteakio/api/pkg/docpars"
)

func TestGetConnectionURI(t *testing.T) {
	received := GetConnectionURI("username", "password", "endpoint")
	expected := "mongodb://username:password@endpoint/sample-database?tls=true&replicaSet=rs0&readpreference=secondaryPreferred"
	if received != expected {
		t.Errorf("incorrect connection uri, received: %s, expected: %s", received, expected)
	}
}

func TestGetTLSConfig(t *testing.T) {
	tlsConfig, err := GetTLSConfig("../../etc/aws/rds-combined-ca-bundle.pem")
	if tlsConfig == nil {
		t.Error("error creating tls config")
	}

	if err != nil {
		t.Errorf("error creating tls config: %s", err.Error())
	}
}

func TestNew(t *testing.T) {
	ddb, err := mongo.NewClient(nil)
	if err != nil {
		t.Fatalf("error creating test session: %s", err.Error())
	}

	client := New(ddb, "databaseName", "collectionName")
	if client == nil {
		t.Error("error creating database client")
	}
}

type mockDocumentDBClient struct {
	updateOneError error
	deleteOneError error
	findOutput     *mongo.Cursor
	findError      error
}

func (m *mockDocumentDBClient) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return nil, m.updateOneError
}

func (m *mockDocumentDBClient) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return nil, m.deleteOneError
}

func (m *mockDocumentDBClient) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	return m.findOutput, m.findError
}

func TestCreateOrUpdateDocuments(t *testing.T) {
	tests := []struct {
		description    string
		updateOneError error
		error          error
	}{
		{
			description:    "error inserting documents into database",
			updateOneError: errors.New("mock update error"),
			error:          &ErrorUpdateDocument{},
		},
		{
			description:    "successful create documents invocation",
			updateOneError: nil,
			error:          nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				documentDBClient: &mockDocumentDBClient{
					updateOneError: test.updateOneError,
				},
			}

			err := client.CreateOrUpdateDocuments(context.Background(), []docpars.Document{{}})

			if err != nil {
				switch test.error.(type) {
				case *ErrorUpdateDocument:
					var testError *ErrorUpdateDocument
					if !errors.As(err, &testError) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, testError)
					}
				default:
					t.Fatalf("unexpected error type: %v", err)
				}
			} else {
				if err != test.error {
					t.Errorf("incorrect error, received: %v, expected: %v", err, test.error)
				}
			}
		})
	}
}

func TestDeleteDocuments(t *testing.T) {
	tests := []struct {
		description    string
		deleteOneError error
		error          error
	}{
		{
			description:    "error deleting documents from database",
			deleteOneError: errors.New("mock delete error"),
			error:          &ErrorDeleteDocuments{},
		},
		{
			description:    "successful delete documents invocation",
			deleteOneError: nil,
			error:          nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				documentDBClient: &mockDocumentDBClient{
					deleteOneError: test.deleteOneError,
				},
			}

			err := client.DeleteDocuments(context.Background(), []DocumentInfo{
				{
					Filename: "filename",
					Filepath: "filepath",
				},
			})

			if err != nil {
				switch test.error.(type) {
				case *ErrorDeleteDocuments:
					var testError *ErrorDeleteDocuments
					if !errors.As(err, &testError) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, testError)
					}
				default:
					t.Fatalf("unexpected error type: %v", err)
				}
			} else {
				if err != test.error {
					t.Errorf("incorrect error, received: %v, expected: %v", err, test.error)
				}
			}
		})
	}
}

func TestQueryDocuments(t *testing.T) {
	tests := []struct {
		description string
		findOutput  *mongo.Cursor
		findError   error
		error       error
	}{
		{
			description: "error querying documents in database",
			findOutput:  nil,
			findError:   errors.New("mock query error"),
			error:       &ErrorQueryDocuments{},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				documentDBClient: &mockDocumentDBClient{
					findOutput: test.findOutput,
					findError:  test.findError,
				},
			}

			documents, err := client.QueryDocuments(context.Background(), []byte("query"))

			if err != nil {
				switch test.error.(type) {
				case *ErrorQueryDocuments:
					var testError *ErrorQueryDocuments
					if !errors.As(err, &testError) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, testError)
					}
				default:
					t.Fatalf("unexpected error type: %v", err)
				}
			} else {

				_ = documents // TEMP

				if err != test.error {
					t.Errorf("incorrect error, received: %v, expected: %v", err, test.error)
				}
			}
		})
	}
}
