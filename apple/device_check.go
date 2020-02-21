package apple

import (
	devicecheck "github.com/snowman-mh/device-check-go"
	"github.com/tidepool-org/platform/config"
	"net/http"
)

type DeviceChecker interface {
	IsValidDeviceToken(string) (bool, error)
}

type deviceCheckerImpl struct {
	client *devicecheck.Client
}

type DeviceCheckerConfig struct {
	PrivateKey                string
	Issuer                    string
	KeyID                     string
	UseDevelopmentEnvironment bool
}

func NewDeviceCheckerConfig() *DeviceCheckerConfig {
	return &DeviceCheckerConfig{}
}

func (c *DeviceCheckerConfig) Load(configReporter config.Reporter) error {
	if err := c.Load(configReporter); err != nil {
		return err
	}

	c.PrivateKey = configReporter.GetWithDefault("private_key", "")
	c.Issuer = configReporter.GetWithDefault("issuer", "")
	c.KeyID = configReporter.GetWithDefault("key_id", "")
	c.UseDevelopmentEnvironment = configReporter.GetWithDefault("use_development_environment", "false") == "true"

	return nil
}

func NewDeviceChecker(cfg *DeviceCheckerConfig, httpClient *http.Client) DeviceChecker {
	cred := devicecheck.NewCredentialString(cfg.PrivateKey)
	env := devicecheck.Production
	if cfg.UseDevelopmentEnvironment {
		env = devicecheck.Development
	}
	devicecheckCfg := devicecheck.NewConfig(cfg.Issuer, cfg.KeyID, env)
	client := devicecheck.NewWithHTTPClient(httpClient, cred, devicecheckCfg)

	return &deviceCheckerImpl{
		client: client,
	}
}

func (d *deviceCheckerImpl) IsValidDeviceToken(token string) (bool, error) {
	err := d.client.ValidateDeviceToken(token)
	if err == nil {
		return true, nil
	} else if err == devicecheck.ErrBadDeviceToken {
		return false, nil
	} else {
		return false, err
	}
}
