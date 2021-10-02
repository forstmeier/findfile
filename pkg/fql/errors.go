package fql

import "fmt"

const errorMessage = "package fql: %s"

// ParseFQLError wraps errors returned by fql.parseFQL in
// the fql.FQLer.ConvertFQL method.
type ParseFQLError struct {
	err error
}

func (e *ParseFQLError) Error() string {
	return fmt.Sprintf(errorMessage, e.err.Error())
}
