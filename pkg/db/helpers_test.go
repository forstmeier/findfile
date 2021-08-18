package db

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/s3"
)

type mockS3Client struct {
	mockListObjectsV2Output *s3.ListObjectsV2Output
	mockListObjectsV2Error  error
	mockPutObjectOutput     *s3.PutObjectOutput
	mockPutObjectError      error
	mockPutObjectBucket     string
	mockPutObjectKey        string
}

func (m *mockS3Client) ListObjectsV2(input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
	return m.mockListObjectsV2Output, m.mockListObjectsV2Error
}

func (m *mockS3Client) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	m.mockPutObjectBucket = *input.Bucket
	m.mockPutObjectKey = *input.Key

	return m.mockPutObjectOutput, m.mockPutObjectError
}

func (m *mockS3Client) DeleteObjects(input *s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error) {
	return nil, nil
}

type mockAthenaClient struct {
	mockStartQueryExecutionOutput *athena.StartQueryExecutionOutput
	mockStartQueryExecutionError  error
	mockGetQueryExecutionOutput   *athena.GetQueryExecutionOutput
	mockGetQueryExecutionError    error
}

func (m *mockAthenaClient) StartQueryExecution(input *athena.StartQueryExecutionInput) (*athena.StartQueryExecutionOutput, error) {
	return m.mockStartQueryExecutionOutput, m.mockStartQueryExecutionError
}

func (m *mockAthenaClient) GetQueryExecution(input *athena.GetQueryExecutionInput) (*athena.GetQueryExecutionOutput, error) {
	return m.mockGetQueryExecutionOutput, m.mockGetQueryExecutionError
}

func (m *mockAthenaClient) GetQueryResults(input *athena.GetQueryResultsInput) (*athena.GetQueryResultsOutput, error) {
	return nil, nil
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

	bucketName := "bucketName"
	body := struct {
		name string
	}{
		name: "name",
	}
	key := "file.json"

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				bucketName: bucketName,
				s3Client: &mockS3Client{
					mockPutObjectError: test.mockPutObjectError,
				},
			}

			err := client.uploadObject(context.Background(), body, key, "entity")

			if err != nil {
				if err.Error() != test.error.Error() {
					t.Errorf("incorrect error, received: %s, expected: %s", err.Error(), test.error.Error())
				}

			} else {
				receivedBucket := client.s3Client.(*mockS3Client).mockPutObjectBucket
				receivedKey := client.s3Client.(*mockS3Client).mockPutObjectKey

				if receivedBucket != bucketName {
					t.Errorf("incorrect bucket name, received: %s, expected: %s", receivedBucket, bucketName)
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
						Key: aws.String("prefix/file.json"),
					},
				},
				IsTruncated: aws.Bool(false),
			},
			mockListObjectsV2Error: nil,
			results:                []string{"prefix/file.json"},
			error:                  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				s3Client: &mockS3Client{
					mockListObjectsV2Output: test.mockListObjectsV2Output,
					mockListObjectsV2Error:  test.mockListObjectsV2Error,
				},
			}

			results, err := client.listDocumentKeys(context.Background(), "bucket", "prefix")

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
			client := &Client{
				athenaClient: &mockAthenaClient{
					mockStartQueryExecutionOutput: test.mockStartQueryExecutionOutput,
					mockStartQueryExecutionError:  test.mockStartQueryExecutionError,
					mockGetQueryExecutionOutput:   test.mockGetQueryExecutionOutput,
					mockGetQueryExecutionError:    test.mockGetQueryExecutionError,
				},
			}

			executionID, state, err := client.executeQuery(context.Background(), []byte("query"))

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
