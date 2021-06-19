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
	err error
}

func (e *ErrorPresignURL) Error() string {
	return fmt.Sprintf("%s: presign: %s", packageName, e.err.Error())
}
