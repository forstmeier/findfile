package fql

import "fmt"

const packageName = "fql"

// ErrorParseFQL wraps errors returned by fql.parseFQL in
// the fql.FQLer.ConvertFQL method.
type ErrorParseFQL struct {
	err error
}

func (e *ErrorParseFQL) Error() string {
	return fmt.Sprintf("[%s] [convert fql]: %s", packageName, e.err.Error())
}
