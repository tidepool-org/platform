package service

import (
	"context"
	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/summary"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/page"
)

type Client struct {
	dataStore dataStore.Store
}

func NewClient(strDEPRECATED dataStore.Store) (*Client, error) {
	if strDEPRECATED == nil {
		return nil, errors.New("data store deprecated is missing")
	}

	return &Client{
		dataStore: strDEPRECATED,
	}, nil
}

func (c *Client) CreateUserDataSet(ctx context.Context, userID string, create *data.DataSetCreate) (*data.DataSet, error) {
	panic("Not Implemented!")
}

func (c *Client) ListUserDataSets(ctx context.Context, userID string, filter *data.DataSetFilter, pagination *page.Pagination) (data.DataSets, error) {
	repository := c.dataStore.NewDataRepository()
	return repository.ListUserDataSets(ctx, userID, filter, pagination)
}

func (c *Client) GetDataSet(ctx context.Context, id string) (*data.DataSet, error) {
	repository := c.dataStore.NewDataRepository()
	return repository.GetDataSet(ctx, id)
}

func (c *Client) BackfillSummaries(ctx context.Context) (int, error) {
	var empty struct{}
	var userIDsReqBackfill []string
	var count = 0

	summaryRepository := c.dataStore.NewSummaryRepository()
	dataRepository := c.dataStore.NewDataRepository()

	distinctDataUserIDs, err := dataRepository.DistinctUserIDs(ctx)
	if err != nil {
		return count, err
	}

	distinctSummaryIDs, err := summaryRepository.DistinctSummaryIDs(ctx)
	if err != nil {
		return count, err
	}

	distinctSummaryIDMap := make(map[string]struct{})
	for _, v := range distinctSummaryIDs {
		distinctSummaryIDMap[v] = empty
	}

	for _, userID := range distinctDataUserIDs {
		if _, exists := distinctSummaryIDMap[userID]; exists {
		} else {
			userIDsReqBackfill = append(userIDsReqBackfill, userID)
		}

		if len(userIDsReqBackfill) >= backfillBatch {
			break
		}
	}

	// todo move following line somewhere else
	types := []string{"cgm", "bgm"}

	var summaries = make([]*summary.Summary, len(userIDsReqBackfill)*2)
	for i, userID := range userIDsReqBackfill {
		for j, typ := range types {
			summaries[i*len(types)+j] = summary.NewSummary(userID, typ)
		}
	}

	if len(summaries) > 0 {
		count, err = summaryRepository.CreateSummaries(ctx, summaries)
		if err != nil {
			return count, err
		}
	}

	return count, nil
}

func (c *Client) GetOutdatedUserIDs(ctx context.Context, pagination *page.Pagination) ([][]string, error) {
	summaryRepository := c.dataStore.NewSummaryRepository()

	return summaryRepository.GetOutdatedUserIDs(ctx, pagination)
}

func (c *Client) UpdateDataSet(ctx context.Context, id string, update *data.DataSetUpdate) (*data.DataSet, error) {
	panic("Not Implemented!")
}

func (c *Client) DeleteDataSet(ctx context.Context, id string) error {
	panic("Not Implemented!")
}

func (c *Client) CreateDataSetsData(ctx context.Context, dataSetID string, datumArray []data.Datum) error {
	panic("Not Implemented!")
}

func (c *Client) DestroyDataForUserByID(ctx context.Context, userID string) error {
	panic("Not Implemented!")
}
