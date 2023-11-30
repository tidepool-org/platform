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
// partners. It may return an error and a valid Secrets object because we may
// not want to use or have the system fail even if one of the partners can't
// be initialized.
func NewPartnerSecrets() (*PartnerSecrets, error) {
	var errs []error
	cs, err := NewCoastalSecrets()
	if err != nil {
		errs = append(errs, err)
	}

	ps, err := NewPalmTreeSecrets()
	if err != nil {
		errs = append(errs, err)
	}

	return &PartnerSecrets{
		cs: cs,
		ps: ps,
	}, errors.Join(errs...)
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
