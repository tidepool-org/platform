package auth

import (
	"time"

	"golang.org/x/oauth2"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
)

type OAuthToken struct {
	AccessToken    string    `json:"accessToken" bson:"accessToken"`
	TokenType      string    `json:"tokenType,omitempty" bson:"tokenType,omitempty"`
	RefreshToken   string    `json:"refreshToken,omitempty" bson:"refreshToken,omitempty"`
	ExpirationTime time.Time `json:"expirationTime,omitempty" bson:"expirationTime,omitempty"`
	IDToken        *string   `json:"idToken,omitempty" bson:"idToken,omitempty"`
}

func NewOAuthToken() *OAuthToken {
	return &OAuthToken{}
}

func NewOAuthTokenFromRawToken(rawToken *oauth2.Token) (*OAuthToken, error) {
	if rawToken == nil {
		return nil, errors.New("raw token is missing")
	}

	var idToken *string
	if extraIDToken, ok := rawToken.Extra("id_token").(string); ok && extraIDToken != "" {
		idToken = &extraIDToken
	}

	return &OAuthToken{
		AccessToken:    rawToken.AccessToken,
		TokenType:      rawToken.TokenType,
		RefreshToken:   rawToken.RefreshToken,
		ExpirationTime: rawToken.Expiry,
		IDToken:        idToken,
	}, nil
}

func (o *OAuthToken) Parse(parser structure.ObjectParser) {
	if accessToken := parser.String("accessToken"); accessToken != nil {
		o.AccessToken = *accessToken
	}
	if tokenType := parser.String("tokenType"); tokenType != nil {
		o.TokenType = *tokenType
	}
	if refreshToken := parser.String("refreshToken"); refreshToken != nil {
		o.RefreshToken = *refreshToken
	}
	if expirationTime := parser.Time("expirationTime", time.RFC3339Nano); expirationTime != nil {
		o.ExpirationTime = *expirationTime
	}
	o.IDToken = parser.String("idToken")
}

func (o *OAuthToken) Validate(validator structure.Validator) {
	validator.String("accessToken", &o.AccessToken).NotEmpty()
	validator.String("idToken", o.IDToken).NotEmpty()
}

func (o *OAuthToken) Expire() {
	o.ExpirationTime = time.Now().Add(-time.Second)
}

func (o *OAuthToken) RawToken() *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  o.AccessToken,
		TokenType:    o.TokenType,
		RefreshToken: o.RefreshToken,
		Expiry:       o.ExpirationTime,
	}
}

func (o *OAuthToken) MatchesRawToken(rawToken *oauth2.Token) bool {
	return rawToken != nil &&
		rawToken.AccessToken == o.AccessToken &&
		rawToken.TokenType == o.TokenType &&
		rawToken.RefreshToken == o.RefreshToken &&
		rawToken.Expiry.Equal(o.ExpirationTime)
}
