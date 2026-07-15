package test

import (
	authTest "github.com/tidepool-org/platform/auth/test"
)

type ProviderSessionRepository struct {
	*authTest.ProviderSessionClient
}

func NewProviderSessionRepository() *ProviderSessionRepository {
	return &ProviderSessionRepository{
		ProviderSessionClient: authTest.NewProviderSessionClient(),
	}
}

func (p *ProviderSessionRepository) Expectations() {
	p.ProviderSessionClient.Expectations()
}
