package main

import (
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/urfave/cli"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/session"
	"github.com/tidepool-org/platform/store/mongo"
	mongoTool "github.com/tidepool-org/platform/tool/mongo"
)

const (
	SecretFlag = "secret"
	DryRunFlag = "dry-run"
)

func main() {
	application.Run(NewTool())
}

type Tool struct {
	*mongoTool.Tool
	secret string
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

	t.CLI().Usage = "Migrate all sessions in database to expanded form"
	t.CLI().Authors = []cli.Author{
		{
			Name:  "Darin Krauss",
			Email: "darin@tidepool.org",
		},
	}
	t.CLI().Flags = append(t.CLI().Flags,
		cli.StringFlag{
			Name:  SecretFlag,
			Usage: "session store secret",
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

	t.secret = t.ConfigReporter().WithScopes("session", "store").GetWithDefault("secret", "")

	t.secret = context.String(SecretFlag)
	t.dryRun = context.Bool(DryRunFlag)

	return true
}

func (t *Tool) execute() error {
	if t.secret == "" {
		return errors.New("main", "secret is missing")
	}

	t.Logger().Debug("Migrating sessions to expanded form")

	t.Logger().Debug("Creating sessions store")

	mongoConfig := t.MongoConfig().Clone()
	mongoConfig.Database = "user"
	mongoConfig.Collection = "tokens"
	sessionsStore, err := mongo.New(t.Logger(), mongoConfig)
	if err != nil {
		return errors.Wrap(err, "main", "unable to create sessions store")
	}
	defer sessionsStore.Close()

	t.Logger().Debug("Creating sessions sessions")

	iterateSessionsSession := sessionsStore.NewSession(t.Logger())
	defer iterateSessionsSession.Close()

	updateSessionsSession := sessionsStore.NewSession(t.Logger())
	defer updateSessionsSession.Close()

	t.Logger().Debug("Iterating sessions")

	iter := iterateSessionsSession.C().Find(bson.M{}).Iter()

	expiredTime := time.Now().Unix()
	expiredSessionCount := 0
	migratedSessionCount := 0
	session := &session.Session{}
	for iter.Next(session) {

		if t.isSessionExpanded(session) {
			continue
		}

		sessionLogger := t.Logger().WithField("session", session)

		sessionID := session.ID
		if sessionID == "" {
			sessionLogger.Warn("Missing session id in result from sessions query")
			continue
		}

		if err = t.expandSession(session, t.secret); err != nil {
			sessionLogger.WithError(err).Error("Unable to expand session")
			continue
		}

		if session.ExpiresAt < expiredTime {
			if !t.dryRun {
				if err = updateSessionsSession.C().RemoveId(sessionID); err != nil {
					sessionLogger.WithError(err).Error("Unable to remove session")
					continue
				}
			}

			expiredSessionCount++

			sessionLogger.Debugf("Expired session (expired %d seconds ago)", expiredTime-session.ExpiresAt)
		} else {
			if !t.dryRun {
				if err = updateSessionsSession.C().UpdateId(sessionID, session); err != nil {
					sessionLogger.WithError(err).Error("Unable to update session")
					continue
				}
			}

			migratedSessionCount++

			sessionLogger.Debugf("Migrated session (expires %d second from now)", session.ExpiresAt-expiredTime)
		}
	}
	if err = iter.Close(); err != nil {
		return errors.Wrap(err, "main", "unable to iterate sessions")
	}

	if !t.dryRun {
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
		var count int
		if count, err = iterateSessionsSession.C().Find(selector).Count(); err != nil {
			t.Logger().WithError(err).Error("Unable to query for unexpanded sessions")
		} else if count != 0 {
			t.Logger().WithField("count", count).Error("Found unexpanded sessions")
		}
	}

	t.Logger().Infof("Deleted %d expired sessions and migrated %d sessions to expanded form", expiredSessionCount, migratedSessionCount)

	return nil
}

func (t *Tool) isSessionExpanded(session *session.Session) bool {
	return session.Duration != 0
}

func (t *Tool) expandSession(session *session.Session, secret string) error {
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
			return errors.Wrap(err, "main", "unexpected error")
		}
		if validationError.Errors != jwt.ValidationErrorExpired {
			return errors.Wrap(validationError, "main", "unexpected validation error")
		}
	}

	session.IsServer = parsedClaims.IsServer == "yes"
	if session.IsServer {
		session.ServerID = parsedClaims.UserID
	} else {
		session.UserID = parsedClaims.UserID
	}
	session.Duration = int64(parsedClaims.Duration)
	session.ExpiresAt = parsedClaims.ExpiresAt

	session.CreatedAt = session.Time

	return nil
}
