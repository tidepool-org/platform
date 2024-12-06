package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/migrations/20231128_jellyfish_migration/utils"
	"github.com/tidepool-org/platform/pointer"
)

type Migration struct {
	ctx           context.Context
	cli           *cli.App
	config        *config
	client        *mongo.Client
	migrationUtil *utils.DataMigration
}

type config struct {
	recordLimit    int
	mongoURI       string
	dryRun         bool
	stopOnErr      bool
	userID         string
	uploadID       string
	lastUpdatedID  string
	nopPercent     int
	queryBatchSize int
	queryLimit     int
}

const DryRunFlag = "dry-run"
const StopOnErrorFlag = "stop-on-error"
const RecordLimitFlag = "record-limit"
const NopPercentFlag = "nop-percent"
const MongoURIFlag = "uri"
const DatumIDFlag = "datum-id"
const UserIDFlag = "user-id"
const UploadIDFlag = "upload-id"
const QueryLimitFlag = "query-limit"
const QueryBatchFlag = "query-batch"

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	migration := NewJellyfishMigration(ctx)
	migration.RunAndExit()
	log.Println("finished migration")
}

func NewJellyfishMigration(ctx context.Context) *Migration {
	return &Migration{
		config: &config{},
		ctx:    ctx,
		cli:    cli.NewApp(),
	}
}

func (c *config) report() string {
	details := "\nMIGRATION DETAILS:\n"
	details += fmt.Sprintf("- %s\t\t\t[%d]\n", RecordLimitFlag, c.recordLimit)
	details += fmt.Sprintf("- %s? \t\t[%t]\n", DryRunFlag, c.dryRun)
	details += fmt.Sprintf("- %s\t\t[%t]\n", StopOnErrorFlag, c.stopOnErr)
	details += fmt.Sprintf("- %s\t[%s]\n", DatumIDFlag, c.lastUpdatedID)
	details += fmt.Sprintf("- %s\t\t[%s]\n", UserIDFlag, c.userID)
	details += fmt.Sprintf("- %s\t\t[%d]\n", QueryBatchFlag, c.queryBatchSize)
	details += fmt.Sprintf("- %s\t\t[%d]\n", QueryLimitFlag, c.queryLimit)
	details += fmt.Sprintf("- %s\t\t[%s]\n", UploadIDFlag, c.uploadID)
	return details
}

func (m *Migration) RunAndExit() {
	if err := m.Initialize(); err != nil {
		os.Exit(1)
	}

	m.CLI().Action = func(ctx *cli.Context) error {

		var err error
		m.client, err = mongo.Connect(m.ctx, options.Client().ApplyURI(strings.TrimSpace(m.config.mongoURI)))
		if err != nil {
			return fmt.Errorf("unable to connect to MongoDB: %w", err)
		}
		defer m.client.Disconnect(m.ctx)

		dbChecker := utils.NewMongoInstanceCheck(
			m.client,
			utils.NewMongoInstanceCheckConfig(&m.config.nopPercent),
		)

		m.migrationUtil, err = utils.NewMigration(
			m.ctx,
			utils.NewSettings(
				&m.config.dryRun,
				&m.config.stopOnErr,
				&m.config.recordLimit,
				&m.config.queryBatchSize,
				&m.config.queryLimit,
				pointer.FromBool(true),
			),
			dbChecker,
			m.client.Database("data").Collection("deviceData"),
			&m.config.lastUpdatedID,
		)

		log.Println(m.config.report())

		if err != nil {
			return fmt.Errorf("unable to create migration utils : %w", err)
		}

		if err := m.migrationUtil.Initialize(); err != nil {
			log.Printf("prepare failed: %s", err)
			return err
		}
		if err := m.migrationUtil.Execute(utils.JellyfishDataQueryFn, utils.JellyfishDataUpdatesFn); err != nil {
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
			Name:        StopOnErrorFlag,
			Usage:       "stop migration on error",
			Destination: &m.config.stopOnErr,
		},
		cli.IntFlag{
			Name:        RecordLimitFlag,
			Usage:       "max number of records migrate",
			Destination: &m.config.recordLimit,
			Required:    false,
		},
		cli.IntFlag{
			Name:        NopPercentFlag,
			Usage:       "how much of the oplog is NOP",
			Destination: &m.config.nopPercent,
			Value:       50,
			Required:    false,
		},
		cli.StringFlag{
			Name:        MongoURIFlag,
			Usage:       "mongo connection URI",
			Destination: &m.config.mongoURI,
			Required:    false,
			//uri string comes from file called `uri`
			FilePath: "./uri",
		},
		cli.StringFlag{
			Name:        DatumIDFlag,
			Usage:       "id of last datum updated",
			Destination: &m.config.lastUpdatedID,
			Required:    false,
			//id of last datum updated read and written to file `lastProcessedId`
			FilePath: "./lastProcessedId",
		},
		cli.StringFlag{
			Name:        UserIDFlag,
			Usage:       "id of single user to migrate",
			Destination: &m.config.userID,
			Required:    false,
		},
		cli.IntFlag{
			Name:        QueryLimitFlag,
			Usage:       "max number of items to return",
			Destination: &m.config.queryLimit,
			Value:       50000,
			Required:    false,
		},
		cli.IntFlag{
			Name:        QueryBatchFlag,
			Usage:       "max number of items in each query batch",
			Destination: &m.config.queryBatchSize,
			Value:       10000,
			Required:    false,
		},
	)
	return nil
}

func (m *Migration) CLI() *cli.App {
	return m.cli
}
