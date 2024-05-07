package auth

import (
	"context"

	"github.com/tidepool-org/platform/devicetokens"
	"github.com/tidepool-org/platform/request"
)

const (
	TidepoolServiceSecretHeaderKey      = "X-Tidepool-Service-Secret"
	TidepoolAuthorizationHeaderKey      = "Authorization"
	TidepoolSessionTokenHeaderKey       = "X-Tidepool-Session-Token"
	TidepoolRestrictedTokenParameterKey = "restricted_token"
)

//go:generate mockgen --build_flags=--mod=mod -source=./auth.go -destination=./test/mock.go -package test -aux_files=github.com/tidepool-org/platform/auth=provider_session.go,github.com/tidepool-org/platform/auth=restricted_token.go Client
type Client interface {
	ProviderSessionAccessor
	RestrictedTokenAccessor
	ExternalAccessor
	DeviceTokensClient
}

type ExternalAccessor interface {
	ServerSessionToken() (string, error)
	ValidateSessionToken(ctx context.Context, token string) (request.AuthDetails, error)
	EnsureAuthorized(ctx context.Context) error
	EnsureAuthorizedService(ctx context.Context) error
	EnsureAuthorizedUser(ctx context.Context, targetUserID string, permission string) (string, error)
}

type contextKey string

const serverSessionTokenContextKey contextKey = "serverSessionToken"

func NewContextWithServerSessionToken(ctx context.Context, serverSessionToken string) context.Context {
	return context.WithValue(ctx, serverSessionTokenContextKey, serverSessionToken)
}

// ServerSessionTokenFromContext returns a JWT access token from a Context.
//
// An empty string is returned if no token is found.
func ServerSessionTokenFromContext(ctx context.Context) string {
	if ctx != nil {
		if serverSessionToken, ok := ctx.Value(serverSessionTokenContextKey).(string); ok {
			return serverSessionToken
		}
	}
	return ""
}

// DeviceTokensClient provides access to the tokens used to authenticate
// mobile device push notifications.
type DeviceTokensClient interface {
	GetDeviceTokens(ctx context.Context, userID string) ([]*devicetokens.DeviceToken, error)
}
