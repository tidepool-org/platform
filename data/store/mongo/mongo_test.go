package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/data/types/base"
	"github.com/tidepool-org/platform/data/types/base/upload"
	"github.com/tidepool-org/platform/log"
	baseMongo "github.com/tidepool-org/platform/store/mongo"
	testMongo "github.com/tidepool-org/platform/test/mongo"
)

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

func NewDataset(userID string, groupID string) *upload.Upload {
	dataset := upload.Init()
	Expect(dataset).ToNot(BeNil())

	dataset.GroupID = groupID
	dataset.UserID = userID

	dataset.ClockDriftOffset = app.IntegerAsPointer(0)
	dataset.ConversionOffset = app.IntegerAsPointer(0)
	dataset.DeviceID = app.StringAsPointer("tesla-aps-4242424242")
	dataset.DeviceTime = app.StringAsPointer("2015-05-06T14:08:09")
	dataset.Time = app.StringAsPointer("2015-05-06T07:08:09-07:00")
	dataset.TimezoneOffset = app.IntegerAsPointer(-420)

	dataset.ComputerTime = app.StringAsPointer("2015-06-07T08:09:10")
	dataset.DeviceManufacturers = app.StringArrayAsPointer([]string{"Tesla"})
	dataset.DeviceModel = app.StringAsPointer("1234")
	dataset.DeviceSerialNumber = app.StringAsPointer("567890")
	dataset.DeviceTags = app.StringArrayAsPointer([]string{"insulin-pump"})
	dataset.TimeProcessing = app.StringAsPointer("utc-bootstrapping")
	dataset.TimeZone = app.StringAsPointer("US/Pacific")
	dataset.Version = app.StringAsPointer("0.260.1")

	return dataset
}

