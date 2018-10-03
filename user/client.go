package user

import "context"

type Client interface {
	EnsureAuthorizedService(ctx context.Context) error
	EnsureAuthorizedUser(ctx context.Context, targetUserID string, permission string) (string, error)
	GetUserPermissions(ctx context.Context, requestUserID string, targetUserID string) (Permissions, error)
}
