package v1

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
	"fmt"
	"net/http"

	"github.com/tidepool-org/platform/service"
)

func ErrorUserIDMissing() *service.Error {
	return &service.Error{
		Code:   "user-id-missing",
		Status: http.StatusBadRequest,
		Title:  "user id is missing",
		Detail: "User id is missing",
	}
}

func ErrorUserIDNotFound(userID string) *service.Error {
	return &service.Error{
		Code:   "user-id-not-found",
		Status: http.StatusNotFound,
		Title:  "user with specified id not found",
		Detail: fmt.Sprintf("User with id %s not found", userID),
	}
}
