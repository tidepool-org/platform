package service

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

func (e *Error) Clone() *Error {
	err := &Error{
		Code:   e.Code,
		Detail: e.Detail,
		Status: e.Status,
		Title:  e.Title,
	}
	if e.Source != nil {
		err.Source = &Source{
			Parameter: e.Source.Parameter,
			Pointer:   e.Source.Pointer,
		}
	}
	return err
}

func NewErrors() *Errors {
	return &Errors{
		errors: []*Error{},
	}
}

func (e *Errors) HasErrors() bool {
	return len(e.errors) > 0
}

func (e *Errors) GetErrors() []*Error {
	return e.errors
}

func (e *Errors) AppendError(err *Error) {
	e.errors = append(e.errors, err)
}
