package service

import (
	"github.com/tidepool-org/platform/prescription/store"
	"github.com/tidepool-org/platform/service"
)

type Service interface {
	service.Service

	PrescriptionStore() store.Store
	Status() *Status
}

type Status struct {
	Version string
	Server  interface{}
	Store   interface{}
}
