package csql

import "context"

// CSQLer defines the method for converting CSQL queries
// into database implementation-compatible queries as
// a byte array.
type CSQLer interface {
	ConvertCSQL(ctx context.Context, accountID string, csqlQuery map[string]interface{}) ([]byte, error)
}
