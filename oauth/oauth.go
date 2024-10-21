package oauth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/provider"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
)

type TokenSourceSource interface {
	TokenSource(ctx context.Context, token *Token) (oauth2.TokenSource, error)
}

type Provider interface {
	provider.Provider
	TokenSourceSource

	CalculateStateForRestrictedToken(restrictedToken string) string // state = crypto of provider name, restrictedToken, secret
	GetAuthorizationCodeURLWithState(state string) string
	ExchangeAuthorizationCodeForToken(ctx context.Context, authorizationCode string) (*Token, error)
}

type HTTPClientSource interface {
	HTTPClient(ctx context.Context, tokenSourceSource TokenSourceSource) (*http.Client, error)
}

type TokenSource interface {
	HTTPClientSource

	RefreshedToken() (*Token, error)
	ExpireToken()
}

type Token struct {
	AccessToken    string    `json:"accessToken" bson:"accessToken"`
	IdToken        string    `json:"idToken,omitempty" bson:"idToken,omitempty"`
	TokenType      string    `json:"tokenType,omitempty" bson:"tokenType,omitempty"`
	RefreshToken   string    `json:"refreshToken,omitempty" bson:"refreshToken,omitempty"`
	ExpirationTime time.Time `json:"expirationTime,omitempty" bson:"expirationTime,omitempty"`
}

func NewToken() *Token {
	return &Token{}
}

func NewTokenFromRawToken(rawToken *oauth2.Token) (*Token, error) {
	if rawToken == nil {
		return nil, errors.New("raw token is missing")
	}
	return &Token{
		AccessToken:    rawToken.AccessToken,
		IdToken:        GetIdToken(rawToken),
		TokenType:      rawToken.TokenType,
		RefreshToken:   rawToken.RefreshToken,
		ExpirationTime: rawToken.Expiry,
	}, nil
}

func (t *Token) Parse(parser structure.ObjectParser) {
	if accessToken := parser.String("accessToken"); accessToken != nil {
		t.AccessToken = *accessToken
	}
	if idToken := parser.String("idToken"); idToken != nil {
		t.IdToken = *idToken
	}
	if tokenType := parser.String("tokenType"); tokenType != nil {
		t.TokenType = *tokenType
	}
	if refreshToken := parser.String("refreshToken"); refreshToken != nil {
		t.RefreshToken = *refreshToken
	}
	if expirationTime := parser.Time("expirationTime", time.RFC3339Nano); expirationTime != nil {
		t.ExpirationTime = *expirationTime
	}
}

func (t *Token) Validate(validator structure.Validator) {
	validator.String("accessToken", &t.AccessToken).NotEmpty()
}

func (t *Token) Normalize(normalizer structure.Normalizer) {}

func (t *Token) Expire() {
	t.ExpirationTime = time.Now().Add(-time.Second)
}

func (t *Token) RawToken() *oauth2.Token {
	token := &oauth2.Token{
		AccessToken:  t.AccessToken,
		TokenType:    t.TokenType,
		RefreshToken: t.RefreshToken,
		Expiry:       t.ExpirationTime,
	}
	if t.IdToken != "" {
		token.WithExtra(map[string]any{
			"id_token": t.IdToken,
		})
	}
	return token
}

func (t *Token) MatchesRawToken(rawToken *oauth2.Token) bool {
	return rawToken != nil &&
		rawToken.AccessToken == t.AccessToken &&
		GetIdToken(rawToken) == t.IdToken &&
		rawToken.TokenType == t.TokenType &&
		rawToken.RefreshToken == t.RefreshToken &&
		rawToken.Expiry.Equal(t.ExpirationTime)
}

func IsAccessTokenError(err error) bool {
	return err != nil && request.IsErrorUnauthenticated(errors.Cause(err))
}

func IsRefreshTokenError(err error) bool {
	return err != nil && strings.Contains(errors.Cause(err).Error(), "oauth2: cannot fetch token: 400 Bad Request")
}

func GetIdToken(rawToken *oauth2.Token) string {
	idToken, _ := rawToken.Extra("id_token").(string)
	return idToken
}

const ErrorAccessDenied = "access_denied"
