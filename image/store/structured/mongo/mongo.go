package mongo

import (
	"context"
	"time"

	mgo "github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/image"
	imageStoreStructured "github.com/tidepool-org/platform/image/store/structured"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
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

func (s *Store) NewSession() imageStoreStructured.Session {
	return s.newSession()
}

func (s *Store) newSession() *Session {
	return &Session{
		Session: s.Store.NewSession("images"),
	}
}

type Session struct {
	*storeStructuredMongo.Session
}

func (s *Session) EnsureIndexes() error {
	return s.EnsureAllIndexes([]mgo.Index{
		{Key: []string{"id"}, Background: true, Unique: true},
		{Key: []string{"userId", "status"}, Background: true},
	})
}

func (s *Session) List(ctx context.Context, userID string, filter *image.Filter, pagination *page.Pagination) (image.ImageArray, error) {
	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"userId": userID, "filter": filter, "pagination": pagination})

	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return nil, errors.New("user id is invalid")
	}
	if filter == nil {
		filter = image.NewFilter()
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

	var status []string
	if filter.Status != nil {
		status = *filter.Status
	} else {
		status = []string{image.StatusAvailable}
	}

	result := image.ImageArray{}
	query := bson.M{
		"userId": userID,
		"status": bson.M{
			"$in": status,
		},
		"deletedTime": bson.M{
			"$exists": false,
		},
	}
	if filter.ContentIntent != nil {
		query["contentIntent"] = bson.M{
			"$in": *filter.ContentIntent,
		}
	}
	err := s.C().Find(query).Sort("-createdTime").Skip(pagination.Page * pagination.Size).Limit(pagination.Size).All(&result)
	if err != nil {
		logger.WithError(err).Error("Unable to list images")
		return nil, errors.Wrap(err, "unable to list images")
	}

	logger.WithFields(log.Fields{"count": len(result), "duration": time.Since(now) / time.Microsecond}).Debug("List")
	return result, nil
}

func (s *Session) Create(ctx context.Context, userID string, metadata *image.Metadata) (*image.Image, error) {
	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"userId": userID, "metadata": metadata})

	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return nil, errors.New("user id is invalid")
	}
	if metadata == nil {
		return nil, errors.New("metadata is missing")
	} else if err := structureValidator.New().Validate(metadata); err != nil {
		return nil, errors.Wrap(err, "metadata is invalid")
	}

	if s.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()

	if metadata.IsEmpty() {
		metadata = nil
	}

	doc := &image.Image{
		UserID:      pointer.FromString(userID),
		Status:      pointer.FromString(image.StatusCreated),
		Metadata:    metadata,
		CreatedTime: pointer.FromTime(now.Truncate(time.Millisecond)),
		Revision:    pointer.FromInt(0),
	}

	var id string
	var err error
	for retry := 0; retry < 3; retry++ {
		id = image.NewID()
		logger = logger.WithField("id", id)

		doc.ID = pointer.FromString(id)
		if err = s.C().Insert(doc); mgo.IsDup(err) {
			logger.WithError(err).Error("Duplicate image id")
		} else {
			break
		}
	}
	if err != nil {
		logger.WithError(err).Error("Unable to create image")
		return nil, errors.Wrap(err, "unable to create image")
	}

	result, err := s.get(logger, id, nil)
	if err != nil {
		return nil, err
	}

	logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("Create")
	return result, nil
}

