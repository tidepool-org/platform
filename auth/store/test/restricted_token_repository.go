package test

import (
	authTest "github.com/tidepool-org/platform/auth/test"
)

type RestrictedTokenRepository struct {
	*authTest.RestrictedTokenAccessor
}

func NewRestrictedTokenRepository() *RestrictedTokenRepository {
	return &RestrictedTokenRepository{
		RestrictedTokenAccessor: authTest.NewRestrictedTokenAccessor(),
	}
}

func (r *RestrictedTokenRepository) Expectations() {
	r.RestrictedTokenAccessor.Expectations()
}
