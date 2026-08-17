package test

import (
	authTest "github.com/tidepool-org/platform/auth/test"
	oauthWork "github.com/tidepool-org/platform/oauth/work"
	"github.com/tidepool-org/platform/test"
)

func RandomTokenMetadata(options ...test.Option) *oauthWork.TokenMetadata {
	return &oauthWork.TokenMetadata{
		OAuthToken: test.RandomOptionalPointer(authTest.RandomToken, options...),
	}
}

func CloneTokenMetadata(datum *oauthWork.TokenMetadata) *oauthWork.TokenMetadata {
	if datum == nil {
		return nil
	}
	return &oauthWork.TokenMetadata{
		OAuthToken: authTest.CloneToken(datum.OAuthToken),
	}
}

func NewObjectFromTokenMetadata(datum *oauthWork.TokenMetadata, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.OAuthToken != nil {
		object[oauthWork.MetadataKeyOAuthToken] = authTest.NewObjectFromToken(datum.OAuthToken, objectFormat)
	}
	return object
}
