package auth

import (
	"context"

	"github.com/tidepool-org/platform/devicetokens"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
)

const (
	TidepoolServiceSecretHeaderKey      = "X-Tidepool-Service-Secret"
	TidepoolAuthorizationHeaderKey      = "Authorization"
	TidepoolSessionTokenHeaderKey       = "X-Tidepool-Session-Token"
	TidepoolRestrictedTokenParameterKey = "restricted_token"
)

//go:generate mockgen --build_flags=--mod=mod -source=./auth.go -destination=./test/mock.go -package test -aux_files=github.com/tidepool-org/platform/auth=provider_session.go,github.com/tidepool-org/platform/auth=restricted_token.go

type Client interface {
	ProviderSessionAccessor
	RestrictedTokenAccessor
	ExternalAccessor
	permission.Client
	DeviceTokensClient
}

type ExternalAccessor interface {
	ServerSessionToken() (string, error)
	ValidateSessionToken(ctx context.Context, token string) (request.AuthDetails, error)
	EnsureAuthorized(ctx context.Context) error
	EnsureAuthorizedService(ctx context.Context) error
	EnsureAuthorizedUser(ctx context.Context, targetUserID string, permission string) (string, error)
}

type ServerSessionTokenProvider interface {
	ServerSessionToken() (string, error)
}

func NewContextWithServerSessionTokenProvider(ctx context.Context, serverSessionTokenProvider ServerSessionTokenProvider) context.Context {
	return context.WithValue(ctx, serverSessionTokenProviderContextKey, serverSessionTokenProvider)
}

func ServerSessionTokenProviderFromContext(ctx context.Context) ServerSessionTokenProvider {
	if ctx != nil {
		if serverSessionTokenProvider, ok := ctx.Value(serverSessionTokenProviderContextKey).(ServerSessionTokenProvider); ok {
			return serverSessionTokenProvider
		}
	}
	return nil
}

type contextKey string

const serverSessionTokenProviderContextKey contextKey = "serverSessionTokenProvider"

// DeviceTokensClient provides access to the tokens used to authenticate
// mobile device push notifications.
type DeviceTokensClient interface {
	GetDeviceTokens(ctx context.Context, userID string) ([]*devicetokens.DeviceToken, error)
}
