package v1

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
	serviceApi "github.com/tidepool-org/platform/service/api"
	"net/http"
)

func (r *Router) ConsentRecordRoutes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/v1/consents", serviceApi.RequireServer(r.ListConsents)),
	}
}

func (r *Router) ListConsentRecords(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	userID := req.PathParam("userId")
	if userID == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("userId"))
		return
	}

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
