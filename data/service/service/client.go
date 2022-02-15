package service

import (
	"context"
	"math"
	"time"

	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/blood/glucose/summary"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
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
	var err error
	var status *summary.UserLastUpdated
	summaryRepository := c.dataStore.NewSummaryRepository()
	dataRepository := c.dataStore.NewDataRepository()

	// we need the original summary object to grab the original for rolling calc
	userSummary, err := summaryRepository.GetSummary(ctx, id)
	if err != nil {
		return nil, err
	} else if userSummary == nil {
		// check to ensure the user has data
		status, err = dataRepository.GetLastUpdatedForUser(ctx, id)
		if err != nil {
			return nil, err
		}

		userSummary = summary.New(id)
	}

	timestamp := time.Now().UTC()
	userSummary.LastUpdated = &timestamp

	if status == nil {
		status, err = dataRepository.GetLastUpdatedForUser(ctx, id)
		if err != nil {
			return nil, err
		}
	}

	// remove 2 weeks for start time
	startTime := status.LastData.AddDate(0, 0, -14)
	firstData := startTime

	var newWeight *summary.WeightingResult
	if userSummary.LastData != nil {
		newWeight, err = summary.CalculateWeight(startTime, status.LastData, *userSummary.LastData)

		if err != nil {
			return nil, err
		}
	} else {
		newWeight = &summary.WeightingResult{Weight: 1.0, StartTime: startTime}
	}

	totalMinutes := float64(math.Round(status.LastData.Sub(newWeight.StartTime).Minutes()))

	// quit here if we dont have a long enough timeblock, and might result in +Inf result
	// 0.5 minutes was chosen to smooth any possible float inaccuracy with large division
	// and avoid calculating on duplicate calls
	// theres nothing actually wrong here, so dont return an error.
	if totalMinutes < 0.5 {
		return userSummary, nil
	}

	userData, err := dataRepository.GetCGMDataRange(ctx, id, newWeight.StartTime, status.LastData)
	if err != nil {
		return nil, err
	}

	stats := summary.CalculateStats(userData, totalMinutes)
	stats, err = summary.ReweightStats(stats, userSummary, newWeight.Weight)
	if err != nil {
		return nil, err
	}

	userSummary.LastUpload = &status.LastUpload
	userSummary.LastData = &status.LastData
	userSummary.FirstData = &firstData
	userSummary.TimeInRange = pointer.FromFloat64(stats.TimeInRange)
	userSummary.TimeBelowRange = pointer.FromFloat64(stats.TimeBelowRange)
	userSummary.TimeAboveRange = pointer.FromFloat64(stats.TimeAboveRange)
	userSummary.TimeCGMUse = pointer.FromFloat64(stats.TimeCGMUse)
	userSummary.AverageGlucose = &summary.Glucose{
		Value: pointer.FromFloat64(stats.AverageGlucose),
		Units: pointer.FromString(units),
	}
	userSummary.LowGlucoseThreshold = pointer.FromFloat64(lowBloodGlucose)
	userSummary.HighGlucoseThreshold = pointer.FromFloat64(highBloodGlucose)

	userSummary, err = summaryRepository.UpdateSummary(ctx, userSummary)
	if err != nil {
		return nil, err
	}

	return userSummary, err
}

func (c *Client) BackfillSummaries(ctx context.Context) (int64, error) {
	var empty struct{}
	userIDsReqUpdate := []string{}
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
			userIDsReqUpdate = append(userIDsReqUpdate, userID)
		}

		if len(userIDsReqUpdate) >= backfillBatch {
			break
		}
	}

	var summaries []*summary.Summary

	for _, userID := range userIDsReqUpdate {
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

func (c *Client) GetAgedUserIDs(ctx context.Context, pagination *page.Pagination) ([]string, error) {
	var empty struct{}
	userIDsReqUpdate := []string{}

	summaryRepository := c.dataStore.NewSummaryRepository()
	dataRepository := c.dataStore.NewDataRepository()

	oldestUpdate, err := summaryRepository.GetOldestUpdate(ctx)
	if err != nil {
		return nil, err
	}

	agedUserIDs, err := summaryRepository.GetUsersWithSummariesBefore(ctx, time.Now().Add(-60*time.Minute))
	if err != nil {
		return nil, err
	}

	freshUserIDs, err := dataRepository.GetUsersWithBGDataSince(ctx, *oldestUpdate)
	if err != nil {
		return nil, err
	}

	freshUserMap := make(map[string]struct{})
	for _, v := range freshUserIDs {
		freshUserMap[v] = empty
	}

	for _, userID := range agedUserIDs {
		if _, exists := freshUserMap[userID]; exists {
			userIDsReqUpdate = append(userIDsReqUpdate, userID)
		} else {
			_, err := summaryRepository.UpdateLastUpdated(ctx, userID)
			if err != nil {
				return nil, err
			}
		}

		if len(userIDsReqUpdate) >= pagination.Size {
			break
		}
	}

	return userIDsReqUpdate, err
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
