package v1

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/consent"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	serviceApi "github.com/tidepool-org/platform/service/api"
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

	details := request.GetAuthDetails(req.Context())
	if !details.IsService() && details.UserID() != userID {
		responder.Error(http.StatusForbidden, request.ErrorUnauthorized())
		return
	}

	filter := consent.NewConsentRecordFilter()
	filter.Latest = pointer.FromAny(true)

	pagination := page.NewPagination()
	if err := request.DecodeRequestQuery(req.Request, filter, pagination); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	consents, err := r.service.ListConsentRecords(req.Context(), userID, filter, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, consents)
}

func (r *Router) GetConsentRecord(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	userID := req.PathParam("userId")
	if userID == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("userId"))
		return
	}

	id := req.PathParam("id")
	if id == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("id"))
		return
	}

	details := request.GetAuthDetails(req.Context())
	if !details.IsService() && details.UserID() != userID {
		responder.Error(http.StatusForbidden, request.ErrorUnauthorized())
		return
	}

	consentRecord, err := r.service.GetConsentRecord(req.Context(), userID, id)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, consentRecord)
}

func (r *Router) CreateConsentRecord(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	userID := req.PathParam("userId")
	if userID == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("userId"))
		return
	}

	details := request.GetAuthDetails(req.Context())
	if !details.IsService() && details.UserID() != userID {
		responder.Error(http.StatusForbidden, request.ErrorUnauthorized())
		return
	}

	create := &consent.RecordCreate{}
	if err := request.DecodeRequestBody(req.Request, create); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}
	create.UserID = userID
	create.Status = consent.RecordStatusActive

	consentRecord, err := r.service.CreateConsentRecord(req.Context(), userID, create)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, consentRecord)
}

func (r *Router) UpdateConsentRecord(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	userID := req.PathParam("userId")
	if userID == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("userId"))
		return
	}

	id := req.PathParam("id")
	if id == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("id"))
		return
	}

	details := request.GetAuthDetails(req.Context())
	if !details.IsService() && details.UserID() != userID {
		responder.Error(http.StatusForbidden, request.ErrorUnauthorized())
		return
	}

	consentRecord, err := r.service.GetConsentRecord(req.Context(), userID, id)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	} else if consentRecord == nil {
		responder.Empty(http.StatusNotFound)
		return
	}

	if err = request.DecodeRequestBody(req.Request, consentRecord.ToUpdate()); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	consentRecord, err = r.service.UpdateConsentRecord(req.Context(), consentRecord)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, consentRecord)
}

func (r *Router) RevokeConsentRecord(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	userID := req.PathParam("userId")
	if userID == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("userId"))
		return
	}

	id := req.PathParam("id")
	if id == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("id"))
		return
	}

	details := request.GetAuthDetails(req.Context())
	if !details.IsService() && details.UserID() != userID {
		responder.Error(http.StatusForbidden, request.ErrorUnauthorized())
		return
	}

	revoke := consent.NewConsentRecordRevoke()
	revoke.ID = id

	err := r.service.RevokeConsentRecord(req.Context(), userID, revoke)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusNoContent)
}
