package provider

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v4"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"golang.org/x/oauth2"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/oauth"
)

type Provider struct {
	name         string
	config       Config
	jwks         jwk.Set
	oauth2Config *oauth2.Config
}

func New(name string, config *Config, jwks jwk.Set) (*Provider, error) {
	if name == "" {
		return nil, errors.New("name is missing")
	}
	if config == nil {
		return nil, errors.New("config is missing")
	} else if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}

	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.AuthorizeURL,
			TokenURL: config.TokenURL,
		},
		RedirectURL: config.RedirectURL,
		Scopes:      config.Scopes,
	}
	if config.AuthStyleInParams {
		oauth2Config.Endpoint.AuthStyle = oauth2.AuthStyleInParams
	}

	return &Provider{
		name:         name,
		config:       *config,
		jwks:         jwks,
		oauth2Config: oauth2Config,
	}, nil
}

func (p *Provider) Type() string {
	return oauth.ProviderType
}

func (p *Provider) Name() string {
	return p.name
}

func (p *Provider) ClientID() string {
	return p.config.ClientID
}

func (p *Provider) ClientSecret() string {
	return p.config.ClientSecret
}

func (p *Provider) OnCreate(ctx context.Context, providerSession *auth.ProviderSession) error {
	return nil
}

func (p *Provider) OnDelete(ctx context.Context, providerSession *auth.ProviderSession) error {
	return nil
}

func (p *Provider) SupportsUserInitiatedAccountUnlinking() bool {
	return true
}

func (p *Provider) ParseToken(token string, claims jwt.Claims) error {
	if token == "" {
		return errors.New("token is missing")
	}
	if claims == nil {
		return errors.New("claims are missing")
	}

	if p.jwks == nil {
		return errors.Newf("jwks is not defined for provider %s", p.name)
	}

	// Only verify the signed JWT, because the jwt package doesn't support validation with a JWK Set
	if _, err := jws.Verify([]byte(token), jws.WithKeySet(p.jwks, jws.WithInferAlgorithmFromKey(true))); err != nil {
		return errors.Wrap(err, "unable to verify id token with jwks")
	}

	// Parse the JWT with the jwt package for consistency with the rest of codebase
	_, _, err := jwt.NewParser().ParseUnverified(token, claims)
	return err
}

func (p *Provider) TokenSource(ctx context.Context, token *auth.OAuthToken) (oauth2.TokenSource, error) {
	if token == nil {
		return nil, errors.New("token is missing")
	}

	tknSrc := p.oauth2Config.TokenSource(ctx, token.RawToken())
	if tknSrc == nil {
		return nil, errors.New("unable to create token source")
	}

	return tknSrc, nil
}

func (p *Provider) CookieDisabled() bool {
	return p.config.CookieDisabled
}

func (p *Provider) CalculateStateForRestrictedToken(restrictedToken string) string {
	if !p.CookieDisabled() {
		return crypto.HexEncodedMD5Hash(fmt.Sprintf("%s:%s:%s:%s", p.Type(), p.Name(), restrictedToken, *p.config.StateSalt))
	} else {
		return restrictedToken
	}
}

func (p *Provider) GetAuthorizationCodeURLWithState(state string) string {
	return p.oauth2Config.AuthCodeURL(state)
}

func (p *Provider) ExchangeAuthorizationCodeForToken(ctx context.Context, authorizationCode string) (*auth.OAuthToken, error) {
	token, err := p.oauth2Config.Exchange(ctx, authorizationCode)
	if err != nil {
		return nil, errors.Wrap(err, "unable to exchange authorization code for token")
	}

	return auth.NewOAuthTokenFromRawToken(token)
}

func (p *Provider) IsErrorCodeAccessDenied(errorCode string) bool {
	return errorCode == "access_denied"
}
