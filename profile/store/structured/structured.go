package structured

import (
	"context"

	"github.com/tidepool-org/platform/profile"
	"github.com/tidepool-org/platform/request"
)

type Store interface {
	NewMetaRepository() MetaRepository
}

type MetaRepository interface {
	Get(ctx context.Context, userID string, condition *request.Condition) (*profile.Profile, error)
	Delete(ctx context.Context, userID string, condition *request.Condition) (bool, error)
	Destroy(ctx context.Context, userID string, condition *request.Condition) (bool, error)
}
