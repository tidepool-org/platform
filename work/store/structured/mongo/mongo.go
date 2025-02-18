package mongo

import (
	"context"
	"maps"
	"slices"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	bsonPrimitive "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	storeStructured "github.com/tidepool-org/platform/store/structured"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/work"
)

func NewStore(config *storeStructuredMongo.Config) (*Store, error) {
	store, err := storeStructuredMongo.NewStore(config)
	if err != nil {
		return nil, err
	}
	return &Store{
		Store:      store,
		Repository: store.GetRepository("work"),
	}, nil
}

type Store struct {
	*storeStructuredMongo.Store
	*storeStructuredMongo.Repository
}

func (s *Store) EnsureIndexes() error {
	return s.CreateAllIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "type", Value: 1},
				{Key: "state", Value: 1},
			},
			Options: mongoOptions.Index().
				SetName("TypeState"),
		},
		{
			Keys: bson.D{
				{Key: "type", Value: 1},
				{Key: "deduplicationId", Value: 1},
			},
			Options: mongoOptions.Index().
				SetName("TypeDeduplicationIdUnique").
				SetPartialFilterExpression(bson.D{{Key: "deduplicationId", Value: bson.M{"$exists": true}}}).
				SetUnique(true),
		},
		// TODO: Test performance, add appropriate indexes
	})
}

