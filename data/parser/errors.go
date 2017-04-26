package parser

import "github.com/tidepool-org/platform/service"

// TODO: Review all errors for consistency and language
// Once shipped, Code and Title cannot change

func ErrorNotParsed() *service.Error {
	return &service.Error{
		Code:   "not-parsed",
		Title:  "not parsed",
		Detail: "Not parsed",
	}
}
