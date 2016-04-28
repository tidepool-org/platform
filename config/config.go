package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jinzhu/configor"

	"github.com/tidepool-org/platform/app"
)

const (
	EnvPrefixDefault = "TIDEPOOL"
)

var (
	_once      sync.Once
	_error     error
	_directory string
)

func Load(name string, config interface{}) error {
	_once.Do(func() {
		if env := os.Getenv("TIDEPOOL_ENV"); env == "" {
			_error = app.Error("config", "TIDEPOOL_ENV not defined")
		} else if err := os.Setenv("CONFIGOR_ENV", env); err != nil {
			_error = app.ExtError(err, "config", "unable to set CONFIGOR_ENV")
		} else if err := os.Setenv("CONFIGOR_ENV_PREFIX", EnvPrefixDefault); err != nil {
			_error = app.ExtError(err, "config", "unable to set CONFIGOR_ENV_PREFIX")
		} else if directory := os.Getenv("TIDEPOOL_CONFIG_DIRECTORY"); directory == "" {
			_error = app.Error("config", "TIDEPOOL_CONFIG_DIRECTORY not defined")
		} else if _directory, err = filepath.Abs(directory); err != nil {
			_error = app.ExtError(err, "config", "unable to determine absolute path to directory")
		}
	})

	if _error != nil {
		return _error
	}

	prefix := fmt.Sprintf("%s_%s", EnvPrefixDefault, strings.ToUpper(name))
	if err := os.Setenv("CONFIGOR_ENV_PREFIX", prefix); err != nil {
		return app.ExtError(err, "config", "unable to set CONFIGOR_ENV_PREFIX")
	}
	defer os.Setenv("CONFIGOR_ENV_PREFIX", EnvPrefixDefault)

	for _, extension := range []string{"json", "yaml"} {
		path := filepath.Join(_directory, fmt.Sprintf("%s.%s", name, extension))
		if _, err := os.Stat(path); err == nil {
			return configor.Load(config, path)
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	return configor.Load(config)
}
