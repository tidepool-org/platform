package api

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

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/environment"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service/api"
	"github.com/tidepool-org/platform/userservices/client"
	"github.com/tidepool-org/platform/userservices/service"
	"github.com/tidepool-org/platform/userservices/service/context"
	"github.com/tidepool-org/platform/version"
)

type Standard struct {
	*api.Standard
	userServicesClient client.Client
}

func NewStandard(versionReporter version.Reporter, environmentReporter environment.Reporter, logger log.Logger, userServicesClient client.Client) (*Standard, error) {
	if versionReporter == nil {
		return nil, app.Error("api", "version reporter is missing")
	}
	if environmentReporter == nil {
		return nil, app.Error("api", "environment reporter is missing")
	}
	if logger == nil {
		return nil, app.Error("api", "logger is missing")
	}
	if userServicesClient == nil {
		return nil, app.Error("api", "user services client is missing")
	}

	standard, err := api.NewStandard(versionReporter, environmentReporter, logger)
	if err != nil {
		return nil, err
	}

	return &Standard{
		Standard:           standard,
		userServicesClient: userServicesClient,
	}, nil
}

func (s *Standard) InitializeRouter(routes []service.Route) error {
	baseRoutes := []service.Route{
		service.MakeRoute("GET", "/status", s.GetStatus),
		service.MakeRoute("GET", "/version", s.GetVersion),
	}

	routes = append(baseRoutes, routes...)

	var contextRoutes []*rest.Route
	for _, route := range routes {
		contextRoutes = append(contextRoutes, &rest.Route{
			HttpMethod: route.Method,
			PathExp:    route.Path,
			Func:       s.withContext(route.Handler),
		})
	}

	router, err := rest.MakeRouter(contextRoutes...)
	if err != nil {
		return app.ExtError(err, "api", "unable to create router")
	}

	s.API().SetApp(router)

	return nil
}

func (s *Standard) withContext(handler service.HandlerFunc) rest.HandlerFunc {
	return context.WithContext(s.userServicesClient, handler)
}
