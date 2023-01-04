package v1

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/appvalidate"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service/api"
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
	err := request.DecodeRequestBody(req.Request, challengeCreate)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}
	if err := structValidator.New().Validate(challengeCreate); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	result, err := r.AppValidator().CreateAttestChallenge(ctx, challengeCreate)
	if responder.RespondIfError(err) {
		return
	}
	responder.Data(http.StatusCreated, result)
}

func (r *Router) CreateAssertionChallenge(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	details := request.DetailsFromContext(req.Context())
	ctx := req.Context()

	challengeCreate := appvalidate.NewChallengeCreate(details.UserID())
	err := request.DecodeRequestBody(req.Request, challengeCreate)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}
	if err := structValidator.New().Validate(challengeCreate); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	result, err := r.AppValidator().CreateAssertChallenge(ctx, challengeCreate)
	if responder.RespondIfError(err) {
		return
	}
	responder.Data(http.StatusCreated, result)
}

func (r *Router) VerifyAttestation(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	details := request.DetailsFromContext(req.Context())
	ctx := req.Context()

	attestVerify := appvalidate.NewAttestationVerify(details.UserID())
	err := request.DecodeRequestBody(req.Request, attestVerify)
	if responder.RespondIfError(err) {
		return
	}

	err = r.AppValidator().VerifyAttestation(ctx, attestVerify)
	if responder.RespondIfError(err) {
		return
	}
	responder.Empty(http.StatusNoContent)
}
