package pars

import "fmt"

const packageName = "pars"

// ErrorAnalyzeDocument wraps errors returned by
// textract.Textract.AnalyzeDocument in the pars.Parser.Parse
// method.
type ErrorAnalyzeDocument struct {
	err error
}

func (e *ErrorAnalyzeDocument) Error() string {
	return fmt.Sprintf("[%s] [analyze document]: %s", packageName, e.err.Error())
}
