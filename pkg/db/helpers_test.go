package db

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/cheesesteakio/api/pkg/docpars"
)

type mockS3Client struct {
	mockPutObjectOutput     *s3.PutObjectOutput
	mockPutObjectError      error
	mockPutObjectBucket     string
	mockPutObjectKey        string
	mockListObjectsV2Output *s3.ListObjectsV2Output
	mockListObjectsV2Error  error
	mockDeleteObjectsError  error
	mockDeleteObjectsBucket string
	mockDeleteObjectsKey    string
}

func (m *mockS3Client) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	m.mockPutObjectBucket = *input.Bucket
	m.mockPutObjectKey = *input.Key

	return m.mockPutObjectOutput, m.mockPutObjectError
}

func (m *mockS3Client) ListObjectsV2(input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
	return m.mockListObjectsV2Output, m.mockListObjectsV2Error
}

func (m *mockS3Client) DeleteObjects(input *s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error) {
	m.mockDeleteObjectsBucket = *input.Bucket
	m.mockDeleteObjectsKey = *input.Delete.Objects[0].Key

	return nil, m.mockDeleteObjectsError
}

type mockAthenaClient struct {
	mockStartQueryExecutionOutput *athena.StartQueryExecutionOutput
	mockStartQueryExecutionError  error
	mockGetQueryExecutionOutput   *athena.GetQueryExecutionOutput
	mockGetQueryExecutionError    error
	mockGetQueryResultsOutput     *athena.GetQueryResultsOutput
	mockGetQueryResultsError      error
}

func (m *mockAthenaClient) StartQueryExecution(input *athena.StartQueryExecutionInput) (*athena.StartQueryExecutionOutput, error) {
	return m.mockStartQueryExecutionOutput, m.mockStartQueryExecutionError
}

func (m *mockAthenaClient) GetQueryExecution(input *athena.GetQueryExecutionInput) (*athena.GetQueryExecutionOutput, error) {
	return m.mockGetQueryExecutionOutput, m.mockGetQueryExecutionError
}

func (m *mockAthenaClient) GetQueryResults(input *athena.GetQueryResultsInput) (*athena.GetQueryResultsOutput, error) {
	return m.mockGetQueryResultsOutput, m.mockGetQueryResultsError
}

func Test_uploadObject(t *testing.T) {
	tests := []struct {
		description        string
		mockPutObjectError error
		error              error
	}{
		{
			description:        "error putting object",
			mockPutObjectError: errors.New("mock put object error"),
			error:              errors.New("mock put object error"),
		},
		{
			description:        "successful upload object invocation",
			mockPutObjectError: nil,
			error:              nil,
		},
	}

	bucket := "bucket"
	body := struct {
		name string
	}{
		name: "name",
	}
	key := "key.json"

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			h := help{
				bucketName: bucket,
				s3Client: &mockS3Client{
					mockPutObjectError: test.mockPutObjectError,
				},
			}

			err := h.uploadObject(context.Background(), body, key)

			if err != nil {
				if err.Error() != test.error.Error() {
					t.Errorf("incorrect error, received: %s, expected: %s", err.Error(), test.error.Error())
				}
			} else {
				receivedBucket := h.s3Client.(*mockS3Client).mockPutObjectBucket
				receivedKey := h.s3Client.(*mockS3Client).mockPutObjectKey

				if receivedBucket != bucket {
					t.Errorf("incorrect bucket name, received: %s, expected: %s", receivedBucket, bucket)
				}

				if receivedKey != key {
					t.Errorf("incorrect key, received: %s, expected: %s", receivedKey, key)
				}
			}
		})
	}
}

func Test_listDocumentKeys(t *testing.T) {
	tests := []struct {
		description             string
		mockListObjectsV2Output *s3.ListObjectsV2Output
		mockListObjectsV2Error  error
		results                 []string
		error                   error
	}{
		{
			description:             "error listing objects",
			mockListObjectsV2Output: nil,
			mockListObjectsV2Error:  errors.New("mock list objects error"),
			results:                 nil,
			error:                   errors.New("mock list objects error"),
		},
		{
			description: "successful list document keys invocation",
			mockListObjectsV2Output: &s3.ListObjectsV2Output{
				Contents: []*s3.Object{
					{
						Key: aws.String("prefix/key.json"),
					},
				},
				IsTruncated: aws.Bool(false),
			},
			mockListObjectsV2Error: nil,
			results:                []string{"prefix/key.json"},
			error:                  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			h := &help{
				s3Client: &mockS3Client{
					mockListObjectsV2Output: test.mockListObjectsV2Output,
					mockListObjectsV2Error:  test.mockListObjectsV2Error,
				},
			}

			results, err := h.listDocumentKeys(context.Background(), "bucket", "prefix")

			if err != nil {
				if err.Error() != test.error.Error() {
					t.Errorf("incorrect error, received: %s, expected: %s", err.Error(), test.error.Error())
				}
			} else {
				for i, received := range results {
					expected := test.results[i]
					if received != expected {
						t.Errorf("incorrect result value, received: %s, expected: %s", received, expected)
					}
				}
			}
		})
	}
}

