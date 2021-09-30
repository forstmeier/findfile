package db

import "fmt"

const packageName = "db"

// AddFolderError wraps errors returned by db.helper.addFolder.
type AddFolderError struct {
	err error
}

func (e *AddFolderError) Error() string {
	return fmt.Sprintf("[%s] [setup database] [add folder]: %s", packageName, e.err.Error())
}

// UploadObjectError wraps errors returned by db.helper.uploadObject.
type UploadObjectError struct {
	err      error
	function string
	entity   string
}

func (e *UploadObjectError) Error() string {
	return fmt.Sprintf("[%s] [%s] [upload object] [entity: %s]: %s", packageName, e.function, e.entity, e.err.Error())
}

// ExecuteQueryError wraps errors returned by db.helper.executeQuery.
type ExecuteQueryError struct {
	err      error
	function string
}

func (e *ExecuteQueryError) Error() string {
	return fmt.Sprintf("[%s] [%s] [execute query]: %s", packageName, e.function, e.err.Error())
}

// GetQueryResultsError wraps errors returned by db.helper.getQueryResultIDs
// and helper.getQueryResultDocuments.
type GetQueryResultsError struct {
	err         error
	function    string
	subfunction string
}

func (e *GetQueryResultsError) Error() string {
	return fmt.Sprintf("[%s] [%s] [%s]: %s", packageName, e.function, e.subfunction, e.err.Error())
}

// DeleteDocumentsByKeysError wraps errors returned by db.helper.deleteDocumentsByKeys.
type DeleteDocumentsByKeysError struct {
	err error
}

func (e *DeleteDocumentsByKeysError) Error() string {
	return fmt.Sprintf("[%s] [delete documents] [delete documents by keys]: %s", packageName, e.err.Error())
}
