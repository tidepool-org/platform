package fetch

import (
	"github.com/tidepool-org/platform/errors"
)

const (
	ErrorCodeAuthenticationFailure = "authentication-failure"
	ErrorCodeInvalidState          = "invalid-state"
	ErrorCodeResourceFailure       = "resource-failure"
)

func ErrorAuthenticationFailureError(err error) error {
	return errors.WrapPrepared(err, ErrorCodeAuthenticationFailure, "authentication failure", "authentication failure")
}

func ErrorInvalidStateError(err error) error {
	return errors.WrapPrepared(err, ErrorCodeInvalidState, "invalid state", "unrecoverable error due to an invalid internal state")
}

func ErrorResourceFailureError(err error) error {
	return errors.WrapPrepared(err, ErrorCodeResourceFailure, "resource failure", "temporary failure due to missing or unreachable resource")
}
