package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/jinzhu/configor"
)

func FromJson(config interface{}, name string) error {

	// we need to get the absolute path and also test it because of the
	// difference of how this code runs in different environments
	const config_path = "./_config/"
	absPath, _ := filepath.Abs(config_path + name)
	_, err := os.Open(absPath)

	if err != nil {
		const config_lib_path = "./../_config/"
		absPath, _ = filepath.Abs(config_lib_path + name)
	}

	return configor.Load(&config, absPath)
}

func FromEnv(name string) (string, error) {

	value := os.Getenv(name)

	if value == "" {
		return "", errors.New(fmt.Sprintf("$%s must be set", name))
	}
	return value, nil
}
