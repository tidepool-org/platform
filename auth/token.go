package auth

import (
	"sort"
	"time"

	"golang.org/x/oauth2"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
)

const (
	KeyIDToken = "id_token"
	KeyScope   = "scope"
)

type OAuthToken struct {
	AccessToken    string    `json:"accessToken" bson:"accessToken"`
	TokenType      string    `json:"tokenType,omitempty" bson:"tokenType,omitempty"`
	RefreshToken   string    `json:"refreshToken,omitempty" bson:"refreshToken,omitempty"`
	ExpirationTime time.Time `json:"expirationTime,omitempty" bson:"expirationTime,omitempty"`
	Scope          *[]string `json:"scope,omitempty" bson:"scope,omitempty"`
	IDToken        *string   `json:"idToken,omitempty" bson:"idToken,omitempty"`
}

func ParseOAuthToken(parser structure.ObjectParser) *OAuthToken {
	if !parser.Exists() {
		return nil
	}
	datum := NewOAuthToken()
	parser.Parse(datum)
	return datum
}

func NewOAuthToken() *OAuthToken {
	return &OAuthToken{}
}

func NewOAuthTokenFromRawToken(rawToken *oauth2.Token) (*OAuthToken, error) {
	if rawToken == nil {
		return nil, errors.New("raw token is missing")
	}

	scope, err := GetScope(rawToken)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get scope from raw token")
	}

	return &OAuthToken{
		AccessToken:    rawToken.AccessToken,
		TokenType:      rawToken.TokenType,
		RefreshToken:   rawToken.RefreshToken,
		ExpirationTime: rawToken.Expiry,
		Scope:          scope,
		IDToken:        GetIDToken(rawToken),
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
	o.Scope = parser.StringArray("scope")
	o.IDToken = parser.String("idToken")
}

func (o *OAuthToken) Validate(validator structure.Validator) {
	validator.String("accessToken", &o.AccessToken).NotEmpty()
	validator.StringArray("scope", o.Scope).EachUsing(ScopeTokenValidator).EachUnique()
	validator.String("idToken", o.IDToken).NotEmpty()
}

func (o *OAuthToken) Normalize(normalizer structure.Normalizer) {
	if o.Scope != nil {
		sort.Strings(*o.Scope)
	}
}

func (o *OAuthToken) Refreshed(rawToken *oauth2.Token) (*OAuthToken, error) {
	if rawToken == nil {
		return nil, errors.New("raw token is missing")
	}

	refreshed := *o
	refreshed.AccessToken = rawToken.AccessToken
	refreshed.TokenType = rawToken.TokenType
	refreshed.RefreshToken = rawToken.RefreshToken
	refreshed.ExpirationTime = rawToken.Expiry

	// Only replace if one provided
	if scope, err := GetScope(rawToken); err != nil {
		return nil, err
	} else if scope != nil {
		refreshed.Scope = scope
	}
	if idToken := GetIDToken(rawToken); idToken != nil {
		refreshed.IDToken = idToken
	}

	return &refreshed, nil
}

func (o *OAuthToken) Expired() *OAuthToken {
	return &OAuthToken{
		AccessToken:    o.AccessToken,
		TokenType:      o.TokenType,
		RefreshToken:   o.RefreshToken,
		ExpirationTime: time.Now().Add(-time.Second),
		Scope:          o.Scope,
		IDToken:        o.IDToken,
	}
}

func (o *OAuthToken) RawToken() *oauth2.Token {
	rawToken := &oauth2.Token{
		AccessToken:  o.AccessToken,
		TokenType:    o.TokenType,
		RefreshToken: o.RefreshToken,
		Expiry:       o.ExpirationTime,
	}

	extra := map[string]any{}
	if o.Scope != nil {
		extra[KeyScope] = JoinScope(*o.Scope)
	}
	if o.IDToken != nil {
		extra[KeyIDToken] = *o.IDToken
	}
	if len(extra) > 0 {
		rawToken = rawToken.WithExtra(extra)
	}

	return rawToken
}

func (o *OAuthToken) MatchesRawToken(rawToken *oauth2.Token) bool {
	scope, err := GetScope(rawToken)
	if err != nil {
		return false
	}
	return rawToken != nil &&
		rawToken.AccessToken == o.AccessToken &&
		rawToken.TokenType == o.TokenType &&
		rawToken.RefreshToken == o.RefreshToken &&
		rawToken.Expiry.Equal(o.ExpirationTime) &&
		pointer.EqualStringArray(scope, o.Scope) &&
		pointer.EqualString(GetIDToken(rawToken), o.IDToken)
}

func GetIDToken(rawToken *oauth2.Token) *string {
	if idToken, ok := rawToken.Extra(KeyIDToken).(string); ok && idToken != "" {
		return &idToken
	}
	return nil
}

func GetScope(rawToken *oauth2.Token) (*[]string, error) {
	if rawScope, ok := rawToken.Extra(KeyScope).(string); !ok {
		return nil, nil
	} else if scope, err := ParseScope(rawScope); err != nil {
		return nil, err
	} else if scope == nil {
		return nil, nil
	} else {
		return pointer.FromStringArray(scope), nil
	}
}
