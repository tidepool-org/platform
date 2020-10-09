package errors

import (
	"errors"
	"fmt"
)

type WrappedError struct {
	error
	m string
}

func (err *WrappedError) Error() string {
	return err.m + ": " + err.error.Error()
}

func New(message string) error {
	return errors.New(message)
}

func Newf(message string, args ...interface{}) error {
	return errors.New(fmt.Sprintf(message, args...))
}

func Wrap(e error, m string) error {
	return &WrappedError{error: e, m: m}
}

func Wrapf(e error, m string, args ...interface{}) error {
	return Wrap(e, fmt.Sprintf(m, args...))
}
