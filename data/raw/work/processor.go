package work

import (
	"context"

	dataRaw "github.com/tidepool-org/platform/data/raw"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	MetadataKeyID = "dataRawId"
)

//go:generate mockgen -source=processor.go -destination=test/processor_mocks.go -package=test Client
type Client interface {
	Get(ctx context.Context, id string, condition *request.Condition) (*dataRaw.Raw, error)
	Update(ctx context.Context, id string, condition *request.Condition, update *dataRaw.Update) (*dataRaw.Raw, error)
}

type Processor struct {
	*workBase.Processor
	Client  Client
	DataRaw *dataRaw.Raw
}

func NewProcessor(processor *workBase.Processor, client Client) (*Processor, error) {
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

func (p *Processor) DataRawIDFromMetadata() (*string, error) {
	parser := p.MetadataParser()
	dataRawID := parser.String(MetadataKeyID)
	if err := parser.Error(); err != nil {
		return nil, errors.Wrap(err, "unable to parse data raw id from metadata")
	}
	return dataRawID, nil
}

func (p *Processor) FetchDataRawFromMetadata() *work.ProcessResult {
	dataRawID, err := p.DataRawIDFromMetadata()
	if err != nil || dataRawID == nil {
		return p.Failed(errors.Wrap(err, "unable to get data raw id from metadata"))
	}
	return p.FetchDataRaw(*dataRawID)
}

func (p *Processor) FetchDataRaw(dataRawID string) *work.ProcessResult {
	dataRaw, err := p.Client.Get(p.Context(), dataRawID, nil)
	if err != nil {
		return p.Failing(errors.Wrap(err, "unable to fetch data raw"))
	} else if dataRaw == nil {
		return p.Failed(errors.New("data raw is missing"))
	}
	p.DataRaw = dataRaw

	p.ContextWithField("dataRaw", log.Fields{"id": p.DataRaw.ID, "dataSetId": p.DataRaw.DataSetID, "userId": p.DataRaw.UserID})

	return nil
}

func (p *Processor) UpdateDataRaw(dataRawUpdate dataRaw.Update) *work.ProcessResult {
	if p.DataRaw == nil {
		return p.Failed(errors.New("data raw is missing"))
	}

	src, err := p.Client.Update(context.WithoutCancel(p.Context()), p.DataRaw.ID, nil, &dataRawUpdate)
	if err != nil {
		return p.Failing(errors.Wrap(err, "unable to update data raw"))
	} else if src == nil {
		return p.Failed(errors.New("data raw is missing"))
	}

	p.DataRaw = src
	return nil
}
