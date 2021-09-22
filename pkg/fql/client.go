package fql

import "context"

var _ FQLer = &Client{}

// Client implements the fql.FQLer methods using AWS Athena.
type Client struct {
	parseFQL func(ctx context.Context, fqlQuery map[string]interface{}) ([]byte, error)
}

// New generates a fql.Client pointer instance for an AWS Athena
// backend implementation.
func New() *Client {
	return &Client{
		parseFQL: parseFQL,
	}
}

// ConvertFQL implements the fql.FQLer.ConvertFQL method
// using AWS Athena.
func (c *Client) ConvertFQL(ctx context.Context, fqlQuery map[string]interface{}) ([]byte, error) {
	query, err := c.parseFQL(ctx, fqlQuery)
	if err != nil {
		return nil, &ErrorParseFQL{err: err}
	}

	return query, nil
}
