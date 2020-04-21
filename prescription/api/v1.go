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
	usr := r.getUserOrRespondWithError(req, responder, user.RoleClinic)
	if usr == nil {
		return
	}

	create := prescription.NewRevisionCreate()
	if err := request.DecodeRequestBody(req.Request, create); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	// TODO: check prescription permission
	prescr, err := r.PrescriptionService().CreatePrescription(ctx, *usr.UserID, create)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusCreated, prescr)
}

func (r *Router) ListPrescriptions(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	responder := request.MustNewResponder(res, req)
	usr := r.getUserOrRespondWithError(req, responder)
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

	prescr, err := r.PrescriptionService().ListPrescriptions(ctx, filter, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, prescr)
}

func (r *Router) GetPrescription(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	responder := request.MustNewResponder(res, req)
	prescriptionID := req.PathParam("prescriptionId")
	usr := r.getUserOrRespondWithError(req, responder)
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
	prescr, err := r.PrescriptionService().ListPrescriptions(ctx, filter, pagination)
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
	responder := request.MustNewResponder(res, req)
	prescriptionID := req.PathParam("prescriptionId")
	usr := r.getUserOrRespondWithError(req, responder, user.RoleClinic)
	if usr == nil {
		return
	}

	success, err := r.PrescriptionService().DeletePrescription(ctx, *usr.UserID, prescriptionID)
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

func (r *Router) AddRevision(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	responder := request.MustNewResponder(res, req)
	prescriptionID := req.PathParam("prescriptionId")
	usr := r.getUserOrRespondWithError(req, responder, user.RoleClinic)
	if usr == nil {
		return
	}

	create := prescription.NewRevisionCreate()
	if err := request.DecodeRequestBody(req.Request, create); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	// TODO: check prescription permission
	prescr, err := r.PrescriptionService().AddRevision(ctx, usr, prescriptionID, create)
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
	responder := request.MustNewResponder(res, req)
	usr := r.getUserOrRespondWithError(req, responder)
	if usr == nil {
		return
	}

	if !usr.IsPatient() {
		responder.Error(http.StatusUnauthorized, request.ErrorUnauthorized())
		return
	}

	claim := prescription.NewPrescriptionClaim()
	if err := request.DecodeRequestBody(req.Request, claim); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	prescr, err := r.PrescriptionService().ClaimPrescription(ctx, usr, claim)
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
	responder := request.MustNewResponder(res, req)
	prescriptionID := req.PathParam("prescriptionId")
	usr := r.getUserOrRespondWithError(req, responder)
	if usr == nil {
		return
	}

	if !usr.IsPatient() {
		responder.Error(http.StatusUnauthorized, request.ErrorUnauthorized())
		return
	}

	update := prescription.NewStateUpdate()
	if err := request.DecodeRequestBody(req.Request, update); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	prescr, err := r.PrescriptionService().UpdatePrescriptionState(ctx, usr, prescriptionID, update)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	} else if prescr == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFound())
		return
	}

	responder.Data(http.StatusOK, prescr)
}

func (r *Router) getUserOrRespondWithError(req *rest.Request, responder *request.Responder, requiredRoles ...string) *user.User {
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

	for i := 0; i < len(requiredRoles); i++ {
		if !usr.HasRole(requiredRoles[i]) {
			responder.Error(http.StatusForbidden, request.ErrorUnauthorized())
			return nil
		}
	}

	return usr
}
