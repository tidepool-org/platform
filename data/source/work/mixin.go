package work

import (
	"context"

	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/work"
)

const MetadataKeyDataSourceID = "dataSourceId"

type Metadata struct {
	DataSourceID *string `json:"dataSourceId,omitempty" bson:"dataSourceId,omitempty"`
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	m.DataSourceID = parser.String(MetadataKeyDataSourceID)
}

func (m *Metadata) Validate(validator structure.Validator) {
	validator.String(MetadataKeyDataSourceID, m.DataSourceID).Using(dataSource.IDValidator)
}

//go:generate mockgen -source=mixin.go -destination=test/mixin_mocks.go -package=test Mixin
type Mixin interface {
	DataSourceClient() dataSource.Client

	HasDataSource() bool
	DataSource() *dataSource.Source
	SetDataSource(dataSource *dataSource.Source) *work.ProcessResult

	FetchDataSource(dataSourceID string) *work.ProcessResult
	FetchDataSourceFromProviderSessionID(providerSessionID string) *work.ProcessResult
	UpdateDataSource(dataSourceUpdate *dataSource.Update) *work.ProcessResult

	AddDataSourceToContext()
}

type FromWork interface {
	HasWorkMetadata() bool

	FetchDataSourceFromWorkMetadata() *work.ProcessResult
	UpdateWorkMetadataFromDataSource() *work.ProcessResult
}

type WithParsedMetadata[M any] interface {
	HasDataSourceMetadata() bool
	DataSourceMetadata() *M
	SetDataSourceMetadata(dataSourceMetadata *M) *work.ProcessResult

	UpdateDataSourceMetadata() *work.ProcessResult
}

//go:generate mockgen -source=mixin.go -destination=test/mixin_mocks.go -package=test MixinFromWork
type MixinFromWork interface {
	Mixin
	FromWork
}

//go:generate mockgen -source=mixin.go -destination=test/mixin_mocks.go -package=test MixinWithParsedMetadata
type MixinWithParsedMetadata[M any] interface {
	Mixin
	WithParsedMetadata[M]
}

//go:generate mockgen -source=mixin.go -destination=test/mixin_mocks.go -package=test MixinFromWorkWithParsedMetadata
type MixinFromWorkWithParsedMetadata[M any] interface {
	Mixin
	FromWork
	WithParsedMetadata[M]
}

func NewMixin(provider work.Provider, dataSourceClient dataSource.Client) (Mixin, error) {
	return NewMixinWithParsedMetadata[map[string]any](provider, dataSourceClient)
}

func NewMixinFromWork(provider work.Provider, dataSourceClient dataSource.Client, workMetadata *Metadata) (MixinFromWork, error) {
	return NewMixinFromWorkWithParsedMetadata[map[string]any](provider, dataSourceClient, workMetadata)
}

func NewMixinWithParsedMetadata[M any](provider work.Provider, dataSourceClient dataSource.Client) (MixinWithParsedMetadata[M], error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}
	if dataSourceClient == nil {
		return nil, errors.New("data source client is missing")
	}
	return &mixin[M]{
		Provider:         provider,
		dataSourceClient: dataSourceClient,
	}, nil
}

func NewMixinFromWorkWithParsedMetadata[M any](provider work.Provider, dataSourceClient dataSource.Client, workMetadata *Metadata) (MixinFromWorkWithParsedMetadata[M], error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}
	if dataSourceClient == nil {
		return nil, errors.New("data source client is missing")
	}
	if workMetadata == nil {
		return nil, errors.New("work metadata is missing")
	}
	return &mixin[M]{
		Provider:         provider,
		dataSourceClient: dataSourceClient,
		workMetadata:     workMetadata,
	}, nil
}

type mixin[M any] struct {
	work.Provider
	dataSourceClient   dataSource.Client
	dataSource         *dataSource.Source
	dataSourceMetadata *M
	workMetadata       *Metadata
}

func (m *mixin[M]) DataSourceClient() dataSource.Client {
	return m.dataSourceClient
}

func (m *mixin[M]) HasDataSource() bool {
	return m.dataSource != nil
}

func (m *mixin[M]) DataSource() *dataSource.Source {
	return m.dataSource
}

func (m *mixin[M]) SetDataSource(dataSrc *dataSource.Source) *work.ProcessResult {
	var dataSrcMetadata *M
	if dataSrc != nil {
		var err error
		if dataSrcMetadata, err = metadata.Decode[M](m.Context(), dataSrc.Metadata); err != nil {
			return m.Failing(errors.Wrap(err, "unable to decode data source metadata"))
		}
	}
	m.dataSource = dataSrc
	m.dataSourceMetadata = dataSrcMetadata
	m.AddDataSourceToContext()
	return nil
}

