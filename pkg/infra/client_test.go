package infra

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/aws/aws-sdk-go/service/s3"
)

type mockS3Client struct {
	createBucketError error
	deleteBucketError error
}

func (m *mockS3Client) CreateBucket(input *s3.CreateBucketInput) (*s3.CreateBucketOutput, error) {
	return nil, m.createBucketError
}

func (m *mockS3Client) DeleteBucket(input *s3.DeleteBucketInput) (*s3.DeleteBucketOutput, error) {
	return nil, m.deleteBucketError
}

type mockDocumentDBClient struct {
	createDocumentDBError error
	deleteDocumentDBError error
}

func (m *mockDocumentDBClient) CreateDBCluster(input *docdb.CreateDBClusterInput) (*docdb.CreateDBClusterOutput, error) {
	return nil, m.createDocumentDBError
}

func (m *mockDocumentDBClient) DeleteDBCluster(input *docdb.DeleteDBClusterInput) (*docdb.DeleteDBClusterOutput, error) {
	return nil, m.deleteDocumentDBError
}

func TestNew(t *testing.T) {
	client := New()

	if client == nil {
		t.Error("error creating database client")
	}
}

func TestCreateFilesystem(t *testing.T) {
	tests := []struct {
		description       string
		createBucketError error
		error             error
	}{
		{
			description:       "error creating s3 bucket",
			createBucketError: errors.New("mock create bucket error"),
			error:             &ErrorCreateFilesystem{},
		},
		{
			description:       "successful create filesystem invocation",
			createBucketError: nil,
			error:             nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				s3Client: &mockS3Client{
					createBucketError: test.createBucketError,
				},
			}

			err := client.CreateFilesystem(context.Background(), "account_id")

			if err != nil {
				switch test.error.(type) {
				case *ErrorCreateFilesystem:
					var testError *ErrorCreateFilesystem
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

func TestDeleteFilesystem(t *testing.T) {
	tests := []struct {
		description       string
		deleteBucketError error
		error             error
	}{
		{
			description:       "error deleting s3 bucket",
			deleteBucketError: errors.New("mock delete bucket error"),
			error:             &ErrorDeleteFilesystem{},
		},
		{
			description:       "successful delete filesystem invocation",
			deleteBucketError: nil,
			error:             nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				s3Client: &mockS3Client{
					deleteBucketError: test.deleteBucketError,
				},
			}

			err := client.DeleteFilesystem(context.Background(), "account_id")

			if err != nil {
				switch test.error.(type) {
				case *ErrorDeleteFilesystem:
					var testError *ErrorDeleteFilesystem
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

func TestCreateDatabase(t *testing.T) {
	tests := []struct {
		description           string
		createDocumentDBError error
		error                 error
	}{
		{
			description:           "error creating documentdb cluster bucket",
			createDocumentDBError: errors.New("mock create cluster error"),
			error:                 &ErrorCreateDatabase{},
		},
		{
			description:           "successful create database invocation",
			createDocumentDBError: nil,
			error:                 nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				documentDBClient: &mockDocumentDBClient{
					createDocumentDBError: test.createDocumentDBError,
				},
			}

			err := client.CreateDatabase(context.Background(), "account_id")

			if err != nil {
				switch test.error.(type) {
				case *ErrorCreateDatabase:
					var testError *ErrorCreateDatabase
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

func TestDeleteDatabase(t *testing.T) {
	tests := []struct {
		description           string
		deleteDocumentDBError error
		error                 error
	}{
		{
			description:           "error deleting documentdb cluster bucket",
			deleteDocumentDBError: errors.New("mock delete cluster error"),
			error:                 &ErrorDeleteDatabase{},
		},
		{
			description:           "successful delete database invocation",
			deleteDocumentDBError: nil,
			error:                 nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				documentDBClient: &mockDocumentDBClient{
					deleteDocumentDBError: test.deleteDocumentDBError,
				},
			}

			err := client.DeleteDatabase(context.Background(), "account_id")

			if err != nil {
				switch test.error.(type) {
				case *ErrorDeleteDatabase:
					var testError *ErrorDeleteDatabase
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
