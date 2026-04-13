package mongo

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	bsonPrimitive "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/compress"
	"github.com/tidepool-org/platform/data"
	dataRaw "github.com/tidepool-org/platform/data/raw"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	storeStructured "github.com/tidepool-org/platform/store/structured"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	IDSeparator  = ":"
	IDDateFormat = time.DateOnly

	CompressionOriginalSizeMinimum = 100 // To prevent unnecessary compression
	CompressionFactorMinimum       = 0.1
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
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "dataSetId", Value: 1},
				{Key: "createdTime", Value: 1},
			},
			Options: mongoOptions.Index().
				SetName("UserIdDataSetIdCreatedTime"),
		},
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
		filter = &dataRaw.Filter{}
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
	if filter.DataSetID != nil {
		query["dataSetId"] = *filter.DataSetID
	}
	if filter.Processed != nil {
		query["processedTime"] = bson.M{"$exists": *filter.Processed}
	}
	if filter.Archivable != nil {
		if *filter.Archivable {
			query["archivableTime"] = bson.M{"$exists": true, "$lt": now}
		} else {
			query["$or"] = bson.A{
				bson.M{"archivableTime": bson.M{"$exists": false}},
				bson.M{"archivableTime": bson.M{"$gte": now}},
			}
		}
	}
	if filter.Archived != nil {
		query["archivedTime"] = bson.M{"$exists": *filter.Archived}
	}

	opts := storeStructuredMongo.FindWithPagination(pagination).
		SetSort(bson.M{"createdTime": 1}).
		SetProjection(bson.M{"data": 0})
	documents, err := s.findMany(ctx, query, opts)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("unable to list raw")
		return nil, errors.Wrap(err, "unable to list raw")
	} else if documents == nil {
		return nil, nil
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

	hasherMD5 := md5.New() //nolint:gosec // MD5 acceptable for data integrity checksums, not used for security
	hasherSHA256 := sha256.New()

	// Setup readers
	originalLimitReader := io.LimitReader(reader, dataRaw.SizeMaximum+1)                                       // Limit original size to prevent abuse
	originalSizeReader := compress.SizeReader(originalLimitReader)                                             // Capture original size
	originalHeadReader := compress.HeadReader(originalSizeReader, dataRaw.SizeStoredMaximum)                   // Capture original head to use if compression not warranted
	hasherMD5Reader := io.TeeReader(originalHeadReader, hasherMD5)                                             // Calculate MD5 digest of original data
	hasherSHA256Reader := io.TeeReader(hasherMD5Reader, hasherSHA256)                                          // Calculate SHA256 digest of original data
	compressedLimitReader := compress.LimitCompressReadCloser(hasherSHA256Reader, dataRaw.SizeStoredMaximum+1) // Limit compressed size to prevent abuse

	// Read all compressed data
	compressedData, err := io.ReadAll(compressedLimitReader)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("unable to read compressed data")
		return nil, errors.Wrap(err, "unable to read compressed data")
	}

	// TODO: BACK-3629 - Respond with HTTP 400 Bad Request when raw data request body exceeds maximum size

	// Ensure original size is valid
	originalSize := originalSizeReader.Size()
	if originalSize > dataRaw.SizeMaximum {
		lgr.Error("data size exceeds maximum allowed size")
		return nil, errors.New("data size exceeds maximum allowed size")
	}

	// Ensure compressed size is valid
	compressedSize := len(compressedData)
	if compressedSize > dataRaw.SizeStoredMaximum {
		lgr.Error("compressed data size exceeds maximum allowed size")
		return nil, errors.New("compressed data size exceeds maximum allowed size")
	}

	// TODO: BACK-3630 - Respond with HTTP 400 Bad Request when raw data request-specified MD5 digest does not match calculated

	// Ensure MD5 digest is valid
	digestMD5 := base64.StdEncoding.EncodeToString(hasherMD5.Sum(nil))
	if create.DigestMD5 != nil && *create.DigestMD5 != digestMD5 {
		lgr.Error("calculated MD5 digest does not match expected")
		return nil, errors.New("calculated MD5 digest does not match expected")
	}

	// Ensure SHA256 digest is valid
	digestSHA256 := base64.StdEncoding.EncodeToString(hasherSHA256.Sum(nil))
	if create.DigestSHA256 != nil && *create.DigestSHA256 != digestSHA256 {
		lgr.Error("calculated SHA256 digest does not match expected")
		return nil, errors.New("calculated SHA256 digest does not match expected")
	}

	// Use compressed data if compression is significant or original data cannot be stored directly
	compressed := originalSize > dataRaw.SizeStoredMaximum ||
		(originalSize >= CompressionOriginalSizeMinimum && compressionFactor(originalSize, int64(compressedSize)) >= CompressionFactorMinimum)

	var data bsonPrimitive.Binary
	if compressed {
		data = bsonPrimitive.Binary{Data: compressedData}
	} else {
		data = bsonPrimitive.Binary{Data: originalHeadReader.Bytes()}
	}

	document := &Document{
		UserID:         userID,
		DataSetID:      dataSetID,
		Metadata:       create.Metadata,
		DigestMD5:      digestMD5,
		DigestSHA256:   pointer.From(digestSHA256),
		MediaType:      pointer.DefaultString(create.MediaType, dataRaw.MediaTypeDefault),
		Size:           int(originalSize),
		Compressed:     compressed,
		Data:           data,
		ArchivableTime: create.ArchivableTime,
		CreatedTime:    now,
		Revision:       1,
	}

	ctx, lgr = log.ContextAndLoggerWithField(ctx, "document", document)

	// From this point forward, the context should not be cancelable
	ctx = context.WithoutCancel(ctx)

	result, err := s.InsertOne(ctx, document)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("unable to create raw")
		return nil, errors.Wrap(err, "unable to create raw")
	}

	document.ID = result.InsertedID.(bsonPrimitive.ObjectID)

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

	opts := mongoOptions.FindOne().
		SetProjection(bson.M{"data": 0})
	document, err := s.findOne(ctx, query, opts)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("unable to get raw")
		return nil, errors.Wrap(err, "unable to get raw")
	} else if document == nil {
		return nil, nil
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
	defer func() { lgr.WithField("duration", time.Since(now)/time.Microsecond).Debug("GetContent") }()

	query := bson.M{"_id": objectID}
	if condition.Revision != nil {
		query["revision"] = *condition.Revision
	}

	document, err := s.findOne(ctx, query)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("unable to get content")
		return nil, errors.Wrap(err, "unable to get content")
	} else if document == nil {
		return nil, nil
	}

	return document.AsContent(), nil
}