func (s *Session) DeleteAll(ctx context.Context, userID string) (bool, error) {
	ctx, logger := log.ContextAndLoggerWithField(ctx, "userId", userID)

	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if userID == "" {
		return false, errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return false, errors.New("user id is invalid")
	}

	if s.IsClosed() {
		return false, errors.New("session closed")
	}

	now := time.Now()

	query := bson.M{
		"userId": userID,
	}
	set := bson.M{
		"modifiedTime": now.Truncate(time.Millisecond),
		"deletedTime":  now.Truncate(time.Millisecond),
	}
	unset := bson.M{}
	changeInfo, err := s.C().UpdateAll(query, s.ConstructUpdate(set, unset))
	if err != nil {
		logger.WithError(err).Error("Unable to delete all images")
		return false, errors.Wrap(err, "unable to delete all images")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("DeleteAll")
	return changeInfo.Updated > 0, nil
}

func (s *Session) DestroyAll(ctx context.Context, userID string) (bool, error) {
	ctx, logger := log.ContextAndLoggerWithField(ctx, "userId", userID)

	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if userID == "" {
		return false, errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return false, errors.New("user id is invalid")
	}

	if s.IsClosed() {
		return false, errors.New("session closed")
	}

	now := time.Now()

	query := bson.M{
		"userId": userID,
	}
	changeInfo, err := s.C().RemoveAll(query)
	if err != nil {
		logger.WithError(err).Error("Unable to destroy all images")
		return false, errors.Wrap(err, "unable to destroy all images")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("DestroyAll")
	return changeInfo.Removed > 0, nil
}

func (s *Session) Get(ctx context.Context, id string, condition *request.Condition) (*image.Image, error) {
	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"id": id, "condition": condition})

	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	} else if !image.IsValidID(id) {
		return nil, errors.New("id is invalid")
	}
	if condition == nil {
		condition = request.NewCondition()
	} else if err := structureValidator.New().Validate(condition); err != nil {
		return nil, errors.Wrap(err, "condition is invalid")
	}

	if s.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()

	result, err := s.get(logger, id, condition, storeStructuredMongo.NotDeleted)
	if err != nil {
		return nil, err
	}

	logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("Get")
	return result, nil
}

func (s *Session) Update(ctx context.Context, id string, condition *request.Condition, update *imageStoreStructured.Update) (*image.Image, error) {
	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"id": id, "condition": condition, "update": update})

	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	} else if !image.IsValidID(id) {
		return nil, errors.New("id is invalid")
	}
	if condition == nil {
		condition = request.NewCondition()
	} else if err := structureValidator.New().Validate(condition); err != nil {
		return nil, errors.Wrap(err, "condition is invalid")
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

	if !update.IsEmpty() {
		query := bson.M{
			"id": id,
			"deletedTime": bson.M{
				"$exists": false,
			},
		}
		if condition.Revision != nil {
			query["revision"] = *condition.Revision
		}
		set := bson.M{
			"modifiedTime": now.Truncate(time.Millisecond),
		}
		unset := bson.M{}
		min := bson.M{}
		addToSet := bson.M{}
		if metadata := update.Metadata; metadata != nil {
			if metadata.Associations != nil {
				set["metadata.associations"] = metadata.Associations
			}
			if metadata.Location != nil {
				set["metadata.location"] = metadata.Location
			}
			if metadata.Metadata != nil {
				set["metadata.metadata"] = metadata.Metadata
			}
			if metadata.Name != nil {
				set["metadata.name"] = *metadata.Name
			}
			if metadata.Origin != nil {
				set["metadata.origin"] = metadata.Origin
			}
		}
		if update.ContentID != nil {
			set["status"] = image.StatusAvailable
			set["contentId"] = *update.ContentID
			set["contentIntent"] = *update.ContentIntent
			set["contentAttributes.digestMD5"] = *update.ContentAttributes.DigestMD5
			set["contentAttributes.mediaType"] = *update.ContentAttributes.MediaType
			set["contentAttributes.width"] = *update.ContentAttributes.Width
			set["contentAttributes.height"] = *update.ContentAttributes.Height
			set["contentAttributes.size"] = *update.ContentAttributes.Size
			min["contentAttributes.createdTime"] = now.Truncate(time.Millisecond)
			set["contentAttributes.modifiedTime"] = now.Truncate(time.Millisecond)
			unset["renditionsId"] = true
			unset["renditions"] = true
		} else if update.RenditionsID != nil {
			set["renditionsId"] = *update.RenditionsID
			set["renditions"] = []string{*update.Rendition}
		} else if update.Rendition != nil {
			addToSet["renditions"] = *update.Rendition
		}
		changeInfo, err := s.C().UpdateAll(query, s.ConstructUpdate(set, unset, map[string]bson.M{"$min": min, "$addToSet": addToSet}))
		if err != nil {
			logger.WithError(err).Error("Unable to update image")
			return nil, errors.Wrap(err, "unable to update image")
		} else if changeInfo.Updated > 0 {
			condition = nil
		} else {
			update = nil
		}

		logger = logger.WithField("changeInfo", changeInfo)
	}

	var result *image.Image
	if update != nil {
		var err error
		if result, err = s.get(logger, id, condition); err != nil {
			return nil, err
		}
	}

	logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("Update")
	return result, nil
}

