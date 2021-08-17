package db

import "fmt"

const packageName = "db"

// ErrorListingObjects wraps errors returned by s3.S3.ListObjectsV2 in
// the db.Databaser.CreateOrUpdateDocuments and
// db.Databaser.DeleteDocuments methods.
type ErrorListingObjects struct {
	err    error
	action string
}

func (e *ErrorListingObjects) Error() string {
	return fmt.Sprintf("%s: %s documents: %s", packageName, e.action, e.err.Error())
}

// ErrorMarshalData wraps errors returned by json.Marshal in the
// db.Databaser.CreateOrUpdateDocuments method.
type ErrorMarshalData struct {
	err    error
	entity string
}

func (e *ErrorMarshalData) Error() string {
	return fmt.Sprintf("%s: create or update documents: [%s] %s", packageName, e.entity, e.err.Error())
}

// ErrorPutObject wraps errors returned by s3.S3.PutObject in the
// db.Databaser.CreateOrUpdateDocuments method.
type ErrorPutObject struct {
	err    error
	entity string
}

func (e *ErrorPutObject) Error() string {
	return fmt.Sprintf("%s: create or update documents: [%s] %s", packageName, e.entity, e.err.Error())
}
