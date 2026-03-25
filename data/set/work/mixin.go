package work

import (
	"context"

	"github.com/tidepool-org/platform/data"
	dataSet "github.com/tidepool-org/platform/data/set"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/work"
)

//go:generate mockgen -source=mixin.go -destination=test/mixin_mocks.go -package=test -typed

const MetadataKeyDataSetID = "dataSetId"

type Metadata struct {
	DataSetID *string `json:"dataSetId,omitempty" bson:"dataSetId,omitempty"`
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	m.DataSetID = parser.String(MetadataKeyDataSetID)
}

func (m *Metadata) Validate(validator structure.Validator) {
	validator.String(MetadataKeyDataSetID, m.DataSetID).Using(data.SetIDValidator)
}

type Mixin interface {
	DataSetClient() dataSet.Client

	HasDataSet() bool
	DataSet() *data.DataSet
	SetDataSet(dataSet *data.DataSet) *work.ProcessResult

	FetchDataSet(dataSetID string) *work.ProcessResult
	UpdateDataSet(dataSetUpdate *data.DataSetUpdate) *work.ProcessResult

	AddDataSetToContext()
}

type MixinFromWork interface {
	Mixin

	HasWorkMetadata() bool

	FetchDataSetFromWorkMetadata() *work.ProcessResult
	UpdateWorkMetadataFromDataSet() *work.ProcessResult
}

func NewMixin(provider work.Provider, dataSetClient dataSet.Client) (Mixin, error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}
	if dataSetClient == nil {
		return nil, errors.New("data set client is missing")
	}
	return &mixin{
		Provider:      provider,
		dataSetClient: dataSetClient,
	}, nil
}

func NewMixinFromWork(provider work.Provider, dataSetClient dataSet.Client, workMetadata *Metadata) (MixinFromWork, error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}
	if dataSetClient == nil {
		return nil, errors.New("data set client is missing")
	}
	if workMetadata == nil {
		return nil, errors.New("work metadata is missing")
	}
	return &mixin{
		Provider:      provider,
		dataSetClient: dataSetClient,
		workMetadata:  workMetadata,
	}, nil
}

type mixin struct {
	work.Provider
	dataSetClient dataSet.Client
	dataSet       *data.DataSet
	workMetadata  *Metadata
}

func (m *mixin) DataSetClient() dataSet.Client {
	return m.dataSetClient
}

func (m *mixin) HasDataSet() bool {
	return m.dataSet != nil
}

func (m *mixin) DataSet() *data.DataSet {
	return m.dataSet
}

func (m *mixin) SetDataSet(dataSet *data.DataSet) *work.ProcessResult {
	m.dataSet = dataSet
	m.AddDataSetToContext()
	return nil
}

func (m *mixin) FetchDataSet(dataSetID string) *work.ProcessResult {
	if dataSt, err := m.dataSetClient.GetDataSet(m.Context(), dataSetID); err != nil {
		return m.Failing(errors.Wrap(err, "unable to get data set"))
	} else if dataSt == nil {
		return m.Failed(errors.New("data set is missing"))
	} else {
		return m.SetDataSet(dataSt)
	}
}

func (m *mixin) UpdateDataSet(dataSetUpdate *data.DataSetUpdate) *work.ProcessResult {
	if dataSetUpdate == nil {
		return m.Failed(errors.New("data set update is missing"))
	}
	if m.dataSet == nil {
		return m.Failed(errors.New("data set is missing"))
	} else if m.dataSet.ID == nil {
		return m.Failed(errors.New("data set id is missing"))
	}

	if dataSt, err := m.dataSetClient.UpdateDataSet(context.WithoutCancel(m.Context()), *m.dataSet.ID, dataSetUpdate); err != nil {
		return m.Failing(errors.Wrap(err, "unable to update data set"))
	} else if dataSt == nil {
		return m.Failed(errors.New("data set is missing"))
	} else {
		return m.SetDataSet(dataSt)
	}
}

func (m *mixin) HasWorkMetadata() bool {
	return m.workMetadata != nil
}

func (m *mixin) FetchDataSetFromWorkMetadata() *work.ProcessResult {
	if m.workMetadata == nil {
		return m.Failed(errors.New("work metadata is missing"))
	} else if m.workMetadata.DataSetID == nil {
		return m.Failed(errors.New("work metadata data set id is missing"))
	} else {
		return m.FetchDataSet(*m.workMetadata.DataSetID)
	}
}

func (m *mixin) UpdateWorkMetadataFromDataSet() *work.ProcessResult {
	if m.dataSet == nil {
		return m.Failed(errors.New("data set is missing"))
	} else if m.dataSet.ID == nil {
		return m.Failed(errors.New("data set id is missing"))
	} else if m.workMetadata == nil {
		return m.Failed(errors.New("work metadata is missing"))
	}
	m.workMetadata.DataSetID = m.dataSet.ID
	return nil
}

func (m *mixin) AddDataSetToContext() {
	m.AddFieldToContext("dataSet", dataSetToFields(m.dataSet))
}

func dataSetToFields(dataSet *data.DataSet) log.Fields {
	if dataSet == nil {
		return nil
	}
	return log.Fields{
		"id":     dataSet.ID,
		"userId": dataSet.UserID,
	}
}
