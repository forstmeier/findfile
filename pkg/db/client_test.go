package db

import (
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/forstmeier/findfile/pkg/pars"
)

func TestNew(t *testing.T) {
	client, err := New(
		session.New(),
		"url",
		"username",
		"password",
	)
	if err != nil {
		t.Errorf("incorrect error, received: %v, expected: nil", err)
	}

	if client == nil {
		t.Error("error creating parser client")
	}
}

type mockHelper struct {
	mockExecuteCreateError error
	mockExecuteBulkBody    io.Reader
	mockExecuteBulkError   error
	mockExecuteDeleteBody  io.Reader
	mockExecuteDeleteError error
	mockExecuteQueryBody   io.Reader
	mockExecuteQueryOutput io.ReadCloser
	mockExecuteQueryError  error
}

func (m *mockHelper) executeCreate(ctx context.Context) error {
	return m.mockExecuteCreateError
}

func (m *mockHelper) executeBulk(ctx context.Context, body io.Reader) error {
	m.mockExecuteBulkBody = body
	return m.mockExecuteBulkError
}

func (m *mockHelper) executeDelete(ctx context.Context, body io.Reader) error {
	m.mockExecuteDeleteBody = body
	return m.mockExecuteDeleteError
}

func (m *mockHelper) executeQuery(ctx context.Context, body io.Reader) (io.ReadCloser, error) {
	m.mockExecuteQueryBody = body
	return m.mockExecuteQueryOutput, m.mockExecuteQueryError
}

func TestUpsertDocuments(t *testing.T) {
	tests := []struct {
		description          string
		mockExecuteBulkBody  string
		mockExecuteBulkError error
		error                error
	}{
		{
			description:          "error executing bulk request",
			mockExecuteBulkBody:  "",
			mockExecuteBulkError: errors.New("mock execute bulk error"),
			error:                &ExecuteBulkError{},
		},
		{
			description: "successful invocation",
			mockExecuteBulkBody: `{ "index": { "_id": "doc_id" } }
{"id":"doc_id","entity":"","file_bucket":"","file_key":""}
`,
			mockExecuteBulkError: nil,
			error:                nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			h := &mockHelper{
				mockExecuteBulkError: test.mockExecuteBulkError,
			}

			c := &Client{
				helper: h,
			}

			err := c.UpsertDocuments(context.Background(), []pars.Document{
				{
					ID: "doc_id",
				},
			})

			if err != nil {
				switch e := test.error.(type) {
				case *ExecuteBulkError:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
					}
				default:
					t.Fatalf("unexpected error type: %v", err)
				}
			} else {
				mockExecuteBulkBody, err := io.ReadAll(c.helper.(*mockHelper).mockExecuteBulkBody)
				if err != nil {
					t.Fatalf("error reading body: %v", err)
				}

				if string(mockExecuteBulkBody) != test.mockExecuteBulkBody {
					t.Errorf("incorrect body, received: %s, expected: %s", mockExecuteBulkBody, test.mockExecuteBulkBody)
				}
			}
		})
	}
}

func TestDeleteDocumentsByIDs(t *testing.T) {
	tests := []struct {
		description            string
		mockExecuteDeleteBody  string
		mockExecuteDeleteError error
		error                  error
	}{
		{
			description:            "error executing delete request",
			mockExecuteDeleteBody:  "",
			mockExecuteDeleteError: errors.New("mock execute delete error"),
			error:                  &ExecuteDeleteError{},
		},
		{
			description:            "successful invocation",
			mockExecuteDeleteBody:  `{ "query": { "bool": { "minimum_should_match": 1, "should": [ { "match": { "file_bucket": "bucket", "file_key": "key.jpeg" } } ] } } }`,
			mockExecuteDeleteError: nil,
			error:                  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			h := &mockHelper{
				mockExecuteDeleteError: test.mockExecuteDeleteError,
			}

			c := &Client{
				helper: h,
			}

			err := c.DeleteDocumentsByIDs(context.Background(), []string{"bucket/key.jpeg"})

			if err != nil {
				switch e := test.error.(type) {
				case *ExecuteDeleteError:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
					}
				default:
					t.Fatalf("unexpected error type: %v", err)
				}
			} else {
				mockExecuteDeleteBody, err := io.ReadAll(c.helper.(*mockHelper).mockExecuteDeleteBody)
				if err != nil {
					t.Fatalf("error reading body: %v", err)
				}

				if string(mockExecuteDeleteBody) != test.mockExecuteDeleteBody {
					t.Errorf("incorrect body, received: %s, expected: %s", mockExecuteDeleteBody, test.mockExecuteDeleteBody)
				}
			}
		})
	}
}

