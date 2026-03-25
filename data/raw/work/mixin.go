package work

import (
	"context"

	dataRaw "github.com/tidepool-org/platform/data/raw"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/work"
)

//go:generate mockgen -source=mixin.go -destination=test/mixin_mocks.go -package=test -typed

const MetadataKeyDataRawID = "dataRawId"

type Metadata struct {
	DataRawID *string `json:"dataRawId,omitempty" bson:"dataRawId,omitempty"`
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	m.DataRawID = parser.String(MetadataKeyDataRawID)
}

func (m *Metadata) Validate(validator structure.Validator) {
	validator.String(MetadataKeyDataRawID, m.DataRawID).Using(dataRaw.DataRawIDValidator)
}

type Mixin interface {
	DataRawClient() dataRaw.Client

	HasDataRaw() bool
	DataRaw() *dataRaw.Raw
	SetDataRaw(dataRaw *dataRaw.Raw) *work.ProcessResult

	FetchDataRaw(dataRawID string) *work.ProcessResult
	UpdateDataRaw(dataRawUpdate *dataRaw.Update) *work.ProcessResult

	HasDataRawContent() bool
	DataRawContent() *dataRaw.Content

	FetchDataRawContent(dataRawID string) *work.ProcessResult
	FetchDataRawContentFromDataRaw() *work.ProcessResult
	CloseDataRawContent() *work.ProcessResult

	AddDataRawToContext()
}

type FromWork interface {
	HasWorkMetadata() bool

	FetchDataRawFromWorkMetadata() *work.ProcessResult
	UpdateWorkMetadataFromDataRaw() *work.ProcessResult
}

type WithParsedMetadata[M any] interface {
	HasDataRawMetadata() bool
	DataRawMetadata() *M
	SetDataRawMetadata(dataRawMetadata *M) *work.ProcessResult

	UpdateDataRawMetadata() *work.ProcessResult
}

type MixinFromWork interface {
	Mixin
	FromWork
}

type MixinWithParsedMetadata[M any] interface {
	Mixin
	WithParsedMetadata[M]
}

type MixinFromWorkWithParsedMetadata[M any] interface {
	Mixin
	FromWork
	WithParsedMetadata[M]
}

func NewMixin(provider work.Provider, dataRawClient dataRaw.Client) (Mixin, error) {
	return NewMixinWithParsedMetadata[map[string]any](provider, dataRawClient)
}

func NewMixinFromWork(provider work.Provider, dataRawClient dataRaw.Client, workMetadata *Metadata) (MixinFromWork, error) {
	return NewMixinFromWorkWithParsedMetadata[map[string]any](provider, dataRawClient, workMetadata)
}

func NewMixinWithParsedMetadata[M any](provider work.Provider, dataRawClient dataRaw.Client) (MixinWithParsedMetadata[M], error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}
	if dataRawClient == nil {
		return nil, errors.New("data raw client is missing")
	}
	return &mixin[M]{
		Provider:      provider,
		dataRawClient: dataRawClient,
	}, nil
}

func NewMixinFromWorkWithParsedMetadata[M any](provider work.Provider, dataRawClient dataRaw.Client, workMetadata *Metadata) (MixinFromWorkWithParsedMetadata[M], error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}
	if dataRawClient == nil {
		return nil, errors.New("data raw client is missing")
	}
	if workMetadata == nil {
		return nil, errors.New("work metadata is missing")
	}
	return &mixin[M]{
		Provider:      provider,
		dataRawClient: dataRawClient,
		workMetadata:  workMetadata,
	}, nil
}

type mixin[M any] struct {
	work.Provider
	dataRawClient   dataRaw.Client
	dataRaw         *dataRaw.Raw
	dataRawContent  *dataRaw.Content
	dataRawMetadata *M
	workMetadata    *Metadata
}

func (m *mixin[M]) DataRawClient() dataRaw.Client {
	return m.dataRawClient
}

func (m *mixin[M]) HasDataRaw() bool {
	return m.dataRaw != nil
}

func (m *mixin[M]) DataRaw() *dataRaw.Raw {
	return m.dataRaw
}

func (m *mixin[M]) SetDataRaw(dataRw *dataRaw.Raw) *work.ProcessResult {
	var dataRwMetadata *M
	if dataRw != nil {
		var err error
		if dataRwMetadata, err = metadata.Decode[M](m.Context(), dataRw.Metadata); err != nil {
			return m.Failing(errors.Wrap(err, "unable to decode data raw metadata"))
		}
	}
	m.dataRaw = dataRw
	m.dataRawMetadata = dataRwMetadata
	m.dataRawContent = nil
	m.AddDataRawToContext()
	return nil
}

func (m *mixin[M]) HasDataRawMetadata() bool {
	return m.dataRawMetadata != nil
}

