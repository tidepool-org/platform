package main

import (
	"context"
	"slices"
	"time"

	"github.com/urfave/cli"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	migrationMongo "github.com/tidepool-org/platform/migration/mongo"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

const (
	dataSourcesCollectionName = "data_sources"
	batchSize                 = 1000
)

func main() {
	application.RunAndExit(NewMigration())
}

type Migration struct {
	*migrationMongo.Migration
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

	m.CLI().Usage = "Move the data source dataSetId into dataSetIds"
	m.CLI().Description = "Between 2026-03 and 2026-10 the platform stored the data set of a data source in a single\n" +
		"   dataSetId field. Data sources reference all of their data sets in dataSetIds again, the last being the\n" +
		"   current one, and the platform reads a stored dataSetId as the last entry. This migration appends dataSetId\n" +
		"   to dataSetIds, unless already present, and removes it, so that no document depends on that read path.\n" +
		"   Run after the platform version with dataSetIds is deployed."
	m.CLI().Authors = []cli.Author{
		{
			Name:  "Todd Kazakov",
			Email: "todd@tidepool.org",
		},
	}
	m.CLI().Action = func(ctx *cli.Context) error {
		if !m.ParseContext(ctx) {
			return nil
		}
		return m.execute(log.NewContextWithLogger(context.Background(), m.Logger()))
	}

	return nil
}

func (m *Migration) execute(ctx context.Context) error {
	mongoConfig := m.NewMongoConfig()
	mongoConfig.Timeout = 60 * time.Minute
	dataStore, err := storeStructuredMongo.NewStore(mongoConfig)
	if err != nil {
		return errors.Wrap(err, "unable to create data store")
	}
	defer dataStore.Terminate(context.Background())

	repository := dataStore.GetRepository(dataSourcesCollectionName)

	var appendedCount, presentCount, errorCount int
	lastID := primitive.NilObjectID
	for {
		selector := bson.M{
			"_id":       bson.M{"$gt": lastID},
			"dataSetId": bson.M{"$type": "string"},
		}
		opts := options.Find().
			SetSort(bson.M{"_id": 1}).
			SetLimit(batchSize).
			SetProjection(bson.M{"_id": 1, "id": 1, "dataSetId": 1, "dataSetIds": 1})
		cursor, err := repository.Find(ctx, selector, opts)
		if err != nil {
			return errors.Wrap(err, "unable to find data sources")
		}

		var dataSources []dataSourceDocument
		if err = cursor.All(ctx, &dataSources); err != nil {
			return errors.Wrap(err, "unable to decode data sources")
		}

		for _, dataSrc := range dataSources {
			if dataSrc.DataSetID == "" {
				m.Logger().WithField("id", dataSrc.ID).Warn("skipping data source with an empty data set id")
				continue
			}
			present := slices.Contains(dataSrc.DataSetIDs, dataSrc.DataSetID)
			if err = m.migrate(ctx, repository, dataSrc, present); err != nil {
				errorCount++
				m.Logger().WithError(err).WithField("id", dataSrc.ID).Error("unable to migrate data source")
			} else if present {
				presentCount++
			} else {
				appendedCount++
			}
		}

		if len(dataSources) < batchSize {
			break
		}
		lastID = dataSources[len(dataSources)-1].ObjectID
	}

	m.Logger().WithFields(log.Fields{"appended": appendedCount, "present": presentCount, "errors": errorCount}).Info("Migrated data sources")
	return nil
}

// migrate moves dataSetId to the end of dataSetIds, as the newest data set, and removes it. $addToSet leaves
// a dataSetIds that already holds it unchanged and creates one when the document has none. The revision is
// bumped as for any update, so a concurrent conditional update of the data source does not apply over it.
func (m *Migration) migrate(ctx context.Context, repository *storeStructuredMongo.Repository, dataSrc dataSourceDocument, present bool) error {
	logger := m.Logger().WithFields(log.Fields{"id": dataSrc.ID, "dataSetId": dataSrc.DataSetID, "dataSetIds": dataSrc.DataSetIDs, "present": present})
	if m.DryRun() {
		logger.Info("[DRY RUN] migrating data source")
		return nil
	}

	update := bson.M{
		"$addToSet": bson.M{"dataSetIds": dataSrc.DataSetID},
		"$unset":    bson.M{"dataSetId": ""},
		"$set":      bson.M{"modifiedTime": time.Now()},
		"$inc":      bson.M{"revision": 1},
	}
	if _, err := repository.UpdateOne(ctx, bson.M{"_id": dataSrc.ObjectID}, update); err != nil {
		return err
	}

	logger.Debug("migrated data source")
	return nil
}

type dataSourceDocument struct {
	ObjectID   primitive.ObjectID `bson:"_id"`
	ID         string             `bson:"id"`
	DataSetID  string             `bson:"dataSetId"`
	DataSetIDs []string           `bson:"dataSetIds"`
}
