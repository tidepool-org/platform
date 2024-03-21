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
	"go.mongodb.org/mongo-driver/mongo/options"
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
	Cap                 *int
	WriteBatchSize      *int64
}

type DataMigrationConfig struct {
	writeToDisk bool
	//apply no changes
	dryRun bool
	//rollback the changes that have been applied
	rollback bool
	//name of section with mongo document that stores the original values
	rollbackSectionName string
	//halt on error
	stopOnErr bool
	// cap for number of items to migrate
	cap *int
}

func NewDataMigrationConfig(dryRun *bool, stopOnErr *bool, rollback *bool, rollbackSectionName *string, cap *int, writeToDisk *bool) *DataMigrationConfig {
	cfg := &DataMigrationConfig{
		writeToDisk:         false,
		rollback:            true,
		rollbackSectionName: "_rollbackMigration",
		dryRun:              true,
		stopOnErr:           true,
	}
	if dryRun != nil {
		cfg.SetDryRun(*dryRun)
	}
	if stopOnErr != nil {
		cfg.SetStopOnErr(*stopOnErr)
	}
	if rollback != nil {
		cfg.SetRollback(*rollback)
	}
	if rollbackSectionName != nil {
		cfg.SetRollbackSectionName(*rollbackSectionName)
	}
	if cap != nil && *cap > 0 {
		cfg.cap = cap
		log.Printf("capped at %d items", *cfg.cap)
	}
	return cfg
}

func (c *DataMigrationConfig) SetDryRun(dryRun bool) *DataMigrationConfig {
	c.dryRun = dryRun
	return c
}
func (c *DataMigrationConfig) SetStopOnErr(stopOnErr bool) *DataMigrationConfig {
	c.stopOnErr = stopOnErr
	return c
}
func (c *DataMigrationConfig) SetRollback(rollback bool) *DataMigrationConfig {
	c.rollback = rollback
	return c
}
func (c *DataMigrationConfig) SetRollbackSectionName(rollbackSectionName string) *DataMigrationConfig {
	c.rollbackSectionName = rollbackSectionName
	return c
}
func (c *DataMigrationConfig) SetWriteToDisk(writeToDisk bool) *DataMigrationConfig {
	c.writeToDisk = writeToDisk
	return c
}

type DataMigration struct {
	ctx                  context.Context
	dataC                *mongo.Collection
	config               *DataMigrationConfig
	updates              []mongo.WriteModel
	groupedDiffs         map[string][]UpdateData
	groupedErrors        groupedErrors
	rawData              []bson.M
	errorsCount          int
	updatedCount         int
	lastUpdatedId        string
	startedAt            time.Time
	mongoInstanceChecker MongoInstanceCheck
}

type DataMigrationQueryFn func(m *DataMigration, selector bson.M, opts ...*options.FindOptions) bool
type DataMigrationUpdateFn func(m *DataMigration) (int, error)

type groupedErrors map[string][]ErrorData

func NewMigration(ctx context.Context, config *DataMigrationConfig, checker MongoInstanceCheck, dataC *mongo.Collection, lastID *string) (*DataMigration, error) {
	var err error
	if config == nil {
		err = errors.Join(err, errors.New("missing required configuration"))
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
		config:               config,
		updates:              []mongo.WriteModel{},
		rawData:              []bson.M{},
		groupedErrors:        groupedErrors{},
		groupedDiffs:         map[string][]UpdateData{},
		errorsCount:          0,
		updatedCount:         0,
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
	return Settings{
		DryRun:              m.config.dryRun,
		Rollback:            m.config.rollback,
		RollbackSectionName: m.config.rollbackSectionName,
		Cap:                 m.config.cap,
		StopOnErr:           m.config.stopOnErr,
		WriteBatchSize:      m.mongoInstanceChecker.GetWriteBatchSize(),
	}
}

func (m *DataMigration) GetDataCollection() *mongo.Collection {
	return m.dataC
}

func (m *DataMigration) Execute(
	selector bson.M,
	selectorOpt *options.FindOptions,
	queryFn DataMigrationQueryFn,
	updateFn DataMigrationUpdateFn) error {
	for queryFn(m, selector, selectorOpt) {
		count, err := updateFn(m)
		if err != nil {
			m.writeErrors(nil)
			return err
		}
		m.updatedCount += count
		if m.capReached() {
			break
		}
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
	m.updates = append(m.updates, data.getMongoUpdates(m.config.rollback, m.config.rollbackSectionName)...)
}

func (m *DataMigration) ResetUpdates() {
	m.updates = []mongo.WriteModel{}
}

func (m *DataMigration) GetUpdates() []mongo.WriteModel {
	return m.updates
}

func (m *DataMigration) SetLastProcessed(lastID string) {
	m.lastUpdatedId = lastID
	m.writeLastProcessed(m.lastUpdatedId)
}

func (m *DataMigration) SetFetched(raw []bson.M) {
	m.rawData = append(m.rawData, raw...)
}

func (m *DataMigration) GetStats() MigrationStats {
	return MigrationStats{
		Errored: m.errorsCount,
		Fetched: len(m.rawData),
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
	var errFormat = "[_id=%s] %s %s\n"

	if m.config.stopOnErr {
		log.Printf(errFormat, data.ItemID, data.Msg, data.Error.Error())
		os.Exit(1)
	}
}

func (m *DataMigration) UpdateChecks() error {
	if err := m.mongoInstanceChecker.BlockUntilDBReady(m.GetCtx()); err != nil {
		return err
	}
	if err := m.mongoInstanceChecker.CheckFreeSpace(m.GetCtx(), m.GetDataCollection()); err != nil {
		return err
	}
	return nil
}

func (m *DataMigration) capReached() bool {
	if m.config.cap != nil {
		stats := m.GetStats()

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
	log.Printf("elapsed [%s] for [%d] fetched [%d] updates applied with [%d] errors\n", c.Elapsed, c.Fetched, c.Applied, c.Errored)
}

func (m *DataMigration) writeErrors(groupLimit *int) {
	if m.config.writeToDisk {
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
}

func (m *DataMigration) writeAudit(groupLimit *int) {

	if !m.config.dryRun {
		m.groupedDiffs = map[string][]UpdateData{}
		return
	}
	if m.config.writeToDisk {

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
}

func (m *DataMigration) writeLastProcessed(itemID string) {
	if m.config.writeToDisk {
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
