package v1

import (
	"net/http"

	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
)

const (
	RequestQueryParameterVerificationToken = "verification_token"
	RequestQueryParameterChallenge         = "challenge"
)

func Subscription(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()
	ctx := req.Context()
	lgr := log.LoggerFromContext(ctx)
	responder := request.MustNewResponder(res, req)

	// Get required query parameters
	query := req.URL.Query()
	verificationToken := query.Get(RequestQueryParameterVerificationToken)
	lgr = lgr.WithField("verificationToken", verificationToken)
	if verificationToken == "" {
		lgr.Error("verification token is missing")
		responder.String(http.StatusForbidden, "verification token is missing")
		return
	}
	challenge := query.Get(RequestQueryParameterChallenge)
	lgr = lgr.WithField("challenge", challenge)
	if challenge == "" {
		lgr.Error("challenge is missing")
		responder.String(http.StatusForbidden, "challenge is missing")
		return
	}

	// Ensure verification token matches partner secret
	if verificationToken != dataServiceContext.OuraClient().PartnerSecret() {
		lgr.Error("verification token is invalid")
		responder.String(http.StatusForbidden, "verification token is invalid")
		return
	}

	// Return challenge
	responder.Data(http.StatusOK, map[string]string{"challenge": challenge})
}
