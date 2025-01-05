package mongo

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"io"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/raw"
	dataRaw "github.com/tidepool-org/platform/data/raw"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	storeStructured "github.com/tidepool-org/platform/store/structured"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	IDSeparator  = ":"
	IDDateFormat = time.DateOnly
)

type Store struct {
	*storeStructuredMongo.Store
	*storeStructuredMongo.Repository
}

func NewStore(config *storeStructuredMongo.Config) (*Store, error) {
	store, err := storeStructuredMongo.NewStore(config)
	if err != nil {
		return nil, err
	}
	return &Store{
		Store:      store,
		Repository: store.GetRepository("raw"),
	}, nil
}

func (s *Store) EnsureIndexes() error {
	return s.CreateAllIndexes(context.Background(), []mongo.IndexModel{
		// TODO
		// {
		// 	Keys:    bson.D{{Key: "_id", Value: 1}, {Key: "revision", Value: 1}},
		// 	Options: options.Index().SetUnique(true),
		// },
		// {
		// 	Keys:    bson.D{{Key: "name", Value: 1}},
		// 	Options: options.Index().SetUnique(true).SetSparse(true),
		// },
		// {
		// 	Keys:    bson.D{{Key: "priority", Value: 1}},
		// 	Options: options.Index(),
		// },
		// {
		// 	Keys:    bson.D{{Key: "availableTime", Value: 1}},
		// 	Options: options.Index(),
		// },
		// {
		// 	Keys:    bson.D{{Key: "expirationTime", Value: 1}},
		// 	Options: options.Index(),
		// },
		// {
		// 	Keys:    bson.D{{Key: "state", Value: 1}},
		// 	Options: options.Index(),
		// },
	})
}

func (s *Store) List(ctx context.Context, userID string, filter *dataRaw.Filter, pagination *page.Pagination) ([]*dataRaw.Raw, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	} else if !auth.IsValidUserID(userID) {
		return nil, errors.New("user id is invalid")
	}
	if filter == nil {
		filter = dataRaw.NewFilter()
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	ctx, lgr := log.ContextAndLoggerWithFields(ctx, log.Fields{"userId": userID, "filter": filter, "pagination": pagination})

	now := time.Now()
	defer func() { lgr.WithField("duration", time.Since(now)/time.Microsecond).Debug("List") }()

	query := bson.M{"userId": userID}
	if createdTime := filter.CreatedTime(); createdTime != nil {
		query["createdTime"] = bson.M{
			"$gte": createdTime,
			"$lt":  createdTime.AddDate(0, 0, 1),
		}
	}
	if filter.DataSetIDs != nil {
		query["dataSetId"] = bson.M{"$in": *filter.DataSetIDs}
	}

	opts := storeStructuredMongo.
		FindWithPagination(pagination).
		SetProjection(bson.M{"data": 0}).
		SetSort(bson.M{"createdTime": 1})
	cursor, err := s.Find(ctx, query, opts)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("Unable to list raw")
		return nil, errors.Wrap(err, "unable to list raw")
	}

	var documents Documents
	err = cursor.All(ctx, &documents)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("Unable to decode raw")
		return nil, errors.Wrap(err, "unable to decode raw")
	}

	lgr = lgr.WithField("count", len(documents))
	return documents.AsRaw(), nil
}

