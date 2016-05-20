package app

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import "fmt"

func Error(pkg string, message string) error {
	return fmt.Errorf("%s: %s", pkg, message)
}

func Errorf(pkg string, format string, a ...interface{}) error {
	return Error(pkg, fmt.Sprintf(format, a...))
}

func ExtError(err error, pkg string, message string) error {
	var errorString string
	if err != nil {
		errorString = err.Error()
	} else {
		errorString = "app: error is nil"
	}
	return fmt.Errorf("%s: %s; %s", pkg, message, errorString)
}

func ExtErrorf(err error, pkg string, format string, a ...interface{}) error {
	return ExtError(err, pkg, fmt.Sprintf(format, a...))
}
