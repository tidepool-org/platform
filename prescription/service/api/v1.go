package api

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/prescription"
	"github.com/tidepool-org/platform/request"
)

func (r *Router) CreatePrescription(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	details := request.DetailsFromContext(req.Context())
	userID := details.UserID()

	// TODO: check prescription role
	if userID == "" {
		responder.Error(http.StatusUnauthorized, request.ErrorUnauthorized())
		return
	}

	create := prescription.NewRevisionCreate()
	if err := request.DecodeRequestBody(req.Request, create); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	prescr, err := r.PrescriptionClient().CreatePrescription(req.Context(), userID, create)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusCreated, prescr)
}
