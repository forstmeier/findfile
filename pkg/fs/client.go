package fs

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var _ Filesystemer = &Client{}

// Client implements the fs.Filesystemer methods using AWS S3.
type Client struct {
	s3Client s3Client
}

type s3Client interface {
	ListObjectsV2(input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error)
}

// New generates a fs.Client pointer instance with AWS S3.
func New(newSession *session.Session) *Client {
	return &Client{
		s3Client: s3.New(newSession),
	}
}

// ListFiles implements the fs.Filesystemer.ListFiles method
// using S3.
func (c *Client) ListFiles(ctx context.Context, bucket string) ([]string, error) {
	keys := []string{}

	var startKey *string
	for {
		output, err := c.s3Client.ListObjectsV2(&s3.ListObjectsV2Input{
			Bucket:     &bucket,
			StartAfter: startKey,
		})

		if err != nil {
			return nil, &ListObjectsError{
				err: err,
			}
		}

		contents := output.Contents
		for _, content := range contents {
			keys = append(keys, *content.Key)
		}
		startKey = output.StartAfter

		if len(contents) < 1000 { // max keys returned by list method
			break
		}
	}

	return keys, nil
}