func TestDeleteDocumentsByBuckets(t *testing.T) {
	tests := []struct {
		description            string
		mockExecuteDeleteBody  string
		mockExecuteDeleteError error
		error                  error
	}{
		{
			description:            "error executing delete request",
			mockExecuteDeleteBody:  "",
			mockExecuteDeleteError: errors.New("mock execute delete error"),
			error:                  &ExecuteDeleteError{},
		},
		{
			description:            "successful invocation",
			mockExecuteDeleteBody:  `{ "query": { "bool": { "minimum_should_match": 1, "should": [ { "match": { "file_bucket": "bucket" } } ] } } }`,
			mockExecuteDeleteError: nil,
			error:                  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			h := &mockHelper{
				mockExecuteDeleteError: test.mockExecuteDeleteError,
			}

			c := &Client{
				helper: h,
			}

			err := c.DeleteDocumentsByBuckets(context.Background(), []string{"bucket"})

			if err != nil {
				switch e := test.error.(type) {
				case *ExecuteDeleteError:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
					}
				default:
					t.Fatalf("unexpected error type: %v", err)
				}
			} else {
				mockExecuteDeleteBody, err := io.ReadAll(c.helper.(*mockHelper).mockExecuteDeleteBody)
				if err != nil {
					t.Fatalf("error reading body: %v", err)
				}

				if string(mockExecuteDeleteBody) != test.mockExecuteDeleteBody {
					t.Errorf("incorrect body, received: %s, expected: %s", mockExecuteDeleteBody, test.mockExecuteDeleteBody)
				}
			}
		})
	}
}

func TestQueryDocuments(t *testing.T) {
	tests := []struct {
		description            string
		mockExecuteQueryBody   string
		mockExecuteQueryOutput io.ReadCloser
		mockExecuteQueryError  error
		documents              []pars.Document
		error                  error
	}{
		{
			description:            "error executing query request",
			mockExecuteQueryBody:   "",
			mockExecuteQueryOutput: nil,
			mockExecuteQueryError:  errors.New("mock execute query error"),
			error:                  &ExecuteQueryError{},
		},
		{
			description:            "successful invocation",
			mockExecuteQueryBody:   `{ "query": { "match": { "pages.lines.text": { "query": "example text", "fuzziness": "AUTO" } } } }`,
			mockExecuteQueryOutput: io.NopCloser(strings.NewReader(`{ "hits": { "hits": [ { "_source": { "id": "doc_id" } } ] } }`)),
			mockExecuteQueryError:  nil,
			documents: []pars.Document{
				{
					ID: "doc_id",
				},
			},
			error: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			h := &mockHelper{
				mockExecuteQueryOutput: test.mockExecuteQueryOutput,
				mockExecuteQueryError:  test.mockExecuteQueryError,
			}

			c := &Client{
				helper: h,
			}

			documents, err := c.QueryDocuments(context.Background(), Query{
				Text: "example text",
			})

			if err != nil {
				switch e := test.error.(type) {
				case *ExecuteQueryError:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
					}
				default:
					t.Fatalf("unexpected error type: %v", err)
				}
			} else {
				mockExecuteQueryBody, err := io.ReadAll(c.helper.(*mockHelper).mockExecuteQueryBody)
				if err != nil {
					t.Fatalf("error reading body: %v", err)
				}

				if string(mockExecuteQueryBody) != test.mockExecuteQueryBody {
					t.Errorf("incorrect body, received: %s, expected: %s", mockExecuteQueryBody, test.mockExecuteQueryBody)
				}

				if !reflect.DeepEqual(documents, test.documents) {
					t.Errorf("incorrect documents, received: %+v, expected: %+v", documents, test.documents)
				}
			}
		})
	}
}
