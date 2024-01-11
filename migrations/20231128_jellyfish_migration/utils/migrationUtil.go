package utils

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MigrationUtilConfig struct {
	//apply no changes
	dryRun bool
	//halt on error
	stopOnErr      bool
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
	// cap for number of items to migrate
	cap *int
}

type migrationUtil struct {
	writeBatchSize *int64
	client         *mongo.Client
	config         *MigrationUtilConfig
	updates        []mongo.WriteModel
	lastUpdatedId  string
}

type MigrationUtil interface {
	Initialize(ctx context.Context, dataC *mongo.Collection) error
	Execute(ctx context.Context, dataC *mongo.Collection, fetchAndUpdateFn func() bool) error
	OnError(reportErr error, id string, msg string)
	SetData(update *mongo.UpdateOneModel, lastID string)
	GetLastID() string
	GetUpdatesCount() int
}

const oplogName = "oplog.rs"

// MigrationUtil helps managed the migration process
// errors written to
func NewMigrationUtil(config *MigrationUtilConfig, client *mongo.Client, lastID *string) (MigrationUtil, error) {
	var err error
	if config == nil {
		err = errors.Join(err, errors.New("missing required configuration"))
	}
	if client == nil {
		err = errors.Join(err, errors.New("missing required mongo client"))
	}

	if err != nil {
		return nil, err
	}

	log.Printf("migration util configuration: %v", config)

	m := &migrationUtil{
		client:  client,
		config:  config,
		updates: []mongo.WriteModel{},
	}
	if lastID != nil {
		m.lastUpdatedId = *lastID
	}
	return m, nil
}

func (m *migrationUtil) Initialize(ctx context.Context, dataC *mongo.Collection) error {
	log.Print("Initialize migrationUtil")
	if err := m.checkFreeSpace(ctx, dataC); err != nil {
		return err
	}
	if err := m.setWriteBatchSize(ctx); err != nil {
		return err
	}
	return nil
}

func (m *migrationUtil) Execute(ctx context.Context, dataC *mongo.Collection, fetchAndUpdateFn func() bool) error {
	totalMigrated := 0
	migrateStart := time.Now()
	for fetchAndUpdateFn() {
		writeStart := time.Now()
		updatedCount, err := m.writeUpdates(ctx, dataC)
		if err != nil {
			log.Printf("failed writing batch: %s", err)
			return err
		}
		log.Printf("4. data write took [%s] for [%d] items", time.Since(writeStart), updatedCount)
		totalMigrated = totalMigrated + updatedCount

		if m.config.cap != nil {
			log.Println("check cap")
			if totalMigrated >= *m.config.cap {
				break
			}
		}
	}
	log.Printf("migration took [%s] for [%d] items ", time.Since(migrateStart), totalMigrated)
	if m.config.dryRun {
		log.Println("dry-run so no changes applied")
	}
	return nil
}

func (m *migrationUtil) SetData(update *mongo.UpdateOneModel, lastID string) {
	m.lastUpdatedId = lastID
	m.updates = append(m.updates, update)
}

func (m *migrationUtil) GetUpdatesCount() int {
	return len(m.updates)
}

func (m *migrationUtil) GetLastID() string {
	return m.lastUpdatedId
}

func NewMigrationUtilConfig(dryRun *bool, stopOnErr *bool, nopPercent *int, cap *int) *MigrationUtilConfig {
	cfg := &MigrationUtilConfig{
		minOplogWindow:         28800, // 8hrs
		minFreePercent:         10,
		expectedOplogEntrySize: 420,

		dryRun:     true,
		stopOnErr:  true,
		nopPercent: 25,
	}
	if dryRun != nil {
		cfg.SetDryRun(*dryRun)
	}
	if stopOnErr != nil {
		cfg.SetStopOnErr(*stopOnErr)
	}
	if nopPercent != nil {
		cfg.SetNopPercent(*nopPercent)
	}
	if cap != nil {
		cfg.cap = cap
	}
	return cfg
}

func (c *MigrationUtilConfig) SetNopPercent(nopPercent int) *MigrationUtilConfig {
	c.nopPercent = nopPercent
	return c
}

func (c *MigrationUtilConfig) SetMinOplogWindow(minOplogWindow int) *MigrationUtilConfig {
	c.minOplogWindow = minOplogWindow
	return c
}
func (c *MigrationUtilConfig) SetExpectedOplogEntrySize(expectedOplogEntrySize int) *MigrationUtilConfig {
	c.expectedOplogEntrySize = expectedOplogEntrySize
	return c
}
func (c *MigrationUtilConfig) SetMinFreePercent(minFreePercent int) *MigrationUtilConfig {
	c.minFreePercent = minFreePercent
	return c
}
func (c *MigrationUtilConfig) SetDryRun(dryRun bool) *MigrationUtilConfig {
	c.dryRun = dryRun
	return c
}
func (c *MigrationUtilConfig) SetStopOnErr(stopOnErr bool) *MigrationUtilConfig {
	c.stopOnErr = stopOnErr
	return c
}

