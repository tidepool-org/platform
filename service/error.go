package service

import "strconv"

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

type Error struct {
	Code   string  `json:"code,omitempty"`
	Title  string  `json:"title,omitempty"`
	Detail string  `json:"detail,omitempty"`
	Status int     `json:"status,string,omitempty"`
	Source *Source `json:"source,omitempty"`
}

type Source struct {
	Parameter string `json:"parameter,omitempty"`
	Pointer   string `json:"pointer,omitempty"`
}

func (e *Error) WithParameter(parameter string) *Error {
	if e.Source == nil {
		e.Source = &Source{}
	}
	e.Source.Parameter = parameter
	return e
}

func (e *Error) WithPointer(pointer string) *Error {
	if e.Source == nil {
		e.Source = &Source{}
	}
	e.Source.Pointer = pointer
	return e
}

func QuoteIfString(interfaceValue interface{}) interface{} {
	if stringValue, ok := interfaceValue.(string); ok {
		return strconv.Quote(stringValue)
	}
	return interfaceValue
}

// TODO: Deprecate below Errors struct

type Errors struct {
	errors []*Error
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
