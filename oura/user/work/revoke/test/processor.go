package test

import (
	"maps"

	oauthWorkTest "github.com/tidepool-org/platform/oauth/work/test"
	ouraUserWorkRevoke "github.com/tidepool-org/platform/oura/user/work/revoke"
	"github.com/tidepool-org/platform/test"
)

func RandomMetadata(options ...test.Option) *ouraUserWorkRevoke.Metadata {
	return &ouraUserWorkRevoke.Metadata{
		TokenMetadata: *oauthWorkTest.RandomTokenMetadata(),
	}
}

func CloneMetadata(datum *ouraUserWorkRevoke.Metadata) *ouraUserWorkRevoke.Metadata {
	if datum == nil {
		return nil
	}
	return &ouraUserWorkRevoke.Metadata{
		TokenMetadata: *oauthWorkTest.CloneTokenMetadata(&datum.TokenMetadata),
	}
}

func NewObjectFromMetadata(datum *ouraUserWorkRevoke.Metadata, format test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	maps.Copy(object, oauthWorkTest.NewObjectFromTokenMetadata(&datum.TokenMetadata, format))
	return object
}
