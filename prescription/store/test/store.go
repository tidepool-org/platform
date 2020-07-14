package test

import (
	"github.com/tidepool-org/platform/prescription/store"
)

type Store struct {
	NewPrescriptionRepositoryInvocations int
	NewPrescriptionRepositoryImpl        *PrescriptionRepository
}

func NewStore() *Store {
	return &Store{
		NewPrescriptionRepositoryImpl: NewPrescriptionSession(),
	}
}

func (s *Store) NewPrescriptionSession() store.PrescriptionRepository {
	s.NewPrescriptionRepositoryInvocations++
	return s.NewPrescriptionRepositoryImpl
}

func (s *Store) Expectations() {
	s.NewPrescriptionRepositoryImpl.Expectations()
}
