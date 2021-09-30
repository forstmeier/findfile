package pars

import "fmt"

const packageName = "pars"

// ParseDocumentError wraps errors returned by
// textract.Textract.AnalyzeDocument in the pars.Parser.Parse
// method.
type ParseDocumentError struct {
	err error
}

func (e *ParseDocumentError) Error() string {
	return fmt.Sprintf("[%s] [parse document]: %s", packageName, e.err.Error())
}
