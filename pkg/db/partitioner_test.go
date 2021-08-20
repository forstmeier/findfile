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
	client := NewPartitionerClient(session.New(), "table", "database", "catalogID")
	if client == nil {
		t.Error("error partition client")
	}
}

func TestAddPartition(t *testing.T) {
	tests := []struct {
		description                    string
		mockCreatePartitionOutputError error
		error                          error
	}{
		{
			description:                    "error creating partition",
			mockCreatePartitionOutputError: errors.New("mock create partition error"),
			error:                          &ErrorCreatePartition{},
		},
		{
			description:                    "successful add partition invocation",
			mockCreatePartitionOutputError: nil,
			error:                          nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &PartitionerClient{
				glueClient: &mockGlueClient{
					mockCreatePartitionOutputError: test.mockCreatePartitionOutputError,
				},
			}

			err := client.AddPartition(context.Background(), "account_id")

			if err != nil {
				switch e := test.error.(type) {
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
