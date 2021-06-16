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
	InsertMany(ctx context.Context, documents []interface{}, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error)
	FindOneAndReplace(ctx context.Context, filter interface{}, replacement interface{}, opts ...*options.FindOneAndReplaceOptions) *mongo.SingleResult
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

// Create implements the db.Databaser.Create method.
func (c *Client) Create(ctx context.Context, documents []docpars.Document) error {
	input := make([]interface{}, len(documents))
	for i, document := range documents {
		input[i] = document
	}

	_, err := c.documentDBClient.InsertMany(ctx, input)
	if err != nil {
		return &ErrorCreateDocuments{err: err}
	}

	return nil
}

// Update implements the db.Databaser.Update method.
func (c *Client) Update(ctx context.Context, documents []docpars.Document) error {
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

		err := c.documentDBClient.FindOneAndReplace(ctx, filter, document).Err()
		if err != nil || err != mongo.ErrNoDocuments {
			return &ErrorUpdateDocuments{err: err}
		}
	}

	return nil
}

// Delete implements the db.Databaser.Delete method.
func (c *Client) Delete(ctx context.Context, documentsInfo []DocumentInfo) error {
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

// Query implements the db.Databaser.Query method.
func (c *Client) Query(ctx context.Context, query []byte) ([]docpars.Document, error) {
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