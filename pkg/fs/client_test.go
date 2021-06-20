package fs

import (
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
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

			presignedURL, err := client.GenerateUploadURL("file.jpg")

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
				if !strings.Contains(presignedURL, "https://cheesesteakstorage-main.s3.amazonaws.com/file.jpg") {
					t.Errorf("incorrect presigned url, received: %s", presignedURL)
				}
			}
		})
	}
}