// OnError
// - write error to file `error.log` in directory cli is running in
// - optionally stop the operation if stopOnErr is true in the config
func (m *migrationUtil) OnError(reportErr error, id string, msg string) {
	var errFormat = "[id=%s] %s %s"
	if reportErr != nil {
		f, err := os.OpenFile("error.log",
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		defer f.Close()
		f.WriteString(fmt.Sprintf(errFormat, id, msg, reportErr.Error()))
		if m.config.stopOnErr {
			log.Printf(errFormat, id, msg, reportErr.Error())
			os.Exit(1)
		}
	}
}

func (m *migrationUtil) getOplogCollection() *mongo.Collection {
	return m.client.Database("local").Collection(oplogName)
}

func (m *migrationUtil) getAdminDB() *mongo.Database {
	return m.client.Database("admin")
}

func writeLastItemUpdate(itemID string, dryRun bool) {
	if strings.TrimSpace(itemID) != "" {
		if dryRun {
			log.Printf("dry run so not setting lastUpdatedId %s", itemID)
			return
		}
		f, err := os.Create("./lastUpdatedId")
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		defer f.Close()
		f.WriteString(itemID)
	}
}

func (m *migrationUtil) getOplogDuration(ctx context.Context) (time.Duration, error) {
	type MongoMetaData struct {
		Wall time.Time `json:"wall"`
	}
	if oplogC := m.getOplogCollection(); oplogC != nil {
		var oldest MongoMetaData
		if err := oplogC.FindOne(
			ctx,
			bson.M{"wall": bson.M{"$exists": true}},
			options.FindOne().SetSort(bson.M{"$natural": 1})).Decode(&oldest); err != nil {
			return 0, err
		}

		var newest MongoMetaData
		if err := oplogC.FindOne(
			ctx,
			bson.M{"wall": bson.M{"$exists": true}},
			options.FindOne().SetSort(bson.M{"$natural": -1})).Decode(&newest); err != nil {
			return 0, err
		}
		oplogDuration := newest.Wall.Sub(oldest.Wall)
		return oplogDuration, nil
	}
	log.Println("Not clustered, not retrieving oplog duration.")
	oplogDuration := time.Duration(m.config.minOplogWindow+1) * time.Second
	return oplogDuration, nil

}

func (m *migrationUtil) setWriteBatchSize(ctx context.Context) error {
	var calculateBatchSize = func(oplogSize int, oplogEntryBytes int, oplogMinWindow int, nopPercent int) int64 {
		return int64(math.Floor(float64(oplogSize) / float64(oplogEntryBytes) / float64(oplogMinWindow) / (float64(nopPercent) / 7)))
	}

	if oplogC := m.getOplogCollection(); oplogC != nil {
		type MongoMetaData struct {
			MaxSize int `json:"maxSize"`
		}
		var metaData MongoMetaData
		if err := oplogC.Database().RunCommand(ctx, bson.M{"collStats": oplogName}).Decode(&metaData); err != nil {
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

func (m *migrationUtil) checkFreeSpace(ctx context.Context, dataC *mongo.Collection) error {
	// pass in config and mongo collection being migrated
	if dataC == nil {
		return errors.New("missing required mongo data collection")
	}

	type MongoMetaData struct {
		FsTotalSize int `json:"fsTotalSize"`
		FsUsedSize  int `json:"fsUsedSize"`
	}
	var metaData MongoMetaData
	if dataC != nil {
		if err := dataC.Database().RunCommand(ctx, bson.M{"dbStats": 1}).Decode(&metaData); err != nil {
			return err
		}
		bytesFree := metaData.FsTotalSize - metaData.FsUsedSize
		percentFree := int(math.Floor(float64(bytesFree) / float64(metaData.FsTotalSize) * 100))
		if m.config.minFreePercent > percentFree {
			return fmt.Errorf("error %d%% is  below minimum free space of %d%%", percentFree, m.config.minFreePercent)
		}
		return nil
	}
	return errors.New("could not get deviceData database")
}

func (m *migrationUtil) getWaitTime(ctx context.Context) (float64, error) {
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
	if err := m.getAdminDB().RunCommand(ctx, bson.M{"replSetGetStatus": 1}).Decode(&metaData); err != nil {
		return 0, err
	}

	for _, member := range metaData.Members {
		if member.State < 1 || member.State > 2 || member.Health != 1 || member.Uptime < 120 {
			log.Printf("DB member %s down or not ready.", member.Name)
			return 240, nil
		}
	}

	oplogDuration, err := m.getOplogDuration(ctx)
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

func (m *migrationUtil) blockUntilDBReady(ctx context.Context) error {
	waitTime, err := m.getWaitTime(ctx)
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
		waitTime, err = m.getWaitTime(ctx)
		if err != nil {
			log.Printf("failed getting wait time  %d", time.Duration(waitTime)*time.Second)
			return err
		}
	}
	return nil
}

func (m *migrationUtil) writeUpdates(ctx context.Context, dataC *mongo.Collection) (int, error) {
	if dataC == nil {
		return 0, errors.New("missing required collection to write updates to")
	}
	if len(m.updates) == 0 {
		return 0, nil
	}
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
		if err := m.blockUntilDBReady(ctx); err != nil {
			return updateCount, err
		}
		if err := m.checkFreeSpace(ctx, dataC); err != nil {
			return updateCount, err
		}

		if m.config.dryRun {
			updateCount += len(batch)
			continue
		}
		results, err := dataC.BulkWrite(ctx, batch)
		if err != nil {
			log.Printf("error writing batch updates %v", err)
			return updateCount, err
		}

		updateCount += int(results.ModifiedCount)
		writeLastItemUpdate(m.lastUpdatedId, m.config.dryRun)
	}
	log.Printf("mongo bulk write took %s", time.Since(start))
	return updateCount, nil
}
