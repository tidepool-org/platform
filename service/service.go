package service

import "github.com/tidepool-org/platform/application"

type Service struct {
	*application.Application
}

func New(prefix string) (*Service, error) {
	app, err := application.New(prefix)
	if err != nil {
		return nil, err
	}

	return &Service{
		Application: app,
	}, nil
}
