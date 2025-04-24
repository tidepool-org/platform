package mocks

import (
	"context"
	"encoding/json"

	"github.com/tidepool-org/platform/permission"
)

// likeT encapsulates some handy methods of testing.T
//
// In ginkgo, the GinkgoT() method will work.
type likeT interface {
	Fatalf(format string, args ...any)
	Logf(format string, args ...any)
}

var (
	TestUserID1 = "62a372fa-7096-4d33-ab3a-1f26d7701f76"
	TestUserID2 = "89d13ccb-32fb-47ef-9a8c-9d45f5d1c145"
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

type PermissionsMapMap map[string]map[string]permission.Permissions

func TestUserPerms() permission.Permissions {
	return permission.Permissions{
		permission.Follow: map[string]interface{}{},
		permission.Read:   map[string]interface{}{},
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
