package confirmation

import "context"

type ConfirmationAccessor interface {
	DeleteUserConfirmations(ctx context.Context, userID string) error
}
