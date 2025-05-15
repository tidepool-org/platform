package postprocessor

import (
	"context"
	"github.com/tidepool-org/platform/data"
	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
	"time"
)

// AddDataSourceDataPostProcessPlanner produces a plan to update a data source
// after data has been added to an associated data set
type AddDataSourceDataPostProcessPlanner struct {
	dataSet    upload.Upload
	dataSource source.Source
}

func NewAddDataSourceDataPlanner(dataSource source.Source) (*AddDataSourceDataPostProcessPlanner, error) {
	if dataSource.ID == nil || *dataSource.ID == "" {
		return nil, errors.New("data source id is missing")
	}

	return &AddDataSourceDataPostProcessPlanner{
		dataSource: dataSource,
	}, nil
}

func (d *AddDataSourceDataPostProcessPlanner) Plan(ctx context.Context, deviceData data.Data) (data.PostProcessPlan, error) {
	var earliestDataTime *time.Time
	var latestDataTime *time.Time

	for _, datum := range deviceData {
		datumTime := datum.GetTime()
		if datumTime == nil {
			continue
		}

		if earliestDataTime == nil || (*datumTime).Before(*earliestDataTime) {
			earliestDataTime = datumTime
		}
		if latestDataTime == nil || (*datumTime).After(*latestDataTime) {
			latestDataTime = datumTime
		}
	}

	return []data.PostProcessAction{
		&SetDataSourceTimestamps{
			DataSourceID:     *d.dataSource.ID,
			LastImportTime:   time.Now(),
			EarliestDataTime: earliestDataTime,
			LatestDataTime:   latestDataTime,
		},
	}, nil
}

type SetDataSourceTimestamps struct {
	DataSourceID string

	EarliestDataTime *time.Time
	LatestDataTime   *time.Time

	LastImportTime time.Time
}

func (s *SetDataSourceTimestamps) Run(ctx context.Context, dataServiceContext dataService.Context) error {
	update := source.NewUpdate()
	update.EarliestDataTime = s.EarliestDataTime
	update.LatestDataTime = s.LatestDataTime
	update.LastImportTime = pointer.FromAny(s.LastImportTime)

	_, err := dataServiceContext.DataSourceClient().Update(ctx, s.DataSourceID, nil, update)
	return err
}
