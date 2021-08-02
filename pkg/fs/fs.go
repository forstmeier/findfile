package fs

import "context"

// Filesystemer defines the methods for interacting with the
// target filesystem.
type Filesystemer interface {
	GenerateUploadURL(ctx context.Context, accountID string, fileInfo FileInfo) (string, error)
	GenerateDownloadURL(ctx context.Context, accountID string, fileInfo FileInfo) (string, error)
	ListFilesByAccountID(ctx context.Context, filepath, accountID string) ([]FileInfo, error)
	DeleteFiles(ctx context.Context, accountID string, filesInfo []FileInfo) error
}

// FileInfo holds data related to a file.
type FileInfo struct {
	Filename string
	Filepath string
}
