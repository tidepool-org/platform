package work

import (
	"context"

	authProviderSessionWork "github.com/tidepool-org/platform/auth/providersession/work"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/work"
)

const (
	MetadataKeyID           = "dataSourceId"
	MetadataKeyDeviceHashes = "deviceHashes"
)

//go:generate mockgen -source=processor.go -destination=test/processor_mocks.go -package=test Client
type Client interface {
	List(ctx context.Context, userID string, filter *dataSource.Filter, pagination *page.Pagination) (dataSource.SourceArray, error)

	Get(ctx context.Context, id string) (*dataSource.Source, error)
	Update(ctx context.Context, id string, condition *request.Condition, update *dataSource.Update) (*dataSource.Source, error)
	Destroy(ctx context.Context, id string, condition *request.Condition) (bool, error)
}

type Processor struct {
	*authProviderSessionWork.Processor
	Client     Client
	DataSource *dataSource.Source
}

func NewProcessor(processor *authProviderSessionWork.Processor, client Client) (*Processor, error) {
	if processor == nil {
		return nil, errors.New("processor is missing")
	}
	if client == nil {
		return nil, errors.New("client is missing")
	}
	return &Processor{
		Processor: processor,
		Client:    client,
	}, nil
}

func (p *Processor) ProviderSessionIDFromDataSource() (*string, error) {
	if p.DataSource == nil {
		return nil, errors.New("data source is missing")
	}
	return p.DataSource.ProviderSessionID, nil
}

func (p *Processor) FetchProviderSessionFromDataSource() *work.ProcessResult {
	if p.DataSource == nil {
		if result := p.FetchDataSourceFromMetadata(); result != nil {
			return result
		}
	}
	providerSessionID, err := p.ProviderSessionIDFromDataSource()
	if err != nil || providerSessionID == nil {
		return p.Failed(errors.Wrap(err, "unable to get provider session id from data source"))
	}
	return p.FetchProviderSession(*providerSessionID)
}

func (p *Processor) DataSourceIDFromMetadata() (*string, error) {
	parser := p.MetadataParser()
	dataSrcID := parser.String(MetadataKeyID)
	if err := parser.Error(); err != nil {
		return nil, errors.Wrap(err, "unable to parse data source id from metadata")
	}
	return dataSrcID, nil
}

func (p *Processor) FetchDataSourceFromMetadata() *work.ProcessResult {
	dataSrcID, err := p.DataSourceIDFromMetadata()
	if err != nil || dataSrcID == nil {
		return p.Failed(errors.Wrap(err, "unable to get data source id from metadata"))
	}
	return p.FetchDataSource(*dataSrcID)
}

func (p *Processor) FetchDataSource(dataSrcID string) *work.ProcessResult {
	dataSrc, err := p.Client.Get(p.Context(), dataSrcID)
	if err != nil {
		return p.Failing(errors.Wrap(err, "unable to fetch data source"))
	} else if dataSrc == nil {
		return p.Failed(errors.New("data source is missing"))
	}
	p.DataSource = dataSrc

	p.ContextWithField("dataSource", log.Fields{"id": p.DataSource.ID, "dataSetIds": p.DataSource.DataSetIDs, "userId": p.DataSource.UserID})

	return nil
}

func (p *Processor) UpdateDataSource(dataSrcUpdate dataSource.Update) *work.ProcessResult {
	if p.DataSource == nil {
		return p.Failed(errors.New("data source is missing"))
	}

	dataSrc, err := p.Client.Update(context.WithoutCancel(p.Context()), *p.DataSource.ID, nil, &dataSrcUpdate)
	if err != nil {
		return p.Failing(errors.Wrap(err, "unable to update data source"))
	} else if dataSrc == nil {
		return p.Failed(errors.New("data source is missing"))
	}

	p.DataSource = dataSrc
	return nil
}

func (p *Processor) ReplaceDataSource(dataSrc *dataSource.Source) *work.ProcessResult {
	if dataSrc == nil {
		return p.Failed(errors.New("replacement data source is missing"))
	}

	ctx := p.Context()

	// If replacement is not disconnected, then delete associated provider session, which will disconnect data source
	if *dataSrc.State != dataSource.StateDisconnected {
		p.Logger().WithField("dataSourceId", dataSrc.ID).Warn("replacement data source not disconnected")
		if dataSrc.ProviderSessionID != nil {
			if err := p.Processor.Client.DeleteProviderSession(ctx, *dataSrc.ProviderSessionID); err != nil {
				return p.Failing(errors.Wrap(err, "unable to delete replacement data source provider session"))
			}
		} else {
			dataSrc.State = pointer.FromString(dataSource.StateDisconnected)
		}
	}

	// Do not interrupt
	ctx = context.WithoutCancel(ctx)

	// If there is a data source, then update the replacement to match provider session and state
	if p.DataSource != nil {
		var err error

		dataSrcUpdate := dataSource.Update{
			ProviderSessionID: p.DataSource.ProviderSessionID,
			State:             p.DataSource.State,
		}
		if dataSrc, err = p.Client.Update(ctx, *dataSrc.ID, nil, &dataSrcUpdate); err != nil {
			return p.Failing(errors.Wrap(err, "unable to update replacement data source"))
		}
	}

	// Update metadata with new data source id, if necessary
	if metadata := p.Metadata(); metadata != nil {
		if _, ok := metadata[MetadataKeyID]; ok {
			metadata[MetadataKeyID] = *dataSrc.ID
			if result := p.Processor.ProcessingUpdate(); result != nil {
				return result
			}
		}
	}

	// No matter what, we are replaced
	defer func() { p.DataSource = dataSrc }()

	// Delete any existing data source (do not disconnect first, just delete, the replacement assumes the provider session)
	if p.DataSource != nil {
		if _, err := p.Client.Destroy(ctx, *p.DataSource.ID, nil); err != nil {
			p.Logger().WithField("dataSourceId", p.DataSource.ID).Warn("unable to delete existing data source")
		}
	}

	return nil
}

func (p *Processor) DeviceHashes() (map[string]string, error) {
	parser := p.MetadataParser().WithReferenceObjectParser(MetadataKeyDeviceHashes)
	deviceHashes := map[string]string{}
	for _, deviceID := range parser.References() {
		if deviceHash := parser.String(deviceID); deviceHash != nil {
			deviceHashes[deviceID] = *deviceHash
		}
	}
	if err := parser.Error(); err != nil {
		return nil, errors.Wrap(err, "unable to parse time range from metadata")
	}
	return deviceHashes, nil
}
