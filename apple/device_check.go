package apple

import (
	"errors"
	"net/http"

	"github.com/kelseyhightower/envconfig"
	devicecheck "github.com/rinchsan/device-check-go"
)

type DeviceCheck interface {
	IsTokenValid(string) (bool, error)
}

type deviceChecker struct {
	client *devicecheck.Client
}

type DeviceCheckConfig struct {
	PrivateKey                string `envconfig:"TIDEPOOL_APPLE_DEVICE_CHECKER_PRIVATE_KEY"`
	Issuer                    string `envconfig:"TIDEPOOL_APPLE_DEVICE_CHECKER_KEY_ISSUER"`
	KeyID                     string `envconfig:"TIDEPOOL_APPLE_DEVICE_CHECKER_KEY_ID"`
	UseDevelopmentEnvironment bool   `envconfig:"TIDEPOOL_APPLE_DEVICE_CHECKER_USE_DEVELOPMENT" default:"true"`
}

func NewDeviceCheckConfig() *DeviceCheckConfig {
	return &DeviceCheckConfig{}
}

func (c *DeviceCheckConfig) Load() error {
	return envconfig.Process("", c)
}

func NewDeviceCheck(cfg *DeviceCheckConfig, httpClient *http.Client) DeviceCheck {
	cred := devicecheck.NewCredentialString(cfg.PrivateKey)
	env := devicecheck.Production
	if cfg.UseDevelopmentEnvironment {
		env = devicecheck.Development
	}
	devicecheckCfg := devicecheck.NewConfig(cfg.Issuer, cfg.KeyID, env)
	client := devicecheck.NewWithHTTPClient(httpClient, cred, devicecheckCfg)

	return &deviceChecker{
		client: client,
	}
}

func (d *deviceChecker) IsTokenValid(token string) (bool, error) {
	err := d.client.ValidateDeviceToken(token)
	if err == nil {
		return true, nil
	} else if errors.Is(err, devicecheck.ErrUnauthorized) || errors.Is(err, devicecheck.ErrBadRequest) {
		return false, nil
	} else {
		return false, err
	}
}
