package db

import "fmt"

const packageName = "db"

// ErrorNewClient wraps errors returned by mongo.NewClient in
// the db.New method.
type ErrorNewClient struct {
	err error
}

func (e *ErrorNewClient) Error() string {
	return fmt.Sprintf("%s: new client: %s", packageName, e.err.Error())
}

// ErrorUpdateDocument wraps errors returned by mongo.Client.UpdateOne
// in the db.Databaser.CreateOrUpdateDocuments method.
type ErrorUpdateDocument struct {
	err error
}

func (e *ErrorUpdateDocument) Error() string {
	return fmt.Sprintf("%s: create or update: %s", packageName, e.err.Error())
}

// ErrorDeleteDocuments wraps errors returned by mongo.Client.DeleteOne
// in the db.Databaser.Delete method.
type ErrorDeleteDocuments struct {
	err error
}

func (e *ErrorDeleteDocuments) Error() string {
	return fmt.Sprintf("%s: delete: %s", packageName, e.err.Error())
}

// ErrorQueryDocuments wraps errors returned by mongo.Client.Find in
// the db.Databaser.Query method.
type ErrorQueryDocuments struct {
	err error
}

func (e *ErrorQueryDocuments) Error() string {
	return fmt.Sprintf("%s: query: %s", packageName, e.err.Error())
}

// ErrorParseQueryResults wraps errors returned by mongo.Cursor.All
// in the db.Databaser.Query method.
type ErrorParseQueryResults struct {
	err error
}

func (e *ErrorParseQueryResults) Error() string {
	return fmt.Sprintf("%s: query: %s", packageName, e.err.Error())
}
