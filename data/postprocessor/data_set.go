package postprocessor

import (
	"context"
	"github.com/tidepool-org/platform/data"
	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"time"
)

type AddDataSetDataPostProcessPlanner struct {
	dataSet    upload.Upload
	dataSource *source.Source
}

func NewAddDataSetDataPlanner(dataSet upload.Upload, dataSource *source.Source) (*AddDataSetDataPostProcessPlanner, error) {
	if dataSet.ID == nil || *dataSet.ID == "" {
		return nil, errors.New("data set id is missing")
	}
	return &AddDataSetDataPostProcessPlanner{
		dataSet:    dataSet,
		dataSource: dataSource,
	}, nil
}

func (d *AddDataSetDataPostProcessPlanner) Plan(ctx context.Context, deviceData data.Data) (data.PostProcessPlan, error) {
	lgr := log.LoggerFromContext(ctx)

	if d.dataSource == nil {
		lgr.Info("updating the time zone offset of data set is required only when a data source exists, skipping")
		return nil, nil
	}

	var latestDataTime *time.Time
	var timeZoneOffset *int

	for _, datum := range deviceData {
		datumTime := datum.GetTime()
		if datumTime == nil {
			continue
		}

		if latestDataTime == nil || (*datumTime).After(*latestDataTime) {
			latestDataTime = datumTime
			if offset := datum.GetTimeZoneOffset(); offset != nil {
				timeZoneOffset = offset
			}
		}
	}

	if timeZoneOffset == nil {
		return nil, nil
	}

	return []data.PostProcessAction{
		&SetDataSetTimezoneOffset{
			DataSetID:      *d.dataSet.ID,
			TimezoneOffset: *timeZoneOffset,
		},
	}, nil
}

type SetDataSetTimezoneOffset struct {
	DataSetID string

	TimezoneOffset int
}

func (s *SetDataSetTimezoneOffset) Run(ctx context.Context, dataServiceContext dataService.Context) error {
	update := data.NewDataSetUpdate()
	update.TimeZoneOffset = &s.TimezoneOffset

	_, err := dataServiceContext.DataRepository().UpdateDataSet(ctx, s.DataSetID, update)
	return err
}
