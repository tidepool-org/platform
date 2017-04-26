package client

import (
	"fmt"
	"net/http"

	"github.com/tidepool-org/platform/service"
)

type AuthenticationDetails interface {
	Token() string

	IsServer() bool
	UserID() string
}

type Client interface {
	ValidateAuthenticationToken(context service.Context, authenticationToken string) (AuthenticationDetails, error)
	GetUserPermissions(context service.Context, requestUserID string, targetUserID string) (Permissions, error)
	GetUserGroupID(context service.Context, userID string) (string, error)

	ServerToken() (string, error)
}

type Permission map[string]interface{}
type Permissions map[string]Permission

const OwnerPermission = "root"
const CustodianPermission = "custodian"
const UploadPermission = "upload"
const ViewPermission = "view"

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
