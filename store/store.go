package store

import (
	"time"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/labix.org/v2/mgo"
	"github.com/tidepool-org/platform/Godeps/_workspace/src/labix.org/v2/mgo/bson"

	"github.com/tidepool-org/platform/config"
	log "github.com/tidepool-org/platform/logger"
)

// Generic store interface for all operations that we need
// Note: there will be a specific implementation for the store
// type we are using, which could also be a mock
type Store interface {
	Save(d interface{}) error
	Update(id StoreIdField, d interface{}) error
	Delete(id StoreIdField) error
	Read(id StoreIdField, result interface{}) error
	ReadAll(id StoreIdField, results interface{}) error
}

type StoreIdField struct {
	Name  string
	Value string
}

type MongoStore struct {
	Session        *mgo.Session
	CollectionName string
	Config         MongoConfig
}

type MongoConfig struct {
	Url     string `json:"connectionUrl"`
	DbName  string `json:"databaseName"`
	Timeout int    `json:"timeout"`
}

func NewMongoStore(name string) *MongoStore {

	store := &MongoStore{CollectionName: name}
	config.FromJson(&store.Config, "mongo.json")

	var err error
	store.Session, err = mgo.DialWithTimeout(store.Config.Url, time.Duration(store.Config.Timeout)*time.Second)

	if err != nil {
		log.Logging.Fatal(err)
	}

	return store
}

func (this *MongoStore) Cleanup() {
	cpy := this.Session.Copy()
	defer cpy.Close()

	cpy.DB(this.Config.DbName).C(this.CollectionName).DropCollection()
}

func (this *MongoStore) Save(d interface{}) error {
	cpy := this.Session.Copy()
	defer cpy.Close()

	if err := cpy.DB(this.Config.DbName).C(this.CollectionName).Insert(d); err != nil {
		return err
	}
	return nil
}

func (this *MongoStore) Update(id StoreIdField, d interface{}) error {
	cpy := this.Session.Copy()
	defer cpy.Close()

	if _, err := cpy.DB(this.Config.DbName).C(this.CollectionName).Upsert(bson.M{id.Name: id.Value}, d); err != nil {
		return err
	}
	return nil
}

func (this *MongoStore) Delete(id StoreIdField) error {

	cpy := this.Session.Copy()
	defer cpy.Close()

	if err := cpy.DB(this.Config.DbName).C(this.CollectionName).Remove(bson.M{id.Name: id.Value}); err != nil {
		return err
	}
	return nil
}

func (this *MongoStore) Read(id StoreIdField, result interface{}) error {
	cpy := this.Session.Copy()
	defer cpy.Close()

	log.Logging.Info("read", id)

	if err := cpy.DB(this.Config.DbName).C(this.CollectionName).Find(bson.M{id.Name: id.Value}).One(result); err != nil {
		return err
	}
	return nil
}

func (this *MongoStore) ReadAll(id StoreIdField, results interface{}) error {
	cpy := this.Session.Copy()
	defer cpy.Close()

	log.Logging.Info("read all", id)

	if err := cpy.DB(this.Config.DbName).C(this.CollectionName).Find(bson.M{id.Name: id.Value}).All(results); err != nil {
		return err
	}
	return nil
}
