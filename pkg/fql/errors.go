package fql

import "fmt"

const packageName = "fql"

// ParseFQLError wraps errors returned by fql.parseFQL in
// the fql.FQLer.ConvertFQL method.
type ParseFQLError struct {
	err error
}

func (e *ParseFQLError) Error() string {
	return fmt.Sprintf("[%s] [convert fql]: %s", packageName, e.err.Error())
}
