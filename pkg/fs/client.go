package fs

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
)

var _ Filesystemer = &Client{}

// Client implements the fs.Filesystemr methods using AWS S3.
type Client struct {
	helper helper
}

// New generates a fs.Client pointer instance with AWS S3 and
// AWS SNS clients.
func New(newSession *session.Session, topicARN, configurationID string) *Client {
	return &Client{
		helper: &help{
			configurationID: configurationID,
			topicARN:        topicARN,
			s3Client:        s3.New(newSession),
			snsClient:       sns.New(newSession),
		},
	}
}

// CreateFileWatcher implements the fs.Filesystemer.CreateFileWatcher
// method using AWS S3 and AWS SNS.
func (c *Client) CreateFileWatcher(ctx context.Context, path string) error {
	if err := c.helper.addOrRemoveNotification(ctx, path, true); err != nil {
		return &ErrorAddNotification{
			err: err,
		}
	}

	if err := c.helper.addOrRemoveTopicPolicyBucketARN(ctx, path, true); err != nil {
		return &ErrorAddTopicPolicyBucketARN{
			err: err,
		}
	}

	return nil
}

// DeleteFileWatcher implements the fs.Filesystemer.DeleteFileWatcher
// method using AWS S3 and AWS SNS.
func (c *Client) DeleteFileWatcher(ctx context.Context, path string) error {
	if err := c.helper.addOrRemoveNotification(ctx, path, false); err != nil {
		return &ErrorRemoveNotification{
			err: err,
		}
	}

	if err := c.helper.addOrRemoveTopicPolicyBucketARN(ctx, path, false); err != nil {
		return &ErrorRemoveTopicPolicyBucketARN{
			err: err,
		}
	}

	return nil
}
