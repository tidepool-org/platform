package v1

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/appvalidate"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service/api"
	"github.com/tidepool-org/platform/structure"
	structValidator "github.com/tidepool-org/platform/structure/validator"
)

func (r *Router) AppValidateRoutes() []*rest.Route {
	return []*rest.Route{
		rest.Post("/v1/attestations/challenges", api.RequireUser(r.CreateAttestationChallenge)),
		rest.Post("/v1/assertions/challenges", api.RequireUser(r.CreateAssertionChallenge)),
		rest.Post("/v1/attestations/verifications", api.RequireUser(r.VerifyAttestation)),
	}
}

func (r *Router) CreateAttestationChallenge(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	details := request.DetailsFromContext(req.Context())
	ctx := req.Context()

	challengeCreate := appvalidate.NewChallengeCreate(details.UserID())
	if decodeValidateBodyFailed(responder, req.Request, challengeCreate) {
		return
	}

	result, err := r.AppValidator().CreateAttestChallenge(ctx, challengeCreate)
	if responder.RespondIfError(err) {
		fields := log.Fields{
			"userID": details.UserID(),
			"keyId":  challengeCreate.KeyID,
		}
		log.LoggerFromContext(ctx).WithFields(fields).WithError(err).Error("unable to create attestation challenge")
		return
	}
	responder.Data(http.StatusCreated, result)
}

func (r *Router) CreateAssertionChallenge(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	details := request.DetailsFromContext(req.Context())
	ctx := req.Context()

	challengeCreate := appvalidate.NewChallengeCreate(details.UserID())
	if decodeValidateBodyFailed(responder, req.Request, challengeCreate) {
		return
	}

	result, err := r.AppValidator().CreateAssertChallenge(ctx, challengeCreate)
	if responder.RespondIfError(err) {
		fields := log.Fields{
			"userID": details.UserID(),
			"keyId":  challengeCreate.KeyID,
		}
		log.LoggerFromContext(ctx).WithFields(fields).WithError(err).Error("unable to create assertion challenge")
		return
	}
	responder.Data(http.StatusCreated, result)
}

func (r *Router) VerifyAttestation(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	details := request.DetailsFromContext(req.Context())
	ctx := req.Context()

	attestVerify := appvalidate.NewAttestationVerify(details.UserID())
	if decodeValidateBodyFailed(responder, req.Request, attestVerify) {
		return
	}

	err := r.AppValidator().VerifyAttestation(ctx, attestVerify)
	if responder.RespondIfError(err) {
		fields := log.Fields{
			"userID": details.UserID(),
			"keyId":  attestVerify.KeyID,
		}
		log.LoggerFromContext(ctx).WithFields(fields).WithError(err).Error("unable to verify attestation")
		return
	}
	responder.Empty(http.StatusNoContent)
}

func decodeValidateBodyFailed(responder *request.Responder, req *http.Request, body structure.Validatable) bool {
	if err := request.DecodeRequestBody(req, body); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return true
	}
	if err := structValidator.New().Validate(body); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return true
	}
	return false
}
