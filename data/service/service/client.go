package service

import (
	"context"

	"github.com/tidepool-org/platform/data"
	dataDeduplicator "github.com/tidepool-org/platform/data/deduplicator"
	dataStore "github.com/tidepool-org/platform/data/store"
	dataSummaryTypes "github.com/tidepool-org/platform/data/summary/types"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
)

type Client struct {
	logger                log.Logger
	dataStore             dataStore.Store
	dataDuplicatorFactory dataDeduplicator.Factory
}

func NewClient(logger log.Logger, strDEPRECATED dataStore.Store, dataDuplicatorFactory dataDeduplicator.Factory) (*Client, error) {
	if logger == nil {
		return nil, errors.New("logger is missing")
	}
	if strDEPRECATED == nil {
		return nil, errors.New("data store deprecated is missing")
	}
	if dataDuplicatorFactory == nil {
		return nil, errors.New("data deduplicator factory is missing")
	}

	return &Client{
		logger:                logger,
		dataStore:             strDEPRECATED,
		dataDuplicatorFactory: dataDuplicatorFactory,
	}, nil
}

func (c *Client) ListUserDataSets(ctx context.Context, userID string, filter *data.DataSetFilter, pagination *page.Pagination) (data.DataSets, error) {
	repository := c.dataStore.NewDataRepository()
	return repository.ListUserDataSets(ctx, userID, filter, pagination)
}

func (c *Client) CreateUserDataSet(ctx context.Context, userID string, create *data.DataSetCreate) (*data.DataSet, error) {
	repository := c.dataStore.NewDataRepository()

	dataSet, err := repository.CreateUserDataSet(ctx, userID, create)
	if err != nil {
		return nil, err
	}

	var deduplicator dataDeduplicator.Deduplicator
	if deduplicator, err = c.dataDuplicatorFactory.New(ctx, dataSet); err != nil {
		err = errors.Wrap(err, "unable to get deduplicator")
	} else if deduplicator == nil {
		err = errors.Wrap(err, "deduplicator not found")
	} else if dataSet, err = deduplicator.Open(ctx, dataSet); err != nil {
		err = errors.Wrap(err, "unable to open")
	} else {
		return dataSet, nil
	}

	c.logger.WithError(err).Error("Unable to create data set")

	if err = repository.DeleteDataSet(ctx, dataSet); err != nil {
		c.logger.WithError(err).Error("Unable to delete data set after unable to create data set")
	}

	return dataSet, nil
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

func (c *Client) GetCGMSummary(ctx context.Context, id string) (*dataSummaryTypes.Summary[*dataSummaryTypes.CGMStats, dataSummaryTypes.CGMStats], error) {
	panic("Not Implemented!")
}

func (c *Client) GetBGMSummary(ctx context.Context, id string) (*dataSummaryTypes.Summary[*dataSummaryTypes.BGMStats, dataSummaryTypes.BGMStats], error) {
	panic("Not Implemented!")
}

func (c *Client) UpdateCGMSummary(ctx context.Context, id string) (*dataSummaryTypes.Summary[*dataSummaryTypes.CGMStats, dataSummaryTypes.CGMStats], error) {
	panic("Not Implemented!")
}

func (c *Client) UpdateBGMSummary(ctx context.Context, id string) (*dataSummaryTypes.Summary[*dataSummaryTypes.BGMStats, dataSummaryTypes.BGMStats], error) {
	panic("Not Implemented!")
}

func (c *Client) GetOutdatedUserIDs(ctx context.Context, t string, pagination *page.Pagination) (*dataSummaryTypes.OutdatedSummariesResponse, error) {
	panic("Not Implemented!")
}

func (c *Client) BackfillSummaries(ctx context.Context, t string) (int, error) {
	panic("Not Implemented!")
}

func (c *Client) GetMigratableUserIDs(ctx context.Context, t string, pagination *page.Pagination) ([]string, error) {
	panic("Not Implemented!")
}

func (c *Client) GetContinuousSummary(ctx context.Context, id string) (*dataSummaryTypes.Summary[*dataSummaryTypes.ContinuousStats, dataSummaryTypes.ContinuousStats], error) {
	panic("Not Implemented!")
}

func (c *Client) UpdateContinuousSummary(ctx context.Context, id string) (*dataSummaryTypes.Summary[*dataSummaryTypes.ContinuousStats, dataSummaryTypes.ContinuousStats], error) {
	panic("Not Implemented!")
}
