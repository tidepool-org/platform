package main

import (
	"encoding/json"
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

	t.CLI().Usage = "Migrate all device data to include user id derived from group id"
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
	metaIDToUserIDMap, err := t.buildMetaIDToUserIDMap()
	if err != nil {
		return errors.Wrap(err, "main", "unable to build meta id to user id map")
	}

	groupIDToUserIDMap, err := t.buildGroupIDToUserIDMap(metaIDToUserIDMap)
	if err != nil {
		return errors.Wrap(err, "main", "unable to build group id to user id map")
	}

	err = t.migrateGroupIDToUserIDForDeviceData(groupIDToUserIDMap)
	if err != nil {
		return errors.Wrap(err, "main", "unable to migrate group id to user id for device data")
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
		return nil, errors.Wrap(err, "main", "unable to create users store")
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
			userLogger.Error("Found multiple users with same user id")
			continue
		}
		userIDMap[userID] = true

		metaID := result.Private.Meta.ID
		if metaID == "" {
			userLogger.Warn("Missing private meta id in result from users query for user id")
			continue
		}

		userLogger = userLogger.WithField("metaId", metaID)

		if _, ok := metaIDToUserIDMap[metaID]; ok {
			userLogger.Error("Found multiple users with same meta id")
			continue
		}
		metaIDToUserIDMap[metaID] = userID
	}
	if err = iter.Close(); err != nil {
		return nil, errors.Wrap(err, "main", "unable to iterate users")
	}

	t.Logger().Debugf("Found %d users with meta", len(metaIDToUserIDMap))

	return metaIDToUserIDMap, nil
}

func (t *Tool) buildGroupIDToUserIDMap(metaIDToUserIDMap map[string]string) (map[string]string, error) {
	t.Logger().Debug("Building group id to user id map")

	metaIDMap := map[string]bool{}
	groupIDToUserIDMap := map[string]string{}

	t.Logger().Debug("Creating meta store")

	mongoConfig := t.MongoConfig().Clone()
	mongoConfig.Database = "seagull"
	mongoConfig.Collection = "seagull"
	metaStore, err := mongo.New(t.Logger(), mongoConfig)
	if err != nil {
		return nil, errors.Wrap(err, "main", "unable to create meta store")
	}
	defer metaStore.Close()

	t.Logger().Debug("Creating meta session")

	metaSession := metaStore.NewSession(t.Logger())
	defer metaSession.Close()

	t.Logger().Debug("Iterating meta")

	iter := metaSession.C().Find(bson.M{}).Iter()

	var result struct {
		MetaID string `bson:"_id"`
		Value  string `bson:"value"`
	}
	for iter.Next(&result) {
		metaLogger := t.Logger()

		metaID := result.MetaID
		if metaID == "" {
			metaLogger.Warn("Missing meta id in result from meta query")
			continue
		}

		metaLogger = metaLogger.WithField("metaId", metaID)

		if _, ok := metaIDMap[metaID]; ok {
			metaLogger.Error("Found multiple metas with same meta id")
			continue
		}
		metaIDMap[metaID] = true

		userID, ok := metaIDToUserIDMap[metaID]
		if !ok {
			metaLogger.Error("Missing user id for meta id")
			continue
		}

		metaLogger = metaLogger.WithField("userId", userID)

		if result.Value == "" {
			metaLogger.Warn("Missing value in result from meta query for meta id")
			continue
		}

		var value struct {
			Private struct {
				Uploads struct {
					ID string `json:"id"`
				} `json:"uploads"`
			} `json:"private"`
		}
		if err = json.Unmarshal([]byte(result.Value), &value); err != nil {
			metaLogger.WithError(err).Warn("Unable to unmarshal value from meta query for meta id")
			continue
		}

		groupID := value.Private.Uploads.ID
		if groupID == "" {
			metaLogger.Debug("Missing group id in value in result from meta query for meta id")
			continue
		}

		metaLogger = metaLogger.WithField("groupId", groupID)

		if _, ok = groupIDToUserIDMap[groupID]; ok {
			metaLogger.Error("Found multiple metas with same group id")
			continue
		}
		groupIDToUserIDMap[groupID] = userID
	}
	if err = iter.Close(); err != nil {
		return nil, errors.Wrap(err, "main", "unable to iterate meta")
	}

	t.Logger().Debugf("Found %d groups with user", len(groupIDToUserIDMap))

	return groupIDToUserIDMap, nil
}

func (t *Tool) migrateGroupIDToUserIDForDeviceData(groupIDToUserIDMap map[string]string) error {
	t.Logger().Debug("Migrating group id to user id for device data")

	var migrateGroupCount int
	var migrateDeviceDataCount int

	t.Logger().Debug("Creating device data store")

	mongoConfig := t.MongoConfig().Clone()
	mongoConfig.Database = "data"
	mongoConfig.Collection = "deviceData"
	deviceDataStore, err := mongo.New(t.Logger(), mongoConfig)
	if err != nil {
		return errors.Wrap(err, "main", "unable to create device data store")
	}
	defer deviceDataStore.Close()

	t.Logger().Debug("Creating device data session")

	deviceDataSession := deviceDataStore.NewSession(t.Logger())
	defer deviceDataSession.Close()

	t.Logger().Debug("Walking group id to user id map")

	var count int
	for groupID, userID := range groupIDToUserIDMap {
		dataLogger := t.Logger().WithFields(log.Fields{"groupId": groupID, "userId": userID})

		dataLogger.Debug("Finding device data for group id with incorrect existing user id")

		query := bson.M{
			"$and": []bson.M{
				{"_groupId": groupID},
				{"_userId": bson.M{"$exists": true}},
				{"_userId": bson.M{"$ne": userID}},
			},
		}
		count, err = deviceDataSession.C().Find(query).Count()
		if err != nil {
			dataLogger.WithError(err).Error("Unable to query for incorrect device data")
			continue
		}

		if count != 0 {
			dataLogger.WithField("count", count).Error("Found device data for group id with incorrect existing user id")
			continue
		}

		dataLogger.Debug("Updating device data for group id with user id")

		selector := bson.M{
			"_groupId": groupID,
			"_userId":  bson.M{"$exists": false},
		}

		if t.dryRun {
			count, err = deviceDataSession.C().Find(selector).Count()
		} else {
			update := bson.M{
				"$set": bson.M{"_userId": userID},
			}

			var changeInfo *mgo.ChangeInfo
			changeInfo, err = deviceDataSession.C().UpdateAll(selector, update)
			if changeInfo != nil {
				count = changeInfo.Updated
			}
		}

		if err != nil {
			dataLogger.WithError(err).Error("Unable to update device data for group id with user id")
			continue
		}

		if count > 0 {
			dataLogger.Infof("Migrated %d device data", count)
			migrateGroupCount++
			migrateDeviceDataCount += count
		}
	}

	if !t.dryRun {
		if count, err = deviceDataSession.C().Find(bson.M{"_userId": bson.M{"$exists": false}}).Count(); err != nil {
			t.Logger().WithError(err).Error("Unable to query for device data without user id")
		} else if count != 0 {
			t.Logger().WithField("count", count).Error("Found device data without user id")
		}
	}

	t.Logger().Infof("Migrated %d device data for %d groups", migrateDeviceDataCount, migrateGroupCount)

	return nil
}
