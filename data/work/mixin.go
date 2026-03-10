package work

import (
	"context"

	providerSession "github.com/tidepool-org/platform/auth/providersession"
	providerSessionWork "github.com/tidepool-org/platform/auth/providersession/work"
	dataRaw "github.com/tidepool-org/platform/data/raw"
	dataRawWork "github.com/tidepool-org/platform/data/raw/work"
	dataSet "github.com/tidepool-org/platform/data/set"
	dataSetWork "github.com/tidepool-org/platform/data/set/work"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceWork "github.com/tidepool-org/platform/data/source/work"
	"github.com/tidepool-org/platform/errors"
	oauthWork "github.com/tidepool-org/platform/oauth/work"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	MetadataKeyTimeRange = "timeRange"
)

type Dependencies struct {
	ProviderSessionClient providerSession.Client
	DataSourceClient      dataSource.Client
	DataRawClient         dataRaw.Client
	DataSetClient         dataSet.Client
}

func (d Dependencies) Validate() error {
	if d.ProviderSessionClient == nil {
		return errors.New("provider session client is missing")
	}
	if d.DataSourceClient == nil {
		return errors.New("data source client is missing")
	}
	if d.DataRawClient == nil {
		return errors.New("data raw client is missing")
	}
	if d.DataSetClient == nil {
		return errors.New("data set client is missing")
	}
	return nil
}

type (
	ProviderSessionMixin = providerSessionWork.Mixin
	OAuthMixin           = oauthWork.Mixin
	DataSourceMixin      = dataSourceWork.Mixin
	DataRawMixin         = dataRawWork.Mixin
	DataSetMixin         = dataSetWork.Mixin
)

type DataSourceProviderSessionMixin struct {
	*workBase.Processor
	dataSourceMixin      *DataSourceMixin
	providerSessionMixin *ProviderSessionMixin
}

func NewDataSourceProviderSessionMixin(processor *workBase.Processor, dataSourceMixin *DataSourceMixin, providerSessionMixin *ProviderSessionMixin) (*DataSourceProviderSessionMixin, error) {
	if processor == nil {
		return nil, errors.New("processor is missing")
	}
	if dataSourceMixin == nil {
		return nil, errors.New("data source mixin is missing")
	}
	if providerSessionMixin == nil {
		return nil, errors.New("provider session mixin is missing")
	}

	return &DataSourceProviderSessionMixin{
		Processor:            processor,
		dataSourceMixin:      dataSourceMixin,
		providerSessionMixin: providerSessionMixin,
	}, nil
}

func (d *DataSourceProviderSessionMixin) ProviderSessionIDFromDataSource() (*string, error) {
	if d.dataSourceMixin.DataSource == nil {
		return nil, errors.New("data source is missing")
	}
	return d.dataSourceMixin.DataSource.ProviderSessionID, nil
}

func (d *DataSourceProviderSessionMixin) FetchProviderSessionFromDataSource() *work.ProcessResult {
	providerSessionID, err := d.ProviderSessionIDFromDataSource()
	if err != nil || providerSessionID == nil {
		return d.Failed(errors.Wrap(err, "unable to get provider session id from data source"))
	}
	return d.providerSessionMixin.FetchProviderSession(*providerSessionID)
}

func (d *DataSourceProviderSessionMixin) ReplaceDataSource(dataSrc *dataSource.Source) *work.ProcessResult {
	if dataSrc == nil {
		return d.Failed(errors.New("replacement data source is missing"))
	}

	ctx := d.Context()

	// If replacement is not disconnected, then delete associated provider session, which will disconnect data source
	if dataSrc.State != dataSource.StateDisconnected {
		d.Logger().WithField("dataSourceId", dataSrc.ID).Warn("replacement data source not disconnected")
		if dataSrc.ProviderSessionID != nil {
			if err := d.providerSessionMixin.Client.DeleteProviderSession(ctx, *dataSrc.ProviderSessionID); err != nil {
				return d.Failing(errors.Wrap(err, "unable to delete replacement data source provider session"))
			}
		} else {
			dataSrc.State = dataSource.StateDisconnected
		}
	}

	// Do not interrupt
	ctx = context.WithoutCancel(ctx)

	// If there is a data source, then update the replacement to match provider session and state
	if d.dataSourceMixin.DataSource != nil {
		var err error

		dataSrcUpdate := dataSource.Update{
			ProviderSessionID: d.dataSourceMixin.DataSource.ProviderSessionID,
			State:             &d.dataSourceMixin.DataSource.State,
		}
		if dataSrc, err = d.dataSourceMixin.Client.Update(ctx, dataSrc.ID, nil, &dataSrcUpdate); err != nil {
			return d.Failing(errors.Wrap(err, "unable to update replacement data source"))
		}
	}

	// Update metadata with new data source id, if necessary
	if metadata := d.Metadata(); metadata != nil {
		if _, ok := metadata[dataSourceWork.MetadataKeyID]; ok {
			metadata[dataSourceWork.MetadataKeyID] = dataSrc.ID
			if result := d.ProcessingUpdate(); result != nil {
				return result
			}
		}
	}

	// No matter what, we are replaced
	defer func() { d.dataSourceMixin.DataSource = dataSrc }()

	// Delete any existing data source (do not disconnect first, just delete, the replacement assumes the provider session)
	if d.dataSourceMixin.DataSource != nil {
		if _, err := d.dataSourceMixin.Client.Delete(ctx, d.dataSourceMixin.DataSource.ID, nil); err != nil {
			d.Logger().WithField("dataSourceId", d.dataSourceMixin.DataSource.ID).Warn("unable to delete existing data source")
		}
	}

	return nil
}

