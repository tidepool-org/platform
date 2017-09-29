package v1

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/notification/service"
)

type Router struct {
	service.Service
}

func NewRouter(svc service.Service) (*Router, error) {
	if svc == nil {
		return nil, errors.New("service is missing")
	}

	return &Router{
		Service: svc,
	}, nil
}

func (r *Router) Routes() []*rest.Route {
	return []*rest.Route{}
}

/*
POST /v1/users/:userId/notifications
	request includes info about type of notification
	web:
		message:
		template?:
	email?:
	sms?:
	push?:

GET /v1/users/:userId/notifications

DELETE /v1/notifications/:notificationid

How to mark as dismissed?
Do we want to retain after dismissed for tracking purposes?
Or add to audit log?
*/