func (s *Store) Poll(ctx context.Context, poll *work.Poll) ([]*work.Work, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if poll == nil {
		return nil, errors.New("poll is missing")
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(poll); err != nil {
		return nil, errors.Wrap(err, "poll is invalid")
	}

	ctx, lgr := log.ContextAndLoggerWithField(ctx, "poll", poll)

	now := time.Now()
	defer func() { lgr.WithField("duration", time.Since(now)/time.Microsecond).Debug("Poll") }()

	types := slices.Sorted(maps.Keys(poll.TypeQuantities))

	var pipeline bson.A

	// Match only requested type(s)
	if len(types) == 1 {
		pipeline = append(pipeline, bson.M{"$match": bson.M{"type": types[0]}})
	} else {
		pipeline = append(pipeline, bson.M{"$match": bson.M{"type": bson.M{"$in": types}}})
	}

	// Match either pending available OR processsing/failing with serial id (to remove pending available with matching serial id later)
	pipeline = append(pipeline, bson.M{"$match": bson.M{"$or": bson.A{
		bson.M{"state": "pending", "processingAvailableTime": bson.M{"$lte": now}},
		bson.M{"state": "processing", "serialId": bson.M{"$exists": true}},
		bson.M{"state": "failing", "failingRetryTime": bson.M{"$lte": now}},
		bson.M{"state": "failing", "serialId": bson.M{"$exists": true}},
	}}})

	// Sort by processing priority and available time
	pipeline = append(pipeline, bson.M{"$sort": bson.D{bson.E{Key: "processingPriority", Value: -1}, bson.E{Key: "processingAvailableTime", Value: 1}}})

	// Group all documents by serial id
	pipeline = append(pipeline, bson.M{"$group": bson.M{"_id": "$serialId", "documents": bson.M{"$push": "$$ROOT"}}})

	// Match any without a serial id or any serial id that does not have one in state processsing or failing with retry time in future
	pipeline = append(pipeline, bson.M{"$match": bson.M{"$or": bson.A{
		bson.M{"_id": bson.M{"$exists": false}},
		bson.M{"$nor": bson.A{
			bson.M{"documents.0.state": "processing"},
			bson.M{"documents.0.state": "failing", "documents.0.failingRetryTime": bson.M{"gt": now}},
		}},
	}}})

	// Only allow one per serial id, or all if no serial id
	pipeline = append(pipeline, bson.M{"$project": bson.M{"documents": bson.M{"$cond": bson.M{
		"if":   bson.M{"$ifNull": bson.A{"$_id", false}},
		"then": bson.M{"$first": "$documents"},
		"else": "$documents",
	}}}})

	// Ungroup all documents
	pipeline = append(pipeline, bson.M{"$unwind": "$documents"})
	pipeline = append(pipeline, bson.M{"$replaceRoot": bson.M{"newRoot": "$documents"}})

	// Sort by processing priority and available time
	pipeline = append(pipeline, bson.M{"$sort": bson.D{bson.E{Key: "processingPriority", Value: -1}, bson.E{Key: "processingAvailableTime", Value: 1}}})

	// If one type, then just simple limit
	// Otherwise, group by type, limit each group by type quantity, and ungroup
	if len(types) == 1 {
		pipeline = append(pipeline, bson.M{"$limit": poll.TypeQuantities[types[0]]})
	} else {

		// Limit each type by quantity
		var branches bson.A
		for _, typ := range types {
			branches = append(branches, bson.M{
				// If matches type
				"case": bson.M{"$eq": bson.A{"$_id", typ}},
				// Then limit to quantity requested
				"then": bson.M{"$firstN": bson.M{"input": "$documents", "n": poll.TypeQuantities[typ]}},
			})
		}

		// Group all documents by type
		pipeline = append(pipeline, bson.M{"$group": bson.M{"_id": "$type", "documents": bson.M{"$push": "$$ROOT"}}})

		// Only capture requested quantity per type (using switch branches)
		pipeline = append(pipeline, bson.M{"$project": bson.M{"documents": bson.M{"$switch": bson.M{"branches": branches}}}})

		// Ungroup all documents
		pipeline = append(pipeline, bson.M{"$unwind": "$documents"})
		pipeline = append(pipeline, bson.M{"$replaceRoot": bson.M{"newRoot": "$documents"}})
	}

	// Only need _id and revision
	pipeline = append(pipeline, bson.M{"$project": bson.M{"_id": true, "revision": true}})

	lgr.WithField("pipeline", pipeline).Debug("Poll")

	// Perform aggregation

	cursor, err := s.Aggregate(ctx, pipeline)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.WithField("pipeline", pipeline).Error("Unable to aggregate poll work")
		return nil, errors.Wrap(err, "unable to aggregate poll work")
	}
	defer cursor.Close(ctx)

	// Get identifiers (_id and revision), if nothing, then bail
	var idAndRevisions []bson.M
	err = cursor.All(ctx, &idAndRevisions)
	lgr = lgr.WithError(err)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		} else {
			lgr.Error("Unable to get all poll work")
			return nil, errors.Wrap(err, "unable to get all poll work")
		}
	} else if len(idAndRevisions) == 0 {
		return nil, nil
	}

	modifiedBatchID := id.Must(id.New(16))

	// Build query and update
	query := bson.M{"$or": idAndRevisions}
	update := bson.A{
		bson.M{"$set": bson.M{
			"processingTime":        now,
			"processingTimeoutTime": bson.M{"$dateAdd": bson.M{"startDate": now, "unit": "second", "amount": "$processingTimeout"}},
			"state":                 work.StateProcessing,
			"modifiedTime":          now,
			"modifiedBatchId":       modifiedBatchID,
			"revision":              bson.M{"$add": bson.A{"$revision", 1}},
		}},
		bson.M{"$unset": bson.A{"processingDuration"}},
	}

	// From this point forward, the context should not be cancelable
	ctx = context.WithoutCancel(ctx)

	// Update, if nothing, then bail
	updateResult, err := s.UpdateMany(ctx, query, update)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("Unable to update all poll work")
		return nil, errors.Wrap(err, "unable to update all poll work")
	} else if updateResult.MatchedCount == 0 || updateResult.ModifiedCount == 0 {
		return nil, nil
	}

	// Return documents that were updated
	query = bson.M{"modifiedBatchId": modifiedBatchID}
	opts := mongoOptions.Find().
		SetSort(bson.M{"processingAvailableTime": 1})
	documents, err := s.findMany(ctx, query, opts)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("Unable to list poll work")
		return nil, errors.Wrap(err, "unable to list poll work")
	} else if documents == nil {
		return nil, nil
	}

	lgr = lgr.WithField("count", len(documents))
	return documents.AsWork(), nil
}

