package db

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/glue"
)

type glueClient interface {
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
	tableName    string
	databaseName string
	catalogID    string
	glueClient   glueClient
}

// NewPartitionerClient returns a db.Partitioner pointer instance.
func NewPartitionerClient(newSession *session.Session, tableName, databaseName, catalogID string) Partitioner {
	return &PartitionerClient{
		tableName:    tableName,
		databaseName: databaseName,
		catalogID:    catalogID,
		glueClient:   glue.New(newSession),
	}
}

// AddPartition adds the required partition(s) in AWS Athena under the
// provided accountID value.
func (p *PartitionerClient) AddPartition(ctx context.Context, accountID string) error {
	input := &glue.CreatePartitionInput{
		CatalogId:    &p.catalogID,
		DatabaseName: &p.databaseName,
		PartitionInput: &glue.PartitionInput{
			Values: []*string{
				&accountID,
			},
		},
		TableName: &p.tableName,
	}

	_, err := p.glueClient.CreatePartition(input)
	if err != nil {
		return &ErrorCreatePartition{
			err: err,
		}
	}

	return nil
}

// RemovePartition removes the required partition(s) in AWS Athena under
// the provided accountID value.
func (p *PartitionerClient) RemovePartition(ctx context.Context, accountID string) error {
	input := &glue.DeletePartitionInput{
		CatalogId:    &p.catalogID,
		DatabaseName: &p.databaseName,
		PartitionValues: []*string{
			&accountID,
		},
		TableName: &p.tableName,
	}

	_, err := p.glueClient.DeletePartition(input)
	if err != nil {
		return &ErrorDeletePartition{
			err: err,
		}
	}

	return nil
}
