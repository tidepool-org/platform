package utils

import (
	"context"
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
	Filter interface{} `json:"-"`
	ItemID string      `json:"_id"`
	Apply  []bson.M    `json:"apply"`
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
	DryRun          bool
	StopOnErr       bool
	WriteBatchSize  *int64
	QueryBatchSize  int
	QueryBatchLimit int

	RecordLimit *int
	writeToDisk bool
}

func NewSettings(dryRun *bool, stopOnErr *bool, recordLimit *int, queryBatchSize *int, queryBatchLimit *int, writeToDisk *bool) *Settings {
	settings := &Settings{
		writeToDisk:     false,
		DryRun:          true,
		StopOnErr:       true,
		QueryBatchSize:  50,
		QueryBatchLimit: 100,
	}
	if dryRun != nil {
		settings.DryRun = *dryRun
	}
	if stopOnErr != nil {
		settings.StopOnErr = *stopOnErr
	}

	if writeToDisk != nil {
		settings.writeToDisk = *writeToDisk
	}
	if queryBatchSize != nil {
		settings.QueryBatchSize = *queryBatchSize
	}
	if queryBatchLimit != nil {
		settings.QueryBatchLimit = *queryBatchLimit
	}
	if recordLimit != nil && *recordLimit > 0 {
		settings.RecordLimit = recordLimit
		log.Printf("capped at %d items", *settings.RecordLimit)
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
	lastUpdatedID        *string
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
		lastUpdatedID:        lastID,
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
			m.writeErrors()
			return err
		}
		m.updatesApplied(count)
		if m.completeUpdates() {
			break
		}
		m.writeErrors()
		m.writeAudit()
	}
	m.GetStats().report()
	m.writeErrors()
	m.writeAudit()
	return nil
}

func (d UpdateData) getMongoUpdates() []mongo.WriteModel {
	updates := []mongo.WriteModel{}
	for _, u := range d.Apply {
		updateOp := mongo.NewUpdateOneModel()
		updateOp.Filter = d.Filter
		updateOp.SetUpdate(u)
		updates = append(updates, updateOp)
	}
	return updates
}

func (m *DataMigration) SetUpdates(data UpdateData) {
	m.updates = append(m.updates, data.getMongoUpdates()...)
}

func (m *DataMigration) updatesApplied(updatedCount int) {
	m.updates = []mongo.WriteModel{}
	m.updatedCount += updatedCount
}

func (m *DataMigration) completeUpdates() bool {
	recordLimit := m.settings.RecordLimit
	stats := m.GetStats()
	if recordLimit == nil {
		log.Printf("updates applied [%d] fetched [%d]", stats.Applied, stats.Fetched)
	} else {
		percent := (float64(stats.Fetched) * float64(100)) / float64(*recordLimit)
		log.Printf("processed %.0f %% of %d records and applied %d changes", percent, *recordLimit, stats.Applied)

		if *recordLimit <= stats.Applied || *recordLimit <= stats.Fetched {
			log.Printf("recordLimit [%d] updates applied [%d] fetched [%d]", *recordLimit, stats.Applied, stats.Fetched)
			return true
		}
	}
	return false
}

func (m *DataMigration) GetUpdates() []mongo.WriteModel {
	return m.updates
}

func (m *DataMigration) SetLastProcessed(lastID string) {
	m.lastUpdatedID = &lastID
	m.writeLastProcessed(*m.lastUpdatedID)
}

func (m *DataMigration) UpdateFetchedCount() {
	m.fetchedCount++
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

func (m *DataMigration) GetLastID() *string {
	return m.lastUpdatedID
}

func (m *DataMigration) OnError(data ErrorData) {
	if m.settings.StopOnErr {
		log.Printf("[_id=%s] %s %s\n", data.ItemID, data.Msg, data.Error.Error())
		os.Exit(1)
	}
	m.errorsCount++
	m.groupedErrors[data.ItemType] = append(m.groupedErrors[data.ItemType], data)
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

func (m *DataMigration) writeErrors() {
	if !m.settings.writeToDisk {
		m.groupedErrors = map[string][]ErrorData{}
	}
	for group, errors := range m.groupedErrors {
		logPath := filepath.Join(".", "error", fmt.Sprintf("%s.log", group))
		writeFileData(errors, logPath)
		// f, err := m.createFile("error", group, "%s.log")
		// if err != nil {
		// 	log.Println(err)
		// 	os.Exit(1)
		// }
		// defer f.Close()
		// for _, data := range errors {
		// 	errJSON, err := json.Marshal(data)
		// 	if err != nil {
		// 		log.Println(err)
		// 		os.Exit(1)
		// 	}
		// 	f.WriteString(string(errJSON) + "\n")
		// }
		m.groupedErrors[group] = []ErrorData{}
	}
}

func (m *DataMigration) writeAudit() {
	if !m.settings.writeToDisk || !m.settings.DryRun {
		m.groupedDiffs = map[string][]UpdateData{}
		return
	}
	for group, diffs := range m.groupedDiffs {

		logPath := filepath.Join(".", "audit", fmt.Sprintf("%s.json", group))
		writeFileData(diffs, logPath)

		// f, err := m.createFile("audit", group, "%s.json")
		// if err != nil {
		// 	log.Println(err)
		// 	os.Exit(1)
		// }
		// defer f.Close()
		// for _, data := range diffs {
		// 	diffJSON, err := json.Marshal(data)
		// 	if err != nil {
		// 		log.Println(err)
		// 		os.Exit(1)
		// 	}
		// 	f.WriteString(string(diffJSON) + "\n")
		// }
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

// func (m *DataMigration) createFile(fileType string, dataGroup string, logName string) (*os.File, error) {

// 	var err error
// 	if fileType == "" {
// 		err = errors.Join(err, errors.New("missing file type"))
// 	}
// 	if dataGroup == "" {
// 		err = errors.Join(err, errors.New("missing data group"))
// 	}
// 	if logName == "" {
// 		err = errors.Join(err, errors.New("missing log group"))
// 	}
// 	if err != nil {
// 		return nil, err
// 	}

// 	logName = fmt.Sprintf(logName, dataGroup)
// 	logPath := filepath.Join(".", fileType)
// 	err = os.MkdirAll(logPath, os.ModePerm)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return os.OpenFile(logPath+"/"+logName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// }