func (m *mixin[M]) DataRawMetadata() *M {
	return m.dataRawMetadata
}

func (m *mixin[M]) SetDataRawMetadata(dataRawMetadata *M) *work.ProcessResult {
	m.dataRawMetadata = dataRawMetadata
	m.AddDataRawToContext()
	return nil
}

func (m *mixin[M]) FetchDataRaw(dataRawID string) *work.ProcessResult {
	if dataRaw, err := m.dataRawClient.Get(m.Context(), dataRawID, nil); err != nil {
		return m.Failing(errors.Wrap(err, "unable to get data raw"))
	} else if dataRaw == nil {
		return m.Failed(errors.New("data raw is missing"))
	} else {
		return m.SetDataRaw(dataRaw)
	}
}

func (m *mixin[M]) UpdateDataRaw(dataRawUpdate *dataRaw.Update) *work.ProcessResult {
	if dataRawUpdate == nil {
		return m.Failed(errors.New("data raw update is missing"))
	}
	if m.dataRaw == nil {
		return m.Failed(errors.New("data raw is missing"))
	}

	if dataRwMetadata, err := metadata.Encode(m.dataRawMetadata); err != nil {
		return m.Failing(errors.Wrap(err, "unable to encode data raw metadata"))
	} else if dataRwMetadata != nil || m.dataRaw.Metadata != nil {
		dataRawUpdate.Metadata = &dataRwMetadata
	}

	if dataRw, err := m.dataRawClient.Update(context.WithoutCancel(m.Context()), m.dataRaw.ID, nil, dataRawUpdate); err != nil {
		return m.Failing(errors.Wrap(err, "unable to update data raw"))
	} else if dataRw == nil {
		return m.Failed(errors.New("data raw is missing"))
	} else {
		return m.SetDataRaw(dataRw)
	}
}

func (m *mixin[M]) UpdateDataRawMetadata() *work.ProcessResult {
	return m.UpdateDataRaw(&dataRaw.Update{})
}

func (m *mixin[M]) HasDataRawContent() bool {
	return m.dataRawContent != nil
}

func (m *mixin[M]) DataRawContent() *dataRaw.Content {
	return m.dataRawContent
}

func (m *mixin[M]) FetchDataRawContent(dataRawID string) *work.ProcessResult {
	dataRwContent, err := m.dataRawClient.GetContent(m.Context(), dataRawID, nil)
	if err != nil {
		return m.Failing(errors.Wrap(err, "unable to get data raw content"))
	} else if dataRwContent == nil {
		return m.Failed(errors.New("data raw content is missing"))
	}
	m.dataRawContent = dataRwContent
	return nil
}

func (m *mixin[M]) FetchDataRawContentFromDataRaw() *work.ProcessResult {
	if m.dataRaw == nil {
		return m.Failed(errors.New("data raw is missing"))
	}
	return m.FetchDataRawContent(m.dataRaw.ID)
}

func (m *mixin[M]) CloseDataRawContent() *work.ProcessResult {
	if m.dataRawContent != nil && m.dataRawContent.ReadCloser != nil {
		if err := m.dataRawContent.ReadCloser.Close(); err != nil {
			log.LoggerFromContext(m.Context()).WithError(err).Warn("unable to close data raw content")
		}
	}
	return nil
}

func (m *mixin[M]) HasWorkMetadata() bool {
	return m.workMetadata != nil
}

func (m *mixin[M]) FetchDataRawFromWorkMetadata() *work.ProcessResult {
	if m.workMetadata == nil {
		return m.Failed(errors.New("work metadata is missing"))
	} else if m.workMetadata.DataRawID == nil {
		return m.Failed(errors.New("work metadata data raw id is missing"))
	} else {
		return m.FetchDataRaw(*m.workMetadata.DataRawID)
	}
}

func (m *mixin[M]) UpdateWorkMetadataFromDataRaw() *work.ProcessResult {
	if m.dataRaw == nil {
		return m.Failed(errors.New("data raw is missing"))
	} else if m.workMetadata == nil {
		return m.Failed(errors.New("work metadata is missing"))
	}
	m.workMetadata.DataRawID = &m.dataRaw.ID
	return nil
}

func (m *mixin[M]) AddDataRawToContext() {
	m.AddFieldsToContext(log.Fields{"dataRaw": dataRawToFields(m.dataRaw), "dataRawMetadata": m.dataRawMetadata})
}

func dataRawToFields(dataRw *dataRaw.Raw) log.Fields {
	if dataRw == nil {
		return nil
	}
	return log.Fields{
		"id":            dataRw.ID,
		"userId":        dataRw.UserID,
		"dataSetId":     dataRw.DataSetID,
		"metadata":      dataRw.Metadata,
		"mediaType":     dataRw.MediaType,
		"size":          dataRw.Size,
		"processedTime": dataRw.ProcessedTime,
	}
}
