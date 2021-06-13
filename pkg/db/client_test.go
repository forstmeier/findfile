package db

import (
	"context"
	"errors"
	"testing"

	"github.com/cheesesteakio/api/pkg/docpars"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestNew(t *testing.T) {
	client, err := New("databaseName", "collectionName")

	if err != nil {
		t.Errorf("error received creating database client, %s:", err.Error())
	}

	if client == nil {
		t.Error("error creating database client")
	}
}

type mockDocumentDBClient struct {
	insertManyError         error
	findOneAndReplaceOutput *mongo.SingleResult
	deleteOneError          error
	findOutput              *mongo.Cursor
	findError               error
}

func (m *mockDocumentDBClient) InsertMany(ctx context.Context, documents []interface{}, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	return nil, m.insertManyError
}

func (m *mockDocumentDBClient) FindOneAndReplace(ctx context.Context, filter interface{}, replacement interface{}, opts ...*options.FindOneAndReplaceOptions) *mongo.SingleResult {
	return m.findOneAndReplaceOutput
}

func (m *mockDocumentDBClient) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return nil, m.deleteOneError
}

func (m *mockDocumentDBClient) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	return m.findOutput, m.findError
}

func TestCreate(t *testing.T) {
	tests := []struct {
		description     string
		insertManyError error
		error           error
	}{
		{
			description:     "error inserting documents into database",
			insertManyError: errors.New("mock insert error"),
			error:           &ErrorCreateDocuments{},
		},
		{
			description:     "successful create documents invocation",
			insertManyError: nil,
			error:           nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				documentDBClient: &mockDocumentDBClient{
					insertManyError: test.insertManyError,
				},
			}

			err := client.Create(context.Background(), []docpars.Document{{}})

			if err != nil {
				switch test.error.(type) {
				case *ErrorCreateDocuments:
					var testError *ErrorCreateDocuments
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

func TestUpdate(t *testing.T) {
	tests := []struct {
		description             string
		findOneAndReplaceOutput *mongo.SingleResult
		error                   error
	}{
		{
			description:             "error updating documents in database",
			findOneAndReplaceOutput: &mongo.SingleResult{},
			error:                   &ErrorUpdateDocuments{},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				documentDBClient: &mockDocumentDBClient{
					findOneAndReplaceOutput: test.findOneAndReplaceOutput,
				},
			}

			err := client.Update(context.Background(), []docpars.Document{{}})

			if err != nil {
				switch test.error.(type) {
				case *ErrorUpdateDocuments:
					var testError *ErrorUpdateDocuments
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

func TestDelete(t *testing.T) {
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

			err := client.Delete(context.Background(), []DocumentInfo{
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

func TestQuery(t *testing.T) {
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

			documents, err := client.Query(context.Background(), []byte("query"))

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
