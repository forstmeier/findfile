package infra

import "fmt"

const packageName = "infra"

// ErrorCreateFilesystem wraps errors returned by aws.s3.CreateBucket
// in the CreateFilesystem method.
type ErrorCreateFilesystem struct {
	err error
}

func (e *ErrorCreateFilesystem) Error() string {
	return fmt.Sprintf("%s: create filesystem: %s", packageName, e.err.Error())
}

// ErrorDeleteFilesystem wraps errors returned by aws.s3.DeleteBucket
// in the DeleteFilesystem method.
type ErrorDeleteFilesystem struct {
	err error
}

func (e *ErrorDeleteFilesystem) Error() string {
	return fmt.Sprintf("%s: delete filesystem: %s", packageName, e.err.Error())
}

// ErrorCreateDatabase wraps errors returned by aws.docdb.CreateDBCluster
// in the CreateDatabase method.
type ErrorCreateDatabase struct {
	err error
}

func (e *ErrorCreateDatabase) Error() string {
	return fmt.Sprintf("%s: create database: %s", packageName, e.err.Error())
}

// ErrorDeleteDatabase wraps errors returned by aws.docdb.DeleteDBCluster
// in the DeleteDatabase method.
type ErrorDeleteDatabase struct {
	err error
}

func (e *ErrorDeleteDatabase) Error() string {
	return fmt.Sprintf("%s: delete database: %s", packageName, e.err.Error())
}
