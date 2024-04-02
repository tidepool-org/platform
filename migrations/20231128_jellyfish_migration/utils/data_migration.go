package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UpdateData struct {
	Filter    interface{} `json:"-"`
	ItemID    string      `json:"_id"`
	UserID    string      `json:"_userId"`
	ItemType  string      `json:"-"`
	Apply     []bson.M    `json:"apply"`
	ApplyLast bson.M      `json:"applyLast"`
	Revert    []bson.M    `json:"revert"`
}

type ErrorData struct {
	Error    error  `json:"error"`
	ItemID   string `json:"_id"`
	ItemType string `json:"-"`
	Msg      string `json:"message,omitempty"`
}

type MigrationStats struct {
	Errored int
	Fetched int
	Applied int
	ToApply int
	Elapsed time.Duration
}

type Settings struct {
	DryRun              bool
	Rollback            bool
	RollbackSectionName string
	StopOnErr           bool
	WriteBatchSize      *int64
	QueryBatchSize      int
	QueryBatchLimit     int

	capacity    *int
	writeToDisk bool
}

func NewSettings(dryRun *bool, stopOnErr *bool, rollback *bool, rollbackSectionName *string, capacity *int, queryBatch *int, queryLimit *int, writeToDisk *bool) *Settings {
	settings := &Settings{
		writeToDisk:         false,
		Rollback:            true,
		RollbackSectionName: "_rollbackMigration",
		DryRun:              true,
		StopOnErr:           true,
		QueryBatchSize:      50,
		QueryBatchLimit:     100,
	}
	if dryRun != nil {
		settings.DryRun = *dryRun
	}
	if stopOnErr != nil {
		settings.StopOnErr = *stopOnErr
	}
	if rollback != nil {
		settings.Rollback = *rollback
	}
	if rollbackSectionName != nil {
		settings.RollbackSectionName = *rollbackSectionName
	}
	if writeToDisk != nil {
		settings.writeToDisk = *writeToDisk
	}
	if queryBatch != nil {
		settings.QueryBatchSize = *queryBatch
	}
	if queryLimit != nil {
		settings.QueryBatchLimit = *queryLimit
	}
	if capacity != nil && *capacity > 0 {
		settings.capacity = capacity
		log.Printf("capped at %d items", *settings.capacity)
	}
	return settings
}

type DataMigration struct {
	ctx                  context.Context
	dataC                *mongo.Collection
	settings             *Settings
	updates              []mongo.WriteModel
	groupedDiffs         map[string][]UpdateData
	groupedErrors        groupedErrors
	errorsCount          int
	updatedCount         int
	fetchedCount         int
	lastUpdatedId        string
	startedAt            time.Time
	mongoInstanceChecker MongoInstanceCheck
}

type DataMigrationQueryFn func(m *DataMigration) bool
type DataMigrationUpdateFn func(m *DataMigration) (int, error)

type groupedErrors map[string][]ErrorData

func NewMigration(ctx context.Context, settings *Settings, checker MongoInstanceCheck, dataC *mongo.Collection, lastID *string) (*DataMigration, error) {
	var err error
	if settings == nil {
		err = errors.Join(err, errors.New("missing required settings"))
	}
	if checker == nil {
		err = errors.Join(err, errors.New("missing required mongo checker"))
	}

	if err != nil {
		return nil, err
	}

	m := &DataMigration{
		ctx:                  ctx,
		dataC:                dataC,
		mongoInstanceChecker: checker,
		settings:             settings,
		updates:              []mongo.WriteModel{},
		groupedErrors:        groupedErrors{},
		groupedDiffs:         map[string][]UpdateData{},
		errorsCount:          0,
		updatedCount:         0,
		fetchedCount:         0,
		startedAt:            time.Now(),
	}
	if lastID != nil {
		m.lastUpdatedId = *lastID
	}
	return m, nil
}

func (m *DataMigration) Initialize() error {
	if err := m.mongoInstanceChecker.CheckFreeSpace(m.ctx, m.dataC); err != nil {
		return err
	}
	if err := m.mongoInstanceChecker.SetWriteBatchSize(m.ctx); err != nil {
		return err
	}
	return nil
}

func (m *DataMigration) GetCtx() context.Context {
	return m.ctx
}

func (m *DataMigration) GetSettings() Settings {
	m.settings.WriteBatchSize = m.mongoInstanceChecker.GetWriteBatchSize()
	return *m.settings
}

func (m *DataMigration) GetDataCollection() *mongo.Collection {
	return m.dataC
}

