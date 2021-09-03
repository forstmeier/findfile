package cql

import "context"

var _ CQLer = &Client{}

// Client implements the cql.CQLer methods using AWS Athena.
type Client struct {
	parseCQL func(ctx context.Context, accountID string, cqlQuery map[string]interface{}) ([]byte, error)
}

// New generates a cql.Client pointer instance for an AWS Athena
// backend implementation.
func New() *Client {
	return &Client{
		parseCQL: parseCQL,
	}
}

// ConvertCQL implements the cql.CQLer.ConvertCQL method.
func (c *Client) ConvertCQL(ctx context.Context, accountID string, cqlQuery map[string]interface{}) ([]byte, error) {
	query, err := c.parseCQL(ctx, accountID, cqlQuery)
	if err != nil {
		return nil, &ErrorParseCQL{err: err}
	}

	return query, nil
}
