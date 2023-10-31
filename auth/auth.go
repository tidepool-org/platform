package auth

import (
	"context"

	"github.com/tidepool-org/platform/request"
)

const (
	TidepoolServiceSecretHeaderKey      = "X-Tidepool-Service-Secret"
	TidepoolAuthorizationHeaderKey      = "Authorization"
	TidepoolSessionTokenHeaderKey       = "X-Tidepool-Session-Token"
	TidepoolRestrictedTokenParameterKey = "restricted_token"
)

type Client interface {
	ExternalAccessor
}

type ExternalAccessor interface {
	ValidateSessionToken(ctx context.Context, token string) (request.Details, error)
}
