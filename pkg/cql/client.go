package cql

import "context"

var _ CQLer = &Client{}

// Client implements the cql.CQLer methods using DocumentDB.
type Client struct {
	parseCQL func(accountID string, cqlQuery map[string]interface{}) ([]byte, error)
}

// New generates a cql.Client pointer instance for a DocumentDB
// backend implementation.
func New() *Client {
	return &Client{
		parseCQL: parseCQL,
	}
}

// ConvertCQL implements the cql.CQLer.ConvertCQL method.
func (c *Client) ConvertCQL(ctx context.Context, accountID string, cqlQuery map[string]interface{}) ([]byte, error) {
	bsonQuery, err := c.parseCQL(accountID, cqlQuery)
	if err != nil {
		return nil, &ErrorConvertCQL{err: err}
	}

	return bsonQuery, nil
}