func (s *Store) Update(ctx context.Context, id string, condition *storeStructured.Condition, update *dataRaw.Update) (*dataRaw.Raw, error) {
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
	if update == nil {
		return nil, errors.New("update is missing")
	} else if err = structureValidator.New(log.LoggerFromContext(ctx)).Validate(update); err != nil {
		return nil, errors.Wrap(err, "update is invalid")
	}

	ctx, lgr := log.ContextAndLoggerWithFields(ctx, log.Fields{"id": id, "condition": condition, "update": update})

	now := time.Now()
	defer func() { lgr.WithField("duration", time.Since(now)/time.Microsecond).Debug("Update") }()

	query := bson.M{"_id": objectID}
	if condition.Revision != nil {
		query["revision"] = *condition.Revision
	}

	set := bson.M{
		"modifiedTime": now,
	}
	if update.ProcessedTime != nil {
		set["processedTime"] = *update.ProcessedTime
	}
	if update.ArchivableTime != nil {
		set["archivableTime"] = *update.ArchivableTime
	}
	if update.ArchivedTime != nil {
		set["archivedTime"] = *update.ArchivedTime
	}
	if update.Metadata != nil {
		set["metadata"] = *update.Metadata
	}

	// From this point forward, the context should not be cancelable
	ctx = context.WithoutCancel(ctx)

	var document *Document
	err = s.FindOneAndUpdate(ctx, query, s.ConstructUpdate(set, nil)).Decode(&document)
	lgr = lgr.WithError(err)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		} else {
			lgr.Error("unable to update raw")
			return nil, errors.Wrap(err, "unable to update raw")
		}
	}

	query = bson.M{"_id": objectID}
	document, err = s.findOne(ctx, query)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("unable to get raw after update")
		return nil, errors.Wrap(err, "unable to get raw after update")
	} else if document == nil {
		return nil, nil
	}

	return document.AsRaw(), nil
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
	} else if err = structureValidator.New(log.LoggerFromContext(ctx)).Validate(condition); err != nil {
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
	err = s.FindOneAndDelete(ctx, query, mongoOptions.FindOneAndDelete().SetProjection(bson.M{"data": 0})).Decode(&document)
	lgr = lgr.WithError(err)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		} else {
			lgr.Error("unable to delete raw")
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
		lgr.Error("unable to delete multiple raw")
		return 0, errors.Wrap(err, "unable to delete multiple raw")
	}

	lgr.WithField("count", deleteResult.DeletedCount)
	return int(deleteResult.DeletedCount), nil
}

