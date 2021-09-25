package db

import "fmt"

const packageName = "db"

// ErrorAddFolder wraps errors returned by db.helper.addFolder.
type ErrorAddFolder struct {
	err error
}

func (e *ErrorAddFolder) Error() string {
	return fmt.Sprintf("[%s] [setup database] [add folder]: %s", packageName, e.err.Error())
}

// ErrorUploadObject wraps errors returned by db.helper.uploadObject.
type ErrorUploadObject struct {
	err      error
	function string
	entity   string
}

func (e *ErrorUploadObject) Error() string {
	return fmt.Sprintf("[%s] [%s] [upload object] [entity: %s]: %s", packageName, e.function, e.entity, e.err.Error())
}

// ErrorExecuteQuery wraps errors returned by db.helper.executeQuery.
type ErrorExecuteQuery struct {
	err      error
	function string
}

func (e *ErrorExecuteQuery) Error() string {
	return fmt.Sprintf("[%s] [%s] [execute query]: %s", packageName, e.function, e.err.Error())
}

// ErrorGetQueryResults wraps errors returned by db.helper.getQueryResultIDs
// and helper.getQueryResultDocuments.
type ErrorGetQueryResults struct {
	err         error
	function    string
	subfunction string
}

func (e *ErrorGetQueryResults) Error() string {
	return fmt.Sprintf("[%s] [%s] [%s]: %s", packageName, e.function, e.subfunction, e.err.Error())
}

// ErrorDeleteDocumentsByKeys wraps errors returned by db.helper.deleteDocumentsByKeys.
type ErrorDeleteDocumentsByKeys struct {
	err error
}

func (e *ErrorDeleteDocumentsByKeys) Error() string {
	return fmt.Sprintf("[%s] [delete documents] [delete documents by keys]: %s", packageName, e.err.Error())
}
