package fs

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var _ Filesystemer = &Client{}

// Client implements the fs.Filesystemer methods using AWS S3.
type Client struct {
	s3Client s3Client
}

type s3Client interface {
	PutObjectRequest(input *s3.PutObjectInput) (req *request.Request, output *s3.PutObjectOutput)
	DeleteObjects(input *s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error)
}

// New generates a Client pointer instance with an AWS S3 client.
func New() (*Client, error) {
	newSession, err := session.NewSession()
	if err != nil {
		return nil, &ErrorNewClient{err: err}
	}
	s3Client := s3.New(newSession)

	return &Client{
		s3Client: s3Client,
	}, nil
}

// GenerateUploadURL implements the fs.Filesystemer.GenerateUploadURL method
// using presigned S3 URLs.
func (c *Client) GenerateUploadURL(ctx context.Context, accountID string, filename string) (string, error) {
	putRequest, _ := c.s3Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(mainBucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", accountID, filename)),
	})

	urlString, err := putRequest.Presign(15 * time.Minute)
	if err != nil {
		return "", &ErrorPresignURL{err: err}
	}

	return urlString, nil
}

// DeleteFiles implements the fs.Filesystemer.DeleteFiles method.
func (c *Client) DeleteFiles(ctx context.Context, accountID string, filenames []string) error {
	objects := make([]*s3.ObjectIdentifier, len(filenames))
	for i, filename := range filenames {
		objects[i] = &s3.ObjectIdentifier{
			Key: aws.String(fmt.Sprintf("%s/%s", accountID, filename)),
		}
	}

	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(mainBucket),
		Delete: &s3.Delete{
			Objects: objects,
		},
	}

	_, err := c.s3Client.DeleteObjects(input)
	if err != nil {
		return &ErrorDeleteObjects{err: err}
	}

	return nil
}