func (s *Store) List(ctx context.Context, filter *work.Filter, pagination *page.Pagination) ([]*work.Work, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if filter == nil {
		return nil, errors.New("filter is missing")
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	ctx, lgr := log.ContextAndLoggerWithFields(ctx, log.Fields{"filter": filter, "pagination": pagination})

	now := time.Now()
	defer func() { lgr.WithField("duration", time.Since(now)/time.Microsecond).Debug("List") }()

	query := bson.M{}
	if filter.Types != nil {
		query["type"] = bson.M{"$in": *filter.Types}
	}
	if filter.GroupID != nil {
		query["groupId"] = *filter.GroupID
	}

	opts := storeStructuredMongo.FindWithPagination(pagination).
		SetSort(bson.M{"createdTime": 1})
	documents, err := s.findMany(ctx, query, opts)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("Unable to list work")
		return nil, errors.Wrap(err, "unable to list work")
	} else if documents == nil {
		return nil, nil
	}

	lgr = lgr.WithField("count", len(documents))
	return documents.AsWork(), nil
}

func (s *Store) Create(ctx context.Context, create *work.Create) (*work.Work, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if create == nil {
		return nil, errors.New("create is missing")
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(create); err != nil {
		return nil, errors.Wrap(err, "create is invalid")
	}

	ctx, lgr := log.ContextAndLoggerWithField(ctx, "create", create)

	now := time.Now()
	defer func() { lgr.WithField("duration", time.Since(now)/time.Microsecond).Debug("Create") }()

	document := &Document{
		Type:                    create.Type,
		GroupID:                 create.GroupID,
		DeduplicationID:         create.DeduplicationID,
		SerialID:                create.SerialID,
		ProcessingAvailableTime: latestTime(create.ProcessingAvailableTime, now),
		ProcessingPriority:      create.ProcessingPriority,
		ProcessingTimeout:       create.ProcessingTimeout,
		Metadata:                create.Metadata,
		PendingTime:             now,
		State:                   work.StatePending,
		CreatedTime:             now,
		Revision:                1,
	}

	ctx, lgr = log.ContextAndLoggerWithField(ctx, "document", document)

	// From this point forward, the context should not be cancelable
	ctx = context.WithoutCancel(ctx)

	result, err := s.InsertOne(ctx, document)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("Unable to create work")
		return nil, errors.Wrap(err, "unable to create work")
	}

	document.ID = result.InsertedID.(bsonPrimitive.ObjectID)

	return document.AsWork(), nil
}

func (s *Store) Get(ctx context.Context, id string, condition *storeStructured.Condition) (*work.Work, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	objectID, err := objectIDFromID(id)
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

	document, err := s.findOne(ctx, query)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("Unable to get work")
		return nil, errors.Wrap(err, "unable to get work")
	} else if document == nil {
		return nil, nil
	}

	return document.AsWork(), nil
}

