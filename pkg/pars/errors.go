package pars

import "fmt"

const errorMessage = "package pars: %s"

// ParseDocumentError wraps errors returned by
// textract.Textract.AnalyzeDocument in the pars.Parser.Parse
// method.
type ParseDocumentError struct {
	err error
}

func (e *ParseDocumentError) Error() string {
	return fmt.Sprintf(errorMessage, e.err.Error())
}
