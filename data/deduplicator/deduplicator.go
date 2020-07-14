package deduplicator

import (
	"context"

	"github.com/tidepool-org/platform/data"
	dataStoreDEPRECATED "github.com/tidepool-org/platform/data/storeDEPRECATED"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
)

type Factory interface {
	New(dataSet *dataTypesUpload.Upload) (Deduplicator, error)
	Get(dataSet *dataTypesUpload.Upload) (Deduplicator, error)
}

type Deduplicator interface {
	Open(ctx context.Context, session dataStoreDEPRECATED.DataRepository, dataSet *dataTypesUpload.Upload) (*dataTypesUpload.Upload, error)
	AddData(ctx context.Context, session dataStoreDEPRECATED.DataRepository, dataSet *dataTypesUpload.Upload, dataSetData data.Data) error
	DeleteData(ctx context.Context, session dataStoreDEPRECATED.DataRepository, dataSet *dataTypesUpload.Upload, selectors *data.Selectors) error
	Close(ctx context.Context, session dataStoreDEPRECATED.DataRepository, dataSet *dataTypesUpload.Upload) error
	Delete(ctx context.Context, session dataStoreDEPRECATED.DataRepository, dataSet *dataTypesUpload.Upload) error
}
