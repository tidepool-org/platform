package load

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"time"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/work"
	workService "github.com/tidepool-org/platform/work/service"
)

type CoordinatorClient struct {
	coordinator    *workService.Coordinator
	workClient     work.Client
	groupWorkItems map[string][]*work.Work
	runFilePath    string
	runStart       time.Time
}

type Data struct {
	OffsetFromStart int64        `json:"offsetFromStart"`
	Create          *work.Create `json:"create"`
}

func NewCoordinatorClient(authClient auth.Client, workClient work.Client) (*CoordinatorClient, error) {
	logger := null.NewLogger()
	workCoordinator := &CoordinatorClient{
		workClient:     workClient,
		groupWorkItems: map[string][]*work.Work{},
	}
	coordinator, err := workService.NewCoordinator(logger, authClient, workCoordinator.workClient)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create work coordinator")
	}
	workCoordinator.coordinator = coordinator

	lp, err := NewLoadProcessors()
	if err != nil {
		return nil, errors.Wrap(err, "unable to load processors")
	}

	if err = workCoordinator.coordinator.RegisterProcessors(lp); err != nil {
		return nil, errors.Wrap(err, "unable to register abbott processors")
	}
	return workCoordinator, nil
}

func (c *CoordinatorClient) Run(ctx context.Context, runFilePath string) error {
	if runFilePath == "" {
		return errors.New("missing required run file")
	}
	c.coordinator.Start()
	c.runFilePath = runFilePath

	jsonFile, err := os.Open(runFilePath)
	if err != nil {
		return errors.Wrapf(err, "unable open run file %s", runFilePath)
	}
	defer jsonFile.Close()
	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		return errors.Wrapf(err, "read file %s", runFilePath)
	}
	var allData []Data
	json.Unmarshal(jsonData, &allData)

	c.runStart = time.Now()
	for _, data := range allData {
		data.Create.ProcessingAvailableTime = c.runStart.Add(time.Second * time.Duration(data.OffsetFromStart))
		workItem, err := c.workClient.Create(ctx, data.Create)
		if err != nil {
			return errors.Wrapf(err, "error creating work %v", data.Create)
		}
		c.groupWorkItems[*data.Create.GroupID] = append(c.groupWorkItems[*data.Create.GroupID], workItem)
	}
	return nil
}

func (c *CoordinatorClient) GetCreatedWork() map[string][]*work.Work {
	return c.groupWorkItems
}

func (c *CoordinatorClient) CleanUp(ctx context.Context) error {
	log.Printf("cleanup for run %s", c.runFilePath)
	for groupID := range c.groupWorkItems {
		count, err := c.workClient.DeleteAllByGroupID(ctx, groupID)
		if err != nil {
			return errors.Wrapf(err, "unable to delete work items for groupId %s", groupID)
		}
		log.Printf("cleanup removed %d items for groupId %s", count, groupID)
	}
	c.coordinator.Stop()
	return nil
}
