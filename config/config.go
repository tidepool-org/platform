package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jinzhu/configor"
)

func FromJSON(config interface{}, name string) error {

	// we need to get the absolute path and also test it because of the
	// difference of how this code runs in different environments
	absPath, err := filepath.Abs(filepath.Join("_config/", name))
	if err != nil {
		return err
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		absPath, err = filepath.Abs(filepath.Join("../_config/", name))
		if err != nil {
			return err
		}
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
