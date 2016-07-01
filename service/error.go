package service

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

type Source struct {
	Parameter string `json:"parameter,omitempty"`
	Pointer   string `json:"pointer,omitempty"`
}

type Error struct {
	Code   string      `json:"code,omitempty"`
	Title  string      `json:"title,omitempty"`
	Detail string      `json:"detail,omitempty"`
	Status int         `json:"status,string,omitempty"`
	Source *Source     `json:"source,omitempty"`
	Meta   interface{} `json:"meta,omitempty"`
}

func (e *Error) WithSourceParameter(parameter string) *Error {
	if e.Source == nil {
		e.Source = &Source{}
	}
	e.Source.Parameter = parameter
	return e
}

func (e *Error) WithSourcePointer(pointer string) *Error {
	if e.Source == nil {
		e.Source = &Source{}
	}
	e.Source.Pointer = pointer
	return e
}

func (e *Error) WithMeta(meta interface{}) *Error {
	e.Meta = meta
	return e
}
