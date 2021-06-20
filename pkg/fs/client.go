package fs

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Client implements the fs.Filesystemer methods using AWS S3.
type Client struct {
	s3Client s3Client
}

type s3Client interface {
	PutObjectRequest(input *s3.PutObjectInput) (req *request.Request, output *s3.PutObjectOutput)
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
func (c *Client) GenerateUploadURL(filename string) (string, error) {
	putRequest, _ := c.s3Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(mainBucket),
		Key:    aws.String(filename),
	})

	urlString, err := putRequest.Presign(15 * time.Minute)
	if err != nil {
		return "", &ErrorPresignURL{err: err}
	}

	return urlString, nil
}
