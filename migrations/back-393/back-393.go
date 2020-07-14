package main

import (
	"context"
	"encoding/base64"
	"os"
	"time"

	"github.com/urfave/cli"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/crypto"
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

	m.CLI().Usage = "BACK393: Add sharedId to existing gatekeeper.perms documents"
	m.CLI().Description = "BACK393: Gatekeeper.perms records which accounts are shared with whom.\n" +
		"   It encrypts the user id of the shared account for some unknown reasson.\n" +
		"   This migration adds a new field, sharerId, which contains the unencrypted value of the shared user id."
	m.CLI().Authors = []cli.Author{
		{
			Name:  "Derrick Burns",
			Email: "derrick@tidepool.org",
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
	m.Logger().Debug("Add sharerId to gatekeeper. ")

	mongoConfig := m.NewMongoConfig()
	mongoConfig.Database = "gatekeeper"
	mongoConfig.Timeout = 60 * time.Minute
	params := storeStructuredMongo.Params{DatabaseConfig: mongoConfig}
	dataStore, err := storeStructuredMongo.NewStore(params)
	if err != nil {
		return errors.Wrap(err, "unable to create data store")
	}
	defer dataStore.Terminate(nil)

	m.Logger().Debug("Creating data repository")

	dataRepository := dataStore.GetRepository("perms")

	numChanged := m.addSharerID(dataRepository)

	m.Logger().Infof("Updated %d shares", numChanged)

	return nil
}

// UserIDFromGroupID decrypt userid
func UserIDFromGroupID(groupID string, secret string) (string, error) {
	if groupID == "" {
		return "", errors.New("group id is missing")
	}
	if secret == "" {
		return "", errors.New("secret is missing")
	}

	groupIDBytes, err := base64.StdEncoding.DecodeString(groupID)
	if err != nil {
		return "", errors.New("unable to decode with Base64")
	}

	userIDBytes, err := crypto.DecryptWithAES256UsingPassphrase(groupIDBytes, []byte(secret))
	if err != nil {
		return "", errors.New("unable to decrypt with AES-256 using passphrase")
	}

	return string(userIDBytes), nil
}

func (m *Migration) addSharerID(dataRepository *storeStructuredMongo.Repository) int {
	logger := m.Logger()

	logger.Debug("Finding shares")

	type doc struct {
		ID      primitive.ObjectID `bson:"_id"`
		GroupID string             `bson:"groupId"`
	}
	docs := make([]doc, 0)
	var numChanged int

	secret := os.Getenv("GATEKEEPER_SECRET")
	opts := options.Find().SetProjection(bson.M{"_id": 1, "groupId": 1})
	c, err := dataRepository.Find(context.Background(), bson.M{}, opts)
	if err != nil {
		logger.WithError(err).Error("Unable to find any shares")
	} else {
		c.All(context.Background(), &docs)
		logger.Infof("Found %d shares", len(docs))
		for _, doc := range docs {
			logger.Debugf("Updating document id %s, groupID %s", doc.ID, doc.GroupID)

			sharerID, err := UserIDFromGroupID(doc.GroupID, secret)
			if err != nil {
				logger.WithError(err).Error("failed to decode groupId")
				continue
			}
			change := bson.M{"$set": bson.M{"sharerId": sharerID}}
			var result interface{}
			opts := options.FindOneAndUpdate().SetUpsert(true)
			err = dataRepository.FindOneAndUpdate(context.Background(), bson.M{"_id": doc.ID}, change, opts).Decode(&result)

			if err != nil {
				logger.WithError(err).Errorf("Could not update share ID %s", doc.ID)
				continue
			}
			numChanged++
		}
	}
	return numChanged
}
