package main

import (
	"time"

	"github.com/urfave/cli"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	mongoMigration "github.com/tidepool-org/platform/migration/mongo"
	"github.com/tidepool-org/platform/store/mongo"
)

func main() {
	application.Run(NewMigration())
}

type Migration struct {
	*mongoMigration.Migration
}

func NewMigration() (*Migration, error) {
	migration, err := mongoMigration.NewMigration("TIDEPOOL")
	if err != nil {
		return nil, err
	}

	return &Migration{
		Migration: migration,
	}, nil
}

func (m *Migration) Initialize() error {
	if err := m.Migration.Initialize(); err != nil {
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

	m.CLI().Action = func(context *cli.Context) error {
		if !m.ParseContext(context) {
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
	mongoConfig.Collection = "deviceData"
	mongoConfig.Timeout = 60 * time.Minute
	dataStore, err := mongo.New(m.Logger(), mongoConfig)
	if err != nil {
		return errors.Wrap(err, "main", "unable to create data store")
	}
	defer dataStore.Close()

	m.Logger().Debug("Creating data session")

	dataStoreSession := dataStore.NewSession(m.Logger())
	defer dataStoreSession.Close()

	var count int
	count += m.migrateUploadDataDeduplicatorDescriptor(dataStoreSession, "truncate", "org.tidepool.truncate")
	count += m.migrateUploadDataDeduplicatorDescriptor(dataStoreSession, "hash-deactivate-old", "org.tidepool.hash-deactivate-old")
	count += m.migrateUploadDataDeduplicatorDescriptor(dataStoreSession, "hash", "org.tidepool.hash-drop-new")
	count += m.migrateNonUploadDataDeduplicatorDescriptor(dataStoreSession)

	m.Logger().Infof("Migrated %d data duplicator descriptors", count)

	return nil
}

func (m *Migration) migrateUploadDataDeduplicatorDescriptor(dataStoreSession *mongo.Session, fromName string, toName string) int {
	logger := m.Logger().WithFields(log.Fields{"fromName": fromName, "toName": toName})

	logger.Debug("Migrating upload data deduplicator descriptors")

	var count int
	var err error

	selector := bson.M{
		"type":               "upload",
		"_deduplicator.name": fromName,
	}

	if m.DryRun() {
		count, err = dataStoreSession.C().Find(selector).Count()
	} else {
		update := bson.M{
			"$set": bson.M{
				"_deduplicator.name":    toName,
				"_deduplicator.version": "0.0.0",
			},
		}

		var changeInfo *mgo.ChangeInfo
		changeInfo, err = dataStoreSession.C().UpdateAll(selector, update)
		if changeInfo != nil {
			count = changeInfo.Updated
		}
	}

	if err != nil {
		logger.WithError(err).Error("Unable to migrate upload data deduplicator descriptors")
	}

	logger.Debugf("Migrated %d upload data deduplicator descriptors", count)

	return count
}

func (m *Migration) migrateNonUploadDataDeduplicatorDescriptor(dataStoreSession *mongo.Session) int {
	m.Logger().Debug("Migrating non-upload data deduplicator descriptors")

	var count int
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
		count, err = dataStoreSession.C().Find(selector).Count()
	} else {
		update := bson.M{
			"$unset": bson.M{
				"_deduplicator.name": 1,
			},
		}

		var changeInfo *mgo.ChangeInfo
		changeInfo, err = dataStoreSession.C().UpdateAll(selector, update)
		if changeInfo != nil {
			count = changeInfo.Updated
		}
	}

	if err != nil {
		m.Logger().WithError(err).Error("Unable to migrate non-upload data deduplicator descriptors")
	}

	m.Logger().Debugf("Migrated %d non-upload data deduplicator descriptors", count)

	return count
}
