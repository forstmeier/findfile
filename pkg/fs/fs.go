package fs

import "context"

// Bucket name constants for use in Filesystemer callers.
const (
	MainBucket = "cheesesteakstorage-main"
	DemoBucket = "cheesesteakstorage-demo"
)

// Filesystemer defines the methods for interacting with the
// target filesystem.
type Filesystemer interface {
	GenerateUploadURL(ctx context.Context, bucketName, accountID, filename string) (string, error)
	GenerateDownloadURL(ctx context.Context, bucketName, accountID, filename string) (string, error)
	DeleteFiles(ctx context.Context, bucketName, accountID string, filenames []string) error
}
