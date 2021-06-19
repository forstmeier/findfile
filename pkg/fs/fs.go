package fs

const mainBucket = "cheesesteakstorage-main"

// Filesystemer defines the methods for interacting with the
// target filesystem.
type Filesystemer interface {
	GenerateUploadURL(filename string) (string, error)
}
