package devicecheck

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const (
	bitStateNotFoundStr = "Failed to find bit state"
)

var (
	ErrBadRequest         = errors.New("bad request")
	ErrUnauthorized       = errors.New("invalid or expired token")
	ErrForbidden          = errors.New("action not allowed")
	ErrMethodNotAllowed   = errors.New("method not allowed")
	ErrTooManyRequests    = errors.New("too many requests")
	ErrServer             = errors.New("server error")
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrUnknown            = errors.New("unknown error")
	ErrBitStateNotFound   = errors.New("bit state not found")
)

func isErrBitStateNotFound(body string) bool {
	return strings.Contains(body, bitStateNotFoundStr)
}

func newError(code int, body string) error {
	switch code {
	case http.StatusBadRequest:
		return fmt.Errorf("%w: %s", ErrBadRequest, body)
	case http.StatusUnauthorized:
		return fmt.Errorf("%w: %s", ErrUnauthorized, body)
	case http.StatusForbidden:
		return fmt.Errorf("%w: %s", ErrForbidden, body)
	case http.StatusMethodNotAllowed:
		return fmt.Errorf("%w: %s", ErrMethodNotAllowed, body)
	case http.StatusTooManyRequests:
		return fmt.Errorf("%w: %s", ErrTooManyRequests, body)
	case http.StatusInternalServerError:
		return fmt.Errorf("%w: %s", ErrServer, body)
	case http.StatusServiceUnavailable:
		return fmt.Errorf("%w: %s", ErrServiceUnavailable, body)
	default:
		return fmt.Errorf("%w: %s", ErrUnknown, body)
	}
}
