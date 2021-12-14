package service

import (
	"context"
	"math"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/blood/glucose"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/summary"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
)

const (
	lowBloodGlucose  = 3.9
	highBloodGlucose = 10
	units            = "mmol/l"
)

type Client struct {
	dataStore dataStore.Store
}

// assumes all except freestyle is 5 minutes
func GetDuration(dataSet *continuous.Continuous) int64 {
	if strings.Contains(*dataSet.DeviceID, "AbbottFreeStyleLibre") {
		return 15
	}
	return 5
}

func CalculateWeight(startTime time.Time, endTime time.Time, lastData time.Time) (float64, time.Time, error) {
	var weight float64 = 1.0

	if endTime.Before(lastData) {
		return weight, startTime, errors.New("Invalid time period for calculation, endTime before lastData.")
	}

	if startTime.Before(lastData) {
		// get ratio between start time and actual start time for weights
		wholeTime := endTime.Sub(startTime)
		newTime := endTime.Sub(lastData)
		weight = newTime.Seconds() / wholeTime.Seconds()

		startTime = lastData
	}

	return weight, startTime, nil
}

func CalculateStats(userData []*continuous.Continuous, totalMinutes float64) *summary.Stats {
	var inRangeMinutes int64 = 0
	var belowRangeMinutes int64 = 0
	var aboveRangeMinutes int64 = 0
	var totalGlucose float64 = 0
	var normalizedValue float64

	for _, r := range userData {
		normalizedValue = *glucose.NormalizeValueForUnits(r.Value, pointer.FromString(units))

		if normalizedValue <= lowBloodGlucose {
			belowRangeMinutes += GetDuration(r)
		} else if normalizedValue >= highBloodGlucose {
			aboveRangeMinutes += GetDuration(r)
		} else {
			inRangeMinutes += GetDuration(r)
		}

		totalGlucose += normalizedValue
	}

	averageGlucose := totalGlucose / float64(len(userData))
	timeInRange := float64(inRangeMinutes) / totalMinutes
	timeBelowRange := float64(belowRangeMinutes) / totalMinutes
	timeAboveRange := float64(aboveRangeMinutes) / totalMinutes

	// return 2 decimals
	return &summary.Stats{
		TimeInRange:    math.Round(timeInRange*100) / 100,
		TimeBelowRange: math.Round(timeBelowRange*100) / 100,
		TimeAboveRange: math.Round(timeAboveRange*100) / 100,
		AverageGlucose: math.Round(averageGlucose*100) / 100,
	}
}

func ReweightStats(stats *summary.Stats, userSummary *summary.Summary, weight float64) (*summary.Stats, error) {
	if weight < 0 || weight > 1 {
		return stats, errors.New("Invalid weight (<0||>1) for stats")
	}
	// if we are rolling in previous averages
	if weight != 1 && weight >= 0 {
		stats.AverageGlucose = stats.AverageGlucose*weight + *userSummary.AverageGlucose.Value*(1-weight)
		stats.TimeInRange = stats.TimeInRange*weight + *userSummary.TimeInRange*(1-weight)
		stats.TimeBelowRange = stats.TimeBelowRange*weight + *userSummary.TimeBelowRange*(1-weight)
		stats.TimeAboveRange = stats.TimeAboveRange*weight + *userSummary.TimeAboveRange*(1-weight)
	}

	return stats, nil
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
	dataRepository := c.dataStore.NewDataRepository()

	userSummary, err := summaryRepository.GetSummary(ctx, id)

	if err == mongo.ErrNoDocuments {
		_, err := dataRepository.GetLastUpdatedForUser(ctx, id)

		if err != nil {
			return nil, nil
		} else {
			return &summary.Summary{}, nil
		}
	}
	return userSummary, err
}

func (c *Client) UpdateSummary(ctx context.Context, id string) (*summary.Summary, error) {
	summaryRepository := c.dataStore.NewSummaryRepository()
	dataRepository := c.dataStore.NewDataRepository()

	// we need the original summary object to grab the original for rolling calc
	userSummary, err := summaryRepository.GetSummary(ctx, id)

	if err == mongo.ErrNoDocuments {
		// check to ensure the user has data
		_, err := dataRepository.GetLastUpdatedForUser(ctx, id)
		if err != nil {
			return nil, nil
		}

		userSummary = summary.New(id)
	} else if err != nil {
		return nil, err
	}

	timestamp := time.Now().UTC()
	userSummary.LastUpdated = &timestamp

	status, err := dataRepository.GetLastUpdatedForUser(ctx, id)

	// remove 2 weeks for start time
	startTime := status.LastData.AddDate(0, 0, -14)
	firstData := startTime

	weight := 1.0
	if userSummary.LastData != nil {
		weight, startTime, err = CalculateWeight(startTime, status.LastData, *userSummary.LastData)

		if err != nil {
			return nil, err
		}
	}

	totalMinutes := float64(math.Round(status.LastData.Sub(startTime).Minutes()))

	// quit here if we dont have a long enough timeblock, and might result in +Inf result
	// 0.5 minutes was chosen to smooth any possible float inaccuracy with large division
	// and avoid calculating on duplicate calls
	// theres nothing actually wrong here, so dont return an error.
	if totalMinutes < 0.5 {
		return userSummary, nil
	}

	userData, err := dataRepository.GetCGMDataRange(ctx, id, startTime, status.LastData)
	if err != nil {
		return nil, err
	}

	stats := CalculateStats(userData, totalMinutes)
	stats, err = ReweightStats(stats, userSummary, weight)
	if err != nil {
		return nil, err
	}

	userSummary.LastUpload = &status.LastUpload
	userSummary.LastData = &status.LastData
	userSummary.FirstData = &firstData
	userSummary.TimeInRange = pointer.FromFloat64(stats.TimeInRange)
	userSummary.TimeBelowRange = pointer.FromFloat64(stats.TimeBelowRange)
	userSummary.TimeAboveRange = pointer.FromFloat64(stats.TimeAboveRange)
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

func (c *Client) GetAgedSummaries(ctx context.Context, pagination *page.Pagination) ([]*summary.Summary, error) {
	summaryRepository := c.dataStore.NewSummaryRepository()
	dataRepository := c.dataStore.NewDataRepository()

	lastUpdated, err := summaryRepository.GetLastUpdated(ctx)
	if err != nil {
		return nil, err
	}

	agedSummaries, err := summaryRepository.GetAgedSummaries(ctx, lastUpdated)
	if err != nil {
		return nil, err
	}

	freshUserIds, err := dataRepository.GetFreshUsers(ctx, lastUpdated)
	if err != nil {
		return nil, err
	}

	freshUserMap := make(map[string]struct{})
	var empty struct{}
	for _, v := range freshUserIds {
		freshUserMap[v] = empty
	}

	var summariesReqUpdate []*summary.Summary

	for _, agedSummary := range agedSummaries {
		if _, exists := freshUserMap[agedSummary.UserID]; exists {
			summariesReqUpdate = append(summariesReqUpdate, agedSummary)
		} else {
			_, err := summaryRepository.UpdateLastUpdated(ctx, agedSummary.UserID)
			if err != nil {
				return nil, err
			}
		}

		if len(summariesReqUpdate) >= pagination.Size {
			break
		}
	}

	if summariesReqUpdate == nil {
		summariesReqUpdate = []*summary.Summary{}
	}

	return summariesReqUpdate, err
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
