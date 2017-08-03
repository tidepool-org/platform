package service

import "github.com/tidepool-org/platform/application"

type Service struct {
	*application.Application
}

func New(name string, prefix string) (*Service, error) {
	app, err := application.New(name, prefix)
	if err != nil {
		return nil, err
	}

	return &Service{
		Application: app,
	}, nil
}
