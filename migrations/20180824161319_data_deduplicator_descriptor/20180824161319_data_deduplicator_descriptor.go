package main

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/urfave/cli"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	migrationMongo "github.com/tidepool-org/platform/migration/mongo"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
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

	m.CLI().Usage = "Migrate all data deduplicator descriptors to the latest format"
	m.CLI().Description = "Migrate all data deduplicator descriptors to the latest format. Specifically, migrate" +
		"\n   the name from:" +
		"\n\n   'org.tidepool.continuous.origin'   => 'org.tidepool.deduplicator.dataset.delete.origin'" +
		"\n   'org.tidepool.hash-deactivate-old' => 'org.tidepool.deduplicator.device.deactivate.hash'" +
		"\n   'org.tidepool.truncate'            => 'org.tidepool.deduplicator.device.truncate.dataset'" +
		"\n   'org.tidepool.continuous'          => 'org.tidepool.deduplicator.none'" +
		"\n\n   This migration is idempotent." +
		"\n\n   NOTE: This migration MUST be executed immediately AFTER upgrading Platform to v1.29.0."
	m.CLI().Authors = []cli.Author{
		{
			Name:  "Darin Krauss",
			Email: "darin@tidepool.org",
		},
	}

	m.CLI().Action = func(ctx *cli.Context) error {
		if !m.ParseContext(ctx) {
			return nil
		}
		return m.execute()
	}

	return nil
}

func (m *Migration) execute() error {
	m.Logger().Debug("Migrating data deduplicator descriptors")

	m.Logger().Debug("Creating data store")

	mongoConfig := m.NewMongoConfig()
	mongoConfig.Database = "data"
	mongoConfig.Timeout = 60 * time.Minute
	params := storeStructuredMongo.Params{DatabaseConfig: mongoConfig}
	dataStore, err := storeStructuredMongo.NewStore(params)
	if err != nil {
		return errors.Wrap(err, "unable to create data store")
	}
	defer dataStore.Terminate(context.Background())

	m.Logger().Debug("Creating data repository")

	deviceDataRepository := dataStore.GetRepository("deviceData")

	var count int64
	count += m.migrateUploadDataDeduplicatorDescriptor(deviceDataRepository, "org.tidepool.continuous.origin", "org.tidepool.deduplicator.dataset.delete.origin")
	count += m.migrateUploadDataDeduplicatorDescriptor(deviceDataRepository, "org.tidepool.hash-deactivate-old", "org.tidepool.deduplicator.device.deactivate.hash")
	count += m.migrateUploadDataDeduplicatorDescriptor(deviceDataRepository, "org.tidepool.truncate", "org.tidepool.deduplicator.device.truncate.dataset")
	count += m.migrateUploadDataDeduplicatorDescriptor(deviceDataRepository, "org.tidepool.continuous", "org.tidepool.deduplicator.none")

	m.Logger().Infof("Migrated %d data duplicator descriptors", count)

	return nil
}

func (m *Migration) migrateUploadDataDeduplicatorDescriptor(dataRepository *storeStructuredMongo.Repository, fromName string, toName string) int64 {
	logger := m.Logger().WithFields(log.Fields{"fromName": fromName, "toName": toName})

	logger.Debug("Migrating upload data deduplicator descriptors")

	var count int64
	var err error

	selector := bson.M{
		"type":               "upload",
		"_deduplicator.name": fromName,
	}

	if m.DryRun() {
		count, err = dataRepository.CountDocuments(context.Background(), selector)
	} else {
		update := bson.M{
			"$set": bson.M{
				"_deduplicator.name": toName,
			},
		}

		var changeInfo *mongo.UpdateResult
		changeInfo, err = dataRepository.UpdateMany(context.Background(), selector, update)
		if changeInfo != nil {
			count = changeInfo.ModifiedCount
		}
	}

	if err != nil {
		logger.WithError(err).Error("Unable to migrate upload data deduplicator descriptors")
	}

	logger.Debugf("Migrated %d upload data deduplicator descriptors", count)

	return count
}
