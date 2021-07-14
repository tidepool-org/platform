package service

import (
	"context"
    "time"

	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
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

func (c *Client) GetSummary(ctx context.Context, id string) (*data.Summary, error) {
	repository := c.dataStore.NewSummaryRepository()
	return repository.GetSummary(ctx, id)
}

func (c *Client) UpdateSummary(ctx context.Context, id string) (*data.Summary, error) {
	summaryRepository := c.dataStore.NewSummaryRepository()
	dataRepository := c.dataStore.NewDataRepository()

    // we need the original summary object to grab the original for rolling calc
    summary, err := summaryRepository.GetSummary(ctx, id)
    if err != nil {
		return nil, err
	}

    summary, err = dataRepository.CalculateSummary(ctx, summary)
    if err != nil {
		return nil, err
	}

    summary, err = summaryRepository.UpdateSummary(ctx, summary)
    if err != nil {
		return nil, err
	}

    return summary, err
}

func (c *Client) GetAgedSummaries(ctx context.Context, minutes uint) ([]*data.Summary, error) {
    summaryRepository := c.dataStore.NewSummaryRepository()
	dataRepository := c.dataStore.NewDataRepository()

    // first get aged summaries
    // then query latest upload of dataset
    // if last upload newer than summary, add to list
    summaries, err := summaryRepository.GetAgedSummaries(ctx, minutes)

    if err != nil {
		return nil, err
	}

	var lastUpdated time.Time
	var freshTime time.Time
	var agedSummaries []*data.Summary

    for _, summary := range summaries {
        lastUpdated, err = dataRepository.GetLastUpdated(ctx, summary.UserID)
        if err != nil {
            return nil, err
        }

        // accept half of interval difference as "fresh" to prevent noise
        freshTime = summary.LastUpdated.Add(time.Minute * -time.Duration(minutes/2))

        if freshTime.Before(lastUpdated) {
            agedSummaries = append(agedSummaries, summary)
        }
    }

    if agedSummaries == nil {
		agedSummaries = []*data.Summary{}
	}

    return agedSummaries, err
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
