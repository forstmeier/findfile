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
	GetObjectRequest(input *s3.GetObjectInput) (req *request.Request, output *s3.GetObjectOutput)
	ListObjectsV2(input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error)
	DeleteObjects(input *s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error)
}

// New generates a Client pointer instance with an AWS S3 client.
func New(newSession *session.Session) *Client {
	s3Client := s3.New(newSession)

	return &Client{
		s3Client: s3Client,
	}
}

// GenerateUploadURL implements the fs.Filesystemer.GenerateUploadURL method
// using presigned S3 URLs.
func (c *Client) GenerateUploadURL(ctx context.Context, accountID string, fileInfo FileInfo) (string, error) {
	putRequest, _ := c.s3Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(fileInfo.Filepath),
		Key:    aws.String(fmt.Sprintf("%s/%s", accountID, fileInfo.Filename)),
	})

	urlString, err := putRequest.Presign(15 * time.Minute)
	if err != nil {
		return "", &ErrorPresignURL{
			err:    err,
			action: "upload",
		}
	}

	return urlString, nil
}

// GenerateDownloadURL implements the fs.Filesystemer.GenerateDownloadURL method
// using presigned S3 URLs.
func (c *Client) GenerateDownloadURL(ctx context.Context, accountID string, fileInfo FileInfo) (string, error) {
	getRequest, _ := c.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(fileInfo.Filepath),
		Key:    aws.String(fmt.Sprintf("%s/%s", accountID, fileInfo.Filename)),
	})

	urlString, err := getRequest.Presign(15 * time.Minute)
	if err != nil {
		return "", &ErrorPresignURL{
			err:    err,
			action: "download",
		}
	}

	return urlString, nil
}

// ListFilesByAccountID implements the fs.Filesystemer.ListFilesByAccountID
// method.
func (c *Client) ListFilesByAccountID(ctx context.Context, filepath, accountID string) ([]FileInfo, error) {
	var results []FileInfo
	var continuationToken string
	for {
		input := &s3.ListObjectsV2Input{
			Bucket:            aws.String(filepath),
			Prefix:            aws.String(accountID),
			ContinuationToken: aws.String(continuationToken),
		}

		output, err := c.s3Client.ListObjectsV2(input)
		if err != nil {
			return nil, &ErrorListObjects{err: err}
		}

		for _, content := range output.Contents {
			results = append(results, FileInfo{
				Filepath: *output.Name,
				Filename: *content.Key,
			})
		}

		if *output.IsTruncated {
			continuationToken = *output.NextContinuationToken
		} else {
			break
		}
	}

	return results, nil
}

// DeleteFiles implements the fs.Filesystemer.DeleteFiles method.
func (c *Client) DeleteFiles(ctx context.Context, accountID string, filesInfo []FileInfo) error {
	chunkSize := 1000 // S3 max delete objects count
	for i := 0; i < len(filesInfo); i += chunkSize {
		end := i + chunkSize
		if end > len(filesInfo) {
			end = len(filesInfo)
		}

		filesInfoSubset := filesInfo[i:end]

		objects := make([]*s3.ObjectIdentifier, len(filesInfoSubset))
		for i, fileInfo := range filesInfoSubset {
			objects[i] = &s3.ObjectIdentifier{
				Key: aws.String(fmt.Sprintf("%s/%s", accountID, fileInfo.Filename)),
			}
		}

		input := &s3.DeleteObjectsInput{
			Bucket: aws.String(filesInfo[0].Filepath),
			Delete: &s3.Delete{
				Objects: objects,
			},
		}

		_, err := c.s3Client.DeleteObjects(input)
		if err != nil {
			return &ErrorDeleteObjects{err: err}
		}
	}

	return nil
}
