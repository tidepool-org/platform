package work

import (
	"context"

	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	MetadataKeyID           = "dataSourceId"
	MetadataKeyDeviceHashes = "deviceHashes"
)

type Mixin struct {
	*workBase.Processor
	Client     dataSource.Client
	DataSource *dataSource.Source
}

func NewMixin(processor *workBase.Processor, client dataSource.Client) (*Mixin, error) {
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

func (m *Mixin) DataSourceIDFromMetadata() (*string, error) {
	parser := m.MetadataParser()
	dataSrcID := parser.String(MetadataKeyID)
	if err := parser.Error(); err != nil {
		return nil, errors.Wrap(err, "unable to parse data source id from metadata")
	}
	return dataSrcID, nil
}

func (m *Mixin) FetchDataSourceFromMetadata() *work.ProcessResult {
	dataSrcID, err := m.DataSourceIDFromMetadata()
	if err != nil || dataSrcID == nil {
		return m.Failed(errors.Wrap(err, "unable to get data source id from metadata"))
	}
	return m.FetchDataSource(*dataSrcID)
}

func (m *Mixin) FetchDataSource(dataSrcID string) *work.ProcessResult {
	dataSrc, err := m.Client.Get(m.Context(), dataSrcID)
	if err != nil {
		return m.Failing(errors.Wrap(err, "unable to fetch data source"))
	} else if dataSrc == nil {
		return m.Failed(errors.New("data source is missing"))
	}
	m.DataSource = dataSrc

	m.AddFieldToContext("dataSource", log.Fields{"id": m.DataSource.ID, "dataSetIds": m.DataSource.DataSetIDs, "userId": m.DataSource.UserID})

	return nil
}

func (m *Mixin) UpdateDataSource(dataSrcUpdate dataSource.Update) *work.ProcessResult {
	if m.DataSource == nil {
		return m.Failed(errors.New("data source is missing"))
	}

	dataSrc, err := m.Client.Update(context.WithoutCancel(m.Context()), *m.DataSource.ID, nil, &dataSrcUpdate)
	if err != nil {
		return m.Failing(errors.Wrap(err, "unable to update data source"))
	} else if dataSrc == nil {
		return m.Failed(errors.New("data source is missing"))
	}

	m.DataSource = dataSrc
	return nil
}

func (m *Mixin) DeviceHashes() (map[string]string, error) {
	parser := m.MetadataParser().WithReferenceObjectParser(MetadataKeyDeviceHashes)
	deviceHashes := map[string]string{}
	for _, deviceID := range parser.References() {
		if deviceHash := parser.String(deviceID); deviceHash != nil {
			deviceHashes[deviceID] = *deviceHash
		}
	}
	if err := parser.Error(); err != nil {
		return nil, errors.Wrap(err, "unable to parse device hashes from metadata")
	}
	return deviceHashes, nil
}
