package api

import (
	"github.com/ant0ine/go-json-rest/rest"

	confirmationStore "github.com/tidepool-org/platform/confirmation/store"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/errors"
	messageStore "github.com/tidepool-org/platform/message/store"
	"github.com/tidepool-org/platform/metric"
	permissionStore "github.com/tidepool-org/platform/permission/store"
	profileStore "github.com/tidepool-org/platform/profile/store"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/api"
	sessionStore "github.com/tidepool-org/platform/session/store"
	"github.com/tidepool-org/platform/user"
	userService "github.com/tidepool-org/platform/user/service"
	userContext "github.com/tidepool-org/platform/user/service/context"
	userStore "github.com/tidepool-org/platform/user/store"
)

type Standard struct {
	*api.API
	dataClient        dataClient.Client
	metricClient      metric.Client
	userClient        user.Client
	confirmationStore confirmationStore.Store
	messageStore      messageStore.Store
	permissionStore   permissionStore.Store
	profileStore      profileStore.Store
	sessionStore      sessionStore.Store
	userStore         userStore.Store
}

func NewStandard(svc service.Service, dataClient dataClient.Client, metricClient metric.Client, userClient user.Client,
	confirmationStore confirmationStore.Store, messageStore messageStore.Store, permissionStore permissionStore.Store,
	profileStore profileStore.Store, sessionStore sessionStore.Store, userStore userStore.Store) (*Standard, error) {
	if dataClient == nil {
		return nil, errors.New("data client is missing")
	}
	if metricClient == nil {
		return nil, errors.New("metric client is missing")
	}
	if userClient == nil {
		return nil, errors.New("user client is missing")
	}
	if confirmationStore == nil {
		return nil, errors.New("confirmation store is missing")
	}
	if messageStore == nil {
		return nil, errors.New("message store is missing")
	}
	if permissionStore == nil {
		return nil, errors.New("permission store is missing")
	}
	if profileStore == nil {
		return nil, errors.New("profile store is missing")
	}
	if sessionStore == nil {
		return nil, errors.New("session store is missing")
	}
	if userStore == nil {
		return nil, errors.New("user store is missing")
	}

	a, err := api.New(svc)
	if err != nil {
		return nil, err
	}

	return &Standard{
		API:               a,
		dataClient:        dataClient,
		metricClient:      metricClient,
		userClient:        userClient,
		confirmationStore: confirmationStore,
		messageStore:      messageStore,
		permissionStore:   permissionStore,
		profileStore:      profileStore,
		sessionStore:      sessionStore,
		userStore:         userStore,
	}, nil
}

func (s *Standard) DEPRECATEDInitializeRouter(routes []userService.Route) error {
	baseRoutes := []userService.Route{
		userService.MakeRoute("GET", "/status", s.StatusGet),
		userService.MakeRoute("GET", "/version", s.VersionGet),
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
		return errors.Wrap(err, "unable to create router")
	}

	s.DEPRECATEDAPI().SetApp(router)

	return nil
}

func (s *Standard) withContext(handler userService.HandlerFunc) rest.HandlerFunc {
	return userContext.WithContext(s.AuthClient(), s.dataClient, s.metricClient, s.userClient,
		s.confirmationStore, s.messageStore, s.permissionStore, s.profileStore,
		s.sessionStore, s.userStore, handler)
}
