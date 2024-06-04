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
	cap            int
	uri            string
	dryRun         bool
	stopOnErr      bool
	userID         string
	lastUpdatedID  string
	nopPercent     int
	queryBatchSize int
	queryLimit     int
}

const DryRunFlag = "dry-run"

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
	details += fmt.Sprintf("- CAP\t\t\t[%d]\n", c.cap)
	details += fmt.Sprintf("- AUDIT? \t\t[%t]\n", c.dryRun)
	details += fmt.Sprintf("- STOP ON ERROR\t\t[%t]\n", c.stopOnErr)
	details += fmt.Sprintf("- LAST PROCESSED ID\t[%s]\n", c.lastUpdatedID)
	details += fmt.Sprintf("- USER ID\t\t[%s]\n", c.userID)
	details += fmt.Sprintf("- QUERY BATCH\t\t[%d]\n", c.queryBatchSize)
	details += fmt.Sprintf("- QUERY LIMIT\t\t[%d]\n", c.queryLimit)
	return details
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

		dbChecker := utils.NewMongoInstanceCheck(
			m.client,
			utils.NewMongoInstanceCheckConfig(&m.config.nopPercent),
		)

		m.migrationUtil, err = utils.NewMigration(
			m.ctx,
			utils.NewSettings(
				&m.config.dryRun,
				&m.config.stopOnErr,
				&m.config.cap,
				&m.config.queryBatchSize,
				&m.config.queryLimit,
				pointer.FromBool(true),
			),
			dbChecker,
			m.client.Database("data").Collection("deviceData"),
			&m.config.lastUpdatedID,
		)

		log.Printf("%s", m.config.report())

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
			Destination: &m.config.lastUpdatedID,
			Required:    false,
			//id of last datum updated read and written to file `lastProcessedId`
			FilePath: "./lastProcessedId",
		},
		cli.StringFlag{
			Name:        "user-id",
			Usage:       "id of single user to migrate",
			Destination: &m.config.userID,
			Required:    false,
		},
		cli.IntFlag{
			Name:        "query-limit",
			Usage:       "max number of items to return",
			Destination: &m.config.queryLimit,
			Value:       50000,
			Required:    false,
		},
		cli.IntFlag{
			Name:        "query-batch",
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
