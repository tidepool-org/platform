package workTestLoad

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/errors"
	logNull "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/work"
	workLoad "github.com/tidepool-org/platform/work/load"
	workService "github.com/tidepool-org/platform/work/service"
)

type CoordinatorClient struct {
	coordinator *workService.Coordinator
	workClient  work.Client
	groupID     string
	createItems []*work.Create
	workItems   []*work.Work
	runStart    time.Time
}

func NewCoordinatorClient(authClient auth.Client, workClient work.Client) (*CoordinatorClient, error) {
	logger := logNull.NewLogger()
	cc := &CoordinatorClient{
		workClient:  workClient,
		workItems:   []*work.Work{},
		createItems: []*work.Create{},
		groupID:     data.NewID(),
	}
	coordinator, err := workService.NewCoordinator(logger, authClient, cc.workClient)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create work coordinator")
	}
	cc.coordinator = coordinator

	lp, err := workLoad.NewLoadProcessors(workClient, cc.coordinator.RegisterProcessor)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load processors")
	}

	if err = cc.coordinator.RegisterProcessors(lp); err != nil {
		return nil, errors.Wrap(err, "unable to register abbott processors")
	}
	return cc, nil
}

func (c *CoordinatorClient) Initialize(ctx context.Context, duplicateID *string, serializeID *string, items ...*work.Create) error {
	for _, create := range items {
		create.GroupID = &c.groupID
		create.DeduplicationID = duplicateID
		create.SerialID = serializeID
		c.createItems = append(c.createItems, create)
	}
	return nil
}

func (c *CoordinatorClient) Run(ctx context.Context) error {
	c.coordinator.Start()
	c.runStart = time.Now()
	for _, create := range c.GetCreate() {
		workItem, err := c.workClient.Create(ctx, create)
		if err != nil {
			return errors.Wrapf(err, "error creating work %v", create)
		}
		c.workItems = append(c.workItems, workItem)
	}
	return nil
}

func (c *CoordinatorClient) GetWorkGroupID() string {
	return c.groupID
}

func (c *CoordinatorClient) GetWork() []*work.Work {
	return c.workItems
}

func (c *CoordinatorClient) GetCreate() []*work.Create {
	return c.createItems
}

func (c *CoordinatorClient) CleanUp(ctx context.Context) error {
	_, err := c.workClient.DeleteAllByGroupID(ctx, c.groupID)
	if err != nil {
		return errors.Wrapf(err, "unable to delete work items for groupId %s", c.groupID)
	}
	c.coordinator.Stop()
	return nil
}
