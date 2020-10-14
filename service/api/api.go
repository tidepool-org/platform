package api

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/middleware"
)

type API struct {
	service.Service
	api              *rest.Api
	statusMiddleware *rest.StatusMiddleware
}

func New(svc service.Service) (*API, error) {
	if svc == nil {
		return nil, errors.New("service is missing")
	}

	return &API{
		Service: svc,
		api:     rest.NewApi(),
	}, nil
}

func (a *API) Status() *rest.Status {
	if a.statusMiddleware == nil {
		return nil
	}
	return a.statusMiddleware.GetStatus()
}

func (a *API) DEPRECATEDAPI() *rest.Api {
	return a.api
}

func (a *API) Handler() http.Handler {
	return a.api.MakeHandler()
}

func (a *API) InitializeMiddleware() error {
	loggerMiddleware, err := middleware.NewLogger(a.Logger())
	if err != nil {
		return err
	}
	errorMiddleware, err := middleware.NewError()
	if err != nil {
		return err
	}
	traceMiddleware, err := middleware.NewTrace()
	if err != nil {
		return err
	}
	//accessLogMiddleware, err := middleware.NewAccessLog()
	//if err != nil {
		//return err
	//}
	recoverMiddleware, err := middleware.NewRecover()
	if err != nil {
		return err
	}
	authMiddleware, err := middleware.NewAuth(a.Secret(), a.AuthClient())
	if err != nil {
		return err
	}

	statusMiddleware := &rest.StatusMiddleware{}
	timerMiddleware := &rest.TimerMiddleware{}
	recorderMiddleware := &rest.RecorderMiddleware{}
	gzipMiddleware := &rest.GzipMiddleware{}

	middlewareStack := []rest.Middleware{
		loggerMiddleware,
		errorMiddleware,
		traceMiddleware,
		accessLogMiddleware,
		statusMiddleware,
		timerMiddleware,
		recorderMiddleware,
		recoverMiddleware,
		authMiddleware,
		gzipMiddleware,
	}

	a.api.Use(middlewareStack...)

	a.statusMiddleware = statusMiddleware

	return nil
}

func (a *API) InitializeRouters(routers ...service.Router) error {
	routes := []*rest.Route{}

	for _, router := range routers {
		routes = append(routes, router.Routes()...)
	}

	router, err := rest.MakeRouter(routes...)
	if err != nil {
		return errors.Wrap(err, "unable to initializer router")
	}

	a.api.SetApp(router)

	return nil
}
