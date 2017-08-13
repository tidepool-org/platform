package service

import (
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/auth/store"
	"github.com/tidepool-org/platform/service"
)

type Service interface {
	service.Service

	AuthStore() store.Store

	Status() *auth.Status
}
