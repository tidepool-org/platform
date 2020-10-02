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
	mongoMigration "github.com/tidepool-org/platform/migration/mongo"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

func main() {
	application.RunAndExit(NewMigration())
}

type Migration struct {
	*mongoMigration.Migration
}

func NewMigration() *Migration {
	return &Migration{
		Migration: mongoMigration.NewMigration(),
	}
}

func (m *Migration) Initialize(provider application.Provider) error {
	if err := m.Migration.Initialize(provider); err != nil {
		return err
	}

	m.CLI().Usage = "Migrate all data deduplicator descriptors to latest format"
	m.CLI().Description = "Migrate all data deduplicator descriptors to latest format. Deduplicator name 'hash'" +
		"\n   renamed to 'hash-drop-new'. All deduplicator names updated to use 'org.tidepool.' prefix." +
		"\n   Upload records without deduplicator version set to '0.0.0'. Non-upload records updated to" +
		"\n   remove extraneous deduplicator name." +
		"\n\n   This migration is idempotent." +
		"\n\n   NOTE: This migration MUST be executed immediately AFTER upgrading Platform to v1.8.0."
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
	count += m.migrateUploadDataDeduplicatorDescriptor(deviceDataRepository, "truncate", "org.tidepool.truncate")
	count += m.migrateUploadDataDeduplicatorDescriptor(deviceDataRepository, "hash-deactivate-old", "org.tidepool.hash-deactivate-old")
	count += m.migrateUploadDataDeduplicatorDescriptor(deviceDataRepository, "hash", "org.tidepool.hash-drop-new")
	count += m.migrateNonUploadDataDeduplicatorDescriptor(deviceDataRepository)

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
				"_deduplicator.name":    toName,
				"_deduplicator.version": "0.0.0",
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

func (m *Migration) migrateNonUploadDataDeduplicatorDescriptor(dataRepository *storeStructuredMongo.Repository) int64 {
	m.Logger().Debug("Migrating non-upload data deduplicator descriptors")

	var count int64
	var err error

	selector := bson.M{
		"type": bson.M{
			"$ne": "upload",
		},
		"_deduplicator.name": bson.M{
			"$exists": true,
		},
	}

	if m.DryRun() {
		count, err = dataRepository.CountDocuments(context.Background(), selector)
	} else {
		update := bson.M{
			"$unset": bson.M{
				"_deduplicator.name": 1,
			},
		}

		var changeInfo *mongo.UpdateResult
		changeInfo, err = dataRepository.UpdateMany(context.Background(), selector, update)
		if changeInfo != nil {
			count = changeInfo.ModifiedCount
		}
	}

	if err != nil {
		m.Logger().WithError(err).Error("Unable to migrate non-upload data deduplicator descriptors")
	}

	m.Logger().Debugf("Migrated %d non-upload data deduplicator descriptors", count)

	return count
}
