package types

// TODO: This needs to be moved to its own /data/types/test package so it is separate

import (
	"errors"
	"fmt"

	"github.com/tidepool-org/platform/validate"
)

type TestingHelper struct {
	ErrorProcessing validate.ErrorProcessing
}

type ExpectedErrorDetails struct {
	Path   string
	Detail string
}

func NewTestingHelper() *TestingHelper {
	return &TestingHelper{
		ErrorProcessing: validate.NewErrorProcessing("0"),
	}
}

func (t *TestingHelper) ValidDataType(builtType interface{}) error {

	if t.ErrorProcessing.HasErrors() {
		fmt.Println("details of unexpected errors: ")
		for _, err := range t.ErrorProcessing.GetErrors() {
			fmt.Println(err.Source.Pointer)
			fmt.Println(err.Detail)
		}
		return errors.New("type errored while being built and validated")
	}
	if builtType == nil {
		return errors.New("type is nil")
	}

	return nil
}

func (t *TestingHelper) ErrorIsExpected(builtType interface{}, expected ExpectedErrorDetails) error {

	if !t.ErrorProcessing.HasErrors() {
		return errors.New("there are no errors when we expected one")
	}

	if errorCount := len(t.ErrorProcessing.GetErrors()); errorCount != 1 {
		fmt.Println("details of unexpected errors: ")
		for _, err := range t.ErrorProcessing.GetErrors() {
			fmt.Println(err.Source.Pointer)
			fmt.Println(err.Detail)
		}

		return fmt.Errorf("we expected only one error but found %d", errorCount)
	}

	err := t.ErrorProcessing.GetErrors()[0]
	if expected.Detail != err.Detail {
		return fmt.Errorf("expected: %s actual: %s", expected.Detail, err.Detail)
	}
	if expected.Path != err.Source.Pointer {
		return fmt.Errorf("expected: %s actual: %s", expected.Path, err.Source.Pointer)
	}

	return nil
}

func (t *TestingHelper) HasExpectedErrors(builtType interface{}, expected map[string]ExpectedErrorDetails) error {

	if !t.ErrorProcessing.HasErrors() {
		return errors.New("there are no errors when we expected them")
	}

	for _, err := range t.ErrorProcessing.GetErrors() {
		if found, ok := expected[err.Source["pointer"]]; !ok {
			return fmt.Errorf("unexpected error source: %s", err.Source["pointer"])
		} else {
			if found.Detail != err.Detail {
				return fmt.Errorf("expected: %s actual: %s", found.Detail, err.Detail)
			}
		}
	}

	return nil
}