func (s *Store) Update(ctx context.Context, id string, condition *storeStructured.Condition, update *work.Update) (*work.Work, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	objectID, err := objectIDFromID(id)
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

	document, err := s.findOne(ctx, query)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("Unable to get work before update")
		return nil, errors.Wrap(err, "unable to get work before update")
	} else if document == nil {
		return nil, nil
	}

	set := bson.M{}
	unset := bson.M{}

	if update.State != document.State {
		switch update.State {
		case work.StatePending:
			set["pendingTime"] = now
			unset["processingTimeoutTime"] = true
			if document.ProcessingTime != nil {
				set["processingDuration"] = now.Sub(*document.ProcessingTime).Seconds()
			}
			unset["failingTime"] = true
			unset["failingError"] = true
			unset["failingRetryCount"] = true
			unset["failingRetryTime"] = true
			unset["failedTime"] = true
			unset["failedError"] = true
			unset["successTime"] = true
		case work.StateProcessing:
			set["processingTime"] = now
			set["processingTimeoutTime"] = now.Add(time.Duration(document.ProcessingTimeout) * time.Second)
			unset["processingDuration"] = true
			unset["failedTime"] = true
			unset["failedError"] = true
			unset["successTime"] = true
		case work.StateFailing:
			unset["processingTimeoutTime"] = true
			if document.ProcessingTime != nil {
				set["processingDuration"] = now.Sub(*document.ProcessingTime).Seconds()
			}
			set["failingTime"] = now
			unset["failedTime"] = true
			unset["failedError"] = true
			unset["successTime"] = true
		case work.StateFailed:
			unset["processingTimeoutTime"] = true
			if document.ProcessingTime != nil {
				set["processingDuration"] = now.Sub(*document.ProcessingTime).Seconds()
			}
			set["failedTime"] = now
			unset["successTime"] = true
		case work.StateSuccess:
			unset["processingTimeoutTime"] = true
			if document.ProcessingTime != nil {
				set["processingDuration"] = now.Sub(*document.ProcessingTime).Seconds()
			}
			unset["failingTime"] = true
			unset["failingError"] = true
			unset["failingRetryCount"] = true
			unset["failingRetryTime"] = true
			unset["failedTime"] = true
			unset["failedError"] = true
			set["successTime"] = now
		}
		set["state"] = update.State
	}

	if update.PendingUpdate != nil {
		set["processingAvailableTime"] = latestTime(update.PendingUpdate.ProcessingAvailableTime, now)
		set["processingPriority"] = update.PendingUpdate.ProcessingPriority
		set["processingTimeout"] = update.PendingUpdate.ProcessingTimeout
		set["metadata"] = update.PendingUpdate.Metadata
	}
	if update.ProcessingUpdate != nil {
		set["metadata"] = update.ProcessingUpdate.Metadata
	}
	if update.FailingUpdate != nil {
		set["failingError"] = update.FailingUpdate.FailingError
		set["failingRetryCount"] = update.FailingUpdate.FailingRetryCount
		set["failingRetryTime"] = latestTime(update.FailingUpdate.FailingRetryTime, now)
		set["metadata"] = update.FailingUpdate.Metadata
	}
	if update.FailedUpdate != nil {
		set["failedError"] = update.FailedUpdate.FailedError
		set["metadata"] = update.FailedUpdate.Metadata
	}
	if update.SuccessUpdate != nil {
		set["metadata"] = update.SuccessUpdate.Metadata
	}

	set["modifiedTime"] = now

	// From this point forward, the context should not be cancelable
	ctx = context.WithoutCancel(ctx)

	err = s.FindOneAndUpdate(ctx, query, s.ConstructUpdate(set, unset)).Decode(&document)
	lgr = lgr.WithError(err)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		} else {
			lgr.Error("Unable to update work")
			return nil, errors.Wrap(err, "unable to update work")
		}
	}

	query = bson.M{"_id": objectID}
	document, err = s.findOne(ctx, query)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("Unable to get work after update")
		return nil, errors.Wrap(err, "unable to get work after update")
	} else if document == nil {
		return nil, nil
	}

	return document.AsWork(), nil
}

func (s *Store) Delete(ctx context.Context, id string, condition *storeStructured.Condition) (*work.Work, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	objectID, err := objectIDFromID(id)
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

	// From this point forward, the context should not be cancelable
	ctx = context.WithoutCancel(ctx)

	var document *Document
	err = s.FindOneAndDelete(ctx, query).Decode(&document)
	lgr = lgr.WithError(err)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		} else {
			lgr.Error("Unable to delete work")
			return nil, errors.Wrap(err, "unable to delete work")
		}
	}

	lgr = lgr.WithField("document", document)

	return document.AsWork(), nil
}

