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
	"os"
	"path/filepath"

	"github.com/tidepool-org/platform/app"
)

var (
	_error     error
	_directory string
)

func init() {
	if env := os.Getenv("TIDEPOOL_ENV"); env == "" {
		_error = app.Error("config", "TIDEPOOL_ENV is not defined")
	} else if err := os.Setenv("CONFIGOR_ENV", env); err != nil {
		_error = app.ExtError(err, "config", "unable to set CONFIGOR_ENV")
	} else if err = os.Setenv("CONFIGOR_ENV_PREFIX", "TIDEPOOL"); err != nil {
		_error = app.ExtError(err, "config", "unable to set CONFIGOR_ENV_PREFIX")
	} else if directory := os.Getenv("TIDEPOOL_CONFIG_DIRECTORY"); directory == "" {
		_error = app.Error("config", "TIDEPOOL_CONFIG_DIRECTORY is not defined")
	} else if _directory, err = filepath.Abs(directory); err != nil {
		_error = app.ExtError(err, "config", "unable to determine absolute path to directory")
	}
}
