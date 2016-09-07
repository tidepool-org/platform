package parser

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

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