func (s *Store) Create(ctx context.Context, userID string, dataSetID string, create *dataRaw.Create, reader io.Reader) (*dataRaw.Raw, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	} else if !auth.IsValidUserID(userID) {
		return nil, errors.New("user id is invalid")
	}
	if dataSetID == "" {
		return nil, errors.New("data set id is missing")
	} else if !data.IsValidSetID(dataSetID) {
		return nil, errors.New("data set id is invalid")
	}
	if create == nil {
		return nil, errors.New("create is missing")
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(create); err != nil {
		return nil, errors.Wrap(err, "create is invalid")
	}
	if reader == nil {
		return nil, errors.New("reader is missing")
	}

	ctx, lgr := log.ContextAndLoggerWithFields(ctx, log.Fields{"userId": userID, "dataSetId": dataSetID, "create": create})

	now := time.Now()
	defer func() { lgr.WithField("duration", time.Since(now)/time.Microsecond).Debug("Create") }()

	hasher := md5.New()
	data, err := io.ReadAll(io.TeeReader(io.LimitReader(reader, raw.DataSizeMaximum+1), hasher))
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("Unable to read data")
		return nil, errors.Wrap(err, "unable to read data")
	}

	// TODO: How to report this back as a Bad Request

	size := len(data)
	if size > raw.DataSizeMaximum {
		lgr.Error("Data size exceeds maximum allowed size")
		return nil, errors.New("data size exceeds maximum allowed size")
	}

	// TODO: How to report this back as a Bad Request

	digestMD5 := base64.StdEncoding.EncodeToString(hasher.Sum(nil))
	if create.DigestMD5 != nil && *create.DigestMD5 != digestMD5 {
		lgr.Error("Calculated MD5 digest does not match expected")
		return nil, errors.New("calculated MD5 digest does not match expected")
	}

	document := &Document{
		UserID:      userID,
		DataSetID:   dataSetID,
		Metadata:    create.Metadata,
		DigestMD5:   digestMD5,
		MediaType:   *pointer.DefaultString(create.MediaType, raw.MediaTypeDefault),
		Size:        size,
		Data:        primitive.Binary{Data: data},
		CreatedTime: now,
		Revision:    1,
	}

	ctx, lgr = log.ContextAndLoggerWithField(ctx, "document", document)

	// From this point forward, the context should not be cancelable
	ctx = context.WithoutCancel(ctx)

	result, err := s.InsertOne(ctx, document)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("Unable to create raw")
		return nil, errors.Wrap(err, "unable to create raw")
	}

	document.ID = result.InsertedID.(primitive.ObjectID)
	lgr = lgr.WithField("document", document)

	return document.AsRaw(), nil
}

func (s *Store) Get(ctx context.Context, id string, condition *storeStructured.Condition) (*dataRaw.Raw, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	objectID, _, err := objectIDAndDateFromID(id)
	if err != nil {
		return nil, err
	}
	if condition == nil {
		condition = storeStructured.NewCondition()
	} else if err = structureValidator.New(log.LoggerFromContext(ctx)).Validate(condition); err != nil {
		return nil, errors.Wrap(err, "condition is invalid")
	}

	ctx, lgr := log.ContextAndLoggerWithFields(ctx, log.Fields{"id": id, "condition": condition})

	now := time.Now()
	defer func() { lgr.WithField("duration", time.Since(now)/time.Microsecond).Debug("Get") }()

	query := bson.M{"_id": objectID}
	if condition.Revision != nil {
		query["revision"] = *condition.Revision
	}

	document, err := s.get(ctx, query, bson.M{"data": 0})
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("Unable to get raw")
		return nil, err
	}

	return document.AsRaw(), nil
}

func (s *Store) GetContent(ctx context.Context, id string, condition *storeStructured.Condition) (*dataRaw.Content, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	objectID, _, err := objectIDAndDateFromID(id)
	if err != nil {
		return nil, err
	}
	if condition == nil {
		condition = storeStructured.NewCondition()
	} else if err = structureValidator.New(log.LoggerFromContext(ctx)).Validate(condition); err != nil {
		return nil, errors.Wrap(err, "condition is invalid")
	}

	ctx, lgr := log.ContextAndLoggerWithFields(ctx, log.Fields{"id": id, "condition": condition})

	now := time.Now()
	defer func() { lgr.WithField("duration", time.Since(now)/time.Microsecond).Debug("Get") }()

	query := bson.M{"_id": objectID}
	if condition.Revision != nil {
		query["revision"] = *condition.Revision
	}

	document, err := s.get(ctx, query, nil)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("Unable to get raw")
		return nil, err
	}

	return document.AsContent(), nil
}

