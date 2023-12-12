package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/migrations/20231128_jellyfish_migration/utils"
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
	ctx            context.Context
	cli            *cli.App
	config         *Config
	client         *mongo.Client
	writeBatchSize *int64
	updates        []mongo.WriteModel
}

const oplogName = "oplog.rs"
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
		ctx:     ctx,
		cli:     cli.NewApp(),
		config:  &Config{},
		updates: []mongo.WriteModel{},
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

		log.Printf("config %#v", m.config)
		if err := m.prepare(); err != nil {
			log.Printf("prepare failed: %s", err)
			return err
		}
		if err := m.execute(); err != nil {
			log.Printf("execute failed: %s", err)
			return err
		}
		log.Println("finished execute")
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
			Name:  fmt.Sprintf("%s,%s", DryRunFlag, "n"),
			Usage: "dry run only; do not migrate",
		},
		cli.Int64Flag{
			Name:        "batch-size",
			Usage:       "number of records to read each time",
			Destination: &m.config.readBatchSize,
			Value:       30,
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
			Required:    false,
			FilePath:    "./uri",
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
func (m *Migration) getOplogCollection() *mongo.Collection {
	return m.client.Database("local").Collection(oplogName)
}

func (m *Migration) prepare() error {
	if err := m.checkFreeSpace(); err != nil {
		return err
	}
	if err := m.setWriteBatchSize(); err != nil {
		return err
	}
	return nil
}

func (m *Migration) execute() error {
	log.Println("about to run execute")
	totalMigrated := 0
	testingCapSize := 100
	for m.fetchAndUpdateBatch() {
		updatedCount, err := m.writeBatchUpdates()
		if err != nil {
			log.Printf("failed writing batch: %s", err)
			return err
		}
		totalMigrated = totalMigrated + updatedCount
		log.Printf("migrated %d for a total of %d migrated items", updatedCount, totalMigrated)
		if totalMigrated >= testingCapSize {
			log.Println("migrated docs up to cap so exiting")
			break
		}
	}
	return nil
}

func (m *Migration) getOplogDuration() (time.Duration, error) {
	log.Println("checking oplog duration ...")
	type MongoMetaData struct {
		Wall time.Time `json:"wall"`
	}
	if oplogC := m.getOplogCollection(); oplogC != nil {
		var oldest MongoMetaData
		if err := oplogC.FindOne(
			m.ctx,
			bson.M{"wall": bson.M{"$exists": true}},
			options.FindOne().SetSort(bson.M{"$natural": 1})).Decode(&oldest); err != nil {
			return 0, err
		}

		var newest MongoMetaData
		if err := oplogC.FindOne(
			m.ctx,
			bson.M{"wall": bson.M{"$exists": true}},
			options.FindOne().SetSort(bson.M{"$natural": -1})).Decode(&newest); err != nil {
			return 0, err
		}
		oplogDuration := newest.Wall.Sub(oldest.Wall)
		log.Printf("oplog duration is currently: %v", oplogDuration)
		return oplogDuration, nil
	}
	log.Println("Not clustered, not retrieving oplog duration.")
	oplogDuration := time.Duration(m.config.minOplogWindow+1) * time.Second
	return oplogDuration, nil

}

func calculateBatchSize(oplogSize int, oplogEntryBytes int, oplogMinWindow int, nopPercent int) int64 {
	return int64(math.Floor(float64(oplogSize) / float64(oplogEntryBytes) / float64(oplogMinWindow) / (float64(nopPercent) / 7)))
}

func (m *Migration) setWriteBatchSize() error {
	log.Println("set writeBatchSize...")
	if oplogC := m.getOplogCollection(); oplogC != nil {
		type MongoMetaData struct {
			MaxSize int `json:"maxSize"`
		}
		var metaData MongoMetaData
		if err := oplogC.Database().RunCommand(m.ctx, bson.M{"collStats": oplogName}).Decode(&metaData); err != nil {
			return err
		}
		log.Printf("oplog maxSize: %d", metaData.MaxSize)
		writeBatchSize := calculateBatchSize(metaData.MaxSize, m.config.expectedOplogEntrySize, m.config.minOplogWindow, m.config.nopPercent)
		m.writeBatchSize = &writeBatchSize
		log.Printf("writeBatchSize: %d", writeBatchSize)
		return nil
	}
	var writeBatchSize = int64(30000)
	log.Printf("MongoDB is not clustered, removing write batch limit, setting to %d documents.", writeBatchSize)
	m.writeBatchSize = &writeBatchSize
	return nil
}

func (m *Migration) checkFreeSpace() error {
	log.Println("check free space...")
	type MongoMetaData struct {
		FsTotalSize int `json:"fsTotalSize"`
		FsUsedSize  int `json:"fsUsedSize"`
	}
	var metaData MongoMetaData
	if dataC := m.getDataCollection(); dataC != nil {
		if err := dataC.Database().RunCommand(m.ctx, bson.M{"dbStats": 1}).Decode(&metaData); err != nil {
			return err
		}
		log.Printf("dbStats: %#v", metaData)
		bytesFree := metaData.FsTotalSize - metaData.FsUsedSize
		percentFree := int(math.Floor(float64(bytesFree) / float64(metaData.FsTotalSize) * 100))
		log.Printf("DB disk currently has %d%% (%d bytes) free.", percentFree, bytesFree)
		if m.config.minFreePercent > percentFree {
			return fmt.Errorf("error %d%% is  below minimum free space of %d%%", percentFree, m.config.minFreePercent)
		}
		return nil
	}
	return errors.New("could not get deviceData database")
}

func (m *Migration) getWaitTime() (float64, error) {
	log.Println("getting wait time ...")

	type Member struct {
		Name   string `json:"name"`
		Health int    `json:"health"`
		Uptime int    `json:"uptime"`
		State  int    `json:"state"`
	}

	type MongoMetaData struct {
		Members []Member `json:"members"`
	}

	var metaData MongoMetaData
	if err := m.client.Database("admin").RunCommand(m.ctx, bson.M{"replSetGetStatus": 1}).Decode(&metaData); err != nil {
		return 0, err
	}

	for _, member := range metaData.Members {
		if member.State < 1 || member.State > 2 || member.Health != 1 || member.Uptime < 120 {
			log.Printf("DB member %s down or not ready.", member.Name)
			return 240, nil
		}
	}

	oplogDuration, err := m.getOplogDuration()
	if err != nil {
		return 0, err
	}
	if oplogDuration.Seconds() < float64(m.config.minOplogWindow) {
		minOplogWindowTime := time.Duration(m.config.minOplogWindow) * time.Second
		log.Printf("DB oplog shorter than requested duration of %s, currently %s.", minOplogWindowTime, oplogDuration)
		waitTime := float64(m.config.minOplogWindow) - oplogDuration.Seconds()
		waitTime *= 1.15
		if waitTime < 600 {
			waitTime = 600
		}
		return waitTime, nil
	}
	return 0, nil
}

func (m *Migration) blockUntilDBReady() error {
	log.Println("blocking until ready...")
	waitTime, err := m.getWaitTime()
	if err != nil {
		return err
	}
	var totalWait float64
	for waitTime > 0 {
		totalWait += waitTime
		if totalWait > 1800 {
			log.Printf("Long total wait of %s, possibly high load, or sustained DB outage. If neither, adjust NOP_PERCENT to reduce overshoot.", time.Duration(totalWait)*time.Second)
		}
		log.Printf("Sleeping for %d", time.Duration(waitTime)*time.Second)
		time.Sleep(time.Duration(waitTime) * time.Second)
		waitTime, err = m.getWaitTime()
		if err != nil {
			log.Printf("failed getting wait time  %d", time.Duration(waitTime)*time.Second)
			return err
		}
	}
	return nil
}

func (m *Migration) fetchAndUpdateBatch() bool {
	selector := bson.M{
		// jellyfish uses a generated _id that is not an mongo objectId
		"_id":           bson.M{"$not": bson.M{"$type": "objectId"}},
		"_deduplicator": bson.M{"$exists": false},
		// testing based on _userId for jamie+qa3_1@tidepool.org
		"_userId": "5e8cac61-6bef-4728-b490-c1d82087ed9c",
	}
	m.updates = []mongo.WriteModel{}
	if dataC := m.getDataCollection(); dataC != nil {
		dDataCursor, err := dataC.Find(m.ctx, selector,
			&options.FindOptions{Limit: &m.config.readBatchSize},
		)
		if err != nil {
			log.Printf("failed to select data: %s", err)
			return false
		}

		var dDataResult bson.M

		defer dDataCursor.Close(m.ctx)
		for dDataCursor.Next(m.ctx) {
			if err = dDataCursor.Decode(&dDataResult); err != nil {
				log.Printf("failed decoding data: %s", err)
				return false
			}

			datumID, err := utils.GetValidatedString(dDataResult, "_id")
			if err != nil {
				log.Printf("failed getting dutum _id: %s", err)
				return false
			}

			updates, err := utils.GetDatumUpdates(dDataResult)
			if err != nil {
				log.Printf("failed getting datum updates: %s", err)
				return false
			}

			m.updates = append(m.updates, mongo.NewUpdateOneModel().SetFilter(
				bson.M{
					"_id":          datumID,
					"modifiedTime": dDataResult["modifiedTime"],
				}).SetUpdate(bson.M{
				"$set": updates,
			}))
		}
		return len(m.updates) > 0
	}
	return false
}

func (m *Migration) writeBatchUpdates() (int, error) {
	var getBatches = func(chunkSize int) [][]mongo.WriteModel {
		log.Printf("updates to apply count: %d", len(m.updates))
		batches := [][]mongo.WriteModel{}
		for i := 0; i < len(m.updates); i += chunkSize {
			end := i + chunkSize
			if end > len(m.updates) {
				end = len(m.updates)
			}
			batches = append(batches, m.updates[i:end])
		}
		return batches
	}
	updateCount := 0
	log.Printf("write batch size %d", *m.writeBatchSize)
	for _, batch := range getBatches(int(*m.writeBatchSize)) {
		if err := m.blockUntilDBReady(); err != nil {
			log.Printf("writeBatchUpdates-blocking error: %s", err)
			return updateCount, err
		}
		if err := m.checkFreeSpace(); err != nil {
			log.Printf("writeBatchUpdates-freespace error: %s", err)
			return updateCount, err
		}
		log.Printf("updates to write %d", len(batch))

		updateCount += len(batch)
		log.Printf("updates applied so far %d", updateCount)
		if deviceC := m.getDataCollection(); deviceC != nil {
			results, err := deviceC.BulkWrite(m.ctx, batch)
			if err != nil {
				log.Printf("error writing batch updates %v", err)
				return updateCount, err
			}
			log.Printf("update resuts %#v", results)
			updateCount = updateCount + int(results.ModifiedCount)
		}
	}
	log.Printf("applied %d updates", updateCount)
	return updateCount, nil
}