func NewDatasetData() []data.Datum {
	datasetData := []data.Datum{}
	for count := 0; count < 3; count++ {
		baseDatum := &base.Base{}
		baseDatum.Init()

		baseDatum.Type = "test"

		baseDatum.ClockDriftOffset = app.IntegerAsPointer(0)
		baseDatum.ConversionOffset = app.IntegerAsPointer(0)
		baseDatum.DeviceID = app.StringAsPointer("tesla-aps-4242424242")
		baseDatum.DeviceTime = app.StringAsPointer("2015-05-06T14:08:09")
		baseDatum.Time = app.StringAsPointer("2015-05-06T07:08:09-07:00")
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
				var datasetExistingOne *upload.Upload
				var datasetExistingTwo *upload.Upload
				var dataset *upload.Upload

				BeforeEach(func() {
					testMongoSession = testMongo.Session().Copy()
					testMongoCollection = testMongoSession.DB(mongoConfig.Database).C(mongoConfig.Collection)
					userID = app.NewID()
					groupID = app.NewID()
					datasetExistingOne = NewDataset(userID, groupID)
					datasetExistingOne.CreatedTime = "2016-09-01T12:00:00Z"
					Expect(testMongoCollection.Insert(datasetExistingOne)).To(Succeed())
					datasetExistingTwo = NewDataset(userID, groupID)
					datasetExistingTwo.CreatedTime = "2016-09-01T10:00:00Z"
					Expect(testMongoCollection.Insert(datasetExistingTwo)).To(Succeed())
					dataset = NewDataset(userID, groupID)
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
						Expect(mongoSession.GetDatasetsForUserByID(userID, filter, pagination)).To(Equal([]*upload.Upload{datasetExistingOne, dataset, datasetExistingTwo}))
					})

					It("succeeds if the filter is not specified", func() {
						Expect(mongoSession.GetDatasetsForUserByID(userID, nil, pagination)).To(Equal([]*upload.Upload{datasetExistingOne, dataset, datasetExistingTwo}))
					})

					It("succeeds if the pagination is not specified", func() {
						Expect(mongoSession.GetDatasetsForUserByID(userID, filter, nil)).To(Equal([]*upload.Upload{datasetExistingOne, dataset, datasetExistingTwo}))
					})

					It("succeeds if the pagination size is not default", func() {
						pagination.Size = 2
						Expect(mongoSession.GetDatasetsForUserByID(userID, filter, pagination)).To(Equal([]*upload.Upload{datasetExistingOne, dataset}))
					})

					It("succeeds if the pagination page and size is not default", func() {
						pagination.Page = 1
						pagination.Size = 2
						Expect(mongoSession.GetDatasetsForUserByID(userID, filter, pagination)).To(Equal([]*upload.Upload{datasetExistingTwo}))
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
						ValidateDataset(testMongoCollection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, datasetExistingOne, datasetExistingTwo)
						Expect(mongoSession.CreateDataset(dataset)).To(Succeed())
						ValidateDataset(testMongoCollection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, datasetExistingOne, datasetExistingTwo, dataset)
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
							ValidateDataset(testMongoCollection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, datasetExistingOne, datasetExistingTwo)
							Expect(mongoSession.CreateDataset(dataset)).To(Succeed())
							ValidateDataset(testMongoCollection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, datasetExistingOne, datasetExistingTwo)
							ValidateDataset(testMongoCollection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": agentUserID}, dataset)
						})
					})
				})

				Context("UpdateDataset", func() {
					BeforeEach(func() {
						dataset.CreatedTime = "2016-09-01T11:00:00Z"
						Expect(testMongoCollection.Insert(dataset)).To(Succeed())
					})

					Context("with data state closed", func() {
						BeforeEach(func() {
							dataset.DataState = "closed"
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
						dataset.DataState = "closed"
						Expect(mongoSession.UpdateDataset(dataset)).To(Succeed())
						Expect(dataset.ModifiedTime).ToNot(BeEmpty())
						Expect(dataset.ModifiedUserID).To(BeEmpty())
					})

					It("has the correct stored datasets", func() {
						ValidateDataset(testMongoCollection, bson.M{}, datasetExistingOne, datasetExistingTwo, dataset)
						ValidateDataset(testMongoCollection, bson.M{"modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}})
						dataset.DataState = "closed"
						Expect(mongoSession.UpdateDataset(dataset)).To(Succeed())
						ValidateDataset(testMongoCollection, bson.M{}, datasetExistingOne, datasetExistingTwo, dataset)
						ValidateDataset(testMongoCollection, bson.M{"modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, dataset)
					})

					Context("with agent specified", func() {
						var agentUserID string

						BeforeEach(func() {
							agentUserID = app.NewID()
							mongoSession.SetAgent(&TestAgent{false, agentUserID})
						})

						It("sets the modified time and modified user id", func() {
							dataset.DataState = "closed"
							Expect(mongoSession.UpdateDataset(dataset)).To(Succeed())
							Expect(dataset.ModifiedTime).ToNot(BeEmpty())
							Expect(dataset.ModifiedUserID).To(Equal(agentUserID))
						})

						It("has the correct stored datasets", func() {
							ValidateDataset(testMongoCollection, bson.M{}, datasetExistingOne, datasetExistingTwo, dataset)
							ValidateDataset(testMongoCollection, bson.M{"modifiedTime": bson.M{"$exists": true}, "modifiedUserId": agentUserID})
							dataset.DataState = "closed"
							Expect(mongoSession.UpdateDataset(dataset)).To(Succeed())
							ValidateDataset(testMongoCollection, bson.M{}, datasetExistingOne, datasetExistingTwo, dataset)
							ValidateDataset(testMongoCollection, bson.M{"modifiedTime": bson.M{"$exists": true}, "modifiedUserId": agentUserID}, dataset)
						})
					})
				})

				Context("with data", func() {
					var datasetExistingOneData []data.Datum
					var datasetExistingTwoData []data.Datum
					var datasetData []data.Datum

					BeforeEach(func() {
						dataset.CreatedTime = "2016-09-01T11:00:00Z"
						Expect(testMongoCollection.Insert(dataset)).To(Succeed())
						datasetExistingOneData = NewDatasetData()
						Expect(mongoSession.CreateDatasetData(datasetExistingOne, datasetExistingOneData)).To(Succeed())
						datasetExistingTwoData = NewDatasetData()
						Expect(mongoSession.CreateDatasetData(datasetExistingTwo, datasetExistingTwoData)).To(Succeed())
						datasetData = NewDatasetData()
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
								baseDatum, ok := datasetDatum.(*base.Base)
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
								baseDatum, ok := datasetDatum.(*base.Base)
								Expect(ok).To(BeTrue())
								Expect(baseDatum).ToNot(BeNil())
								Expect(baseDatum.Active).To(BeFalse())
							}
						})

						It("sets the created time on the dataset data", func() {
							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
							for _, datasetDatum := range datasetData {
								baseDatum, ok := datasetDatum.(*base.Base)
								Expect(ok).To(BeTrue())
								Expect(baseDatum).ToNot(BeNil())
								Expect(baseDatum.CreatedTime).ToNot(BeEmpty())
								Expect(baseDatum.CreatedUserID).To(BeEmpty())
							}
						})

						It("has the correct stored dataset data", func() {
							datasetBeforeCreateData := append(datasetExistingOneData, datasetExistingTwoData...)
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
									baseDatum, ok := datasetDatum.(*base.Base)
									Expect(ok).To(BeTrue())
									Expect(baseDatum).ToNot(BeNil())
									Expect(baseDatum.CreatedTime).ToNot(BeEmpty())
									Expect(baseDatum.CreatedUserID).To(Equal(agentUserID))
								}
							})

							It("has the correct stored dataset data", func() {
								datasetBeforeCreateData := append(datasetExistingOneData, datasetExistingTwoData...)
								ValidateDatasetData(testMongoCollection, bson.M{"type": bson.M{"$ne": "upload"}, "createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, datasetBeforeCreateData)
								Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
								ValidateDatasetData(testMongoCollection, bson.M{"type": bson.M{"$ne": "upload"}, "createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, datasetBeforeCreateData)
								ValidateDatasetData(testMongoCollection, bson.M{"type": bson.M{"$ne": "upload"}, "createdTime": bson.M{"$exists": true}, "createdUserId": agentUserID}, datasetData)
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
							ValidateDataset(testMongoCollection, bson.M{}, dataset, datasetExistingOne, datasetExistingTwo)
							ValidateDataset(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": false}, "deletedUserId": bson.M{"$exists": false}}, dataset, datasetExistingOne, datasetExistingTwo)
							Expect(mongoSession.DeleteOtherDatasetData(dataset)).To(Succeed())
							Expect(testMongoCollection.Find(bson.M{"type": "upload"}).Count()).To(Equal(3))
							ValidateDataset(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": false}, "deletedUserId": bson.M{"$exists": false}}, dataset)
						})

						It("has the correct stored active dataset data", func() {
							datasetAfterRemoveData := append(datasetData, dataset, datasetExistingOne, datasetExistingTwo)
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
								ValidateDataset(testMongoCollection, bson.M{}, dataset, datasetExistingOne, datasetExistingTwo)
								ValidateDataset(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": false}, "deletedUserId": bson.M{"$exists": false}}, dataset, datasetExistingOne, datasetExistingTwo)
								Expect(mongoSession.DeleteOtherDatasetData(dataset)).To(Succeed())
								Expect(testMongoCollection.Find(bson.M{"type": "upload"}).Count()).To(Equal(3))
								ValidateDataset(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": false}, "deletedUserId": bson.M{"$exists": false}}, dataset)
							})

							It("has the correct stored active dataset data", func() {
								datasetAfterRemoveData := append(datasetData, dataset, datasetExistingOne, datasetExistingTwo)
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
						var deleteDataset *upload.Upload
						var deleteDatasetData []data.Datum

						BeforeEach(func() {
							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
							deleteUserID = app.NewID()
							deleteGroupID = app.NewID()
							deleteDataset = NewDataset(deleteUserID, deleteGroupID)
							deleteDataset.CreatedTime = "2016-09-01T11:00:00Z"
							Expect(testMongoCollection.Insert(deleteDataset)).To(Succeed())
							deleteDatasetData = NewDatasetData()
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
							ValidateDataset(testMongoCollection, bson.M{}, dataset, datasetExistingOne, datasetExistingTwo, deleteDataset)
							Expect(mongoSession.DestroyDataForUserByID(deleteUserID)).To(Succeed())
							ValidateDataset(testMongoCollection, bson.M{}, dataset, datasetExistingOne, datasetExistingTwo)
						})

						It("has the correct stored dataset data", func() {
							datasetAfterRemoveData := append(append(append(datasetData, dataset, datasetExistingOne, datasetExistingTwo), datasetExistingOneData...), datasetExistingTwoData...)
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
