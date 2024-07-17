package permission

import (
	"context"
)

type Permission map[string]interface{}

// Permissions are permissions that are keyed depending on the type of permissions that are being retrieved.
//
// If it is a one to one user to user permission check, then it is keyed by permssion type (Follow, Custodian, etc):
//
//	Permissions{"follow": struct{}{}, "upload": struct{}{}}
//
// If it is a grouped set of permissions, it is keyed by userId:
//
//	Permissions{"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa":{"root":{}},"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb":{"note":{},"upload":{},"view":{}}}
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
	// GroupsForUser returns permissions that have been shared with granteeUserID. It is keyed by the user that has shared something with granteeUserID. It includes the user themself.
	GroupsForUser(ctx context.Context, granteeUserID string) (Permissions, error)
	// UsersInGroup returns permissions that the user with id sharerID has shared with others, keyed by user id. It includes the user themself.
	UsersInGroup(ctx context.Context, sharerID string) (Permissions, error)
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

func (p Permission) Has(permissionType string) bool {
	_, exists := p[permissionType]
	return exists
}

func (p Permission) HasAny(permissionTypes ...string) bool {
	for _, perm := range permissionTypes {
		if p.Has(perm) {
			return true
		}
	}
	return false
}
