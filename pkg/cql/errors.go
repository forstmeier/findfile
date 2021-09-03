package cql

import "fmt"

const packageName = "cql"

// ErrorParseCQL wraps errors returned by cql.parseCQL in
// the ConvertCQL method.
type ErrorParseCQL struct {
	err error
}

func (e *ErrorParseCQL) Error() string {
	return fmt.Sprintf("[%s] [convert cql]: %s", packageName, e.err.Error())
}
