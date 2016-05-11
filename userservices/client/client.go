package client

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

type Client interface {
	Start() error
	Close()
	ValidateUserSession(context *service.Context, sessionToken string) (string, error)
	ValidateTargetUserPermissions(context *service.Context, requestUserID string, targetUserID string, targetPermissions Permissions) error
	GetUserGroupID(context *service.Context, userID string) (string, error)
}

type Permission map[string]interface{}
type Permissions map[string]Permission

var UploadPermissions = Permissions{"upload": {}}
var ViewPermissions = Permissions{"view": {}}

type UnauthorizedError struct{}

func NewUnauthorizedError() *UnauthorizedError {
	return &UnauthorizedError{}
}

func (u *UnauthorizedError) Error() string {
	return "client: unauthorized"
}

func IsUnauthorizedError(err error) bool {
	_, ok := err.(*UnauthorizedError)
	return ok
}

type UnexpectedResponseError struct {
	Method     string
	URL        string
	StatusCode int
}

func NewUnexpectedResponseError(response *http.Response, request *http.Request) *UnexpectedResponseError {
	return &UnexpectedResponseError{
		Method:     request.Method,
		URL:        request.URL.String(),
		StatusCode: response.StatusCode,
	}
}

func (u *UnexpectedResponseError) Error() string {
	return fmt.Sprintf("client: unexpected response status code %d from %s %s", u.StatusCode, u.Method, u.URL)
}
