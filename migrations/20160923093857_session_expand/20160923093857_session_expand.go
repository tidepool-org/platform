package main

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/urfave/cli"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/errors"
	mongoMigration "github.com/tidepool-org/platform/migration/mongo"
	"github.com/tidepool-org/platform/session"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

const (
	SecretFlag = "secret"
)

func main() {
	application.RunAndExit(NewMigration())
}

type Migration struct {
	*mongoMigration.Migration
	secret string
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

	m.CLI().Usage = "Migrate all sessions to expanded form"
	m.CLI().Description = "Migrate all sessions to expanded form, including additional fields such as 'isServer', 'serverId', 'userId', 'duration', 'createdAt', and 'expiresAt'." +
		"\n\n   This migration is idempotent." +
		"\n\n   NOTE: This migration MUST be executed immediately AFTER upgrading Shoreline to v0.9.1."
	m.CLI().Authors = []cli.Author{
		{
			Name:  "Darin Krauss",
			Email: "darin@tidepool.org",
		},
	}
	m.CLI().Flags = append(m.CLI().Flags,
		cli.StringFlag{
			Name:  SecretFlag,
			Usage: "session store secret",
		},
	)

	m.CLI().Action = func(ctx *cli.Context) error {
		if !m.ParseContext(ctx) {
			return nil
		}
		return m.execute()
	}

	return nil
}

func (m *Migration) ParseContext(ctx *cli.Context) bool {
	if parsed := m.Migration.ParseContext(ctx); !parsed {
		return parsed
	}

	m.secret = m.ConfigReporter().WithScopes("session", "store").GetWithDefault("secret", "")

	m.secret = ctx.String(SecretFlag)

	return true
}

func (m *Migration) Secret() string {
	return m.secret
}

func (m *Migration) execute() error {
	if m.Secret() == "" {
		return errors.New("secret is missing")
	}

	m.Logger().Debug("Migrating sessions to expanded form")

	m.Logger().Debug("Creating sessions store")

	mongoConfig := m.NewMongoConfig()
	mongoConfig.Database = "user"
	sessionsStore, err := storeStructuredMongo.NewStore(mongoConfig)
	if err != nil {
		return errors.Wrap(err, "unable to create sessions store")
	}
	defer sessionsStore.Terminate(context.Background())

	m.Logger().Debug("Creating sessions repositories")

	iterateTokenRepository := sessionsStore.GetRepository("tokens")

	updateTokenRepository := sessionsStore.GetRepository("tokens")

	m.Logger().Debug("Iterating sessions")

	cursor, err := iterateTokenRepository.Find(context.Background(), bson.M{})

	now := time.Now()
	expiredSessionCount := 0
	migratedSessionCount := 0
	var repository session.Session
	for cursor.Next(context.Background()) {
		if err = cursor.Decode(&repository); err != nil {
			return errors.Wrap(err, "unable to decode session")
		}

		if m.isSessionExpanded(&repository) {
			continue
		}

		sessionLogger := m.Logger().WithField("session", repository)

		sessionID := repository.ID
		if sessionID == "" {
			sessionLogger.Warn("Missing session id in result from sessions query")
			continue
		}

		if err = m.expandSession(&repository, m.Secret()); err != nil {
			sessionLogger.WithError(err).Error("Unable to expand session")
			continue
		}

		if repository.ExpiresAt.Before(now) {
			if !m.DryRun() {
				if _, err = updateTokenRepository.DeleteOne(context.Background(), bson.M{"_id": sessionID}); err != nil {
					sessionLogger.WithError(err).Error("Unable to remove session")
					continue
				}
			}

			expiredSessionCount++

			sessionLogger.Debugf("Expired session (expired %d seconds ago)", now.Unix()-repository.ExpiresAt.Unix())
		} else {
			if !m.DryRun() {
				if _, err = updateTokenRepository.UpdateOne(context.Background(), bson.M{"_id": sessionID}, repository); err != nil {
					sessionLogger.WithError(err).Error("Unable to update session")
					continue
				}
			}

			migratedSessionCount++

			sessionLogger.Debugf("Migrated session (expires %d second from now)", repository.ExpiresAt.Unix()-now.Unix())
		}
	}

	if !m.DryRun() {
		selector := bson.M{
			"$or": []bson.M{
				{"_id": bson.M{"$exists": false}},
				{"isServer": bson.M{"$exists": false}},
				{"$and": []bson.M{
					{"isServer": true},
					{"serverId": bson.M{"$exists": false}},
				}},
				{"$and": []bson.M{
					{"isServer": false},
					{"userId": bson.M{"$exists": false}},
				}},
				{"duration": bson.M{"$exists": false}},
				{"expiresAt": bson.M{"$exists": false}},
				{"time": bson.M{"$exists": false}},
				{"createdAt": bson.M{"$exists": false}},
			},
		}
		var count int64
		if count, err = iterateTokenRepository.CountDocuments(context.Background(), selector); err != nil {
			m.Logger().WithError(err).Error("Unable to query for unexpanded sessions")
		} else if count != 0 {
			m.Logger().WithField("count", count).Error("Found unexpanded sessions")
		}
	}

	m.Logger().Infof("Deleted %d expired sessions and migrated %d sessions to expanded form", expiredSessionCount, migratedSessionCount)

	return nil
}

func (m *Migration) isSessionExpanded(session *session.Session) bool {
	return session.Duration != 0
}

func (m *Migration) expandSession(session *session.Session, secret string) error {
	parsedClaims := struct {
		jwt.StandardClaims
		IsServer string  `json:"svr"`
		UserID   string  `json:"usr"`
		Duration float64 `json:"dur"`
	}{}

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}
	_, err := jwt.ParseWithClaims(session.ID, &parsedClaims, keyFunc)
	if err != nil {
		validationError, ok := err.(*jwt.ValidationError)
		if !ok {
			return errors.Wrap(err, "unexpected error")
		}
		if validationError.Errors != jwt.ValidationErrorExpired {
			return errors.Wrap(validationError, "unexpected validation error")
		}
	}

	session.IsServer = parsedClaims.IsServer == "yes"
	if session.IsServer {
		session.ServerID = parsedClaims.UserID
	} else {
		session.UserID = parsedClaims.UserID
	}
	session.Duration = int64(parsedClaims.Duration)
	session.ExpiresAt = time.Unix(parsedClaims.ExpiresAt, 0)

	session.CreatedAt = session.Time

	return nil
}
