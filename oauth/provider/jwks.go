package provider

import (
	"context"

	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
)

func NewJWKS(configReporter config.Reporter) (jwk.Set, error) {
	jwksURL := configReporter.GetWithDefault("jwks_url", "")
	if jwksURL == "" {
		return nil, nil
	}

	// Provider life-cycle is tied to the application life-cycle. Use a background context
	// to keep refreshing the cache until the application is terminated.
	jwkCache := jwk.NewCache(context.Background())

	if err := jwkCache.Register(jwksURL); err != nil {
		return nil, errors.New("unable to register jwks url")
	}

	return jwk.NewCachedSet(jwkCache, jwksURL), nil
}