func (s *Store) Delete(ctx context.Context, id string, condition *storeStructured.Condition) (*dataRaw.Raw, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	objectID, _, err := objectIDAndDateFromID(id)
	if err != nil {
		return nil, err
	}
	if condition == nil {
		condition = storeStructured.NewCondition()
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(condition); err != nil {
		return nil, errors.Wrap(err, "condition is invalid")
	}

	ctx, lgr := log.ContextAndLoggerWithFields(ctx, log.Fields{"id": id, "condition": condition})

	now := time.Now()
	defer func() { lgr.WithField("duration", time.Since(now)/time.Microsecond).Debug("Delete") }()

	query := bson.M{"_id": objectID}
	if condition.Revision != nil {
		query["revision"] = *condition.Revision
	}

	var document *Document
	err = s.FindOneAndDelete(ctx, query, options.FindOneAndDelete().SetProjection(bson.M{"data": 0})).Decode(&document)
	lgr = lgr.WithError(err)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		} else {
			lgr.Error("Unable to delete raw")
			return nil, errors.Wrap(err, "unable to delete raw")
		}
	}

	return document.AsRaw(), nil
}

func (s *Store) DeleteMultiple(ctx context.Context, ids []string) (int, error) {
	if ctx == nil {
		return 0, errors.New("context is missing")
	}
	objectIDs, err := objectIDsFromIDs(ids)
	if err != nil {
		return 0, err
	}

	ctx, lgr := log.ContextAndLoggerWithField(ctx, "ids", ids)

	now := time.Now()
	defer func() { lgr.WithField("duration", time.Since(now)/time.Microsecond).Debug("DeleteMultiple") }()

	query := bson.M{"_id": bson.M{"$in": objectIDs}}

	deleteResult, err := s.DeleteMany(ctx, query)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("Unable to delete multiple raw")
		return 0, errors.Wrap(err, "unable to delete multiple raw")
	}

	lgr.WithField("count", deleteResult.DeletedCount)
	return int(deleteResult.DeletedCount), nil
}

func (s *Store) DeleteAllByDataSetID(ctx context.Context, dataSetID string) (int, error) {
	if ctx == nil {
		return 0, errors.New("context is missing")
	}
	if dataSetID == "" {
		return 0, errors.New("data set id is missing")
	} else if !data.IsValidSetID(dataSetID) {
		return 0, errors.New("data set id is invalid")
	}

	ctx, lgr := log.ContextAndLoggerWithField(ctx, "dataSetId", dataSetID)

	now := time.Now()
	defer func() { lgr.WithField("duration", time.Since(now)/time.Microsecond).Debug("DeleteAllByDataSetID") }()

	query := bson.M{"dataSetId": dataSetID}

	deleteResult, err := s.DeleteMany(ctx, query)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("Unable to delete all by data set id raw")
		return 0, errors.Wrap(err, "unable to delete all by data set id raw")
	}

	lgr.WithField("count", deleteResult.DeletedCount)
	return int(deleteResult.DeletedCount), nil
}

func (s *Store) DeleteAllByUserID(ctx context.Context, userID string) (int, error) {
	if ctx == nil {
		return 0, errors.New("context is missing")
	}
	if userID == "" {
		return 0, errors.New("user id is missing")
	} else if !auth.IsValidUserID(userID) {
		return 0, errors.New("user id is invalid")
	}

	ctx, lgr := log.ContextAndLoggerWithField(ctx, "userId", userID)

	now := time.Now()
	defer func() { lgr.WithField("duration", time.Since(now)/time.Microsecond).Debug("DeleteAllByUserID") }()

	query := bson.M{"userId": userID}

	deleteResult, err := s.DeleteMany(ctx, query)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("Unable to delete all by user id raw")
		return 0, errors.Wrap(err, "unable to delete all by user id raw")
	}

	lgr.WithField("count", deleteResult.DeletedCount)
	return int(deleteResult.DeletedCount), nil
}

