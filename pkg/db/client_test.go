package db

import (
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/elastic/go-elasticsearch/v7/esapi"

	"github.com/forstmeier/findfile/pkg/pars"
)

func TestNew(t *testing.T) {
	client, err := New(session.New())
	if err != nil {
		t.Errorf("incorrect error, received: %v, expected: nil", err)
	}

	if client == nil {
		t.Error("error creating parser client")
	}
}

type mockHelper struct {
	mockExecuteIndexMappingOutput *esapi.Response
	mockExecuteIndexMappingError  error
	mockExecuteBulkBody           io.Reader
	mockExecuteBulkOutput         *esapi.Response
	mockExecuteBulkError          error
	mockExecuteDeleteBody         io.Reader
	mockExecuteDeleteOutput       *esapi.Response
	mockExecuteDeleteError        error
	mockExecuteQueryBody          io.Reader
	mockExecuteQueryOutput        *esapi.Response
	mockExecuteQueryError         error
}

func (mh *mockHelper) executeIndexMapping(ctx context.Context, request *esapi.IndicesPutMappingRequest) (*esapi.Response, error) {
	return mh.mockExecuteIndexMappingOutput, mh.mockExecuteIndexMappingError
}

func (mh *mockHelper) executeBulk(ctx context.Context, request *esapi.BulkRequest) (*esapi.Response, error) {
	mh.mockExecuteBulkBody = request.Body
	return mh.mockExecuteBulkOutput, mh.mockExecuteBulkError
}

func (mh *mockHelper) executeDelete(ctx context.Context, request *esapi.DeleteByQueryRequest) (*esapi.Response, error) {
	mh.mockExecuteDeleteBody = request.Body
	return mh.mockExecuteDeleteOutput, mh.mockExecuteDeleteError
}

func (mh *mockHelper) executeQuery(ctx context.Context, request *esapi.SearchRequest) (*esapi.Response, error) {
	mh.mockExecuteQueryBody = request.Body
	return mh.mockExecuteQueryOutput, mh.mockExecuteQueryError
}

func TestUpsertDocuments(t *testing.T) {
	tests := []struct {
		description           string
		mockExecuteBulkBody   string
		mockExecuteBulkOutput *esapi.Response
		mockExecuteBulkError  error
		error                 error
	}{
		{
			description:         "error executing bulk request",
			mockExecuteBulkBody: "",
			mockExecuteBulkOutput: &esapi.Response{
				StatusCode: 300,
			},
			mockExecuteBulkError: errors.New("mock execute bulk error"),
			error:                &ExecuteBulkError{},
		},
		{
			description: "successful invocation",
			mockExecuteBulkBody: `{ "index": { "_id": "doc_id" } }
{"id":"doc_id","entity":"","file_bucket":"","file_key":"","pages":null}
`,
			mockExecuteBulkOutput: &esapi.Response{
				StatusCode: 200,
			},
			mockExecuteBulkError: nil,
			error:                nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			h := &mockHelper{
				mockExecuteBulkOutput: test.mockExecuteBulkOutput,
				mockExecuteBulkError:  test.mockExecuteBulkError,
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

func TestDeleteDocuments(t *testing.T) {
	tests := []struct {
		description             string
		mockExecuteDeleteBody   string
		mockExecuteDeleteOutput *esapi.Response
		mockExecuteDeleteError  error
		error                   error
	}{
		{
			description:           "error executing delete request",
			mockExecuteDeleteBody: "",
			mockExecuteDeleteOutput: &esapi.Response{
				StatusCode: 300,
			},
			mockExecuteDeleteError: errors.New("mock execute delete error"),
			error:                  &ExecuteDeleteError{},
		},
		{
			description:           "successful invocation",
			mockExecuteDeleteBody: `{ "query": { "bool": { "minimum_should_match": 1, "should": [ { "match": { "file_bucket": "bucket", "file_key": "key.jpeg" } } ] } } }`,
			mockExecuteDeleteOutput: &esapi.Response{
				StatusCode: 200,
			},
			mockExecuteDeleteError: nil,
			error:                  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			h := &mockHelper{
				mockExecuteDeleteOutput: test.mockExecuteDeleteOutput,
				mockExecuteDeleteError:  test.mockExecuteDeleteError,
			}

			c := &Client{
				helper: h,
			}

			err := c.DeleteDocuments(context.Background(), []string{"bucket/key.jpeg"})

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
		mockExecuteQueryOutput *esapi.Response
		mockExecuteQueryError  error
		documents              []pars.Document
		error                  error
	}{
		{
			description:          "error executing query request",
			mockExecuteQueryBody: "",
			mockExecuteQueryOutput: &esapi.Response{
				StatusCode: 300,
			},
			mockExecuteQueryError: errors.New("mock execute query error"),
			error:                 &ExecuteQueryError{},
		},
		{
			description:          "successful invocation",
			mockExecuteQueryBody: `{ "query": { "nested": { "path": "pages.lines", "query": { "match": { "lines.text": "example text" } } } } }`,
			mockExecuteQueryOutput: &esapi.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{ "hits": { "hits": [ { "id": "doc_id" } ] } }`)),
			},
			mockExecuteQueryError: nil,
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