func (m *mixin[M]) HasDataSourceMetadata() bool {
	return m.dataSourceMetadata != nil
}

func (m *mixin[M]) DataSourceMetadata() *M {
	return m.dataSourceMetadata
}

func (m *mixin[M]) SetDataSourceMetadata(dataSourceMetadata *M) *work.ProcessResult {
	m.dataSourceMetadata = dataSourceMetadata
	m.AddDataSourceToContext()
	return nil
}

func (m *mixin[M]) FetchDataSource(dataSourceID string) *work.ProcessResult {
	if dataSrc, err := m.dataSourceClient.Get(m.Context(), dataSourceID); err != nil {
		return m.Failing(errors.Wrap(err, "unable to get data source"))
	} else if dataSrc == nil {
		return m.Failed(errors.New("data source is missing"))
	} else {
		return m.SetDataSource(dataSrc)
	}
}

func (m *mixin[M]) FetchDataSourceFromProviderSessionID(providerSessionID string) *work.ProcessResult {
	if dataSrc, err := m.dataSourceClient.GetFromProviderSession(m.Context(), providerSessionID); err != nil {
		return m.Failing(errors.Wrap(err, "unable to get data source from provider session"))
	} else if dataSrc == nil {
		return m.Failed(errors.New("data source is missing"))
	} else {
		return m.SetDataSource(dataSrc)
	}
}

func (m *mixin[M]) UpdateDataSource(dataSourceUpdate *dataSource.Update) *work.ProcessResult {
	if dataSourceUpdate == nil {
		return m.Failed(errors.New("data source update is missing"))
	}
	if m.dataSource == nil {
		return m.Failed(errors.New("data source is missing"))
	}

	if dataSrcMetadata, err := metadata.Encode(m.dataSourceMetadata); err != nil {
		return m.Failing(errors.Wrap(err, "unable to encode data source metadata"))
	} else if dataSrcMetadata != nil || m.dataSource.Metadata != nil {
		dataSourceUpdate.Metadata = &dataSrcMetadata
	}

	if dataSrc, err := m.dataSourceClient.Update(context.WithoutCancel(m.Context()), m.dataSource.ID, nil, dataSourceUpdate); err != nil {
		return m.Failing(errors.Wrap(err, "unable to update data source"))
	} else if dataSrc == nil {
		return m.Failed(errors.New("data source is missing"))
	} else {
		return m.SetDataSource(dataSrc)
	}
}

func (m *mixin[M]) UpdateDataSourceMetadata() *work.ProcessResult {
	return m.UpdateDataSource(&dataSource.Update{})
}

func (m *mixin[M]) HasWorkMetadata() bool {
	return m.workMetadata != nil
}

func (m *mixin[M]) FetchDataSourceFromWorkMetadata() *work.ProcessResult {
	if m.workMetadata == nil {
		return m.Failed(errors.New("work metadata is missing"))
	} else if m.workMetadata.DataSourceID == nil {
		return m.Failed(errors.New("work metadata data source id is missing"))
	} else {
		return m.FetchDataSource(*m.workMetadata.DataSourceID)
	}
}

func (m *mixin[M]) UpdateWorkMetadataFromDataSource() *work.ProcessResult {
	if m.dataSource == nil {
		return m.Failed(errors.New("data source is missing"))
	} else if m.workMetadata == nil {
		return m.Failed(errors.New("work metadata is missing"))
	}
	m.workMetadata.DataSourceID = &m.dataSource.ID
	return nil
}

func (m *mixin[M]) AddDataSourceToContext() {
	m.AddFieldsToContext(log.Fields{"dataSource": dataSourceToFields(m.dataSource), "dataSourceMetadata": m.dataSourceMetadata})
}

func dataSourceToFields(dataSrc *dataSource.Source) log.Fields {
	if dataSrc == nil {
		return nil
	}
	return log.Fields{
		"id":                 dataSrc.ID,
		"userId":             dataSrc.UserID,
		"providerType":       dataSrc.ProviderType,
		"providerName":       dataSrc.ProviderName,
		"providerExternalId": dataSrc.ProviderExternalID,
		"providerSessionId":  dataSrc.ProviderSessionID,
		"state":              dataSrc.State,
		"metadata":           dataSrc.Metadata,
		"dataSetId":          dataSrc.DataSetID,
	}
}
