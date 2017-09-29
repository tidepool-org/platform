package test

import (
	testAuth "github.com/tidepool-org/platform/auth/test"
	testStore "github.com/tidepool-org/platform/store/test"
)

type ProviderSessionSession struct {
	*testStore.Session
	*testAuth.ProviderSessionAccessor
}

func NewProviderSessionSession() *ProviderSessionSession {
	return &ProviderSessionSession{
		Session:                 testStore.NewSession(),
		ProviderSessionAccessor: testAuth.NewProviderSessionAccessor(),
	}
}

func (p *ProviderSessionSession) Expectations() {
	p.Session.Expectations()
	p.ProviderSessionAccessor.Expectations()
}
