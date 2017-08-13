package test

import (
	testStore "github.com/tidepool-org/platform/store/test"
)

type StoreSession struct {
	*testStore.Session
}

func NewStoreSession() *StoreSession {
	return &StoreSession{
		Session: testStore.NewSession(),
	}
}

func (s *StoreSession) UnusedOutputsCount() int {
	return s.Session.UnusedOutputsCount()
}
