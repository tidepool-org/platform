package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/urfave/cli"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	uri            string
	minOplogWindow int
	// these values are used to determine writes batches, first dividing the oplog's size with the desired duration and
	// expected entry size, then adding a divisor to account for NOP overshoot in the oplog
	expectedOplogEntrySize int
	// how much of the oplog is NOP, this adjusts the batch to account for an oplog that is very change sensitive
	// must be > 0
	// prod 0.6
	// idle 100
	nopPercent int
	// minimum free disk space percent
	minFreePercent int
	readBatchSize  int64
}

type Migration struct {
	cli            *cli.App
	ctx            context.Context
	config         *Config
	client         *mongo.Client
	oplogC         *mongo.Collection
	deviceDataC    *mongo.Collection
	writeBatchSize *int64
	updates        []mongo.WriteModel
}

const oplogName = "oplog.rs"

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	migration := NewMigration(ctx)
	migration.RunAndExit()
}

func NewMigration(ctx context.Context) *Migration {
	return &Migration{
		cli:     cli.NewApp(),
		ctx:     ctx,
		config:  &Config{},
		updates: []mongo.WriteModel{},
	}
}

func (m *Migration) RunAndExit() {
	if err := m.Initialize(); err != nil {
		os.Exit(1)
	}

	m.CLI().Action = func(ctx *cli.Context) error {
		log.Println("before prepare")
		if err := m.prepare(); err != nil {
			return err
		}
		return nil
	}

	if err := m.CLI().Run(os.Args); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}

func (m *Migration) CLI() *cli.App {
	return m.cli
}

func (m *Migration) Initialize() error {

	log.Println("Initialize...")

	m.CLI().Usage = "BACK-37: Migrate all existing data to add required Platform deduplication hash fields"
	m.CLI().Description = "BACK-37: To fully migrate devices from the `jellyfish` upload API to the `platform` upload API"
	m.CLI().Authors = []cli.Author{
		{
			Name:  "J H BATE",
			Email: "jamie@tidepool.org",
		},
	}
	m.CLI().Flags = append(m.CLI().Flags,
		cli.Int64Flag{
			Name:        "batch-size",
			Usage:       "number of records to read each time",
			Destination: &m.config.readBatchSize,
			Value:       3000,
			Required:    false,
		},
		cli.IntFlag{
			Name:        "min-free-percent",
			Usage:       "minimum free disk space percent",
			Destination: &m.config.minFreePercent,
			Value:       10,
			Required:    false,
		},
		cli.IntFlag{
			Name:        "nop-percent",
			Usage:       "how much of the oplog is NOP",
			Destination: &m.config.nopPercent,
			Value:       100,
			Required:    false,
		},
		cli.IntFlag{
			Name:        "oplog-entry-size",
			Usage:       "minimum free disk space percent",
			Destination: &m.config.expectedOplogEntrySize,
			Value:       420,
			Required:    false,
		},
		cli.IntFlag{
			Name:        "oplog-window",
			Usage:       "minimum oplog window in seconds",
			Destination: &m.config.minOplogWindow,
			Value:       28800, // 8hrs
			Required:    false,
		},
		cli.StringFlag{
			Name:        "uri",
			Usage:       "mongo connection URI",
			Destination: &m.config.uri,
			Value:       "mongodb://localhost:27017",
			Required:    false,
		},
	)
	return nil
}

func (m *Migration) prepare() error {
	log.Println("running prepare ...")
	var err error
	m.client, err = mongo.Connect(m.ctx, options.Client().ApplyURI(m.config.uri))
	if err != nil {
		return fmt.Errorf("unable to connect to MongoDB: %w", err)
	}
	defer m.client.Disconnect(m.ctx)

	m.oplogC = m.client.Database("local").Collection(oplogName)
	m.deviceDataC = m.client.Database("data").Collection("deviceData")

	if err := m.checkFreeSpace(); err != nil {
		return err
	}

	err = m.setWriteBatchSize()
	if err != nil {
		return err
	}
	return nil
}

func (m *Migration) setWriteBatchSize() error {
	if m.oplogC != nil {
		log.Println("Getting Write Batch Size...")
		type MongoMetaData struct {
			MaxSize int `json:"maxSize"`
		}
		var metaData MongoMetaData
		if err := m.oplogC.Database().RunCommand(m.ctx, bson.M{"collStats": oplogName}).Decode(&metaData); err != nil {
			return err
		}
		log.Printf("oplogSize... %v", metaData.MaxSize)
		writeBatchSize := int64(math.Floor(
			float64(metaData.MaxSize) /
				float64(m.config.expectedOplogEntrySize) /
				float64(m.config.minOplogWindow) /
				(float64(m.config.nopPercent) / float64(7))))
		log.Printf("writeBatchSize... %v", writeBatchSize)
		m.writeBatchSize = &writeBatchSize
		return nil
	}
	var writeBatchSize = int64(30000)
	log.Printf("MongoDB is not clustered, removing write batch limit, setting to %d documents.", writeBatchSize)
	m.writeBatchSize = &writeBatchSize
	return nil
}

func (m *Migration) checkFreeSpace() error {
	type MongoMetaData struct {
		FsTotalSize int `json:"fsTotalSize"`
		FsUsedSize  int `json:"fsUsedSize"`
	}
	var metaData MongoMetaData
	log.Println("Getting DB free space...")
	err := m.deviceDataC.Database().RunCommand(m.ctx, bson.M{"dbStats": 1}).Decode(&metaData)
	if err != nil {
		return err
	}

	log.Printf("Stats ... %v ", metaData)
	bytesFree := metaData.FsTotalSize - metaData.FsUsedSize
	percentFree := int(math.Floor(float64(bytesFree) / float64(metaData.FsTotalSize) * 100))
	log.Printf("DB disk currently has %d%% (%d) free.", percentFree*100, bytesFree)

	if percentFree > m.config.minFreePercent {
		return fmt.Errorf("error %d%% is  below minimum free space of %d%%", percentFree, m.config.minFreePercent)
	}
	return nil
}
