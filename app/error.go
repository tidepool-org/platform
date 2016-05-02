package app

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
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