func Test_deleteDocumentsByKeys(t *testing.T) {
	tests := []struct {
		description            string
		mockDeleteObjectsError error
		error                  error
	}{
		{
			description:            "error deleting objects",
			mockDeleteObjectsError: errors.New("mock delete object error"),
			error:                  errors.New("mock delete object error"),
		},
		{
			description:            "successful delete objects by keys invocation",
			mockDeleteObjectsError: nil,
			error:                  nil,
		},
	}

	bucket := "bucket"
	key := "key.json"

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			h := &help{
				bucketName: bucket,
				s3Client: &mockS3Client{
					mockDeleteObjectsError: test.mockDeleteObjectsError,
				},
			}

			err := h.deleteDocumentsByKeys(context.Background(), []string{key})

			if err != nil {
				if err.Error() != test.error.Error() {
					t.Errorf("incorrect error, received: %s, expected: %s", err.Error(), test.error.Error())
				}
			} else {
				receivedBucket := h.s3Client.(*mockS3Client).mockDeleteObjectsBucket
				receivedKey := h.s3Client.(*mockS3Client).mockDeleteObjectsKey

				if receivedBucket != bucket {
					t.Errorf("incorrect bucket name, received: %s, expected: %s", receivedBucket, bucket)
				}

				if receivedKey != key {
					t.Errorf("incorrect key, received: %s, expected: %s", receivedKey, key)
				}
			}
		})
	}
}

func Test_executeQuery(t *testing.T) {
	tests := []struct {
		description                   string
		mockStartQueryExecutionOutput *athena.StartQueryExecutionOutput
		mockStartQueryExecutionError  error
		mockGetQueryExecutionOutput   *athena.GetQueryExecutionOutput
		mockGetQueryExecutionError    error
		executionID                   string
		state                         string
		error                         error
	}{
		{
			description:                   "error starting query execution",
			mockStartQueryExecutionOutput: nil,
			mockStartQueryExecutionError:  errors.New("mock start query execution error"),
			mockGetQueryExecutionOutput:   nil,
			mockGetQueryExecutionError:    nil,
			executionID:                   "",
			state:                         "",
			error:                         errors.New("mock start query execution error"),
		},
		{
			description: "error getting query execution",
			mockStartQueryExecutionOutput: &athena.StartQueryExecutionOutput{
				QueryExecutionId: aws.String("query_execution_id"),
			},
			mockStartQueryExecutionError: nil,
			mockGetQueryExecutionOutput:  nil,
			mockGetQueryExecutionError:   errors.New("mock get query execution error"),
			executionID:                  "",
			state:                        "",
			error:                        errors.New("mock get query execution error"),
		},
		{
			description: "successful execute query invocation",
			mockStartQueryExecutionOutput: &athena.StartQueryExecutionOutput{
				QueryExecutionId: aws.String("query_execution_id"),
			},
			mockStartQueryExecutionError: nil,
			mockGetQueryExecutionOutput: &athena.GetQueryExecutionOutput{
				QueryExecution: &athena.QueryExecution{
					Status: &athena.QueryExecutionStatus{
						State: aws.String("NOT_RUNNING"),
					},
				},
			},
			mockGetQueryExecutionError: nil,
			executionID:                "query_execution_id",
			state:                      "NOT_RUNNING",
			error:                      nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			h := &help{
				athenaClient: &mockAthenaClient{
					mockStartQueryExecutionOutput: test.mockStartQueryExecutionOutput,
					mockStartQueryExecutionError:  test.mockStartQueryExecutionError,
					mockGetQueryExecutionOutput:   test.mockGetQueryExecutionOutput,
					mockGetQueryExecutionError:    test.mockGetQueryExecutionError,
				},
			}

			executionID, state, err := h.executeQuery(context.Background(), []byte("query"))

			if err != nil {
				if err.Error() != test.error.Error() {
					t.Errorf("incorrect error, received: %s, expected: %s", err.Error(), test.error.Error())
				}
			} else {
				if *executionID != test.executionID {
					t.Errorf("incorrect execution id, received: %s, expected: %s", *executionID, test.executionID)
				}

				if *state != test.state {
					t.Errorf("incorrect state, received: %s, expected: %s", *state, test.state)
				}
			}
		})
	}
}

