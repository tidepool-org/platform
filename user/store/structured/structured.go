package structured

import (
	"context"

	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/user"
)

type Store interface {
	NewUserRepository() UserRepository
}

type UserRepository interface {
	Get(ctx context.Context, id string, condition *request.Condition) (*user.User, error)
	Delete(ctx context.Context, id string, condition *request.Condition) (bool, error)
	Destroy(ctx context.Context, id string, condition *request.Condition) (bool, error)
}
