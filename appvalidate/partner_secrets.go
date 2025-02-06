package appvalidate

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrUnsupportedPartner        = errors.New("unsupported partner")
	ErrPartnerUnitialized        = errors.New("partner not initialized")
	ErrInvalidPartnerCredentials = errors.New("invalid partner credentials")
	ErrEmptySecretsConfig        = errors.New("empty secrets config")
	ErrInvalidPartnerPayload     = errors.New("invalid partner payload")
)

type PartnerSecrets struct {
	cs *CoastalSecrets
	ps *PalmTreeSecrets
}

// NewPartnerSecrets creates a new Secrets wrapper with all supported
// partners.
func NewPartnerSecrets(cs *CoastalSecrets, ps *PalmTreeSecrets) *PartnerSecrets {
	return &PartnerSecrets{
		cs: cs,
		ps: ps,
	}
}

func (s *PartnerSecrets) GetSecret(ctx context.Context, payload AssertionClientData) (response any, err error) {
	switch payload.Partner {
	case PartnerCoastal:
		if s.cs == nil {
			return nil, fmt.Errorf("%w: %v", ErrPartnerUnitialized, payload.Partner)
		}
		return s.cs.GetSecret(ctx, []byte(payload.PartnerData))
	case PartnerPalmTree:
		if s.ps == nil {
			return nil, fmt.Errorf("%w: %v", ErrPartnerUnitialized, payload.Partner)
		}
		return s.ps.GetSecret(ctx, []byte(payload.PartnerData))
	default:
		return nil, fmt.Errorf("%w: %v", ErrUnsupportedPartner, payload.Partner)
	}
}
