package main

import (
	"context"
	"fmt"
	"time"

	"github.com/urfave/cli"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/application"
	dataDeduplicatorDeduplicator "github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	dataRawStoreStructuredMongo "github.com/tidepool-org/platform/data/raw/store/structured/mongo"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	migrationMongo "github.com/tidepool-org/platform/migration/mongo"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"github.com/tidepool-org/platform/work"
	workStoreStructuredMongo "github.com/tidepool-org/platform/work/store/structured/mongo"
)

// These mirror the Tandem plugin, which a migration cannot import as the public build does not carry it.
const (
	tandemDataSetClientName = "org.tidepool.tandem.eventhub"
	tandemProviderName      = "tandem"
	tandemWorkGroupPrefix   = "org.tidepool.tandem"
	tandemIngestWorkType    = "org.tidepool.tandem.work.data.ingest"
	tandemIngestTimeout     = 10 * time.Minute

	// The lease the ingest work holds on a data set while it writes, in the data set metadata; the migration takes
	// it the same way, so no ingest work writes the data set while it is migrated, and expired leases are reclaimed.
	leaseAttempt  = "migration"
	leaseDuration = 1 * time.Hour

	userIDFlag           = "user-id"
	dataDatabaseFlag     = "data-database"
	tidepoolDatabaseFlag = "tidepool-database"

	batchSize = 100
)

// The keys the ingest work writes in the raw data metadata, which a re-ingestion starts without.
var rawProcessingMetadataKeys = []string{"processingStartedTime", "attempt", "processedOrdinal", "rewindOrdinal", "processError", "skipRequested"}

func main() {
	application.RunAndExit(NewMigration())
}

type Migration struct {
	*migrationMongo.Migration
	userID           string
	dataDatabase     string
	tidepoolDatabase string
}

func NewMigration() *Migration {
	return &Migration{
		Migration: migrationMongo.NewMigration(),
	}
}

func (m *Migration) Initialize(provider application.Provider) error {
	if err := m.Migration.Initialize(provider); err != nil {
		return err
	}

	m.CLI().Usage = "Re-ingest every Tandem data set from its raw data with the revise deduplicator"
	m.CLI().Description = "The Tandem ingest work now records the order it processes raw data in, rewinds raw data that\n" +
		"   arrived out of order, and keeps the history of each datum with the revise deduplicator, none of which raw\n" +
		"   data processed before it can take part in. This migration finds every Tandem data set, takes its ingest\n" +
		"   lease so no ingest work writes it meanwhile, destroys its data, returns its raw data to unprocessed,\n" +
		"   switches it to the revise deduplicator, clears its ingest state, which releases the lease, and creates\n" +
		"   ingest work for the data set, which ingests all of its raw data in order. Archived raw data, whose content\n" +
		"   is gone, is left as it is. Run after the platform version with the ordered ingest is deployed, for all\n" +
		"   users or for one with --user-id."
	m.CLI().Authors = []cli.Author{
		{
			Name:  "Todd Kazakov",
			Email: "todd@tidepool.org",
		},
	}
	m.CLI().Flags = append(m.CLI().Flags,
		cli.StringFlag{
			Name:  fmt.Sprintf("%s,%s", userIDFlag, "u"),
			Usage: "migrate the data sets of this user only",
		},
		cli.StringFlag{
			Name:  dataDatabaseFlag,
			Usage: "the database of the data sets, data and raw data",
			Value: "data",
		},
		cli.StringFlag{
			Name:  tidepoolDatabaseFlag,
			Usage: "the database of the data sources and work",
			Value: "tidepool",
		},
	)
	m.CLI().Action = func(ctx *cli.Context) error {
		if !m.ParseContext(ctx) {
			return nil
		}
		return m.execute(log.NewContextWithLogger(context.Background(), m.Logger()))
	}

	return nil
}

func (m *Migration) ParseContext(ctx *cli.Context) bool {
	if parsed := m.Migration.ParseContext(ctx); !parsed {
		return parsed
	}
	m.userID = ctx.String(userIDFlag)
	m.dataDatabase = ctx.String(dataDatabaseFlag)
	m.tidepoolDatabase = ctx.String(tidepoolDatabaseFlag)
	return true
}

type stores struct {
	dataSets    *storeStructuredMongo.Repository
	data        *storeStructuredMongo.Repository
	raw         *storeStructuredMongo.Repository
	dataSources *storeStructuredMongo.Repository
	work        *workStoreStructuredMongo.Store
}

