package service

import (
	"context"
	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
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

func (c *Client) GetCGMSummary(ctx context.Context, id string) (*summary.CGMSummary, error) {
	summaryRepository := c.dataStore.NewSummaryRepository()
	return summaryRepository.GetCGMSummary(ctx, id)
}

func (c *Client) GetBGMSummary(ctx context.Context, id string) (*summary.BGMSummary, error) {
	summaryRepository := c.dataStore.NewSummaryRepository()
	return summaryRepository.GetBGMSummary(ctx, id)
}

func (c *Client) UpdateCGMSummary(ctx context.Context, id string) (*summary.CGMSummary, error) {
	var err error
	var status *summary.UserLastUpdated
	var userSummary *summary.CGMSummary
	var timestamp time.Time
	var userData []*glucoseDatum.Glucose
	timestamp = time.Now().UTC()
	logger := log.LoggerFromContext(ctx)
	logger.Debugf("Starting summary calculation for %s", id)
	summaryRepository := c.dataStore.NewSummaryRepository()
	dataRepository := c.dataStore.NewDataRepository()

	// we need the original summary object to grab the original for rolling calc
	userSummary, err = summaryRepository.GetCGMSummary(ctx, id)
	if err != nil {
		return nil, err
	}

	status, err = dataRepository.GetLastUpdatedForUser(ctx, id, "cgm")
	if err != nil {
		return nil, err
	}

	// this filters out users which require no update, as they have no cgm data, but have an outdated summary
	if status.LastData.IsZero() {
		if userSummary != nil {
			// user's data is inactive/deleted, or this summary shouldn't have been created
			logger.Warnf("User %s has an outdated summary with no data, skipping calc.", id)
			userSummary, err = summaryRepository.UpdateCGMSummary(ctx, userSummary)
			if err != nil {
				return nil, err
			}
		}
		return userSummary, nil
	}

	// user exists (has relevant data), but no summary, create a blank one
	if userSummary == nil {
		userSummary = summary.NewCGMSummary(id)
	}

	if !status.LastData.IsZero() {
		// remove 30 days for start time
		startTime := status.LastData.AddDate(0, 0, -30)
		endTime := status.LastData

		if userSummary.LastData != nil {
			// if summary already exists with a last data checkpoint, start data pull there
			if startTime.Before(*userSummary.LastData) {
				startTime = *userSummary.LastData
			}

			// ensure endTime does not move backwards by capping it at summary LastData
			if !status.LastData.After(*userSummary.LastData) {
				endTime = *userSummary.LastData
			}
		}

		userData, err = dataRepository.GetDataRange(ctx, id, "cgm", startTime, endTime)
		if err != nil {
			return nil, err
		}
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
	userSummary, err = summaryRepository.UpdateCGMSummary(ctx, userSummary)

	return userSummary, err
}

func (c *Client) UpdateBGMSummary(ctx context.Context, id string) (*summary.BGMSummary, error) {
	var err error
	var status *summary.UserLastUpdated
	var userSummary *summary.BGMSummary
	var timestamp time.Time
	var userData []*glucoseDatum.Glucose
	timestamp = time.Now().UTC()
	logger := log.LoggerFromContext(ctx)
	logger.Debugf("Starting summary calculation for %s", id)
	summaryRepository := c.dataStore.NewSummaryRepository()
	dataRepository := c.dataStore.NewDataRepository()

	// we need the original summary object to grab the original for rolling calc
	userSummary, err = summaryRepository.GetBGMSummary(ctx, id)
	if err != nil {
		return nil, err
	}

	status, err = dataRepository.GetLastUpdatedForUser(ctx, id, "bgm")
	if err != nil {
		return nil, err
	}

	// this filters out users which require no update, as they have no bgm data, but have an outdated summary
	if status.LastData.IsZero() {
		if userSummary != nil {
			// user's data is inactive/deleted, or this summary shouldn't have been created
			logger.Warnf("User %s has an outdated summary with no data, skipping calc.", id)
			userSummary, err = summaryRepository.UpdateBGMSummary(ctx, userSummary)
			if err != nil {
				return nil, err
			}
		}
		return userSummary, nil
	}

	// user exists (has relevant data), but no summary, create a blank one
	if userSummary == nil {
		userSummary = summary.NewBGMSummary(id)
	}

	if !status.LastData.IsZero() {
		// remove 30 days for start time
		startTime := status.LastData.AddDate(0, 0, -30)
		endTime := status.LastData

		if userSummary.LastData != nil {
			// if summary already exists with a last data checkpoint, start data pull there
			if startTime.Before(*userSummary.LastData) {
				startTime = *userSummary.LastData
			}

			// ensure endTime does not move backwards by capping it at summary LastData
			if !status.LastData.After(*userSummary.LastData) {
				endTime = *userSummary.LastData
			}
		}

		userData, err = dataRepository.GetDataRange(ctx, id, "bgm", startTime, endTime)
		if err != nil {
			return nil, err
		}
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
	userSummary, err = summaryRepository.UpdateBGMSummary(ctx, userSummary)

	return userSummary, err
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
