package store

import (
	"time"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/labix.org/v2/mgo"
	"github.com/tidepool-org/platform/Godeps/_workspace/src/labix.org/v2/mgo/bson"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/logger"
)

var (
	log = logger.Log.GetNamed("store")
)

//Store is a generic store interface for all operations that we need
// Note: there will be a specific implementation for the store
// type we are using, which could also be a mock
type Store interface {
	Save(d interface{}) error
	Update(id IDField, d interface{}) error
	Delete(id IDField) error
	Read(id IDField, result interface{}) error
	ReadAll(id IDField, results interface{}) error
}

//IDField used for finding an item based on its specific ID
type IDField struct {
	Name  string
	Value string
}

//MongoStore is the mongo implementation of Store
type MongoStore struct {
	Session        *mgo.Session
	CollectionName string
	Config         MongoConfig
}

//MongoConfig is the required config for the MongoStore
type MongoConfig struct {
	URL     string `json:"connectionUrl"`
	DbName  string `json:"databaseName"`
	Timeout int    `json:"timeout"`
}

//NewMongoStore returns an initailised instance of MongoStore
func NewMongoStore(name string) *MongoStore {

	store := &MongoStore{CollectionName: name}
	config.FromJSON(&store.Config, "mongo.json")

	var err error
	store.Session, err = mgo.DialWithTimeout(store.Config.URL, time.Duration(store.Config.Timeout)*time.Second)

	if err != nil {
		log.Fatal(err)
	}

	return store
}

//Cleanup will cleanup the collection, used for testing purposes
func (mongoStore *MongoStore) Cleanup() {
	cpy := mongoStore.Session.Copy()
	defer cpy.Close()

	cpy.DB(mongoStore.Config.DbName).C(mongoStore.CollectionName).DropCollection()
}

//Save will save the specified data
func (mongoStore *MongoStore) Save(d interface{}) error {
	cpy := mongoStore.Session.Copy()
	defer cpy.Close()

	if err := cpy.DB(mongoStore.Config.DbName).C(mongoStore.CollectionName).Insert(d); err != nil {
		return err
	}
	return nil
}

//Update will update the specified data based on its IDField
func (mongoStore *MongoStore) Update(id IDField, d interface{}) error {
	cpy := mongoStore.Session.Copy()
	defer cpy.Close()

	if _, err := cpy.DB(mongoStore.Config.DbName).C(mongoStore.CollectionName).Upsert(bson.M{id.Name: id.Value}, d); err != nil {
		return err
	}
	return nil
}

//Delete will delete the specified data based on its IDField
func (mongoStore *MongoStore) Delete(id IDField) error {
	cpy := mongoStore.Session.Copy()
	defer cpy.Close()

	if err := cpy.DB(mongoStore.Config.DbName).C(mongoStore.CollectionName).Remove(bson.M{id.Name: id.Value}); err != nil {
		return err
	}
	return nil
}

//Read will get the specified data based on its IDField
func (mongoStore *MongoStore) Read(id IDField, result interface{}) error {
	cpy := mongoStore.Session.Copy()
	defer cpy.Close()

	if err := cpy.DB(mongoStore.Config.DbName).C(mongoStore.CollectionName).Find(bson.M{id.Name: id.Value}).One(result); err != nil {
		return err
	}
	return nil
}

//ReadAll all data that matches the specified IDField
func (mongoStore *MongoStore) ReadAll(id IDField, results interface{}) error {
	cpy := mongoStore.Session.Copy()
	defer cpy.Close()

	if err := cpy.DB(mongoStore.Config.DbName).C(mongoStore.CollectionName).Find(bson.M{id.Name: id.Value}).All(results); err != nil {
		return err
	}
	return nil
}