func (m *Migration) execute(ctx context.Context) error {
	dataConfig := m.NewMongoConfig()
	dataConfig.Database = m.dataDatabase
	dataConfig.Timeout = 60 * time.Minute
	dataStore, err := storeStructuredMongo.NewStore(dataConfig)
	if err != nil {
		return errors.Wrap(err, "unable to create data store")
	}
	defer dataStore.Terminate(context.Background())

	tidepoolConfig := m.NewMongoConfig()
	tidepoolConfig.Database = m.tidepoolDatabase
	tidepoolConfig.Timeout = 60 * time.Minute
	tidepoolStore, err := storeStructuredMongo.NewStore(tidepoolConfig)
	if err != nil {
		return errors.Wrap(err, "unable to create tidepool store")
	}
	defer tidepoolStore.Terminate(context.Background())
	workStore, err := workStoreStructuredMongo.NewStore(tidepoolConfig)
	if err != nil {
		return errors.Wrap(err, "unable to create work store")
	}
	defer workStore.Terminate(context.Background())

	s := &stores{
		dataSets:    dataStore.GetRepository("deviceDataSets"),
		data:        dataStore.GetRepository("deviceData"),
		raw:         dataStore.GetRepository("raw"),
		dataSources: tidepoolStore.GetRepository("data_sources"),
		work:        workStore,
	}

	var migratedCount, skippedCount, errorCount int
	lastID := primitive.NilObjectID
	for {
		selector := bson.M{
			"_id":         bson.M{"$gt": lastID},
			"client.name": tandemDataSetClientName,
			"deletedTime": bson.M{"$exists": false},
		}
		if m.userID != "" {
			selector["_userId"] = m.userID
		}
		opts := options.Find().
			SetSort(bson.M{"_id": 1}).
			SetLimit(batchSize).
			SetProjection(bson.M{"_id": 1, "uploadId": 1, "_userId": 1})
		cursor, err := s.dataSets.Find(ctx, selector, opts)
		if err != nil {
			return errors.Wrap(err, "unable to find data sets")
		}

		var dataSets []dataSetDocument
		if err = cursor.All(ctx, &dataSets); err != nil {
			return errors.Wrap(err, "unable to decode data sets")
		}

		for _, dataSet := range dataSets {
			if migrated, err := m.migrate(ctx, s, dataSet); err != nil {
				errorCount++
				m.Logger().WithError(err).WithFields(log.Fields{"userId": dataSet.UserID, "dataSetId": dataSet.UploadID}).Error("unable to migrate data set")
			} else if migrated {
				migratedCount++
			} else {
				skippedCount++
			}
		}

		if len(dataSets) < batchSize {
			break
		}
		lastID = dataSets[len(dataSets)-1].ObjectID
	}

	m.Logger().WithFields(log.Fields{"migrated": migratedCount, "skipped": skippedCount, "errors": errorCount}).Info("Migrated data sets")
	return nil
}

// migrate takes the ingest lease on the data set, destroys its data, returns its raw data to unprocessed, switches it
// to the revise deduplicator, clears its ingest state, which releases the lease, and creates ingest work for its raw
// data of the highest sequence, which the work ingests last, so the work stays pending until every raw data is
// ingested. It reports whether the data set was migrated, which a data set without a data source, without raw data
// to re-ingest, or leased by an ingest run is not.
func (m *Migration) migrate(ctx context.Context, s *stores, dataSet dataSetDocument) (bool, error) {
	logger := m.Logger().WithFields(log.Fields{"userId": dataSet.UserID, "dataSetId": dataSet.UploadID})

	dataSource, err := m.findDataSource(ctx, s, dataSet)
	if err != nil {
		return false, err
	} else if dataSource == nil {
		logger.Warn("skipping data set without a data source")
		return false, nil
	}
	lastRaw, archivedCount, err := m.findLastRaw(ctx, s, dataSet)
	if err != nil {
		return false, err
	} else if lastRaw == nil {
		logger.WithField("archived", archivedCount).Warn("skipping data set without raw data to re-ingest")
		return false, nil
	}
	lastRawID := dataRawStoreStructuredMongo.IDFromObjectIDAndDate(lastRaw.ObjectID, lastRaw.CreatedTime)
	logger = logger.WithFields(log.Fields{"dataSourceId": dataSource.ID, "lastDataRawId": lastRawID, "archived": archivedCount})

	if m.DryRun() {
		logger.Info("[DRY RUN] migrating data set")
		return true, nil
	}

	now := time.Now()
	if leased, err := m.acquireLease(ctx, s, dataSet, now); err != nil {
		return false, err
	} else if !leased {
		logger.Warn("skipping data set leased by an ingest run; run again later")
		return false, nil
	}

	if _, err := s.data.DeleteMany(ctx, bson.M{"_userId": dataSet.UserID, "uploadId": dataSet.UploadID}); err != nil {
		return false, errors.Wrap(err, "unable to destroy data")
	}

	unset := bson.M{"processedTime": "", "archivableTime": ""}
	for _, key := range rawProcessingMetadataKeys {
		unset["metadata."+key] = ""
	}
	rawSelector := bson.M{"userId": dataSet.UserID, "dataSetId": dataSet.UploadID, "archivedTime": bson.M{"$exists": false}}
	rawUpdate := bson.M{"$unset": unset, "$set": bson.M{"modifiedTime": now}, "$inc": bson.M{"revision": 1}}
	if _, err := s.raw.UpdateMany(ctx, rawSelector, rawUpdate); err != nil {
		return false, errors.Wrap(err, "unable to reset raw data")
	}

	// Clearing the metadata releases the lease; the next ingest run starts from the switched data set
	dataSetUpdate := bson.M{
		"$set":   bson.M{"_deduplicator": bson.M{"name": dataDeduplicatorDeduplicator.DataSetOriginReviseName, "version": dataDeduplicatorDeduplicator.DataSetOriginReviseVersion}, "modifiedTime": now},
		"$unset": bson.M{"metadata": ""},
		"$inc":   bson.M{"revision": 1},
	}
	if _, err := s.dataSets.UpdateOne(ctx, bson.M{"_id": dataSet.ObjectID}, dataSetUpdate); err != nil {
		return false, errors.Wrap(err, "unable to switch the data set and clear its ingest state")
	}

	create := &work.Create{
		Type:              tandemIngestWorkType,
		GroupID:           pointer.From(fmt.Sprintf("%s:%s", tandemWorkGroupPrefix, dataSource.ID)),
		DeduplicationID:   pointer.From(fmt.Sprintf("%s:reingest:%d", dataSource.ID, now.Unix())),
		SerialID:          pointer.From(fmt.Sprintf("%s:%s", tandemIngestWorkType, dataSet.UploadID)),
		ProcessingTimeout: int(tandemIngestTimeout.Seconds()),
		Metadata: map[string]any{
			"dataSourceId": dataSource.ID,
			"dataSetId":    dataSet.UploadID,
			"dataRawId":    lastRawID,
		},
	}
	if wrk, err := s.work.Create(ctx, create); err != nil {
		return false, errors.Wrap(err, "unable to create ingest work")
	} else if wrk == nil {
		return false, errors.New("ingest work already exists")
	} else {
		logger = logger.WithField("workId", wrk.ID)
	}

	logger.Info("migrated data set")
	return true, nil
}

