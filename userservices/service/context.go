package service

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	metricservicesClient "github.com/tidepool-org/platform/metricservices/client"
	"github.com/tidepool-org/platform/service"
	userservicesClient "github.com/tidepool-org/platform/userservices/client"
)

type Context interface {
	service.Context

	MetricServicesClient() metricservicesClient.Client
	UserServicesClient() userservicesClient.Client

	AuthenticationDetails() userservicesClient.AuthenticationDetails
	SetAuthenticationDetails(authenticationDetails userservicesClient.AuthenticationDetails)
}

type HandlerFunc func(context Context)
