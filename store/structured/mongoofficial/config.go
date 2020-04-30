package mongoofficial

import (
	"context"

	"go.uber.org/fx"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/store/structured/mongo"
)

func NewConfig(reporter config.Reporter, lifecycle fx.Lifecycle) (*mongo.Config, error) {
	cfg := mongo.NewConfig()

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return cfg.Load(reporter)
		},
	})

	return cfg, nil
}
