package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/jinzhu/configor"
)

//FromJSON will load the requested JSON config into the given config interface
func FromJSON(config interface{}, name string) error {

	// we need to get the absolute path and also test it because of the
	// difference of how this code runs in different environments
	const configPath = "./_config/"
	absPath, _ := filepath.Abs(configPath + name)
	_, err := os.Open(absPath)

	if err != nil {
		const configLibPath = "./../_config/"
		absPath, _ = filepath.Abs(configLibPath + name)
	}

	return configor.Load(&config, absPath)
}

//FromEnv will load the requested config value
func FromEnv(name string) (string, error) {

	value := os.Getenv(name)

	if value == "" {
		return "", fmt.Errorf("$%s must be set", name)
	}
	return value, nil
}
