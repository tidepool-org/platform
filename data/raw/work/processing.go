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

//go:generate mockgen -source=processing.go -destination=test/processing_mocks.go -package=test Client
type Client interface {
	Get(ctx context.Context, id string, condition *request.Condition) (*dataRaw.Raw, error)
	Update(ctx context.Context, id string, condition *request.Condition, update *dataRaw.Update) (*dataRaw.Raw, error)
}

type Processing struct {
	*workBase.Processing
	Client  Client
	DataRaw *dataRaw.Raw
}

func NewProcessing(processing *workBase.Processing, client Client) (*Processing, error) {
	if processing == nil {
		return nil, errors.New("processing is missing")
	}
	if client == nil {
		return nil, errors.New("client is missing")
	}
	return &Processing{
		Processing: processing,
		Client:     client,
	}, nil
}

func (p *Processing) DataRawIDFromMetadata() (*string, error) {
	parser := p.MetadataParser()
	dataRawID := parser.String(MetadataKeyID)
	if err := parser.Error(); err != nil {
		return nil, errors.Wrap(err, "unable to parse data raw id from metadata")
	}
	return dataRawID, nil
}

func (p *Processing) FetchDataRawFromMetadata() *work.ProcessResult {
	dataRawID, err := p.DataRawIDFromMetadata()
	if err != nil || dataRawID == nil {
		return p.Failed(errors.Wrap(err, "unable to get data raw id from metadata"))
	}
	return p.FetchDataRaw(*dataRawID)
}

func (p *Processing) FetchDataRaw(dataRawID string) *work.ProcessResult {
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

func (p *Processing) UpdateDataRaw(dataRawUpdate dataRaw.Update) *work.ProcessResult {
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
