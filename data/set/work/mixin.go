package work

import (
	"github.com/tidepool-org/platform/data"
	dataSet "github.com/tidepool-org/platform/data/set"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/pointer"
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
	EnsureDataSet() *work.ProcessResult
	DataSet() *data.DataSet
	SetDataSet(dataSet *data.DataSet) *work.ProcessResult
	ClearDataSet()

	FetchDataSet(dataSetID string) *work.ProcessResult
	CreateDataSet(userID string, dataSetCreate *data.DataSetCreate) *work.ProcessResult
	UpdateDataSet(dataSetUpdate *data.DataSetUpdate) *work.ProcessResult

	AddDataSetToContext()
}

type FromWork interface {
	HasWorkMetadata() bool

	FetchDataSetFromWorkMetadata() *work.ProcessResult
	UpdateWorkMetadataFromDataSet() *work.ProcessResult
}

type WithParsedMetadata[M any] interface {
	HasDataSetMetadata() bool
	DataSetMetadata() *M
	SetDataSetMetadata(dataSetMetadata *M) *work.ProcessResult

	UpdateDataSetMetadata() *work.ProcessResult
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

func NewMixin(provider work.Provider, dataSetClient dataSet.Client) (Mixin, error) {
	return NewMixinWithParsedMetadata[map[string]any](provider, dataSetClient)
}

func NewMixinFromWork(provider work.Provider, dataSetClient dataSet.Client, workMetadata *Metadata) (MixinFromWork, error) {
	return NewMixinFromWorkWithParsedMetadata[map[string]any](provider, dataSetClient, workMetadata)
}

func NewMixinWithParsedMetadata[M any](provider work.Provider, dataSetClient dataSet.Client) (MixinWithParsedMetadata[M], error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}
	if dataSetClient == nil {
		return nil, errors.New("data set client is missing")
	}
	return &mixin[M]{
		Provider:      provider,
		dataSetClient: dataSetClient,
	}, nil
}

func NewMixinFromWorkWithParsedMetadata[M any](provider work.Provider, dataSetClient dataSet.Client, workMetadata *Metadata) (MixinFromWorkWithParsedMetadata[M], error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}
	if dataSetClient == nil {
		return nil, errors.New("data set client is missing")
	}
	if workMetadata == nil {
		return nil, errors.New("work metadata is missing")
	}
	return &mixin[M]{
		Provider:      provider,
		dataSetClient: dataSetClient,
		workMetadata:  workMetadata,
	}, nil
}

type mixin[M any] struct {
	work.Provider
	dataSetClient   dataSet.Client
	dataSet         *data.DataSet
	dataSetMetadata *M
	workMetadata    *Metadata
}

func (m *mixin[M]) DataSetClient() dataSet.Client {
	return m.dataSetClient
}

func (m *mixin[M]) HasDataSet() bool {
	return m.dataSet != nil
}

func (m *mixin[M]) EnsureDataSet() *work.ProcessResult {
	if m.dataSet == nil {
		return m.Failed(errors.New("data set is missing"))
	} else {
		return nil
	}
}

func (m *mixin[M]) DataSet() *data.DataSet {
	return m.dataSet
}

func (m *mixin[M]) SetDataSet(dataSt *data.DataSet) *work.ProcessResult {
	var dataStMetadata *M
	if dataSt != nil {
		var err error
		if dataStMetadata, err = metadata.Decode[M](m.Context(), dataSt.Metadata); err != nil {
			return m.Failed(errors.Wrap(err, "unable to decode data set metadata"))
		}
	}
	m.dataSet = dataSt
	m.dataSetMetadata = dataStMetadata
	m.AddDataSetToContext()
	return nil
}

func (m *mixin[M]) ClearDataSet() {
	m.dataSet = nil
	m.dataSetMetadata = nil
	m.AddDataSetToContext()
}

func (m *mixin[M]) HasDataSetMetadata() bool {
	return m.dataSetMetadata != nil
}

func (m *mixin[M]) DataSetMetadata() *M {
	return m.dataSetMetadata
}

func (m *mixin[M]) SetDataSetMetadata(dataSetMetadata *M) *work.ProcessResult {
	m.dataSetMetadata = dataSetMetadata
	m.AddDataSetToContext()
	return nil
}

