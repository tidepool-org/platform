package api

import (
	"net/http"

	"github.com/tidepool-org/platform/page"

	"github.com/tidepool-org/platform/user"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/prescription"
	"github.com/tidepool-org/platform/request"
)

func (r *Router) CreatePrescription(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Context()
	details := request.DetailsFromContext(ctx)
	userID := details.UserID()

	if userID == "" {
		responder.Error(http.StatusUnauthorized, request.ErrorUnauthorized())
		return
	}

	usr, err := r.UserClient().Get(ctx, userID)
	if err != nil {
		responder.Error(http.StatusInternalServerError, request.ErrorInternalServerError(err))
		return
	}

	if usr == nil || !usr.HasRole(user.RoleClinic) {
		responder.Error(http.StatusUnauthorized, request.ErrorUnauthorized())
		return
	}

	create := prescription.NewRevisionCreate()
	if err := request.DecodeRequestBody(req.Request, create); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	// TODO: check prescription permission
	prescr, err := r.PrescriptionClient().CreatePrescription(req.Context(), userID, create)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusCreated, prescr)
}

func (r *Router) ListPrescriptions(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Context()
	details := request.DetailsFromContext(ctx)
	userID := details.UserID()

	pagination := page.NewPagination()
	if err := request.DecodeRequestQuery(req.Request, pagination); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	if userID == "" {
		responder.Error(http.StatusUnauthorized, request.ErrorUnauthorized())
		return
	}

	usr, err := r.UserClient().Get(ctx, userID)
	if err != nil {
		responder.Error(http.StatusInternalServerError, request.ErrorInternalServerError(err))
		return
	}

	if usr == nil || usr.UserID == nil {
		responder.Error(http.StatusUnauthorized, request.ErrorUnauthorized())
		return
	}

	// TODO: handle clinic access
	filter := prescription.NewFilter()
	if usr.HasRole(user.RoleClinic) {
		filter.ClinicianID = *usr.UserID
	} else {
		filter.PatientID = *usr.UserID
	}

	prescr, err := r.PrescriptionClient().ListPrescriptions(ctx, filter, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, prescr)
}

func (r *Router) GetUnclaimedPrescription(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Context()
	details := request.DetailsFromContext(ctx)
	userID := details.UserID()
	accessCode := req.PathParam("accessCode")

	if accessCode == "" {
		responder.Error(http.StatusBadRequest, request.ErrorBadRequest())
		return
	}

	if userID == "" {
		responder.Error(http.StatusUnauthorized, request.ErrorUnauthorized())
		return
	}

	usr, err := r.UserClient().Get(ctx, userID)
	if err != nil {
		responder.Error(http.StatusInternalServerError, request.ErrorInternalServerError(err))
		return
	}

	if usr == nil || usr.UserID == nil || usr.HasRole(user.RoleClinic) {
		responder.Error(http.StatusUnauthorized, request.ErrorUnauthorized())
		return
	}

	prescr, err := r.PrescriptionClient().GetUnclaimedPrescription(ctx, accessCode)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, prescr)
}
