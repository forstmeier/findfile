package csql

import "context"

var _ CSQLer = &Client{}

// Client implements the csql.CSQLer methods using DocumentDB.
type Client struct {
	parseCSQL func(csqlQuery map[string]interface{}) ([]byte, error)
}

// New generates a csql.Client pointer instance for a DocumentDB
// backend implementation.
func New() *Client {
	return &Client{
		parseCSQL: parseCSQL,
	}
}

// ConvertCSQL implements the csql.CSQLer.ConvertCSQL method.
func (c *Client) ConvertCSQL(ctx context.Context, csqlQuery map[string]interface{}) ([]byte, error) {
	bsonQuery, err := c.parseCSQL(csqlQuery)
	if err != nil {
		return nil, &ErrorConvertCSQL{err: err}
	}

	return bsonQuery, nil
}
