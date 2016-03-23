package data

import "fmt"

//ErrorSet contains a list of Error's
type ErrorSet struct {
	errors []*Error
}

//NewErrorSet creates an initialised ErrorSet
func NewErrorSet() *ErrorSet {
	return &ErrorSet{}
}

//AppendError appends an error to the ErrorSet
func (e *ErrorSet) AppendError(err *Error) {
	if err == nil || err.IsEmpty() {
		return
	}
	e.errors = append(e.errors, err)
	return
}

//Error returns a string representation of the errors
func (e *ErrorSet) Error() string {

	errorsStr := ""
	if e != nil {
		for i := range e.errors {
			if e.errors[i] != nil {
				errorsStr = fmt.Sprintln(errorsStr, e.errors[i].Error())
			}
		}
	}

	return errorsStr
}

//Error contains a list or errors and asscociated data
type Error struct {
	errors []error
	data   map[string]interface{}
}

//NewError creates and instance of Error
func NewError(data map[string]interface{}) *Error {
	return &Error{data: data}
}

//AppendError appends an error to the Error
func (e *Error) AppendError(err error) {
	if err == nil {
		return
	}
	e.errors = append(e.errors, err)
	return
}

//AppendFieldError appends a error and associated field name
func (e *Error) AppendFieldError(name string, detail interface{}) {

	e.errors = append(e.errors,
		fmt.Errorf("encountered an error on type field %s when given %v ", name, detail),
	)
	return
}

//IsEmpty return true if there are no errors contained
func (e *Error) IsEmpty() bool {
	return len(e.errors) == 0
}

//Error returns a string representation of the errors
func (e *Error) Error() string {

	errorsStr := ""

	for i := range e.errors {
		if errorsStr == "" {
			errorsStr = e.errors[i].Error()
			break
		}
		errorsStr = fmt.Sprintln(errorsStr, e.errors[i].Error())
	}

	return fmt.Sprintf("processing %v found: %s", e.data, errorsStr)

}
