package types

import (
	"errors"
	"fmt"

	"github.com/tidepool-org/platform/validate"
)

type TestingHelper struct {
	validate.ErrorProcessing
}

type ExpectedErrorDetails struct {
	Path   string
	Detail string
}

func NewTestingHelper() *TestingHelper {
	return &TestingHelper{
		ErrorProcessing: validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()},
	}
}

func (t *TestingHelper) ValidDataType(builtType interface{}) error {

	if t.HasErrors() {
		fmt.Println("details of unexpected errors: ")
		for i := range t.Errors {
			fmt.Println(t.Errors[i].Source["pointer"])
			fmt.Println(t.Errors[i].Detail)
		}
		return errors.New("type errored while being built and validated")
	}
	if builtType == nil {
		return errors.New("type is nil")
	}

	return nil
}

func (t *TestingHelper) ErrorIsExpected(builtType interface{}, expected ExpectedErrorDetails) error {

	if !t.HasErrors() {
		return errors.New("there are no errors when we expected one")
	}

	if len(t.Errors) != 1 {

		fmt.Println("details of unexpected errors: ")
		for i := range t.Errors {
			fmt.Println(t.Errors[i].Source["pointer"])
			fmt.Println(t.Errors[i].Detail)
		}

		return fmt.Errorf("we expected only one error but found %d", len(t.Errors))
	}

	if expected.Detail != t.Errors[0].Detail {
		return fmt.Errorf("expected: %s actual: %s", expected.Detail, t.Errors[0].Detail)
	}

	if expected.Path != t.Errors[0].Source["pointer"] {
		return fmt.Errorf("expected: %s actual: %s", expected.Path, t.Errors[0].Source["pointer"])
	}

	return nil
}
