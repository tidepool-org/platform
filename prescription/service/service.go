package service

import (
	"github.com/tidepool-org/platform/prescription"
	"github.com/tidepool-org/platform/prescription/store"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/user"
)

type Service interface {
	service.Service

	PrescriptionStore() store.Store
	PrescriptionClient() prescription.Client
	UserClient() user.Client

	Status() *Status
}

type Status struct {
	Version string
	Server  interface{}
	Store   interface{}
}
