package fs

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func TestNew(t *testing.T) {
	client, err := New()

	if err != nil {
		t.Errorf("error received creating filesystem client, %s:", err.Error())
	}

	if client == nil {
		t.Error("error creating filesystem client")
	}
}

func TestGenerateUploadURL(t *testing.T) {
	tests := []struct {
		description string
		s3Client    *s3.S3
		error       error
	}{
		{
			description: "error presigning url",
			s3Client:    s3.New(session.Must(session.NewSession())),
			error:       &ErrorPresignURL{},
		},
		{
			description: "successful generate presigned url invocation",
			s3Client: s3.New(session.New(&aws.Config{
				Region: aws.String("us-east-1"),
			})),
			error: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				s3Client: test.s3Client,
			}

			presignedURL, err := client.GenerateUploadURL(context.Background(), "account_id", "file.jpg")

			if err != nil {
				switch test.error.(type) {
				case *ErrorPresignURL:
					var testError *ErrorPresignURL
					if !errors.As(err, &testError) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, testError)
					}
				default:
					t.Fatalf("unexpected error type: %v", err)
				}
			} else {
				if !strings.Contains(presignedURL, "https://cheesesteakstorage-main.s3.amazonaws.com/account_id/file.jpg") {
					t.Errorf("incorrect presigned url, received: %s", presignedURL)
				}
			}
		})
	}
}

type mockS3Client struct {
	deleteObjectsError error
}

func (m *mockS3Client) PutObjectRequest(input *s3.PutObjectInput) (req *request.Request, output *s3.PutObjectOutput) {
	return nil, nil
}

func (m *mockS3Client) DeleteObjects(input *s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error) {
	return nil, m.deleteObjectsError
}

func TestDeleteFiles(t *testing.T) {
	tests := []struct {
		description        string
		deleteObjectsError error
		error              error
	}{
		{
			description:        "s3 client delete object error",
			deleteObjectsError: errors.New("mock delete object error"),
			error:              &ErrorDeleteObjects{},
		},
		{
			description:        "successful delete files invocation",
			deleteObjectsError: nil,
			error:              nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				s3Client: &mockS3Client{
					deleteObjectsError: test.deleteObjectsError,
				},
			}

			err := client.DeleteFiles(context.Background(), "account_id", []string{"filename.jpg"})

			if err != nil {
				switch test.error.(type) {
				case *ErrorDeleteObjects:
					var testError *ErrorDeleteObjects
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
