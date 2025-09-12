package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/golang-jwt/jwt/v4"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"golang.org/x/oauth2"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/errors"
)

const ProviderType = "oauth"

type Provider struct {
	name      string
	config    *oauth2.Config
	stateSalt string
	jwks      jwk.Set
}

func New(name string, configReporter config.Reporter, jwks jwk.Set) (*Provider, error) {
	if name == "" {
		return nil, errors.New("name is missing")
	}
	if configReporter == nil {
		return nil, errors.New("config reporter is missing")
	}

	cfg := &oauth2.Config{}
	cfg.ClientID = configReporter.GetWithDefault("client_id", "")
	if cfg.ClientID == "" {
		return nil, errors.New("client id is missing")
	}
	cfg.ClientSecret = configReporter.GetWithDefault("client_secret", "")
	if cfg.ClientSecret == "" {
		return nil, errors.New("client secret is missing")
	}
	cfg.Endpoint.AuthURL = configReporter.GetWithDefault("authorize_url", "")
	if cfg.Endpoint.AuthURL == "" {
		return nil, errors.New("authorize url is missing")
	}
	cfg.Endpoint.TokenURL = configReporter.GetWithDefault("token_url", "")
	if cfg.Endpoint.TokenURL == "" {
		return nil, errors.New("token url is missing")
	}
	cfg.RedirectURL = configReporter.GetWithDefault("redirect_url", "")
	if cfg.RedirectURL == "" {
		return nil, errors.New("redirect url is missing")
	}
	cfg.Scopes = SplitScopes(configReporter.GetWithDefault("scopes", ""))

	if authStyleInParams, err := strconv.ParseBool(configReporter.GetWithDefault("auth_style_in_params", "false")); err != nil {
		return nil, errors.New("auth style in params is invalid")
	} else if authStyleInParams {
		cfg.Endpoint.AuthStyle = oauth2.AuthStyleInParams
	}

	var stateSalt string
	if useCookie, err := strconv.ParseBool(configReporter.GetWithDefault("use_cookie", "true")); err != nil {
		return nil, errors.New("use cookie is invalid")
	} else if useCookie {
		if stateSalt = configReporter.GetWithDefault("state_salt", ""); stateSalt == "" {
			return nil, errors.New("state salt is missing")
		}
	}

	return &Provider{
		name:      name,
		config:    cfg,
		stateSalt: stateSalt,
		jwks:      jwks,
	}, nil
}

func (p *Provider) Type() string {
	return ProviderType
}

func (p *Provider) Name() string {
	return p.name
}

func (p *Provider) ClientID() string {
	return p.config.ClientID
}

func (p *Provider) OnCreate(ctx context.Context, providerSession *auth.ProviderSession) error {
	return nil
}

func (p *Provider) OnDelete(ctx context.Context, providerSession *auth.ProviderSession) error {
	return nil
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

	tknSrc := p.config.TokenSource(ctx, token.RawToken())
	if tknSrc == nil {
		return nil, errors.New("unable to create token source")
	}

	return tknSrc, nil
}

func (p *Provider) UseCookie() bool {
	return p.stateSalt != ""
}

func (p *Provider) CalculateStateForRestrictedToken(restrictedToken string) string {
	if p.stateSalt != "" {
		return crypto.HexEncodedMD5Hash(fmt.Sprintf("%s:%s:%s:%s", p.Type(), p.Name(), restrictedToken, p.stateSalt))
	} else {
		return restrictedToken
	}
}

func (p *Provider) GetAuthorizationCodeURLWithState(state string) string {
	return p.config.AuthCodeURL(state)
}

func (p *Provider) ExchangeAuthorizationCodeForToken(ctx context.Context, authorizationCode string) (*auth.OAuthToken, error) {
	token, err := p.config.Exchange(ctx, authorizationCode)
	if err != nil {
		return nil, errors.Wrap(err, "unable to exchange authorization code for token")
	}

	return auth.NewOAuthTokenFromRawToken(token)
}

func (p *Provider) IsErrorCodeAccessDenied(errorCode string) bool {
	return errorCode == "access_denied"
}

func SplitScopes(scopes string) []string {
	return config.SplitTrimCompact(scopes)
}