func (m *DataMigration) Execute(
	queryFn DataMigrationQueryFn,
	updateFn DataMigrationUpdateFn) error {
	for queryFn(m) {
		count, err := updateFn(m)
		if err != nil {
			m.writeErrors(nil)
			return err
		}
		m.updatesApplied(count)
		if m.completeUpdates() {
			break
		}
		m.writeErrors(nil)
		m.writeAudit(nil)
	}
	m.GetStats().report()
	m.writeErrors(nil)
	m.writeAudit(nil)
	return nil
}

func (d UpdateData) getMongoUpdates(rollback bool, rollbackSectionName string) []mongo.WriteModel {
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

func (m *DataMigration) SetUpdates(data UpdateData) {
	m.groupedDiffs[data.ItemType] = append(m.groupedDiffs[data.ItemType], data)
	m.updates = append(m.updates, data.getMongoUpdates(m.settings.Rollback, m.settings.RollbackSectionName)...)
}

func (m *DataMigration) updatesApplied(updatedCount int) {
	m.updates = []mongo.WriteModel{}
	m.updatedCount += updatedCount
}

func (m *DataMigration) completeUpdates() bool {
	capacity := m.settings.capacity
	if capacity != nil {
		stats := m.GetStats()

		percent := (float64(stats.Fetched) * float64(100)) / float64(*capacity)

		log.Printf("processed %.0f %% of %d records and applied %d changes", percent, *capacity, stats.Applied)

		if *capacity <= stats.Applied || *capacity <= stats.Fetched {
			log.Printf("cap [%d] updates applied [%d] fetched [%d]", *capacity, stats.Applied, stats.Fetched)
			return true
		}
	}
	return false
}

func (m *DataMigration) GetUpdates() []mongo.WriteModel {
	return m.updates
}

func (m *DataMigration) SetLastProcessed(lastID string) {
	m.lastUpdatedId = lastID
	m.writeLastProcessed(m.lastUpdatedId)
}

func (m *DataMigration) SetFetched(raw []bson.M) {
	m.fetchedCount += len(raw)
}

func (m *DataMigration) GetStats() MigrationStats {
	return MigrationStats{
		Errored: m.errorsCount,
		Fetched: m.fetchedCount,
		ToApply: len(m.updates),
		Applied: m.updatedCount,
		Elapsed: time.Since(m.startedAt).Truncate(time.Millisecond),
	}
}

func (m *DataMigration) GetLastID() string {
	return m.lastUpdatedId
}

func (m *DataMigration) OnError(data ErrorData) {
	m.errorsCount++
	m.groupedErrors[data.ItemType] = append(m.groupedErrors[data.ItemType], data)
	if m.settings.StopOnErr {
		log.Printf("[_id=%s] %s %s\n", data.ItemID, data.Msg, data.Error.Error())
		os.Exit(1)
	}
}

func (m *DataMigration) CheckMongoInstance() error {
	if err := m.mongoInstanceChecker.BlockUntilDBReady(m.GetCtx()); err != nil {
		return err
	}
	if err := m.mongoInstanceChecker.CheckFreeSpace(m.GetCtx(), m.GetDataCollection()); err != nil {
		return err
	}
	return nil
}

func (c MigrationStats) report() {
	if c.Applied == 0 && c.Fetched > 0 {
		log.Printf("elapsed [%s] for [%d] items fetched with [%d] errors\n", c.Elapsed, c.Fetched, c.Errored)
		return
	}
	log.Printf("elapsed [%s] for [%d] fetched [%d] updates applied with [%d] errors\n", c.Elapsed, c.Fetched, c.Applied, c.Errored)
}

func (m *DataMigration) writeErrors(groupLimit *int) {
	if !m.settings.writeToDisk {
		m.groupedErrors = map[string][]ErrorData{}
	}
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

func (m *DataMigration) writeAudit(groupLimit *int) {
	if !m.settings.writeToDisk || !m.settings.DryRun {
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

func (m *DataMigration) writeLastProcessed(itemID string) {
	if m.settings.writeToDisk {
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
}

func createFile(fileType string, dataGroup string, logName string) (*os.File, error) {

	var err error
	if fileType == "" {
		err = errors.Join(err, errors.New("missing file type"))
	}
	if dataGroup == "" {
		err = errors.Join(err, errors.New("missing data group"))
	}
	if logName == "" {
		err = errors.Join(err, errors.New("missing log group"))
	}
	if err != nil {
		return nil, err
	}

	logName = fmt.Sprintf(logName, dataGroup)
	dateContainer := time.Now().Round(6 * time.Hour).Format("2006-01-02T15-04-05")
	logPath := filepath.Join(".", fileType, dateContainer)

	err = os.MkdirAll(logPath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return os.OpenFile(logPath+"/"+logName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}