func (s *Store) DeleteAllByDataSetID(ctx context.Context, userID string, dataSetID string) (int, error) {
	if ctx == nil {
		return 0, errors.New("context is missing")
	}
	if userID == "" {
		return 0, errors.New("user id is missing")
	} else if !auth.IsValidUserID(userID) {
		return 0, errors.New("user id is invalid")
	}
	if dataSetID == "" {
		return 0, errors.New("data set id is missing")
	} else if !data.IsValidSetID(dataSetID) {
		return 0, errors.New("data set id is invalid")
	}

	ctx, lgr := log.ContextAndLoggerWithFields(ctx, log.Fields{"userId": userID, "dataSetId": dataSetID})

	now := time.Now()
	defer func() { lgr.WithField("duration", time.Since(now)/time.Microsecond).Debug("DeleteAllByDataSetID") }()

	query := bson.M{
		"userId":    userID,
		"dataSetId": dataSetID,
	}

	deleteResult, err := s.DeleteMany(ctx, query)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("unable to delete all by data set id raw")
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
		lgr.Error("unable to delete all by user id raw")
		return 0, errors.Wrap(err, "unable to delete all by user id raw")
	}

	lgr.WithField("count", deleteResult.DeletedCount)
	return int(deleteResult.DeletedCount), nil
}

func (s *Store) findOne(ctx context.Context, query bson.M, opts ...*mongoOptions.FindOneOptions) (*Document, error) {
	result := s.FindOne(ctx, query, opts...)
	if err := result.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	var document *Document
	if err := result.Decode(&document); err != nil {
		return nil, err
	}

	return document, nil
}

func (s *Store) findMany(ctx context.Context, query bson.M, opts ...*mongoOptions.FindOptions) (Documents, error) {
	cursor, err := s.Find(ctx, query, opts...)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	var documents Documents
	if err = cursor.All(ctx, &documents); err != nil {
		return nil, err
	}

	return documents, nil
}

