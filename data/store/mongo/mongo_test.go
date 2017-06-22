package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"sync"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/log"
	baseMongo "github.com/tidepool-org/platform/store/mongo"
	testMongo "github.com/tidepool-org/platform/test/mongo"
)

var _sampleTimeMutex sync.Mutex
var _sampleTimeOnce sync.Once
var _sampleTime time.Time

func SampleTime() time.Time {
	_sampleTimeMutex.Lock()
	defer _sampleTimeMutex.Unlock()

	_sampleTimeOnce.Do(func() {
		_sampleTime, _ = time.Parse(time.RFC3339, "2016-08-30T23:59:50-07:00")
	})

	_sampleTime = _sampleTime.Add(time.Second)
	return _sampleTime
}

type TestAgent struct {
	TestIsServer bool
	TestUserID   string
}

func (t *TestAgent) IsServer() bool {
	return t.TestIsServer
}

func (t *TestAgent) UserID() string {
	return t.TestUserID
}

func NewDataset(userID string, groupID string, deviceID string) *upload.Upload {
	dataset := upload.Init()
	Expect(dataset).ToNot(BeNil())

	dataset.Deduplicator = &data.DeduplicatorDescriptor{Name: "test-deduplicator"}
	dataset.GroupID = groupID
	dataset.UserID = userID

	dataset.ClockDriftOffset = app.IntegerAsPointer(0)
	dataset.ConversionOffset = app.IntegerAsPointer(0)
	dataset.DeviceID = app.StringAsPointer(deviceID)
	dataset.DeviceTime = app.StringAsPointer(SampleTime().Format("2006-01-02T15:04:05"))
	dataset.Time = app.StringAsPointer(SampleTime().UTC().Format("2006-01-02T15:04:05Z07:00"))
	dataset.TimezoneOffset = app.IntegerAsPointer(-420)

	dataset.ComputerTime = app.StringAsPointer(SampleTime().Format("2006-01-02T15:04:05"))
	dataset.DeviceManufacturers = app.StringArrayAsPointer([]string{"Tesla"})
	dataset.DeviceModel = app.StringAsPointer("1234")
	dataset.DeviceSerialNumber = app.StringAsPointer("567890")
	dataset.DeviceTags = app.StringArrayAsPointer([]string{"insulin-pump"})
	dataset.TimeProcessing = app.StringAsPointer("utc-bootstrapping")
	dataset.TimeZone = app.StringAsPointer("US/Pacific")
	dataset.Version = app.StringAsPointer("0.260.1")

	return dataset
}

func NewDatasetData(deviceID string) []data.Datum {
	datasetData := []data.Datum{}
	for count := 0; count < 3; count++ {
		baseDatum := &types.Base{}
		baseDatum.Init()

		baseDatum.Deduplicator = &data.DeduplicatorDescriptor{Hash: app.NewID()}
		baseDatum.Type = "test"

		baseDatum.ClockDriftOffset = app.IntegerAsPointer(0)
		baseDatum.ConversionOffset = app.IntegerAsPointer(0)
		baseDatum.DeviceID = app.StringAsPointer(deviceID)
		baseDatum.DeviceTime = app.StringAsPointer(SampleTime().Format("2006-01-02T15:04:05"))
		baseDatum.Time = app.StringAsPointer(SampleTime().UTC().Format("2006-01-02T15:04:05Z07:00"))
		baseDatum.TimezoneOffset = app.IntegerAsPointer(-420)

		datasetData = append(datasetData, baseDatum)
	}
	return datasetData
}

func ValidateDataset(testMongoCollection *mgo.Collection, selector bson.M, expectedDatasets ...*upload.Upload) {
	selector["type"] = "upload"
	var actualDatasets []*upload.Upload
	Expect(testMongoCollection.Find(selector).All(&actualDatasets)).To(Succeed())
	Expect(actualDatasets).To(ConsistOf(expectedDatasets))
}

// TODO: Actually compare dataset data here, not just count. Needs code to read actual data types from Mongo.

func ValidateDatasetData(testMongoCollection *mgo.Collection, selector bson.M, expectedDatasetData []data.Datum) {
	var actualDatasetData []interface{}
	Expect(testMongoCollection.Find(selector).All(&actualDatasetData)).To(Succeed())
	Expect(actualDatasetData).To(HaveLen(len(expectedDatasetData)))
}

