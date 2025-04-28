package appvalidate

import (
	"context"

	"github.com/tidepool-org/platform/errors"
)

var (
	ErrDuplicateKeyId = errors.New("duplicate key id")
	ErrKeyIdNotFound  = errors.New("key id not found")
)

type Filter struct {
	UserID string
	KeyID  string
}

//go:generate mockgen -source=repository.go -destination=test/repository_mocks.go -package=test Repository
type Repository interface {
	IsVerified(ctx context.Context, f Filter) (bool, error)
	Get(ctx context.Context, f Filter) (*AppValidation, error)
	GetAttestationChallenge(ctx context.Context, f Filter) (string, error)
	Upsert(ctx context.Context, v *AppValidation) error
	UpdateAttestation(ctx context.Context, f Filter, u AttestationUpdate) error
	UpdateAssertion(ctx context.Context, f Filter, u AssertionUpdate) error
}
