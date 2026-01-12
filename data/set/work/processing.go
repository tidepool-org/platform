package work

import (
	"context"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	MetadataKeyID = "dataSetId"
)

//go:generate mockgen -source=processing.go -destination=test/processing_mocks.go -package=test Client
type Client interface {
	Get(ctx context.Context, id string, condition *request.Condition) (*data.DataSet, error)
	Update(ctx context.Context, id string, condition *request.Condition, update *data.DataSetUpdate) (*data.DataSet, error)
}

type Processing struct {
	*workBase.Processing
	Client  Client
	DataSet *data.DataSet
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

func (p *Processing) DataSetIDFromMetadata() (*string, error) {
	parser := p.MetadataParser()
	dataSetID := parser.String(MetadataKeyID)
	if err := parser.Error(); err != nil {
		return nil, errors.Wrap(err, "unable to parse data set id from metadata")
	}
	return dataSetID, nil
}

func (p *Processing) FetchDataSetFromMetadata() *work.ProcessResult {
	dataSetID, err := p.DataSetIDFromMetadata()
	if err != nil || dataSetID == nil {
		return p.Failed(errors.Wrap(err, "unable to get data set id from metadata"))
	}
	return p.FetchDataSet(*dataSetID)
}

func (p *Processing) FetchDataSet(dataSetID string) *work.ProcessResult {
	dataSet, err := p.Client.Get(p.Context(), dataSetID, nil)
	if err != nil {
		return p.Failing(errors.Wrap(err, "unable to fetch data set"))
	} else if dataSet == nil {
		return p.Failed(errors.New("data set is missing"))
	}
	p.DataSet = dataSet

	p.ContextWithField("dataSet", log.Fields{"id": p.DataSet.ID, "userId": p.DataSet.UserID})

	return nil
}

func (p *Processing) UpdateDataSet(dataSetUpdate data.DataSetUpdate) *work.ProcessResult {
	if p.DataSet == nil {
		return p.Failed(errors.New("data set is missing"))
	}

	src, err := p.Client.Update(context.WithoutCancel(p.Context()), *p.DataSet.ID, nil, &dataSetUpdate)
	if err != nil {
		return p.Failing(errors.Wrap(err, "unable to update data set"))
	} else if src == nil {
		return p.Failed(errors.New("data set is missing"))
	}

	p.DataSet = src
	return nil
}
