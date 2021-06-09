package api

import (
	"net/http"

	structureValidator "github.com/tidepool-org/platform/structure/validator"

	"github.com/tidepool-org/platform/page"

	"github.com/tidepool-org/platform/user"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/prescription"
	"github.com/tidepool-org/platform/request"
)

func (r *Router) CreatePrescription(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	details := request.DetailsFromContext(ctx)
	responder := request.MustNewResponder(res, req)

	userID := details.UserID()
	usr := r.getUserOrRespondWithError(req, responder, userID, user.RoleClinic)
	if usr == nil {
		return
	}

	clinicId := req.PathParam("clinicId")
	create := prescription.NewRevisionCreate(clinicId, userID, true)
	if err := request.DecodeRequestBody(req.Request, create); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	validator := structureValidator.New().WithReference("initialSettings")
	if err := r.deviceSettingsValidator.Validate(ctx, create.InitialSettings, validator); err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}
	if err := validator.Error(); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	// TODO: check prescription permission
	prescr, err := r.prescriptionService.CreatePrescription(ctx, create)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusCreated, prescr)
}

func (r *Router) ListCurrentUserPrescriptions(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	details := request.DetailsFromContext(ctx)
	responder := request.MustNewResponder(res, req)

	userID := details.UserID()
	r.listPrescriptionsForUserID(req, responder, userID)
}

func (r *Router) ListUserPrescriptions(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	details := request.DetailsFromContext(ctx)
	responder := request.MustNewResponder(res, req)

	userID := req.PathParam("userId")
	if !r.canAccessPrescriptionsForRequestUserID(details, userID) {
		responder.Error(http.StatusUnauthorized, request.ErrorUnauthorized())
		return
	}

	r.listPrescriptionsForUserID(req, responder, userID)
}

func (r *Router) GetPrescription(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	details := request.DetailsFromContext(ctx)
	responder := request.MustNewResponder(res, req)

	prescriptionID := req.PathParam("prescriptionId")
	userID := details.UserID()
	usr := r.getUserOrRespondWithError(req, responder, userID)
	if usr == nil {
		return
	}

	// TODO: handle clinic access
	filter, err := prescription.NewFilter(usr)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}
	filter.ID = prescriptionID

	pagination := &page.Pagination{Page: 0, Size: 1}
	prescr, err := r.prescriptionService.ListPrescriptions(ctx, filter, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	if prescr == nil || len(prescr) == 0 {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFound())
		return
	}

	responder.Data(http.StatusOK, prescr[0])
}

func (r *Router) DeletePrescription(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	details := request.DetailsFromContext(ctx)
	responder := request.MustNewResponder(res, req)

	clinicId := req.PathParam("clinicId")
	prescriptionID := req.PathParam("prescriptionId")
	userID := details.UserID()
	usr := r.getUserOrRespondWithError(req, responder, userID, user.RoleClinic)
	if usr == nil {
		return
	}

	success, err := r.prescriptionService.DeletePrescription(ctx, clinicId, prescriptionID, *usr.UserID)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	if !success {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFound())
		return
	}

	responder.Empty(http.StatusOK)
}

func (r *Router) AddRevision(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	details := request.DetailsFromContext(ctx)
	responder := request.MustNewResponder(res, req)

	clinicId := req.PathParam("clinicId")
	prescriptionID := req.PathParam("prescriptionId")
	userID := details.UserID()

	usr := r.getUserOrRespondWithError(req, responder, userID, user.RoleClinic)
	if usr == nil {
		return
	}

	create := prescription.NewRevisionCreate(clinicId, userID, true)
	if err := request.DecodeRequestBody(req.Request, create); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	validator := structureValidator.New().WithReference("initialSettings")
	if err := r.deviceSettingsValidator.Validate(ctx, create.InitialSettings, validator); err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}
	if err := validator.Error(); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	// TODO: check prescription permission
	prescr, err := r.prescriptionService.AddRevision(ctx, prescriptionID, create)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	} else if prescr == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFound())
		return
	}

	responder.Data(http.StatusOK, prescr)
}

func (r *Router) ClaimPrescription(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	details := request.DetailsFromContext(ctx)
	responder := request.MustNewResponder(res, req)

	userID := details.UserID()
	usr := r.getUserOrRespondWithError(req, responder, userID)
	if usr == nil {
		return
	}

	if !usr.IsPatient() {
		responder.Error(http.StatusForbidden, request.ErrorUnauthorized())
		return
	}

	claim := prescription.NewPrescriptionClaim(*usr.UserID)
	if err := request.DecodeRequestBody(req.Request, claim); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	prescr, err := r.prescriptionService.ClaimPrescription(ctx, claim)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	} else if prescr == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFound())
		return
	}

	responder.Data(http.StatusOK, prescr)
}

func (r *Router) UpdateState(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	details := request.DetailsFromContext(ctx)
	responder := request.MustNewResponder(res, req)

	prescriptionID := req.PathParam("prescriptionId")
	userID := details.UserID()
	usr := r.getUserOrRespondWithError(req, responder, userID)
	if usr == nil {
		return
	}

	if !usr.IsPatient() {
		responder.Error(http.StatusUnauthorized, request.ErrorUnauthorized())
		return
	}

	update := prescription.NewStateUpdate(userID)
	if err := request.DecodeRequestBody(req.Request, update); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	prescr, err := r.prescriptionService.UpdatePrescriptionState(ctx, prescriptionID, update)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	} else if prescr == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFound())
		return
	}

	responder.Data(http.StatusOK, prescr)
}

func (r *Router) listPrescriptionsForUserID(req *rest.Request, responder *request.Responder, userID string) {
	ctx := req.Context()
	usr := r.getUserOrRespondWithError(req, responder, userID)
	if usr == nil {
		return
	}

	filter, err := prescription.NewFilter(usr)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	pagination := page.NewPagination()
	if err := request.DecodeRequestQuery(req.Request, filter, pagination); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	prescr, err := r.prescriptionService.ListPrescriptions(ctx, filter, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, prescr)
}

func (r *Router) canAccessPrescriptionsForRequestUserID(details request.Details, requestedUserID string) bool {
	currentUserID := details.UserID()
	return details.IsService() || currentUserID == requestedUserID
}

func (r *Router) getUserOrRespondWithError(req *rest.Request, responder *request.Responder, userID string, requiredRoles ...string) *user.User {
	ctx := req.Context()
	if userID == "" {
		responder.Error(http.StatusUnauthorized, request.ErrorUnauthorized())
		return nil
	}

	usr, err := r.userClient.Get(ctx, userID)
	if err != nil {
		responder.Error(http.StatusInternalServerError, request.ErrorInternalServerError(err))
		return nil
	}

	if usr == nil || usr.UserID == nil || userID != *usr.UserID {
		responder.Error(http.StatusUnauthorized, request.ErrorUnauthorized())
		return nil
	}

	for i := 0; i < len(requiredRoles); i++ {
		if !usr.HasRole(requiredRoles[i]) {
			responder.Error(http.StatusForbidden, request.ErrorUnauthorized())
			return nil
		}
	}

	return usr
}
