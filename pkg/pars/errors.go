package pars

import "fmt"

const packageName = "pars"

// ErrorParseDocument wraps errors returned by
// textract.Textract.AnalyzeDocument in the pars.Parser.Parse
// method.
type ErrorParseDocument struct {
	err error
}

func (e *ErrorParseDocument) Error() string {
	return fmt.Sprintf("[%s] [parse document]: %s", packageName, e.err.Error())
}