func (m *mixin[M]) FetchDataSet(dataSetID string) *work.ProcessResult {
	if dataSt, err := m.dataSetClient.GetDataSet(m.Context(), dataSetID); err != nil {
		return m.Failing(errors.Wrap(err, "unable to get data set"))
	} else if dataSt == nil {
		return m.Failed(errors.New("data set is missing"))
	} else {
		return m.SetDataSet(dataSt)
	}
}

func (m *mixin[M]) CreateDataSet(userID string, dataSetCreate *data.DataSetCreate) *work.ProcessResult {
	if m.dataSet != nil {
		return m.Failed(errors.New("data set already exists"))
	}

	// The work's context is passed as is so the create cannot outlive the work's deadline
	if dataSt, err := m.dataSetClient.CreateUserDataSet(m.Context(), userID, dataSetCreate); err != nil {
		return m.Failing(errors.Wrap(err, "unable to create data set"))
	} else if dataSt == nil {
		return m.Failed(errors.New("data set is missing"))
	} else if result := m.SetDataSet(dataSt); result != nil {
		return result
	}

	log.LoggerFromContext(m.Context()).Debug("created data set")
	return nil
}

// UpdateDataSet writes the update with the parsed metadata, conditioned on the version of the data set as read
// unless the update carries its own condition; a concurrent write of the data set is a conflict.
func (m *mixin[M]) UpdateDataSet(dataSetUpdate *data.DataSetUpdate) *work.ProcessResult {
	if dataSetUpdate == nil {
		return m.Failed(errors.New("data set update is missing"))
	}
	if m.dataSet == nil {
		return m.Failed(errors.New("data set is missing"))
	} else if m.dataSet.ID == nil {
		return m.Failed(errors.New("data set id is missing"))
	}

	if dataStMetadata, err := metadata.Encode(m.dataSetMetadata); err != nil {
		return m.Failed(errors.Wrap(err, "unable to encode data set metadata"))
	} else if dataStMetadata != nil || m.dataSet.Metadata != nil {
		dataSetUpdate.Metadata = &dataStMetadata
	}
	if dataSetUpdate.Condition == nil {
		dataSetUpdate.Condition = data.NewDataSetConditionWithVersionInternal(pointer.FromInt(m.dataSet.VersionInternal))
	}

	// The work's context is passed as is so the update cannot outlive the work's deadline
	if dataSt, err := m.dataSetClient.UpdateDataSet(m.Context(), *m.dataSet.ID, dataSetUpdate); err != nil {
		return m.Failing(errors.Wrap(err, "unable to update data set"))
	} else if dataSt == nil {
		return m.Failing(errors.Wrap(data.ErrorDataSetUpdateConflict(), "unable to update data set"))
	} else {
		return m.SetDataSet(dataSt)
	}
}

func (m *mixin[M]) UpdateDataSetMetadata() *work.ProcessResult {
	return m.UpdateDataSet(&data.DataSetUpdate{})
}

func (m *mixin[M]) HasWorkMetadata() bool {
	return m.workMetadata != nil
}

func (m *mixin[M]) FetchDataSetFromWorkMetadata() *work.ProcessResult {
	if m.workMetadata == nil {
		return m.Failed(errors.New("work metadata is missing"))
	} else if m.workMetadata.DataSetID == nil {
		return m.Failed(errors.New("work metadata data set id is missing"))
	} else {
		return m.FetchDataSet(*m.workMetadata.DataSetID)
	}
}

func (m *mixin[M]) UpdateWorkMetadataFromDataSet() *work.ProcessResult {
	if m.dataSet == nil {
		return m.Failed(errors.New("data set is missing"))
	} else if m.dataSet.ID == nil {
		return m.Failed(errors.New("data set id is missing"))
	} else if m.workMetadata == nil {
		return m.Failed(errors.New("work metadata is missing"))
	}
	m.workMetadata.DataSetID = pointer.Clone(m.dataSet.ID)
	return nil
}

func (m *mixin[M]) AddDataSetToContext() {
	m.AddFieldsToContext(log.Fields{"dataSet": dataSetToFields(m.dataSet), "dataSetMetadata": m.dataSetMetadata})
}

func dataSetToFields(dataSet *data.DataSet) log.Fields {
	if dataSet == nil {
		return nil
	}
	return log.Fields{
		"id":       dataSet.ID,
		"userId":   dataSet.UserID,
		"metadata": dataSet.Metadata,
		"version":  dataSet.VersionInternal,
	}
}
