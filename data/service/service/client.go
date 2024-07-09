package service

import (
	"context"

	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/summary/types"
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

func (c *Client) GetCGMSummary(ctx context.Context, id string) (*types.Summary[*types.CGMStats, types.CGMStats], error) {
	panic("Not Implemented!")
}

func (c *Client) GetBGMSummary(ctx context.Context, id string) (*types.Summary[*types.BGMStats, types.BGMStats], error) {
	panic("Not Implemented!")
}

func (c *Client) UpdateCGMSummary(ctx context.Context, id string) (*types.Summary[*types.CGMStats, types.CGMStats], error) {
	panic("Not Implemented!")
}

func (c *Client) UpdateBGMSummary(ctx context.Context, id string) (*types.Summary[*types.BGMStats, types.BGMStats], error) {
	panic("Not Implemented!")
}

func (c *Client) GetOutdatedUserIDs(ctx context.Context, t string, pagination *page.Pagination) (*types.OutdatedSummariesResponse, error) {
	panic("Not Implemented!")
}

func (c *Client) BackfillSummaries(ctx context.Context, t string) (int, error) {
	panic("Not Implemented!")
}

func (c *Client) GetMigratableUserIDs(ctx context.Context, t string, pagination *page.Pagination) ([]string, error) {
	panic("Not Implemented!")
}

func (c *Client) GetContinuousSummary(ctx context.Context, id string) (*types.Summary[*types.ContinuousStats, types.ContinuousStats], error) {
	panic("Not Implemented!")
}

func (c *Client) UpdateContinuousSummary(ctx context.Context, id string) (*types.Summary[*types.ContinuousStats, types.ContinuousStats], error) {
	panic("Not Implemented!")
}
