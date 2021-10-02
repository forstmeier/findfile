package db

import "fmt"

const errorMessage = "package db: %s"

// AddFolderError wraps errors returned by db.helper.addFolder.
type AddFolderError struct {
	err error
}

func (e *AddFolderError) Error() string {
	return fmt.Sprintf(errorMessage, e.err.Error())
}

// UploadObjectError wraps errors returned by db.helper.uploadObject.
type UploadObjectError struct {
	err error
}

func (e *UploadObjectError) Error() string {
	return fmt.Sprintf(errorMessage, e.err.Error())
}

// ExecuteQueryError wraps errors returned by db.helper.executeQuery.
type ExecuteQueryError struct {
	err error
}

func (e *ExecuteQueryError) Error() string {
	return fmt.Sprintf(errorMessage, e.err.Error())
}

// GetQueryResultsError wraps errors returned by db.helper.getQueryResultIDs
// and helper.getQueryResultDocuments.
type GetQueryResultsError struct {
	err error
}

func (e *GetQueryResultsError) Error() string {
	return fmt.Sprintf(errorMessage, e.err.Error())
}

// DeleteDocumentsByKeysError wraps errors returned by db.helper.deleteDocumentsByKeys.
type DeleteDocumentsByKeysError struct {
	err error
}

func (e *DeleteDocumentsByKeysError) Error() string {
	return fmt.Sprintf(errorMessage, e.err.Error())
}
