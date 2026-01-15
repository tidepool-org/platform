package work

import (
	"context"

	dataRaw "github.com/tidepool-org/platform/data/raw"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	MetadataKeyID              = "dataRawId"
	MetadataKeyIngestionOffset = "ingestionOffset"
)

type Mixin struct {
	*workBase.Processor
	Client  dataRaw.Client
	DataRaw *dataRaw.Raw
}

func NewMixin(processor *workBase.Processor, client dataRaw.Client) (*Mixin, error) {
	if processor == nil {
		return nil, errors.New("processor is missing")
	}
	if client == nil {
		return nil, errors.New("client is missing")
	}
	return &Mixin{
		Processor: processor,
		Client:    client,
	}, nil
}

func (m *Mixin) DataRawIDFromMetadata() (*string, error) {
	parser := m.MetadataParser()
	dataRawID := parser.String(MetadataKeyID)
	if err := parser.Error(); err != nil {
		return nil, errors.Wrap(err, "unable to parse data raw id from metadata")
	}
	return dataRawID, nil
}

func (m *Mixin) FetchDataRawFromMetadata() *work.ProcessResult {
	dataRawID, err := m.DataRawIDFromMetadata()
	if err != nil || dataRawID == nil {
		return m.Failed(errors.Wrap(err, "unable to get data raw id from metadata"))
	}
	return m.FetchDataRaw(*dataRawID)
}

func (m *Mixin) FetchDataRaw(dataRawID string) *work.ProcessResult {
	dataRaw, err := m.Client.Get(m.Context(), dataRawID, nil)
	if err != nil {
		return m.Failing(errors.Wrap(err, "unable to fetch data raw"))
	} else if dataRaw == nil {
		return m.Failed(errors.New("data raw is missing"))
	}
	m.DataRaw = dataRaw

	m.AddFieldToContext("dataRaw", log.Fields{"id": m.DataRaw.ID, "dataSetId": m.DataRaw.DataSetID, "userId": m.DataRaw.UserID})

	return nil
}

func (m *Mixin) UpdateDataRaw(dataRawUpdate dataRaw.Update) *work.ProcessResult {
	if m.DataRaw == nil {
		return m.Failed(errors.New("data raw is missing"))
	}

	src, err := m.Client.Update(context.WithoutCancel(m.Context()), m.DataRaw.ID, nil, &dataRawUpdate)
	if err != nil {
		return m.Failing(errors.Wrap(err, "unable to update data raw"))
	} else if src == nil {
		return m.Failed(errors.New("data raw is missing"))
	}

	m.DataRaw = src
	return nil
}

func (m *Mixin) IngestionOffset() (*int, error) {
	parser := m.MetadataParser()
	ingestionOffset := parser.Int(MetadataKeyIngestionOffset)
	if err := parser.Error(); err != nil {
		return nil, errors.Wrap(err, "unable to parse ingestion offset from metadata")
	}
	return ingestionOffset, nil
}
