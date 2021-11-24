package db

import "fmt"

const errorMessage = "package db: %v"

// NewClientError wraps errors returned by db.New.
type NewClientError struct {
	err error
}

func (e *NewClientError) Error() string {
	return fmt.Sprintf(errorMessage, e.err)
}

// MarshalDocumentError wraps errors returned by json.Marshal
// in db.Databaser.UpsertDocuments.
type MarshalDocumentError struct {
	err error
}

func (e *MarshalDocumentError) Error() string {
	return fmt.Sprintf(errorMessage, e.err)
}

// WriteDocumentDataError wraps errors returned by
// bytes.Buffer.Write in db.Databaser.UpsertDocuments.
type WriteDocumentDataError struct {
	err error
}

func (e *WriteDocumentDataError) Error() string {
	return fmt.Sprintf(errorMessage, e.err)
}

// ExecuteBulkError wraps errors returned by db.helper.executeBulk
// in db.Databaser.UpsertDocuments.
type ExecuteBulkError struct {
	err error
}

func (e *ExecuteBulkError) Error() string {
	return fmt.Sprintf(errorMessage, e.err)
}

// ExecuteDeleteError wraps errors returned by db.helper.executeDelete
// in db.Databaser.DeleteDocuments.
type ExecuteDeleteError struct {
	err error
}

func (e *ExecuteDeleteError) Error() string {
	return fmt.Sprintf(errorMessage, e.err)
}

// ExecuteQueryError wraps errors returned by db.helper.executeQuery
// in db.Databaser.QueryDocuments.
type ExecuteQueryError struct {
	err error
}

func (e *ExecuteQueryError) Error() string {
	return fmt.Sprintf(errorMessage, e.err)
}

// ReadQueryResponseBodyError wraps errors returned by io.ReadAll
// in db.Databaser.QueryDocuments.
type ReadQueryResponseBodyError struct {
	err error
}

func (e *ReadQueryResponseBodyError) Error() string {
	return fmt.Sprintf(errorMessage, e.err)
}

// UnmarshalQueryResponseBodyError wraps errors returned by json.Unmarshal
// in db.Databaser.QueryDocuments.
type UnmarshalQueryResponseBodyError struct {
	err error
}

func (e *UnmarshalQueryResponseBodyError) Error() string {
	return fmt.Sprintf(errorMessage, e.err)
}
