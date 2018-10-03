package file

import (
	"context"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	storeUnstructured "github.com/tidepool-org/platform/store/unstructured"
)

const Type = "file"

type Store struct {
	directory string
}

func NewStore(cfg *Config) (*Store, error) {
	if cfg == nil {
		return nil, errors.New("config is missing")
	} else if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}

	return &Store{
		directory: cfg.Directory,
	}, nil
}

func (s *Store) Exists(ctx context.Context, key string) (bool, error) {
	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if key == "" {
		return false, errors.New("key is missing")
	} else if !storeUnstructured.IsValidKey(key) {
		return false, errors.New("key is invalid")
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"directory": s.directory, "key": key})
	filePath := s.resolveKey(key)

	var exists bool
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			logger.WithError(err).Errorf("Unable to stat file at path %q", filePath)
			return false, errors.Wrapf(err, "unable to stat file at path %q", filePath)
		}
	} else if !fileInfo.Mode().IsRegular() {
		logger.Errorf("Unexpected directory or irregular file at path %q", filePath)
		return false, errors.Newf("unexpected directory or irregular file at path %q", filePath)
	} else {
		exists = true
	}

	logger.WithField("exists", exists).Debug("Exists")
	return exists, nil
}

func (s *Store) Put(ctx context.Context, key string, reader io.Reader) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if key == "" {
		return errors.New("key is missing")
	} else if !storeUnstructured.IsValidKey(key) {
		return errors.New("key is invalid")
	}
	if reader == nil {
		return errors.New("reader is missing")
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"directory": s.directory, "key": key})
	filePath := s.resolveKey(key)
	directoryPath := filepath.Dir(filePath)

	err := os.MkdirAll(directoryPath, 0777)
	if err != nil {
		logger.WithError(err).Errorf("Unable to create directories at path %q", directoryPath)
		return errors.Wrapf(err, "unable to create directories at path %q", directoryPath)
	}

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		logger.WithError(err).Errorf("Unable to create file at path %q", filePath)
		return errors.Wrapf(err, "unable to create file at path %q", filePath)
	}

	_, err = io.Copy(file, reader)
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		logger.WithError(err).Errorf("Unable to write file at path %q", filePath)
		return errors.Wrapf(err, "unable to write file at path %q", filePath)
	}

	logger.Debug("Put")
	return nil
}

func (s *Store) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if key == "" {
		return nil, errors.New("key is missing")
	} else if !storeUnstructured.IsValidKey(key) {
		return nil, errors.New("key is invalid")
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"directory": s.directory, "key": key})
	filePath := s.resolveKey(key)

	var reader io.ReadCloser
	if file, openErr := os.Open(filePath); openErr != nil {
		if !os.IsNotExist(openErr) {
			logger.WithError(openErr).Errorf("Unable to open file at path %q", filePath)
			return nil, errors.Wrapf(openErr, "unable to open file at path %q", filePath)
		}
	} else if fileInfo, statErr := file.Stat(); statErr != nil {
		file.Close()
		logger.WithError(statErr).Errorf("Unable to stat file at path %q", filePath)
		return nil, errors.Wrapf(statErr, "unable to stat file at path %q", filePath)
	} else if !fileInfo.Mode().IsRegular() {
		file.Close()
		logger.Errorf("Unexpected directory or irregular file at path %q", filePath)
		return nil, errors.Newf("unexpected directory or irregular file at path %q", filePath)
	} else {
		reader = file
	}

	logger.WithField("exists", reader != nil).Debug("Get")
	return reader, nil
}

func (s *Store) Delete(ctx context.Context, key string) (bool, error) {
	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if key == "" {
		return false, errors.New("key is missing")
	} else if !storeUnstructured.IsValidKey(key) {
		return false, errors.New("key is invalid")
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"directory": s.directory, "key": key})
	filePath := s.resolveKey(key)

	var exists bool
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			logger.WithError(err).Errorf("Unable to stat file at path %q", filePath)
			return false, errors.Wrapf(err, "unable to stat file at path %q", filePath)
		}
	} else if !fileInfo.Mode().IsRegular() {
		logger.Errorf("Unexpected directory or irregular file at path %q", filePath)
		return false, errors.Newf("unexpected directory or irregular file at path %q", filePath)
	} else if removeErr := os.Remove(filePath); removeErr != nil {
		if !os.IsNotExist(removeErr) {
			logger.WithError(removeErr).Errorf("Unable to remove file at path %q", filePath)
			return false, errors.Wrapf(removeErr, "unable to remove file at path %q", filePath)
		}
	} else {
		exists = true
		for key = path.Dir(key); key != "."; key = path.Dir(key) {
			if err = os.Remove(s.resolveKey(key)); err != nil {
				break
			}
		}
	}

	logger.WithField("exists", exists).Debug("Delete")
	return exists, nil
}

func (s *Store) resolveKey(key string) string {
	return filepath.Join(s.directory, filepath.FromSlash(key))
}
