package infra

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/aws/aws-sdk-go/service/s3"
)

const nameRoot = "cheesesteakio"

var _ Infrastructurer = &Client{}

// Client implements the infra.Infrastructurer methods using AWS.
type Client struct {
	s3Client         s3Client
	documentDBClient documentDBClient
}

type s3Client interface {
	CreateBucket(input *s3.CreateBucketInput) (*s3.CreateBucketOutput, error)
	DeleteBucket(input *s3.DeleteBucketInput) (*s3.DeleteBucketOutput, error)
}

type documentDBClient interface {
	CreateDBCluster(input *docdb.CreateDBClusterInput) (*docdb.CreateDBClusterOutput, error)
	DeleteDBCluster(input *docdb.DeleteDBClusterInput) (*docdb.DeleteDBClusterOutput, error)
}

// New generates a Client pointer instance with a AWS resource clients
// for S3 and DocumentDB.
func New() *Client {
	newSession := session.Must(session.NewSession())
	s3Service := s3.New(newSession)
	documentDBService := docdb.New(newSession)

	return &Client{
		s3Client:         s3Service,
		documentDBClient: documentDBService,
	}
}

// CreateFilesystem implements the infra.Infrastructurer.CreateFilesystem
// interface method using AWS S3.
func (c *Client) CreateFilesystem(ctx context.Context, accountID string) error {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(entityName("bucket", accountID)),
	}

	_, err := c.s3Client.CreateBucket(input)
	if err != nil {
		return &ErrorCreateFilesystem{err: err}
	}

	return nil
}

// DeleteFilesystem implements the infra.Infrastructurer.DeleteFilesystem
// interface method using AWS S3.
func (c *Client) DeleteFilesystem(ctx context.Context, accountID string) error {
	input := &s3.DeleteBucketInput{
		Bucket: aws.String(entityName("bucket", accountID)),
	}

	_, err := c.s3Client.DeleteBucket(input)
	if err != nil {
		return &ErrorDeleteFilesystem{err: err}
	}

	return nil
}

// CreateDatabase implements the infra.Infrastructurer.CreateDatabase
// interface method using AWS DocumentDB.
func (c *Client) CreateDatabase(ctx context.Context, accountID string) error {
	input := &docdb.CreateDBClusterInput{
		DBClusterIdentifier: aws.String(entityName("cluster", accountID)),
		Engine:              aws.String("docdb"),
		MasterUsername:      aws.String("cheesesteakio-cluster-admin"), // TEMP
		MasterUserPassword:  aws.String("passwordHere"),                // TEMP
	}

	_, err := c.documentDBClient.CreateDBCluster(input)
	if err != nil {
		return &ErrorCreateDatabase{err: err}
	}

	return nil
}

// DeleteDatabase implements the infra.Infrastructurer.DeleteDatabase
// interface method using AWS DocumentDB.
func (c *Client) DeleteDatabase(ctx context.Context, accountID string) error {
	input := &docdb.DeleteDBClusterInput{
		DBClusterIdentifier: aws.String(entityName("cluster", accountID)),
	}

	_, err := c.documentDBClient.DeleteDBCluster(input)
	if err != nil {
		return &ErrorDeleteDatabase{err: err}
	}

	return nil
}

func entityName(entity, accountID string) string {
	return fmt.Sprintf("%s-%s-%s", nameRoot, entity, accountID)
}
