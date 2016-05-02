package config

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jinzhu/configor"

	"github.com/tidepool-org/platform/app"
)

func Load(name string, config interface{}) error {
	if name == "" {
		return app.Error("config", "name is not specified")
	}
	if config == nil {
		return app.Error("config", "config is not specified")
	}

	if _error != nil {
		return _error
	}

	for _, extension := range []string{"json", "yaml"} {
		path := filepath.Join(_directory, fmt.Sprintf("%s.%s", name, extension))
		if _, err := os.Stat(path); err == nil {
			return loadWithPrefix(name, config, path)
		} else if !os.IsNotExist(err) {
			return app.ExtError(err, "config", "unable to find config file")
		}
	}

	return loadWithPrefix(name, config)
}

func loadWithPrefix(name string, config interface{}, args ...string) error {
	oldPrefix := os.Getenv("CONFIGOR_ENV_PREFIX")
	newPrefix := fmt.Sprintf("%s_%s", oldPrefix, strings.ToUpper(name))

	// TODO: This is NOT concurrent safe!

	if err := os.Setenv("CONFIGOR_ENV_PREFIX", newPrefix); err != nil {
		return app.ExtError(err, "config", "unable to set new CONFIGOR_ENV_PREFIX")
	}

	err := configor.Load(config, args...)

	if err := os.Setenv("CONFIGOR_ENV_PREFIX", oldPrefix); err != nil {
		return app.ExtError(err, "config", "unable to set old CONFIGOR_ENV_PREFIX")
	}

	if err != nil {
		return app.ExtError(err, "config", "unable to load config")
	}

	return nil
}
