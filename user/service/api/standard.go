package api

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	confirmationStore "github.com/tidepool-org/platform/confirmation/store"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	messageStore "github.com/tidepool-org/platform/message/store"
	metricClient "github.com/tidepool-org/platform/metric/client"
	permissionStore "github.com/tidepool-org/platform/permission/store"
	profileStore "github.com/tidepool-org/platform/profile/store"
	"github.com/tidepool-org/platform/service/api"
	sessionStore "github.com/tidepool-org/platform/session/store"
	userClient "github.com/tidepool-org/platform/user/client"
	"github.com/tidepool-org/platform/user/service"
	"github.com/tidepool-org/platform/user/service/context"
	userStore "github.com/tidepool-org/platform/user/store"
	"github.com/tidepool-org/platform/version"
)

type Standard struct {
	*api.Standard
	dataClient        dataClient.Client
	metricClient      metricClient.Client
	userClient        userClient.Client
	confirmationStore confirmationStore.Store
	messageStore      messageStore.Store
	permissionStore   permissionStore.Store
	profileStore      profileStore.Store
	sessionStore      sessionStore.Store
	userStore         userStore.Store
}

func NewStandard(versionReporter version.Reporter, logger log.Logger,
	authClient auth.Client, dataClient dataClient.Client, metricClient metricClient.Client, userClient userClient.Client,
	confirmationStore confirmationStore.Store, messageStore messageStore.Store, permissionStore permissionStore.Store,
	profileStore profileStore.Store, sessionStore sessionStore.Store, userStore userStore.Store) (*Standard, error) {
	if dataClient == nil {
		return nil, errors.New("api", "data client is missing")
	}
	if metricClient == nil {
		return nil, errors.New("api", "metric client is missing")
	}
	if userClient == nil {
		return nil, errors.New("api", "user client is missing")
	}
	if confirmationStore == nil {
		return nil, errors.New("api", "confirmation store is missing")
	}
	if messageStore == nil {
		return nil, errors.New("api", "message store is missing")
	}
	if permissionStore == nil {
		return nil, errors.New("api", "permission store is missing")
	}
	if profileStore == nil {
		return nil, errors.New("api", "profile store is missing")
	}
	if sessionStore == nil {
		return nil, errors.New("api", "session store is missing")
	}
	if userStore == nil {
		return nil, errors.New("api", "user store is missing")
	}

	standard, err := api.NewStandard(versionReporter, logger, authClient)
	if err != nil {
		return nil, err
	}

	return &Standard{
		Standard:          standard,
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

func (s *Standard) DEPRECATEDInitializeRouter(routes []service.Route) error {
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
		return errors.Wrap(err, "api", "unable to create router")
	}

	s.API().SetApp(router)

	return nil
}

func (s *Standard) withContext(handler service.HandlerFunc) rest.HandlerFunc {
	return context.WithContext(s.AuthClient(), s.dataClient, s.metricClient, s.userClient,
		s.confirmationStore, s.messageStore, s.permissionStore, s.profileStore,
		s.sessionStore, s.userStore, handler)
}
