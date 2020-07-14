package devices

import (
	"context"
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/tidepool-org/devices/api"
	"go.uber.org/fx"
	"google.golang.org/grpc"

	"github.com/tidepool-org/platform/errors"
)

var ClientModule = fx.Provide(NewClient)

type Config struct {
	DevicesClientAddress string `envconfig:"TIDEPOOL_DEVICES_CLIENT_ADDRESS" required:"true"`
}

type Client struct {
	api.DevicesClient
	conn *grpc.ClientConn
}

func NewClient(lifecycle fx.Lifecycle) api.DevicesClient {
	client := &Client{}
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return client.Initialize(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return client.Stop()
		},
	})

	return client
}

func (c *Client) Initialize(ctx context.Context) error {
	cfg := &Config{}
	if err := envconfig.Process("", cfg); err != nil {
		return err
	}

	conn, err := grpc.DialContext(ctx, cfg.DevicesClientAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return errors.New(fmt.Sprintf("could not connect to devices service: %v", err))
	}

	c.DevicesClient = api.NewDevicesClient(conn)
	return nil
}

func (c *Client) Stop() error {
	return c.conn.Close()
}