var _ = Describe("Mongo", func() {
	var mongoConfig *baseMongo.Config
	var mongoStore *mongo.Store
	var mongoSession store.Session

	BeforeEach(func() {
		mongoConfig = &baseMongo.Config{
			Addresses:  testMongo.Address(),
			Database:   testMongo.Database(),
			Collection: testMongo.NewCollectionName(),
			Timeout:    app.DurationAsPointer(5 * time.Second),
		}
	})

	AfterEach(func() {
		if mongoSession != nil {
			mongoSession.Close()
		}
		if mongoStore != nil {
			mongoStore.Close()
		}
	})

	Context("New", func() {
		It("returns an error if unsuccessful", func() {
			var err error
			mongoStore, err = mongo.New(nil, nil)
			Expect(err).To(HaveOccurred())
			Expect(mongoStore).To(BeNil())
		})

		It("returns a new store and no error if successful", func() {
			var err error
			mongoStore, err = mongo.New(log.NewNull(), mongoConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			mongoStore, err = mongo.New(log.NewNull(), mongoConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})

		Context("NewSession", func() {
			It("returns an error if unsuccessful", func() {
				var err error
				mongoSession, err = mongoStore.NewSession(nil)
				Expect(err).To(HaveOccurred())
				Expect(mongoSession).To(BeNil())
			})

			It("returns a new session and no error if successful", func() {
				var err error
				mongoSession, err = mongoStore.NewSession(log.NewNull())
				Expect(err).ToNot(HaveOccurred())
				Expect(mongoSession).ToNot(BeNil())
			})
		})

		Context("with a new session", func() {
			BeforeEach(func() {
				var err error
				mongoSession, err = mongoStore.NewSession(log.NewNull())
				Expect(err).ToNot(HaveOccurred())
				Expect(mongoSession).ToNot(BeNil())
			})

			Context("with persisted data", func() {
				var testMongoSession *mgo.Session
				var testMongoCollection *mgo.Collection
				var userID string
				var groupID string
				var deviceID string
				var datasetExistingOther *upload.Upload
				var datasetExistingOne *upload.Upload
				var datasetExistingTwo *upload.Upload
				var dataset *upload.Upload

				BeforeEach(func() {
					testMongoSession = testMongo.Session().Copy()
					testMongoCollection = testMongoSession.DB(mongoConfig.Database).C(mongoConfig.Collection)
					userID = app.NewID()
					groupID = app.NewID()
					deviceID = app.NewID()
					datasetExistingOther = NewDataset(app.NewID(), app.NewID(), app.NewID())
					datasetExistingOther.CreatedTime = "2016-09-01T12:00:00Z"
					Expect(testMongoCollection.Insert(datasetExistingOther)).To(Succeed())
					datasetExistingOne = NewDataset(userID, groupID, deviceID)
					datasetExistingOne.CreatedTime = "2016-09-01T12:30:00Z"
					Expect(testMongoCollection.Insert(datasetExistingOne)).To(Succeed())
					datasetExistingTwo = NewDataset(userID, groupID, deviceID)
					datasetExistingTwo.CreatedTime = "2016-09-01T10:00:00Z"
					Expect(testMongoCollection.Insert(datasetExistingTwo)).To(Succeed())
					dataset = NewDataset(userID, groupID, deviceID)
				})

				AfterEach(func() {
					if testMongoSession != nil {
						testMongoSession.Close()
					}
				})

				Context("GetDatasetsForUserByID", func() {
					var filter *store.Filter
					var pagination *store.Pagination

					BeforeEach(func() {
						dataset.CreatedTime = "2016-09-01T11:00:00Z"
						Expect(testMongoCollection.Insert(dataset)).To(Succeed())
						filter = store.NewFilter()
						pagination = store.NewPagination()
					})

					It("succeeds if it successfully finds the user datasets", func() {
						Expect(mongoSession.GetDatasetsForUserByID(userID, filter, pagination)).To(ConsistOf([]*upload.Upload{datasetExistingOne, dataset, datasetExistingTwo}))
					})

					It("succeeds if the filter is not specified", func() {
						Expect(mongoSession.GetDatasetsForUserByID(userID, nil, pagination)).To(ConsistOf([]*upload.Upload{datasetExistingOne, dataset, datasetExistingTwo}))
					})

					It("succeeds if the pagination is not specified", func() {
						Expect(mongoSession.GetDatasetsForUserByID(userID, filter, nil)).To(ConsistOf([]*upload.Upload{datasetExistingOne, dataset, datasetExistingTwo}))
					})

					It("succeeds if the pagination size is not default", func() {
						pagination.Size = 2
						Expect(mongoSession.GetDatasetsForUserByID(userID, filter, pagination)).To(ConsistOf([]*upload.Upload{datasetExistingOne, dataset}))
					})

					It("succeeds if the pagination page and size is not default", func() {
						pagination.Page = 1
						pagination.Size = 2
						Expect(mongoSession.GetDatasetsForUserByID(userID, filter, pagination)).To(ConsistOf([]*upload.Upload{datasetExistingTwo}))
					})

					It("succeeds if it successfully does not find another user datasets", func() {
						resultDatasets, err := mongoSession.GetDatasetsForUserByID(app.NewID(), filter, pagination)
						Expect(err).ToNot(HaveOccurred())
						Expect(resultDatasets).ToNot(BeNil())
						Expect(resultDatasets).To(BeEmpty())
					})

					It("returns an error if the user id is missing", func() {
						resultDatasets, err := mongoSession.GetDatasetsForUserByID("", filter, pagination)
						Expect(err).To(MatchError("mongo: user id is missing"))
						Expect(resultDatasets).To(BeNil())
					})

					It("returns an error if the pagination page is less than minimum", func() {
						pagination.Page = -1
						resultDatasets, err := mongoSession.GetDatasetsForUserByID(userID, filter, pagination)
						Expect(err).To(MatchError("mongo: pagination is invalid; store: page is invalid"))
						Expect(resultDatasets).To(BeNil())
					})

					It("returns an error if the pagination size is less than minimum", func() {
						pagination.Size = 0
						resultDatasets, err := mongoSession.GetDatasetsForUserByID(userID, filter, pagination)
						Expect(err).To(MatchError("mongo: pagination is invalid; store: size is invalid"))
						Expect(resultDatasets).To(BeNil())
					})

					It("returns an error if the pagination size is greater than maximum", func() {
						pagination.Size = 101
						resultDatasets, err := mongoSession.GetDatasetsForUserByID(userID, filter, pagination)
						Expect(err).To(MatchError("mongo: pagination is invalid; store: size is invalid"))
						Expect(resultDatasets).To(BeNil())
					})

					It("returns an error if the session is closed", func() {
						mongoSession.Close()
						resultDatasets, err := mongoSession.GetDatasetsForUserByID(userID, filter, pagination)
						Expect(err).To(MatchError("mongo: session closed"))
						Expect(resultDatasets).To(BeNil())
					})

					Context("with deleted dataset", func() {
						BeforeEach(func() {
							dataset.DeletedTime = "2016-09-01T13:00:00Z"
							Expect(testMongoCollection.Update(bson.M{"id": dataset.ID}, dataset)).To(Succeed())
						})

						It("succeeds if it successfully finds the non-deleted user datasets", func() {
							Expect(mongoSession.GetDatasetsForUserByID(userID, filter, pagination)).To(ConsistOf([]*upload.Upload{datasetExistingOne, datasetExistingTwo}))
						})

						It("succeeds if it successfully finds all the user datasets", func() {
							filter.Deleted = true
							Expect(mongoSession.GetDatasetsForUserByID(userID, filter, pagination)).To(ConsistOf([]*upload.Upload{datasetExistingOne, dataset, datasetExistingTwo}))
						})
					})
				})

				Context("GetDatasetByID", func() {
					BeforeEach(func() {
						dataset.CreatedTime = "2016-09-01T11:00:00Z"
						Expect(testMongoCollection.Insert(dataset)).To(Succeed())
					})

					It("succeeds if it successfully finds the dataset", func() {
						Expect(mongoSession.GetDatasetByID(dataset.UploadID)).To(Equal(dataset))
					})

					It("returns an error if the dataset id is missing", func() {
						resultDataset, err := mongoSession.GetDatasetByID("")
						Expect(err).To(MatchError("mongo: dataset id is missing"))
						Expect(resultDataset).To(BeNil())
					})

					It("returns an error if the session is closed", func() {
						mongoSession.Close()
						resultDataset, err := mongoSession.GetDatasetByID(dataset.UploadID)
						Expect(err).To(MatchError("mongo: session closed"))
						Expect(resultDataset).To(BeNil())
					})

					It("returns no dataset successfully if the dataset cannot be found", func() {
						resultDataset, err := mongoSession.GetDatasetByID("not-found")
						Expect(err).ToNot(HaveOccurred())
						Expect(resultDataset).To(BeNil())
					})
				})

				Context("FindPreviousActiveDatasetForDevice", func() {
					var datasetExistingThree *upload.Upload

					BeforeEach(func() {
						datasetExistingThree = NewDataset(userID, groupID, deviceID)
						datasetExistingThree.CreatedTime = "2016-09-01T11:30:00Z"
						Expect(testMongoCollection.Insert(datasetExistingThree)).To(Succeed())
						dataset.CreatedTime = "2016-09-01T12:15:00Z"
						Expect(testMongoCollection.Insert(dataset)).To(Succeed())
					})

					It("succeeds if it successfully finds the previous dataset", func() {
						previousDataset, err := mongoSession.FindPreviousActiveDatasetForDevice(dataset)
						Expect(err).ToNot(HaveOccurred())
						Expect(previousDataset).To(Equal(datasetExistingThree))
					})

					It("returns an error if the dataset is missing", func() {
						previousDataset, err := mongoSession.FindPreviousActiveDatasetForDevice(nil)
						Expect(err).To(MatchError("mongo: dataset is missing"))
						Expect(previousDataset).To(BeNil())
					})

					It("returns an error if the dataset user id is missing", func() {
						dataset.UserID = ""
						previousDataset, err := mongoSession.FindPreviousActiveDatasetForDevice(dataset)
						Expect(err).To(MatchError("mongo: dataset user id is missing"))
						Expect(previousDataset).To(BeNil())
					})

					It("returns an error if the dataset group id is missing", func() {
						dataset.GroupID = ""
						previousDataset, err := mongoSession.FindPreviousActiveDatasetForDevice(dataset)
						Expect(err).To(MatchError("mongo: dataset group id is missing"))
						Expect(previousDataset).To(BeNil())
					})

					It("returns an error if the dataset device id is missing", func() {
						dataset.DeviceID = nil
						previousDataset, err := mongoSession.FindPreviousActiveDatasetForDevice(dataset)
						Expect(err).To(MatchError("mongo: dataset device id is missing"))
						Expect(previousDataset).To(BeNil())
					})

					It("returns an error if the dataset device id is empty", func() {
						dataset.DeviceID = app.StringAsPointer("")
						previousDataset, err := mongoSession.FindPreviousActiveDatasetForDevice(dataset)
						Expect(err).To(MatchError("mongo: dataset device id is missing"))
						Expect(previousDataset).To(BeNil())
					})

					It("returns an error if the dataset deduplicator descriptor is missing", func() {
						dataset.Deduplicator = nil
						previousDataset, err := mongoSession.FindPreviousActiveDatasetForDevice(dataset)
						Expect(err).To(MatchError("mongo: dataset deduplicator name is missing"))
						Expect(previousDataset).To(BeNil())
					})

					It("returns an error if the dataset deduplicator descriptor name is empty", func() {
						dataset.Deduplicator.Name = ""
						previousDataset, err := mongoSession.FindPreviousActiveDatasetForDevice(dataset)
						Expect(err).To(MatchError("mongo: dataset deduplicator name is missing"))
						Expect(previousDataset).To(BeNil())
					})

					It("returns an error if the dataset created time is missing", func() {
						dataset.CreatedTime = ""
						previousDataset, err := mongoSession.FindPreviousActiveDatasetForDevice(dataset)
						Expect(err).To(MatchError("mongo: dataset created time is missing"))
						Expect(previousDataset).To(BeNil())
					})

					It("returns an error if the session is closed", func() {
						mongoSession.Close()
						previousDataset, err := mongoSession.FindPreviousActiveDatasetForDevice(dataset)
						Expect(err).To(MatchError("mongo: session closed"))
						Expect(previousDataset).To(BeNil())
					})

					It("returns no previous dataset if different user id", func() {
						dataset.UserID = app.NewID()
						previousDataset, err := mongoSession.FindPreviousActiveDatasetForDevice(dataset)
						Expect(err).ToNot(HaveOccurred())
						Expect(previousDataset).To(BeNil())
					})

					It("returns no previous dataset if different group id", func() {
						dataset.GroupID = app.NewID()
						previousDataset, err := mongoSession.FindPreviousActiveDatasetForDevice(dataset)
						Expect(err).ToNot(HaveOccurred())
						Expect(previousDataset).To(BeNil())
					})

					It("returns no previous dataset if different device id", func() {
						dataset.DeviceID = app.StringAsPointer(app.NewID())
						previousDataset, err := mongoSession.FindPreviousActiveDatasetForDevice(dataset)
						Expect(err).ToNot(HaveOccurred())
						Expect(previousDataset).To(BeNil())
					})

					It("returns no previous dataset if different deduplicator name", func() {
						dataset.Deduplicator.Name = app.NewID()
						previousDataset, err := mongoSession.FindPreviousActiveDatasetForDevice(dataset)
						Expect(err).ToNot(HaveOccurred())
						Expect(previousDataset).To(BeNil())
					})

					It("ignores previous dataset if it does not have created time", func() {
						Expect(testMongoCollection.Update(bson.M{"id": datasetExistingThree.ID}, bson.M{"$unset": bson.M{"createdTime": 1}})).To(Succeed())
						previousDataset, err := mongoSession.FindPreviousActiveDatasetForDevice(dataset)
						Expect(err).ToNot(HaveOccurred())
						Expect(previousDataset).To(Equal(datasetExistingTwo))
					})

					It("ignores previous dataset if created time is empty", func() {
						Expect(testMongoCollection.Update(bson.M{"id": datasetExistingThree.ID}, bson.M{"$set": bson.M{"createdTime": ""}})).To(Succeed())
						previousDataset, err := mongoSession.FindPreviousActiveDatasetForDevice(dataset)
						Expect(err).ToNot(HaveOccurred())
						Expect(previousDataset).To(Equal(datasetExistingTwo))
					})

					It("ignores previous dataset if created time equals created time of dataset", func() {
						Expect(testMongoCollection.Update(bson.M{"id": datasetExistingThree.ID}, bson.M{"$set": bson.M{"createdTime": "2016-09-01T12:15:00Z"}})).To(Succeed())
						previousDataset, err := mongoSession.FindPreviousActiveDatasetForDevice(dataset)
						Expect(err).ToNot(HaveOccurred())
						Expect(previousDataset).To(Equal(datasetExistingTwo))
					})

					It("ignores previous dataset if created time after created time of dataset", func() {
						Expect(testMongoCollection.Update(bson.M{"id": datasetExistingThree.ID}, bson.M{"$set": bson.M{"createdTime": "2016-09-01T12:15:01Z"}})).To(Succeed())
						previousDataset, err := mongoSession.FindPreviousActiveDatasetForDevice(dataset)
						Expect(err).ToNot(HaveOccurred())
						Expect(previousDataset).To(Equal(datasetExistingTwo))
					})
				})

				Context("CreateDataset", func() {
					It("succeeds if it successfully creates the dataset", func() {
						Expect(mongoSession.CreateDataset(dataset)).To(Succeed())
					})

					It("returns an error if the dataset is missing", func() {
						Expect(mongoSession.CreateDataset(nil)).To(MatchError("mongo: dataset is missing"))
					})

					It("returns an error if the user id is missing", func() {
						dataset.UserID = ""
						Expect(mongoSession.CreateDataset(dataset)).To(MatchError("mongo: dataset user id is missing"))
					})

					It("returns an error if the group id is missing", func() {
						dataset.GroupID = ""
						Expect(mongoSession.CreateDataset(dataset)).To(MatchError("mongo: dataset group id is missing"))
					})

					It("returns an error if the upload id is missing", func() {
						dataset.UploadID = ""
						Expect(mongoSession.CreateDataset(dataset)).To(MatchError("mongo: dataset upload id is missing"))
					})

					It("returns an error if the session is closed", func() {
						mongoSession.Close()
						Expect(mongoSession.CreateDataset(dataset)).To(MatchError("mongo: session closed"))
					})

					It("returns an error if the dataset with the same id already exists", func() {
						Expect(mongoSession.CreateDataset(dataset)).To(Succeed())
						Expect(mongoSession.CreateDataset(dataset)).To(MatchError("mongo: unable to create dataset; mongo: dataset already exists"))
					})

					It("sets the created time", func() {
						Expect(mongoSession.CreateDataset(dataset)).To(Succeed())
						Expect(dataset.CreatedTime).ToNot(BeEmpty())
						Expect(dataset.CreatedUserID).To(BeEmpty())
						Expect(dataset.ByUser).To(BeEmpty())
					})

					It("has the correct stored datasets", func() {
						ValidateDataset(testMongoCollection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, datasetExistingOther, datasetExistingOne, datasetExistingTwo)
						Expect(mongoSession.CreateDataset(dataset)).To(Succeed())
						ValidateDataset(testMongoCollection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, datasetExistingOther, datasetExistingOne, datasetExistingTwo, dataset)
					})

					Context("with agent specified", func() {
						var agentUserID string

						BeforeEach(func() {
							agentUserID = app.NewID()
							mongoSession.SetAgent(&TestAgent{false, agentUserID})
						})

						It("sets the created time and created user id", func() {
							Expect(mongoSession.CreateDataset(dataset)).To(Succeed())
							Expect(dataset.CreatedTime).ToNot(BeEmpty())
							Expect(dataset.CreatedUserID).To(Equal(agentUserID))
							Expect(dataset.ByUser).To(Equal(agentUserID))
						})

						It("has the correct stored datasets", func() {
							ValidateDataset(testMongoCollection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, datasetExistingOther, datasetExistingOne, datasetExistingTwo)
							Expect(mongoSession.CreateDataset(dataset)).To(Succeed())
							ValidateDataset(testMongoCollection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, datasetExistingOther, datasetExistingOne, datasetExistingTwo)
							ValidateDataset(testMongoCollection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": agentUserID}, dataset)
						})
					})
				})

				Context("UpdateDataset", func() {
					BeforeEach(func() {
						dataset.CreatedTime = "2016-09-01T11:00:00Z"
						Expect(testMongoCollection.Insert(dataset)).To(Succeed())
					})

					Context("with state closed", func() {
						BeforeEach(func() {
							dataset.State = "closed"
						})

						It("succeeds if it successfully updates the dataset", func() {
							Expect(mongoSession.UpdateDataset(dataset)).To(Succeed())
						})

						It("returns an error if the dataset is missing", func() {
							Expect(mongoSession.UpdateDataset(nil)).To(MatchError("mongo: dataset is missing"))
						})

						It("returns an error if the user id is missing", func() {
							dataset.UserID = ""
							Expect(mongoSession.UpdateDataset(dataset)).To(MatchError("mongo: dataset user id is missing"))
						})

						It("returns an error if the group id is missing", func() {
							dataset.GroupID = ""
							Expect(mongoSession.UpdateDataset(dataset)).To(MatchError("mongo: dataset group id is missing"))
						})

						It("returns an error if the upload id is missing", func() {
							dataset.UploadID = ""
							Expect(mongoSession.UpdateDataset(dataset)).To(MatchError("mongo: dataset upload id is missing"))
						})

						It("returns an error if the session is closed", func() {
							mongoSession.Close()
							Expect(mongoSession.UpdateDataset(dataset)).To(MatchError("mongo: session closed"))
						})

						It("returns an error if the dataset with the same user id, group id, and upload id does not yet exist", func() {
							dataset.UploadID = app.NewID()
							Expect(mongoSession.UpdateDataset(dataset)).To(MatchError("mongo: unable to update dataset; not found"))
						})
					})

					It("sets the modified time", func() {
						dataset.State = "closed"
						Expect(mongoSession.UpdateDataset(dataset)).To(Succeed())
						Expect(dataset.ModifiedTime).ToNot(BeEmpty())
						Expect(dataset.ModifiedUserID).To(BeEmpty())
					})

					It("has the correct stored datasets", func() {
						ValidateDataset(testMongoCollection, bson.M{}, datasetExistingOther, datasetExistingOne, datasetExistingTwo, dataset)
						ValidateDataset(testMongoCollection, bson.M{"modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}})
						dataset.State = "closed"
						Expect(mongoSession.UpdateDataset(dataset)).To(Succeed())
						ValidateDataset(testMongoCollection, bson.M{}, datasetExistingOther, datasetExistingOne, datasetExistingTwo, dataset)
						ValidateDataset(testMongoCollection, bson.M{"modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, dataset)
					})

					Context("with agent specified", func() {
						var agentUserID string

						BeforeEach(func() {
							agentUserID = app.NewID()
							mongoSession.SetAgent(&TestAgent{false, agentUserID})
						})

						It("sets the modified time and modified user id", func() {
							dataset.State = "closed"
							Expect(mongoSession.UpdateDataset(dataset)).To(Succeed())
							Expect(dataset.ModifiedTime).ToNot(BeEmpty())
							Expect(dataset.ModifiedUserID).To(Equal(agentUserID))
						})

						It("has the correct stored datasets", func() {
							ValidateDataset(testMongoCollection, bson.M{}, datasetExistingOther, datasetExistingOne, datasetExistingTwo, dataset)
							ValidateDataset(testMongoCollection, bson.M{"modifiedTime": bson.M{"$exists": true}, "modifiedUserId": agentUserID})
							dataset.State = "closed"
							Expect(mongoSession.UpdateDataset(dataset)).To(Succeed())
							ValidateDataset(testMongoCollection, bson.M{}, datasetExistingOther, datasetExistingOne, datasetExistingTwo, dataset)
							ValidateDataset(testMongoCollection, bson.M{"modifiedTime": bson.M{"$exists": true}, "modifiedUserId": agentUserID}, dataset)
						})
					})
				})

				Context("with data", func() {
					var datasetExistingOtherData []data.Datum
					var datasetExistingOneData []data.Datum
					var datasetExistingTwoData []data.Datum
					var datasetData []data.Datum

					BeforeEach(func() {
						dataset.CreatedTime = "2016-09-01T11:00:00Z"
						Expect(testMongoCollection.Insert(dataset)).To(Succeed())
						datasetExistingOtherData = NewDatasetData(app.NewID())
						Expect(mongoSession.CreateDatasetData(datasetExistingOther, datasetExistingOtherData)).To(Succeed())
						datasetExistingOneData = NewDatasetData(deviceID)
						Expect(mongoSession.CreateDatasetData(datasetExistingOne, datasetExistingOneData)).To(Succeed())
						datasetExistingTwoData = NewDatasetData(deviceID)
						Expect(mongoSession.CreateDatasetData(datasetExistingTwo, datasetExistingTwoData)).To(Succeed())
						datasetData = NewDatasetData(deviceID)
					})

					Context("DeleteDataset", func() {
						BeforeEach(func() {
							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
						})

						It("succeeds if it successfully deletes the dataset", func() {
							Expect(mongoSession.DeleteDataset(dataset)).To(Succeed())
						})

						It("returns an error if the dataset is missing", func() {
							Expect(mongoSession.DeleteDataset(nil)).To(MatchError("mongo: dataset is missing"))
						})

						It("returns an error if the user id is missing", func() {
							dataset.UserID = ""
							Expect(mongoSession.DeleteDataset(dataset)).To(MatchError("mongo: dataset user id is missing"))
						})

						It("returns an error if the group id is missing", func() {
							dataset.GroupID = ""
							Expect(mongoSession.DeleteDataset(dataset)).To(MatchError("mongo: dataset group id is missing"))
						})

						It("returns an error if the upload id is missing", func() {
							dataset.UploadID = ""
							Expect(mongoSession.DeleteDataset(dataset)).To(MatchError("mongo: dataset upload id is missing"))
						})

						It("returns an error if the session is closed", func() {
							mongoSession.Close()
							Expect(mongoSession.DeleteDataset(dataset)).To(MatchError("mongo: session closed"))
						})

						It("sets the deleted time on the dataset", func() {
							Expect(mongoSession.DeleteDataset(dataset)).To(Succeed())
							Expect(dataset.DeletedTime).ToNot(BeEmpty())
							Expect(dataset.DeletedUserID).To(BeEmpty())
						})

						It("has the correct stored dataset", func() {
							ValidateDataset(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": true}, "deletedUserId": bson.M{"$exists": false}})
							Expect(mongoSession.DeleteDataset(dataset)).To(Succeed())
							ValidateDataset(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": true}, "deletedUserId": bson.M{"$exists": false}}, dataset)
						})

						It("has the correct stored dataset data", func() {
							ValidateDatasetData(testMongoCollection, bson.M{"uploadId": dataset.UploadID}, append(datasetData, dataset))
							Expect(mongoSession.DeleteDataset(dataset)).To(Succeed())
							ValidateDatasetData(testMongoCollection, bson.M{"uploadId": dataset.UploadID}, []data.Datum{dataset})
						})

						Context("with agent specified", func() {
							var agentUserID string

							BeforeEach(func() {
								agentUserID = app.NewID()
								mongoSession.SetAgent(&TestAgent{false, agentUserID})
							})

							It("sets the deleted time and deleted user id on the dataset", func() {
								Expect(mongoSession.DeleteDataset(dataset)).To(Succeed())
								Expect(dataset.DeletedTime).ToNot(BeEmpty())
								Expect(dataset.DeletedUserID).To(Equal(agentUserID))
							})

							It("has the correct stored dataset", func() {
								ValidateDataset(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": true}, "deletedUserId": agentUserID})
								Expect(mongoSession.DeleteDataset(dataset)).To(Succeed())
								ValidateDataset(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": true}, "deletedUserId": agentUserID}, dataset)
							})

							It("has the correct stored dataset data", func() {
								ValidateDatasetData(testMongoCollection, bson.M{"uploadId": dataset.UploadID}, append(datasetData, dataset))
								Expect(mongoSession.DeleteDataset(dataset)).To(Succeed())
								ValidateDatasetData(testMongoCollection, bson.M{"uploadId": dataset.UploadID}, []data.Datum{dataset})
							})
						})
					})

					Context("with deduplicator hashes", func() {
						var dataHashes []string

						BeforeEach(func() {
							dataHashes = []string{}
							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
							for _, dataDatum := range datasetData {
								dataHashes = append(dataHashes, dataDatum.DeduplicatorDescriptor().Hash)
							}
						})

						Context("GetDatasetDataDeduplicatorHashes", func() {
							It("succeeds if it successfully gets the data hashes", func() {
								foundHashes, err := mongoSession.GetDatasetDataDeduplicatorHashes(dataset, false)
								Expect(err).ToNot(HaveOccurred())
								Expect(foundHashes).To(ConsistOf(dataHashes))
							})

							It("returns an error if the dataset is missing", func() {
								foundHashes, err := mongoSession.GetDatasetDataDeduplicatorHashes(nil, false)
								Expect(err).To(MatchError("mongo: dataset is missing"))
								Expect(foundHashes).To(BeNil())
							})

							It("returns an error if the dataset user id is missing", func() {
								dataset.UserID = ""
								foundHashes, err := mongoSession.GetDatasetDataDeduplicatorHashes(dataset, false)
								Expect(err).To(MatchError("mongo: dataset user id is missing"))
								Expect(foundHashes).To(BeNil())
							})

							It("returns an error if the dataset group id is missing", func() {
								dataset.GroupID = ""
								foundHashes, err := mongoSession.GetDatasetDataDeduplicatorHashes(dataset, false)
								Expect(err).To(MatchError("mongo: dataset group id is missing"))
								Expect(foundHashes).To(BeNil())
							})

							It("returns an error if the dataset upload id is missing", func() {
								dataset.UploadID = ""
								foundHashes, err := mongoSession.GetDatasetDataDeduplicatorHashes(dataset, false)
								Expect(err).To(MatchError("mongo: dataset upload id is missing"))
								Expect(foundHashes).To(BeNil())
							})

							It("returns an error if the session is closed", func() {
								mongoSession.Close()
								foundHashes, err := mongoSession.GetDatasetDataDeduplicatorHashes(dataset, false)
								Expect(err).To(MatchError("mongo: session closed"))
								Expect(foundHashes).To(BeNil())
							})

							It("returns no hashes if different user id", func() {
								dataset.UserID = app.NewID()
								foundHashes, err := mongoSession.GetDatasetDataDeduplicatorHashes(dataset, false)
								Expect(err).ToNot(HaveOccurred())
								Expect(foundHashes).To(BeNil())
							})

							It("returns no hashes if different group id", func() {
								dataset.GroupID = app.NewID()
								foundHashes, err := mongoSession.GetDatasetDataDeduplicatorHashes(dataset, false)
								Expect(err).ToNot(HaveOccurred())
								Expect(foundHashes).To(BeNil())
							})

							It("returns no hashes if different upload id", func() {
								dataset.UploadID = app.NewID()
								foundHashes, err := mongoSession.GetDatasetDataDeduplicatorHashes(dataset, false)
								Expect(err).ToNot(HaveOccurred())
								Expect(foundHashes).To(BeNil())
							})

							It("returns no hashes if active", func() {
								foundHashes, err := mongoSession.GetDatasetDataDeduplicatorHashes(dataset, true)
								Expect(err).ToNot(HaveOccurred())
								Expect(foundHashes).To(BeNil())
							})
						})

						Context("FindAllDatasetDataDeduplicatorHashesForDevice", func() {
							var queryHashes []string

							BeforeEach(func() {
								queryHashes = append(dataHashes, dataHashes...)
								queryHashes = append(queryHashes, app.NewID(), app.NewID())
							})

							It("succeeds if it successfully finds the data hashes with the specified query hashes", func() {
								foundHashes, err := mongoSession.FindAllDatasetDataDeduplicatorHashesForDevice(userID, deviceID, queryHashes)
								Expect(err).ToNot(HaveOccurred())
								Expect(foundHashes).To(ConsistOf(dataHashes))
							})

							It("returns an error if the user id is missing", func() {
								foundHashes, err := mongoSession.FindAllDatasetDataDeduplicatorHashesForDevice("", deviceID, queryHashes)
								Expect(err).To(MatchError("mongo: user id is missing"))
								Expect(foundHashes).To(BeEmpty())
							})

							It("returns an error if the device id is missing", func() {
								foundHashes, err := mongoSession.FindAllDatasetDataDeduplicatorHashesForDevice(userID, "", queryHashes)
								Expect(err).To(MatchError("mongo: device id is missing"))
								Expect(foundHashes).To(BeEmpty())
							})

							It("returns an error if the session is closed", func() {
								mongoSession.Close()
								foundHashes, err := mongoSession.FindAllDatasetDataDeduplicatorHashesForDevice(userID, deviceID, queryHashes)
								Expect(err).To(MatchError("mongo: session closed"))
								Expect(foundHashes).To(BeEmpty())
							})

							It("returns no data hashes if the query hashes is empty", func() {
								foundHashes, err := mongoSession.FindAllDatasetDataDeduplicatorHashesForDevice(userID, deviceID, []string{})
								Expect(err).ToNot(HaveOccurred())
								Expect(foundHashes).To(BeEmpty())
							})
						})
					})

					Context("CreateDatasetData", func() {
						It("succeeds if it successfully creates the dataset data", func() {
							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
						})

						It("returns an error if the dataset is missing", func() {
							Expect(mongoSession.CreateDatasetData(nil, datasetData)).To(MatchError("mongo: dataset is missing"))
						})

						It("returns an error if the dataset data is missing", func() {
							Expect(mongoSession.CreateDatasetData(dataset, nil)).To(MatchError("mongo: dataset data is missing"))
						})

						It("returns an error if the user id is missing", func() {
							dataset.UserID = ""
							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(MatchError("mongo: dataset user id is missing"))
						})

						It("returns an error if the group id is missing", func() {
							dataset.GroupID = ""
							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(MatchError("mongo: dataset group id is missing"))
						})

						It("returns an error if the upload id is missing", func() {
							dataset.UploadID = ""
							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(MatchError("mongo: dataset upload id is missing"))
						})

						It("returns an error if the session is closed", func() {
							mongoSession.Close()
							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(MatchError("mongo: session closed"))
						})

						It("sets the user id, group id, and upload id on the dataset data to match the dataset", func() {
							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
							for _, datasetDatum := range datasetData {
								baseDatum, ok := datasetDatum.(*types.Base)
								Expect(ok).To(BeTrue())
								Expect(baseDatum).ToNot(BeNil())
								Expect(baseDatum.UserID).To(Equal(dataset.UserID))
								Expect(baseDatum.GroupID).To(Equal(dataset.GroupID))
								Expect(baseDatum.UploadID).To(Equal(dataset.UploadID))
							}
						})

						It("leaves the dataset data not active", func() {
							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
							for _, datasetDatum := range datasetData {
								baseDatum, ok := datasetDatum.(*types.Base)
								Expect(ok).To(BeTrue())
								Expect(baseDatum).ToNot(BeNil())
								Expect(baseDatum.Active).To(BeFalse())
							}
						})

						It("sets the created time on the dataset data", func() {
							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
							for _, datasetDatum := range datasetData {
								baseDatum, ok := datasetDatum.(*types.Base)
								Expect(ok).To(BeTrue())
								Expect(baseDatum).ToNot(BeNil())
								Expect(baseDatum.CreatedTime).ToNot(BeEmpty())
								Expect(baseDatum.CreatedUserID).To(BeEmpty())
							}
						})

						It("has the correct stored dataset data", func() {
							datasetBeforeCreateData := append(append(datasetExistingOtherData, datasetExistingOneData...), datasetExistingTwoData...)
							ValidateDatasetData(testMongoCollection, bson.M{"type": bson.M{"$ne": "upload"}, "createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, datasetBeforeCreateData)
							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
							ValidateDatasetData(testMongoCollection, bson.M{"type": bson.M{"$ne": "upload"}, "createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, append(datasetBeforeCreateData, datasetData...))
						})

						Context("with agent specified", func() {
							var agentUserID string

							BeforeEach(func() {
								agentUserID = app.NewID()
								mongoSession.SetAgent(&TestAgent{false, agentUserID})
							})

							It("sets the created time and created user id on the dataset data", func() {
								Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
								for _, datasetDatum := range datasetData {
									baseDatum, ok := datasetDatum.(*types.Base)
									Expect(ok).To(BeTrue())
									Expect(baseDatum).ToNot(BeNil())
									Expect(baseDatum.CreatedTime).ToNot(BeEmpty())
									Expect(baseDatum.CreatedUserID).To(Equal(agentUserID))
								}
							})

							It("has the correct stored dataset data", func() {
								datasetBeforeCreateData := append(append(datasetExistingOtherData, datasetExistingOneData...), datasetExistingTwoData...)
								ValidateDatasetData(testMongoCollection, bson.M{"type": bson.M{"$ne": "upload"}, "createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, datasetBeforeCreateData)
								Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
								ValidateDatasetData(testMongoCollection, bson.M{"type": bson.M{"$ne": "upload"}, "createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, datasetBeforeCreateData)
								ValidateDatasetData(testMongoCollection, bson.M{"type": bson.M{"$ne": "upload"}, "createdTime": bson.M{"$exists": true}, "createdUserId": agentUserID}, datasetData)
							})
						})
					})

					Context("FindEarliestDatasetDataTime", func() {
						It("succeeds with no earliest time if it finds there is no data time", func() {
							earliestTime, err := mongoSession.FindEarliestDatasetDataTime(dataset)
							Expect(err).ToNot(HaveOccurred())
							Expect(earliestTime).To(Equal(""))
						})

						Context("with dataset data", func() {
							var expectedEarliestTime string

							BeforeEach(func() {
								Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
								expectedEarliestTime = *datasetData[0].(*types.Base).Time
							})

							It("succeeds if it finds the earliest dataset data time", func() {
								earliestTime, err := mongoSession.FindEarliestDatasetDataTime(dataset)
								Expect(err).ToNot(HaveOccurred())
								Expect(earliestTime).To(Equal(expectedEarliestTime))
							})

							It("returns an error if the dataset is missing", func() {
								earliestTime, err := mongoSession.FindEarliestDatasetDataTime(nil)
								Expect(err).To(MatchError("mongo: dataset is missing"))
								Expect(earliestTime).To(Equal(""))
							})

							It("returns an error if the user id is missing", func() {
								dataset.UserID = ""
								earliestTime, err := mongoSession.FindEarliestDatasetDataTime(dataset)
								Expect(err).To(MatchError("mongo: dataset user id is missing"))
								Expect(earliestTime).To(Equal(""))
							})

							It("returns an error if the group id is missing", func() {
								dataset.GroupID = ""
								earliestTime, err := mongoSession.FindEarliestDatasetDataTime(dataset)
								Expect(err).To(MatchError("mongo: dataset group id is missing"))
								Expect(earliestTime).To(Equal(""))
							})

							It("returns an error if the upload id is missing", func() {
								dataset.UploadID = ""
								earliestTime, err := mongoSession.FindEarliestDatasetDataTime(dataset)
								Expect(err).To(MatchError("mongo: dataset upload id is missing"))
								Expect(earliestTime).To(Equal(""))
							})

							It("returns an error if the session is closed", func() {
								mongoSession.Close()
								earliestTime, err := mongoSession.FindEarliestDatasetDataTime(dataset)
								Expect(err).To(MatchError("mongo: session closed"))
								Expect(earliestTime).To(Equal(""))
							})
						})
					})

					Context("ActivateDatasetData", func() {
						BeforeEach(func() {
							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
						})

						It("succeeds if it successfully activates the dataset", func() {
							Expect(mongoSession.ActivateDatasetData(dataset)).To(Succeed())
						})

						It("returns an error if the dataset is missing", func() {
							Expect(mongoSession.ActivateDatasetData(nil)).To(MatchError("mongo: dataset is missing"))
						})

						It("returns an error if the user id is missing", func() {
							dataset.UserID = ""
							Expect(mongoSession.ActivateDatasetData(dataset)).To(MatchError("mongo: dataset user id is missing"))
						})

						It("returns an error if the group id is missing", func() {
							dataset.GroupID = ""
							Expect(mongoSession.ActivateDatasetData(dataset)).To(MatchError("mongo: dataset group id is missing"))
						})

						It("returns an error if the upload id is missing", func() {
							dataset.UploadID = ""
							Expect(mongoSession.ActivateDatasetData(dataset)).To(MatchError("mongo: dataset upload id is missing"))
						})

						It("returns an error if the session is closed", func() {
							mongoSession.Close()
							Expect(mongoSession.ActivateDatasetData(dataset)).To(MatchError("mongo: session closed"))
						})

						It("sets the active on the dataset", func() {
							Expect(mongoSession.ActivateDatasetData(dataset)).To(Succeed())
							Expect(dataset.Active).To(BeTrue())
						})

						It("sets the modified time on the dataset", func() {
							Expect(mongoSession.ActivateDatasetData(dataset)).To(Succeed())
							Expect(dataset.ModifiedTime).ToNot(BeEmpty())
							Expect(dataset.ModifiedUserID).To(BeEmpty())
						})

						It("has the correct stored active dataset", func() {
							ValidateDataset(testMongoCollection, bson.M{"_active": true, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}})
							Expect(mongoSession.ActivateDatasetData(dataset)).To(Succeed())
							ValidateDataset(testMongoCollection, bson.M{"_active": true, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, dataset)
						})

						It("has the correct stored active dataset data", func() {
							ValidateDatasetData(testMongoCollection, bson.M{"_active": true, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, []data.Datum{})
							Expect(mongoSession.ActivateDatasetData(dataset)).To(Succeed())
							ValidateDatasetData(testMongoCollection, bson.M{"_active": true, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, append(datasetData, dataset))
						})

						Context("with agent specified", func() {
							var agentUserID string

							BeforeEach(func() {
								agentUserID = app.NewID()
								mongoSession.SetAgent(&TestAgent{false, agentUserID})
							})

							It("sets the modified time and modified user id on the dataset", func() {
								Expect(mongoSession.ActivateDatasetData(dataset)).To(Succeed())
								Expect(dataset.ModifiedTime).ToNot(BeEmpty())
								Expect(dataset.ModifiedUserID).To(Equal(agentUserID))
							})

							It("has the correct stored active dataset", func() {
								ValidateDataset(testMongoCollection, bson.M{"_active": true, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": agentUserID})
								Expect(mongoSession.ActivateDatasetData(dataset)).To(Succeed())
								ValidateDataset(testMongoCollection, bson.M{"_active": true, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": agentUserID}, dataset)
							})

							It("has the correct stored active dataset data", func() {
								ValidateDatasetData(testMongoCollection, bson.M{"_active": true, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": agentUserID}, []data.Datum{})
								Expect(mongoSession.ActivateDatasetData(dataset)).To(Succeed())
								ValidateDatasetData(testMongoCollection, bson.M{"_active": true, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": agentUserID}, append(datasetData, dataset))
							})
						})
					})

					Context("SetDatasetDataActiveUsingHashes", func() {
						var targetData []data.Datum
						var queryHashes []string
						var uploadID string

						BeforeEach(func() {
							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
							Expect(mongoSession.ActivateDatasetData(dataset)).To(Succeed())
							targetData = []data.Datum{datasetData[0], datasetData[2]}
							queryHashes = []string{}
							for _, targetDatum := range targetData {
								queryHashes = append(queryHashes, targetDatum.DeduplicatorDescriptor().Hash)
							}
							uploadID = dataset.UploadID
						})

						It("succeeds if it successfully sets dataset data active using hashes", func() {
							Expect(mongoSession.SetDatasetDataActiveUsingHashes(dataset, queryHashes, false)).To(Succeed())
						})

						It("returns an error if the dataset is missing", func() {
							Expect(mongoSession.SetDatasetDataActiveUsingHashes(nil, queryHashes, false)).To(MatchError("mongo: dataset is missing"))
						})

						It("returns an error if the dataset user id is missing", func() {
							dataset.UserID = ""
							Expect(mongoSession.SetDatasetDataActiveUsingHashes(dataset, queryHashes, false)).To(MatchError("mongo: dataset user id is missing"))
						})

						It("returns an error if the dataset group id is missing", func() {
							dataset.GroupID = ""
							Expect(mongoSession.SetDatasetDataActiveUsingHashes(dataset, queryHashes, false)).To(MatchError("mongo: dataset group id is missing"))
						})

						It("returns an error if the dataset upload id is missing", func() {
							dataset.UploadID = ""
							Expect(mongoSession.SetDatasetDataActiveUsingHashes(dataset, queryHashes, false)).To(MatchError("mongo: dataset upload id is missing"))
						})

						It("returns an error if the session is closed", func() {
							mongoSession.Close()
							Expect(mongoSession.SetDatasetDataActiveUsingHashes(dataset, queryHashes, false)).To(MatchError("mongo: session closed"))
						})

						It("succeeds if hashes is missing", func() {
							Expect(mongoSession.SetDatasetDataActiveUsingHashes(dataset, nil, false)).To(Succeed())
						})

						It("succeeds if hashes is empty", func() {
							Expect(mongoSession.SetDatasetDataActiveUsingHashes(dataset, []string{}, false)).To(Succeed())
						})

						Context("validating all data is active before", func() {
							BeforeEach(func() {
								ValidateDatasetData(testMongoCollection, bson.M{"uploadId": dataset.UploadID, "_active": false}, nil)
							})

							Context("validating data is deactivated after", func() {
								AfterEach(func() {
									ValidateDatasetData(testMongoCollection, bson.M{"uploadId": dataset.UploadID, "_active": false}, targetData)
								})

								It("succeeds if it successfully deactivates dataset data using hashes", func() {
									Expect(mongoSession.SetDatasetDataActiveUsingHashes(dataset, queryHashes, false)).To(Succeed())
									ValidateDatasetData(testMongoCollection, bson.M{"uploadId": dataset.UploadID, "_active": false, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, targetData)
								})

								Context("with agent specified", func() {
									var agentUserID string

									BeforeEach(func() {
										agentUserID = app.NewID()
										mongoSession.SetAgent(&TestAgent{false, agentUserID})
									})

									It("succeeds if it successfully deactivates dataset data using hashes", func() {
										Expect(mongoSession.SetDatasetDataActiveUsingHashes(dataset, queryHashes, false)).To(Succeed())
										ValidateDatasetData(testMongoCollection, bson.M{"uploadId": dataset.UploadID, "_active": false, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": agentUserID}, targetData)
									})
								})
							})

							Context("validating data is unchanged after", func() {
								AfterEach(func() {
									ValidateDatasetData(testMongoCollection, bson.M{"uploadId": uploadID, "_active": false}, nil)
								})

								It("does not deactivate any data if hashes is missing", func() {
									Expect(mongoSession.SetDatasetDataActiveUsingHashes(dataset, nil, false)).To(Succeed())
								})

								It("does not deactivate any data if hashes is empty", func() {
									Expect(mongoSession.SetDatasetDataActiveUsingHashes(dataset, []string{}, false)).To(Succeed())
								})

								It("does not deactivate any data if different user id", func() {
									dataset.UserID = app.NewID()
									Expect(mongoSession.SetDatasetDataActiveUsingHashes(dataset, queryHashes, false)).To(Succeed())
								})

								It("does not deactivate any data if different group id", func() {
									dataset.GroupID = app.NewID()
									Expect(mongoSession.SetDatasetDataActiveUsingHashes(dataset, queryHashes, false)).To(Succeed())
								})

								It("does not deactivate any data if different upload id", func() {
									dataset.UploadID = app.NewID()
									Expect(mongoSession.SetDatasetDataActiveUsingHashes(dataset, queryHashes, false)).To(Succeed())
								})

								It("does not deactivate any data if active", func() {
									Expect(mongoSession.SetDatasetDataActiveUsingHashes(dataset, queryHashes, true)).To(Succeed())
								})
							})
						})
					})

					Context("SetDeviceDataActiveUsingHashes", func() {
						var targetData []data.Datum
						var queryHashes []string
						var deactiveData []data.Datum

						BeforeEach(func() {
							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
							Expect(mongoSession.ActivateDatasetData(dataset)).To(Succeed())
							targetData = []data.Datum{datasetData[0], datasetData[2]}
							queryHashes = []string{}
							for _, targetDatum := range targetData {
								queryHashes = append(queryHashes, targetDatum.DeduplicatorDescriptor().Hash)
							}
							deactiveData = append(append(datasetExistingOneData, datasetExistingOne, datasetExistingTwo), datasetExistingTwoData...)
						})

						It("succeeds if it successfully sets dataset data active using hashes", func() {
							Expect(mongoSession.SetDeviceDataActiveUsingHashes(dataset, queryHashes, false)).To(Succeed())
						})

						It("returns an error if the dataset is missing", func() {
							Expect(mongoSession.SetDeviceDataActiveUsingHashes(nil, queryHashes, false)).To(MatchError("mongo: dataset is missing"))
						})

						It("returns an error if the dataset user id is missing", func() {
							dataset.UserID = ""
							Expect(mongoSession.SetDeviceDataActiveUsingHashes(dataset, queryHashes, false)).To(MatchError("mongo: dataset user id is missing"))
						})

						It("returns an error if the dataset group id is missing", func() {
							dataset.GroupID = ""
							Expect(mongoSession.SetDeviceDataActiveUsingHashes(dataset, queryHashes, false)).To(MatchError("mongo: dataset group id is missing"))
						})

						It("returns an error if the dataset device id is missing", func() {
							dataset.DeviceID = nil
							Expect(mongoSession.SetDeviceDataActiveUsingHashes(dataset, queryHashes, false)).To(MatchError("mongo: dataset device id is missing"))
						})

						It("returns an error if the dataset device id is empty", func() {
							dataset.DeviceID = app.StringAsPointer("")
							Expect(mongoSession.SetDeviceDataActiveUsingHashes(dataset, queryHashes, false)).To(MatchError("mongo: dataset device id is missing"))
						})

						It("returns an error if the session is closed", func() {
							mongoSession.Close()
							Expect(mongoSession.SetDeviceDataActiveUsingHashes(dataset, queryHashes, false)).To(MatchError("mongo: session closed"))
						})

						It("succeeds if hashes is missing", func() {
							Expect(mongoSession.SetDeviceDataActiveUsingHashes(dataset, nil, false)).To(Succeed())
						})

						It("succeeds if hashes is empty", func() {
							Expect(mongoSession.SetDeviceDataActiveUsingHashes(dataset, []string{}, false)).To(Succeed())
						})

						Context("validating all data is active before", func() {
							BeforeEach(func() {
								ValidateDatasetData(testMongoCollection, bson.M{"deviceId": *dataset.DeviceID, "_active": false}, deactiveData)
							})

							Context("validating data is deactivated after", func() {
								AfterEach(func() {
									ValidateDatasetData(testMongoCollection, bson.M{"deviceId": *dataset.DeviceID, "_active": false}, append(deactiveData, targetData...))
								})

								It("succeeds if it successfully deactivates dataset data using hashes", func() {
									Expect(mongoSession.SetDeviceDataActiveUsingHashes(dataset, queryHashes, false)).To(Succeed())
									ValidateDatasetData(testMongoCollection, bson.M{"deviceId": *dataset.DeviceID, "_active": false, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, targetData)
								})

								Context("with agent specified", func() {
									var agentUserID string

									BeforeEach(func() {
										agentUserID = app.NewID()
										mongoSession.SetAgent(&TestAgent{false, agentUserID})
									})

									It("succeeds if it successfully deactivates dataset data using hashes", func() {
										Expect(mongoSession.SetDeviceDataActiveUsingHashes(dataset, queryHashes, false)).To(Succeed())
										ValidateDatasetData(testMongoCollection, bson.M{"deviceId": *dataset.DeviceID, "_active": false, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": agentUserID}, targetData)
									})
								})
							})

							Context("validating data is unchanged after", func() {
								AfterEach(func() {
									ValidateDatasetData(testMongoCollection, bson.M{"deviceId": deviceID, "_active": false}, deactiveData)
								})

								It("does not deactivate any data if hashes is missing", func() {
									Expect(mongoSession.SetDeviceDataActiveUsingHashes(dataset, nil, false)).To(Succeed())
								})

								It("does not deactivate any data if hashes is empty", func() {
									Expect(mongoSession.SetDeviceDataActiveUsingHashes(dataset, []string{}, false)).To(Succeed())
								})

								It("does not deactivate any data if different user id", func() {
									dataset.UserID = app.NewID()
									Expect(mongoSession.SetDeviceDataActiveUsingHashes(dataset, queryHashes, false)).To(Succeed())
								})

								It("does not deactivate any data if different group id", func() {
									dataset.GroupID = app.NewID()
									Expect(mongoSession.SetDeviceDataActiveUsingHashes(dataset, queryHashes, false)).To(Succeed())
								})

								It("does not deactivate any data if different device id", func() {
									dataset.DeviceID = app.StringAsPointer(app.NewID())
									Expect(mongoSession.SetDeviceDataActiveUsingHashes(dataset, queryHashes, false)).To(Succeed())
								})

								It("does not deactivate any data if active", func() {
									Expect(mongoSession.SetDeviceDataActiveUsingHashes(dataset, queryHashes, true)).To(Succeed())
								})
							})
						})
					})

					Context("DeactivateOtherDatasetDataAfterTime", func() {
						var afterTime string

						BeforeEach(func() {
							Expect(mongoSession.ActivateDatasetData(datasetExistingOther)).To(Succeed())
							Expect(mongoSession.ActivateDatasetData(datasetExistingOne)).To(Succeed())
							Expect(mongoSession.ActivateDatasetData(datasetExistingTwo)).To(Succeed())
							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
							afterTime = *datasetExistingTwoData[0].(*types.Base).Time
						})

						It("succeeds if it successfully deactivates other dataset data after time", func() {
							Expect(mongoSession.DeactivateOtherDatasetDataAfterTime(dataset, afterTime)).To(Succeed())
						})

						It("returns an error if the dataset is missing", func() {
							Expect(mongoSession.DeactivateOtherDatasetDataAfterTime(nil, afterTime)).To(MatchError("mongo: dataset is missing"))
						})

						It("returns an error if the user id is missing", func() {
							dataset.UserID = ""
							Expect(mongoSession.DeactivateOtherDatasetDataAfterTime(dataset, afterTime)).To(MatchError("mongo: dataset user id is missing"))
						})

						It("returns an error if the group id is missing", func() {
							dataset.GroupID = ""
							Expect(mongoSession.DeactivateOtherDatasetDataAfterTime(dataset, afterTime)).To(MatchError("mongo: dataset group id is missing"))
						})

						It("returns an error if the upload id is missing", func() {
							dataset.UploadID = ""
							Expect(mongoSession.DeactivateOtherDatasetDataAfterTime(dataset, afterTime)).To(MatchError("mongo: dataset upload id is missing"))
						})

						It("returns an error if the device id is missing (nil)", func() {
							dataset.DeviceID = nil
							Expect(mongoSession.DeactivateOtherDatasetDataAfterTime(dataset, afterTime)).To(MatchError("mongo: dataset device id is missing"))
						})

						It("returns an error if the device id is missing (empty)", func() {
							dataset.DeviceID = app.StringAsPointer("")
							Expect(mongoSession.DeactivateOtherDatasetDataAfterTime(dataset, afterTime)).To(MatchError("mongo: dataset device id is missing"))
						})

						It("returns an error if the after time is missing", func() {
							Expect(mongoSession.DeactivateOtherDatasetDataAfterTime(dataset, "")).To(MatchError("mongo: after time is missing"))
						})

						It("returns an error if the session is closed", func() {
							mongoSession.Close()
							Expect(mongoSession.DeactivateOtherDatasetDataAfterTime(dataset, afterTime)).To(MatchError("mongo: session closed"))
						})

						It("has the correct stored inactive dataset", func() {
							ValidateDataset(testMongoCollection, bson.M{"_active": false}, dataset)
							Expect(mongoSession.DeactivateOtherDatasetDataAfterTime(dataset, afterTime)).To(Succeed())
							ValidateDataset(testMongoCollection, bson.M{"_active": false}, dataset)
							ValidateDataset(testMongoCollection, bson.M{"_active": false, "modifiedTime": bson.M{"$exists": false}, "modifiedUserId": bson.M{"$exists": false}}, dataset)
						})

						It("has the correct stored active dataset data", func() {
							ValidateDatasetData(testMongoCollection, bson.M{"_active": false}, append(datasetData, dataset))
							Expect(mongoSession.DeactivateOtherDatasetDataAfterTime(dataset, afterTime)).To(Succeed())
							ValidateDatasetData(testMongoCollection, bson.M{"_active": false}, append(append(datasetData, dataset), datasetExistingTwoData...))
							ValidateDatasetData(testMongoCollection, bson.M{"_active": false, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, datasetExistingTwoData)
						})

						Context("with agent specified", func() {
							var agentUserID string

							BeforeEach(func() {
								agentUserID = app.NewID()
								mongoSession.SetAgent(&TestAgent{false, agentUserID})
							})

							It("has the correct stored inactive dataset", func() {
								ValidateDataset(testMongoCollection, bson.M{"_active": false}, dataset)
								Expect(mongoSession.DeactivateOtherDatasetDataAfterTime(dataset, afterTime)).To(Succeed())
								ValidateDataset(testMongoCollection, bson.M{"_active": false}, dataset)
								ValidateDataset(testMongoCollection, bson.M{"_active": false, "modifiedTime": bson.M{"$exists": false}, "modifiedUserId": bson.M{"$exists": false}}, dataset)
							})

							It("has the correct stored active dataset data", func() {
								ValidateDatasetData(testMongoCollection, bson.M{"_active": false}, append(datasetData, dataset))
								Expect(mongoSession.DeactivateOtherDatasetDataAfterTime(dataset, afterTime)).To(Succeed())
								ValidateDatasetData(testMongoCollection, bson.M{"_active": false}, append(append(datasetData, dataset), datasetExistingTwoData...))
								ValidateDatasetData(testMongoCollection, bson.M{"_active": false, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": agentUserID}, datasetExistingTwoData)
							})
						})
					})

					Context("DeleteOtherDatasetData", func() {
						BeforeEach(func() {
							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
						})

						It("succeeds if it successfully removes all other dataset data", func() {
							Expect(mongoSession.DeleteOtherDatasetData(dataset)).To(Succeed())
						})

						It("returns an error if the dataset is missing", func() {
							Expect(mongoSession.DeleteOtherDatasetData(nil)).To(MatchError("mongo: dataset is missing"))
						})

						It("returns an error if the user id is missing", func() {
							dataset.UserID = ""
							Expect(mongoSession.DeleteOtherDatasetData(dataset)).To(MatchError("mongo: dataset user id is missing"))
						})

						It("returns an error if the group id is missing", func() {
							dataset.GroupID = ""
							Expect(mongoSession.DeleteOtherDatasetData(dataset)).To(MatchError("mongo: dataset group id is missing"))
						})

						It("returns an error if the upload id is missing", func() {
							dataset.UploadID = ""
							Expect(mongoSession.DeleteOtherDatasetData(dataset)).To(MatchError("mongo: dataset upload id is missing"))
						})

						It("returns an error if the device id is missing (nil)", func() {
							dataset.DeviceID = nil
							Expect(mongoSession.DeleteOtherDatasetData(dataset)).To(MatchError("mongo: dataset device id is missing"))
						})

						It("returns an error if the device id is missing (empty)", func() {
							dataset.DeviceID = app.StringAsPointer("")
							Expect(mongoSession.DeleteOtherDatasetData(dataset)).To(MatchError("mongo: dataset device id is missing"))
						})

						It("returns an error if the session is closed", func() {
							mongoSession.Close()
							Expect(mongoSession.DeleteOtherDatasetData(dataset)).To(MatchError("mongo: session closed"))
						})

						It("has the correct stored active dataset", func() {
							ValidateDataset(testMongoCollection, bson.M{}, dataset, datasetExistingOther, datasetExistingOne, datasetExistingTwo)
							ValidateDataset(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": false}, "deletedUserId": bson.M{"$exists": false}}, dataset, datasetExistingOther, datasetExistingOne, datasetExistingTwo)
							Expect(mongoSession.DeleteOtherDatasetData(dataset)).To(Succeed())
							Expect(testMongoCollection.Find(bson.M{"type": "upload"}).Count()).To(Equal(4))
							ValidateDataset(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": false}, "deletedUserId": bson.M{"$exists": false}}, dataset, datasetExistingOther)
						})

						It("has the correct stored active dataset data", func() {
							datasetAfterRemoveData := append(append(datasetData, dataset, datasetExistingOther, datasetExistingOne, datasetExistingTwo), datasetExistingOtherData...)
							datasetBeforeRemoveData := append(append(datasetAfterRemoveData, datasetExistingOneData...), datasetExistingTwoData...)
							ValidateDatasetData(testMongoCollection, bson.M{}, datasetBeforeRemoveData)
							Expect(mongoSession.DeleteOtherDatasetData(dataset)).To(Succeed())
							ValidateDatasetData(testMongoCollection, bson.M{}, datasetAfterRemoveData)
						})

						Context("with agent specified", func() {
							var agentUserID string

							BeforeEach(func() {
								agentUserID = app.NewID()
								mongoSession.SetAgent(&TestAgent{false, agentUserID})
							})

							It("has the correct stored active dataset", func() {
								ValidateDataset(testMongoCollection, bson.M{}, dataset, datasetExistingOther, datasetExistingOne, datasetExistingTwo)
								ValidateDataset(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": false}, "deletedUserId": bson.M{"$exists": false}}, dataset, datasetExistingOther, datasetExistingOne, datasetExistingTwo)
								Expect(mongoSession.DeleteOtherDatasetData(dataset)).To(Succeed())
								Expect(testMongoCollection.Find(bson.M{"type": "upload"}).Count()).To(Equal(4))
								ValidateDataset(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": false}, "deletedUserId": bson.M{"$exists": false}}, dataset, datasetExistingOther)
							})

							It("has the correct stored active dataset data", func() {
								datasetAfterRemoveData := append(append(datasetData, dataset, datasetExistingOther, datasetExistingOne, datasetExistingTwo), datasetExistingOtherData...)
								datasetBeforeRemoveData := append(append(datasetAfterRemoveData, datasetExistingOneData...), datasetExistingTwoData...)
								ValidateDatasetData(testMongoCollection, bson.M{}, datasetBeforeRemoveData)
								Expect(mongoSession.DeleteOtherDatasetData(dataset)).To(Succeed())
								ValidateDatasetData(testMongoCollection, bson.M{}, datasetAfterRemoveData)
							})
						})
					})

					Context("DestroyDataForUserByID", func() {
						var deleteUserID string
						var deleteGroupID string
						var deleteDeviceID string
						var deleteDataset *upload.Upload
						var deleteDatasetData []data.Datum

						BeforeEach(func() {
							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
							deleteUserID = app.NewID()
							deleteGroupID = app.NewID()
							deleteDeviceID = app.NewID()
							deleteDataset = NewDataset(deleteUserID, deleteGroupID, deleteDeviceID)
							deleteDataset.CreatedTime = "2016-09-01T11:00:00Z"
							Expect(testMongoCollection.Insert(deleteDataset)).To(Succeed())
							deleteDatasetData = NewDatasetData(deleteDeviceID)
							Expect(mongoSession.CreateDatasetData(deleteDataset, deleteDatasetData)).To(Succeed())
						})

						It("succeeds if it successfully removes all data", func() {
							Expect(mongoSession.DestroyDataForUserByID(deleteUserID)).To(Succeed())
						})

						It("returns an error if the user id is missing", func() {
							Expect(mongoSession.DestroyDataForUserByID("")).To(MatchError("mongo: user id is missing"))
						})

						It("returns an error if the session is closed", func() {
							mongoSession.Close()
							Expect(mongoSession.DestroyDataForUserByID(deleteUserID)).To(MatchError("mongo: session closed"))
						})

						It("has the correct stored dataset", func() {
							ValidateDataset(testMongoCollection, bson.M{}, dataset, datasetExistingOther, datasetExistingOne, datasetExistingTwo, deleteDataset)
							Expect(mongoSession.DestroyDataForUserByID(deleteUserID)).To(Succeed())
							ValidateDataset(testMongoCollection, bson.M{}, dataset, datasetExistingOther, datasetExistingOne, datasetExistingTwo)
						})

						It("has the correct stored dataset data", func() {
							datasetAfterRemoveData := append(append(append(append(datasetData, dataset, datasetExistingOther, datasetExistingOne, datasetExistingTwo), datasetExistingOtherData...), datasetExistingOneData...), datasetExistingTwoData...)
							datasetBeforeRemoveData := append(append(datasetAfterRemoveData, deleteDataset), deleteDatasetData...)
							ValidateDatasetData(testMongoCollection, bson.M{}, datasetBeforeRemoveData)
							Expect(mongoSession.DestroyDataForUserByID(deleteUserID)).To(Succeed())
							ValidateDatasetData(testMongoCollection, bson.M{}, datasetAfterRemoveData)
						})
					})
				})
			})
		})
	})
})
