package v1

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/udi"
)

const ErrorCodeUDIUnavailable = "udi-unavailable"

func ErrorUDIUnavailable(err error) error {
	return errors.WrapPrepared(err, ErrorCodeUDIUnavailable, "udi unavailable", "unable to identify the release in service")
}

func (r *Router) UDIRoutes() []*rest.Route {
	return []*rest.Route{
		rest.Post("/v1/udi", r.CreateUDI),
	}
}

func (r *Router) CreateUDI(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	udiRequest := udi.NewRequest()
	if err := request.DecodeRequestBody(req.Request, udiRequest); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	release, err := r.ClusterReleaseReporter().Release(req.Context())
	if err != nil {
		responder.Error(http.StatusServiceUnavailable, ErrorUDIUnavailable(err))
		return
	}

	result, err := udi.New(release, *udiRequest.ClientVersion)
	if err != nil {
		responder.Error(http.StatusServiceUnavailable, ErrorUDIUnavailable(err))
		return
	}

	responder.Data(http.StatusOK, result)
}
