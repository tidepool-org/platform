package devicecheck

import (
	"errors"
	"net/http"

	devicecheck "github.com/snowman-mh/device-check-go"

	"github.com/tidepool-org/platform/apple"
	"github.com/tidepool-org/platform/config"
)

type deviceChecker struct {
	client *devicecheck.Client
}

type Config struct {
	PrivateKey                string
	Issuer                    string
	KeyID                     string
	UseDevelopmentEnvironment bool
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Load(configReporter config.Reporter) error {
	if configReporter == nil {
		return errors.New("config reporter is missing")
	}

	c.PrivateKey = configReporter.GetWithDefault("private_key", "")
	c.Issuer = configReporter.GetWithDefault("issuer", "")
	c.KeyID = configReporter.GetWithDefault("key_id", "")
	c.UseDevelopmentEnvironment = configReporter.GetWithDefault("use_development_environment", "false") == "true"

	return nil
}

func New(cfg *Config, httpClient *http.Client) apple.DeviceCheck {
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

func (d *deviceChecker) IsValidDeviceToken(token string) (bool, error) {
	err := d.client.ValidateDeviceToken(token)
	if err == nil {
		return true, nil
	} else if err == devicecheck.ErrBadDeviceToken {
		return false, nil
	} else {
		return false, err
	}
}
