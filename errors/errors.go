package errors

import "fmt"

func New(pkg string, message string) error {
	return fmt.Errorf("%s: %s", pkg, message)
}

func Newf(pkg string, format string, a ...interface{}) error {
	return New(pkg, fmt.Sprintf(format, a...))
}

func Wrap(err error, pkg string, message string) error {
	var errorString string
	if err != nil {
		errorString = err.Error()
	} else {
		errorString = "errors: error is nil"
	}
	return fmt.Errorf("%s: %s; %s", pkg, message, errorString)
}

func Wrapf(err error, pkg string, format string, a ...interface{}) error {
	return Wrap(err, pkg, fmt.Sprintf(format, a...))
}