type Document struct {
	ID             bsonPrimitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"` // Database only
	UserID         string                 `json:"userId,omitempty" bson:"userId,omitempty"`
	DataSetID      string                 `json:"dataSetId,omitempty" bson:"dataSetId,omitempty"`
	Metadata       map[string]any         `json:"metadata,omitempty" bson:"metadata,omitempty"`
	DigestMD5      string                 `json:"digestMD5,omitempty" bson:"digestMD5,omitempty"`
	DigestSHA256   *string                `json:"digestSHA256,omitempty" bson:"digestSHA256,omitempty"` // FUTURE: Optional until data migrated
	MediaType      string                 `json:"mediaType,omitempty" bson:"mediaType,omitempty"`
	Size           int                    `json:"size,omitempty" bson:"size,omitempty"`
	Compressed     bool                   `json:"compressed,omitempty" bson:"compressed,omitempty"`
	Data           bsonPrimitive.Binary   `json:"-" bson:"data,omitempty"`
	ProcessedTime  *time.Time             `json:"processedTime,omitempty" bson:"processedTime,omitempty"`
	ArchivableTime *time.Time             `json:"archivableTime,omitempty" bson:"archivableTime,omitempty"`
	ArchivedTime   *time.Time             `json:"archivedTime,omitempty" bson:"archivedTime,omitempty"`
	CreatedTime    time.Time              `json:"createdTime,omitzero" bson:"createdTime,omitempty"`
	ModifiedTime   *time.Time             `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	Revision       int                    `json:"revision,omitempty" bson:"revision,omitempty"`
}

func (d *Document) AsRaw() *dataRaw.Raw {
	return &dataRaw.Raw{
		ID:             idFromObjectIDAndDate(d.ID, d.CreatedTime),
		UserID:         d.UserID,
		DataSetID:      d.DataSetID,
		Metadata:       d.Metadata,
		DigestMD5:      d.DigestMD5,
		DigestSHA256:   d.DigestSHA256,
		MediaType:      d.MediaType,
		Size:           d.Size,
		ProcessedTime:  d.ProcessedTime,
		ArchivableTime: d.ArchivableTime,
		ArchivedTime:   d.ArchivedTime,
		CreatedTime:    d.CreatedTime,
		ModifiedTime:   d.ModifiedTime,
		Revision:       d.Revision,
	}
}

func (d *Document) AsContent() *dataRaw.Content {
	readCloser := io.NopCloser(bytes.NewReader(d.Data.Data))
	if d.Compressed {
		readCloser = compress.DecompressReadCloser(readCloser)
	}
	return &dataRaw.Content{
		DigestMD5:    d.DigestMD5,
		DigestSHA256: d.DigestSHA256,
		MediaType:    d.MediaType,
		ReadCloser:   readCloser,
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

func objectIDsFromIDs(ids []string) ([]bsonPrimitive.ObjectID, error) {
	if ids == nil {
		return nil, nil
	}
	objectIDs := make([]bsonPrimitive.ObjectID, len(ids))
	for index, id := range ids {
		if objectID, _, err := objectIDAndDateFromID(id); err != nil {
			return nil, err
		} else {
			objectIDs[index] = objectID
		}
	}
	return objectIDs, nil
}

func objectIDAndDateFromID(id string) (bsonPrimitive.ObjectID, time.Time, error) {
	if id == "" {
		return bsonPrimitive.NilObjectID, time.Time{}, errors.New("id is missing")
	} else if parts := strings.SplitN(id, ":", 2); len(parts) != 2 {
		return bsonPrimitive.NilObjectID, time.Time{}, errors.New("id is invalid")
	} else if objectID, err := bsonPrimitive.ObjectIDFromHex(parts[0]); err != nil {
		return bsonPrimitive.NilObjectID, time.Time{}, errors.New("id is invalid")
	} else if date, err := time.Parse(IDDateFormat, parts[1]); err != nil {
		return bsonPrimitive.NilObjectID, time.Time{}, errors.New("id is invalid")
	} else {
		return objectID, date, nil
	}
}

func idFromObjectIDAndDate(objectID bsonPrimitive.ObjectID, date time.Time) string {
	return strings.Join([]string{objectID.Hex(), date.Format(IDDateFormat)}, IDSeparator)
}

func compressionFactor(originalSize int64, compressedSize int64) float64 {
	if originalSize == 0 {
		return 1.0
	}
	return 1.0 - float64(compressedSize)/float64(originalSize)
}
