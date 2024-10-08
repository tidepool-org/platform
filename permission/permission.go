package permission

import (
	"context"
)

type Permission map[string]interface{}
type Permissions map[string]Permission

const (
	Follow    = "follow"
	Custodian = "custodian"
	Owner     = "root"
	Read      = "view"
	Write     = "upload"
)

type Client interface {
	GetUserPermissions(ctx context.Context, requestUserID string, targetUserID string) (Permissions, error)
}

func FixOwnerPermissions(permissions Permissions) Permissions {
	if ownerPermission, ok := permissions[Owner]; ok {
		if _, ok = permissions[Write]; !ok {
			permissions[Write] = ownerPermission
		}
		if _, ok = permissions[Read]; !ok {
			permissions[Read] = ownerPermission
		}
	}
	return permissions
}

// HasExplicitMembershipRelationship return whether a grantor has given a
// grantee explicit rights. This is need in some places where we want to test a
// user's permission. It is called "Explicit" because in most middleware, a
// service call already has implicit rights and this would then only be called
// to check if a user has explicit writes if the AuthDetails were not from a
// service.
func HasExplicitMembershipRelationship(ctx context.Context, client Client, granteeUserID, grantorUserID string) (has bool, err error) {
	fromTo, err := client.GetUserPermissions(ctx, granteeUserID, grantorUserID)
	if err != nil {
		return false, err
	}
	if len(fromTo) > 0 {
		return true, nil
	}
	toFrom, err := client.GetUserPermissions(ctx, grantorUserID, granteeUserID)
	if err != nil {
		return false, err
	}
	if len(toFrom) > 0 {
		return true, nil
	}
	return false, nil
}

func HasExplicitWritePermissions(ctx context.Context, c Client, granteeUserID, grantorUserID string) (has bool, err error) {
	if granteeUserID != "" && granteeUserID == grantorUserID {
		return true, nil
	}
	perms, err := c.GetUserPermissions(ctx, granteeUserID, grantorUserID)
	if err != nil {
		return false, err
	}
	if _, ok := perms[Custodian]; ok {
		return true, nil
	}
	if _, ok := perms[Write]; ok {
		return true, nil
	}
	if _, ok := perms[Owner]; ok {
		return true, nil
	}
	return false, nil
}
