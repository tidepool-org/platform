package service

import (
	"github.com/tidepool-org/platform/notification"
	"github.com/tidepool-org/platform/notification/store"
	"github.com/tidepool-org/platform/service"
)

type Service interface {
	service.Service

	NotificationStore() store.Store

	Status() *notification.Status
}