// acquireLease writes the ingest lease of the data set for the migration, unless an ingest run holds one that has not
// expired, in which case it reports false.
func (m *Migration) acquireLease(ctx context.Context, s *stores, dataSet dataSetDocument, now time.Time) (bool, error) {
	selector := bson.M{
		"_id": dataSet.ObjectID,
		"$or": bson.A{
			bson.M{"metadata.lease": bson.M{"$exists": false}},
			bson.M{"metadata.lease.expiresTime": bson.M{"$lt": now.UTC().Format(time.RFC3339Nano)}},
		},
	}
	lease := bson.M{
		"attempt":     leaseAttempt,
		"workId":      fmt.Sprintf("%s:%d", leaseAttempt, now.Unix()),
		"expiresTime": now.Add(leaseDuration).UTC().Format(time.RFC3339Nano),
	}
	update := bson.M{"$set": bson.M{"metadata.lease": lease, "modifiedTime": now}, "$inc": bson.M{"revision": 1}}
	result, err := s.dataSets.UpdateOne(ctx, selector, update)
	if err != nil {
		return false, errors.Wrap(err, "unable to lease data set")
	}
	return result.MatchedCount > 0, nil
}

func (m *Migration) findDataSource(ctx context.Context, s *stores, dataSet dataSetDocument) (*dataSourceDocument, error) {
	selector := bson.M{"userId": dataSet.UserID, "providerName": tandemProviderName, "dataSetIds": dataSet.UploadID}
	var dataSource dataSourceDocument
	if err := s.dataSources.FindOne(ctx, selector, options.FindOne().SetProjection(bson.M{"id": 1})).Decode(&dataSource); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "unable to find data source")
	}
	return &dataSource, nil
}

// findLastRaw is the raw data of the data set with the highest pump logs sequence, or the newest when none has a
// sequence, among those not archived, with the count of the archived ones.
func (m *Migration) findLastRaw(ctx context.Context, s *stores, dataSet dataSetDocument) (*rawDocument, int, error) {
	selector := bson.M{"userId": dataSet.UserID, "dataSetId": dataSet.UploadID}
	archivedCount, err := s.raw.CountDocuments(ctx, bson.M{"userId": dataSet.UserID, "dataSetId": dataSet.UploadID, "archivedTime": bson.M{"$exists": true}})
	if err != nil {
		return nil, 0, errors.Wrap(err, "unable to count archived raw data")
	}
	selector["archivedTime"] = bson.M{"$exists": false}

	opts := options.FindOne().
		SetSort(bson.D{{Key: "metadata.minimumSequenceGroup", Value: -1}, {Key: "metadata.minimumSequenceNumber", Value: -1}, {Key: "createdTime", Value: -1}}).
		SetProjection(bson.M{"_id": 1, "createdTime": 1})
	var lastRaw rawDocument
	if err := s.raw.FindOne(ctx, selector, opts).Decode(&lastRaw); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, int(archivedCount), nil
		}
		return nil, 0, errors.Wrap(err, "unable to find raw data")
	}
	return &lastRaw, int(archivedCount), nil
}

type dataSetDocument struct {
	ObjectID primitive.ObjectID `bson:"_id"`
	UploadID string             `bson:"uploadId"`
	UserID   string             `bson:"_userId"`
}

type dataSourceDocument struct {
	ID string `bson:"id"`
}

type rawDocument struct {
	ObjectID    primitive.ObjectID `bson:"_id"`
	CreatedTime time.Time          `bson:"createdTime"`
}
