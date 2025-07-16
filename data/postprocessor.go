package data

import (
	"context"
	dataService "github.com/tidepool-org/platform/data/service"
)

type PostProcessPlan []PostProcessAction

func (p PostProcessPlan) Run(ctx context.Context, dataServiceContext dataService.Context) error {
	for _, action := range p {
		if err := action.Run(ctx, dataServiceContext); err != nil {
			return err
		}
	}
	return nil
}

type PostProcessAction interface {
	Run(context.Context, dataService.Context) error
}
