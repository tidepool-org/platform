package config

import "github.com/tidepool-org/platform/errors"

const ErrorCodeKeyNotFound = "key-not-found"

func ErrorKeyNotFound(key string) error {
	return errors.Preparedf(ErrorCodeKeyNotFound, "key not found", "key %q not found", key)
}
