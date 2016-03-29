package data

import (
	"fmt"
	"strings"
)

type ErrorArray struct {
	errors []*Error
}

func NewErrorArray() *ErrorArray {
	return &ErrorArray{}
}

func (e *ErrorArray) AppendError(err *Error) {
	if err != nil && !err.IsEmpty() {
		e.errors = append(e.errors, err)
	}
}

func (e *ErrorArray) Error() string {
	if e == nil || len(e.errors) == 0 {
		return ""
	}

	errorStrings := []string{}
	for i := range e.errors {
		errorStrings = append(errorStrings, e.errors[i].Error())
	}
	return strings.Join(errorStrings, ";")
}

type Error struct {
	errors []error
	datum  Datum
}

func NewError(datum Datum) *Error {
	return &Error{datum: datum}
}

func (e *Error) AppendError(err error) {
	if err != nil {
		e.errors = append(e.errors, err)
	}
}

func (e *Error) AppendFieldError(name string, detail interface{}) {
	e.errors = append(e.errors,
		fmt.Errorf("encountered an error on type field %s when given %v ", name, detail),
	)
	return
}

func (e *Error) IsEmpty() bool {
	return e == nil || len(e.errors) == 0
}

func (e *Error) Error() string {
	if e == nil || len(e.errors) == 0 {
		return ""
	}
	errorStrings := []string{}
	for i := range e.errors {
		errorStrings = append(errorStrings, e.errors[i].Error())
	}
	return fmt.Sprintf("%#v %s", e.datum, strings.Join(errorStrings, ";"))
}
