package mongo

import (
	"context"
	"maps"
	"slices"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/metadata"
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

	// TODO: Can this be changed to bson.M? I think so.

	pipeline := P(
		// Match only requested types
		DE("$match", DE("type", DE("$in", types))),
		// Match either pending available OR running with serial id (to remove pending available with matching serial id later)
		DE("$match", DE("$or", []bson.D{
			D(E("state", "pending"), E("processingAvailableTime", DE("$lte", now))),
			D(E("state", "running"), E("serialId", DE("$exists", true))),
		})),
		// Sort by processing priority and available time
		DE("$sort", D(E("processingPriority", -1), E("processingAvailableTime", 1))),
		// Group all documents by type and serial id
		DE("$group", D(E("_id", D(E("type", "$type"), E("serialId", "$serialId"))), E("documents", DE("$push", "$$ROOT")))),
		// Match any without a serial id or any serial id that does not have one in state running
		DE("$match", DE("$or", []bson.D{
			DE("_id.serialId", DE("$exists", false)),
			DE("documents.state", DE("$ne", "running")),
		})),
		// Only allow one per serial id, or all if no serial id
		DE("$project", DE("documents", DE("$cond", D(
			E("if", DE("$ifNull", A("$_id.serialId", false))),
			E("then", DE("$first", "$documents")),
			E("else", "$documents"),
		)))),
		// Ungroup all documents
		DE("$unwind", "$documents"),
		DE("$replaceRoot", DE("newRoot", "$documents")),
		// Sort by processing priority and available time
		DE("$sort", D(E("processingPriority", -1), E("processingAvailableTime", 1))),
	)

	// If one type, then just simple limit
	// Otherwise, group by type, limit each group by type quantity, and ungroup
	if len(types) == 1 {
		pipeline = slices.Concat(pipeline, P(DE("$limit", poll.TypeQuantities[types[0]])))
	} else {
		var branches []bson.D
		for _, typ := range types {
			branch := D(
				// If matches type
				E("case", DE("$eq", A("$_id", typ))),
				// Then limit to quantity requested
				E("then", DE("$firstN", D(E("input", "$documents"), E("n", poll.TypeQuantities[typ])))),
			)
			branches = append(branches, branch)
		}
		pipeline = slices.Concat(pipeline, P(
			// Group all documents by type
			DE("$group", D(E("_id", "$type"), E("documents", DE("$push", "$$ROOT")))),
			// Only capture requested quantity per type (using switch branches)
			DE("$project", DE("documents", DE("$switch", DE("branches", branches)))),
			// Ungroup all documents
			DE("$unwind", "$documents"),
			DE("$replaceRoot", DE("newRoot", "$documents")),
		))
	}

	// Only need _id and revision
	pipeline = slices.Concat(pipeline, P(DE("$project", D(E("_id", true), E("revision", true)))))

	// Perform aggregation
	cursor, err := s.Aggregate(ctx, pipeline)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.WithField("pipeline", pipeline).Error("Unable to aggregate poll work")
		return nil, errors.Wrap(err, "unable to aggregate poll work")
	}
	defer cursor.Close(ctx)

	// Get identifiers (_id and revision), if nothing, then bail
	var identifiers []bson.M
	err = cursor.All(ctx, &identifiers)
	lgr = lgr.WithError(err)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		} else {
			lgr.Error("unable to get all poll work")
			return nil, errors.Wrap(err, "unable to get all poll work")
		}
	} else if len(identifiers) == 0 {
		return nil, nil
	}

	// Build query and update
	query := bson.M{"$or": identifiers}
	update := bson.M{
		"$set": bson.M{
			"processingTime":        now,
			"processingTimeoutTime": bson.M{"$dateAdd": bson.M{"startDate": now, "unit": "second", "amount": "$processingTimeout"}},
			"state":                 work.StateProcessing,
			"modifiedTime":          now,
			"revision":              bson.M{"$add": bson.A{bson.M{"$ifNull": bson.A{"$revision", 0}}, 1}},
		},
		"$unset": bson.A{"processingDuration"},
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

	// to avoid getting work that another process has claimed, we need
	// some unique identifier to match on. It should use also _id to reduce scope,
	// but cannot use revision+1 (because the other process would have bumped it, too),
	// So, could key off modifiedTime (as it is down to the ms), but that could, in  occur, too.
	// Need UUID?

	// TODO: get only changed
	// TODO: log any error?
	//	lgr = lgr.WithField("count", len(documents))

	// return wrks, nil
	return nil, nil
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
		return nil, errors.New("pagination is missing")
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

	opts := storeStructuredMongo.
		FindWithPagination(pagination).
		SetSort(bson.M{"createdTime": 1})
	cursor, err := s.Find(ctx, query, opts)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("Unable to list work")
		return nil, errors.Wrap(err, "unable to list work")
	}

	var documents Documents
	err = cursor.All(ctx, &documents)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("Unable to decode work")
		return nil, errors.Wrap(err, "unable to decode work")
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
		ProcessingAvailableTime: create.ProcessingAvailableTime,
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

	document.ID = result.InsertedID.(primitive.ObjectID)
	lgr = lgr.WithField("document", document)

	prometheusWorkTypeStateCurrent.WithLabelValues(document.Type, document.State).Inc()

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

	document, err := s.get(ctx, query)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("Unable to get work")
		return nil, err
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
	} else if update.IsEmpty() {
		return nil, errors.New("update is empty")
	}

	ctx, lgr := log.ContextAndLoggerWithFields(ctx, log.Fields{"id": id, "condition": condition, "update": update})

	now := time.Now()
	defer func() { lgr.WithField("duration", time.Since(now)/time.Microsecond).Debug("Update") }()

	query := bson.M{"_id": objectID}
	if condition.Revision != nil {
		query["revision"] = *condition.Revision
	}

	document, err := s.get(ctx, query)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("Unable to get work")
		return nil, err
	}

	set := bson.M{}
	unset := bson.M{}
	if update.PendingUpdate != nil {
		set["processingAvailableTime"] = update.PendingUpdate.ProcessingAvailableTime
		set["processingPriority"] = update.PendingUpdate.ProcessingPriority
		set["processingTimeout"] = update.PendingUpdate.ProcessingTimeout
		set["metadata"] = update.PendingUpdate.Metadata
	}
	if update.ProcessingUpdate != nil {
		set["metadata"] = update.ProcessingUpdate.Metadata
	}
	if update.StateUpdate != nil && update.StateUpdate.State != document.State {
		switch update.StateUpdate.State {
		case work.StatePending:
			set["pendingTime"] = now
			unset["processingTimeoutTime"] = true
			if document.ProcessingTime != nil {
				set["processingDuration"] = now.Sub(*document.ProcessingTime).Seconds()
			}
		case work.StateProcessing:
			set["processingTime"] = now
			set["processingTimeoutTime"] = now.Add(time.Duration(document.ProcessingTimeout) * time.Second)
			unset["processingDuration"] = true
		}
		set["state"] = update.StateUpdate.State
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

	if update.StateUpdate != nil {
		prometheusWorkTypeStateCurrent.WithLabelValues(document.Type, document.State).Dec()
		prometheusWorkTypeStateCurrent.WithLabelValues(document.Type, update.StateUpdate.State).Inc()
	}

	query = bson.M{"_id": objectID}
	document, err = s.get(ctx, query)
	lgr = lgr.WithError(err)
	if err != nil {
		lgr.Error("Unable to get work")
		return nil, err
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
	if document.State != work.StatePending {
		lgr.Warn("Deleted work with unexpected state")
	}

	prometheusWorkTypeStateCurrent.WithLabelValues(document.Type, document.State).Dec()

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

func (s *Store) get(ctx context.Context, query bson.M) (*Document, error) {
	var document *Document
	if err := s.FindOne(ctx, query).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		} else {
			return nil, errors.Wrap(err, "unable to get work")
		}
	}
	return document, nil
}

