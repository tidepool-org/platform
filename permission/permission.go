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

func HasMembershipRelationship(ctx context.Context, client Client, granteeUserID, grantorUserID string) (has bool, err error) {
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
