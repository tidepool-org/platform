package oauth

import (
	"context"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/provider"
	"github.com/tidepool-org/platform/structure"
)

type Provider interface {
	provider.Provider

	Config() *oauth2.Config

	State(restrictedToken string) string // state = crypto of provider name, restrictedToken, secret
}

type TokenSource interface {
	HTTPClient(ctx context.Context, prvdr Provider) (*http.Client, error)
	RefreshedToken() (*Token, error)
}

type Token struct {
	AccessToken    string    `json:"accessToken" bson:"accessToken"`
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
		TokenType:      rawToken.TokenType,
		RefreshToken:   rawToken.RefreshToken,
		ExpirationTime: rawToken.Expiry,
	}, nil
}

func (t *Token) Parse(parser structure.ObjectParser) {
	if accessToken := parser.String("accessToken"); accessToken != nil {
		t.AccessToken = *accessToken
	}
	if tokenType := parser.String("tokenType"); tokenType != nil {
		t.TokenType = *tokenType
	}
	if refreshToken := parser.String("refreshToken"); refreshToken != nil {
		t.RefreshToken = *refreshToken
	}
	if expirationTime := parser.Time("expirationTime", time.RFC3339); expirationTime != nil {
		t.ExpirationTime = *expirationTime
	}
}

func (t *Token) Validate(validator structure.Validator) {
	validator.String("accessToken", &t.AccessToken).NotEmpty()
}

func (t *Token) Normalize(normalizer structure.Normalizer) {
	t.ExpirationTime = t.ExpirationTime.UTC()
}

func (t *Token) RawToken() *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  t.AccessToken,
		TokenType:    t.TokenType,
		RefreshToken: t.RefreshToken,
		Expiry:       t.ExpirationTime,
	}
}

func (t *Token) MatchesRawToken(rawToken *oauth2.Token) bool {
	return rawToken != nil &&
		rawToken.AccessToken == t.AccessToken &&
		rawToken.TokenType == t.TokenType &&
		rawToken.RefreshToken == t.RefreshToken &&
		rawToken.Expiry == t.ExpirationTime
}

const ErrorAccessDenied = "access_denied"
