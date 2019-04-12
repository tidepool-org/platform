package main

import (
	"time"

	mgo "github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
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
	dataStore, err := storeStructuredMongo.NewStore(mongoConfig, m.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create data store")
	}
	defer dataStore.Close()

	m.Logger().Debug("Creating data session")

	dataSession := dataStore.NewSession("deviceData")
	defer dataSession.Close()

	var count int
	count += m.migrateUploadDataDeduplicatorDescriptor(dataSession, "org.tidepool.continuous.origin", "org.tidepool.deduplicator.dataset.delete.origin")
	count += m.migrateUploadDataDeduplicatorDescriptor(dataSession, "org.tidepool.hash-deactivate-old", "org.tidepool.deduplicator.device.deactivate.hash")
	count += m.migrateUploadDataDeduplicatorDescriptor(dataSession, "org.tidepool.truncate", "org.tidepool.deduplicator.device.truncate.dataset")
	count += m.migrateUploadDataDeduplicatorDescriptor(dataSession, "org.tidepool.continuous", "org.tidepool.deduplicator.none")

	m.Logger().Infof("Migrated %d data duplicator descriptors", count)

	return nil
}

func (m *Migration) migrateUploadDataDeduplicatorDescriptor(dataSession *storeStructuredMongo.Session, fromName string, toName string) int {
	logger := m.Logger().WithFields(log.Fields{"fromName": fromName, "toName": toName})

	logger.Debug("Migrating upload data deduplicator descriptors")

	var count int
	var err error

	selector := bson.M{
		"type":               "upload",
		"_deduplicator.name": fromName,
	}

	if m.DryRun() {
		count, err = dataSession.C().Find(selector).Count()
	} else {
		update := bson.M{
			"$set": bson.M{
				"_deduplicator.name": toName,
			},
		}

		var changeInfo *mgo.ChangeInfo
		changeInfo, err = dataSession.C().UpdateAll(selector, update)
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
