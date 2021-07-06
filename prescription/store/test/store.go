package test

import (
	"context"

	"github.com/tidepool-org/platform/prescription/store"
)

type Store struct {
	GetPrescriptionRepositoryInvocations int
	GetPrescriptionRepositoryImpl        *PrescriptionRepository
}

func NewStore() *Store {
	return &Store{
		GetPrescriptionRepositoryImpl: NewPrescriptionRepository(),
	}
}

func (s *Store) GetPrescriptionRepository() store.PrescriptionRepository {
	s.GetPrescriptionRepositoryInvocations++
	return s.GetPrescriptionRepositoryImpl
}

func (s *Store) Status(context.Context) interface{} {
	return nil
}

func (s *Store) CreateIndexes(context.Context) error {
	return nil
}

func (s *Store) Expectations() {
	s.GetPrescriptionRepositoryImpl.Expectations()
}
