package permission

import (
	"context"
)

type Permission map[string]interface{}
type Permissions map[string]Permission

// GroupedPermissions are permissions that are keyed by userID. As an example a response may be {"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa":{"root":{}},"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb":{"note":{},"upload":{},"view":{}}}
type GroupedPermissions map[string]Permissions

type TrustPermissions struct {
	TrustorPermissions *Permission
	TrusteePermissions *Permission
}

const (
	Follow    = "follow"
	Custodian = "custodian"
	Owner     = "root"
	Read      = "view"
	Write     = "upload"
)

type Client interface {
	GetUserPermissions(ctx context.Context, requestUserID string, targetUserID string) (Permissions, error)
	// GroupsForUser returns permissions that have been shared with granteeUserID. It is keyed by the user that has shared something with granteeUserID
	GroupsForUser(ctx context.Context, granteeUserID string) (GroupedPermissions, error)
	// UsersInGroup returns permissions that the user with id sharerID has shared with others, keyed by user id.
	UsersInGroup(ctx context.Context, sharerID string) (GroupedPermissions, error)
	HasMembershipRelationship(ctx context.Context, granteeUserID, grantorUserID string) (has bool, err error)
	HasCustodianPermissions(ctx context.Context, granteeUserID, grantorUserID string) (has bool, err error)
	HasWritePermissions(ctx context.Context, granteeUserID, grantorUserID string) (has bool, err error)
}

func FixGroupedOwnerPermissions(groupPermissions GroupedPermissions) GroupedPermissions {
	for key, perms := range groupPermissions {
		groupPermissions[key] = FixOwnerPermissions(perms)
	}
	return groupPermissions
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

func (p Permissions) HasReadPermissions() bool {
	for _, perm := range []string{Custodian, Owner, Read, Write} {
		if _, ok := p[perm]; ok {
			return true
		}
	}
	return false
}
