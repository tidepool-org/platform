package client

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/summary/types"
)

// TODO: Move interface to data package once upload dependency broken
// TODO: Once above complete, rename ClientImpl to Client

type Client interface {
	data.DataSetAccessor

	CreateDataSetsData(ctx context.Context, dataSetID string, datumArray []data.Datum) error
	DestroyDataForUserByID(ctx context.Context, userID string) error
	GetCGMSummary(ctx context.Context, id string) (*types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket], error)
	GetBGMSummary(ctx context.Context, id string) (*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket], error)
	GetContinuousSummary(ctx context.Context, id string) (*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket], error)
	UpdateCGMSummary(ctx context.Context, id string) (*types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket], error)
	UpdateBGMSummary(ctx context.Context, id string) (*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket], error)
	UpdateContinuousSummary(ctx context.Context, id string) (*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket], error)
	GetOutdatedUserIDs(ctx context.Context, t string, pagination *page.Pagination) (*types.OutdatedSummariesResponse, error)
	GetMigratableUserIDs(ctx context.Context, t string, pagination *page.Pagination) ([]string, error)
}

type ClientImpl struct {
	client *platform.Client
}

func New(cfg *platform.Config, authorizeAs platform.AuthorizeAs) (*ClientImpl, error) {
	clnt, err := platform.NewClientWithErrorResponseParser(cfg, authorizeAs, NewSerializableDataErrorResponseParser())
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
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(pagination); err != nil {
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
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(create); err != nil {
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

func (c *ClientImpl) GetCGMSummary(ctx context.Context, userId string) (*types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket], error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userId == "" {
		return nil, errors.New("userId is missing")
	}

	url := c.client.ConstructURL("v1", "summaries", "cgm", userId)
	summary := &types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]{}
	if err := c.client.RequestData(ctx, http.MethodGet, url, nil, nil, summary); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return summary, nil
}

func (c *ClientImpl) GetBGMSummary(ctx context.Context, userId string) (*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket], error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userId == "" {
		return nil, errors.New("id is missing")
	}

	url := c.client.ConstructURL("v1", "summaries", "bgm", userId)
	summary := &types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{}
	if err := c.client.RequestData(ctx, http.MethodGet, url, nil, nil, summary); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return summary, nil
}

func (c *ClientImpl) GetContinuousSummary(ctx context.Context, userId string) (*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket], error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userId == "" {
		return nil, errors.New("id is missing")
	}

	url := c.client.ConstructURL("v1", "summaries", "con", userId)
	summary := &types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{}
	if err := c.client.RequestData(ctx, http.MethodGet, url, nil, nil, summary); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return summary, nil
}

func (c *ClientImpl) UpdateCGMSummary(ctx context.Context, userId string) (*types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket], error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userId == "" {
		return nil, errors.New("id is missing")
	}

	url := c.client.ConstructURL("v1", "summaries", "cgm", userId)
	summary := &types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]{}
	if err := c.client.RequestData(ctx, http.MethodPost, url, nil, nil, summary); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, errors.Cause(err)
	}

	return summary, nil
}

func (c *ClientImpl) UpdateBGMSummary(ctx context.Context, userId string) (*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket], error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userId == "" {
		return nil, errors.New("id is missing")
	}

	url := c.client.ConstructURL("v1", "summaries", "bgm", userId)
	summary := &types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{}
	if err := c.client.RequestData(ctx, http.MethodPost, url, nil, nil, summary); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, errors.Cause(err)
	}

	return summary, nil
}

func (c *ClientImpl) UpdateContinuousSummary(ctx context.Context, userId string) (*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket], error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userId == "" {
		return nil, errors.New("id is missing")
	}

	url := c.client.ConstructURL("v1", "summaries", "con", userId)
	summary := &types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{}
	if err := c.client.RequestData(ctx, http.MethodPost, url, nil, nil, summary); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, errors.Cause(err)
	}

	return summary, nil
}

func (c *ClientImpl) GetOutdatedUserIDs(ctx context.Context, typ string, pagination *page.Pagination) (*types.OutdatedSummariesResponse, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if typ == "" {
		return nil, errors.New("type is missing")
	}
	url := c.client.ConstructURL("v1", "summaries", "outdated", typ)

	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	response := &types.OutdatedSummariesResponse{}
	err := c.client.RequestData(ctx, http.MethodGet, url, []request.RequestMutator{pagination}, nil, response)

	return response, err
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
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	var userIDs []string
	err := c.client.RequestData(ctx, http.MethodGet, url, []request.RequestMutator{pagination}, nil, &userIDs)

	return userIDs, err
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
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(update); err != nil {
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

func NewSerializableDataErrorResponseParser() *SerializableDataErrorResponseParser {
	return &SerializableDataErrorResponseParser{}
}

type SerializableDataErrorResponseParser struct{}

func (s *SerializableDataErrorResponseParser) ParseErrorResponse(ctx context.Context, res *http.Response, req *http.Request) error {
	serializable := &struct {
		Errors errors.Serializable `json:"errors,omitempty"`
	}{}
	if err := json.NewDecoder(res.Body).Decode(serializable); err != nil {
		return nil
	}
	if serializable.Errors.Error != nil {
		return serializable.Errors.Error
	}
	return nil
}
