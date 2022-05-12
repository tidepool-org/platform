package service

import (
	"context"

	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/blood/glucose/summary"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
)

const (
	backfillBatch    = 10000
	lowBloodGlucose  = 3.9
	highBloodGlucose = 10
	units            = "mmol/l"
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

func (c *Client) GetSummary(ctx context.Context, id string) (*summary.Summary, error) {
	summaryRepository := c.dataStore.NewSummaryRepository()

	userSummary, err := summaryRepository.GetSummary(ctx, id)
	if err != nil {
		return nil, err
	}

	return userSummary, err
}

func (c *Client) UpdateSummary(ctx context.Context, id string) (*summary.Summary, error) {
	logger := log.LoggerFromContext(ctx)
	logger.Debugf("Starting summary calculation for %s", id)
	var err error
	var status *summary.UserLastUpdated
	summaryRepository := c.dataStore.NewSummaryRepository()
	dataRepository := c.dataStore.NewDataRepository()

	// we need the original summary object to grab the original for rolling calc
	userSummary, err := summaryRepository.GetSummary(ctx, id)
	if err != nil {
		return nil, err
	}

	status, err = dataRepository.GetLastUpdatedForUser(ctx, id)
	if err != nil {
		return nil, err
	}

	// remove 2 weeks for start time
	startTime := status.LastData.AddDate(0, 0, -14)

	// if summary already exists with a last data checkpoint, start data pull there
	if userSummary.LastData != nil {
		if startTime.Before(*userSummary.LastData) {
			startTime = *userSummary.LastData
		}
	}

	userData, err := dataRepository.GetCGMDataRange(ctx, id, startTime, status.LastData)
	if err != nil {
		return nil, err
	}

	newSummary, err := summary.Update(ctx, userSummary, status, userData)
	if err != nil {
		return nil, err
	}

	userSummary, err = summaryRepository.UpdateSummary(ctx, newSummary)

	return userSummary, err
}

func (c *Client) BackfillSummaries(ctx context.Context) (int64, error) {
	var empty struct{}
	userIDsReqBackfill := []string{}
	var count int64 = 0

	summaryRepository := c.dataStore.NewSummaryRepository()
	dataRepository := c.dataStore.NewDataRepository()

	distinctSummaryIDs, err := summaryRepository.DistinctSummaryIDs(ctx)
	if err != nil {
		return count, err
	}

	distinctDataUserIDs, err := dataRepository.DistinctCGMUserIDs(ctx)
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

	var summaries []*summary.Summary

	for _, userID := range userIDsReqBackfill {
		summaries = append(summaries, summary.New(userID))
	}

	if len(summaries) > 0 {
		count, err = summaryRepository.CreateSummaries(ctx, summaries)
		if err != nil {
			return count, err
		}
	}

	return count, nil
}

func (c *Client) GetOutdatedUserIDs(ctx context.Context, pagination *page.Pagination) ([]string, error) {
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