func (s *Store) get(ctx context.Context, query bson.M, projection bson.M) (*Document, error) {
	var document *Document
	if err := s.FindOne(ctx, query, options.FindOne().SetProjection(projection)).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		} else {
			return nil, errors.Wrap(err, "unable to get raw")
		}
	}
	return document, nil
}

type Document struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserID       string             `json:"userId,omitempty" bson:"userId,omitempty"`
	DataSetID    string             `json:"dataSetId,omitempty" bson:"dataSetId,omitempty"`
	Metadata     *metadata.Metadata `json:"metadata,omitempty" bson:"metadata,omitempty"`
	DigestMD5    string             `json:"digestMD5,omitempty" bson:"digestMD5,omitempty"`
	MediaType    string             `json:"mediaType,omitempty" bson:"mediaType,omitempty"`
	Size         int                `json:"size,omitempty" bson:"size,omitempty"`
	Data         primitive.Binary   `json:"-" bson:"data,omitempty"`
	CreatedTime  time.Time          `json:"createdTime,omitempty" bson:"createdTime,omitempty"`
	ModifiedTime *time.Time         `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	Revision     int                `json:"revision,omitempty" bson:"revision,omitempty"`
}

func (d *Document) AsRaw() *dataRaw.Raw {
	return &dataRaw.Raw{
		ID:           idFromObjectIDAndDate(d.ID, d.CreatedTime),
		UserID:       d.UserID,
		DataSetID:    d.DataSetID,
		Metadata:     d.Metadata,
		DigestMD5:    d.DigestMD5,
		MediaType:    d.MediaType,
		Size:         d.Size,
		CreatedTime:  d.CreatedTime,
		ModifiedTime: d.ModifiedTime,
		Revision:     d.Revision,
	}
}

func (d *Document) AsContent() *dataRaw.Content {
	return &dataRaw.Content{
		DigestMD5:  d.DigestMD5,
		MediaType:  d.MediaType,
		ReadCloser: io.NopCloser(bytes.NewReader(d.Data.Data)),
	}
}

type Documents []*Document

func (d Documents) AsRaw() []*dataRaw.Raw {
	rws := make([]*dataRaw.Raw, len(d))
	for index, document := range d {
		rws[index] = document.AsRaw()
	}
	return rws
}

func objectIDsFromIDs(ids []string) ([]primitive.ObjectID, error) {
	if ids == nil {
		return nil, nil
	}
	objectIDs := make([]primitive.ObjectID, len(ids))
	for index, id := range ids {
		if objectID, _, err := objectIDAndDateFromID(id); err != nil {
			return nil, err
		} else {
			objectIDs[index] = objectID
		}
	}
	return objectIDs, nil
}

func objectIDAndDateFromID(id string) (primitive.ObjectID, time.Time, error) {
	if id == "" {
		return primitive.NilObjectID, time.Time{}, errors.New("id is missing")
	} else if parts := strings.SplitN(id, ":", 2); len(parts) != 2 {
		return primitive.NilObjectID, time.Time{}, errors.New("id is invalid")
	} else if objectID, err := primitive.ObjectIDFromHex(parts[0]); err != nil {
		return primitive.NilObjectID, time.Time{}, errors.New("id is invalid")
	} else if date, err := time.Parse(IDDateFormat, parts[1]); err != nil {
		return primitive.NilObjectID, time.Time{}, errors.New("id is invalid")
	} else {
		return objectID, date, nil
	}
}

func idFromObjectIDAndDate(objectID primitive.ObjectID, date time.Time) string {
	return strings.Join([]string{objectID.String(), date.Format(IDDateFormat)}, IDSeparator)
}
