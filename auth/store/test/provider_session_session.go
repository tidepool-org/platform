package test

import (
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/test"
)

type ProviderSessionSession struct {
	*test.Closer
	*authTest.ProviderSessionAccessor
}

func NewProviderSessionSession() *ProviderSessionSession {
	return &ProviderSessionSession{
		Closer:                  test.NewCloser(),
		ProviderSessionAccessor: authTest.NewProviderSessionAccessor(),
	}
}

func (p *ProviderSessionSession) Expectations() {
	p.Closer.AssertOutputsEmpty()
	p.ProviderSessionAccessor.Expectations()
}
