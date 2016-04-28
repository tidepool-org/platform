package app

import "fmt"

func Error(pkg string, message string) error {
	return fmt.Errorf("%s: %s", pkg, message)
}

func Errorf(pkg string, format string, a ...interface{}) error {
	return Error(pkg, fmt.Sprintf(format, a...))
}

func ExtError(err error, pkg string, message string) error {
	return fmt.Errorf("%s: %s; %s", pkg, message, err.Error())
}

func ExtErrorf(err error, pkg string, format string, a ...interface{}) error {
	return ExtError(err, pkg, fmt.Sprintf(format, a...))
}
