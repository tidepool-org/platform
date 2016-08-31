package context

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"github.com/ant0ine/go-json-rest/rest"

	metricservicesClient "github.com/tidepool-org/platform/metricservices/client"
	commonService "github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/context"
	userservicesClient "github.com/tidepool-org/platform/userservices/client"
	"github.com/tidepool-org/platform/userservices/service"
)

type Standard struct {
	commonService.Context
	metricServicesClient  metricservicesClient.Client
	userServicesClient    userservicesClient.Client
	authenticationDetails userservicesClient.AuthenticationDetails
}

func WithContext(metricServicesClient metricservicesClient.Client, userServicesClient userservicesClient.Client, handler service.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		context, err := context.NewStandard(response, request)
		if err != nil {
			context.RespondWithInternalServerFailure("Unable to create new context for request", err)
			return
		}

		handler(&Standard{
			Context:              context,
			metricServicesClient: metricServicesClient,
			userServicesClient:   userServicesClient,
		})
	}
}

func (s *Standard) MetricServicesClient() metricservicesClient.Client {
	return s.metricServicesClient
}

func (s *Standard) UserServicesClient() userservicesClient.Client {
	return s.userServicesClient
}

func (s *Standard) AuthenticationDetails() userservicesClient.AuthenticationDetails {
	return s.authenticationDetails
}

func (s *Standard) SetAuthenticationDetails(authenticationDetails userservicesClient.AuthenticationDetails) {
	s.authenticationDetails = authenticationDetails
}
