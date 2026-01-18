package work

import (
	"context"

	"github.com/tidepool-org/platform/data"
	dataSet "github.com/tidepool-org/platform/data/set"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const MetadataKeyID = "dataSetId"

type Client interface {
	Get(ctx context.Context, id string, condition *request.Condition) (*data.DataSet, error)
	Update(ctx context.Context, id string, condition *request.Condition, update *data.DataSetUpdate) (*data.DataSet, error)
}

type Mixin struct {
	*workBase.Processor
	Client  dataSet.Client
	DataSet *data.DataSet
}

func NewMixin(processor *workBase.Processor, client dataSet.Client) (*Mixin, error) {
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

func (m *Mixin) DataSetIDFromMetadata() (*string, error) {
	parser := m.MetadataParser()
	dataSetID := parser.String(MetadataKeyID)
	if err := parser.Error(); err != nil {
		return nil, errors.Wrap(err, "unable to parse data set id from metadata")
	}
	return dataSetID, nil
}

func (m *Mixin) FetchDataSetFromMetadata() *work.ProcessResult {
	dataSetID, err := m.DataSetIDFromMetadata()
	if err != nil || dataSetID == nil {
		return m.Failed(errors.Wrap(err, "unable to get data set id from metadata"))
	}
	return m.FetchDataSet(*dataSetID)
}

func (m *Mixin) FetchDataSet(dataSetID string) *work.ProcessResult {
	dataSet, err := m.Client.GetDataSet(m.Context(), dataSetID)
	if err != nil {
		return m.Failing(errors.Wrap(err, "unable to fetch data set"))
	} else if dataSet == nil {
		return m.Failed(errors.New("data set is missing"))
	}
	m.DataSet = dataSet

	m.AddFieldToContext("dataSet", log.Fields{"id": m.DataSet.ID, "userId": m.DataSet.UserID})

	return nil
}

func (m *Mixin) UpdateDataSet(dataSetUpdate data.DataSetUpdate) *work.ProcessResult {
	if m.DataSet == nil {
		return m.Failed(errors.New("data set is missing"))
	}

	src, err := m.Client.UpdateDataSet(context.WithoutCancel(m.Context()), *m.DataSet.ID, &dataSetUpdate)
	if err != nil {
		return m.Failing(errors.Wrap(err, "unable to update data set"))
	} else if src == nil {
		return m.Failed(errors.New("data set is missing"))
	}

	m.DataSet = src
	return nil
}
