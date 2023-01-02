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

//go:generate mockgen -build_flags=--mod=mod -destination=./auth_client_mock.go -package=auth -mock_names Client=MockAuthClient github.com/tidepool-org/platform/auth Client

type Client interface {
	ProviderSessionAccessor
	RestrictedTokenAccessor
	ExternalAccessor
}

type ExternalAccessor interface {
	ServerSessionToken() (string, error)
	ValidateSessionToken(ctx context.Context, token string) (request.Details, error)
	EnsureAuthorized(ctx context.Context) error
	EnsureAuthorizedService(ctx context.Context) error
	EnsureAuthorizedUser(ctx context.Context, targetUserID string, permission string) (string, error)
}

type contextKey string

const serverSessionTokenContextKey contextKey = "serverSessionToken"

func NewContextWithServerSessionToken(ctx context.Context, serverSessionToken string) context.Context {
	return context.WithValue(ctx, serverSessionTokenContextKey, serverSessionToken)
}

func ServerSessionTokenFromContext(ctx context.Context) string {
	if ctx != nil {
		if serverSessionToken, ok := ctx.Value(serverSessionTokenContextKey).(string); ok {
			return serverSessionToken
		}
	}
	return ""
}
