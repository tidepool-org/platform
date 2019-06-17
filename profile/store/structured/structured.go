package structured

import (
	"context"
	"io"

	"github.com/tidepool-org/platform/profile"
	"github.com/tidepool-org/platform/request"
)

type Store interface {
	NewSession() Session
}

type Session interface {
	io.Closer

	Get(ctx context.Context, userID string, condition *request.Condition) (*profile.Profile, error)
	Delete(ctx context.Context, userID string, condition *request.Condition) (bool, error)
	Destroy(ctx context.Context, userID string, condition *request.Condition) (bool, error)
}
