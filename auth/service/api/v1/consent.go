package v1

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
	serviceApi "github.com/tidepool-org/platform/service/api"
	"net/http"
)

func (r *Router) ConsentRoutes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/v1/consents", serviceApi.RequireServer(r.ListConsents)),
	}
}

func (r *Router) ListConsents(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	filter := auth.NewConsentFilter()
	pagination := page.NewPagination()
	if err := request.DecodeRequestQuery(req.Request, filter, pagination); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	consents, err := r.AuthClient().ListConsents(req.Context(), filter, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, consents)
}

func (r *Router) GetConsentByType(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	filter := auth.NewConsentFilter()
	pagination := page.NewPagination()
	if err := request.DecodeRequestQuery(req.Request, filter, pagination); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	consents, err := r.AuthClient().ListConsents(req.Context(), filter, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, consents)
}
