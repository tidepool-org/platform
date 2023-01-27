package v1

import (
	"errors"
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
		rest.Post("/v1/attestations/verifications", api.RequireUser(r.VerifyAttestation)),
		rest.Post("/v1/assertions/challenges", api.RequireUser(r.CreateAssertionChallenge)),
		// Rename this route to show actual intent of retrieving secret when secret retrieval is implemented.
		rest.Post("/v1/assertions/verifications", api.RequireUser(r.VerifyAssertion)),
	}
}

func (r *Router) CreateAttestationChallenge(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	details := request.GetAuthDetails(req.Context())
	ctx := req.Context()

	challengeCreate := appvalidate.NewChallengeCreate(details.UserID())
	if decodeValidateBodyFailed(responder, req.Request, challengeCreate) {
		return
	}

	result, err := r.AppValidator().CreateAttestChallenge(ctx, challengeCreate)
	if err != nil {
		fields := log.Fields{
			"userID": details.UserID(),
			"keyId":  challengeCreate.KeyID,
		}
		log.LoggerFromContext(ctx).WithFields(fields).WithError(err).Error("unable to create attestation challenge")
		if errors.Is(err, appvalidate.ErrDuplicateKeyId) {
			responder.Error(http.StatusBadRequest, errors.New("invalid key id"))
			return
		}
		responder.InternalServerError(err)
		return
	}
	responder.Data(http.StatusCreated, result)
}

func (r *Router) CreateAssertionChallenge(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	details := request.GetAuthDetails(req.Context())
	ctx := req.Context()

	challengeCreate := appvalidate.NewChallengeCreate(details.UserID())
	if decodeValidateBodyFailed(responder, req.Request, challengeCreate) {
		return
	}

	result, err := r.AppValidator().CreateAssertChallenge(ctx, challengeCreate)
	if err != nil {
		fields := log.Fields{
			"userID": details.UserID(),
			"keyId":  challengeCreate.KeyID,
		}
		log.LoggerFromContext(ctx).WithFields(fields).WithError(err).Error("unable to create assertion challenge")
		if errors.Is(err, appvalidate.ErrNotVerified) {
			responder.Error(http.StatusBadRequest, err)
			return
		}
		responder.InternalServerError(err)
		return
	}
	responder.Data(http.StatusCreated, result)
}

func (r *Router) VerifyAttestation(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	details := request.GetAuthDetails(req.Context())
	ctx := req.Context()

	attestVerify := appvalidate.NewAttestationVerify(details.UserID())
	if decodeValidateBodyFailed(responder, req.Request, attestVerify) {
		return
	}

	err := r.AppValidator().VerifyAttestation(ctx, attestVerify)
	if err != nil {
		fields := log.Fields{
			"userID": details.UserID(),
			"keyId":  attestVerify.KeyID,
		}
		log.LoggerFromContext(ctx).WithFields(fields).WithError(err).Error("unable to verify attestation")
		if errors.Is(err, appvalidate.ErrAttestationVerificationFailed) {
			responder.Error(http.StatusBadRequest, err)
			return
		}
		responder.InternalServerError(err)
		return
	}

	responder.Empty(http.StatusNoContent)
}

func (r *Router) VerifyAssertion(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	details := request.GetAuthDetails(req.Context())
	ctx := req.Context()

	assertVerify := appvalidate.NewAssertionVerify(details.UserID())
	if decodeValidateBodyFailed(responder, req.Request, assertVerify) {
		return
	}

	if err := r.AppValidator().VerifyAssertion(ctx, assertVerify); err != nil {
		fields := log.Fields{
			"userID": details.UserID(),
			"keyId":  assertVerify.KeyID,
		}
		log.LoggerFromContext(ctx).WithFields(fields).WithError(err).Error("unable to verify assertion")

		if errors.Is(err, appvalidate.ErrAssertionVerificationFailed) || errors.Is(err, appvalidate.ErrNotVerified) {
			responder.Error(http.StatusBadRequest, err)
			return
		}
		responder.InternalServerError(err)
		return
	}
	// Assertion has succeeded, at this point, we would access some secret
	// from a DB, partner API, etc, depending on the AssertionVerify object.
	// r.SecretGetter.GetSecret(...)
	responder.Empty(http.StatusOK)
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