type DataSourceDataSetMixin struct {
	*workBase.Processor
	dataSourceMixin *DataSourceMixin
	dataSetMixin    *DataSetMixin
}

func NewDataSourceDataSetMixin(processor *workBase.Processor, dataSourceMixin *DataSourceMixin, dataSetMixin *DataSetMixin) (*DataSourceDataSetMixin, error) {
	if processor == nil {
		return nil, errors.New("processor is missing")
	}
	if dataSourceMixin == nil {
		return nil, errors.New("data source mixin is missing")
	}
	if dataSetMixin == nil {
		return nil, errors.New("data set mixin is missing")
	}

	return &DataSourceDataSetMixin{
		Processor:       processor,
		dataSourceMixin: dataSourceMixin,
		dataSetMixin:    dataSetMixin,
	}, nil
}

func (d *DataSourceDataSetMixin) DataSetIDFromDataSource() (*string, error) {
	if d.dataSourceMixin.DataSource == nil {
		return nil, errors.New("data source is missing")
	}
	return d.dataSourceMixin.DataSource.DataSetID, nil
}

func (d *DataSourceDataSetMixin) FetchDataSetFromDataSource() *work.ProcessResult {
	providerSessionID, err := d.DataSetIDFromDataSource()
	if err != nil || providerSessionID == nil {
		return d.Failed(errors.Wrap(err, "unable to get provider session id from data source"))
	}
	return d.dataSetMixin.FetchDataSet(*providerSessionID)
}

type Mixin struct {
	*workBase.Processor
	*ProviderSessionMixin
	*DataSourceMixin
	*DataRawMixin
	*DataSetMixin
	*OAuthMixin
	*DataSourceProviderSessionMixin
	*DataSourceDataSetMixin
}

func NewMixin(processor *workBase.Processor, dependencies Dependencies) (*Mixin, error) {
	if processor == nil {
		return nil, errors.New("processor is missing")
	}
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}

	providerSessionMixin, err := providerSessionWork.NewMixin(processor, dependencies.ProviderSessionClient)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create provider session mixin")
	}
	dataSourceMixin, err := dataSourceWork.NewMixin(processor, dependencies.DataSourceClient)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create data source mixin")
	}
	dataRawMixin, err := dataRawWork.NewMixin(processor, dependencies.DataRawClient)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create data raw mixin")
	}
	dataSetMixin, err := dataSetWork.NewMixin(processor, dependencies.DataSetClient)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create data set mixin")
	}

	oauthMixin, err := oauthWork.NewMixin(processor, providerSessionMixin)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create oauth mixin")
	}
	dataSourceProviderSessionMixin, err := NewDataSourceProviderSessionMixin(processor, dataSourceMixin, providerSessionMixin)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create data source provider session mixin")
	}
	dataSourceDataSetMixin, err := NewDataSourceDataSetMixin(processor, dataSourceMixin, dataSetMixin)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create data source data set mixin")
	}

	return &Mixin{
		Processor:                      processor,
		ProviderSessionMixin:           providerSessionMixin,
		DataSourceMixin:                dataSourceMixin,
		DataRawMixin:                   dataRawMixin,
		DataSetMixin:                   dataSetMixin,
		OAuthMixin:                     oauthMixin,
		DataSourceProviderSessionMixin: dataSourceProviderSessionMixin,
		DataSourceDataSetMixin:         dataSourceDataSetMixin,
	}, nil
}

func (m *Mixin) TimeRangeFromMetadata() (*TimeRange, error) {
	parser := m.MetadataParser()
	timeRange := ParseTimeRange(parser.WithReferenceObjectParser(MetadataKeyTimeRange))
	if err := parser.Error(); err != nil {
		return nil, errors.Wrap(err, "unable to parse time range from metadata")
	}
	return timeRange, nil
}
