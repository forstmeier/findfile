package fs

import "fmt"

const packageName = "fs"

// ErrorNewClient wrap errors returned by session.NewSession in the
// New helper function.
type ErrorNewClient struct {
	err error
}

func (e *ErrorNewClient) Error() string {
	return fmt.Sprintf("%s: new client: %s", packageName, e.err.Error())
}

// ErrorPresignURL wraps errors returned by request.Request.Presign
// in the GenerateUploadURL method.
type ErrorPresignURL struct {
	err    error
	action string
}

func (e *ErrorPresignURL) Error() string {
	return fmt.Sprintf("%s: presign %s: %s", packageName, e.action, e.err.Error())
}

// ErrorListObjects wraps errors returned by s3.S3.ListObjectsV2
// in the DeleteFiles method.
type ErrorListObjects struct {
	err error
}

func (e *ErrorListObjects) Error() string {
	return fmt.Sprintf("%s: list files: %s", packageName, e.err.Error())
}

// ErrorDeleteObjects wraps errors returned by s3.S3.DeleteObjects
// in the DeleteFiles method.
type ErrorDeleteObjects struct {
	err error
}

func (e *ErrorDeleteObjects) Error() string {
	return fmt.Sprintf("%s: delete files: %s", packageName, e.err.Error())
}
