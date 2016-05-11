package service

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import "net/http"

type Error struct {
	Code   string  `json:"code,omitempty"`
	Detail string  `json:"detail,omitempty"`
	Source *Source `json:"source,omitempty"`
	Status int     `json:"status,string,omitempty"`
	Title  string  `json:"title,omitempty"`
}

type Source struct {
	Parameter string `json:"parameter,omitempty"`
	Pointer   string `json:"pointer,omitempty"`
}

type Errors struct {
	errors []*Error
}

const (
	ErrorInternalServerFailure = "internal-server-failure"
)

var (
	InternalServerFailure = &Error{
		Code:   ErrorInternalServerFailure,
		Status: http.StatusInternalServerError,
		Detail: "Internal server failure",
		Title:  "internal server failure",
	}
)

func (e *Error) WithParameter(parameter string) *Error {
	e.Source = &Source{Parameter: parameter}
	return e
}

func (e *Error) WithPointer(pointer string) *Error {
	e.Source = &Source{Pointer: pointer}
	return e
}

func NewErrors() *Errors {
	return &Errors{
		errors: []*Error{},
	}
}

func (e *Errors) HasErrors() bool {
	return len(e.errors) > 0
}

func (e *Errors) GetError(index int) *Error {
	if index < 0 || index >= len(e.errors) {
		return nil
	}

	return e.errors[index]
}

func (e *Errors) GetErrors() []*Error {
	return e.errors
}

func (e *Errors) AppendError(err *Error) {
	e.errors = append(e.errors, err)
}
