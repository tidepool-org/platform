package v1

import (
	"errors"
	"maps"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/appvalidate"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service/api"
	"github.com/tidepool-org/platform/structure"
	structValidator "github.com/tidepool-org/platform/structure/validator"
)

var (
	ErrPartnerSecretsNotInitialized = errors.New("partner secrets not initialized")
)

func (r *Router) AppValidateRoutes() []*rest.Route {
	return []*rest.Route{
		rest.Post("/v1/attestations/challenges", api.RequireUser(r.CreateAttestationChallenge)),
		rest.Post("/v1/attestations/verifications", api.RequireUser(r.VerifyAttestation)),
		rest.Post("/v1/assertions/challenges", api.RequireUser(r.CreateAssertionChallenge)),
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
		if errors.Is(err, appvalidate.ErrKeyIdNotFound) {
			responder.Error(http.StatusNotFound, err)
			return
		}
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

	logFields := log.Fields{
		"userID": details.UserID(),
		"keyId":  assertVerify.KeyID,
	}

	// log debug fields (only in qa environments)
	debugFields := log.Fields{
		"PartnerData": string(assertVerify.ClientData.PartnerData),
	}
	maps.Copy(debugFields, logFields)
	log.LoggerFromContext(ctx).WithFields(debugFields).Debug("appvalidate input")

	if err := r.AppValidator().VerifyAssertion(ctx, assertVerify); err != nil {
		log.LoggerFromContext(ctx).WithFields(logFields).WithError(err).Error("unable to verify assertion")

		if errors.Is(err, appvalidate.ErrAssertionVerificationFailed) || errors.Is(err, appvalidate.ErrNotVerified) {
			responder.Error(http.StatusBadRequest, err)
			return
		}
		responder.InternalServerError(err)
		return
	}

	// Assertion has succeeded, at this point, we would access some secret
	// from a DB, partner API, etc, depending on the AssertionVerify object.
	ps := r.PartnerSecrets()
	if ps == nil {
		responder.Error(http.StatusInternalServerError, ErrPartnerSecretsNotInitialized)
		return
	}
	secret, err := ps.GetSecret(ctx, assertVerify.ClientData)
	if err != nil {
		log.LoggerFromContext(ctx).WithFields(logFields).WithError(err).Errorf("unable to create fetch %v secrets", assertVerify.ClientData.Partner)
		if errors.Is(err, appvalidate.ErrInvalidPartnerPayload) {
			responder.Error(http.StatusBadRequest, err)
			return
		}
		if errors.Is(err, appvalidate.ErrInvalidPartnerCredentials) {
			responder.InternalServerError(err)
			return
		}
		responder.InternalServerError(err)
		return
	}
	log.LoggerFromContext(ctx).WithFields(logFields).Debug("successfully retrieved partner certificates")
	responder.Data(http.StatusOK, appvalidate.AssertionResponse{
		Data: secret,
	})
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
