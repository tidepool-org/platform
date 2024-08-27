package api

import (
	context2 "context"
	"net/http"

	clinic "github.com/tidepool-org/clinic/client"

	"github.com/tidepool-org/platform/clinics"

	structureValidator "github.com/tidepool-org/platform/structure/validator"

	"github.com/tidepool-org/platform/page"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/prescription"
	"github.com/tidepool-org/platform/request"
)

func (r *Router) CreatePrescription(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	details := request.GetAuthDetails(ctx)
	responder := request.MustNewResponder(res, req)

	userID := details.UserID()
	clinicID := req.PathParam("clinicId")
	clinician := r.getClinicianOrRespondWithError(ctx, clinicID, userID, responder)
	if clinician == nil {
		return
	}

	create := prescription.NewRevisionCreate(clinicID, userID, clinics.IsPrescriber(clinician))
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

	prescr, err := r.prescriptionService.CreatePrescription(ctx, create)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusCreated, prescr)
}

func (r *Router) ListClinicPrescriptions(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	details := request.GetAuthDetails(ctx)
	responder := request.MustNewResponder(res, req)

	userID := details.UserID()
	clinicID := req.PathParam("clinicId")
	clinician := r.getClinicianOrRespondWithError(ctx, clinicID, userID, responder)
	if clinician == nil {
		return
	}

	filter, err := prescription.NewClinicFilter(clinicID)
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

func (r *Router) ListUserPrescriptions(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	details := request.GetAuthDetails(ctx)
	responder := request.MustNewResponder(res, req)

	userID := req.PathParam("userId")
	if !r.canAccessPrescriptionsForRequestUserID(details, userID) {
		responder.Error(http.StatusForbidden, request.ErrorUnauthorized())
		return
	}

	filter, err := prescription.NewPatientFilter(userID)
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

func (r *Router) GetClinicPrescription(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	details := request.GetAuthDetails(ctx)
	responder := request.MustNewResponder(res, req)

	userID := details.UserID()
	clinicID := req.PathParam("clinicId")
	prescriptionID := req.PathParam("prescriptionId")
	clinician := r.getClinicianOrRespondWithError(ctx, clinicID, userID, responder)
	if clinician == nil {
		return
	}

	filter, err := prescription.NewClinicFilter(clinicID)
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

	if len(prescr) == 0 {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFound())
		return
	}

	responder.Data(http.StatusOK, prescr[0])
}

func (r *Router) GetPatientPrescription(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	details := request.GetAuthDetails(ctx)
	responder := request.MustNewResponder(res, req)

	userID := req.PathParam("userId")
	prescriptionID := req.PathParam("prescriptionId")
	if !r.canAccessPrescriptionsForRequestUserID(details, userID) {
		responder.Error(http.StatusForbidden, request.ErrorUnauthorized())
		return
	}

	filter, err := prescription.NewPatientFilter(userID)
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

	if len(prescr) == 0 {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFound())
		return
	}

	responder.Data(http.StatusOK, prescr[0])
}

func (r *Router) DeletePrescription(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	details := request.GetAuthDetails(ctx)
	responder := request.MustNewResponder(res, req)

	userID := details.UserID()
	clinicID := req.PathParam("clinicId")
	prescriptionID := req.PathParam("prescriptionId")
	clinician := r.getClinicianOrRespondWithError(ctx, clinicID, userID, responder)
	if clinician == nil {
		return
	}

	success, err := r.prescriptionService.DeletePrescription(ctx, clinicID, prescriptionID, userID)
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
	details := request.GetAuthDetails(ctx)
	responder := request.MustNewResponder(res, req)

	userID := details.UserID()
	clinicID := req.PathParam("clinicId")
	prescriptionID := req.PathParam("prescriptionId")

	clinician := r.getClinicianOrRespondWithError(ctx, clinicID, userID, responder)
	if clinician == nil {
		return
	}

	create := prescription.NewRevisionCreate(clinicID, userID, true)
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
	details := request.GetAuthDetails(ctx)
	responder := request.MustNewResponder(res, req)

	userID := details.UserID()
	if userID != req.PathParam("userId") {
		responder.Error(http.StatusUnauthorized, request.ErrorUnauthorized())
		return
	}

	claim := prescription.NewPrescriptionClaim(userID)
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
	details := request.GetAuthDetails(ctx)
	responder := request.MustNewResponder(res, req)

	userID := details.UserID()
	prescriptionID := req.PathParam("prescriptionId")
	if userID != req.PathParam("userId") {
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

func (r *Router) canAccessPrescriptionsForRequestUserID(details request.AuthDetails, requestedUserID string) bool {
	currentUserID := details.UserID()
	return details.IsService() || currentUserID == requestedUserID
}

func (r *Router) getClinicianOrRespondWithError(ctx context2.Context, clinicID, clinicianID string, responder *request.Responder) *clinic.Clinician {
	clinician, err := r.clinicsClient.GetClinician(ctx, clinicID, clinicianID)
	if err != nil {
		responder.Error(http.StatusInternalServerError, request.ErrorInternalServerError(err))
		return nil
	}
	if clinician == nil {
		responder.Error(http.StatusForbidden, request.ErrorUnauthorized())
		return nil
	}
	return clinician
}
