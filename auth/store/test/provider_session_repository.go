package test

import (
	authTest "github.com/tidepool-org/platform/auth/test"
)

type ProviderSessionRepository struct {
	*authTest.ProviderSessionAccessor
}

func NewProviderSessionRepository() *ProviderSessionRepository {
	return &ProviderSessionRepository{
		ProviderSessionAccessor: authTest.NewProviderSessionAccessor(),
	}
}

func (p *ProviderSessionRepository) Expectations() {
	p.ProviderSessionAccessor.Expectations()
}