func Test_getQueryResultIDs(t *testing.T) {
	tests := []struct {
		description               string
		state                     string
		executionID               string
		mockGetQueryResultsOutput *athena.GetQueryResultsOutput
		mockGetQueryResultsError  error
		accountID                 string
		documentID                string
		error                     error
	}{
		{
			description:               "non-success state received",
			state:                     "NOT_SUCCEEDED",
			executionID:               "execution_id",
			mockGetQueryResultsOutput: nil,
			mockGetQueryResultsError:  nil,
			accountID:                 "",
			documentID:                "",
			error:                     errors.New("incorrect query state [NOT_SUCCEEDED]"),
		},
		{
			description:               "error getting query results",
			state:                     "SUCCEEDED",
			executionID:               "execution_id",
			mockGetQueryResultsOutput: nil,
			mockGetQueryResultsError:  errors.New("mock get query results error"),
			accountID:                 "",
			documentID:                "",
			error:                     errors.New("mock get query results error"),
		},
		{
			description: "successful get query result ids invocation",
			state:       "SUCCEEDED",
			executionID: "execution_id",
			mockGetQueryResultsOutput: &athena.GetQueryResultsOutput{
				ResultSet: &athena.ResultSet{
					Rows: []*athena.Row{
						{
							Data: []*athena.Datum{
								{
									VarCharValue: aws.String("account_id"),
								},
								{
									VarCharValue: aws.String("document_id"),
								},
							},
						},
					},
				},
			},
			mockGetQueryResultsError: nil,
			accountID:                "account_id",
			documentID:               "document_id",
			error:                    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			h := &help{
				athenaClient: &mockAthenaClient{
					mockGetQueryResultsOutput: test.mockGetQueryResultsOutput,
					mockGetQueryResultsError:  test.mockGetQueryResultsError,
				},
			}

			accountID, documentID, err := h.getQueryResultIDs(test.state, test.executionID)

			if err != nil {
				if err.Error() != test.error.Error() {
					t.Errorf("incorrect error, received: %s, expected: %s", err.Error(), test.error.Error())
				}
			} else {
				if *accountID != test.accountID {
					t.Errorf("incorrect account id, received: %s, expected: %s", *accountID, test.accountID)
				}

				if *documentID != test.documentID {
					t.Errorf("incorrect document id, received: %s, expected: %s", *documentID, test.documentID)
				}
			}
		})
	}
}

func Test_getQueryResultDocuments(t *testing.T) {
	tests := []struct {
		description               string
		state                     string
		executionID               string
		mockGetQueryResultsOutput *athena.GetQueryResultsOutput
		mockGetQueryResultsError  error
		documents                 []docpars.Document
		error                     error
	}{
		{
			description:               "non-success state received",
			state:                     "NOT_SUCCEEDED",
			executionID:               "execution_id",
			mockGetQueryResultsOutput: nil,
			mockGetQueryResultsError:  nil,
			documents:                 nil,
			error:                     errors.New("incorrect query state [NOT_SUCCEEDED]"),
		},
		{
			description:               "error getting query results",
			state:                     "SUCCEEDED",
			executionID:               "execution_id",
			mockGetQueryResultsOutput: nil,
			mockGetQueryResultsError:  errors.New("mock get query results error"),
			documents:                 nil,
			error:                     errors.New("mock get query results error"),
		},
		{
			description: "successful get query result ids invocation",
			state:       "SUCCEEDED",
			executionID: "execution_id",
			mockGetQueryResultsOutput: &athena.GetQueryResultsOutput{
				ResultSet: &athena.ResultSet{
					Rows: []*athena.Row{
						{
							Data: []*athena.Datum{
								{
									VarCharValue: aws.String("account_id"),
								},
								{
									VarCharValue: aws.String("bucket"),
								},
								{
									VarCharValue: aws.String("key.json"),
								},
							},
						},
					},
				},
			},
			mockGetQueryResultsError: nil,
			documents: []docpars.Document{
				{
					AccountID: "account_id",
					Filepath:  "bucket",
					Filename:  "key.json",
				},
			},
			error: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			h := &help{
				athenaClient: &mockAthenaClient{
					mockGetQueryResultsOutput: test.mockGetQueryResultsOutput,
					mockGetQueryResultsError:  test.mockGetQueryResultsError,
				},
			}

			documents, err := h.getQueryResultDocuments(test.state, test.executionID)

			if err != nil {
				if err.Error() != test.error.Error() {
					t.Errorf("incorrect error, received: %s, expected: %s", err.Error(), test.error.Error())
				}
			} else {
				received := documents[0]
				expected := test.documents[0]

				if received.AccountID != expected.AccountID {
					t.Errorf("incorrect account id, received: %s, expected: %s", received.AccountID, expected.AccountID)
				}

				if received.Filepath != expected.Filepath {
					t.Errorf("incorrect filepath, received: %s, expected: %s", received.Filepath, expected.Filepath)
				}

				if received.Filename != expected.Filename {
					t.Errorf("incorrect filename, received: %s, expected: %s", received.Filename, expected.Filename)
				}
			}
		})
	}
}
