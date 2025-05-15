package postprocessor

import (
	"context"
	"github.com/tidepool-org/platform/data"
	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/summary"
	"github.com/tidepool-org/platform/summary/types"
)

// AddDataPlanner creates a post-processing plan for triggering summary recalculation after data is ingested
type AddDataPlanner struct {
	dataSet upload.Upload
}

func NewAddDataPlanner(dataSet upload.Upload) (*AddDataPlanner, error) {
	if dataSet.ID == nil {
		return nil, errors.New("data set id is missing")
	}
	if dataSet.UserID == nil {
		return nil, errors.New("user id is missing")
	}
	return &AddDataPlanner{dataSet: dataSet}, nil
}

func (f *AddDataPlanner) Plan(ctx context.Context, deviceData data.Data) (data.PostProcessPlan, error) {
	if f.dataSet.UserID == nil {
		return nil, errors.New("data set user id is missing")
	}

	updateSummaryTypes := summary.GetUpdatedSummaryTypes(deviceData)

	return []data.PostProcessAction{
		&SetOutdatedSummary{
			UpdateSummaryTypes: updateSummaryTypes,
			UserID:             *f.dataSet.UserID,
			Reason:             types.OutdatedReasonDataAdded,
		},
	}, nil
}

// CloseDataSetPlanner creates a post-processing plan for triggering summary recalculation after data set is closed
type CloseDataSetPlanner struct {
	dataSet upload.Upload
}

func NewCloseDataSetPlanner(dataSet upload.Upload) (*CloseDataSetPlanner, error) {
	if dataSet.ID == nil {
		return nil, errors.New("data set id is missing")
	}
	if dataSet.UserID == nil {
		return nil, errors.New("user id is missing")
	}
	return &CloseDataSetPlanner{dataSet: dataSet}, nil
}

func (f *CloseDataSetPlanner) Plan(ctx context.Context, deviceData data.Data) (data.PostProcessPlan, error) {
	if f.dataSet.UserID == nil {
		return nil, errors.New("data set user id is missing")
	}

	updateSummaryTypes := map[string]struct{}{}
	for _, typ := range types.AllSummaryTypes {
		updateSummaryTypes[typ] = struct{}{}
	}

	return []data.PostProcessAction{
		&SetOutdatedSummary{
			UpdateSummaryTypes: updateSummaryTypes,
			UserID:             *f.dataSet.UserID,
			Reason:             types.OutdatedReasonUploadCompleted,
		},
	}, nil
}

type SetOutdatedSummary struct {
	UpdateSummaryTypes map[string]struct{}
	UserID             string
	Reason             string
}

func (s *SetOutdatedSummary) Run(ctx context.Context, dataServiceContext dataService.Context) error {
	summary.MaybeUpdateSummary(ctx, dataServiceContext.SummarizerRegistry(), s.UpdateSummaryTypes, s.UserID, s.Reason)
	return nil
}
