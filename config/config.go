package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/tidepool-org/configor"
	"github.com/tidepool-org/platform/environment"
	"github.com/tidepool-org/platform/errors"
)

type Loader interface {
	Load(name string, config interface{}) error
}

func NewLoader(environmentReporter environment.Reporter, directory string) (Loader, error) {
	if environmentReporter == nil {
		return nil, errors.New("config", "environment reporter is missing")
	}
	if directory == "" {
		return nil, errors.New("config", "directory is missing")
	}

	if fileInfo, err := os.Stat(directory); err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrap(err, "config", "unable to stat directory")
		}
		return nil, errors.New("config", "directory does not exist")
	} else if !fileInfo.IsDir() {
		return nil, errors.New("config", "directory is not a directory")
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
		return errors.New("config", "name is missing")
	}
	if config == nil {
		return errors.New("config", "config is missing")
	}

	paths := []string{}

	path := filepath.Join(l.directory, fmt.Sprintf("%s.json", name))
	if fileInfo, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrap(err, "config", "unable to stat file")
		}
	} else if fileInfo.IsDir() {
		return errors.New("config", "file is a directory")
	} else {
		paths = append(paths, path)
	}

	_mutex.Lock()
	defer _mutex.Unlock()

	if err := os.Setenv("CONFIGOR_ENV", l.environmentReporter.Name()); err != nil {
		return errors.Wrap(err, "config", "unable to set CONFIGOR_ENV")
	}

	if err := os.Setenv("CONFIGOR_ENV_PREFIX", l.environmentReporter.GetKey(strings.ToUpper(name))); err != nil {
		return errors.Wrap(err, "config", "unable to set CONFIGOR_ENV_PREFIX")
	}

	if err := configor.Load(config, paths...); err != nil {
		return errors.Wrap(err, "config", "unable to load config")
	}

	return nil
}
