package store

import (
	"log"
	"strconv"
	"time"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

	"github.com/tidepool-org/platform/config"
)

// Generic store interface for all operations that we need
// Note: there will be a specific implementation for the store
// type we are using, which could also be a mock
type Store interface {
	Save(d interface{}) error
	Update(d interface{}) error
	Delete(id string) error
	Read(id string, result interface{}) error
	ReadAll(results []interface{}) error
}

const (
	MONGO_STORE_URL     = "MONGO_URL"
	MONGO_STORE_TIMEOUT = "MONGO_TIMEOUT"
	MONGO_STORE_DB_NAME = "MONGO_DB_NAME"
)

type MongoStore struct {
	Session        *mgo.Session
	DbName         string
	CollectionName string
}

func NewMongoStore(name string) *MongoStore {

	//CONFIG
	url := config.FromEnv(MONGO_STORE_URL)
	timeoutStr := config.FromEnv(MONGO_STORE_TIMEOUT)
	dbName := config.FromEnvWithDefault(MONGO_STORE_DB_NAME, "")

	timeout := 20 * time.Second

	if timeoutStr != "" {
		secs, err := strconv.Atoi(timeoutStr)
		if err != nil {
			log.Fatal(err)
		}
		timeout = time.Duration(secs) * time.Second
	}
	//END CONFIG

	mgoSession, err := mgo.DialWithTimeout(url, timeout)

	if err != nil {
		log.Fatal(err)
	}

	return &MongoStore{Session: mgoSession, CollectionName: name, DbName: dbName}
}

func (this *MongoStore) Cleanup() error {
	return this.Session.DB(this.DbName).C(this.CollectionName).DropCollection()
}

func (this *MongoStore) Save(d interface{}) error {
	cpy := this.Session.Copy()
	defer cpy.Close()

	if err := cpy.DB(this.DbName).C(this.CollectionName).Insert(d); err != nil {
		return err
	}
	return nil
}

func (this *MongoStore) Update(id string, d interface{}) error {
	cpy := this.Session.Copy()
	defer cpy.Close()

	if _, err := cpy.DB(this.DbName).C(this.CollectionName).Upsert(bson.M{"id": id}, d); err != nil {
		return err
	}
	return nil
}

func (this *MongoStore) Delete(id string) error {

	cpy := this.Session.Copy()
	defer cpy.Close()

	if err := cpy.DB(this.DbName).C(this.CollectionName).Remove(bson.M{"id": id}); err != nil {
		return err
	}
	return nil
}

func (this *MongoStore) Read(id string, result interface{}) error {
	cpy := this.Session.Copy()
	defer cpy.Close()

	if err := cpy.DB(this.DbName).C(this.CollectionName).Find(bson.M{"id": id}).One(&result); err != nil {
		return err
	}
	return nil
}

func (this *MongoStore) ReadAll(results []interface{}) error {
	cpy := this.Session.Copy()
	defer cpy.Close()

	if err := cpy.DB(this.DbName).C(this.CollectionName).Find(bson.M{}).All(&results); err != nil {
		return err
	}
	return nil
}
