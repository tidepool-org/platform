package client

import (
	"context"
	"net/http"
	"time"

	"github.com/tidepool-org/platform/data/summary/types"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	// ExtendedTimeout is used by requests that we expect to take longer than
	// usual.
	ExtendedTimeout = 5 * time.Minute
)

// TODO: Move interface to data package once upload dependency broken
// TODO: Once above complete, rename ClientImpl to Client

type Client interface {
	data.DataSetAccessor

	CreateDataSetsData(ctx context.Context, dataSetID string, datumArray []data.Datum) error

	DestroyDataForUserByID(ctx context.Context, userID string) error

	GetCGMSummary(ctx context.Context, id string) (*types.Summary[types.CGMStats, *types.CGMStats], error)
	GetBGMSummary(ctx context.Context, id string) (*types.Summary[types.BGMStats, *types.BGMStats], error)
	UpdateCGMSummary(ctx context.Context, id string) (*types.Summary[types.CGMStats, *types.CGMStats], error)
	UpdateBGMSummary(ctx context.Context, id string) (*types.Summary[types.BGMStats, *types.BGMStats], error)
	GetOutdatedUserIDs(ctx context.Context, t string, pagination *page.Pagination) ([]string, error)
	GetMigratableUserIDs(ctx context.Context, t string, pagination *page.Pagination) ([]string, error)
	BackfillSummaries(ctx context.Context, t string) (int, error)
}

type ClientImpl struct {
	client                *platform.Client
	extendedTimeoutClient *platform.Client
}

func New(cfg *platform.Config, authorizeAs platform.AuthorizeAs) (*ClientImpl, error) {
	clnt, err := platform.NewClient(cfg, authorizeAs)
	if err != nil {
		return nil, err
	}

	return &ClientImpl{
		client: clnt,
	}, nil
}

func (c *ClientImpl) ListUserDataSets(ctx context.Context, userID string, filter *data.DataSetFilter, pagination *page.Pagination) (data.DataSets, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if filter == nil {
		filter = data.NewDataSetFilter()
	} else if err := structureValidator.New().Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New().Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	url := c.client.ConstructURL("v1", "users", userID, "data_sets")
	dataSets := data.DataSets{}
	if err := c.client.RequestData(ctx, http.MethodGet, url, []request.RequestMutator{filter, pagination}, nil, &dataSets); err != nil {
		return nil, err
	}

	return dataSets, nil
}

func (c *ClientImpl) CreateUserDataSet(ctx context.Context, userID string, create *data.DataSetCreate) (*data.DataSet, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if create == nil {
		return nil, errors.New("create is missing")
	} else if err := structureValidator.New().Validate(create); err != nil {
		return nil, errors.Wrap(err, "create is invalid")
	}

	url := c.client.ConstructURL("v1", "users", userID, "data_sets")
	response := struct {
		Data   *data.DataSet    `json:"data,omitempty"`
		Errors []*service.Error `json:"errors,omitempty"`
		Meta   *interface{}     `json:"meta,omitempty"`
	}{} // TODO: Remove response wrapper once service is updated
	if err := c.client.RequestData(ctx, http.MethodPost, url, nil, create, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}

func (c *ClientImpl) GetDataSet(ctx context.Context, id string) (*data.DataSet, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	}

	url := c.client.ConstructURL("v1", "data_sets", id)
	dataSet := &data.DataSet{}
	if err := c.client.RequestData(ctx, http.MethodGet, url, nil, nil, dataSet); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return dataSet, nil
}

func (c *ClientImpl) GetCGMSummary(ctx context.Context, userId string) (*types.Summary[types.CGMStats, *types.CGMStats], error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userId == "" {
		return nil, errors.New("userId is missing")
	}

	url := c.client.ConstructURL("v1", "summaries", "cgm", userId)
	summary := &types.Summary[types.CGMStats, *types.CGMStats]{}
	if err := c.client.RequestData(ctx, http.MethodGet, url, nil, nil, summary); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return summary, nil
}

func (c *ClientImpl) GetBGMSummary(ctx context.Context, userId string) (*types.Summary[types.BGMStats, *types.BGMStats], error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userId == "" {
		return nil, errors.New("id is missing")
	}

	url := c.client.ConstructURL("v1", "summaries", "bgm", userId)
	summary := &types.Summary[types.BGMStats, *types.BGMStats]{}
	if err := c.client.RequestData(ctx, http.MethodGet, url, nil, nil, summary); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return summary, nil
}

