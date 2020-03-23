package api

import (
	"net/http"

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
	if err != nil || !usr.HasRole(user.RoleClinic) {
		responder.Error(http.StatusUnauthorized, request.ErrorUnauthorized())
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