func (s *Session) Delete(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"id": id, "condition": condition})

	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if id == "" {
		return false, errors.New("id is missing")
	} else if !image.IsValidID(id) {
		return false, errors.New("id is invalid")
	}
	if condition == nil {
		condition = request.NewCondition()
	} else if err := structureValidator.New().Validate(condition); err != nil {
		return false, errors.Wrap(err, "condition is invalid")
	}

	if s.IsClosed() {
		return false, errors.New("session closed")
	}

	now := time.Now()

	query := bson.M{
		"id": id,
	}
	if condition.Revision != nil {
		query["revision"] = *condition.Revision
	}
	set := bson.M{
		"modifiedTime": now.Truncate(time.Millisecond),
		"deletedTime":  now.Truncate(time.Millisecond),
	}
	unset := bson.M{}
	changeInfo, err := s.C().UpdateAll(query, s.ConstructUpdate(set, unset))
	if err != nil {
		logger.WithError(err).Error("Unable to delete image")
		return false, errors.Wrap(err, "unable to delete image")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("Delete")
	return changeInfo.Updated > 0, nil
}

func (s *Session) Destroy(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"id": id, "condition": condition})

	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if id == "" {
		return false, errors.New("id is missing")
	} else if !image.IsValidID(id) {
		return false, errors.New("id is invalid")
	}
	if condition == nil {
		condition = request.NewCondition()
	} else if err := structureValidator.New().Validate(condition); err != nil {
		return false, errors.Wrap(err, "condition is invalid")
	}

	if s.IsClosed() {
		return false, errors.New("session closed")
	}

	now := time.Now()

	query := bson.M{
		"id": id,
	}
	if condition.Revision != nil {
		query["revision"] = *condition.Revision
	}
	changeInfo, err := s.C().RemoveAll(query)
	if err != nil {
		logger.WithError(err).Error("Unable to destroy image")
		return false, errors.Wrap(err, "unable to destroy image")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("Destroy")
	return changeInfo.Removed > 0, nil
}

func (s *Session) get(logger log.Logger, id string, condition *request.Condition, queryModifiers ...storeStructuredMongo.QueryModifier) (*image.Image, error) {
	logger = logger.WithFields(log.Fields{"id": id, "condition": condition})

	results := image.ImageArray{}
	query := bson.M{
		"id": id,
	}
	if condition != nil && condition.Revision != nil {
		query["revision"] = *condition.Revision
	}
	query = storeStructuredMongo.ModifyQuery(query, queryModifiers...)
	err := s.C().Find(query).Limit(2).All(&results)
	if err != nil {
		logger.WithError(err).Error("Unable to get image")
		return nil, errors.Wrap(err, "unable to get image")
	}

	var result *image.Image
	switch len(results) {
	case 0:
		return nil, nil
	case 1:
		result = results[0]
	default:
		logger.Error("Multiple images found")
		result = results[0]
	}

	if result.Revision == nil {
		result.Revision = pointer.FromInt(0)
	}

	return result, nil
}
