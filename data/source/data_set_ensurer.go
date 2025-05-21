package source

import (
	"context"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/errors"
)

type DataSetEnsurerClient interface {
	CreateUserDataSet(ctx context.Context, userID string, create *data.DataSetCreate) (*data.DataSet, error)
	GetDataSet(ctx context.Context, id string) (*data.DataSet, error)
}

type DataSetEnsurerFactory interface {
	NewDataSetCreate(dataSrc Source) data.DataSetCreate
}

type DataSetEnsurer struct {
	Client  DataSetEnsurerClient
	Factory DataSetEnsurerFactory
}

func (d DataSetEnsurer) Ensure(ctx context.Context, dataSrc Source) (*data.DataSet, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	if d.Client == nil {
		return nil, errors.New("client is missing")
	}
	if d.Factory == nil {
		return nil, errors.New("factory is missing")
	}

	if dataSetIDs := dataSrc.DataSetIDs; dataSetIDs != nil {
		for _, dataSetID := range *dataSetIDs {
			if dataSet, err := d.Client.GetDataSet(ctx, dataSetID); err != nil {
				return nil, errors.Wrap(err, "unable to get data set")
			} else if dataSet != nil && dataSet.IsOpen() {
				return dataSet, nil
			}
		}
	}

	dataSetCreate := d.Factory.NewDataSetCreate(dataSrc)
	dataSet, err := d.Client.CreateUserDataSet(ctx, *dataSrc.UserID, &dataSetCreate)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create data set")
	}

	return dataSet, nil
}
