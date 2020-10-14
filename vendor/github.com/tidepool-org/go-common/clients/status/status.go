package status

import (
	"fmt"
	"net/http"
)

// Type status holds the status return from an http request.
type Status struct {
	Code   int    `json:"code"`
	Error  *int   `json:"error,omitempty"`
	Reason string `json:"reason"`
}

// NewStatus constructs a Status object; if no reason is provided, it uses the
// standard one.
func NewStatus(statusCode int, reason string) Status {
	s := Status{Code: statusCode, Reason: reason}
	if s.Reason == "" {
		s.Reason = http.StatusText(statusCode)
	}
	return s
}

func NewStatusWithError(statusCode int, errorCode int, reason string) Status {
	s := Status{Code: statusCode, Error: &errorCode, Reason: reason}
	if s.Reason == "" {
		s.Reason = http.StatusText(statusCode)
	}
	return s
}

// NewStatus constructs a Status object; if no reason is provided, it uses the
// standard one.
func NewStatusf(statusCode int, reason string, args ...interface{}) Status {
	return Status{Code: statusCode, Reason: fmt.Sprintf(reason, args...)}
}

func StatusFromResponse(res *http.Response) Status {
	return Status{Code: res.StatusCode, Reason: res.Status}
}

// String() converts a status to a printable string.
func (s Status) String() string {
	return fmt.Sprintf("%d %s", s.Code, s.Reason)
}

// StatusError represents a Status as an error object.
type StatusError struct {
	Status
}

// Error() renders a StatusError.
func (s *StatusError) Error() string {
	return s.Status.String()
}
