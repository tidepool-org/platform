package provider

import (
	"context"

	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
)

func NewJWKS(configReporter config.Reporter) (jwk.Set, error) {
	var jwks jwk.Set
	jwksURL := configReporter.GetWithDefault("jwks_url", "")

	if jwksURL != "" {
		// Provider life-cycle is tied to the application life-cycle. Use a background context
		// to keep refreshing the cache until the application is terminated.
		jwkCache := jwk.NewCache(context.Background())

		err := jwkCache.Register(jwksURL)
		if err != nil {
			return nil, errors.New("unable to register jwks url")
		}

		jwks = jwk.NewCachedSet(jwkCache, jwksURL)
	}

	return jwks, nil
}
