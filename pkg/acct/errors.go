package acct

import "fmt"

const packageName = "acct"

// ErrorPutItem wraps errors returned by dynamodb.DynamoDB.PutItem
// in the CreateAccount method.
type ErrorPutItem struct {
	err error
}

func (e *ErrorPutItem) Error() string {
	return fmt.Sprintf("%s: create account: %s", packageName, e.err.Error())
}

// ErrorGetItem wraps errors returned by dynamodb.DynamoDB.GetItem
// in the ReadAccount method.
type ErrorGetItem struct {
	err error
}

func (e *ErrorGetItem) Error() string {
	return fmt.Sprintf("%s: read account: %s", packageName, e.err.Error())
}

// ErrorIncorrectValue wraps errors returned by checkValues in the
// UpdateAccount method.
type ErrorIncorrectValue struct {
	key string
}

func (e *ErrorIncorrectValue) Error() string {
	return fmt.Sprintf("%s: update account: incorrect key [%s] in values", packageName, e.key)
}

// ErrorUpdateItem wraps errors returned by dynamodb.DynamoDB.UpdateItem
// in the UpdateAccount method.
type ErrorUpdateItem struct {
	err error
}

func (e *ErrorUpdateItem) Error() string {
	return fmt.Sprintf("%s: update account: %s", packageName, e.err.Error())
}

// ErrorDeleteItem wraps errors returned by dynamodb.DynamoDB.DeleteItem
// in the DeleteAccount method.
type ErrorDeleteItem struct {
	err error
}

func (e *ErrorDeleteItem) Error() string {
	return fmt.Sprintf("%s: delete account: %s", packageName, e.err.Error())
}
