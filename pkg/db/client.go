package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/cheesesteakio/api/pkg/docpars"
)

var _ Databaser = &Client{}

// Client implements the db.Databaser methods using DocumentDB.
type Client struct {
	documentDBClient documentDBClient
}

type documentDBClient interface {
	UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
	Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error)
}

// New generates a db.Client pointer instance with a DocumentDB client.
func New(databaseName, collectionName string) (*Client, error) {
	ddb, err := mongo.NewClient(nil)
	if err != nil {
		return nil, &ErrorNewClient{err: err}
	}

	return &Client{
		documentDBClient: ddb.Database(databaseName).Collection(collectionName),
	}, nil
}

// CreateOrUpdateDocuments implements the db.Databaser.CreateOrUpdateDocuments
// method.
func (c *Client) CreateOrUpdateDocuments(ctx context.Context, documents []docpars.Document) error {
	for _, document := range documents {
		filter := bson.D{
			primitive.E{
				Key:   "filename",
				Value: document.Filename,
			},
			primitive.E{
				Key:   "filepath",
				Value: document.Filepath,
			},
		}

		upsert := true
		option := &options.UpdateOptions{
			Upsert: &upsert,
		}

		_, err := c.documentDBClient.UpdateOne(ctx, filter, document, option)
		if err != nil {
			return &ErrorUpdateDocument{err: err}
		}
	}

	return nil
}

// DeleteDocuments implements the db.Databaser.DeleteDocuments method.
func (c *Client) DeleteDocuments(ctx context.Context, documentsInfo []DocumentInfo) error {
	for _, documentInfo := range documentsInfo {
		filter := bson.D{
			primitive.E{
				Key:   "filename",
				Value: documentInfo.Filename,
			},
			primitive.E{
				Key:   "filepath",
				Value: documentInfo.Filepath,
			},
		}

		_, err := c.documentDBClient.DeleteOne(ctx, filter)
		if err != nil {
			return &ErrorDeleteDocuments{err: err}
		}
	}

	return nil
}

// QueryDocuments implements the db.Databaser.QueryDocuments method.
func (c *Client) QueryDocuments(ctx context.Context, query []byte) ([]docpars.Document, error) {
	cursor, err := c.documentDBClient.Find(ctx, query)
	if err != nil {
		return nil, &ErrorQueryDocuments{err: err}
	}

	var documents []docpars.Document
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, &ErrorParseQueryResults{err: err}
	}

	return documents, nil
}
