package mongo

import (
	"context"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/blob"
	blobStoreStructured "github.com/tidepool-org/platform/blob/store/structured"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

type Store struct {
	*storeStructuredMongo.Store
}

func NewStore(config *storeStructuredMongo.Config, logger log.Logger) (*Store, error) {
	store, err := storeStructuredMongo.NewStore(config, logger)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: store,
	}, nil
}

func (s *Store) EnsureIndexes() error {
	session := s.newSession()
	defer session.Close()
	return session.EnsureIndexes()
}

func (s *Store) NewSession() blobStoreStructured.Session {
	return s.newSession()
}

func (s *Store) newSession() *Session {
	return &Session{
		Session: s.Store.NewSession("blobs"),
	}
}

type Session struct {
	*storeStructuredMongo.Session
}

func (s *Session) EnsureIndexes() error {
	return s.EnsureAllIndexes([]mgo.Index{
		{Key: []string{"id"}, Background: true, Unique: true},
		{Key: []string{"userId"}, Background: true},
		{Key: []string{"mediaType"}, Background: true},
		{Key: []string{"status"}, Background: true},
	})
}

func (s *Session) List(ctx context.Context, userID string, filter *blob.Filter, pagination *page.Pagination) (blob.Blobs, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return nil, errors.New("user id is invalid")
	}
	if filter == nil {
		filter = blob.NewFilter()
	} else if err := structureValidator.New().Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New().Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	if s.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "filter": filter, "pagination": pagination})

	blbs := blob.Blobs{}
	query := bson.M{
		"userId": userID,
	}
	if filter.MediaType != nil {
		query["mediaType"] = bson.M{
			"$in": *filter.MediaType,
		}
	}
	if filter.Status != nil {
		query["status"] = bson.M{
			"$in": *filter.Status,
		}
	} else {
		query["status"] = blob.StatusAvailable
	}
	err := s.C().Find(query).Sort("-createdTime").Skip(pagination.Page * pagination.Size).Limit(pagination.Size).All(&blbs)
	if err != nil {
		logger.WithError(err).Error("Unable to list blobs")
		return nil, errors.Wrap(err, "unable to list blobs")
	}

	logger.WithFields(log.Fields{"count": len(blbs), "duration": time.Since(now) / time.Microsecond}).Debug("List")
	return blbs, nil
}

func (s *Session) Create(ctx context.Context, userID string, create *blobStoreStructured.Create) (*blob.Blob, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return nil, errors.New("user id is invalid")
	}
	if create == nil {
		return nil, errors.New("create is missing")
	} else if err := structureValidator.New().Validate(create); err != nil {
		return nil, errors.Wrap(err, "create is invalid")
	}

	if s.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "create": create})

	doc := &blob.Blob{
		UserID:      pointer.FromString(userID),
		MediaType:   create.MediaType,
		Status:      pointer.FromString(blob.StatusCreated),
		CreatedTime: pointer.FromTime(now.Truncate(time.Second)),
	}

	var id string
	var err error
	for retry := 0; retry < 3; retry++ {
		id = blob.NewID()
		logger = logger.WithField("id", id)

		doc.ID = pointer.FromString(id)
		if err = s.C().Insert(doc); mgo.IsDup(err) {
			logger.WithError(err).Error("Duplicate blob id")
		} else {
			break
		}
	}
	if err != nil {
		logger.WithError(err).Error("Unable to create blob")
		return nil, errors.Wrap(err, "unable to create blob")
	}

	blb, err := s.get(logger, id)
	if err != nil {
		return nil, err
	}

	logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("Create")
	return blb, nil
}

func (s *Session) Get(ctx context.Context, id string) (*blob.Blob, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	} else if !blob.IsValidID(id) {
		return nil, errors.New("id is invalid")
	}

	if s.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("id", id)

	blb, err := s.get(logger, id)
	if err != nil {
		return nil, err
	}

	logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("Get")
	return blb, nil
}

func (s *Session) Update(ctx context.Context, id string, update *blobStoreStructured.Update) (*blob.Blob, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	} else if !blob.IsValidID(id) {
		return nil, errors.New("id is invalid")
	}
	if update == nil {
		return nil, errors.New("update is missing")
	} else if err := structureValidator.New().Validate(update); err != nil {
		return nil, errors.Wrap(err, "update is invalid")
	}

	if s.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"id": id, "update": update})

	if update.HasUpdates() {
		set := bson.M{
			"modifiedTime": pointer.FromTime(now.Truncate(time.Second)),
		}
		unset := bson.M{}
		if update.MediaType != nil {
			set["mediaType"] = *update.MediaType
		}
		if update.DigestMD5 != nil {
			set["digestMD5"] = *update.DigestMD5
		}
		if update.Size != nil {
			set["size"] = *update.Size
		}
		if update.Status != nil {
			set["status"] = *update.Status
		}
		changeInfo, err := s.C().UpdateAll(bson.M{"id": id}, s.ConstructUpdate(set, unset))
		if err != nil {
			logger.WithError(err).Error("Unable to update blob")
			return nil, errors.Wrap(err, "unable to update blob")
		}

		logger = logger.WithField("changeInfo", changeInfo)
	}

	blb, err := s.get(logger, id)
	if err != nil {
		return nil, err
	}

	logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("Update")
	return blb, nil
}

func (s *Session) Delete(ctx context.Context, id string) (bool, error) {
	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if id == "" {
		return false, errors.New("id is missing")
	} else if !blob.IsValidID(id) {
		return false, errors.New("id is invalid")
	}

	if s.IsClosed() {
		return false, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("id", id)

	changeInfo, err := s.C().RemoveAll(bson.M{"id": id})
	if err != nil {
		logger.WithError(err).Error("Unable to delete blob")
		return false, errors.Wrap(err, "unable to delete blob")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("Delete")
	return changeInfo.Removed > 0, nil
}

func (s *Session) get(logger log.Logger, id string) (*blob.Blob, error) {
	blbs := blob.Blobs{}
	err := s.C().Find(bson.M{"id": id}).Limit(2).All(&blbs)
	if err != nil {
		logger.WithError(err).Error("Unable to get blob")
		return nil, errors.Wrap(err, "unable to get blob")
	}

	switch len(blbs) {
	case 0:
		return nil, nil
	case 1:
		return blbs[0], nil
	default:
		logger.Error("Multiple blobs found")
		return blbs[0], nil
	}
}
