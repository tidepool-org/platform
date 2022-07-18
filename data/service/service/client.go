package service

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/data/summary"

	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
)

const (
	backfillBatch = 100000
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
	return summaryRepository.GetSummary(ctx, id)
}

func (c *Client) UpdateSummary(ctx context.Context, id string) (*summary.Summary, error) {
	var err error
	var status *summary.UserLastUpdated
	var userSummary *summary.Summary
	var timestamp time.Time
	timestamp = time.Now().UTC()
	logger := log.LoggerFromContext(ctx)
	logger.Debugf("Starting summary calculation for %s", id)
	summaryRepository := c.dataStore.NewSummaryRepository()
	dataRepository := c.dataStore.NewDataRepository()

	// we need the original summary object to grab the original for rolling calc
	userSummary, err = summaryRepository.GetSummary(ctx, id)
	if err != nil {
		return nil, err
	}

	status, err = dataRepository.GetLastUpdatedForUser(ctx, id)
	if err != nil {
		return nil, err
	}

	if status == nil {
		if userSummary != nil {
			// user's data is inactive/deleted, or this summary shouldn't have been created
			logger.Warnf("User %s has an outdated summary with no data, skipping calc.", id)
			userSummary.OutdatedSince = nil
			userSummary.LastUpdatedDate = timestamp

			userSummary, err = summaryRepository.UpdateSummary(ctx, userSummary)
			if err != nil {
				return nil, err
			}
		}
		return userSummary, nil
	}

	// user exists (has relevant data), but no summary, create a blank one
	if userSummary == nil {
		userSummary = summary.New(id)
	}

	// remove 2 weeks for start time
	startTime := status.LastData.AddDate(0, 0, -14)

	// check status.LastData for going back in time to prevent deleted data from causing issues
	if userSummary.LastData != nil {
		if status.LastData.Before(*userSummary.LastData) || status.LastData.Equal(*userSummary.LastData) {
			userSummary.OutdatedSince = nil
			userSummary.LastUpdatedDate = timestamp

			userSummary, err = summaryRepository.UpdateSummary(ctx, userSummary)
			if err != nil {
				return nil, err
			}
		}

		// if summary already exists with a last data checkpoint, start data pull there
		if startTime.Before(*userSummary.LastData) {
			startTime = *userSummary.LastData
		}
	}

	userData, err := dataRepository.GetCGMDataRange(ctx, id, startTime, status.LastData)
	if err != nil {
		return nil, err
	}

	// if there is new data
	if len(userData) > 0 {
		err = userSummary.Update(ctx, status, userData)
		if err != nil {
			return nil, err
		}
	} else {
		// "new" data must be in the past, don't update, just remove flags and set new date
		logger.Infof("User %s has an outdated summary with no forward data, skipping calc.", id)
		userSummary.OutdatedSince = nil
		userSummary.LastUpdatedDate = timestamp
	}
	userSummary, err = summaryRepository.UpdateSummary(ctx, userSummary)

	return userSummary, err
}

func (c *Client) BackfillSummaries(ctx context.Context) (int, error) {
	var empty struct{}
	var userIDsReqBackfill []string
	var count = 0

	summaryRepository := c.dataStore.NewSummaryRepository()
	dataRepository := c.dataStore.NewDataRepository()

	distinctDataUserIDs, err := dataRepository.DistinctCGMUserIDs(ctx)
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

	var summaries = make([]*summary.Summary, len(userIDsReqBackfill))
	for i, userID := range userIDsReqBackfill {
		summaries[i] = summary.New(userID)
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
