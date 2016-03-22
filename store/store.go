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
	Delete(id Field) error
	Read(id Field, filter Filter, result interface{}) error
	ReadAll(id Field, query Query, filter Filter) Iterator
}

const (
	GreaterThanEquals string = "$gte"
	LessThanEquals    string = "$lte"
	In                string = "$in"
)

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

//Filter for removing unwanted fields from the return data
type Filter []string

//Query for querying of the dataset
// e.g. ["time"] {"$gte":"2010-01-01", "$lte":"2016-01-01"}
type Query map[string]map[string]interface{}

//Field used when accessing data
type Field struct {
	Name  string
	Value interface{}
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

func buildFilter(fields Filter) bson.M {
	mapped := bson.M{}
	for _, val := range fields {
		mapped[val] = 0
	}
	return mapped
}

func buildQuery(id Field, query Query) bson.M {

	builtQuery := bson.M{
		id.Name: id.Value,
		//TODO: specify scheme version
		//"_active":        true,
		//"_schemaVersion": bson.M{GreaterThanEquals: 0, LessThanEquals: 10},
	}

	//Example so its not too abstract
	//["type"] {"$in": ["basal","bolus"]}
	//["time"] {"$gte":"2010", "$lte":"2016"}

	for fieldName, opParams := range query {
		fieldQuery := bson.M{}

		for op, vals := range opParams {
			fieldQuery[op] = vals
		}

		builtQuery[fieldName] = fieldQuery
	}
	return builtQuery
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
// id - `Field` name and value that represents the id for the data
//		e.g. {"userid":"123"}
func (mongoStore *MongoStore) Delete(id Field) error {
	cpy := mongoStore.Session.Copy()
	defer cpy.Close()

	if err := cpy.DB(mongoStore.Config.DbName).
		C(mongoStore.CollectionName).
		Remove(bson.M{id.Name: id.Value}); err != nil {
		return err
	}
	return nil
}

//Read will get the specified data based on its find Fields
// id - `Field` name and value that represents the id for the data
//		e.g. {"userid":"123"}
// filter - field names that will be filtered out of the returned data
// 		e.g. `_version` means that the `_version` field will not be returned in the results
// results is the interface that the result set will be saved into
func (mongoStore *MongoStore) Read(id Field, filter Filter, result interface{}) error {
	cpy := mongoStore.Session.Copy()
	defer cpy.Close()

	if err := cpy.DB(mongoStore.Config.DbName).
		C(mongoStore.CollectionName).
		Find(buildQuery(id, Query{})).
		Select(buildFilter(filter)).
		One(result); err != nil {
		return err
	}
	return nil
}

//ReadAll all data that matches the specified find Fields
// id - `Field` name and value that represents the id for the data
//		e.g. {"userid":"123"}
// query -  `Query` is the query data the will be used in getting the data
//		e.g.  ["type"] {"$in": ["basal","bolus"]} would return all datum of the specified type
// filter - field names that will be filtered out of the returned data
// 		e.g. `_version` means that the `_version` field will not be returned in the results
func (mongoStore *MongoStore) ReadAll(id Field, query Query, filter Filter) Iterator {
	cpy := mongoStore.Session.Copy()
	defer cpy.Close()

	iter := cpy.DB(mongoStore.Config.DbName).
		C(mongoStore.CollectionName).
		Find(buildQuery(id, query)).
		Select(buildFilter(filter)).
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
