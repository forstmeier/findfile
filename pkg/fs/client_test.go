package fs

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type mockS3Client struct {
	mockListObjectsV2Output *s3.ListObjectsV2Output
	mockListObjectsV2Error  error
}

func (m *mockS3Client) ListObjectsV2(input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
	return m.mockListObjectsV2Output, m.mockListObjectsV2Error
}

func TestListFiles(t *testing.T) {
	tests := []struct {
		description             string
		mockListObjectsV2Output *s3.ListObjectsV2Output
		mockListObjectsV2Error  error
		files                   []string
		error                   error
	}{
		{
			description:             "error listing files",
			mockListObjectsV2Output: nil,
			mockListObjectsV2Error:  errors.New("mock list objects error"),
			files:                   nil,
			error:                   &ListObjectsError{},
		},
		{
			description: "successful invocation",
			mockListObjectsV2Output: &s3.ListObjectsV2Output{
				Contents: []*s3.Object{
					{
						Key: aws.String("key.jpeg"),
					},
				},
			},
			mockListObjectsV2Error: nil,
			files:                  []string{"key.jpeg"},
			error:                  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			s3Client := &mockS3Client{
				mockListObjectsV2Output: test.mockListObjectsV2Output,
				mockListObjectsV2Error:  test.mockListObjectsV2Error,
			}

			client := &Client{
				s3Client: s3Client,
			}

			files, err := client.ListFiles(context.Background(), "bucket")

			if err != nil {
				switch e := test.error.(type) {
				case *ListObjectsError:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
					}
				default:
					t.Fatalf("unexpected error type: %v", err)
				}
			} else {
				if !reflect.DeepEqual(files, test.files) {
					t.Errorf("incorrect output, received: %v, expected: %v", files, test.files)
				}
			}
		})
	}
}
