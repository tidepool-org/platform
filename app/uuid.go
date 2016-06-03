package app

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"strings"

	uuid "github.com/satori/go.uuid"
)

func NewUUID() string {
	return uuid.NewV4().String()
}

func NewID() string {
	return strings.Replace(NewUUID(), "-", "", -1)
}
