package db

import "fmt"

const packageName = "db"

// ErrorUploadObject wraps errors returned by helper.uploadObject.
type ErrorUploadObject struct {
	err      error
	function string
}

func (e *ErrorUploadObject) Error() string {
	return fmt.Sprintf("[%s] [%s] [upload object]: %s", packageName, e.function, e.err.Error())
}

// ErrorExecuteQuery wraps errors returned by helper.executeQuery.
type ErrorExecuteQuery struct {
	err      error
	function string
}

func (e *ErrorExecuteQuery) Error() string {
	return fmt.Sprintf("[%s] [%s] [execute query]: %s", packageName, e.function, e.err.Error())
}

// ErrorGetQueryResults wraps errors returned by helper.getQueryResultIDs
// and helper.getQueryResultDocuments.
type ErrorGetQueryResults struct {
	err         error
	function    string
	subfunction string
}

func (e *ErrorGetQueryResults) Error() string {
	return fmt.Sprintf("[%s] [%s] [%s]: %s", packageName, e.function, e.subfunction, e.err.Error())
}

// ErrorListDocumentKeys wraps errors returned by helper.listDocumentKeys.
type ErrorListDocumentKeys struct {
	err error
}

func (e *ErrorListDocumentKeys) Error() string {
	return fmt.Sprintf("[%s] [delete documents] [list document keys]: %s", packageName, e.err.Error())
}

// ErrorDeleteDocumentsByKeys wraps errors returned by helper.deleteDocumentsByKeys.
type ErrorDeleteDocumentsByKeys struct {
	err error
}

func (e *ErrorDeleteDocumentsByKeys) Error() string {
	return fmt.Sprintf("[%s] [delete documents] [delete documents by keys]: %s", packageName, e.err.Error())
}
