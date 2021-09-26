package fql

import "context"

// FQLer defines the method for converting FQL queries
// into database implementation-compatible queries as
// a byte array.
type FQLer interface {
	ConvertFQL(ctx context.Context, fqlQuery []byte) ([]byte, error)
}
