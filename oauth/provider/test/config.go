package test

import (
	authTest "github.com/tidepool-org/platform/auth/test"
	oauthProvider "github.com/tidepool-org/platform/oauth/provider"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

func RandomAcceptURL() string {
	return testHttp.NewURL().String()
}

func RandomAuthorizeURL() string {
	return testHttp.NewURL().String()
}

func RandomRedirectURL() string {
	return testHttp.NewURL().String()
}

func RandomTokenURL() string {
	return testHttp.NewURL().String()
}

func RandomRevokeURL() string {
	return testHttp.NewURL().String()
}

func RandomStateSalt() string {
	return test.RandomStringFromCharset(test.CharsetAlphaNumeric)
}

func RandomConfig(options ...test.Option) *oauthProvider.Config {
	cookiesDisabled := test.RandomBool()
	return &oauthProvider.Config{
		ClientID:          RandomClientID(),
		ClientSecret:      RandomClientSecret(),
		AcceptURL:         test.RandomOptional(RandomAcceptURL, options...),
		AuthorizeURL:      RandomAuthorizeURL(),
		RedirectURL:       RandomRedirectURL(),
		TokenURL:          RandomTokenURL(),
		RevokeURL:         test.RandomOptional(RandomRevokeURL, options...),
		Scopes:            authTest.RandomScope(),
		AuthStyleInParams: test.RandomBool(),
		CookieDisabled:    cookiesDisabled,
		StateSalt:         test.Conditional(RandomStateSalt, !cookiesDisabled),
	}
}

func CloneConfig(config *oauthProvider.Config) *oauthProvider.Config {
	if config == nil {
		return nil
	}
	return &oauthProvider.Config{
		ClientID:          config.ClientID,
		ClientSecret:      config.ClientSecret,
		AcceptURL:         pointer.Clone(config.AcceptURL),
		AuthorizeURL:      config.AuthorizeURL,
		RedirectURL:       config.RedirectURL,
		TokenURL:          config.TokenURL,
		RevokeURL:         pointer.Clone(config.RevokeURL),
		Scopes:            authTest.CloneScope(config.Scopes),
		AuthStyleInParams: config.AuthStyleInParams,
		CookieDisabled:    config.CookieDisabled,
		StateSalt:         pointer.Clone(config.StateSalt),
	}
}

func NewObjectFromConfig(config *oauthProvider.Config, objectFormat test.ObjectFormat) map[string]any {
	if config == nil {
		return nil
	}
	object := map[string]any{}
	object["client_id"] = test.NewObjectFromString(config.ClientID, objectFormat)
	object["client_secret"] = test.NewObjectFromString(config.ClientSecret, objectFormat)
	if config.AcceptURL != nil {
		object["accept_url"] = test.NewObjectFromString(*config.AcceptURL, objectFormat)
	}
	object["authorize_url"] = test.NewObjectFromString(config.AuthorizeURL, objectFormat)
	object["redirect_url"] = test.NewObjectFromString(config.RedirectURL, objectFormat)
	object["token_url"] = test.NewObjectFromString(config.TokenURL, objectFormat)
	if config.RevokeURL != nil {
		object["revoke_url"] = test.NewObjectFromString(*config.RevokeURL, objectFormat)
	}
	if config.Scopes != nil {
		object["scopes"] = authTest.NewObjectFromScope(config.Scopes, objectFormat)
	}
	if config.AuthStyleInParams {
		object["auth_style_in_params"] = test.NewObjectFromBool(config.AuthStyleInParams, objectFormat)
	}
	if config.CookieDisabled {
		object["cookie_disabled"] = test.NewObjectFromBool(config.CookieDisabled, objectFormat)
	}
	if config.StateSalt != nil {
		object["state_salt"] = test.NewObjectFromString(*config.StateSalt, objectFormat)
	}
	return object
}
