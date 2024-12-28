package oauth

import (
	"context"
	"net/http"
	"strings"

	"golang.org/x/oauth2"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/provider"
	"github.com/tidepool-org/platform/request"
)

type TokenSourceSource interface {
	TokenSource(ctx context.Context, token *auth.OAuthToken) (oauth2.TokenSource, error)
}

type Provider interface {
	provider.Provider
	TokenSourceSource

	CalculateStateForRestrictedToken(restrictedToken string) string // state = crypto of provider name, restrictedToken, secret
	GetAuthorizationCodeURLWithState(state string) string
	ExchangeAuthorizationCodeForToken(ctx context.Context, authorizationCode string) (*auth.OAuthToken, error)
}

type HTTPClientSource interface {
	HTTPClient(ctx context.Context, tokenSourceSource TokenSourceSource) (*http.Client, error)
}

type TokenSource interface {
	HTTPClientSource

	RefreshedToken() (*auth.OAuthToken, error)
	ExpireToken()
}

func IsAccessTokenError(err error) bool {
	return err != nil && request.IsErrorUnauthenticated(errors.Cause(err))
}

func IsRefreshTokenError(err error) bool {
	if err == nil {
		return false
	} else if err = errors.Cause(err); err == nil {
		return false
	} else if errString := err.Error(); !strings.Contains(errString, `oauth2: "invalid_grant"`) {
		return false
	} else {
		return true
	}
}

func GetIdToken(rawToken *oauth2.Token) string {
	idToken, _ := rawToken.Extra("id_token").(string)
	return idToken
}

const ErrorAccessDenied = "access_denied"