func (s *Store) DeleteAllByGroupID(ctx context.Context, groupID string) (int, error) {
	if ctx == nil {
		return 0, errors.New("context is missing")
	}
	if groupID == "" {
		return 0, errors.New("group id is missing")
	} else if len(groupID) > work.GroupIDLengthMaximum {
		return 0, errors.New("group id is invalid")
	}

	ctx, lgr := log.ContextAndLoggerWithField(ctx, "groupId", groupID)

	now := time.Now()
	defer func() { lgr.WithField("duration", time.Since(now)/time.Microsecond).Debug("DeleteAllByGroupID") }()

	query := bson.M{"groupId": groupID}

	deleteResult, err := s.DeleteMany(ctx, query)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("Unable to delete all by group id work")
		return 0, errors.Wrap(err, "unable to delete all by group id work")
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
	ID                      bsonPrimitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"` // Database only
	Type                    string                 `json:"type" bson:"type"`
	GroupID                 *string                `json:"groupId,omitempty" bson:"groupId,omitempty"`
	DeduplicationID         *string                `json:"deduplicationId,omitempty" bson:"deduplicationId,omitempty"`
	SerialID                *string                `json:"serialId,omitempty" bson:"serialId,omitempty"`
	ProcessingAvailableTime time.Time              `json:"processingAvailableTime" bson:"processingAvailableTime"`
	ProcessingPriority      int                    `json:"processingPriority" bson:"processingPriority"`
	ProcessingTimeout       int                    `json:"processingTimeout" bson:"processingTimeout"`
	Metadata                map[string]any         `json:"metadata,omitempty" bson:"metadata,omitempty"` // Database only
	PendingTime             time.Time              `json:"pendingTime" bson:"pendingTime"`
	ProcessingTime          *time.Time             `json:"processingTime,omitempty" bson:"processingTime,omitempty"`
	ProcessingTimeoutTime   *time.Time             `json:"processingTimeoutTime,omitempty" bson:"processingTimeoutTime,omitempty"`
	ProcessingDuration      *float64               `json:"processingDuration,omitempty" bson:"processingDuration,omitempty"`
	FailingTime             *time.Time             `json:"failingTime,omitempty" bson:"failingTime,omitempty"`
	FailingError            *errors.Serializable   `json:"failingError,omitempty" bson:"failingError,omitempty"`
	FailingRetryCount       *int                   `json:"failingRetryCount,omitempty" bson:"failingRetryCount,omitempty"`
	FailingRetryTime        *time.Time             `json:"failingRetryTime,omitempty" bson:"failingRetryTime,omitempty"`
	FailedTime              *time.Time             `json:"failedTime,omitempty" bson:"failedTime,omitempty"`
	FailedError             *errors.Serializable   `json:"failedError,omitempty" bson:"failedError,omitempty"`
	SuccessTime             *time.Time             `json:"successTime,omitempty" bson:"successTime,omitempty"`
	State                   string                 `json:"state" bson:"state"`
	CreatedTime             time.Time              `json:"createdTime" bson:"createdTime"`
	ModifiedTime            *time.Time             `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	ModifiedBatchID         *string                `json:"modifiedBatchId,omitempty" bson:"modifiedBatchId,omitempty"` // Database only
	Revision                int                    `json:"revision" bson:"revision"`
}

func (d *Document) AsWork() *work.Work {
	return &work.Work{
		ID:                      d.ID.Hex(),
		Type:                    d.Type,
		GroupID:                 d.GroupID,
		DeduplicationID:         d.DeduplicationID,
		SerialID:                d.SerialID,
		ProcessingAvailableTime: d.ProcessingAvailableTime,
		ProcessingPriority:      d.ProcessingPriority,
		ProcessingTimeout:       d.ProcessingTimeout,
		Metadata:                d.Metadata,
		PendingTime:             d.PendingTime,
		ProcessingTime:          d.ProcessingTime,
		ProcessingTimeoutTime:   d.ProcessingTimeoutTime,
		ProcessingDuration:      d.ProcessingDuration,
		FailingTime:             d.FailingTime,
		FailingError:            d.FailingError,
		FailingRetryCount:       d.FailingRetryCount,
		FailingRetryTime:        d.FailingRetryTime,
		FailedTime:              d.FailedTime,
		FailedError:             d.FailedError,
		SuccessTime:             d.SuccessTime,
		State:                   d.State,
		CreatedTime:             d.CreatedTime,
		ModifiedTime:            d.ModifiedTime,
		Revision:                d.Revision,
	}
}

type Documents []*Document

func (d Documents) AsWork() []*work.Work {
	wrks := make([]*work.Work, len(d))
	for index, document := range d {
		wrks[index] = document.AsWork()
	}
	return wrks
}

func objectIDFromID(id string) (bsonPrimitive.ObjectID, error) {
	if id == "" {
		return bsonPrimitive.NilObjectID, errors.New("id is missing")
	} else if objectID, err := bsonPrimitive.ObjectIDFromHex(id); err != nil {
		return bsonPrimitive.NilObjectID, errors.New("id is invalid")
	} else {
		return objectID, nil
	}
}

func latestTime(tms ...time.Time) time.Time {
	var latestTime time.Time
	for _, tm := range tms {
		if tm.After(latestTime) {
			latestTime = tm
		}
	}
	return latestTime
}
