package factory

import (
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/provider"
)

type Factory struct {
	providers []provider.Provider
}

func New() (*Factory, error) {
	return &Factory{}, nil
}

func (f *Factory) Get(typ string, name string) provider.Provider {
	for _, prvdr := range f.providers {
		if typ == prvdr.Type() && name == prvdr.Name() {
			return prvdr
		}
	}
	return nil
}

func (f *Factory) Add(prvdr provider.Provider) error {
	if prvdr == nil {
		return errors.New("provider is missing")
	}
	if prvdr.Type() == "" {
		return errors.New("provider type is missing")
	}
	if prvdr.Name() == "" {
		return errors.New("provider name is missing")
	}

	f.providers = append(f.providers, prvdr)
	return nil
}
