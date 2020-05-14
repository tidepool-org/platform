package test

import "github.com/tidepool-org/platform/errors"

type ErrorReporter struct {
	err     error
	warning error
}

func NewErrorReporter() *ErrorReporter {
	return &ErrorReporter{}
}

func (e *ErrorReporter) HasError() bool {
	return e.err != nil
}

func (e *ErrorReporter) Error() error {
	return e.err
}

func (e *ErrorReporter) HasWarning() bool {
	return e.warning != nil
}

func (e *ErrorReporter) Warning() error {
	return e.warning
}

func (e *ErrorReporter) ReportError(err error) {
	if err != nil {
		e.err = errors.Append(e.err, err)
	}
}
