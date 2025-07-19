package deduplicator

import (
	"context"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/net"
	"github.com/tidepool-org/platform/pointer"
)

type DataSetStore interface {
	UpdateDataSet(ctx context.Context, id string, update *data.DataSetUpdate) (*data.DataSet, error)
	DeleteDataSet(ctx context.Context, dataSet *data.DataSet) error
}

type DataStore interface {
	CreateDataSetData(ctx context.Context, dataSet *data.DataSet, dataSetData []data.Datum) error
	ExistingDataSetData(ctx context.Context, dataSet *data.DataSet, selectors *data.Selectors) (*data.Selectors, error)
	ActivateDataSetData(ctx context.Context, dataSet *data.DataSet, selectors *data.Selectors) error
	ArchiveDataSetData(ctx context.Context, dataSet *data.DataSet, selectors *data.Selectors) error
	DeleteDataSetData(ctx context.Context, dataSet *data.DataSet, selectors *data.Selectors) error
	DestroyDeletedDataSetData(ctx context.Context, dataSet *data.DataSet, selectors *data.Selectors) error
	DestroyDataSetData(ctx context.Context, dataSet *data.DataSet, selectors *data.Selectors) error

	ArchiveDeviceDataUsingHashesFromDataSet(ctx context.Context, dataSet *data.DataSet) error
	UnarchiveDeviceDataUsingHashesFromDataSet(ctx context.Context, dataSet *data.DataSet) error
	DeleteOtherDataSetData(ctx context.Context, dataSet *data.DataSet) error
}

type Dependencies struct {
	DataSetStore DataSetStore
	DataStore    DataStore
}

func (d Dependencies) Validate() error {
	if d.DataSetStore == nil {
		return errors.New("data set store is missing")
	}
	if d.DataStore == nil {
		return errors.New("data store is missing")
	}
	return nil
}

type Base struct {
	Dependencies
	name    string
	version string
}

func NewBase(dependencies Dependencies, name string, version string) (*Base, error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}
	if name == "" {
		return nil, errors.New("name is missing")
	} else if !net.IsValidReverseDomain(name) {
		return nil, errors.New("name is invalid")
	}
	if version == "" {
		return nil, errors.New("version is missing")
	} else if !net.IsValidSemanticVersion(version) {
		return nil, errors.New("version is invalid")
	}

	return &Base{
		Dependencies: dependencies,
		name:         name,
		version:      version,
	}, nil
}

func (b *Base) New(ctx context.Context, dataSet *data.DataSet) (bool, error) {
	return b.Get(ctx, dataSet)
}

func (b *Base) Get(ctx context.Context, dataSet *data.DataSet) (bool, error) {
	if dataSet == nil {
		return false, errors.New("data set is missing")
	}

	return dataSet.HasDeduplicatorNameMatch(b.name), nil
}

func (b *Base) Open(ctx context.Context, dataSet *data.DataSet) (*data.DataSet, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if dataSet == nil {
		return nil, errors.New("data set is missing")
	}

	if dataSet.HasDataSetTypeContinuous() {
		dataSet.Active = true
	}

	update := data.NewDataSetUpdate()
	update.Active = pointer.FromBool(dataSet.Active)
	update.Deduplicator = data.NewDeduplicatorDescriptor()
	update.Deduplicator.Name = pointer.FromString(b.name)
	update.Deduplicator.Version = pointer.FromString(b.version)
	return b.DataSetStore.UpdateDataSet(ctx, *dataSet.UploadID, update)
}

func (b *Base) AddData(ctx context.Context, dataSet *data.DataSet, dataSetData data.Data) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if dataSet == nil {
		return errors.New("data set is missing")
	}
	if dataSetData == nil {
		return errors.New("data set data is missing")
	}

	if dataSet.HasDataSetTypeContinuous() {
		dataSetData.SetActive(true)
	}

	return b.DataStore.CreateDataSetData(ctx, dataSet, dataSetData)
}

func (b *Base) DeleteData(ctx context.Context, dataSet *data.DataSet, selectors *data.Selectors) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if dataSet == nil {
		return errors.New("data set is missing")
	}
	if selectors == nil {
		return errors.New("selectors is missing")
	}

	return b.DataStore.DestroyDataSetData(ctx, dataSet, selectors)
}

func (b *Base) Close(ctx context.Context, dataSet *data.DataSet) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if dataSet == nil {
		return errors.New("data set is missing")
	}

	if dataSet.HasDataSetTypeContinuous() {
		return nil
	}

	update := data.NewDataSetUpdate()
	update.Active = pointer.FromBool(true)
	if _, err := b.DataSetStore.UpdateDataSet(ctx, *dataSet.UploadID, update); err != nil {
		return err
	}

	return b.DataStore.ActivateDataSetData(ctx, dataSet, nil)
}

func (b *Base) Delete(ctx context.Context, dataSet *data.DataSet) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if dataSet == nil {
		return errors.New("data set is missing")
	}

	return b.DataSetStore.DeleteDataSet(ctx, dataSet)
}

func MapDataSetDataToSelectors(dataSetData data.Data, mapper func(datum data.Datum) *data.Selector) *data.Selectors {
	if len(dataSetData) == 0 {
		return nil
	}
	selectors := make(data.Selectors, len(dataSetData))
	for index, dataSetDatum := range dataSetData {
		selectors[index] = mapper(dataSetDatum)
	}
	return &selectors
}
