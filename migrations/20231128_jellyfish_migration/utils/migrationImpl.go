package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type migrationImpl struct {
	ctx            context.Context
	dataC          *mongo.Collection
	writeBatchSize *int64
	client         *mongo.Client
	config         *MigrationConfig
	updates        []mongo.WriteModel
	groupedDiffs   map[string][]UpdateData
	groupedErrors  groupedErrors
	rawData        []bson.M
	errorsCount    int
	updatedCount   int
	lastUpdatedId  string
	startedAt      time.Time
}

type groupedErrors map[string][]ErrorData

func NewMigration(ctx context.Context, config *MigrationConfig, client *mongo.Client, dataC *mongo.Collection, lastID *string) (Migration, error) {
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

	m := &migrationImpl{
		ctx:           ctx,
		dataC:         dataC,
		client:        client,
		config:        config,
		updates:       []mongo.WriteModel{},
		rawData:       []bson.M{},
		groupedErrors: groupedErrors{},
		groupedDiffs:  map[string][]UpdateData{},
		errorsCount:   0,
		updatedCount:  0,
		startedAt:     time.Now(),
	}
	if lastID != nil {
		m.lastUpdatedId = *lastID
	}
	return m, nil
}

func (m *migrationImpl) Initialize() error {
	if err := m.checkFreeSpace(m.ctx, m.dataC); err != nil {
		return err
	}
	if err := m.setWriteBatchSize(m.ctx); err != nil {
		return err
	}
	return nil
}

func (m *migrationImpl) GetCtx() context.Context {
	return m.ctx
}

func (m *migrationImpl) GetSettings() Settings {
	return Settings{
		DryRun:              m.config.dryRun,
		Rollback:            m.config.rollback,
		RollbackSectionName: m.config.rollbackSectionName,
		Cap:                 m.config.cap,
		StopOnErr:           m.config.stopOnErr,
		WriteBatchSize:      m.writeBatchSize,
	}
}

func (m *migrationImpl) GetDataCollection() *mongo.Collection {
	return m.dataC
}

func (m *migrationImpl) Execute(
	selector bson.M,
	selectorOpt *options.FindOptions,
	queryFn MigrationQueryFn,
	updateFn MigrationUpdateFn) error {
	for queryFn(m, selector, selectorOpt) {
		var err error
		m.updatedCount, err = updateFn(m)
		if err != nil {
			m.writeErrors(nil)
			return err
		}
		if m.capReached() {
			break
		}
	}
	m.getStats().report()
	m.writeErrors(nil)
	m.writeAudit(nil)
	return nil
}

func (d UpdateData) GetMongoUpdates(rollback bool, rollbackSectionName string) []mongo.WriteModel {
	updates := []mongo.WriteModel{}
	for _, u := range d.Apply {
		updateOp := mongo.NewUpdateOneModel()
		updateOp.Filter = d.Filter
		updateOp.SetUpdate(u)
		updates = append(updates, updateOp)
	}
	updateOp := mongo.NewUpdateOneModel()
	updateOp.Filter = d.Filter
	if !rollback && len(d.Revert) > 0 {
		updateOp.SetUpdate(bson.M{"$set": bson.M{rollbackSectionName: d.Revert}})
	} else if rollback {
		updateOp.SetUpdate(bson.M{"$unset": bson.M{rollbackSectionName: ""}})
	}
	updates = append(updates, updateOp)
	return updates
}

func (m *migrationImpl) SetUpdates(data UpdateData) {
	m.groupedDiffs[data.ItemType] = append(m.groupedDiffs[data.ItemType], data)
	m.updates = append(m.updates, data.GetMongoUpdates(m.config.rollback, m.config.rollbackSectionName)...)
}

func (m *migrationImpl) ResetUpdates() {
	m.updates = []mongo.WriteModel{}
}

func (m *migrationImpl) GetUpdates() []mongo.WriteModel {
	return m.updates
}

func (m *migrationImpl) SetLastProcessed(lastID string) {
	m.lastUpdatedId = lastID
	writeLastProcessed(m.lastUpdatedId)
}

func (m *migrationImpl) SetFetched(raw []bson.M) {
	m.rawData = append(m.rawData, raw...)
}

func (m *migrationImpl) getStats() MigrationStats {
	return MigrationStats{
		Errored: m.errorsCount,
		Fetched: len(m.rawData),
		ToApply: len(m.updates),
		Applied: m.updatedCount,
		Elapsed: time.Since(m.startedAt).Truncate(time.Millisecond),
	}
}

func (m *migrationImpl) GetLastID() string {
	return m.lastUpdatedId
}

func (m *migrationImpl) OnError(data ErrorData) {
	m.errorsCount++
	m.groupedErrors[data.ItemType] = append(m.groupedErrors[data.ItemType], data)
	var errFormat = "[_id=%s] %s %s\n"

	if m.config.stopOnErr {
		log.Printf(errFormat, data.ItemID, data.Msg, data.Error.Error())
		os.Exit(1)
	}
}

