package docpars

import "fmt"

const packageName = "docpars"

// ErrorAnalyzeDocument wraps errors returned by textract.Textract.AnalyzeDocument
// in the Convert method.
type ErrorAnalyzeDocument struct {
	err error
}

func (e *ErrorAnalyzeDocument) Error() string {
	return fmt.Sprintf("[%s] [analyze document]: %s", packageName, e.err.Error())
}
