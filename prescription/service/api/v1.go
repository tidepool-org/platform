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
	ctx := req.Context()
	responder := request.MustNewResponder(res, req)
	usr := r.getUserOrRespondWithError(req, responder)
	if usr == nil {
		return
	}

	create := prescription.NewRevisionCreate()
	if err := request.DecodeRequestBody(req.Request, create); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	// TODO: check prescription permission
	prescr, err := r.PrescriptionClient().CreatePrescription(ctx, *usr.UserID, create)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusCreated, prescr)
}

func (r *Router) ListPrescriptions(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	responder := request.MustNewResponder(res, req)

	pagination := page.NewPagination()
	if err := request.DecodeRequestQuery(req.Request, pagination); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	usr := r.getUserOrRespondWithError(req, responder)
	if usr == nil {
		return
	}

	// TODO: handle clinic access
	filter := prescription.NewFilter()
	filter.ClinicianID = *usr.UserID

	prescr, err := r.PrescriptionClient().ListPrescriptions(ctx, filter, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, prescr)
}

func (r *Router) GetUnclaimedPrescription(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	responder := request.MustNewResponder(res, req)
	accessCode := req.PathParam("accessCode")

	if accessCode == "" {
		responder.Error(http.StatusBadRequest, request.ErrorBadRequest())
		return
	}

	usr := r.getUserOrRespondWithError(req, responder)
	if usr == nil {
		return
	}

	prescr, err := r.PrescriptionClient().GetUnclaimedPrescription(ctx, accessCode)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, prescr)
}

func (r *Router) GetPrescription(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	responder := request.MustNewResponder(res, req)
	PrescriptionID := req.PathParam("id")
	usr := r.getUserOrRespondWithError(req, responder)
	if usr == nil {
		return
	} else if !usr.HasRole(user.RoleClinic) {
		responder.Error(http.StatusUnauthorized, request.ErrorUnauthorized())
		return
	}

	// TODO: handle clinic access
	filter := prescription.NewFilter()
	filter.ID = PrescriptionID
	if usr.HasRole(user.RoleClinic) {
		filter.ClinicianID = *usr.UserID
	}

	pagination := &page.Pagination{Page: 0, Size: 1}
	prescr, err := r.PrescriptionClient().ListPrescriptions(ctx, filter, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	if prescr == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFound())
		return
	}

	responder.Data(http.StatusOK, prescr)
}

func (r *Router) DeletePrescription(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	responder := request.MustNewResponder(res, req)
	prescriptionID := req.PathParam("id")
	usr := r.getUserOrRespondWithError(req, responder)
	if usr == nil {
		return
	}

	if !usr.HasRole(user.RoleClinic) {
		responder.Error(http.StatusUnauthorized, request.ErrorUnauthorized())
	}

	success, err := r.PrescriptionClient().DeletePrescription(ctx, *usr.UserID, prescriptionID)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	if success {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFound())
		return
	}

	responder.Empty(http.StatusOK)
}

func (r *Router) getUserOrRespondWithError(req *rest.Request, responder *request.Responder) *user.User {
	ctx := req.Context()
	details := request.DetailsFromContext(ctx)
	userID := details.UserID()

	if userID == "" {
		responder.Error(http.StatusUnauthorized, request.ErrorUnauthorized())
		return nil
	}

	usr, err := r.UserClient().Get(ctx, userID)
	if err != nil {
		responder.Error(http.StatusInternalServerError, request.ErrorInternalServerError(err))
		return nil
	}

	if usr == nil || usr.UserID == nil || userID != *usr.UserID {
		responder.Error(http.StatusUnauthorized, request.ErrorUnauthorized())
		return nil
	}

	return usr
}
