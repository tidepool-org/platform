package task

import (
	"github.com/tidepool-org/platform/errors"
)

const (
	ErrorCodeCalculationFailure = "calculation-failure"
)

func ErrorResourceFailureError(err error) error {
	return errors.WrapPrepared(err, ErrorCodeCalculationFailure, "resource failure", "temporary failure due to upstream summary calculation error")
}
