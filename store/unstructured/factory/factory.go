package factory

import (
	"github.com/tidepool-org/platform/aws"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
	storeUnstructured "github.com/tidepool-org/platform/store/unstructured"
	storeUnstructuredFile "github.com/tidepool-org/platform/store/unstructured/file"
	storeUnstructuredS3 "github.com/tidepool-org/platform/store/unstructured/s3"
)

func NewStore(configReporter config.Reporter, awsAPI aws.API) (storeUnstructured.Store, error) {
	if configReporter == nil {
		return nil, errors.New("config reporter is missing")
	}
	if awsAPI == nil {
		return nil, errors.New("aws api is missing")
	}

	typ, err := configReporter.Get("type")
	if err != nil {
		return nil, errors.New("type is missing")
	}

	switch typ {
	case storeUnstructuredFile.Type:
		return NewFileStore(configReporter.WithScopes(storeUnstructuredFile.Type))
	case storeUnstructuredS3.Type:
		return NewS3Store(configReporter.WithScopes(storeUnstructuredS3.Type), awsAPI)
	}
	return nil, errors.New("type is invalid")
}

func NewFileStore(configReporter config.Reporter) (storeUnstructured.Store, error) {
	cfg := storeUnstructuredFile.NewConfig()
	if err := cfg.Load(configReporter); err != nil {
		return nil, errors.Wrap(err, "unable to load config")
	}
	return storeUnstructuredFile.NewStore(cfg)
}

func NewS3Store(configReporter config.Reporter, awsAPI aws.API) (storeUnstructured.Store, error) {
	cfg := storeUnstructuredS3.NewConfig()
	if err := cfg.Load(configReporter); err != nil {
		return nil, errors.Wrap(err, "unable to load config")
	}
	return storeUnstructuredS3.NewStore(cfg, awsAPI)
}
