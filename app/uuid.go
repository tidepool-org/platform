package app

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"strings"

	uuid "github.com/satori/go.uuid"
)

func NewUUID() string {
	return strings.Replace(uuid.NewV4().String(), "-", "", -1)
}
