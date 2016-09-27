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
	dataservicesClient "github.com/tidepool-org/platform/dataservices/client"
	"github.com/tidepool-org/platform/environment"
	"github.com/tidepool-org/platform/log"
	messageStore "github.com/tidepool-org/platform/message/store"
	metricservicesClient "github.com/tidepool-org/platform/metricservices/client"
	notificationStore "github.com/tidepool-org/platform/notification/store"
	permissionStore "github.com/tidepool-org/platform/permission/store"
	profileStore "github.com/tidepool-org/platform/profile/store"
	"github.com/tidepool-org/platform/service/api"
	sessionStore "github.com/tidepool-org/platform/session/store"
	userStore "github.com/tidepool-org/platform/user/store"
	userservicesClient "github.com/tidepool-org/platform/userservices/client"
	"github.com/tidepool-org/platform/userservices/service"
	"github.com/tidepool-org/platform/userservices/service/context"
	"github.com/tidepool-org/platform/version"
)

type Standard struct {
	*api.Standard
	metricServicesClient metricservicesClient.Client
	userServicesClient   userservicesClient.Client
	dataServicesClient   dataservicesClient.Client
	messageStore         messageStore.Store
	notificationStore    notificationStore.Store
	permissionStore      permissionStore.Store
	profileStore         profileStore.Store
	sessionStore         sessionStore.Store
	userStore            userStore.Store
}

func NewStandard(versionReporter version.Reporter, environmentReporter environment.Reporter, logger log.Logger,
	metricServicesClient metricservicesClient.Client, userServicesClient userservicesClient.Client, dataServicesClient dataservicesClient.Client,
	messageStore messageStore.Store, notificationStore notificationStore.Store, permissionStore permissionStore.Store,
	profileStore profileStore.Store, sessionStore sessionStore.Store, userStore userStore.Store) (*Standard, error) {
	if versionReporter == nil {
		return nil, app.Error("api", "version reporter is missing")
	}
	if environmentReporter == nil {
		return nil, app.Error("api", "environment reporter is missing")
	}
	if logger == nil {
		return nil, app.Error("api", "logger is missing")
	}
	if metricServicesClient == nil {
		return nil, app.Error("api", "metric services client is missing")
	}
	if userServicesClient == nil {
		return nil, app.Error("api", "user services client is missing")
	}
	if dataServicesClient == nil {
		return nil, app.Error("api", "data services client is missing")
	}
	if messageStore == nil {
		return nil, app.Error("api", "message store is missing")
	}
	if notificationStore == nil {
		return nil, app.Error("api", "notification store is missing")
	}
	if permissionStore == nil {
		return nil, app.Error("api", "permission store is missing")
	}
	if profileStore == nil {
		return nil, app.Error("api", "profile store is missing")
	}
	if sessionStore == nil {
		return nil, app.Error("api", "session store is missing")
	}
	if userStore == nil {
		return nil, app.Error("api", "user store is missing")
	}

	standard, err := api.NewStandard(versionReporter, environmentReporter, logger)
	if err != nil {
		return nil, err
	}

	return &Standard{
		Standard:             standard,
		metricServicesClient: metricServicesClient,
		userServicesClient:   userServicesClient,
		dataServicesClient:   dataServicesClient,
		messageStore:         messageStore,
		notificationStore:    notificationStore,
		permissionStore:      permissionStore,
		profileStore:         profileStore,
		sessionStore:         sessionStore,
		userStore:            userStore,
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
	return context.WithContext(s.metricServicesClient, s.userServicesClient, s.dataServicesClient,
		s.messageStore, s.notificationStore, s.permissionStore, s.profileStore,
		s.sessionStore, s.userStore, handler)
}
