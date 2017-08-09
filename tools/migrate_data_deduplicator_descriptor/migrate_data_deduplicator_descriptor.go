package main

import (
	"fmt"
	"time"

	"github.com/urfave/cli"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store/mongo"
	mongoTool "github.com/tidepool-org/platform/tool/mongo"
)

const (
	DryRunFlag = "dry-run"
)

func main() {
	application.Run(NewTool())
}

type Tool struct {
	*mongoTool.Tool
	dryRun bool
}

func NewTool() (*Tool, error) {
	tuel, err := mongoTool.NewTool("TIDEPOOL")
	if err != nil {
		return nil, err
	}

	return &Tool{
		Tool: tuel,
	}, nil
}

func (t *Tool) Initialize() error {
	if err := t.Tool.Initialize(); err != nil {
		return err
	}

	t.CLI().Usage = "Migrate all data deduplicators to latest format"
	t.CLI().Authors = []cli.Author{
		{
			Name:  "Darin Krauss",
			Email: "darin@tidepool.org",
		},
	}
	t.CLI().Flags = append(t.CLI().Flags,
		cli.BoolFlag{
			Name:  fmt.Sprintf("%s,%s", DryRunFlag, "n"),
			Usage: "dry run only, do not update database",
		},
	)

	t.CLI().Action = func(context *cli.Context) error {
		if !t.ParseContext(context) {
			return nil
		}
		return t.execute()
	}

	return nil
}

func (t *Tool) ParseContext(context *cli.Context) bool {
	if parsed := t.Tool.ParseContext(context); !parsed {
		return parsed
	}

	t.dryRun = context.Bool(DryRunFlag)

	return true
}

func (t *Tool) execute() error {
	t.Logger().Debug("Migrating data deduplicator descriptors")

	t.Logger().Debug("Creating data store")

	mongoConfig := t.MongoConfig().Clone()
	mongoConfig.Database = "data"
	mongoConfig.Collection = "deviceData"
	mongoConfig.Timeout = 60 * time.Minute
	dataStore, err := mongo.New(t.Logger(), mongoConfig)
	if err != nil {
		return errors.Wrap(err, "main", "unable to create data store")
	}
	defer dataStore.Close()

	t.Logger().Debug("Creating data session")

	dataStoreSession := dataStore.NewSession(t.Logger())
	defer dataStoreSession.Close()

	var count int
	count += t.migrateUploadDataDeduplicatorDescriptor(dataStoreSession, "truncate", "org.tidepool.truncate")
	count += t.migrateUploadDataDeduplicatorDescriptor(dataStoreSession, "hash-deactivate-old", "org.tidepool.hash-deactivate-old")
	count += t.migrateUploadDataDeduplicatorDescriptor(dataStoreSession, "hash", "org.tidepool.hash-drop-new")
	count += t.migrateNonUploadDataDeduplicatorDescriptor(dataStoreSession)

	t.Logger().Infof("Migrated %d data duplicator descriptors", count)

	return nil
}

func (t *Tool) migrateUploadDataDeduplicatorDescriptor(dataStoreSession *mongo.Session, fromName string, toName string) int {
	logger := t.Logger().WithFields(log.Fields{"fromName": fromName, "toName": toName})

	logger.Debug("Migrating upload data deduplicator descriptors")

	var count int
	var err error

	selector := bson.M{
		"type":               "upload",
		"_deduplicator.name": fromName,
	}

	if t.dryRun {
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

func (t *Tool) migrateNonUploadDataDeduplicatorDescriptor(dataStoreSession *mongo.Session) int {
	t.Logger().Debug("Migrating non-upload data deduplicator descriptors")

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

	if t.dryRun {
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
		t.Logger().WithError(err).Error("Unable to migrate non-upload data deduplicator descriptors")
	}

	t.Logger().Debugf("Migrated %d non-upload data deduplicator descriptors", count)

	return count
}
