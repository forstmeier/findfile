package db

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/glue"
)

type mockGlueClient struct {
	mockCreatePartitionOutputError error
	mockDeletePartitionOutputError error
}

func (m *mockGlueClient) CreatePartition(input *glue.CreatePartitionInput) (*glue.CreatePartitionOutput, error) {
	return nil, m.mockCreatePartitionOutputError
}

func (m *mockGlueClient) DeletePartition(input *glue.DeletePartitionInput) (*glue.DeletePartitionOutput, error) {
	return nil, m.mockDeletePartitionOutputError
}

func TestNewPartitioner(t *testing.T) {
	client := NewPartitionerClient(session.New(), "bucket", "table", "database", "catalogID")
	if client == nil {
		t.Error("error partition client")
	}
}

func TestAddPartition(t *testing.T) {
	tests := []struct {
		description                    string
		mockPutObjectError             error
		mockCreatePartitionOutputError error
		error                          error
	}{
		{
			description:                    "error uploading folder",
			mockPutObjectError:             errors.New("mock put object error"),
			mockCreatePartitionOutputError: nil,
			error:                          &ErrorPutObject{},
		},
		{
			description:                    "error creating partition",
			mockPutObjectError:             nil,
			mockCreatePartitionOutputError: errors.New("mock create partition error"),
			error:                          &ErrorCreatePartition{},
		},
		{
			description:                    "successful add partition invocation",
			mockPutObjectError:             nil,
			mockCreatePartitionOutputError: nil,
			error:                          nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &PartitionerClient{
				s3Client: &mockS3Client{
					mockPutObjectOutput: nil,
					mockPutObjectError:  test.mockPutObjectError,
				},
				glueClient: &mockGlueClient{
					mockCreatePartitionOutputError: test.mockCreatePartitionOutputError,
				},
			}

			err := client.AddPartition(context.Background(), "account_id")

			if err != nil {
				switch e := test.error.(type) {
				case *ErrorPutObject:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
					}
				case *ErrorCreatePartition:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
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

func TestRemovePartition(t *testing.T) {
	tests := []struct {
		description                    string
		mockDeletePartitionOutputError error
		error                          error
	}{
		{
			description:                    "error deleting partition",
			mockDeletePartitionOutputError: errors.New("mock delete partition error"),
			error:                          &ErrorDeletePartition{},
		},
		{
			description:                    "successful remove partition invocation",
			mockDeletePartitionOutputError: nil,
			error:                          nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &PartitionerClient{
				glueClient: &mockGlueClient{
					mockDeletePartitionOutputError: test.mockDeletePartitionOutputError,
				},
			}

			err := client.RemovePartition(context.Background(), "account_id")

			if err != nil {
				switch e := test.error.(type) {
				case *ErrorDeletePartition:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
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
