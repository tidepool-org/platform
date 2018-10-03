package test

import (
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/test"
)

type RestrictedTokenSession struct {
	*test.Closer
	*authTest.RestrictedTokenAccessor
}

func NewRestrictedTokenSession() *RestrictedTokenSession {
	return &RestrictedTokenSession{
		Closer:                  test.NewCloser(),
		RestrictedTokenAccessor: authTest.NewRestrictedTokenAccessor(),
	}
}

func (r *RestrictedTokenSession) Expectations() {
	r.Closer.AssertOutputsEmpty()
	r.RestrictedTokenAccessor.Expectations()
}
