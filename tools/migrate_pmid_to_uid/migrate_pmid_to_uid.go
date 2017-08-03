package main

import (
	"fmt"

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
	IndexFlag  = "index"
	DryRunFlag = "dry-run"
)

func main() {
	application.Run(NewTool())
}

type Tool struct {
	*mongoTool.Tool
	index  bool
	dryRun bool
}

func NewTool() (*Tool, error) {
	tuel, err := mongoTool.NewTool("migrate_pmid_to_uid", "TIDEPOOL")
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

	t.CLI().Usage = "Migrate all metadata to include user id derived from _id"
	t.CLI().Authors = []cli.Author{
		{
			Name:  "Darin Krauss",
			Email: "darin@tidepool.org",
		},
	}
	t.CLI().Flags = append(t.CLI().Flags,
		cli.BoolFlag{
			Name:  fmt.Sprintf("%s,%s", IndexFlag, "i"),
			Usage: "add unique index after migration",
		},
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

	t.index = context.Bool(IndexFlag)
	t.dryRun = context.Bool(DryRunFlag)

	return true
}

func (t *Tool) execute() error {
	if t.index && t.dryRun {
		return errors.New(t.Name(), "cannot specify --index with --dry-run")
	}

	metaIDToUserIDMap, err := t.buildMetaIDToUserIDMap()
	if err != nil {
		return errors.Wrap(err, t.Name(), "unable to build meta id to user id map")
	}

	err = t.migrateMetaIDToUserIDForMetadata(metaIDToUserIDMap)
	if err != nil {
		return errors.Wrap(err, t.Name(), "unable to migrate meta id to user id for metadata")
	}

	return nil
}

func (t *Tool) buildMetaIDToUserIDMap() (map[string]string, error) {
	t.Logger().Debug("Building meta id to user id map")

	userIDMap := map[string]bool{}
	metaIDToUserIDMap := map[string]string{}

	t.Logger().Debug("Creating users store")

	mongoConfig := t.MongoConfig().Clone()
	mongoConfig.Database = "user"
	mongoConfig.Collection = "users"
	usersStore, err := mongo.New(t.Logger(), mongoConfig)
	if err != nil {
		return nil, errors.Wrap(err, t.Name(), "unable to create users store")
	}
	defer usersStore.Close()

	t.Logger().Debug("Creating users session")

	usersSession := usersStore.NewSession(t.Logger())
	defer usersSession.Close()

	t.Logger().Debug("Iterating users")

	iter := usersSession.C().Find(bson.M{}).Select(bson.M{"_id": 0, "userid": 1, "private.meta.id": 1}).Iter()

	var result struct {
		UserID  string `bson:"userid"`
		Private struct {
			Meta struct {
				ID string `bson:"id"`
			} `bson:"meta"`
		} `bson:"private"`
	}
	for iter.Next(&result) {
		userLogger := t.Logger()

		userID := result.UserID
		if userID == "" {
			userLogger.Warn("Missing user id in result from users query")
			continue
		}

		userLogger = userLogger.WithField("userId", userID)

		if _, ok := userIDMap[userID]; ok {
			userLogger.Error("Multiple users found with same user id")
			continue
		}
		userIDMap[userID] = true

		metaID := result.Private.Meta.ID
		if metaID == "" {
			userLogger.Warn("User found without meta id")
			continue
		}

		userLogger = userLogger.WithField("metaId", metaID)

		if _, ok := metaIDToUserIDMap[metaID]; ok {
			userLogger.Error("Multiple users found with same meta id")
			continue
		}
		metaIDToUserIDMap[metaID] = userID
	}
	if err = iter.Close(); err != nil {
		return nil, errors.Wrap(err, t.Name(), "unable to iterate users")
	}

	t.Logger().Debugf("Found %d users with meta", len(metaIDToUserIDMap))

	return metaIDToUserIDMap, nil
}

func (t *Tool) migrateMetaIDToUserIDForMetadata(metaIDToUserIDMap map[string]string) error {
	t.Logger().Debug("Migrating meta id to user id for metadata")

	var migrateMetaCount int
	var migrateMetadataCount int

	t.Logger().Debug("Creating metadata data store")

	mongoConfig := t.MongoConfig().Clone()
	mongoConfig.Database = "seagull"
	mongoConfig.Collection = "seagull"
	metadataStore, err := mongo.New(t.Logger(), mongoConfig)
	if err != nil {
		return errors.Wrap(err, t.Name(), "unable to create metadata store")
	}
	defer metadataStore.Close()

	t.Logger().Debug("Creating metadata session")

	metadataSession := metadataStore.NewSession(t.Logger())
	defer metadataSession.Close()

	t.Logger().Debug("Walking meta id to user id map")

	var count int
	for metaID, userID := range metaIDToUserIDMap {
		metadataLogger := t.Logger().WithFields(log.Fields{"metaId": metaID, "userId": userID})

		metadataLogger.Debug("Finding metadata for meta id")

		var results []struct {
			ID     string  `bson:"_id"`
			UserID *string `bson:"userId"`
			Value  *string `bson:"value"`
		}
		err = metadataSession.C().Find(bson.M{"_id": metaID}).All(&results)
		if err != nil {
			metadataLogger.WithError(err).Error("Unable to query for metadata")
			continue
		}

		resultsCount := len(results)
		switch resultsCount {
		case 0:
			metadataLogger.Error("Metadata not found for meta id")
			continue
		case 1:
			break
		default:
			metadataLogger.WithField("count", resultsCount).Error("More than one metadata found for meta id")
			continue
		}

		if result := results[0]; result.UserID != nil {
			if existingUserID := *result.UserID; existingUserID != userID {
				metadataLogger.WithField("existingUserId", existingUserID).Error("Metadata found for meta id with incorrect existing user id")
			}
			continue
		}

		metadataLogger.Debug("Updating metadata for meta id with user id")

		selector := bson.M{
			"_id":    metaID,
			"userId": bson.M{"$exists": false},
		}

		if t.dryRun {
			count, err = metadataSession.C().Find(selector).Count()
		} else {
			update := bson.M{
				"$set": bson.M{"userId": userID},
			}

			var changeInfo *mgo.ChangeInfo
			changeInfo, err = metadataSession.C().UpdateAll(selector, update)
			if changeInfo != nil {
				count = changeInfo.Updated
			}
		}

		if err != nil {
			metadataLogger.WithError(err).Error("Unable to update metadata for meta id with user id")
			continue
		}

		if count > 0 {
			metadataLogger.Infof("Migrated %d metadata", count)
			migrateMetaCount++
			migrateMetadataCount += count
		}
	}

	if !t.dryRun {
		iter := metadataSession.C().Find(bson.M{"userId": bson.M{"$exists": false}}).Iter()
		var result map[string]interface{}
		for iter.Next(&result) {
			t.Logger().WithField("metaId", result["_id"]).Error("Metadata found without user id")
		}
		if err = iter.Close(); err != nil {
			return errors.Wrap(err, t.Name(), "unable to iterate metadata without user id")
		}
	}

	t.Logger().Infof("Migrated %d metadata for %d meta", migrateMetadataCount, migrateMetaCount)

	if t.index {
		t.Logger().Info("Creating unique index on user id")

		index := mgo.Index{
			Key:        []string{"userId"},
			Unique:     true,
			Background: false,
		}
		err = metadataSession.C().EnsureIndex(index)
		if err != nil {
			return errors.Wrap(err, t.Name(), "unable to create metadata index on user id")
		}
	}

	return nil
}
