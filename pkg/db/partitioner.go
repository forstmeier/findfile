package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/aws/aws-sdk-go/service/s3"
)

type glueClient interface {
	StartCrawler(input *glue.StartCrawlerInput) (*glue.StartCrawlerOutput, error)
	CreatePartition(input *glue.CreatePartitionInput) (*glue.CreatePartitionOutput, error)
	DeletePartition(input *glue.DeletePartitionInput) (*glue.DeletePartitionOutput, error)
}

// Partitioner defines the methods for managing partitions in AwS Athena.
type Partitioner interface {
	AddPartition(ctx context.Context, accountID string) error
	RemovePartition(ctx context.Context, accountID string) error
}

// PartitionerClient implements the db.Partitioner methods using AWS Athena.
type PartitionerClient struct {
	bucketName   string
	tableName    string
	databaseName string
	catalogID    string
	crawlerName  string
	s3Client     s3Client
	glueClient   glueClient
}

var paths = []string{
	"documents",
	"pages",
	"lines",
	"coordinates",
}

// NewPartitionerClient returns a db.Partitioner pointer instance.
func NewPartitionerClient(newSession *session.Session, bucketName, tableName, databaseName, catalogID, crawlerName string) Partitioner {
	return &PartitionerClient{
		bucketName:   bucketName,
		tableName:    tableName,
		databaseName: databaseName,
		catalogID:    catalogID,
		crawlerName:  crawlerName,
		s3Client:     s3.New(newSession),
		glueClient:   glue.New(newSession),
	}
}

// AddPartition adds the required partition in AWS Athena under the
// provided accountID value.
func (p *PartitionerClient) AddPartition(ctx context.Context, accountID string) error {
	for _, path := range paths {
		_, err := p.s3Client.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(p.bucketName),
			Key:    aws.String(fmt.Sprintf("%s/%s", path, accountID)),
		})
		if err != nil {
			return &ErrorPutObject{
				err:  err,
				path: path,
			}
		}

		_, err = p.glueClient.CreatePartition(&glue.CreatePartitionInput{
			CatalogId:    aws.String(p.catalogID),
			DatabaseName: aws.String(p.databaseName),
			PartitionInput: &glue.PartitionInput{
				Values: []*string{
					aws.String(fmt.Sprintf("%s/%s", path, accountID)),
				},
			},
			TableName: &p.tableName,
		})
		if err != nil {
			return &ErrorCreatePartition{
				err: err,
			}
		}
	}

	if err := p.startCrawler(ctx); err != nil {
		return &ErrorStartCrawler{
			err: err,
		}
	}

	return nil
}

func (p *PartitionerClient) startCrawler(ctx context.Context) error {
	count := 1
	for count <= 3 {
		_, err := p.glueClient.StartCrawler(&glue.StartCrawlerInput{
			Name: aws.String(p.crawlerName),
		})
		if errors.Is(err, &glue.CrawlerRunningException{}) {
			time.Sleep(time.Duration(count * 5 * int(time.Second)))
		} else if err != nil {
			return err
		}
		count++
	}

	return nil
}

// RemovePartition removes the required partition(s) in AWS Athena under
// the provided accountID value.
func (p *PartitionerClient) RemovePartition(ctx context.Context, accountID string) error {
	for _, path := range paths {
		input := &glue.DeletePartitionInput{
			CatalogId:    aws.String(p.catalogID),
			DatabaseName: aws.String(p.databaseName),
			PartitionValues: []*string{
				aws.String(fmt.Sprintf("%s/%s", path, accountID)),
			},
			TableName: aws.String(p.tableName),
		}

		_, err := p.glueClient.DeletePartition(input)
		if err != nil {
			return &ErrorDeletePartition{
				err: err,
			}
		}
	}

	return nil
}
