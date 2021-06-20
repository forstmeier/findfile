package fs

import "context"

const mainBucket = "cheesesteakstorage-main"

// Filesystemer defines the methods for interacting with the
// target filesystem.
type Filesystemer interface {
	GenerateUploadURL(ctx context.Context, accountID string, filename string) (string, error)
	DeleteFiles(ctx context.Context, accountID string, filenames []string) error
}
