package work

import (
	providerSession "github.com/tidepool-org/platform/auth/providersession"
	dataRaw "github.com/tidepool-org/platform/data/raw"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/oura"
	workBase "github.com/tidepool-org/platform/work/base"
)

type (
	ProviderSessionClient = providerSession.Client
	DataSourceClient      = dataSource.Client
	DataRawClient         = dataRaw.Client
	OuraClient            = oura.Client
)

type Dependencies struct {
	workBase.Dependencies
	ProviderSessionClient
	DataSourceClient
	DataRawClient
	OuraClient
}

func (d Dependencies) Validate() error {
	if err := d.Dependencies.Validate(); err != nil {
		return err
	}
	if d.ProviderSessionClient == nil {
		return errors.New("provider session client is missing")
	}
	if d.DataSourceClient == nil {
		return errors.New("data source client is missing")
	}
	if d.DataRawClient == nil {
		return errors.New("data raw client is missing")
	}
	if d.OuraClient == nil {
		return errors.New("oura client is missing")
	}
	return nil
}
