package work

import (
	"context"
	"io"

	providerSession "github.com/tidepool-org/platform/auth/providersession"
	providerSessionWork "github.com/tidepool-org/platform/auth/providersession/work"
	"github.com/tidepool-org/platform/data"
	dataRaw "github.com/tidepool-org/platform/data/raw"
	dataRawWork "github.com/tidepool-org/platform/data/raw/work"
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
	if prvdrSession := p.ProviderSession(); prvdrSession == nil {
		return p.Failed(errors.New("provider session is missing"))
	} else {
		return p.FetchDataSourceFromProviderSessionID(prvdrSession.ID)
	}
}

func (p *providerSessionDataSourceMixin) FetchProviderSessionFromDataSource() *work.ProcessResult {
	if dataSrc := p.DataSource(); dataSrc == nil {
		return p.Failed(errors.New("data source is missing"))
	} else if providerSessionID := dataSrc.ProviderSessionID; providerSessionID == nil {
		return p.Failed(errors.New("data source provider session id is missing"))
	} else {
		return p.FetchProviderSession(*providerSessionID)
	}
}

type DataSourceDataSetMixin interface {
	FetchDataSetFromDataSource() *work.ProcessResult
	CreateDataSetForDataSource(dataSetCreate *data.DataSetCreate) *work.ProcessResult
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
	if dataSrc := d.DataSource(); dataSrc == nil {
		return d.Failed(errors.New("data source is missing"))
	} else if dataSetID := dataSrc.DataSetID; dataSetID == nil {
		return d.Failed(errors.New("data source data set id is missing"))
	} else {
		return d.FetchDataSet(*dataSetID)
	}
}

func (d *dataSourceDataSetMixin) CreateDataSetForDataSource(dataSetCreate *data.DataSetCreate) *work.ProcessResult {
	if dataSrc := d.DataSource(); dataSrc == nil {
		return d.Failed(errors.New("data source is missing"))
	} else if dataSetID := dataSrc.DataSetID; dataSetID != nil {
		return d.Failed(errors.New("data source data set id already exists"))
	} else if result := d.CreateDataSet(dataSrc.UserID, dataSetCreate); result != nil {
		return result
	} else {
		return d.UpdateDataSource(&dataSource.Update{DataSetID: d.DataSet().ID})
	}
}

type DataSourceDataRawMixin interface {
	CreateDataRawForDataSource(dataRawCreate *dataRaw.Create, reader io.Reader) *work.ProcessResult
}

func NewDataSourceDataRawMixin(provider work.Provider, dataSourceMixin dataSourceWork.Mixin, dataRawMixin dataRawWork.Mixin) (DataSourceDataRawMixin, error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}
	if dataSourceMixin == nil {
		return nil, errors.New("data source mixin is missing")
	}
	if dataRawMixin == nil {
		return nil, errors.New("data raw mixin is missing")
	}
	return &dataSourceDataRawMixin{
		Provider:        provider,
		dataSourceMixin: dataSourceMixin,
		dataRawMixin:    dataRawMixin,
	}, nil
}

type dataRawMixin = dataRawWork.Mixin

type dataSourceDataRawMixin struct {
	work.Provider
	dataSourceMixin
	dataRawMixin
}

func (d *dataSourceDataRawMixin) CreateDataRawForDataSource(dataRawCreate *dataRaw.Create, reader io.Reader) *work.ProcessResult {
	if dataSrc := d.DataSource(); dataSrc == nil {
		return d.Failed(errors.New("data source is missing"))
	} else if dataSetID := dataSrc.DataSetID; dataSetID == nil {
		return d.Failed(errors.New("data source data set id is missing"))
	} else if result := d.CreateDataRaw(dataSrc.UserID, *dataSrc.DataSetID, dataRawCreate, reader); result != nil {
		return result
	} else {
		return d.UpdateDataSource(&dataSource.Update{LastImportTime: pointer.From(d.DataRaw().CreatedTime)})
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
				State: pointer.From(dataSource.StateDisconnected),
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
			State:             pointer.From(originalDataSource.State),
		}
	}

	// Set the replacement data source
	if result := d.SetDataSource(replacementDataSource); result != nil {
		return result
	}

	// Delete any existing data source (do not disconnect first, just delete, the replacement already assumed the provider session)
	if originalDataSource != nil {
		if _, err := d.DataSourceClient().Delete(ctx, originalDataSource.ID, nil); err != nil {
			log.LoggerFromContext(ctx).WithError(err).WithField("dataSourceId", originalDataSource.ID).Warn("unable to delete existing data source")
		}
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

	return nil
}
