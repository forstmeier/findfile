package cql

import "context"

// CQLer defines the method for converting CQL queries
// into database implementation-compatible queries as
// a byte array.
type CQLer interface {
	ConvertCQL(ctx context.Context, accountID string, cqlQuery map[string]interface{}) ([]byte, error)
}
