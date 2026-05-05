package v1

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oauth"
	"github.com/tidepool-org/platform/oura"
	ouraDataWorkEvent "github.com/tidepool-org/platform/oura/data/work/event"
	ouraWebhook "github.com/tidepool-org/platform/oura/webhook"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
)

const (
	RequestHeaderSignature = "X-Oura-Signature"
	RequestHeaderTimestamp = "X-Oura-Timestamp"

	RequestBodySizeMaximum = 1024 * 1024
)

func (r *Router) Event(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Context()
	lgr := log.LoggerFromContext(ctx)

	// Get required headers
	headers := req.Header
	signature := headers.Get(RequestHeaderSignature)
	lgr = lgr.WithField("signature", signature)
	if signature == "" {
		lgr.Error("signature is missing")
		responder.String(http.StatusBadRequest, "signature is missing")
		return
	}
	timestamp := headers.Get(RequestHeaderTimestamp)
	lgr = lgr.WithField("timestamp", timestamp)
	if timestamp == "" {
		lgr.Error("timestamp is missing")
		responder.String(http.StatusBadRequest, "timestamp is missing")
		return
	}

	// Read entire buffer since we need to use for authorization and for parsing
	body, err := io.ReadAll(io.LimitReader(req.Body, RequestBodySizeMaximum+1))
	if err != nil {
		lgr.WithError(err).Error("unable to read request body")
		responder.InternalServerError(errors.Wrap(err, "unable to read request body")) // HTTP failure to force retry later
		return
	}
	bodySize := len(body)
	lgr = lgr.WithField("bodySize", bodySize)
	if bodySize > RequestBodySizeMaximum {
		lgr.Error("request body size exceeds maximum allowed size")
		responder.String(http.StatusBadRequest, "request body size exceeds maximum allowed size") // HTTP failure to force retry later
		return
	}

	// Calculate the signature and authorize
	calculatedSignature, err := CalculateSignature(r.OuraClient.ClientSecret(), timestamp, body)
	if err != nil {
		lgr.WithError(err).Error("unable to calculate signature")
		responder.InternalServerError(errors.Wrap(err, "unable to calculate signature"))
		return
	} else if !hmac.Equal([]byte(calculatedSignature), []byte(signature)) {
		lgr.WithField("calculatedSignature", calculatedSignature).Error("signature is invalid")
		responder.String(http.StatusForbidden, "signature is invalid")
		return
	}

	// Parse the event
	event := &ouraWebhook.Event{}
	if err = request.DecodeStream(ctx, structure.NewPointerSource(), bytes.NewReader(body), event); err != nil {
		lgr.WithError(err).Error("unable to parse request body")
		responder.String(http.StatusBadRequest, "unable to parse request body") // HTTP failure to force retry later
		return
	}
	lgr = lgr.WithField("event", event)

	// Find the associated provider session
	providerSessionFilter := &auth.ProviderSessionFilter{
		Type:       pointer.From(oauth.ProviderType),
		Name:       pointer.From(oura.ProviderName),
		ExternalID: event.UserID,
	}
	providerSessions, err := page.Collect(func(pagination page.Pagination) (auth.ProviderSessions, error) {
		return r.AuthClient.ListProviderSessions(ctx, providerSessionFilter, &pagination)
	})
	if err != nil {
		lgr.WithError(err).Error("unable to get provider sessions")
		responder.InternalServerError(errors.Wrap(err, "unable to get provider sessions")) // HTTP failure to force retry later
		return
	} else if providerSessionsCount := len(providerSessions); providerSessionsCount < 1 {
		lgr.Error("provider session is missing")
		responder.InternalServerError(errors.New("provider session is missing")) // HTTP failure to force retry later
		return
	}

	for _, providerSession := range providerSessions {
		lgr = lgr.WithField("providerSessionId", providerSession.ID)

		// Create the work
		if workCreate, err := ouraDataWorkEvent.NewWorkCreate(providerSession.ID, event); err != nil {
			lgr.WithError(err).Error("unable to create work create")
			responder.InternalServerError(errors.Wrap(err, "unable to create work create"))
			return
		} else if _, err := r.WorkClient.Create(ctx, workCreate); err != nil {
			lgr.WithError(err).Error("unable to create work")
			responder.InternalServerError(errors.Wrap(err, "unable to create work"))
			return
		}
	}

	responder.String(http.StatusOK, "OK")
}

func CalculateSignature(secret string, timestamp string, bytes []byte) (string, error) {
	hash := hmac.New(sha256.New, []byte(secret))
	if _, err := hash.Write([]byte(timestamp)); err != nil {
		return "", errors.Wrap(err, "unable to write timestamp to hash")
	} else if _, err := hash.Write(bytes); err != nil {
		return "", errors.Wrap(err, "unable to write bytes to hash")
	}
	return strings.ToUpper(hex.EncodeToString(hash.Sum(nil))), nil
}
