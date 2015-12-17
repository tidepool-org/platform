package data

import (
	"errors"
	"fmt"
)

type DataError struct {
	errors []error
	data   map[string]interface{}
}

func NewDataError(data map[string]interface{}) *DataError {
	return &DataError{data: data}
}

func (e *DataError) AppendError(err error) {
	if err == nil {
		return
	}
	e.errors = append(e.errors, err)
	return
}

func (e *DataError) AppendFieldError(name string, detail interface{}) {

	e.errors = append(e.errors, errors.New(
		fmt.Sprintf("encountered an error on type field %s when given %v ", name, detail),
	))
	return
}

func (e *DataError) IsEmpty() bool {
	return len(e.errors) == 0
}

func (e *DataError) Error() string {

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
