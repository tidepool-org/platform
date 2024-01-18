package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/migrations/20231128_jellyfish_migration/utils"
)

type Migration struct {
	ctx           context.Context
	cli           *cli.App
	config        *config
	client        *mongo.Client
	migrationUtil utils.MigrationUtil
}

type config struct {
	cap           int
	uri           string
	dryRun        bool
	stopOnErr     bool
	userID        string
	lastUpdatedId string
	nopPercent    int
}

const DryRunFlag = "dry-run"

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	migration := NewMigration(ctx)
	migration.RunAndExit()
	log.Println("finished migration")
}

func NewMigration(ctx context.Context) *Migration {
	return &Migration{
		config: &config{},
		ctx:    ctx,
		cli:    cli.NewApp(),
	}
}

func (m *Migration) RunAndExit() {
	if err := m.Initialize(); err != nil {
		os.Exit(1)
	}

	m.CLI().Action = func(ctx *cli.Context) error {

		var err error
		m.client, err = mongo.Connect(m.ctx, options.Client().ApplyURI(strings.TrimSpace(m.config.uri)))
		if err != nil {
			return fmt.Errorf("unable to connect to MongoDB: %w", err)
		}
		defer m.client.Disconnect(m.ctx)
		//TODO: just capping while doing test runs, but probably good to have as a general ability
		cap := m.config.cap // while testing
		m.migrationUtil, err = utils.NewMigrationUtil(
			utils.NewMigrationUtilConfig(&m.config.dryRun, &m.config.stopOnErr, &m.config.nopPercent, &cap),
			m.client,
			&m.config.lastUpdatedId,
		)
		if err != nil {
			return fmt.Errorf("unable init migration utils : %w", err)
		}

		if err := m.migrationUtil.Initialize(m.ctx, m.getDataCollection()); err != nil {
			log.Printf("prepare failed: %s", err)
			return err
		}
		if err := m.migrationUtil.Execute(m.ctx, m.getDataCollection(), m.fetchAndUpdateBatch); err != nil {
			log.Printf("execute failed: %s", err)
			return err
		}
		return nil
	}

	if err := m.CLI().Run(os.Args); err != nil {
		if m.client != nil {
			m.client.Disconnect(m.ctx)
		}
		os.Exit(1)
	}
}

func (m *Migration) Initialize() error {
	m.CLI().Usage = "BACK-37: Migrate all existing data to add required Platform deduplication hash fields"
	m.CLI().Description = "BACK-37: To fully migrate devices from the `jellyfish` upload API to the `platform` upload API"
	m.CLI().Authors = []cli.Author{
		{
			Name:  "J H BATE",
			Email: "jamie@tidepool.org",
		},
	}
	m.CLI().Flags = append(m.CLI().Flags,
		cli.BoolFlag{
			Name:        fmt.Sprintf("%s,%s", DryRunFlag, "n"),
			Usage:       "dry run only; do not migrate",
			Destination: &m.config.dryRun,
		},
		cli.BoolFlag{
			Name:        "stop-error",
			Usage:       "stop migration on error",
			Destination: &m.config.stopOnErr,
		},
		cli.IntFlag{
			Name:        "cap",
			Usage:       "max number of records migrate",
			Destination: &m.config.cap,
			Required:    false,
		},
		cli.IntFlag{
			Name:        "nop-percent",
			Usage:       "how much of the oplog is NOP",
			Destination: &m.config.nopPercent,
			Value:       50,
			Required:    false,
		},
		cli.StringFlag{
			Name:        "uri",
			Usage:       "mongo connection URI",
			Destination: &m.config.uri,
			Required:    false,
			//uri string comes from file called `uri`
			FilePath: "./uri",
		},
		cli.StringFlag{
			Name:        "datum-id",
			Usage:       "id of last datum updated",
			Destination: &m.config.lastUpdatedId,
			Required:    false,
			//id of last datum updated read and written to file `lastUpdatedId`
			FilePath: "./lastUpdatedId",
		},
		cli.StringFlag{
			Name:        "user-id",
			Usage:       "id of single user to migrate",
			Destination: &m.config.userID,
			Required:    false,
		},
	)
	return nil
}

func (m *Migration) CLI() *cli.App {
	return m.cli
}

func (m *Migration) getDataCollection() *mongo.Collection {
	return m.client.Database("data").Collection("deviceData")
}

func (m *Migration) onError(errToReport error, id string, msg string) {
	m.migrationUtil.OnError(errToReport, id, msg)
}

func (m *Migration) fetchAndUpdateBatch() bool {

	selector := bson.M{
		"_deduplicator": bson.M{"$exists": false},
	}

	if strings.TrimSpace(m.config.userID) != "" {
		log.Printf("fetching for user %s", m.config.userID)
		selector["_userId"] = m.config.userID
	}

	// jellyfish uses a generated _id that is not an mongo objectId
	idNotObjectID := bson.M{"$not": bson.M{"$type": "objectId"}}

	if lastID := m.migrationUtil.GetLastID(); lastID != "" {
		selector["$and"] = []interface{}{
			bson.M{"_id": bson.M{"$gt": lastID}},
			bson.M{"_id": idNotObjectID},
		}
	} else {
		selector["_id"] = idNotObjectID
	}

	batchSize := int32(5000)
	limit := int64(10000)

	if dataC := m.getDataCollection(); dataC != nil {
		fetchStart := time.Now()

		dDataCursor, err := dataC.Find(m.ctx, selector,
			&options.FindOptions{
				Sort:      bson.M{"_id": 1},
				BatchSize: &batchSize,
				Limit:     &limit,
			},
		)
		if err != nil {
			log.Printf("failed to select data: %s", err)
			return false
		}

		defer dDataCursor.Close(m.ctx)

		log.Printf("fetch [%v] took [%s]", selector, time.Since(fetchStart))

		updateStart := time.Now()

		for dDataCursor.Next(m.ctx) {

			item := bson.M{}
			if err := dDataCursor.Decode(&item); err != nil {
				log.Printf("error decoding data: %s", err)
				return false
			}

			datumID, datumUpdates, err := utils.GetDatumUpdates(item)
			if err != nil {
				m.onError(err, datumID, "failed getting updates")
				continue
			}
			for _, update := range datumUpdates {
				updateOp := mongo.NewUpdateOneModel()
				if update["$rename"] != nil {
					log.Printf("rename op, 2 ops for same datum")
					updateOp.SetFilter(bson.M{"_id": datumID})
				} else {
					updateOp.SetFilter(bson.M{"_id": datumID, "modifiedTime": item["modifiedTime"]})
				}
				updateOp.SetUpdate(update)
				m.migrationUtil.SetData(updateOp, datumID)
			}
		}
		updated, errored := m.migrationUtil.GetUpdateCounts()
		log.Printf("update took [%s] for [%d] items with [%d] errors", time.Since(updateStart), updated, errored)
		return updated > 0
	}
	return false
}
