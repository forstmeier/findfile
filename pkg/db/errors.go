package db

import "fmt"

const packageName = "db"

// ErrorUploadObject wraps errors returned by db.uploadObject in the
// db.Databaser.UpsertDocuments method.
type ErrorUploadObject struct {
	err      error
	function string
}

func (e *ErrorUploadObject) Error() string {
	return fmt.Sprintf("[%s] [%s] [upload object]: %s", packageName, e.function, e.err.Error())
}
