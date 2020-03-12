package test

import (
	"github.com/tidepool-org/platform/prescription/store"
)

type Store struct {
	NewPrescriptionSessionInvocations int
	NewPrescriptionSessionImpl        *PrescriptionSession
}

func NewStore() *Store {
	return &Store{
		NewPrescriptionSessionImpl: NewPrescriptionSession(),
	}
}

func (s *Store) NewPrescriptionSession() store.PrescriptionSession {
	s.NewPrescriptionSessionInvocations++
	return s.NewPrescriptionSessionImpl
}

func (s *Store) Expectations() {
	s.NewPrescriptionSessionImpl.Expectations()
}
