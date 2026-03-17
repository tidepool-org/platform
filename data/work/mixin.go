package work

import (
	"context"

	providerSession "github.com/tidepool-org/platform/auth/providersession"
	providerSessionWork "github.com/tidepool-org/platform/auth/providersession/work"
	dataSetWork "github.com/tidepool-org/platform/data/set/work"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceWork "github.com/tidepool-org/platform/data/source/work"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
)

type ProviderSessionDataSourceMixin interface {
	FetchDataSourceFromProviderSession() *work.ProcessResult
	FetchProviderSessionFromDataSource() *work.ProcessResult
}

func NewProviderSessionDataSourceMixin(provider work.Provider, providerSessionMixin providerSessionWork.Mixin, dataSourceMixin dataSourceWork.Mixin) (ProviderSessionDataSourceMixin, error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}
	if providerSessionMixin == nil {
		return nil, errors.New("provider session mixin is missing")
	}
	if dataSourceMixin == nil {
		return nil, errors.New("data source mixin is missing")
	}
	return &providerSessionDataSourceMixin{
		Provider:             provider,
		providerSessionMixin: providerSessionMixin,
		dataSourceMixin:      dataSourceMixin,
	}, nil
}

type (
	providerSessionMixin = providerSessionWork.Mixin
	dataSourceMixin      = dataSourceWork.Mixin
)

type providerSessionDataSourceMixin struct {
	work.Provider
	providerSessionMixin
	dataSourceMixin
}

func (p *providerSessionDataSourceMixin) FetchDataSourceFromProviderSession() *work.ProcessResult {
	if !p.HasProviderSession() {
		return p.Failed(errors.New("provider session is missing"))
	} else {
		return p.FetchDataSourceFromProviderSessionID(p.ProviderSession().ID)
	}
}

func (p *providerSessionDataSourceMixin) FetchProviderSessionFromDataSource() *work.ProcessResult {
	if !p.HasDataSource() {
		return p.Failed(errors.New("data source is missing"))
	} else if providerSessionID := p.DataSource().ProviderSessionID; providerSessionID == nil {
		return p.Failed(errors.New("data source provider session id is missing"))
	} else {
		return p.FetchProviderSession(*providerSessionID)
	}
}

type DataSourceDataSetMixin interface {
	FetchDataSetFromDataSource() *work.ProcessResult
}

func NewDataSourceDataSetMixin(provider work.Provider, dataSourceMixin dataSourceWork.Mixin, dataSetMixin dataSetWork.Mixin) (DataSourceDataSetMixin, error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}
	if dataSourceMixin == nil {
		return nil, errors.New("data source mixin is missing")
	}
	if dataSetMixin == nil {
		return nil, errors.New("data set mixin is missing")
	}
	return &dataSourceDataSetMixin{
		Provider:        provider,
		dataSourceMixin: dataSourceMixin,
		dataSetMixin:    dataSetMixin,
	}, nil
}

type dataSetMixin = dataSetWork.Mixin

type dataSourceDataSetMixin struct {
	work.Provider
	dataSourceMixin
	dataSetMixin
}

func (d *dataSourceDataSetMixin) FetchDataSetFromDataSource() *work.ProcessResult {
	if !d.HasDataSource() {
		return d.Failed(errors.New("data source is missing"))
	} else if dataSetID := d.DataSource().DataSetID; dataSetID == nil {
		return d.Failed(errors.New("data source data set id is missing"))
	} else {
		return d.FetchDataSet(*dataSetID)
	}
}

type DataSourceReplacerMixin interface {
	ReplaceDataSource(replacementDataSource *dataSource.Source) *work.ProcessResult
}

func NewDataSourceReplacerMixin(provider work.Provider, dataSourceMixin dataSourceWork.Mixin, providerSessionClient providerSession.Client) (DataSourceReplacerMixin, error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}
	if dataSourceMixin == nil {
		return nil, errors.New("data source mixin is missing")
	}
	if providerSessionClient == nil {
		return nil, errors.New("provider session client is missing")
	}
	return &dataSourceReplacerMixin{
		Provider:              provider,
		dataSourceMixin:       dataSourceMixin,
		providerSessionClient: providerSessionClient,
	}, nil
}

type providerSessionClient = providerSession.Client

type dataSourceReplacerMixin struct {
	work.Provider
	dataSourceMixin
	providerSessionClient
}

// FUTURE: Consider moving this functionality to a transaction
func (d *dataSourceReplacerMixin) ReplaceDataSource(replacementDataSource *dataSource.Source) *work.ProcessResult {
	if replacementDataSource == nil {
		return d.Failed(errors.New("replacement data source is missing"))
	}

	ctx := d.Context()

	// If replacement is not disconnected, then delete associated provider session, which will disconnect data source
	var replacementDataSourceUpdate *dataSource.Update
	if replacementDataSource.State != dataSource.StateDisconnected {
		if replacementDataSource.ProviderSessionID != nil {
			if err := d.DeleteProviderSession(ctx, *replacementDataSource.ProviderSessionID); err != nil {
				return d.Failing(errors.Wrap(err, "unable to delete replacement data source provider session"))
			} else if replacementDataSource, err = d.DataSourceClient().Get(ctx, replacementDataSource.ID); err != nil {
				return d.Failing(errors.Wrap(err, "unable to get replacement data source after deleting provider session"))
			}
		} else {
			log.LoggerFromContext(ctx).WithField("replacementDataSourceId", replacementDataSource.ID).Warn("replacement data source not disconnected and without provider session id")
			replacementDataSourceUpdate = &dataSource.Update{
				State: pointer.FromString(dataSource.StateDisconnected),
			}
		}
	}

	// Do not interrupt
	ctx = context.WithoutCancel(ctx)

	// If there is a data source, then update the replacement to match provider session and state
	originalDataSource := d.DataSource()
	if originalDataSource != nil {
		replacementDataSourceUpdate = &dataSource.Update{
			ProviderSessionID: originalDataSource.ProviderSessionID,
			State:             pointer.FromString(originalDataSource.State),
		}
	}

	// Set the replacement data source
	if result := d.SetDataSource(replacementDataSource); result != nil {
		return result
	}

	// Update the replacement data source
	if replacementDataSourceUpdate != nil {
		if result := d.UpdateDataSource(replacementDataSourceUpdate); result != nil {
			return result
		}
	}

	// If there is work metadata, then update that
	if dataSourceMixinFromWork, ok := d.dataSourceMixin.(dataSourceWork.MixinFromWork); ok && dataSourceMixinFromWork.HasWorkMetadata() {
		if result := dataSourceMixinFromWork.UpdateWorkMetadataFromDataSource(); result != nil {
			return result
		}
	}

	// Delete any existing data source (do not disconnect first, just delete, the replacement already assumed the provider session)
	if originalDataSource != nil {
		if _, err := d.DataSourceClient().Delete(ctx, originalDataSource.ID, nil); err != nil {
			log.LoggerFromContext(ctx).WithError(err).WithField("dataSourceId", originalDataSource.ID).Warn("unable to delete existing data source")
		}
	}

	return nil
}
