package service

import (
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/task/store"
)

type Service interface {
	service.Service

	TaskStore() store.Store
	TaskClient() task.Client

	Status() *Status
}

type Status struct {
	Version   string
	Server    interface{}
	TaskStore interface{}
}
