package utils

import (
	"context"
	"log"
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

type MigrationQueryFn = func(mUtil Migration, selector bson.M, opts ...*options.FindOptions) bool

type MigrationUpdateFn = func(mUtil Migration) (int, error)

type Migration interface {
	Initialize() error
	Execute(selector bson.M, opt *options.FindOptions, queryFn MigrationQueryFn, updateFn MigrationUpdateFn) error
	GetSettings() Settings
	GetLastID() string
	UpdateChecks() error
	GetCtx() context.Context
	GetDataCollection() *mongo.Collection
	OnError(data ErrorData)
	SetUpdates(data UpdateData)
	GetUpdates() []mongo.WriteModel
	ResetUpdates()
	SetLastProcessed(lastID string)
	SetFetched(raw []bson.M)
	GetStats() MigrationStats
}

type Settings struct {
	DryRun              bool
	Rollback            bool
	RollbackSectionName string
	StopOnErr           bool
	Cap                 *int
	WriteBatchSize      *int64
}

type MigrationConfig struct {
	//apply no changes
	dryRun bool
	//rollback the changes that have been applied
	rollback bool
	//name of section with mongo document that stores the original values
	rollbackSectionName string
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

func NewMigrationConfig(dryRun *bool, stopOnErr *bool, rollback *bool, rollbackSectionName *string, nopPercent *int, cap *int) *MigrationConfig {
	cfg := &MigrationConfig{
		minOplogWindow:         28800, // 8hrs
		minFreePercent:         10,
		expectedOplogEntrySize: 420,
		rollback:               true,
		rollbackSectionName:    "_rollbackMigration",
		dryRun:                 true,
		stopOnErr:              true,
		nopPercent:             25,
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
	if nopPercent != nil {
		cfg.SetNopPercent(*nopPercent)
	}
	if cap != nil && *cap > 0 {
		cfg.cap = cap
		log.Printf("capped at %d items", *cfg.cap)
	}
	return cfg
}

func (c *MigrationConfig) SetNopPercent(nopPercent int) *MigrationConfig {
	c.nopPercent = nopPercent
	return c
}
func (c *MigrationConfig) SetMinOplogWindow(minOplogWindow int) *MigrationConfig {
	c.minOplogWindow = minOplogWindow
	return c
}
func (c *MigrationConfig) SetExpectedOplogEntrySize(expectedOplogEntrySize int) *MigrationConfig {
	c.expectedOplogEntrySize = expectedOplogEntrySize
	return c
}
func (c *MigrationConfig) SetMinFreePercent(minFreePercent int) *MigrationConfig {
	c.minFreePercent = minFreePercent
	return c
}
func (c *MigrationConfig) SetDryRun(dryRun bool) *MigrationConfig {
	c.dryRun = dryRun
	return c
}
func (c *MigrationConfig) SetStopOnErr(stopOnErr bool) *MigrationConfig {
	c.stopOnErr = stopOnErr
	return c
}
func (c *MigrationConfig) SetRollback(rollback bool) *MigrationConfig {
	c.rollback = rollback
	return c
}
func (c *MigrationConfig) SetRollbackSectionName(rollbackSectionName string) *MigrationConfig {
	c.rollbackSectionName = rollbackSectionName
	return c
}
