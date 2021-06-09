package apple

import (
	"net/http"

	"github.com/kelseyhightower/envconfig"
	devicecheck "github.com/rinchsan/device-check-go"
)

type DeviceCheck interface {
	IsTokenValid(string) (bool, error)
}

type deviceChecker struct {
	enabled bool
	client  *devicecheck.Client
}

type Config struct {
	PrivateKey                string `envconfig:"DEVICE_CHECK_PRIVATE_KEY"`
	Issuer                    string `envconfig:"DEVICE_CHECK_KEY_ISSUER"`
	KeyID                     string `envconfig:"DEVICE_CHECK_KEY_ID"`
	UseDevelopmentEnvironment bool   `envconfig:"DEVICE_CHECK_USE_DEVELOPMENT" default:"true"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Load() error {
	return envconfig.Process("", c)
}

func New(cfg *Config, httpClient *http.Client) DeviceCheck {
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
	} else if err == devicecheck.ErrUnauthorized {
		return false, nil
	} else {
		return false, err
	}
}