type Document struct {
	ID                      primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Type                    string             `json:"type,omitempty" bson:"type,omitempty"`
	GroupID                 *string            `json:"groupId,omitempty" bson:"groupId,omitempty"`
	DeduplicationID         *string            `json:"deduplicationId,omitempty" bson:"deduplicationId,omitempty"`
	SerialID                *string            `json:"serialId,omitempty" bson:"serialId,omitempty"`
	ProcessingAvailableTime time.Time          `json:"processingAvailableTime,omitempty" bson:"processingAvailableTime,omitempty"`
	ProcessingPriority      int                `json:"processingPriority,omitempty" bson:"processingPriority,omitempty"`
	ProcessingTimeout       int                `json:"processingTimeout,omitempty" bson:"processingTimeout,omitempty"`
	Metadata                *metadata.Metadata `json:"metadata,omitempty" bson:"metadata,omitempty"`
	PendingTime             time.Time          `json:"pendingTime,omitempty" bson:"pendingTime,omitempty"`
	ProcessingTime          *time.Time         `json:"processingTime,omitempty" bson:"processingTime,omitempty"`
	ProcessingTimeoutTime   *time.Time         `json:"processingTimeoutTime,omitempty" bson:"processingTimeoutTime,omitempty"`
	ProcessingDuration      *float64           `json:"processingDuration,omitempty" bson:"processingDuration,omitempty"`
	State                   string             `json:"state,omitempty" bson:"state,omitempty"`
	CreatedTime             time.Time          `json:"createdTime,omitempty" bson:"createdTime,omitempty"`
	ModifiedTime            *time.Time         `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	Revision                int                `json:"revision,omitempty" bson:"revision,omitempty"`
}

func (d *Document) AsWork() *work.Work {
	return &work.Work{
		ID:                      d.ID.String(),
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

func objectIDFromID(id string) (primitive.ObjectID, error) {
	if id == "" {
		return primitive.NilObjectID, errors.New("id is missing")
	} else if objectID, err := primitive.ObjectIDFromHex(id); err != nil {
		return primitive.NilObjectID, errors.New("id is invalid")
	} else {
		return objectID, nil
	}
}

func A(a ...any) bson.A {
	return a
}

func E(key string, value any) bson.E {
	return bson.E{Key: key, Value: value}
}

func D(e ...bson.E) bson.D {
	return e
}

func DE(key string, value any) bson.D {
	return D(E(key, value))
}

func P(d ...bson.D) mongo.Pipeline {
	return d
}

var prometheusWorkTypeStateCurrent = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "tidepool_work_type_state_current",
	Help: "The current number of work sorted by type and state",
}, []string{"type", "state"})
