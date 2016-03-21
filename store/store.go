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
	Update(selector interface{}, d interface{}) error
	Delete(find Fields) error
	Read(find Fields, remove Fields, result interface{}) error
	ReadAll(find Fields, remove Fields) Iterator
}

//Iterator for the query iterator
type Iterator interface {
	Next(result interface{}) bool
	Close() error
}

//ClosingSessionIterator so we can manage the session for our iterator
type ClosingSessionIterator struct {
	*mgo.Session
	*mgo.Iter
}

//Fields for finding or updating an item
type Fields map[string]interface{}

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

func mapFields(fields Fields) bson.M {
	mapped := bson.M{}
	for key, val := range fields {
		mapped[key] = val
	}
	return mapped
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
func (mongoStore *MongoStore) Update(selector interface{}, d interface{}) error {
	cpy := mongoStore.Session.Copy()
	defer cpy.Close()

	if _, err := cpy.DB(mongoStore.Config.DbName).
		C(mongoStore.CollectionName).
		Upsert(selector, d); err != nil {
		return err
	}
	return nil
}

//Delete will delete the specified data based on its Fields
// find `IDFields` are the feilds used to find the data to be deleted
func (mongoStore *MongoStore) Delete(find Fields) error {
	cpy := mongoStore.Session.Copy()
	defer cpy.Close()

	if err := cpy.DB(mongoStore.Config.DbName).
		C(mongoStore.CollectionName).
		Remove(mapFields(find)); err != nil {
		return err
	}
	return nil
}

//Read will get the specified data based on its find Fields
// find `Fields` are the feilds used to find the data
// remove `Fields` are the feilds used to remove feilds from being returned.
//   e.g. _version:0 means that the `_version` feild will not be returned in the results
// results is the interface that the result set will be saved into
func (mongoStore *MongoStore) Read(find Fields, remove Fields, result interface{}) error {
	cpy := mongoStore.Session.Copy()
	defer cpy.Close()

	if err := cpy.DB(mongoStore.Config.DbName).
		C(mongoStore.CollectionName).
		Find(mapFields(find)).
		Select(mapFields(remove)).
		One(result); err != nil {
		return err
	}
	return nil
}

//ReadAll all data that matches the specified find Fields
// find `Fields` are the feilds used to find the data
// remove `Fields` are the feilds used to remove fields from being returned.
//   e.g. _version:0 means that the `_version` field will not be returned in the results
func (mongoStore *MongoStore) ReadAll(find Fields, remove Fields) Iterator {
	cpy := mongoStore.Session.Copy()
	defer cpy.Close()

	iter := cpy.DB(mongoStore.Config.DbName).
		C(mongoStore.CollectionName).
		Find(mapFields(find)).
		Select(mapFields(remove)).
		Iter()

	return &ClosingSessionIterator{Session: cpy, Iter: iter}
}

func (i *ClosingSessionIterator) Next(result interface{}) bool {
	if i.Iter != nil {
		return i.Iter.Next(result)
	}
	return false
}

func (i *ClosingSessionIterator) Close() (err error) {
	if i.Iter != nil {
		err = i.Iter.Close()
		i.Iter = nil
	}
	if i.Session != nil {
		i.Session.Close()
		i.Session = nil
	}
	return err
}
