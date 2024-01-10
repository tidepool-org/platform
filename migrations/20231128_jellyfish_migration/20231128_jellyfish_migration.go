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
	dryRun         bool
	stopOnErr      bool
	userID         *string
	lastUpdatedId  *string
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
		ctx:       ctx,
		cli:       cli.NewApp(),
		config:    &Config{},
		updates:   []mongo.WriteModel{},
		stopOnErr: false,
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
		if err := m.prepare(); err != nil {
			log.Printf("prepare failed: %s", err)
			return err
		}
		if err := m.execute(); err != nil {
			log.Printf("execute failed: %s", err)
			return err
		}
		return nil
	}

	log.Printf("args %v", os.Args)

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
			Destination: &m.dryRun,
		},
		cli.BoolFlag{
			Name:        "stop-error",
			Usage:       "stop migration on error",
			Destination: &m.stopOnErr,
		},
		cli.Int64Flag{
			Name:        "batch-size",
			Usage:       "number of records to read each time",
			Destination: &m.config.readBatchSize,
			Value:       300,
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
			Value:       50,
			Required:    false,
		},
		cli.IntFlag{
			Name:        "oplog-entry-size",
			Usage:       "expected oplog entry size",
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
			//uri string comes from file called `uri`
			FilePath: "./uri",
		},
		cli.StringFlag{
			Name:        "datum-id",
			Usage:       "id of last datum updated",
			Destination: m.lastUpdatedId,
			Required:    false,
			//id of last datum updated read and written to file `lastUpdatedId`
			FilePath: "./lastUpdatedId",
		},
		cli.StringFlag{
			Name:        "user-id",
			Usage:       "id of single user to migrate",
			Destination: m.userID,
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

func (m *Migration) getOplogCollection() *mongo.Collection {
	return m.client.Database("local").Collection(oplogName)
}

func (m *Migration) onError(errToReport error, id string, msg string) {
	if errToReport != nil {
		var errFormat = "[id=%s] %s %s"
		f, err := os.OpenFile("error.log",
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		defer f.Close()
		f.WriteString(fmt.Sprintf(errFormat, id, msg, errToReport.Error()))
		writeLastItemUpdate(m.lastUpdatedId, m.dryRun)
		if m.stopOnErr {
			log.Printf(errFormat, id, msg, errToReport.Error())
			os.Exit(1)
		}
	}
}

func writeLastItemUpdate(itemID *string, dryRun bool) {
	if itemID == nil {
		return
	}
	if dryRun {
		log.Printf("dry run so not setting lastUpdatedId %s", *itemID)
		return
	}
	f, err := os.OpenFile("./lastUpdatedId",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer f.Close()
	f.WriteString(*itemID)
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
	totalMigrated := 0
	migrateStart := time.Now()
	for m.fetchAndUpdateBatch() {
		writeStart := time.Now()
		updatedCount, err := m.writeBatchUpdates()
		if err != nil {
			log.Printf("failed writing batch: %s", err)
			return err
		}
		log.Printf("4. data write took [%s] for [%d] items", time.Since(writeStart), updatedCount)
		totalMigrated = totalMigrated + updatedCount
	}
	log.Printf("migration took [%s] for [%d] items ", time.Since(migrateStart), totalMigrated)
	return nil
}

func (m *Migration) getOplogDuration() (time.Duration, error) {
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
		log.Printf("current oplog duration: %v", oplogDuration)
		return oplogDuration, nil
	}
	log.Println("Not clustered, not retrieving oplog duration.")
	oplogDuration := time.Duration(m.config.minOplogWindow+1) * time.Second
	return oplogDuration, nil

}

func (m *Migration) setWriteBatchSize() error {
	// pass in config and mongo oplog collection

	var calculateBatchSize = func(oplogSize int, oplogEntryBytes int, oplogMinWindow int, nopPercent int) int64 {
		return int64(math.Floor(float64(oplogSize) / float64(oplogEntryBytes) / float64(oplogMinWindow) / (float64(nopPercent) / 7)))
	}

	if oplogC := m.getOplogCollection(); oplogC != nil {
		type MongoMetaData struct {
			MaxSize int `json:"maxSize"`
		}
		var metaData MongoMetaData
		if err := oplogC.Database().RunCommand(m.ctx, bson.M{"collStats": oplogName}).Decode(&metaData); err != nil {
			return err
		}
		writeBatchSize := calculateBatchSize(metaData.MaxSize, m.config.expectedOplogEntrySize, m.config.minOplogWindow, m.config.nopPercent)
		m.writeBatchSize = &writeBatchSize
		log.Printf("calculated writeBatchSize: %d", writeBatchSize)
		return nil
	}
	var writeBatchSize = int64(30000)
	log.Printf("MongoDB is not clustered, removing write batch limit, setting to %d documents.", writeBatchSize)
	m.writeBatchSize = &writeBatchSize
	return nil
}

func (m *Migration) checkFreeSpace() error {
	// pass in config and mongo collection being migrated

	type MongoMetaData struct {
		FsTotalSize int `json:"fsTotalSize"`
		FsUsedSize  int `json:"fsUsedSize"`
	}
	var metaData MongoMetaData
	if dataC := m.getDataCollection(); dataC != nil {
		if err := dataC.Database().RunCommand(m.ctx, bson.M{"dbStats": 1}).Decode(&metaData); err != nil {
			return err
		}
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
	// pass config and mongo admin db
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
		"_deduplicator": bson.M{"$exists": false},
	}

	if m.userID != nil {
		log.Print("focused test so we need a user id `--user-id=`")
		selector["_userId"] = *m.userID
	} else {
		log.Print("for testing we need a single user to migrate `--user-id=`")
		return false
	}

	// jellyfish uses a generated _id that is not an mongo objectId
	idNotObjectID := bson.M{"$not": bson.M{"$type": "objectId"}}

	if m.lastUpdatedId != nil {
		selector["$and"] = []interface{}{
			bson.M{"_id": bson.M{"$gt": *m.lastUpdatedId}},
			bson.M{"_id": idNotObjectID},
		}
	} else {
		selector["_id"] = idNotObjectID
	}

	m.updates = []mongo.WriteModel{}

	// TODO: balance with batch write size?
	size := int32(1000)

	if dataC := m.getDataCollection(); dataC != nil {
		fetchStart := time.Now()

		dDataCursor, err := dataC.Find(m.ctx, selector,
			&options.FindOptions{
				Sort:      bson.M{"_id": 1},
				BatchSize: &size,
			},
		)
		if err != nil {
			log.Printf("failed to select data: %s", err)
			return false
		}

		defer dDataCursor.Close(m.ctx)

		log.Printf("1. data fetch [%v] took [%s]", selector, time.Since(fetchStart))

		// decodeStart := time.Now()
		// var dataSet []bson.M
		// if err := dDataCursor.All(m.ctx, &dataSet); err != nil {
		// 	log.Printf("error decoding data: %s", err)
		// 	return false
		// }
		// log.Printf("2. data decode took [%s] for [%d] items", time.Since(decodeStart), len(dataSet))

		updateStart := time.Now()
		// for _, item := range dataSet {

		for dDataCursor.Next(m.ctx) {

			item := bson.M{}
			if err := dDataCursor.Decode(&item); err != nil {
				log.Printf("error decoding data: %s", err)
				return false
			}
			log.Printf("got %v", item)

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
				m.updates = append(m.updates, updateOp)
			}
			m.lastUpdatedId = &datumID
		}

		log.Printf("3. data update took [%s] for [%d] items", time.Since(updateStart), len(m.updates))
		return len(m.updates) > 0
	}
	return false
}

func (m *Migration) writeBatchUpdates() (int, error) {
	if len(m.updates) == 0 {
		return 0, nil
	}
	writeLastItemUpdate(m.lastUpdatedId, m.dryRun)
	start := time.Now()
	var getBatches = func(chunkSize int) [][]mongo.WriteModel {
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
	for _, batch := range getBatches(int(*m.writeBatchSize)) {
		if err := m.blockUntilDBReady(); err != nil {
			log.Printf("writeBatchUpdates-blocking error: %s", err)
			return updateCount, err
		}
		if err := m.checkFreeSpace(); err != nil {
			log.Printf("writeBatchUpdates-freespace error: %s", err)
			return updateCount, err
		}
		log.Printf("batch size to write %d", len(batch))

		if m.dryRun {
			updateCount += len(batch)
			log.Println("dry run so not applying changes")
			continue
		}

		if deviceC := m.getDataCollection(); deviceC != nil {
			results, err := deviceC.BulkWrite(m.ctx, batch)
			if err != nil {
				log.Printf("error writing batch updates %v", err)
				return updateCount, err
			}
			updateCount += int(results.ModifiedCount)
		}
	}
	log.Printf("mongo bulk write took %s", time.Since(start))
	return updateCount, nil
}
