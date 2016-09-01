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
	commonMongo "github.com/tidepool-org/platform/store/mongo"
)

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

	dataset.ByUser = userID

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

func ValidateDataset(mongoTestCollection *mgo.Collection, selector bson.M, expectedDatasets ...*upload.Upload) {
	selector["type"] = "upload"
	var actualDatasets []*upload.Upload
	Expect(mongoTestCollection.Find(selector).All(&actualDatasets)).To(Succeed())
	Expect(actualDatasets).To(ConsistOf(expectedDatasets))
}

// TODO: Actually compare dataset data here, not just count. Needs code to read actual data types from Mongo.

func ValidateDatasetData(mongoTestCollection *mgo.Collection, selector bson.M, expectedDatasetData []data.Datum) {
	var actualDatasetData []interface{}
	Expect(mongoTestCollection.Find(selector).All(&actualDatasetData)).To(Succeed())
	Expect(actualDatasetData).To(HaveLen(len(expectedDatasetData)))
}

var _ = Describe("Mongo", func() {
	var logger log.Logger
	var mongoConfig *commonMongo.Config
	var mongoStore *mongo.Store
	var mongoSession store.Session

	BeforeEach(func() {
		logger = log.NewNull()
		mongoConfig = &commonMongo.Config{
			Addresses:  MongoTestAddress(),
			Database:   MongoTestDatabase(),
			Collection: NewTestSuiteID(),
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
		It("returns no error if successful", func() {
			var err error
			mongoStore, err = mongo.New(logger, mongoConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})

		It("returns an error if unsuccessful", func() {
			var err error
			mongoStore, err = mongo.New(nil, nil)
			Expect(err).To(HaveOccurred())
			Expect(mongoStore).To(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			mongoStore, err = mongo.New(logger, mongoConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})

		Context("NewSession", func() {
			It("returns no error if successful", func() {
				var err error
				mongoSession, err = mongoStore.NewSession(log.NewNull())
				Expect(err).ToNot(HaveOccurred())
				Expect(mongoSession).ToNot(BeNil())
			})

			It("returns an error if unsuccessful", func() {
				var err error
				mongoSession, err = mongoStore.NewSession(nil)
				Expect(err).To(HaveOccurred())
				Expect(mongoSession).To(BeNil())
			})
		})

		Context("with a new session", func() {
			BeforeEach(func() {
				var err error
				mongoSession, err = mongoStore.NewSession(log.NewNull())
				Expect(err).ToNot(HaveOccurred())
				Expect(mongoSession).ToNot(BeNil())
			})

			Context("with a dataset", func() {
				var mongoTestSession *mgo.Session
				var mongoTestCollection *mgo.Collection
				var userID string
				var groupID string
				var datasetExistingOne *upload.Upload
				var datasetExistingTwo *upload.Upload
				var dataset *upload.Upload

				BeforeEach(func() {
					mongoTestSession = MongoTestSession().Copy()
					mongoTestCollection = mongoTestSession.DB(mongoConfig.Database).C(mongoConfig.Collection)
					userID = app.NewID()
					groupID = app.NewID()
					datasetExistingOne = NewDataset(userID, groupID)
					Expect(mongoTestCollection.Insert(datasetExistingOne)).To(Succeed())
					datasetExistingTwo = NewDataset(userID, groupID)
					Expect(mongoTestCollection.Insert(datasetExistingTwo)).To(Succeed())
					dataset = NewDataset(userID, groupID)
				})

				AfterEach(func() {
					if mongoTestSession != nil {
						mongoTestSession.Close()
					}
				})

				Context("GetDatasetsForUser", func() {
					BeforeEach(func() {
						Expect(mongoTestCollection.Insert(dataset)).To(Succeed())
					})

					It("returns no error if it successfully finds the user datasets", func() {
						Expect(mongoSession.GetDatasetsForUser(userID)).To(Equal([]*upload.Upload{datasetExistingOne, datasetExistingTwo, dataset}))
					})

					It("returns no error if it successfully does not find another user datasets", func() {
						resultDatasets, err := mongoSession.GetDatasetsForUser(app.NewID())
						Expect(err).ToNot(HaveOccurred())
						Expect(resultDatasets).ToNot(BeNil())
						Expect(resultDatasets).To(BeEmpty())
					})

					It("returns an error if the user id is missing", func() {
						resultDatasets, err := mongoSession.GetDatasetsForUser("")
						Expect(err).To(MatchError("mongo: user id is missing"))
						Expect(resultDatasets).To(BeNil())
					})

					It("returns an error if the session is closed", func() {
						mongoSession.Close()
						resultDatasets, err := mongoSession.GetDatasetsForUser(userID)
						Expect(err).To(MatchError("mongo: session closed"))
						Expect(resultDatasets).To(BeNil())
					})
				})

				Context("GetDataset", func() {
					BeforeEach(func() {
						Expect(mongoTestCollection.Insert(dataset)).To(Succeed())
					})

					It("returns no error if it successfully finds the dataset", func() {
						Expect(mongoSession.GetDataset(dataset.UploadID)).To(Equal(dataset))
					})

					It("returns an error if the dataset id is missing", func() {
						resultDataset, err := mongoSession.GetDataset("")
						Expect(err).To(MatchError("mongo: dataset id is missing"))
						Expect(resultDataset).To(BeNil())
					})

					It("returns an error if the session is closed", func() {
						mongoSession.Close()
						resultDataset, err := mongoSession.GetDataset(dataset.UploadID)
						Expect(err).To(MatchError("mongo: session closed"))
						Expect(resultDataset).To(BeNil())
					})

					It("returns an error if the dataset cannot be found", func() {
						resultDataset, err := mongoSession.GetDataset("not-found")
						Expect(err).To(MatchError("mongo: unable to get dataset; not found"))
						Expect(resultDataset).To(BeNil())
					})
				})

				Context("CreateDataset", func() {
					It("returns no error if it successfully creates the dataset", func() {
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

					It("has the correct stored datasets", func() {
						ValidateDataset(mongoTestCollection, bson.M{}, datasetExistingOne, datasetExistingTwo)
						Expect(mongoSession.CreateDataset(dataset)).To(Succeed())
						ValidateDataset(mongoTestCollection, bson.M{}, datasetExistingOne, datasetExistingTwo, dataset)
					})
				})

				Context("UpdateDataset", func() {
					BeforeEach(func() {
						Expect(mongoSession.CreateDataset(dataset)).To(Succeed())
					})

					Context("with data state closed", func() {
						BeforeEach(func() {
							dataset.DataState = "closed"
						})

						It("returns no error if it successfully updates the dataset", func() {
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

					It("has the correct stored datasets", func() {
						ValidateDataset(mongoTestCollection, bson.M{}, datasetExistingOne, datasetExistingTwo, dataset)
						dataset.DataState = "closed"
						Expect(mongoSession.UpdateDataset(dataset)).To(Succeed())
						ValidateDataset(mongoTestCollection, bson.M{}, datasetExistingOne, datasetExistingTwo, dataset)
					})
				})

				Context("DeleteDataset", func() {
					BeforeEach(func() {
						Expect(mongoTestCollection.Insert(dataset)).To(Succeed())
					})

					It("returns no error if it successfully deletes the dataset", func() {
						Expect(mongoSession.DeleteDataset(dataset.UploadID)).To(Succeed())
						ValidateDataset(mongoTestCollection, bson.M{}, datasetExistingOne, datasetExistingTwo)
					})

					It("returns no error if it successfully ignores an unknown dataset", func() {
						Expect(mongoSession.DeleteDataset(app.NewID())).To(Succeed())
						ValidateDataset(mongoTestCollection, bson.M{}, datasetExistingOne, datasetExistingTwo, dataset)
					})

					It("returns an error if the dataset id is missing", func() {
						Expect(mongoSession.DeleteDataset("")).To(MatchError("mongo: dataset id is missing"))
					})

					It("returns an error if the session is closed", func() {
						mongoSession.Close()
						Expect(mongoSession.DeleteDataset(userID)).To(MatchError("mongo: session closed"))
					})
				})

				Context("with data", func() {
					var datasetExistingOneData []data.Datum
					var datasetExistingTwoData []data.Datum
					var datasetData []data.Datum

					BeforeEach(func() {
						Expect(mongoSession.CreateDataset(dataset)).To(Succeed())
						datasetExistingOneData = NewDatasetData()
						Expect(mongoSession.CreateDatasetData(datasetExistingOne, datasetExistingOneData)).To(Succeed())
						datasetExistingTwoData = NewDatasetData()
						Expect(mongoSession.CreateDatasetData(datasetExistingTwo, datasetExistingTwoData)).To(Succeed())
						datasetData = NewDatasetData()
					})

					Context("CreateDatasetData", func() {
						It("returns no error if it successfully creates the dataset data", func() {
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

						It("has the correct stored dataset data", func() {
							datasetBeforeCreateData := append(datasetExistingOneData, datasetExistingTwoData...)
							ValidateDatasetData(mongoTestCollection, bson.M{"type": bson.M{"$ne": "upload"}}, datasetBeforeCreateData)
							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
							ValidateDatasetData(mongoTestCollection, bson.M{"type": bson.M{"$ne": "upload"}}, append(datasetBeforeCreateData, datasetData...))
						})
					})

					Context("ActivateDatasetData", func() {
						BeforeEach(func() {
							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
						})

						It("returns no error if it successfully activates the dataset", func() {
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

						It("has the correct stored active dataset", func() {
							ValidateDataset(mongoTestCollection, bson.M{"_active": true})
							Expect(mongoSession.ActivateDatasetData(dataset)).To(Succeed())
							ValidateDataset(mongoTestCollection, bson.M{"_active": true}, dataset)
						})

						It("has the correct stored active dataset data", func() {
							ValidateDatasetData(mongoTestCollection, bson.M{"_active": true}, []data.Datum{})
							Expect(mongoSession.ActivateDatasetData(dataset)).To(Succeed())
							ValidateDatasetData(mongoTestCollection, bson.M{"_active": true}, append(datasetData, dataset))
						})
					})

					Context("DeleteOtherDatasetData", func() {
						BeforeEach(func() {
							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
						})

						It("returns no error if it successfully removes all other dataset data", func() {
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
							ValidateDataset(mongoTestCollection, bson.M{}, dataset, datasetExistingOne, datasetExistingTwo)
							Expect(mongoSession.DeleteOtherDatasetData(dataset)).To(Succeed())
							ValidateDataset(mongoTestCollection, bson.M{}, dataset, datasetExistingOne, datasetExistingTwo)
						})

						It("has the correct stored active dataset data", func() {
							datasetAfterRemoveData := append(datasetData, dataset, datasetExistingOne, datasetExistingTwo)
							datasetBeforeRemoveData := append(append(datasetAfterRemoveData, datasetExistingOneData...), datasetExistingTwoData...)
							ValidateDatasetData(mongoTestCollection, bson.M{}, datasetBeforeRemoveData)
							Expect(mongoSession.DeleteOtherDatasetData(dataset)).To(Succeed())
							ValidateDatasetData(mongoTestCollection, bson.M{}, datasetAfterRemoveData)
						})
					})
				})
			})
		})
	})
})
