package evt

import "fmt"

const errorMessage = "package evt: %s"

// GetEventValuesError wraps errors returned by evt.helper.getEventValues.
type GetEventValuesError struct {
	err error
}

func (e *GetEventValuesError) Error() string {
	return fmt.Sprintf(errorMessage, e.err.Error())
}

// PutEventValuesError wraps errors returned by evt.helper.putEventValues.
type PutEventValuesError struct {
	err error
}

func (e *PutEventValuesError) Error() string {
	return fmt.Sprintf(errorMessage, e.err.Error())
}
