package devicecheck

import (
	"fmt"
	"net/http"
	"strings"
	"unsafe"

	"github.com/pkg/errors"
)

// Errors from DeviceCheck API
var (
	ErrBitStateNotFound          = errors.New("Failed to find bit state")
	ErrBadDeviceToken            = errors.New("Missing or incorrectly formatted device token payload")
	ErrBadBits                   = errors.New("Missing or incorrectly formatted bits")
	ErrBadTimestamp              = errors.New("Missing or incorrectly formatted time stamp")
	ErrInvalidAuthorizationToken = errors.New("Unable to verify authorization token")
	ErrMethodNotAllowed          = errors.New("Method Not Allowed")
)

func newError(code int, body []byte) error {
	if body == nil {
		return newUnknownError(code, body)
	}

	switch strings.TrimSpace(*(*string)(unsafe.Pointer(&body))) {
	case ErrBitStateNotFound.Error():
		return ErrBitStateNotFound
	case ErrBadDeviceToken.Error():
		return ErrBadDeviceToken
	case ErrBadBits.Error():
		return ErrBadBits
	case ErrBadTimestamp.Error():
		return ErrBadTimestamp
	case ErrInvalidAuthorizationToken.Error():
		return ErrInvalidAuthorizationToken
	case ErrMethodNotAllowed.Error():
		return ErrMethodNotAllowed
	default:
		return newUnknownError(code, body)
	}
}

func newUnknownError(code int, body []byte) error {
	if code == http.StatusOK {
		return nil
	}

	if body == nil {
		return fmt.Errorf("Unknown error (code: %d)", code)
	}

	return fmt.Errorf("%s (code: %d)", *(*string)(unsafe.Pointer(&body)), code)
}
