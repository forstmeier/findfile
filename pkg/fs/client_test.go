package fs

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func TestNew(t *testing.T) {
	client := New(session.New())

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

			fileInfo := FileInfo{
				Filepath: "bucket",
				Filename: "file.jpg",
			}

			presignedURL, err := client.GenerateUploadURL(context.Background(), "account_id", fileInfo)

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
				if !strings.Contains(presignedURL, "https://bucket.s3.amazonaws.com/account_id/files/file.jpg") {
					t.Errorf("incorrect presigned url, received: %s", presignedURL)
				}
			}
		})
	}
}

func TestGenerateDownloadURL(t *testing.T) {
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

			fileInfo := FileInfo{
				Filepath: "bucket",
				Filename: "file.jpg",
			}

			presignedURL, err := client.GenerateDownloadURL(context.Background(), "account_id", fileInfo)

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
				if !strings.Contains(presignedURL, "https://bucket.s3.amazonaws.com/account_id/files/file.jpg") {
					t.Errorf("incorrect presigned url, received: %s", presignedURL)
				}
			}
		})
	}
}

type mockS3Client struct {
	deleteObjectsError  error
	listObjectsV2Output *s3.ListObjectsV2Output
	listObjectsV2Error  error
}

func (m *mockS3Client) PutObjectRequest(input *s3.PutObjectInput) (req *request.Request, output *s3.PutObjectOutput) {
	return nil, nil
}

func (m *mockS3Client) GetObjectRequest(input *s3.GetObjectInput) (req *request.Request, output *s3.GetObjectOutput) {
	return nil, nil
}

func (m *mockS3Client) ListObjectsV2(input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
	return m.listObjectsV2Output, m.listObjectsV2Error
}

func (m *mockS3Client) DeleteObjects(input *s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error) {
	return nil, m.deleteObjectsError
}

func TestListFilesByAccountID(t *testing.T) {
	tests := []struct {
		description         string
		listObjectsV2Output *s3.ListObjectsV2Output
		listObjectsV2Error  error
		output              []FileInfo
		error               error
	}{
		{
			description:         "error listing files in s3",
			listObjectsV2Output: nil,
			listObjectsV2Error:  errors.New("mock list objects error"),
			output:              nil,
			error:               &ErrorListObjects{},
		},
		{
			description: "successful list files invocation",
			listObjectsV2Output: &s3.ListObjectsV2Output{
				Name: aws.String("bucket"),
				Contents: []*s3.Object{
					{
						Key: aws.String("file.jpg"),
					},
				},
				IsTruncated: aws.Bool(false),
			},
			listObjectsV2Error: nil,
			output: []FileInfo{
				{
					Filepath: "bucket",
					Filename: "file.jpg",
				},
			},
			error: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				s3Client: &mockS3Client{
					listObjectsV2Output: test.listObjectsV2Output,
					listObjectsV2Error:  test.listObjectsV2Error,
				},
			}

			output, err := client.ListFilesByAccountID(context.Background(), "bucket", "account_id")

			if err != nil {
				switch test.error.(type) {
				case *ErrorListObjects:
					var testError *ErrorListObjects
					if !errors.As(err, &testError) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, testError)
					}
				default:
					t.Fatalf("unexpected error type: %v", err)
				}
			} else {
				if !reflect.DeepEqual(output, test.output) {
					t.Errorf("incorrect output, received: %+v, expected: %+v", output, test.output)
				}
			}
		})
	}
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

			filesInfo := []FileInfo{
				{
					Filepath: "bucket",
					Filename: "file.jpg",
				},
			}

			err := client.DeleteFiles(context.Background(), "account_id", filesInfo)

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
