package fs

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var _ Filesystemer = &Client{}

// Client implements the fs.Filesystemr methods using AWS S3.
type Client struct {
	helper helper
}

// New generates a fs.Client pointer instance with an AWS S3 client.
func New(newSession *session.Session, topicARN string) *Client {
	return &Client{
		helper: &help{
			topicARN: topicARN,
			s3Client: s3.New(newSession),
		},
	}
}

// CreateFileWatcher implements the fs.Filesystemer.CreateFileWatcher method.
func (c *Client) CreateFileWatcher(ctx context.Context, path string) error {
	if err := c.helper.addOrRemoveNotification(ctx, path, true); err != nil {
		return &ErrorAddNotification{
			err: err,
		}
	}

	return nil
}

// DeleteFileWatcher implements the fs.Filesystemer.DeleteFileWatcher method.
func (c *Client) DeleteFileWatcher(ctx context.Context, path string) error {
	if err := c.helper.addOrRemoveNotification(ctx, path, false); err != nil {
		return &ErrorRemoveNotification{
			err: err,
		}
	}

	return nil
}
