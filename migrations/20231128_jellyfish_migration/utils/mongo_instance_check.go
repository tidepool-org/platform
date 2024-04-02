package utils

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoInstanceCheckConfig struct {
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
}

func NewMongoInstanceCheckConfig(nopPercent *int) *MongoInstanceCheckConfig {
	cfg := &MongoInstanceCheckConfig{
		minOplogWindow:         28800, // 8hrs
		minFreePercent:         10,
		expectedOplogEntrySize: 420,
		nopPercent:             25,
	}
	if nopPercent != nil {
		cfg.SetNopPercent(*nopPercent)
	}
	return cfg
}

func (c *MongoInstanceCheckConfig) SetNopPercent(nopPercent int) *MongoInstanceCheckConfig {
	c.nopPercent = nopPercent
	return c
}
func (c *MongoInstanceCheckConfig) SetMinOplogWindow(minOplogWindow int) *MongoInstanceCheckConfig {
	c.minOplogWindow = minOplogWindow
	return c
}
func (c *MongoInstanceCheckConfig) SetExpectedOplogEntrySize(expectedOplogEntrySize int) *MongoInstanceCheckConfig {
	c.expectedOplogEntrySize = expectedOplogEntrySize
	return c
}
func (c *MongoInstanceCheckConfig) SetMinFreePercent(minFreePercent int) *MongoInstanceCheckConfig {
	c.minFreePercent = minFreePercent
	return c
}

type MongoInstanceCheck interface {
	BlockUntilDBReady(ctx context.Context) error
	CheckFreeSpace(ctx context.Context, dataC *mongo.Collection) error
	GetWriteBatchSize() *int64
	SetWriteBatchSize(ctx context.Context) error
}

type mongoInstance struct {
	client         *mongo.Client
	config         *MongoInstanceCheckConfig
	writeBatchSize *int64
}

func NewMongoInstanceCheck(client *mongo.Client, config *MongoInstanceCheckConfig) MongoInstanceCheck {
	return &mongoInstance{
		client: client,
		config: config,
	}
}

func (m *mongoInstance) GetWriteBatchSize() *int64 {
	return m.writeBatchSize
}

func (m *mongoInstance) SetWriteBatchSize(ctx context.Context) error {
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

func (m *mongoInstance) CheckFreeSpace(ctx context.Context, dataC *mongo.Collection) error {
	if dataC == nil {
		return errors.New("missing required mongo data collection")
	}

	type MongoMetaData struct {
		FsTotalSize int `json:"fsTotalSize"`
		FsUsedSize  int `json:"fsUsedSize"`
	}
	var metaData MongoMetaData

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

func (m *mongoInstance) BlockUntilDBReady(ctx context.Context) error {
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
func (m *mongoInstance) getWaitTime(ctx context.Context) (float64, error) {
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

func (m *mongoInstance) getOplogCollection() *mongo.Collection {
	return m.client.Database("local").Collection("oplog.rs")
}

func (m *mongoInstance) getAdminDB() *mongo.Database {
	return m.client.Database("admin")
}

func (m *mongoInstance) getOplogDuration(ctx context.Context) (time.Duration, error) {
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
