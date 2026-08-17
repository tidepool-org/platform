package v1

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oura"
	ouraWebhook "github.com/tidepool-org/platform/oura/webhook"
	"github.com/tidepool-org/platform/request"
)

const (
	QueryParameterVerificationToken = "verification_token"
	QueryParameterChallenge         = "challenge"
)

func (r *Router) Subscription(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	query := req.URL.Query()
	ctx := req.Context()
	lgr := log.LoggerFromContext(ctx)

	// Get the required path parameters
	eventType, err := request.DecodeRequestPathParameter(req, PathParameterEventType, oura.IsValidEventType)
	if err != nil {
		lgr.WithError(err).Error("event type is invalid")
		responder.String(http.StatusForbidden, err.Error())
		return
	}
	dataType, err := request.DecodeRequestPathParameter(req, PathParameterDataType, oura.IsValidEventDataType)
	if err != nil {
		lgr.WithError(err).Error("data type is invalid")
		responder.String(http.StatusForbidden, err.Error())
		return
	}

	// Get the required query parameters
	verificationToken := query.Get(QueryParameterVerificationToken)
	if verificationToken == "" {
		lgr.Error("verification token is missing")
		responder.String(http.StatusForbidden, "verification token is missing")
		return
	}
	challenge := query.Get(QueryParameterChallenge)
	if challenge == "" {
		lgr.Error("challenge is missing")
		responder.String(http.StatusForbidden, "challenge is missing")
		return
	}

	// Ensure verification token matches expected
	expectedCallbackURL := ouraWebhook.CallbackURLForEvent(r.OuraClient.PartnerURL(), eventType, dataType)
	expectedVerificationToken := ouraWebhook.VerificationTokenForCallbackURL(expectedCallbackURL, r.OuraClient.PartnerSecret())
	if verificationToken != expectedVerificationToken {
		lgr.Error("verification token is invalid")
		responder.String(http.StatusForbidden, "verification token is invalid")
		return
	}

	// Return challenge
	responder.Data(http.StatusOK, map[string]string{"challenge": challenge})
}
