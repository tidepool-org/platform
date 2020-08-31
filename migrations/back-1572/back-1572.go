package main

import (
	"time"

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
	dataSession *storeStructuredMongo.Session
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

	m.dataSession = dataStore.NewSession("deviceData")
	defer m.dataSession.Close()

	hashUpdatedCount, archivedCount, errorCount := m.migrateOmnipodDocuments()

	m.Logger().Infof("Migrated %d duplicate Omnipod documents", hashUpdatedCount)
	m.Logger().Infof("Archived %d duplicate Omnipod documents", archivedCount)
	m.Logger().Infof("%d errors occurred", errorCount)

	return nil
}

func (m *Migration) migrateOmnipodDocuments() (int, int, int) {
	logger := m.Logger()

	logger.Debug("Finding distinct users")

	var userIDs []string
	var hashUpdatedCount, archivedCount, errorCount int

	err := m.dataSession.C().Find(bson.M{}).Distinct("_userId", &userIDs)
	if err != nil {
		logger.WithError(err).Error("Unable to execute distinct query")
	} else {
		logger.Debugf("Finding Omnipod records for %d users", len(userIDs))

		for _, userID := range userIDs {
			logger.Debugf("Finding Omnipod records for user ID %s", userID)
			selector := bson.M{
				"_userId": userID,
				"_active": true,
				// Don't need to change the IDs for uploads, since uploads aren't de-duped.
				// All uploads have new `time` fields, and therefore won't have collisions.
				// We avoid trying to change the IDs for `upload` fields, because there's a unique index
				// on `upload` types on `uploadId` (UniqueUploadId). This would then require us to _delete_
				// the old `upload` record first, and outright deleting data seems scary.
				"type":     bson.M{"$ne": "upload"},
				"deviceId": bson.M{"$regex": bson.RegEx{Pattern: `^InsOmn`}},
			}

			var omnipodResult bson.M
			omnipodDocCursor := m.dataSession.C().Find(selector).Iter()
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
					dupCursor := m.dataSession.C().Find(dupQuery).Iter()
					if dupCursor.Done() {
						// No duplicate. Update the ID Hashes.
						// Because `_id` is immutable, we need to insert the new document, then make the old one inactive.
						logger.Debugf("Migrating Omnipod Document ID %s to %s (type: %s)", omnipodResult["id"], expectedID, omnipodResult["type"])
						if m.DryRun() {
							hashUpdatedCount++
						} else {
							err = m.migrateDocument(omnipodResult, expectedID, expectedObjectID)

							if err != nil {
								logger.WithError(err).Errorf("Could not migrate Omnipod Document ID %s.", omnipodResult["id"])
								errorCount++
							} else {
								hashUpdatedCount++
							}
						}
					} else {
						// Got a duplicate. Archive the document with the incorrect ID.
						logger.Debugf("Archiving Omnipod Document ID %s", omnipodResult["id"])

						if m.DryRun() {
							archivedCount++
						} else {
							err := m.archiveDocument(omnipodResult["_id"])

							if err != nil {
								logger.WithError(err).Errorf("Could not archive Omnipod Document ID %s.", omnipodResult["id"])
								errorCount++
							} else {
								archivedCount++
							}
						}
					}
					err = dupCursor.Close()
				} else if expectedObjectID != omnipodResult["_id"] {
					logger.Debugf("Migrating Object ID %s to %s", omnipodResult["_id"], expectedObjectID)
					if m.DryRun() {
						hashUpdatedCount++
					} else {
						err = m.migrateDocument(omnipodResult, expectedID, expectedObjectID)

						if err != nil {
							logger.WithError(err).Errorf("Could not migrate Omnipod Object ID %s.", omnipodResult["_id"])
							errorCount++
						} else {
							hashUpdatedCount++
						}
					}
				}
			}
			if omnipodDocCursor.Timeout() {
				logger.WithError(err).Error("Got a cursor timeout. Please re-run to complete the migration.")
				errorCount++
			}
			err = omnipodDocCursor.Close()
		}
	}

	if err != nil {
		logger.WithError(err).Error("Unable to migrate Omnipod documents")
		errorCount++
	}

	return hashUpdatedCount, archivedCount, errorCount
}

func (m *Migration) archiveDocument(objectId interface{}) error {
	archiveUpdate := bson.M{
		"$set": bson.M{
			"_active":       false,
			"_archivedTime": time.Now().UnixNano() / int64(time.Millisecond),
		},
	}

	return m.dataSession.C().UpdateId(objectId, archiveUpdate)
}

func (m *Migration) migrateDocument(originalDocument bson.M, expectedID string, expectedObjectID string) error {
	newDocument := make(bson.M, len(originalDocument))
	for key, value := range originalDocument {
		if key == "id" {
			value = expectedID
		} else if key == "_id" {
			value = expectedObjectID
		}
		newDocument[key] = value
	}

	err := m.dataSession.C().Insert(newDocument)
	if err != nil {
		m.Logger().WithError(err).Errorf("Could not add new document for Omnipod Document ID %s.", newDocument["id"])
		return err
	} else {
		err := m.archiveDocument(originalDocument["_id"])

		if err != nil {
			m.Logger().WithError(err).Errorf("Could not archive Omnipod Object ID %s.", originalDocument["_id"])
			return err
		}
	}

	return nil
}
