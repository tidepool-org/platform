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
	UpdateUserPermissions(ctx context.Context, requestUserID string, targetUserID string, permissions Permissions) error
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

func (p Permissions) HasReadPermissions() bool {
	for _, perm := range []string{Custodian, Owner, Read, Write} {
		if _, ok := p[perm]; ok {
			return true
		}
	}
	return false
}
