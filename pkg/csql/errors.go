package csql

import "fmt"

const packageName = "csql"

// ErrorParseCSQLJSON wraps errors returned by the parseJSONObject
// helper function in the CSQLToES method.
//
// No sub-types are wrapped and returned and the error message
// returned by parseJSONObject should be made explicit enough
// for debugging purposes.
type ErrorParseCSQLJSON struct {
	err error
}

func (e *ErrorParseCSQLJSON) Error() string {
	return fmt.Sprintf("%s: %s", packageName, e.err.Error())
}
