package fs

import "context"

// Filesystemer defines the methods for monitoring the target
// file system.
type Filesystemer interface {
	CreateFileWatcher(ctx context.Context, path string) error
	DeleteFileWatcher(ctx context.Context, path string) error
}
