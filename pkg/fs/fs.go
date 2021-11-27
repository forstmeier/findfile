package fs

import "context"

// Filesystemer defines methods for interacting with the
// target filesystem.
type Filesystemer interface {
	ListFiles(ctx context.Context, bucket string) ([]string, error)
}
