package test

import testStore "github.com/tidepool-org/platform/store/test"

type AuthsSession struct {
	*testStore.Session
}

func NewAuthsSession() *AuthsSession {
	return &AuthsSession{
		Session: testStore.NewSession(),
	}
}

func (a *AuthsSession) UnusedOutputsCount() int {
	return a.Session.UnusedOutputsCount()
}
