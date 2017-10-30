package user

import "context"

type Client interface {
	GetUserPermissions(ctx context.Context, requestUserID string, targetUserID string) (Permissions, error)
}
