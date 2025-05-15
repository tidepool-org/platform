package test

import (
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomToken() *auth.OAuthToken {
	datum := auth.NewOAuthToken()
	datum.AccessToken = test.RandomString()
	datum.TokenType = test.RandomString()
	datum.RefreshToken = test.RandomString()
	datum.ExpirationTime = test.RandomTime()
	datum.IDToken = pointer.FromString(test.RandomString())
	return datum
}

func CloneToken(datum *auth.OAuthToken) *auth.OAuthToken {
	if datum == nil {
		return nil
	}
	clone := auth.NewOAuthToken()
	clone.AccessToken = datum.AccessToken
	clone.TokenType = datum.TokenType
	clone.RefreshToken = datum.RefreshToken
	clone.ExpirationTime = datum.ExpirationTime
	clone.IDToken = pointer.CloneString(datum.IDToken)
	return clone
}

func NewObjectFromToken(datum *auth.OAuthToken, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	object["accessToken"] = test.NewObjectFromString(datum.AccessToken, objectFormat)
	object["tokenType"] = test.NewObjectFromString(datum.TokenType, objectFormat)
	object["refreshToken"] = test.NewObjectFromString(datum.RefreshToken, objectFormat)
	object["expirationTime"] = test.NewObjectFromTime(datum.ExpirationTime, objectFormat)
	if datum.IDToken != nil {
		object["idToken"] = test.NewObjectFromString(*datum.IDToken, objectFormat)
	}
	return object
}
