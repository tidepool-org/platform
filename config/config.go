package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/jinzhu/configor"
)

func FromJSON(config interface{}, name string) error {

	// we need to get the absolute path and also test it because of the
	// difference of how this code runs in different environments
	absPath, _ := filepath.Abs(filepath.Join("./_config/", name))

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		absPath, _ = filepath.Abs(filepath.Join("./../_config/", name))
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return err
		}
	}

	return configor.Load(&config, absPath)
}

func FromEnv(name string) (string, error) {

	value := os.Getenv(name)

	if value == "" {
		return "", fmt.Errorf("$%s must be set", name)
	}
	return value, nil
}
