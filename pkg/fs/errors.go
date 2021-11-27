package fs

import "fmt"

const errorMessage = "package fs: %s"

// ListObjectsError wraps errors returned by fs.ListFiles.
type ListObjectsError struct {
	err error
}

func (e *ListObjectsError) Error() string {
	return fmt.Sprintf(errorMessage, e.err.Error())
}
