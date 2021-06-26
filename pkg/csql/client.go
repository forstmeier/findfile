package csql

import "context"

var _ CSQLer = &Client{}

// Client implements the csql.CSQLer methods using DocumentDB.
type Client struct {
	parseCSQL func(accountID string, csqlQuery map[string]interface{}) ([]byte, error)
}

// New generates a csql.Client pointer instance for a DocumentDB
// backend implementation.
func New() *Client {
	return &Client{
		parseCSQL: parseCSQL,
	}
}

// ConvertCSQL implements the csql.CSQLer.ConvertCSQL method.
func (c *Client) ConvertCSQL(ctx context.Context, accountID string, csqlQuery map[string]interface{}) ([]byte, error) {
	bsonQuery, err := c.parseCSQL(accountID, csqlQuery)
	if err != nil {
		return nil, &ErrorConvertCSQL{err: err}
	}

	return bsonQuery, nil
}
