package cql

import "fmt"

const packageName = "cql"

// ErrorConvertCQL wraps errors returned by cql.parseCQL in
// the ConvertCQL method.
type ErrorConvertCQL struct {
	err error
}

func (e *ErrorConvertCQL) Error() string {
	return fmt.Sprintf("%s: convert cql: %s", packageName, e.err.Error())
}
