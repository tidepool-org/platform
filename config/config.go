package config

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/tidepool-org/configor"
	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/environment"
)

type Loader interface {
	Load(name string, config interface{}) error
}

func NewLoader(environmentReporter environment.Reporter, directory string) (Loader, error) {
	if environmentReporter == nil {
		return nil, app.Error("config", "environment reporter is missing")
	}
	if directory == "" {
		return nil, app.Error("config", "directory is missing")
	}

	if fileInfo, err := os.Stat(directory); err != nil {
		if !os.IsNotExist(err) {
			return nil, app.ExtError(err, "config", "unable to stat directory")
		}
		return nil, app.Error("config", "directory does not exist")
	} else if !fileInfo.IsDir() {
		return nil, app.Error("config", "directory is not a directory")
	}

	return &loader{
		environmentReporter: environmentReporter,
		directory:           directory,
	}, nil
}

type loader struct {
	environmentReporter environment.Reporter
	directory           string
}

var (
	_mutex sync.Mutex
)

func (l *loader) Load(name string, config interface{}) error {
	if name == "" {
		return app.Error("config", "name is missing")
	}
	if config == nil {
		return app.Error("config", "config is missing")
	}

	paths := []string{}

	path := filepath.Join(l.directory, fmt.Sprintf("%s.json", name))
	if fileInfo, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return app.ExtError(err, "config", "unable to stat file")
		}
	} else if fileInfo.IsDir() {
		return app.Error("config", "file is a directory")
	} else {
		paths = append(paths, path)
	}

	_mutex.Lock()
	defer _mutex.Unlock()

	if err := os.Setenv("CONFIGOR_ENV", l.environmentReporter.Name()); err != nil {
		return app.ExtError(err, "config", "unable to set CONFIGOR_ENV")
	}

	if err := os.Setenv("CONFIGOR_ENV_PREFIX", l.environmentReporter.GetKey(strings.ToUpper(name))); err != nil {
		return app.ExtError(err, "config", "unable to set CONFIGOR_ENV_PREFIX")
	}

	if err := configor.Load(config, paths...); err != nil {
		return app.ExtError(err, "config", "unable to load config")
	}

	return nil
}
