package csql

import "fmt"

const packageName = "csql"

// ErrorConvertCSQL wraps errors returned by csql.parseCSQL in
// the ConvertCSQL method.
type ErrorConvertCSQL struct {
	err error
}

func (e *ErrorConvertCSQL) Error() string {
	return fmt.Sprintf("%s: convert csql: %s", packageName, e.err.Error())
}
