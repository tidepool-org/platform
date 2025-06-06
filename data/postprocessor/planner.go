package postprocessor

import (
	"context"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/summary/postprocessor"
)

// AddDataPlanner creates a post-processing plan for updating associated resources after data has been ingested
type AddDataPlanner struct {
	dataSet    upload.Upload
	dataSource *source.Source
}

func NewAddDataPlanner(dataSet upload.Upload, dataSource *source.Source) *AddDataPlanner {
	return &AddDataPlanner{
		dataSet:    dataSet,
		dataSource: dataSource,
	}
}

func (p *AddDataPlanner) Plan(ctx context.Context, dataSetData data.Data) (data.PostProcessPlan, error) {
	planFns := make([]func(context.Context, data.Data) (data.PostProcessPlan, error), 0)

	summaryAddDataSetDataPlanner, err := postprocessor.NewAddDataPlanner(p.dataSet)
	if err != nil {
		return nil, err
	}
	planFns = append(planFns, summaryAddDataSetDataPlanner.Plan)

	dataSetPlanner, err := NewAddDataSetDataPlanner(p.dataSet, p.dataSource)
	if err != nil {
		return nil, err
	}
	planFns = append(planFns, dataSetPlanner.Plan)

	if p.dataSource != nil {
		dataSourcePlanner, err := NewAddDataSourceDataPlanner(*p.dataSource)
		if err != nil {
			return nil, err
		}

		planFns = append(planFns, dataSourcePlanner.Plan)
	}

	var plan data.PostProcessPlan
	for _, fn := range planFns {
		partialPlan, err := fn(ctx, dataSetData)
		if err != nil {
			return nil, err
		}
		plan = append(plan, partialPlan...)
	}

	return plan, nil
}

// CloseDataSetPlanner creates a post-processing plan when a data set is closed
type CloseDataSetPlanner struct {
	dataSet upload.Upload
}

func NewCloseDataSetPlanner(dataSet upload.Upload) *CloseDataSetPlanner {
	return &CloseDataSetPlanner{
		dataSet: dataSet,
	}
}

func (p *CloseDataSetPlanner) Plan(ctx context.Context, deviceData data.Data) (data.PostProcessPlan, error) {
	summaryCloseDataSetPlanner, err := postprocessor.NewCloseDataSetPlanner(p.dataSet)
	if err != nil {
		return nil, err
	}

	return summaryCloseDataSetPlanner.Plan(ctx, deviceData)
}