func (m *migrationImpl) UpdateChecks() error {
	if err := m.blockUntilDBReady(m.GetCtx()); err != nil {
		return err
	}
	if err := m.checkFreeSpace(m.GetCtx(), m.GetDataCollection()); err != nil {
		return err
	}
	return nil
}

func (m *migrationImpl) capReached() bool {
	if m.config.cap != nil {
		stats := m.getStats()

		percent := (float64(stats.Fetched) * float64(100)) / float64(*m.config.cap)
		log.Printf("processed %.0f %% of %d records and applied %d changes", percent, *m.config.cap, stats.Applied)

		if *m.config.cap <= stats.Applied || *m.config.cap <= stats.Fetched {
			log.Printf("cap [%d] updates applied [%d] fetched [%d]", *m.config.cap, stats.Applied, stats.Fetched)
			return true
		}
	}
	return false
}

func (c MigrationStats) report() {
	if c.Applied == 0 && c.Fetched > 0 {
		log.Printf("elapsed [%s] for [%d] items fetched with [%d] errors\n", c.Elapsed, c.Fetched, c.Errored)
		return
	}
	log.Printf("elapsed [%s] for [%d] items migrated with [%d] errors\n", c.Elapsed, c.Applied, c.Errored)
}

func (m *migrationImpl) getOplogCollection() *mongo.Collection {
	return m.client.Database("local").Collection("oplog.rs")
}

func (m *migrationImpl) getAdminDB() *mongo.Database {
	return m.client.Database("admin")
}

func (m *migrationImpl) getOplogDuration(ctx context.Context) (time.Duration, error) {
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

func (m *migrationImpl) setWriteBatchSize(ctx context.Context) error {
	var calculateBatchSize = func(oplogSize int, oplogEntryBytes int, oplogMinWindow int, nopPercent int) int64 {
		return int64(math.Floor(float64(oplogSize) / float64(oplogEntryBytes) / float64(oplogMinWindow) / (float64(nopPercent) / 7)))
	}

	if oplogC := m.getOplogCollection(); oplogC != nil {
		type MongoMetaData struct {
			MaxSize int `json:"maxSize"`
		}
		var metaData MongoMetaData
		if err := oplogC.Database().RunCommand(ctx, bson.M{"collStats": "oplog.rs"}).Decode(&metaData); err != nil {
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

func (m *migrationImpl) checkFreeSpace(ctx context.Context, dataC *mongo.Collection) error {
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

func (m *migrationImpl) getWaitTime(ctx context.Context) (float64, error) {
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

func (m *migrationImpl) blockUntilDBReady(ctx context.Context) error {
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

func (m *migrationImpl) writeErrors(groupLimit *int) {
	for group, errors := range m.groupedErrors {
		if groupLimit != nil {
			if len(errors) < *groupLimit {
				continue
			}
		}
		f, err := createFile("error", group, "%s.log")
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		defer f.Close()
		for _, data := range errors {
			errJSON, err := json.Marshal(data)
			if err != nil {
				log.Println(err)
				os.Exit(1)
			}
			f.WriteString(string(errJSON) + "\n")
		}
		m.groupedErrors[group] = []ErrorData{}
	}
}

func (m *migrationImpl) writeAudit(groupLimit *int) {

	if !m.config.dryRun {
		m.groupedDiffs = map[string][]UpdateData{}
		return
	}

	for group, diffs := range m.groupedDiffs {
		if groupLimit != nil {
			if len(diffs) < *groupLimit {
				continue
			}
		}
		f, err := createFile("audit", group, "%s.json")
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		defer f.Close()
		for _, data := range diffs {
			diffJSON, err := json.Marshal(data)
			if err != nil {
				log.Println(err)
				os.Exit(1)
			}
			f.WriteString(string(diffJSON) + "\n")
		}
		m.groupedDiffs[group] = []UpdateData{}
	}
}

func writeLastProcessed(itemID string) {
	if strings.TrimSpace(itemID) != "" {
		f, err := os.Create("./lastProcessedId")
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		defer f.Close()
		f.WriteString(itemID)
	}
}

func createFile(fileType string, dataGroup string, logName string) (*os.File, error) {
	var err error
	if fileType == "" {
		errors.Join(err, errors.New("missing file type"))
	}
	if dataGroup == "" {
		errors.Join(err, errors.New("missing data group"))
	}
	if logName == "" {
		errors.Join(err, errors.New("missing log group"))
	}
	if err != nil {
		return nil, err
	}
	logName = fmt.Sprintf(logName, dataGroup)
	logPath := filepath.Join(".", fileType)

	err = os.MkdirAll(logPath, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return os.OpenFile(logPath+"/"+logName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}
