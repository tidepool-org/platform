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

func (f *Factory) Get(typ string, name string) (provider.Provider, error) {
	if typ == "" {
		return nil, errors.New("type is missing")
	}
	if name == "" {
		return nil, errors.New("name is missing")
	}

	for _, prvdr := range f.providers {
		if typ == prvdr.Type() && name == prvdr.Name() {
			return prvdr, nil
		}
	}

	return nil, errors.Newf("provider with provider type %q and provider name %q not found", typ, name)
}

func (f *Factory) Add(prvdr provider.Provider) error {
	if prvdr == nil {
		return errors.New("provider is missing")
	}

	f.providers = append(f.providers, prvdr)
	return nil
}
