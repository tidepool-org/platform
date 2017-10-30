package mongo

import (
	"context"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/store/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type DataSourceSession struct {
	*mongo.Session
}

func (d *DataSourceSession) EnsureIndexes() error {
	return d.EnsureAllIndexes([]mgo.Index{
		{Key: []string{"id"}, Unique: true, Background: true},
		{Key: []string{"userId"}, Background: true},
	})
}

func (d *DataSourceSession) ListUserDataSources(ctx context.Context, userID string, filter *data.DataSourceFilter, pagination *page.Pagination) (data.DataSources, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if filter == nil {
		filter = data.NewDataSourceFilter()
	} else if err := structureValidator.New().Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New().Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	if d.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "filter": filter, "pagination": pagination})

	dataSources := data.DataSources{}
	selector := bson.M{
		"userId": userID,
	}
	if filter.ProviderType != nil {
		selector["providerType"] = *filter.ProviderType
	}
	if filter.ProviderName != nil {
		selector["providerName"] = *filter.ProviderName
	}
	if filter.ProviderSessionID != nil {
		selector["providerSessionId"] = *filter.ProviderSessionID
	}
	if filter.State != nil {
		selector["state"] = *filter.State
	}
	err := d.C().Find(selector).Sort("-createdTime").Skip(pagination.Page * pagination.Size).Limit(pagination.Size).All(&dataSources)
	logger.WithFields(log.Fields{"count": len(dataSources), "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("ListUserDataSources")
	if err != nil {
		return nil, errors.Wrap(err, "unable to list user data sources")
	}

	if dataSources == nil {
		dataSources = data.DataSources{}
	}

	return dataSources, nil
}

func (d *DataSourceSession) CreateUserDataSource(ctx context.Context, userID string, create *data.DataSourceCreate) (*data.DataSource, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	dataSource, err := data.NewDataSource(userID, create)
	if err != nil {
		return nil, err
	} else if err = structureValidator.New().Validate(dataSource); err != nil {
		return nil, errors.Wrap(err, "data source is invalid")
	}

	if d.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "create": create})

	err = d.C().Insert(dataSource)
	logger.WithFields(log.Fields{"id": dataSource.ID, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("CreateUserDataSource")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create user data source")
	}

	return dataSource, nil
}

func (d *DataSourceSession) GetDataSource(ctx context.Context, id string) (*data.DataSource, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	}

	if d.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("id", id)

	dataSources := data.DataSources{}
	err := d.C().Find(bson.M{"id": id}).Limit(2).All(&dataSources)
	logger.WithField("duration", time.Since(now)/time.Microsecond).WithError(err).Debug("GetDataSource")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get data source")
	}

	switch count := len(dataSources); count {
	case 0:
		return nil, nil
	case 1:
		return dataSources[0], nil
	default:
		logger.WithField("count", count).Warnf("Multiple data sources found for id %q", id)
		return dataSources[0], nil
	}
}

func (d *DataSourceSession) UpdateDataSource(ctx context.Context, id string, update *data.DataSourceUpdate) (*data.DataSource, error) {
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

	if d.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"id": id, "update": update})

	set := bson.M{
		"modifiedTime": now.Truncate(time.Second),
	}
	unset := bson.M{}
	if update.State != nil {
		set["state"] = *update.State
		switch *update.State {
		case data.DataSourceStateDisconnected:
			unset["providerSessionId"] = true
			unset["error"] = true
		case data.DataSourceStateConnected:
			unset["error"] = true
		}
	}
	if update.Error != nil {
		set["error"] = *update.Error
	}
	if update.DataSetIDs != nil {
		set["dataSetIds"] = *update.DataSetIDs
	}
	if update.EarliestDataTime != nil {
		set["earliestDataTime"] = (*update.EarliestDataTime).Truncate(time.Second)
	}
	if update.LatestDataTime != nil {
		set["latestDataTime"] = (*update.LatestDataTime).Truncate(time.Second)
	}
	if update.LastImportTime != nil {
		set["lastImportTime"] = (*update.LastImportTime).Truncate(time.Second)
	}
	changeInfo, err := d.C().UpdateAll(bson.M{"id": id}, d.ConstructUpdate(set, unset))
	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("UpdateDataSource")
	if err != nil {
		return nil, errors.Wrap(err, "unable to update data source")
	}

	return d.GetDataSource(ctx, id)
}

func (d *DataSourceSession) DeleteDataSource(ctx context.Context, id string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if id == "" {
		return errors.New("id is missing")
	}

	if d.IsClosed() {
		return errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("id", id)

	changeInfo, err := d.C().RemoveAll(bson.M{"id": id})
	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("DeleteDataSource")
	if err != nil {
		return errors.Wrap(err, "unable to delete data source")
	}

	return nil
}
