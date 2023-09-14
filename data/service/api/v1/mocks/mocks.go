package mocks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
)

// likeT encapsulates some handy methods of testing.T
//
// In ginkgo, the GinkgoT() method will work.
type likeT interface {
	Fatalf(format string, args ...any)
	Logf(format string, args ...any)
}

var (
	TestUserID1 = "user1"
	TestUserID2 = "user2"
	TestToken1  = "token1"
	TestToken2  = "token2"
)

func MustMarshalJSON(t likeT, v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("error marshaling JSON: %s", err)
	}
	return data
}

// Details implements request.Details with test helpers.
type Details struct {
	request.Details
}

func NewDetailsDefault() *Details {
	return NewDetails(request.MethodSessionToken, TestUserID1, TestToken1)
}

func NewDetails(authMethod request.Method, userID, token string) *Details {
	return &Details{
		Details: request.NewDetails(authMethod, userID, token),
	}
}

// mockResponseWriter extends http.ResponseWriter with test utility.
type mockResponseWriter struct {
	http.ResponseWriter
}

func NewResponseWriter(w http.ResponseWriter) *mockResponseWriter {
	return &mockResponseWriter{
		ResponseWriter: w,
	}
}

// WriteJson is a method of rest.ResponseWriter that is useful to override.
func (w *mockResponseWriter) WriteJson(object interface{}) error {
	data, err := w.EncodeJson(object)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w.ResponseWriter, string(data))
	return err
}

// EncodeJson is a method of rest.ResponseWriter that is useful to override.
func (c *mockResponseWriter) EncodeJson(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

type PermissionsMapMap map[string]map[string]permission.Permissions

func TestUserPerms() permission.Permissions {
	return permission.Permissions{
		permission.Alerting: map[string]interface{}{},
		permission.Read:     map[string]interface{}{},
	}
}

func TestPerms() PermissionsMapMap {
	return PermissionsMapMap{
		TestUserID1: {
			TestUserID2: TestUserPerms(),
		},
	}
}

type Permission struct {
	Perms   PermissionsMapMap
	Default permission.Permissions
	Error   error
}

func NewPermission(perms PermissionsMapMap, def permission.Permissions, err error) *Permission {
	if def == nil {
		def = TestUserPerms()
	}
	return &Permission{
		Perms:   perms,
		Default: def,
		Error:   err,
	}
}

func NewPermissionDefault() *Permission {
	return NewPermission(nil, TestUserPerms(), nil)
}

func NewPermissionError(err error) *Permission {
	return NewPermission(nil, nil, err)
}

func (p *Permission) GetUserPermissions(ctx context.Context, requestUserID string, targetUserID string) (permission.Permissions, error) {
	if p.Error != nil {
		return nil, p.Error
	}
	if p, found := p.Perms[requestUserID][targetUserID]; found {
		return p, nil
	}

	return p.Default, nil
}
