package v1

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
)

const (
	QueryVerificationToken = "verification_token"
	QueryChallenge         = "challenge"
)

func (r *Router) Subscription(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	query := req.URL.Query()
	ctx := req.Context()
	lgr := log.LoggerFromContext(ctx)

	verificationToken := query.Get(QueryVerificationToken)
	if verificationToken == "" {
		lgr.Error("verification token is missing")
		responder.String(http.StatusForbidden, "verification token is missing")
		return
	}

	challenge := query.Get(QueryChallenge)
	if challenge == "" {
		lgr.Error("challenge is missing")
		responder.String(http.StatusForbidden, "challenge is missing")
		return
	}
	lgr = lgr.WithField("challenge", challenge)

	// Ensure verification token matches partner secret
	if verificationToken != r.OuraClient.PartnerSecret() {
		lgr.Error("verification token is invalid")
		responder.String(http.StatusForbidden, "verification token is invalid")
		return
	}

	// Return challenge
	responder.Data(http.StatusOK, map[string]string{"challenge": challenge})
}
