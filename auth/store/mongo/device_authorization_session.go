package mongo

import (
	"context"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type DeviceAuthorizationSession struct {
	*storeStructuredMongo.Session
}

func (d *DeviceAuthorizationSession) EnsureIndexes() error {
	return d.EnsureAllIndexes([]mgo.Index{
		{Key: []string{"id"}, Unique: true, Background: true},
		{Key: []string{"token"}, Unique: true, Background: true},
	})
}

func (d *DeviceAuthorizationSession) CreateUserDeviceAuthorization(ctx context.Context, userID string, create *auth.DeviceAuthorizationCreate) (*auth.DeviceAuthorization, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if d.IsClosed() {
		return nil, errors.New("session is closed")
	}

	deviceAuthorization, err := auth.NewDeviceAuthorization(userID, create)
	if err != nil {
		return nil, err
	} else if err = structureValidator.New().Validate(deviceAuthorization); err != nil {
		return nil, errors.Wrap(err, "device authorization is invalid")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "create": create})

	err = d.C().Insert(deviceAuthorization)
	logger.WithFields(log.Fields{"id": deviceAuthorization.ID, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("CreateDeviceAuthorization")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create user device authorization")
	}

	return deviceAuthorization, nil
}

func (d *DeviceAuthorizationSession) ListUserDeviceAuthorizations(ctx context.Context, userID string, pagination *page.Pagination) (auth.DeviceAuthorizations, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if d.IsClosed() {
		return nil, errors.New("session is closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "pagination": pagination})

	selector := bson.M{
		"userId": userID,
	}

	deviceAuthorizations := auth.DeviceAuthorizations{}
	err := d.C().Find(selector).Sort("-createdTime").Skip(pagination.Page * pagination.Size).Limit(pagination.Size).All(&deviceAuthorizations)
	logger.WithFields(log.Fields{"count": len(deviceAuthorizations), "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("ListUserDeviceAuthorizations")
	if err != nil {
		return nil, errors.Wrap(err, "unable to list device authorizations for user")
	}

	if deviceAuthorizations == nil {
		deviceAuthorizations = auth.DeviceAuthorizations{}
	}

	return deviceAuthorizations, nil
}

func (d *DeviceAuthorizationSession) GetUserDeviceAuthorization(ctx context.Context, userID string, id string) (*auth.DeviceAuthorization, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if d.IsClosed() {
		return nil, errors.New("session is closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "id": id})

	selector := bson.M{
		"userId": userID,
		"id":     id,
	}

	deviceAuthorization := &auth.DeviceAuthorization{}
	err := d.C().Find(selector).One(deviceAuthorization)
	logger.WithFields(log.Fields{"duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("GetUserDeviceAuthorizations")
	if err == mgo.ErrNotFound {
		deviceAuthorization = nil
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to get user device authorization")
	}

	return deviceAuthorization, nil
}

func (d *DeviceAuthorizationSession) GetDeviceAuthorization(ctx context.Context, id string) (*auth.DeviceAuthorization, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	if d.IsClosed() {
		return nil, errors.New("session is closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"id": id})

	selector := bson.M{
		"id": id,
	}

	deviceAuthorization := &auth.DeviceAuthorization{}
	err := d.C().Find(selector).One(deviceAuthorization)
	logger.WithFields(log.Fields{"duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("GetDeviceAuthorization")
	if err == mgo.ErrNotFound {
		deviceAuthorization = nil
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to get device authorization")
	}

	return deviceAuthorization, nil
}

func (d *DeviceAuthorizationSession) GetDeviceAuthorizationByToken(ctx context.Context, token string) (*auth.DeviceAuthorization, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if d.IsClosed() {
		return nil, errors.New("session is closed")
	}

	deviceAuthorization := &auth.DeviceAuthorization{}
	now := time.Now()

	// Log only a prefix, because the token is used for authorization
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"token": token[0:5]})

	selector := bson.M{
		"token": token,
	}
	err := d.C().Find(selector).One(deviceAuthorization)
	logger.WithFields(log.Fields{"duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("GetDeviceAuthorizationByToken")
	if err == mgo.ErrNotFound {
		deviceAuthorization = nil
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to get device authorization by token")
	}

	return deviceAuthorization, nil
}

func (d *DeviceAuthorizationSession) UpdateDeviceAuthorization(ctx context.Context, id string, update *auth.DeviceAuthorizationUpdate) (*auth.DeviceAuthorization, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if d.IsClosed() {
		return nil, errors.New("session is closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"id": id})

	selector := bson.M{
		"id":     id,
		"status": auth.DeviceAuthorizationPending,
	}
	set := bson.M{
		"bundleId":         update.BundleID,
		"deviceCheckToken": update.DeviceCheckToken,
		"modifiedTime":     now,
		"status":           update.Status,
		"verificationCode": update.VerificationCode,
	}

	changeInfo, err := d.C().UpdateAll(selector, d.ConstructUpdate(set, bson.M{}))
	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("UpdateProviderSession")
	if err != nil {
		return nil, errors.Wrap(err, "unable to update device authorization")
	}

	if changeInfo.Updated == 0 {
		return nil, errors.New("unable to update non-existing or completed device authorization")
	}

	return d.GetDeviceAuthorization(ctx, id)
}
