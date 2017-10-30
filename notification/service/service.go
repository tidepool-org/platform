package service

import (
	"github.com/tidepool-org/platform/notification/store"
	"github.com/tidepool-org/platform/service"
)

type Service interface {
	service.Service

	NotificationStore() store.Store

	Status() *Status
}

type Status struct {
	Version           string
	Server            interface{}
	NotificationStore interface{}
}
