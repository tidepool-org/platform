package appvalidate

import (
	"context"
)

type Filter struct {
	UserID string
	KeyID  string
}

type Repository interface {
	IsVerified(ctx context.Context, f Filter) (bool, error)
	Get(ctx context.Context, f Filter) (*AppValidation, error)
	GetAttestationChallenge(ctx context.Context, f Filter) (string, error)
	Upsert(ctx context.Context, v *AppValidation) error
	UpdateAttestation(ctx context.Context, f Filter, u AttestationUpdate) error
	UpdateAssertion(ctx context.Context, f Filter, u AssertionUpdate) error
}