func (c *ClientImpl) UpdateCGMSummary(ctx context.Context, userId string) (*types.Summary[types.CGMStats, *types.CGMStats], error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userId == "" {
		return nil, errors.New("id is missing")
	}

	url := c.client.ConstructURL("v1", "summaries", "cgm", userId)
	summary := &types.Summary[types.CGMStats, *types.CGMStats]{}
	if err := c.client.RequestData(ctx, http.MethodPost, url, nil, nil, summary); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return summary, err
	}

	return summary, nil
}

func (c *ClientImpl) UpdateBGMSummary(ctx context.Context, userId string) (*types.Summary[types.BGMStats, *types.BGMStats], error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userId == "" {
		return nil, errors.New("id is missing")
	}

	url := c.client.ConstructURL("v1", "summaries", "bgm", userId)
	summary := &types.Summary[types.BGMStats, *types.BGMStats]{}
	if err := c.client.RequestData(ctx, http.MethodPost, url, nil, nil, summary); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return summary, err
	}

	return summary, nil
}

func (c *ClientImpl) BackfillSummaries(ctx context.Context, typ string) (int, error) {
	var count int
	url := c.client.ConstructURL("v1", "summaries", "backfill", typ)

	ctxWithTimeout, cancel := context.WithTimeout(ctx, ExtendedTimeout)
	defer cancel()
	if err := c.client.RequestData(ctxWithTimeout, http.MethodPost, url, nil, nil, &count); err != nil {
		return count, errors.Wrap(err, "backfill request returned an error")
	}

	return count, nil
}

func (c *ClientImpl) GetOutdatedUserIDs(ctx context.Context, typ string, pagination *page.Pagination) ([]string, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if typ == "" {
		return nil, errors.New("type is missing")
	}
	url := c.client.ConstructURL("v1", "summaries", "outdated", typ)

	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New().Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	var userIDs []string
	if err := c.client.RequestData(ctx, http.MethodGet, url, []request.RequestMutator{pagination}, nil, &userIDs); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return userIDs, err
	}

	return userIDs, nil
}

func (c *ClientImpl) GetMigratableUserIDs(ctx context.Context, typ string, pagination *page.Pagination) ([]string, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if typ == "" {
		return nil, errors.New("type is missing")
	}
	url := c.client.ConstructURL("v1", "summaries", "migratable", typ)

	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New().Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	var userIDs []string
	if err := c.client.RequestData(ctx, http.MethodGet, url, []request.RequestMutator{pagination}, nil, &userIDs); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return userIDs, err
	}

	return userIDs, nil
}

func (c *ClientImpl) UpdateDataSet(ctx context.Context, id string, update *data.DataSetUpdate) (*data.DataSet, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	}
	if update == nil {
		return nil, errors.New("update is missing")
	} else if err := structureValidator.New().Validate(update); err != nil {
		return nil, errors.Wrap(err, "update is invalid")
	}

	url := c.client.ConstructURL("v1", "data_sets", id)
	response := struct {
		Data   *data.DataSet    `json:"data,omitempty"`
		Errors []*service.Error `json:"errors,omitempty"`
		Meta   *interface{}     `json:"meta,omitempty"`
	}{} // TODO: Remove response wrapper once service is updated
	if err := c.client.RequestData(ctx, http.MethodPut, url, nil, update, &response); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return response.Data, nil
}

func (c *ClientImpl) DeleteDataSet(ctx context.Context, id string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if id == "" {
		return errors.New("id is missing")
	}

	url := c.client.ConstructURL("v1", "data_sets", id)
	return c.client.RequestData(ctx, http.MethodDelete, url, nil, nil, nil)
}

// TODO: Rename for consistency

func (c *ClientImpl) CreateDataSetsData(ctx context.Context, dataSetID string, datumArray []data.Datum) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if dataSetID == "" {
		return errors.New("data set id is missing")
	}
	if datumArray == nil {
		return errors.New("datum array is missing")
	}

	// TODO: Remove response wrapper once service is updated
	url := c.client.ConstructURL("v1", "data_sets", dataSetID, "data")
	response := struct {
		Data   *interface{}     `json:"data,omitempty"`
		Errors []*service.Error `json:"errors,omitempty"`
		Meta   *interface{}     `json:"meta,omitempty"`
	}{}
	return c.client.RequestData(ctx, http.MethodPost, url, nil, datumArray, &response)
}

// TODO: Rename for consistency

func (c *ClientImpl) DestroyDataForUserByID(ctx context.Context, userID string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userID == "" {
		return errors.New("user id is missing")
	}

	url := c.client.ConstructURL("v1", "users", userID, "data")
	return c.client.RequestData(ctx, http.MethodDelete, url, nil, nil, nil)
}
