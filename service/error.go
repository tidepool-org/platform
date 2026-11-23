package service

import "github.com/tidepool-org/platform/errors"

type Error struct {
	Code   string `json:"code,omitempty"`
	Title  string `json:"title,omitempty"`
	Detail string `json:"detail,omitempty"`
	Status int    `json:"status,string,omitempty"`
	Meta   any    `json:"meta,omitempty"`
}

func (e *Error) WithMeta(meta any) *Error {
	e.Meta = meta
	return e
}

func (e *Error) AsError() error {
	return errors.WithMeta(errors.Prepared(e.Code, e.Title, e.Detail), e.Meta)
}
