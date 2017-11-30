package service

import (
	"context"

	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
	dataStoreDEPRECATED "github.com/tidepool-org/platform/data/storeDEPRECATED"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/page"
)

type Client struct {
	dataStore           dataStore.Store
	dataStoreDEPRECATED dataStoreDEPRECATED.Store
}

func NewClient(str dataStore.Store, strDEPRECATED dataStoreDEPRECATED.Store) (*Client, error) {
	if str == nil {
		return nil, errors.New("data store is missing")
	}
	if strDEPRECATED == nil {
		return nil, errors.New("data store deprecated is missing")
	}

	return &Client{
		dataStore:           str,
		dataStoreDEPRECATED: strDEPRECATED,
	}, nil
}

func (c *Client) ListUserDataSources(ctx context.Context, userID string, filter *data.DataSourceFilter, pagination *page.Pagination) (data.DataSources, error) {
	ssn := c.dataStore.NewDataSourceSession()
	defer ssn.Close()

	return ssn.ListUserDataSources(ctx, userID, filter, pagination)
}

func (c *Client) CreateUserDataSource(ctx context.Context, userID string, create *data.DataSourceCreate) (*data.DataSource, error) {
	ssn := c.dataStore.NewDataSourceSession()
	defer ssn.Close()

	return ssn.CreateUserDataSource(ctx, userID, create)
}

func (c *Client) GetDataSource(ctx context.Context, id string) (*data.DataSource, error) {
	ssn := c.dataStore.NewDataSourceSession()
	defer ssn.Close()

	return ssn.GetDataSource(ctx, id)
}

func (c *Client) UpdateDataSource(ctx context.Context, id string, update *data.DataSourceUpdate) (*data.DataSource, error) {
	ssn := c.dataStore.NewDataSourceSession()
	defer ssn.Close()

	return ssn.UpdateDataSource(ctx, id, update)
}

func (c *Client) DeleteDataSource(ctx context.Context, id string) error {
	ssn := c.dataStore.NewDataSourceSession()
	defer ssn.Close()

	return ssn.DeleteDataSource(ctx, id)
}

func (c *Client) CreateUserDataSet(ctx context.Context, userID string, create *data.DataSetCreate) (*data.DataSet, error) {
	panic("Not Implemented!")
}

func (c *Client) ListUserDataSets(ctx context.Context, userID string, filter *data.DataSetFilter, pagination *page.Pagination) (data.DataSets, error) {
	ssn := c.dataStoreDEPRECATED.NewDataSession()
	defer ssn.Close()

	return ssn.ListUserDataSets(ctx, userID, filter, pagination)
}

func (c *Client) GetDataSet(ctx context.Context, id string) (*data.DataSet, error) {
	ssn := c.dataStoreDEPRECATED.NewDataSession()
	defer ssn.Close()

	return ssn.GetDataSet(ctx, id)
}

func (c *Client) UpdateDataSet(ctx context.Context, id string, update *data.DataSetUpdate) (*data.DataSet, error) {
	panic("Not Implemented!")
}

func (c *Client) DeleteDataSet(ctx context.Context, id string) error {
	panic("Not Implemented!")
}

func (c *Client) CreateDataSetsData(ctx context.Context, datasetID string, datumArray []data.Datum) error {
	panic("Not Implemented!")
}

func (c *Client) DestroyDataForUserByID(ctx context.Context, userID string) error {
	panic("Not Implemented!")
}
