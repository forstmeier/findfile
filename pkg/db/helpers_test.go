package db

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/findfiledev/api/pkg/pars"
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
	m.mockDeleteObjectsKey = *input.Delete.Objects[0].Key
	m.mockDeleteObjectsBucket = *input.Bucket

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

type mockGlueClient struct {
	mockStartCrawlerName  string
	mockStartCrawlerError error
}

func (m *mockGlueClient) StartCrawler(input *glue.StartCrawlerInput) (*glue.StartCrawlerOutput, error) {
	m.mockStartCrawlerName = *input.Name

	return nil, m.mockStartCrawlerError
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
				databaseBucket: bucket,
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
				databaseBucket: bucket,
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
				receivedKey := h.s3Client.(*mockS3Client).mockDeleteObjectsKey
				receivedBucket := h.s3Client.(*mockS3Client).mockDeleteObjectsBucket

				if receivedKey != key {
					t.Errorf("incorrect key, received: %s, expected: %s", receivedKey, key)
				}

				if receivedBucket != bucket {
					t.Errorf("incorrect bucket name, received: %s, expected: %s", receivedBucket, bucket)
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

func Test_getQueryResultDocuments(t *testing.T) {
	tests := []struct {
		description               string
		state                     string
		executionID               string
		mockGetQueryResultsOutput *athena.GetQueryResultsOutput
		mockGetQueryResultsError  error
		documents                 []pars.Document
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
			description: "successful get query result documents",
			state:       "SUCCEEDED",
			executionID: "execution_id",
			mockGetQueryResultsOutput: &athena.GetQueryResultsOutput{
				ResultSet: &athena.ResultSet{
					Rows: []*athena.Row{
						{
							Data: []*athena.Datum{
								{
									VarCharValue: aws.String("id"),
								},
								{
									VarCharValue: aws.String("key.json"),
								},
								{
									VarCharValue: aws.String("bucket"),
								},
							},
						},
					},
				},
			},
			mockGetQueryResultsError: nil,
			documents: []pars.Document{
				{
					FileKey:    "key.json",
					FileBucket: "bucket",
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

			documents, err := h.getQueryResultDocuments(context.Background(), test.state, test.executionID)

			if err != nil {
				if err.Error() != test.error.Error() {
					t.Errorf("incorrect error, received: %s, expected: %s", err.Error(), test.error.Error())
				}
			} else {
				received := documents[0]
				expected := test.documents[0]

				if received.FileKey != expected.FileKey {
					t.Errorf("incorrect file key, received: %s, expected: %s", received.FileKey, expected.FileKey)
				}

				if received.FileBucket != expected.FileBucket {
					t.Errorf("incorrect file bucket, received: %s, expected: %s", received.FileBucket, expected.FileBucket)
				}
			}
		})
	}
}

func Test_getQueryResultKeys(t *testing.T) {
	tests := []struct {
		description               string
		state                     string
		executionID               string
		mockGetQueryResultsOutput *athena.GetQueryResultsOutput
		mockGetQueryResultsError  error
		keys                      []string
		error                     error
	}{
		{
			description:               "non-success state received",
			state:                     "NOT_SUCCEEDED",
			executionID:               "execution_id",
			mockGetQueryResultsOutput: nil,
			mockGetQueryResultsError:  nil,
			keys:                      nil,
			error:                     errors.New("incorrect query state [NOT_SUCCEEDED]"),
		},
		{
			description:               "error getting query results",
			state:                     "SUCCEEDED",
			executionID:               "execution_id",
			mockGetQueryResultsOutput: nil,
			mockGetQueryResultsError:  errors.New("mock get query results error"),
			keys:                      nil,
			error:                     errors.New("mock get query results error"),
		},
		{
			description: "successful get query result documents",
			state:       "SUCCEEDED",
			executionID: "execution_id",
			mockGetQueryResultsOutput: &athena.GetQueryResultsOutput{
				ResultSet: &athena.ResultSet{
					Rows: []*athena.Row{
						{
							Data: []*athena.Datum{
								{
									VarCharValue: aws.String("document_id"),
								},
								{
									VarCharValue: aws.String("page_id"),
								},
								{
									VarCharValue: aws.String("line_id"),
								},
								{
									VarCharValue: aws.String("coordinates_id"),
								},
							},
						},
					},
				},
			},
			mockGetQueryResultsError: nil,
			keys: []string{
				"documents/document_id.json",
				"pages/page_id.json",
				"lines/line_id.json",
				"coordinates/coordinates_id.json",
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

			keys, err := h.getQueryResultKeys(context.Background(), test.state, test.executionID)

			if err != nil {
				if err.Error() != test.error.Error() {
					t.Errorf("incorrect error, received: %s, expected: %s", err.Error(), test.error.Error())
				}
			} else {
				if len(keys) != len(test.keys) {
					t.Errorf("incorrect keys count, received: %d, expected: %d", len(keys), len(test.keys))
				}

				received := keys[0]
				expected := test.keys[0]
				if received != expected {
					t.Errorf("incorrect key, received: %s, expected: %s", received, expected)
				}
			}
		})
	}
}

func Test_addFolder(t *testing.T) {
	tests := []struct {
		description        string
		mockPutObjectError error
		error              error
	}{
		{
			description:        "error putting folder",
			mockPutObjectError: errors.New("mock put object error"),
			error:              errors.New("mock put object error"),
		},
		{
			description:        "successful invocation",
			mockPutObjectError: nil,
			error:              nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			h := &help{
				databaseBucket: "bucket",
				s3Client: &mockS3Client{
					mockPutObjectOutput: nil,
					mockPutObjectError:  test.mockPutObjectError,
				},
			}

			expectedBucket := "bucket"
			expectedKey := "folder"

			err := h.addFolder(context.Background(), expectedKey)

			if err != nil {
				if err.Error() != test.error.Error() {
					t.Errorf("incorrect error, received: %s, expected: %s", err.Error(), test.error.Error())
				}
			} else {
				receivedBucket := h.s3Client.(*mockS3Client).mockPutObjectBucket
				receivedKey := h.s3Client.(*mockS3Client).mockPutObjectKey

				if receivedBucket != expectedBucket {
					t.Errorf("incorrect bucket name, received: %s, expected: %s", receivedBucket, expectedBucket)
				}

				if receivedKey != expectedKey {
					t.Errorf("incorrect key, received: %s, expected: %s", receivedKey, expectedKey)
				}
			}
		})
	}
}

func Test_startCrawler(t *testing.T) {
	tests := []struct {
		description           string
		mockStartCrawlerError error
		error                 error
	}{
		{
			description:           "start crawler error",
			mockStartCrawlerError: errors.New("mock start crawler error"),
			error:                 errors.New("mock start crawler error"),
		},
		{
			description:           "successful invocation",
			mockStartCrawlerError: nil,
			error:                 nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			expectedName := "crawler"

			h := &help{
				crawlerName: expectedName,
				glueClient: &mockGlueClient{
					mockStartCrawlerError: test.mockStartCrawlerError,
				},
			}

			err := h.startCrawler(context.Background())

			if err != nil {
				if err.Error() != test.error.Error() {
					t.Errorf("incorrect error, received: %s, expected: %s", err.Error(), test.error.Error())
				}
			} else {
				receivedName := h.glueClient.(*mockGlueClient).mockStartCrawlerName

				if receivedName != expectedName {
					t.Errorf("incorrect crawler name, received: %s, expected: %s", receivedName, expectedName)
				}
			}
		})
	}
}
