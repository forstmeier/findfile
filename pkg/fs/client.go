package fs

import (
	"log"
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
	GetObjectRequest(input *s3.GetObjectInput) (req *request.Request, output *s3.GetObjectOutput)
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
	getRequest, _ := c.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(mainBucket),
		Key:    aws.String(filename),
	})

	urlString, err := getRequest.Presign(15 * time.Minute)
	if err != nil {
		log.Println("ERROR:", err)
		return "", &ErrorPresignURL{err: err}
	}

	return urlString, nil
}
