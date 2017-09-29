package test

import (
	testAuth "github.com/tidepool-org/platform/auth/test"
	testStore "github.com/tidepool-org/platform/store/test"
)

type RestrictedTokenSession struct {
	*testStore.Session
	*testAuth.RestrictedTokenAccessor
}

func NewRestrictedTokenSession() *RestrictedTokenSession {
	return &RestrictedTokenSession{
		Session:                 testStore.NewSession(),
		RestrictedTokenAccessor: testAuth.NewRestrictedTokenAccessor(),
	}
}

func (r *RestrictedTokenSession) Expectations() {
	r.Session.Expectations()
	r.RestrictedTokenAccessor.Expectations()
}
