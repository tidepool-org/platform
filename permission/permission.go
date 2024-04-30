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
	HasMembershipRelationship(ctx context.Context, granteeUserID, grantorUserID string) (has bool, err error)
	HasCustodianPermissions(ctx context.Context, granteeUserID, grantorUserID string) (has bool, err error)
	HasWritePermissions(ctx context.Context, granteeUserID, grantorUserID string) (has bool, err error)
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
