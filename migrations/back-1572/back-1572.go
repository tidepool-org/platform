package main

import (
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/urfave/cli"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/errors"
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

	m.CLI().Usage = "BACK-1572: Deduplicate data sets affected by UPLOAD-323/BACK-1379"
	m.CLI().Description = "BACK-1572: Deduplicate data sets affected by UPLOAD-323/BACK-1379. Specifically,\n" +
		"   Find all data where the `deviceId` field starts with 'InsOmn', and check whether the Jellyfish\n" +
		"   generated `id` field matches the expected hash.\n" +
		"   If the `id` field does not match:\n" +
		"     * Update the `id` and `_id` fields to the expected hashes\n" +
		"     * Search for any duplicate documents, and archive the document with the initially incorrect hash"
	m.CLI().Authors = []cli.Author{
		{
			Name:  "Lennart Goedhart",
			Email: "lennart@tidepool.org",
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
	m.Logger().Debug("Deduplicate data sets affected by UPLOAD-323/BACK-1379")

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

	hashUpdatedCount, archivedCount := m.migrateOmnipodDocuments(dataSession)

	m.Logger().Infof("Migrated %d duplicate Omnipod documents", hashUpdatedCount)
	m.Logger().Infof("Archived %d duplicate Omnipod documents", archivedCount)

	return nil
}

func (m *Migration) migrateOmnipodDocuments(dataSession *storeStructuredMongo.Session) (int, int) {
	logger := m.Logger()

	logger.Debug("Finding distinct users")

	var userIDs []string
	var hashUpdatedCount, archivedCount int

	err := dataSession.C().Find(bson.M{}).Distinct("_userId", &userIDs)
	if err != nil {
		logger.WithError(err).Error("Unable to execute distinct query")
	} else {
		logger.Debugf("Finding Omnipod records for %d users", len(userIDs))

		for _, userID := range userIDs {
			logger.Debugf("Finding Omnipod records for user ID %s", userID)
			selector := bson.M{
				"_userId":  userID,
				"_active":  true,
				"deviceId": bson.M{"$regex": bson.RegEx{Pattern: `^InsOmn`}},
			}

			var omnipodResult bson.M
			omnipodDocCursor := dataSession.C().Find(selector).Iter()
			for omnipodDocCursor.Next(&omnipodResult) {
				expectedID := JellyfishIDHash(omnipodResult)
				expectedObjectID := JellyfishObjectIDHash(omnipodResult)

				if expectedID != omnipodResult["id"] {
					logger.Debugf("Expected Omnipod Document ID to be %s, got %s", expectedID, omnipodResult["id"])
					dupQuery := bson.M{
						"_userId":  userID,
						"_active":  true,
						"time":     omnipodResult["time"],
						"type":     omnipodResult["type"],
						"deviceId": omnipodResult["deviceId"],
						"id":       expectedID,
						"_groupId": omnipodResult["_groupId"],
					}
					dupCursor := dataSession.C().Find(dupQuery).Iter()
					if dupCursor.Done() {
						// No duplicate. Update the ID Hashes.
						logger.Debugf("Changing Omnipod Document ID %s to %s (type: %s)", omnipodResult["id"], expectedID, omnipodResult["type"])
						logger.Debugf("Changing _id to %s", expectedObjectID)
						if m.DryRun() {
							hashUpdatedCount++
						} else {
							update := bson.M{
								"$set": bson.M{
									"_id": expectedObjectID,
									"id":  expectedID,
								},
							}

							var changeInfo *mgo.ChangeInfo
							changeInfo, err = dataSession.C().UpdateAll(bson.M{"_id": omnipodResult["_id"]}, update)
							if err != nil {
								logger.WithError(err).Errorf("Could not update ID Hashes for Omnipod Document ID %s.", omnipodResult["id"])
							}
							if changeInfo != nil {
								hashUpdatedCount += changeInfo.Updated
							}
						}
					} else {
						// Got a duplicate. Archive the document with the incorrect ID.
						logger.Debugf("Archiving Omnipod Document ID %s", omnipodResult["id"])

						var dupResult bson.M
						var updateDupObjectID bool
						if dupCursor.Next(&dupResult) {
							// Jellyfish de-duplicates based on the generated ObjectID.
							// If we found a duplicate, we also need to make sure that the ObjectID of
							// the document we're keeping matches what Jellyfish expects it to be.
							updateDupObjectID = (dupResult["_id"] != expectedObjectID)
						}

						if m.DryRun() {
							archivedCount++
							if updateDupObjectID {
								logger.Debugf("Updating Object ID %s to %s", dupResult["_id"], expectedObjectID)
							}
						} else {
							archiveUpdate := bson.M{
								"$set": bson.M{
									"_active":      false,
									"archivedTime": time.Now().Truncate(time.Millisecond).Format(time.RFC3339Nano),
								},
							}

							changeInfo, err := dataSession.C().UpdateAll(bson.M{"_id": omnipodResult["_id"]}, archiveUpdate)
							if err != nil {
								logger.WithError(err).Errorf("Could not archive Omnipod Document ID %s.", omnipodResult["id"])
							}

							if updateDupObjectID {
								updateObjectID := bson.M{
									"$set": bson.M{
										"_id": expectedObjectID,
									},
								}
								err = dataSession.C().UpdateId(dupResult["_id"], updateObjectID)
								if err != nil {
									logger.WithError(err).Errorf("Could not update Object ID %s to %s.", dupResult["_id"], expectedObjectID)
								}
							}

							if changeInfo != nil {
								archivedCount += changeInfo.Updated
							}
						}
					}
					err = dupCursor.Close()
				}
			}
			if omnipodDocCursor.Timeout() {
				logger.WithError(err).Error("Got a cursor timeout. Please re-run to complete the migration.")
			}
			err = omnipodDocCursor.Close()
		}
	}

	if err != nil {
		logger.WithError(err).Error("Unable to migrate Omnipod documents")
	}

	return hashUpdatedCount, archivedCount
}
