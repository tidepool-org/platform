package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"context"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/storeDEPRECATED"
	"github.com/tidepool-org/platform/data/storeDEPRECATED/mongo"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/data/types/upload"
	dataTypesUploadTest "github.com/tidepool-org/platform/data/types/upload/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

func NewDataSet(userID string, deviceID string) *upload.Upload {
	dataSet := dataTypesUploadTest.RandomUpload()
	dataSet.Active = true
	dataSet.ArchivedDataSetID = nil
	dataSet.ArchivedTime = nil
	dataSet.CreatedTime = nil
	dataSet.CreatedUserID = nil
	dataSet.DeletedTime = nil
	dataSet.DeletedUserID = nil
	dataSet.DeviceID = pointer.FromString(deviceID)
	dataSet.Location.GPS.Origin.Time = nil
	dataSet.ModifiedTime = nil
	dataSet.ModifiedUserID = nil
	dataSet.Origin.Time = nil
	dataSet.UserID = pointer.FromString(userID)
	return dataSet
}

func NewDataSetData(deviceID string) []data.Datum {
	dataSetData := []data.Datum{}
	for count := 0; count < 3; count++ {
		datum := dataTypesTest.NewBase()
		datum.Active = false
		datum.ArchivedDataSetID = nil
		datum.ArchivedTime = nil
		datum.CreatedTime = nil
		datum.CreatedUserID = nil
		datum.DeletedTime = nil
		datum.DeletedUserID = nil
		datum.DeviceID = pointer.FromString(deviceID)
		datum.ModifiedTime = nil
		datum.ModifiedUserID = nil
		dataSetData = append(dataSetData, datum)
	}
	return dataSetData
}

func CloneDataSetData(dataSetData []data.Datum) []data.Datum {
	clonedDataSetData := []data.Datum{}
	for _, dataSetDatum := range dataSetData {
		if datum, ok := dataSetDatum.(*types.Base); ok {
			clonedDataSetData = append(clonedDataSetData, dataTypesTest.CloneBase(datum))
		}
	}
	return clonedDataSetData
}

func ValidateDataSet(testMongoCollection *mgo.Collection, query bson.M, filter bson.M, expectedDataSets ...*upload.Upload) {
	query["type"] = "upload"
	filter["_id"] = 0
	var actualDataSets []*upload.Upload
	Expect(testMongoCollection.Find(query).Select(filter).All(&actualDataSets)).To(Succeed())
	Expect(actualDataSets).To(ConsistOf(DataSetsAsInterface(expectedDataSets)...))
}

func DataSetsAsInterface(dataSets []*upload.Upload) []interface{} {
	var dataSetsAsInterface []interface{}
	for _, dataSet := range dataSets {
		dataSetsAsInterface = append(dataSetsAsInterface, dataSet)
	}
	return dataSetsAsInterface
}

func ValidateDataSetData(testMongoCollection *mgo.Collection, query bson.M, filter bson.M, expectedDataSetData []data.Datum) {
	query["type"] = bson.M{"$ne": "upload"}
	filter["_id"] = 0
	var actualDataSetData []interface{}
	Expect(testMongoCollection.Find(query).Select(filter).All(&actualDataSetData)).To(Succeed())
	Expect(actualDataSetData).To(ConsistOf(DataSetDataAsInterface(expectedDataSetData)...))
}

func DataSetDataAsInterface(dataSetData []data.Datum) []interface{} {
	var dataSetDataAsInterface []interface{}
	for _, dataSetDatum := range dataSetData {
		dataSetDataAsInterface = append(dataSetDataAsInterface, DataSetDatumAsInterface(dataSetDatum))
	}
	return dataSetDataAsInterface
}

func DataSetDatumAsInterface(dataSetDatum data.Datum) interface{} {
	bytes, err := bson.Marshal(dataSetDatum)
	Expect(err).ToNot(HaveOccurred())
	Expect(bytes).ToNot(BeNil())
	var dataSetDatumAsInterface interface{}
	Expect(bson.Unmarshal(bytes, &dataSetDatumAsInterface)).To(Succeed())
	return dataSetDatumAsInterface
}

var _ = Describe("Mongo", func() {
	var logger log.Logger
	var mongoConfig *storeStructuredMongo.Config
	var mongoStore *mongo.Store
	var mongoSession storeDEPRECATED.DataSession

	BeforeEach(func() {
		logger = logTest.NewLogger()
		mongoConfig = storeStructuredMongoTest.NewConfig()
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
			mongoStore, err = mongo.NewStore(nil, nil)
			Expect(err).To(HaveOccurred())
			Expect(mongoStore).To(BeNil())
		})

		It("returns a new store and no error if successful", func() {
			var err error
			mongoStore, err = mongo.NewStore(mongoConfig, logger)
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			mongoStore, err = mongo.NewStore(mongoConfig, logger)
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})

		Context("NewDataSession", func() {
			It("returns a new session", func() {
				mongoSession = mongoStore.NewDataSession()
				Expect(mongoSession).ToNot(BeNil())
			})
		})

		Context("with a new session", func() {
			BeforeEach(func() {
				mongoSession = mongoStore.NewDataSession()
				Expect(mongoSession).ToNot(BeNil())
			})

			Context("with persisted data", func() {
				var ctx context.Context
				var testMongoSession *mgo.Session
				var testMongoCollection *mgo.Collection
				var userID string
				var deviceID string
				var dataSetExistingOther *upload.Upload
				var dataSetExistingOne *upload.Upload
				var dataSetExistingTwo *upload.Upload
				var dataSet *upload.Upload

				BeforeEach(func() {
					ctx = log.NewContextWithLogger(context.Background(), logger)
					testMongoSession = storeStructuredMongoTest.Session().Copy()
					testMongoCollection = testMongoSession.DB(mongoConfig.Database).C(mongoConfig.CollectionPrefix + "deviceData")
					userID = userTest.RandomID()
					deviceID = dataTest.NewDeviceID()
					dataSetExistingOther = NewDataSet(userTest.RandomID(), dataTest.NewDeviceID())
					dataSetExistingOther.CreatedTime = pointer.FromString("2016-09-01T12:00:00Z")
					Expect(testMongoCollection.Insert(dataSetExistingOther)).To(Succeed())
					dataSetExistingOne = NewDataSet(userID, deviceID)
					dataSetExistingOne.CreatedTime = pointer.FromString("2016-09-01T12:30:00Z")
					Expect(testMongoCollection.Insert(dataSetExistingOne)).To(Succeed())
					dataSetExistingTwo = NewDataSet(userID, deviceID)
					dataSetExistingTwo.CreatedTime = pointer.FromString("2016-09-01T10:00:00Z")
					Expect(testMongoCollection.Insert(dataSetExistingTwo)).To(Succeed())
					dataSet = NewDataSet(userID, deviceID)
				})

				AfterEach(func() {
					if testMongoSession != nil {
						testMongoSession.Close()
					}
				})

				Context("GetDataSetsForUserByID", func() {
					var filter *storeDEPRECATED.Filter
					var pagination *page.Pagination

					BeforeEach(func() {
						dataSet.CreatedTime = pointer.FromString("2016-09-01T11:00:00Z")
						Expect(testMongoCollection.Insert(dataSet)).To(Succeed())
						filter = storeDEPRECATED.NewFilter()
						pagination = page.NewPagination()
					})

					It("succeeds if it successfully finds the user data sets", func() {
						Expect(mongoSession.GetDataSetsForUserByID(ctx, userID, filter, pagination)).To(ConsistOf([]*upload.Upload{dataSetExistingOne, dataSet, dataSetExistingTwo}))
					})

					It("succeeds if the filter is not specified", func() {
						Expect(mongoSession.GetDataSetsForUserByID(ctx, userID, nil, pagination)).To(ConsistOf([]*upload.Upload{dataSetExistingOne, dataSet, dataSetExistingTwo}))
					})

					It("succeeds if the pagination is not specified", func() {
						Expect(mongoSession.GetDataSetsForUserByID(ctx, userID, filter, nil)).To(ConsistOf([]*upload.Upload{dataSetExistingOne, dataSet, dataSetExistingTwo}))
					})

					It("succeeds if the pagination size is not default", func() {
						pagination.Size = 2
						Expect(mongoSession.GetDataSetsForUserByID(ctx, userID, filter, pagination)).To(ConsistOf([]*upload.Upload{dataSetExistingOne, dataSet}))
					})

					It("succeeds if the pagination page and size is not default", func() {
						pagination.Page = 1
						pagination.Size = 2
						Expect(mongoSession.GetDataSetsForUserByID(ctx, userID, filter, pagination)).To(ConsistOf([]*upload.Upload{dataSetExistingTwo}))
					})

					It("succeeds if it successfully does not find another user data sets", func() {
						resultDataSets, err := mongoSession.GetDataSetsForUserByID(ctx, userTest.RandomID(), filter, pagination)
						Expect(err).ToNot(HaveOccurred())
						Expect(resultDataSets).ToNot(BeNil())
						Expect(resultDataSets).To(BeEmpty())
					})

					It("returns an error if the user id is missing", func() {
						resultDataSets, err := mongoSession.GetDataSetsForUserByID(ctx, "", filter, pagination)
						Expect(err).To(MatchError("user id is missing"))
						Expect(resultDataSets).To(BeNil())
					})

					It("returns an error if the pagination page is less than minimum", func() {
						pagination.Page = -1
						resultDataSets, err := mongoSession.GetDataSetsForUserByID(ctx, userID, filter, pagination)
						Expect(err).To(MatchError("pagination is invalid; value -1 is not greater than or equal to 0"))
						Expect(resultDataSets).To(BeNil())
					})

					It("returns an error if the pagination size is less than minimum", func() {
						pagination.Size = 0
						resultDataSets, err := mongoSession.GetDataSetsForUserByID(ctx, userID, filter, pagination)
						Expect(err).To(MatchError("pagination is invalid; value 0 is not between 1 and 100"))
						Expect(resultDataSets).To(BeNil())
					})

					It("returns an error if the pagination size is greater than maximum", func() {
						pagination.Size = 101
						resultDataSets, err := mongoSession.GetDataSetsForUserByID(ctx, userID, filter, pagination)
						Expect(err).To(MatchError("pagination is invalid; value 101 is not between 1 and 100"))
						Expect(resultDataSets).To(BeNil())
					})

					It("returns an error if the session is closed", func() {
						mongoSession.Close()
						resultDataSets, err := mongoSession.GetDataSetsForUserByID(ctx, userID, filter, pagination)
						Expect(err).To(MatchError("session closed"))
						Expect(resultDataSets).To(BeNil())
					})

					Context("with deleted data set", func() {
						BeforeEach(func() {
							dataSet.DeletedTime = pointer.FromString("2016-09-01T13:00:00Z")
							Expect(testMongoCollection.Update(bson.M{"id": dataSet.ID}, dataSet)).To(Succeed())
						})

						It("succeeds if it successfully finds the non-deleted user data sets", func() {
							Expect(mongoSession.GetDataSetsForUserByID(ctx, userID, filter, pagination)).To(ConsistOf([]*upload.Upload{dataSetExistingOne, dataSetExistingTwo}))
						})

						It("succeeds if it successfully finds all the user data sets", func() {
							filter.Deleted = true
							Expect(mongoSession.GetDataSetsForUserByID(ctx, userID, filter, pagination)).To(ConsistOf([]*upload.Upload{dataSetExistingOne, dataSet, dataSetExistingTwo}))
						})
					})
				})

				Context("GetDataSetByID", func() {
					BeforeEach(func() {
						dataSet.CreatedTime = pointer.FromString("2016-09-01T11:00:00Z")
						Expect(testMongoCollection.Insert(dataSet)).To(Succeed())
					})

					It("succeeds if it successfully finds the data set", func() {
						Expect(mongoSession.GetDataSetByID(ctx, *dataSet.UploadID)).To(Equal(dataSet))
					})

					It("returns an error if the data set id is missing", func() {
						resultDataSet, err := mongoSession.GetDataSetByID(ctx, "")
						Expect(err).To(MatchError("data set id is missing"))
						Expect(resultDataSet).To(BeNil())
					})

					It("returns an error if the session is closed", func() {
						mongoSession.Close()
						resultDataSet, err := mongoSession.GetDataSetByID(ctx, *dataSet.UploadID)
						Expect(err).To(MatchError("session closed"))
						Expect(resultDataSet).To(BeNil())
					})

					It("returns no data set successfully if the data set cannot be found", func() {
						resultDataSet, err := mongoSession.GetDataSetByID(ctx, "not-found")
						Expect(err).ToNot(HaveOccurred())
						Expect(resultDataSet).To(BeNil())
					})
				})

				Context("CreateDataSet", func() {
					It("succeeds if it successfully creates the data set", func() {
						Expect(mongoSession.CreateDataSet(ctx, dataSet)).To(Succeed())
					})

					It("returns an error if the data set is missing", func() {
						Expect(mongoSession.CreateDataSet(ctx, nil)).To(MatchError("data set is missing"))
					})

					It("returns an error if the user id is missing", func() {
						dataSet.UserID = nil
						Expect(mongoSession.CreateDataSet(ctx, dataSet)).To(MatchError("data set user id is missing"))
					})

					It("returns an error if the user id is empty", func() {
						dataSet.UserID = pointer.FromString("")
						Expect(mongoSession.CreateDataSet(ctx, dataSet)).To(MatchError("data set user id is empty"))
					})

					It("returns an error if the upload id is missing", func() {
						dataSet.UploadID = nil
						Expect(mongoSession.CreateDataSet(ctx, dataSet)).To(MatchError("data set upload id is missing"))
					})

					It("returns an error if the upload id is empty", func() {
						dataSet.UploadID = pointer.FromString("")
						Expect(mongoSession.CreateDataSet(ctx, dataSet)).To(MatchError("data set upload id is empty"))
					})

					It("returns an error if the session is closed", func() {
						mongoSession.Close()
						Expect(mongoSession.CreateDataSet(ctx, dataSet)).To(MatchError("session closed"))
					})

					It("returns an error if the data set with the same id already exists", func() {
						Expect(mongoSession.CreateDataSet(ctx, dataSet)).To(Succeed())
						Expect(mongoSession.CreateDataSet(ctx, dataSet)).To(MatchError("unable to create data set; data set already exists"))
					})

					It("sets the created time", func() {
						Expect(mongoSession.CreateDataSet(ctx, dataSet)).To(Succeed())
						Expect(dataSet.CreatedTime).ToNot(BeNil())
						Expect(*dataSet.CreatedTime).ToNot(BeEmpty())
						Expect(dataSet.CreatedUserID).To(BeNil())
						Expect(dataSet.ByUser).To(BeNil())
					})

					It("has the correct stored data sets", func() {
						ValidateDataSet(testMongoCollection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, bson.M{}, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo)
						Expect(mongoSession.CreateDataSet(ctx, dataSet)).To(Succeed())
						ValidateDataSet(testMongoCollection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, bson.M{}, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo, dataSet)
					})
				})

				// Context("UpdateDataSet", func() {
				// 	BeforeEach(func() {
				// 		dataSet.CreatedTime = pointer.FromString("2016-09-01T11:00:00Z")
				// 		Expect(testMongoCollection.Insert(dataSet)).To(Succeed())
				// 	})

				// 	Context("with state closed", func() {
				// 		BeforeEach(func() {
				// 			dataSet.State = pointer.FromString("closed")
				// 		})

				// 		It("succeeds if it successfully updates the data set", func() {
				// 			Expect(mongoSession.UpdateDataSet(ctx, dataSet)).To(Succeed())
				// 		})

				// 		It("returns an error if the data set is missing", func() {
				// 			Expect(mongoSession.UpdateDataSet(ctx, nil)).To(MatchError("data set is missing"))
				// 		})

				// 		It("returns an error if the user id is missing", func() {
				// 			dataSet.UserID = ""
				// 			Expect(mongoSession.UpdateDataSet(ctx, dataSet)).To(MatchError("data set user id is missing"))
				// 		})

				// 		It("returns an error if the upload id is missing", func() {
				// 			dataSet.UploadID = ""
				// 			Expect(mongoSession.UpdateDataSet(ctx, dataSet)).To(MatchError("data set upload id is missing"))
				// 		})

				// 		It("returns an error if the device id is missing (nil)", func() {
				// 			dataSet.DeviceID = nil
				// 			Expect(mongoSession.UpdateDataSet(ctx, dataSet)).To(MatchError("data set device id is missing"))
				// 		})

				// 		It("returns an error if the device id is missing (empty)", func() {
				// 			dataSet.DeviceID = pointer.FromString("")
				// 			Expect(mongoSession.UpdateDataSet(ctx, dataSet)).To(MatchError("data set device id is missing"))
				// 		})

				// 		It("returns an error if the session is closed", func() {
				// 			mongoSession.Close()
				// 			Expect(mongoSession.UpdateDataSet(ctx, dataSet)).To(MatchError("session closed"))
				// 		})

				// 		It("returns an error if the data set with the same user id and upload id does not yet exist", func() {
				// 			dataSet.UploadID = pointer.FromString(dataTest.RandomSetID())
				// 			Expect(mongoSession.UpdateDataSet(ctx, dataSet)).To(MatchError("unable to update data set; not found"))
				// 		})
				// 	})

				// 	It("sets the modified time", func() {
				// 		dataSet.State = "closed"
				// 		Expect(mongoSession.UpdateDataSet(ctx, dataSet)).To(Succeed())
				// 		Expect(dataSet.ModifiedTime).ToNot(BeEmpty())
				// 		Expect(dataSet.ModifiedUserID).To(BeEmpty())
				// 	})

				// 	It("has the correct stored data sets", func() {
				// 		ValidateDataSet(testMongoCollection, bson.M{}, bson.M{}, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo, dataSet)
				// 		ValidateDataSet(testMongoCollection, bson.M{"modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, bson.M{})
				// 		dataSet.State = "closed"
				// 		Expect(mongoSession.UpdateDataSet(ctx, dataSet)).To(Succeed())
				// 		ValidateDataSet(testMongoCollection, bson.M{}, bson.M{}, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo, dataSet)
				// 		ValidateDataSet(testMongoCollection, bson.M{"modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, bson.M{}, dataSet)
				// 	})
				// })

				Context("with data", func() {
					var dataSetExistingOtherData []data.Datum
					var dataSetExistingOneData []data.Datum
					var dataSetExistingTwoData []data.Datum
					var dataSetData []data.Datum

					BeforeEach(func() {
						dataSet.CreatedTime = pointer.FromString("2016-09-01T11:00:00Z")
						Expect(testMongoCollection.Insert(dataSet)).To(Succeed())
						dataSetExistingOtherData = NewDataSetData(dataTest.NewDeviceID())
						Expect(mongoSession.CreateDataSetData(ctx, dataSetExistingOther, dataSetExistingOtherData)).To(Succeed())
						dataSetExistingOneData = NewDataSetData(deviceID)
						Expect(mongoSession.CreateDataSetData(ctx, dataSetExistingOne, dataSetExistingOneData)).To(Succeed())
						dataSetExistingTwoData = NewDataSetData(deviceID)
						Expect(mongoSession.CreateDataSetData(ctx, dataSetExistingTwo, dataSetExistingTwoData)).To(Succeed())
						dataSetData = NewDataSetData(deviceID)
					})

					Context("DeleteDataSet", func() {
						BeforeEach(func() {
							Expect(mongoSession.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
						})

						It("succeeds if it successfully deletes the data set", func() {
							Expect(mongoSession.DeleteDataSet(ctx, dataSet)).To(Succeed())
						})

						It("returns an error if the data set is missing", func() {
							Expect(mongoSession.DeleteDataSet(ctx, nil)).To(MatchError("data set is missing"))
						})

						It("returns an error if the user id is missing", func() {
							dataSet.UserID = nil
							Expect(mongoSession.DeleteDataSet(ctx, dataSet)).To(MatchError("data set user id is missing"))
						})

						It("returns an error if the user id is empty", func() {
							dataSet.UserID = pointer.FromString("")
							Expect(mongoSession.DeleteDataSet(ctx, dataSet)).To(MatchError("data set user id is empty"))
						})

						It("returns an error if the upload id is missing", func() {
							dataSet.UploadID = nil
							Expect(mongoSession.DeleteDataSet(ctx, dataSet)).To(MatchError("data set upload id is missing"))
						})

						It("returns an error if the upload id is empty", func() {
							dataSet.UploadID = pointer.FromString("")
							Expect(mongoSession.DeleteDataSet(ctx, dataSet)).To(MatchError("data set upload id is empty"))
						})

						It("returns an error if the session is closed", func() {
							mongoSession.Close()
							Expect(mongoSession.DeleteDataSet(ctx, dataSet)).To(MatchError("session closed"))
						})

						It("sets the deleted time on the data set", func() {
							Expect(mongoSession.DeleteDataSet(ctx, dataSet)).To(Succeed())
							Expect(dataSet.DeletedTime).ToNot(BeNil())
							Expect(*dataSet.DeletedTime).ToNot(BeEmpty())
							Expect(dataSet.DeletedUserID).To(BeNil())
						})

						It("has the correct stored data sets", func() {
							ValidateDataSet(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": true}, "deletedUserId": bson.M{"$exists": false}}, bson.M{})
							Expect(mongoSession.DeleteDataSet(ctx, dataSet)).To(Succeed())
							ValidateDataSet(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": true}, "deletedUserId": bson.M{"$exists": false}}, bson.M{}, dataSet)
						})

						It("has the correct stored data set data", func() {
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSetData)
							Expect(mongoSession.DeleteDataSet(ctx, dataSet)).To(Succeed())
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, []data.Datum{})
						})
					})

					Context("CreateDataSetData", func() {
						It("succeeds if data set data is empty", func() {
							dataSetData = []data.Datum{}
							Expect(mongoSession.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
						})

						It("succeeds if it successfully creates the data set data", func() {
							Expect(mongoSession.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
						})

						It("returns an error if the data set is missing", func() {
							Expect(mongoSession.CreateDataSetData(ctx, nil, dataSetData)).To(MatchError("data set is missing"))
						})

						It("returns an error if the data set data is missing", func() {
							Expect(mongoSession.CreateDataSetData(ctx, dataSet, nil)).To(MatchError("data set data is missing"))
						})

						It("returns an error if the user id is missing", func() {
							dataSet.UserID = nil
							Expect(mongoSession.CreateDataSetData(ctx, dataSet, dataSetData)).To(MatchError("data set user id is missing"))
						})

						It("returns an error if the user id is empty", func() {
							dataSet.UserID = pointer.FromString("")
							Expect(mongoSession.CreateDataSetData(ctx, dataSet, dataSetData)).To(MatchError("data set user id is empty"))
						})

						It("returns an error if the upload id is missing", func() {
							dataSet.UploadID = nil
							Expect(mongoSession.CreateDataSetData(ctx, dataSet, dataSetData)).To(MatchError("data set upload id is missing"))
						})

						It("returns an error if the upload id is empty", func() {
							dataSet.UploadID = pointer.FromString("")
							Expect(mongoSession.CreateDataSetData(ctx, dataSet, dataSetData)).To(MatchError("data set upload id is empty"))
						})

						It("returns an error if the session is closed", func() {
							mongoSession.Close()
							Expect(mongoSession.CreateDataSetData(ctx, dataSet, dataSetData)).To(MatchError("session closed"))
						})

						It("sets the user id and upload id on the data set data to match the data set", func() {
							Expect(mongoSession.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
							for _, dataSetDatum := range dataSetData {
								baseDatum, ok := dataSetDatum.(*types.Base)
								Expect(ok).To(BeTrue())
								Expect(baseDatum).ToNot(BeNil())
								Expect(baseDatum.UserID).To(Equal(dataSet.UserID))
								Expect(baseDatum.UploadID).To(Equal(dataSet.UploadID))
							}
						})

						It("leaves the data set data not active", func() {
							Expect(mongoSession.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
							for _, dataSetDatum := range dataSetData {
								baseDatum, ok := dataSetDatum.(*types.Base)
								Expect(ok).To(BeTrue())
								Expect(baseDatum).ToNot(BeNil())
								Expect(baseDatum.Active).To(BeFalse())
							}
						})

						It("sets the created time on the data set data", func() {
							Expect(mongoSession.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
							for _, dataSetDatum := range dataSetData {
								baseDatum, ok := dataSetDatum.(*types.Base)
								Expect(ok).To(BeTrue())
								Expect(baseDatum).ToNot(BeNil())
								Expect(baseDatum.CreatedTime).ToNot(BeNil())
								Expect(*baseDatum.CreatedTime).ToNot(BeEmpty())
								Expect(baseDatum.CreatedUserID).To(BeNil())
							}
						})

						It("has the correct stored data set data", func() {
							dataSetBeforeCreateData := append(append(dataSetExistingOtherData, dataSetExistingOneData...), dataSetExistingTwoData...)
							ValidateDataSetData(testMongoCollection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, bson.M{}, dataSetBeforeCreateData)
							Expect(mongoSession.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
							ValidateDataSetData(testMongoCollection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, bson.M{}, append(dataSetBeforeCreateData, dataSetData...))
						})
					})

					Context("DeleteDataSetData", func() {
						var deletes *data.Deletes

						BeforeEach(func() {
							Expect(mongoSession.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
							deletes = &data.Deletes{}
						})

						var deleteDataSetDataAssertions = func() {
							It("succeeds if it successfully deletes the data set", func() {
								Expect(mongoSession.DeleteDataSetData(ctx, dataSet, deletes)).To(Succeed())
							})

							It("returns an error if the context is missing", func() {
								ctx = nil
								Expect(mongoSession.DeleteDataSetData(ctx, dataSet, deletes)).To(MatchError("context is missing"))
							})

							It("returns an error if the data set is missing", func() {
								dataSet = nil
								Expect(mongoSession.DeleteDataSetData(ctx, dataSet, deletes)).To(MatchError("data set is missing"))
							})

							It("returns an error if the user id is missing", func() {
								dataSet.UserID = nil
								Expect(mongoSession.DeleteDataSetData(ctx, dataSet, deletes)).To(MatchError("data set user id is missing"))
							})

							It("returns an error if the user id is empty", func() {
								dataSet.UserID = pointer.FromString("")
								Expect(mongoSession.DeleteDataSetData(ctx, dataSet, deletes)).To(MatchError("data set user id is empty"))
							})

							It("returns an error if the upload id is missing", func() {
								dataSet.UploadID = nil
								Expect(mongoSession.DeleteDataSetData(ctx, dataSet, deletes)).To(MatchError("data set upload id is missing"))
							})

							It("returns an error if the upload id is empty", func() {
								dataSet.UploadID = pointer.FromString("")
								Expect(mongoSession.DeleteDataSetData(ctx, dataSet, deletes)).To(MatchError("data set upload id is empty"))
							})

							It("returns an error if the deletes is missing", func() {
								deletes = nil
								Expect(mongoSession.DeleteDataSetData(ctx, dataSet, deletes)).To(MatchError("deletes is missing"))
							})

							It("returns an error if the deletes is invalid", func() {
								(*deletes)[0].ID = pointer.FromString("")
								Expect(mongoSession.DeleteDataSetData(ctx, dataSet, deletes)).To(MatchError("deletes is invalid; value is empty"))
							})

							It("returns an error if the session is closed", func() {
								mongoSession.Close()
								Expect(mongoSession.DeleteDataSetData(ctx, dataSet, deletes)).To(MatchError("session closed"))
							})

							It("has the correct stored data sets", func() {
								ValidateDataSet(testMongoCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSet)
								Expect(mongoSession.DeleteDataSetData(ctx, dataSet, deletes)).To(Succeed())
								ValidateDataSet(testMongoCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSet)
							})

							It("has the correct stored data set data", func() {
								ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSetData)
								Expect(mongoSession.DeleteDataSetData(ctx, dataSet, deletes)).To(Succeed())
								ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, []data.Datum{})
							})
						}

						Context("by id", func() {
							BeforeEach(func() {
								for _, datum := range dataSetData {
									base, _ := datum.(*types.Base)
									*deletes = append(*deletes, &data.Delete{ID: base.ID})
								}
							})

							deleteDataSetDataAssertions()
						})

						Context("by origin id", func() {
							BeforeEach(func() {
								for _, datum := range dataSetData {
									base, _ := datum.(*types.Base)
									*deletes = append(*deletes, &data.Delete{Origin: &data.DeleteOrigin{ID: base.Origin.ID}})
								}
							})

							deleteDataSetDataAssertions()
						})

						Context("by both id and origin id", func() {
							BeforeEach(func() {
								for _, datum := range dataSetData {
									base, _ := datum.(*types.Base)
									*deletes = append(*deletes, &data.Delete{ID: base.ID, Origin: &data.DeleteOrigin{ID: base.Origin.ID}})
								}
							})

							deleteDataSetDataAssertions()
						})

						Context("with neither id nor origin id it deletes nothing", func() {
							BeforeEach(func() {
								for range dataSetData {
									*deletes = append(*deletes, &data.Delete{})
								}
							})

							It("has the correct stored data sets", func() {
								ValidateDataSet(testMongoCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSet)
								Expect(mongoSession.DeleteDataSetData(ctx, dataSet, deletes)).To(Succeed())
								ValidateDataSet(testMongoCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSet)
							})

							It("has the correct stored data set data", func() {
								ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSetData)
								Expect(mongoSession.DeleteDataSetData(ctx, dataSet, deletes)).To(Succeed())
								ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSetData)
							})
						})
					})

					Context("ActivateDataSetData", func() {
						BeforeEach(func() {
							Expect(mongoSession.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
						})

						It("succeeds if it successfully activates the data set", func() {
							Expect(mongoSession.ActivateDataSetData(ctx, dataSet)).To(Succeed())
						})

						It("returns an error if the data set is missing", func() {
							Expect(mongoSession.ActivateDataSetData(ctx, nil)).To(MatchError("data set is missing"))
						})

						It("returns an error if the user id is missing", func() {
							dataSet.UserID = nil
							Expect(mongoSession.ActivateDataSetData(ctx, dataSet)).To(MatchError("data set user id is missing"))
						})

						It("returns an error if the user id is empty", func() {
							dataSet.UserID = pointer.FromString("")
							Expect(mongoSession.ActivateDataSetData(ctx, dataSet)).To(MatchError("data set user id is empty"))
						})

						It("returns an error if the upload id is missing", func() {
							dataSet.UploadID = nil
							Expect(mongoSession.ActivateDataSetData(ctx, dataSet)).To(MatchError("data set upload id is missing"))
						})

						It("returns an error if the upload id is empty", func() {
							dataSet.UploadID = pointer.FromString("")
							Expect(mongoSession.ActivateDataSetData(ctx, dataSet)).To(MatchError("data set upload id is empty"))
						})

						It("returns an error if the session is closed", func() {
							mongoSession.Close()
							Expect(mongoSession.ActivateDataSetData(ctx, dataSet)).To(MatchError("session closed"))
						})

						It("sets the active on the data set", func() {
							Expect(mongoSession.ActivateDataSetData(ctx, dataSet)).To(Succeed())
							Expect(dataSet.Active).To(BeTrue())
						})

						It("sets the modified time on the data set", func() {
							Expect(mongoSession.ActivateDataSetData(ctx, dataSet)).To(Succeed())
							Expect(dataSet.ModifiedTime).ToNot(BeNil())
							Expect(*dataSet.ModifiedTime).ToNot(BeEmpty())
							Expect(dataSet.ModifiedUserID).To(BeNil())
						})

						It("has the correct stored active data set", func() {
							ValidateDataSet(testMongoCollection, bson.M{"_active": true, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, bson.M{})
							Expect(mongoSession.ActivateDataSetData(ctx, dataSet)).To(Succeed())
							ValidateDataSet(testMongoCollection, bson.M{"_active": true, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, bson.M{}, dataSet)
						})

						It("has the correct stored active data set data", func() {
							ValidateDataSetData(testMongoCollection, bson.M{"_active": true, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, bson.M{}, []data.Datum{})
							Expect(mongoSession.ActivateDataSetData(ctx, dataSet)).To(Succeed())
							for _, dataSetDatum := range dataSetData {
								dataSetDatum.SetActive(true)
							}
							ValidateDataSetData(testMongoCollection, bson.M{"_active": true, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, bson.M{"modifiedTime": 0}, dataSetData)
						})
					})

					Context("ArchiveDeviceDataUsingHashesFromDataSet", func() {
						var dataSetExistingOneDataCloned []data.Datum

						BeforeEach(func() {
							dataSetExistingOneDataCloned = CloneDataSetData(dataSetData)
							Expect(mongoSession.CreateDataSetData(ctx, dataSetExistingOne, dataSetExistingOneDataCloned)).To(Succeed())
							Expect(mongoSession.ActivateDataSetData(ctx, dataSetExistingOne)).To(Succeed())
							Expect(mongoSession.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
							for _, dataSetDatum := range append(dataSetExistingOneData, dataSetExistingOneDataCloned...) {
								dataSetDatum.SetActive(true)
							}
						})

						It("succeeds if it successfully archives device data using hashes from data set", func() {
							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(Succeed())
						})

						It("returns an error if the data set is missing", func() {
							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataSet(ctx, nil)).To(MatchError("data set is missing"))
						})

						It("returns an error if the user id is missing", func() {
							dataSet.UserID = nil
							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("data set user id is missing"))
						})

						It("returns an error if the user id is empty", func() {
							dataSet.UserID = pointer.FromString("")
							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("data set user id is empty"))
						})

						It("returns an error if the upload id is missing", func() {
							dataSet.UploadID = nil
							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("data set upload id is missing"))
						})

						It("returns an error if the upload id is empty", func() {
							dataSet.UploadID = pointer.FromString("")
							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("data set upload id is empty"))
						})

						It("returns an error if the device id is missing (nil)", func() {
							dataSet.DeviceID = nil
							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("data set device id is missing"))
						})

						It("returns an error if the device id is missing (empty)", func() {
							dataSet.DeviceID = pointer.FromString("")
							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("data set device id is missing"))
						})

						It("returns an error if the session is closed", func() {
							mongoSession.Close()
							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("session closed"))
						})

						It("has the correct stored data sets", func() {
							ValidateDataSet(testMongoCollection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true}, bson.M{}, dataSetExistingOne)
							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(Succeed())
							ValidateDataSet(testMongoCollection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true}, bson.M{}, dataSetExistingOne)
						})

						It("has the correct stored archived data set data", func() {
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": false}, bson.M{}, []data.Datum{})
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true}, bson.M{"modifiedTime": 0}, append(dataSetExistingOneData, dataSetExistingOneDataCloned...))
							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(Succeed())
							for _, dataSetDatum := range dataSetExistingOneDataCloned {
								dataSetDatum.SetActive(false)
							}
							ValidateDataSetData(testMongoCollection,
								bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}},
								bson.M{"modifiedTime": 0},
								dataSetExistingOneData)
							ValidateDataSetData(testMongoCollection,
								bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": false, "archivedTime": bson.M{"$exists": true}, "archivedDatasetId": dataSet.UploadID, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}},
								bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0},
								dataSetExistingOneDataCloned)
							ValidateDataSetData(testMongoCollection,
								bson.M{"uploadId": dataSet.UploadID, "_active": false, "archivedTime": bson.M{"$exists": false}, "archivedDatasetId": bson.M{"$exists": false}},
								bson.M{},
								dataSetData)
						})
					})

					Context("UnarchiveDeviceDataUsingHashesFromDataSet", func() {
						var dataSetExistingTwoDataCloned []data.Datum
						var dataSetExistingOneDataCloned []data.Datum

						BeforeEach(func() {
							dataSetExistingTwoDataCloned = CloneDataSetData(dataSetData)
							dataSetExistingOneDataCloned = CloneDataSetData(dataSetData)
							Expect(mongoSession.CreateDataSetData(ctx, dataSetExistingTwo, dataSetExistingTwoDataCloned)).To(Succeed())
							Expect(mongoSession.ActivateDataSetData(ctx, dataSetExistingTwo)).To(Succeed())
							Expect(mongoSession.CreateDataSetData(ctx, dataSetExistingOne, dataSetExistingOneDataCloned)).To(Succeed())
							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSetExistingOne)).To(Succeed())
							Expect(mongoSession.ActivateDataSetData(ctx, dataSetExistingOne)).To(Succeed())
							Expect(mongoSession.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(Succeed())
							Expect(mongoSession.ActivateDataSetData(ctx, dataSet)).To(Succeed())
						})

						It("succeeds if it successfully unarchives device data using hashes from data set", func() {
							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(Succeed())
						})

						It("returns an error if the data set is missing", func() {
							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, nil)).To(MatchError("data set is missing"))
						})

						It("returns an error if the user id is missing", func() {
							dataSet.UserID = nil
							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("data set user id is missing"))
						})

						It("returns an error if the user id is empty", func() {
							dataSet.UserID = pointer.FromString("")
							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("data set user id is empty"))
						})

						It("returns an error if the upload id is missing", func() {
							dataSet.UploadID = nil
							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("data set upload id is missing"))
						})

						It("returns an error if the upload id is empty", func() {
							dataSet.UploadID = pointer.FromString("")
							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("data set upload id is empty"))
						})

						It("returns an error if the device id is missing (nil)", func() {
							dataSet.DeviceID = nil
							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("data set device id is missing"))
						})

						It("returns an error if the device id is missing (empty)", func() {
							dataSet.DeviceID = pointer.FromString("")
							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("data set device id is missing"))
						})

						It("returns an error if the session is closed", func() {
							mongoSession.Close()
							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("session closed"))
						})

						It("has the correct stored data sets", func() {
							ValidateDataSet(testMongoCollection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true}, bson.M{}, dataSetExistingOne)
							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(Succeed())
							ValidateDataSet(testMongoCollection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true}, bson.M{}, dataSetExistingOne)
						})

						It("has the correct stored unarchived data set data", func() {
							for _, dataSetDatum := range append(dataSetData, dataSetExistingOneData...) {
								dataSetDatum.SetActive(true)
							}
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": false}, bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0}, dataSetExistingOneDataCloned)
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true}, bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0}, dataSetExistingOneData)
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID, "_active": true}, bson.M{"modifiedTime": 0}, dataSetData)
							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(Succeed())
							for _, dataSetDatum := range dataSetExistingOneDataCloned {
								dataSetDatum.SetActive(true)
							}
							ValidateDataSetData(testMongoCollection,
								bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true, "archivedTime": bson.M{"$exists": false}, "archivedDatasetId": bson.M{"$exists": false}, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}},
								bson.M{"modifiedTime": 0},
								append(dataSetExistingOneData, dataSetExistingOneDataCloned...))
							ValidateDataSetData(testMongoCollection,
								bson.M{"uploadId": dataSet.UploadID, "_active": true, "archivedTime": bson.M{"$exists": false}, "archivedDatasetId": bson.M{"$exists": false}},
								bson.M{"modifiedTime": 0},
								dataSetData)
						})

						It("has the correct stored data sets if an intermediary is unarchived", func() {
							ValidateDataSet(testMongoCollection, bson.M{"uploadId": dataSetExistingTwo.UploadID, "_active": true}, bson.M{}, dataSetExistingTwo)
							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSetExistingOne)).To(Succeed())
							ValidateDataSet(testMongoCollection, bson.M{"uploadId": dataSetExistingTwo.UploadID, "_active": true}, bson.M{}, dataSetExistingTwo)
						})

						It("has the correct stored unarchived data set data if an intermediary is unarchived", func() {
							for _, dataSetDatum := range append(dataSetExistingOneData, dataSetExistingTwoData...) {
								dataSetDatum.SetActive(true)
							}
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSetExistingTwo.UploadID, "_active": false}, bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0}, dataSetExistingTwoDataCloned)
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSetExistingTwo.UploadID, "_active": true}, bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0}, dataSetExistingTwoData)
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true}, bson.M{"modifiedTime": 0}, dataSetExistingOneData)
							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSetExistingOne)).To(Succeed())
							ValidateDataSetData(testMongoCollection,
								bson.M{"uploadId": dataSetExistingTwo.UploadID, "_active": true, "archivedTime": bson.M{"$exists": false}, "archivedDatasetId": bson.M{"$exists": false}},
								bson.M{"modifiedTime": 0},
								dataSetExistingTwoData)
							ValidateDataSetData(testMongoCollection,
								bson.M{"uploadId": dataSetExistingTwo.UploadID, "_active": false, "archivedTime": bson.M{"$exists": true}, "archivedDatasetId": dataSet.UploadID, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}},
								bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0},
								dataSetExistingTwoDataCloned)
							ValidateDataSetData(testMongoCollection,
								bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true, "archivedTime": bson.M{"$exists": false}, "archivedDatasetId": bson.M{"$exists": false}},
								bson.M{"modifiedTime": 0},
								dataSetExistingOneData)
							ValidateDataSetData(testMongoCollection,
								bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": false, "archivedTime": bson.M{"$exists": true}, "archivedDatasetId": dataSet.UploadID},
								bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0},
								dataSetExistingOneDataCloned)
						})
					})

					Context("ArchiveDataSetDataUsingOriginIDs", func() {
						var originIDs []string

						BeforeEach(func() {
							Expect(mongoSession.CreateDataSetData(ctx, dataSet, dataSetExistingOneData)).To(Succeed())
							Expect(mongoSession.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
							Expect(mongoSession.ActivateDataSetData(ctx, dataSet)).To(Succeed())
							originIDs = []string{}
							for _, datum := range dataSetData {
								baseDatum := datum.(*types.Base)
								if baseDatum.Origin != nil && baseDatum.Origin.ID != nil {
									originIDs = append(originIDs, *baseDatum.Origin.ID)
								}
							}
							for _, datum := range append(dataSetData, dataSetExistingOneData...) {
								datum.SetActive(true)
							}
						})

						It("returns an error if the data set is missing", func() {
							Expect(mongoSession.ArchiveDataSetDataUsingOriginIDs(nil, dataSet, originIDs)).To(MatchError("context is missing"))
						})

						It("returns an error if the data set is missing", func() {
							Expect(mongoSession.ArchiveDataSetDataUsingOriginIDs(ctx, nil, originIDs)).To(MatchError("data set is missing"))
						})

						It("returns an error if the user id is missing", func() {
							dataSet.UserID = nil
							Expect(mongoSession.ArchiveDataSetDataUsingOriginIDs(ctx, dataSet, originIDs)).To(MatchError("data set user id is missing"))
						})

						It("returns an error if the user id is empty", func() {
							dataSet.UserID = pointer.FromString("")
							Expect(mongoSession.ArchiveDataSetDataUsingOriginIDs(ctx, dataSet, originIDs)).To(MatchError("data set user id is empty"))
						})

						It("returns an error if the upload id is missing", func() {
							dataSet.UploadID = nil
							Expect(mongoSession.ArchiveDataSetDataUsingOriginIDs(ctx, dataSet, originIDs)).To(MatchError("data set upload id is missing"))
						})

						It("returns an error if the upload id is empty", func() {
							dataSet.UploadID = pointer.FromString("")
							Expect(mongoSession.ArchiveDataSetDataUsingOriginIDs(ctx, dataSet, originIDs)).To(MatchError("data set upload id is empty"))
						})

						It("returns an error if the session is closed", func() {
							mongoSession.Close()
							Expect(mongoSession.ArchiveDataSetDataUsingOriginIDs(ctx, dataSet, originIDs)).To(MatchError("session closed"))
						})

						It("archives all datum in the data set matching the origin id", func() {
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID, "_active": false}, bson.M{"modifiedTime": 0}, []data.Datum{})
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID, "_active": true}, bson.M{"modifiedTime": 0}, append(dataSetData, dataSetExistingOneData...))
							Expect(mongoSession.ArchiveDataSetDataUsingOriginIDs(ctx, dataSet, originIDs)).To(Succeed())
							for _, datum := range dataSetData {
								datum.SetActive(false)
							}
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID, "_active": false}, bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0}, dataSetData)
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID, "_active": true}, bson.M{"modifiedTime": 0}, dataSetExistingOneData)
						})

						It("does nothing if the origin ids is missing", func() {
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID, "_active": false}, bson.M{"modifiedTime": 0}, []data.Datum{})
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID, "_active": true}, bson.M{"modifiedTime": 0}, append(dataSetData, dataSetExistingOneData...))
							Expect(mongoSession.ArchiveDataSetDataUsingOriginIDs(ctx, dataSet, nil)).To(Succeed())
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID, "_active": false}, bson.M{"modifiedTime": 0}, []data.Datum{})
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID, "_active": true}, bson.M{"modifiedTime": 0}, append(dataSetData, dataSetExistingOneData...))
						})

						It("does nothing if the origin ids is empty", func() {
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID, "_active": false}, bson.M{"modifiedTime": 0}, []data.Datum{})
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID, "_active": true}, bson.M{"modifiedTime": 0}, append(dataSetData, dataSetExistingOneData...))
							Expect(mongoSession.ArchiveDataSetDataUsingOriginIDs(ctx, dataSet, []string{})).To(Succeed())
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID, "_active": false}, bson.M{"modifiedTime": 0}, []data.Datum{})
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID, "_active": true}, bson.M{"modifiedTime": 0}, append(dataSetData, dataSetExistingOneData...))
						})
					})

					Context("DeleteArchivedDataSetData", func() {
						BeforeEach(func() {
							Expect(mongoSession.CreateDataSetData(ctx, dataSet, dataSetExistingOneData)).To(Succeed())
							Expect(mongoSession.ActivateDataSetData(ctx, dataSet)).To(Succeed())
							Expect(mongoSession.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
							for _, datum := range dataSetExistingOneData {
								datum.SetActive(true)
							}
						})

						It("returns an error if the data set is missing", func() {
							Expect(mongoSession.DeleteArchivedDataSetData(nil, dataSet)).To(MatchError("context is missing"))
						})

						It("returns an error if the data set is missing", func() {
							Expect(mongoSession.DeleteArchivedDataSetData(ctx, nil)).To(MatchError("data set is missing"))
						})

						It("returns an error if the user id is missing", func() {
							dataSet.UserID = nil
							Expect(mongoSession.DeleteArchivedDataSetData(ctx, dataSet)).To(MatchError("data set user id is missing"))
						})

						It("returns an error if the user id is empty", func() {
							dataSet.UserID = pointer.FromString("")
							Expect(mongoSession.DeleteArchivedDataSetData(ctx, dataSet)).To(MatchError("data set user id is empty"))
						})

						It("returns an error if the upload id is missing", func() {
							dataSet.UploadID = nil
							Expect(mongoSession.DeleteArchivedDataSetData(ctx, dataSet)).To(MatchError("data set upload id is missing"))
						})

						It("returns an error if the upload id is empty", func() {
							dataSet.UploadID = pointer.FromString("")
							Expect(mongoSession.DeleteArchivedDataSetData(ctx, dataSet)).To(MatchError("data set upload id is empty"))
						})

						It("returns an error if the session is closed", func() {
							mongoSession.Close()
							Expect(mongoSession.DeleteArchivedDataSetData(ctx, dataSet)).To(MatchError("session closed"))
						})

						It("deletes all of the archived data set data", func() {
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID, "_active": false}, bson.M{"modifiedTime": 0}, dataSetData)
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID, "_active": true}, bson.M{"modifiedTime": 0}, dataSetExistingOneData)
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{"modifiedTime": 0}, append(dataSetData, dataSetExistingOneData...))
							Expect(mongoSession.DeleteArchivedDataSetData(ctx, dataSet)).To(Succeed())
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID, "_active": false}, bson.M{"modifiedTime": 0}, []data.Datum{})
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID, "_active": true}, bson.M{"modifiedTime": 0}, dataSetExistingOneData)
							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{"modifiedTime": 0}, dataSetExistingOneData)
						})
					})

					Context("DeleteOtherDataSetData", func() {
						BeforeEach(func() {
							Expect(mongoSession.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
						})

						It("succeeds if it successfully deletes all other data set data", func() {
							Expect(mongoSession.DeleteOtherDataSetData(ctx, dataSet)).To(Succeed())
						})

						It("returns an error if the data set is missing", func() {
							Expect(mongoSession.DeleteOtherDataSetData(ctx, nil)).To(MatchError("data set is missing"))
						})

						It("returns an error if the user id is missing", func() {
							dataSet.UserID = nil
							Expect(mongoSession.DeleteOtherDataSetData(ctx, dataSet)).To(MatchError("data set user id is missing"))
						})

						It("returns an error if the user id is empty", func() {
							dataSet.UserID = pointer.FromString("")
							Expect(mongoSession.DeleteOtherDataSetData(ctx, dataSet)).To(MatchError("data set user id is empty"))
						})

						It("returns an error if the upload id is missing", func() {
							dataSet.UploadID = nil
							Expect(mongoSession.DeleteOtherDataSetData(ctx, dataSet)).To(MatchError("data set upload id is missing"))
						})

						It("returns an error if the upload id is empty", func() {
							dataSet.UploadID = pointer.FromString("")
							Expect(mongoSession.DeleteOtherDataSetData(ctx, dataSet)).To(MatchError("data set upload id is empty"))
						})

						It("returns an error if the device id is missing (nil)", func() {
							dataSet.DeviceID = nil
							Expect(mongoSession.DeleteOtherDataSetData(ctx, dataSet)).To(MatchError("data set device id is missing"))
						})

						It("returns an error if the device id is missing (empty)", func() {
							dataSet.DeviceID = pointer.FromString("")
							Expect(mongoSession.DeleteOtherDataSetData(ctx, dataSet)).To(MatchError("data set device id is missing"))
						})

						It("returns an error if the session is closed", func() {
							mongoSession.Close()
							Expect(mongoSession.DeleteOtherDataSetData(ctx, dataSet)).To(MatchError("session closed"))
						})

						It("has the correct stored active data set", func() {
							ValidateDataSet(testMongoCollection, bson.M{}, bson.M{}, dataSet, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo)
							ValidateDataSet(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": false}, "deletedUserId": bson.M{"$exists": false}}, bson.M{}, dataSet, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo)
							Expect(mongoSession.DeleteOtherDataSetData(ctx, dataSet)).To(Succeed())
							Expect(testMongoCollection.Find(bson.M{"type": "upload"}).Count()).To(Equal(4))
							ValidateDataSet(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": true}, "deletedUserId": bson.M{"$exists": false}}, bson.M{"deletedTime": 0}, dataSetExistingTwo, dataSetExistingOne)
							ValidateDataSet(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": false}, "deletedUserId": bson.M{"$exists": false}}, bson.M{}, dataSet, dataSetExistingOther)
						})

						It("has the correct stored active data set data", func() {
							dataSetDataAfterRemoveData := append(dataSetData, dataSetExistingOtherData...)
							dataSetDataBeforeRemoveData := append(append(dataSetDataAfterRemoveData, dataSetExistingOneData...), dataSetExistingTwoData...)
							ValidateDataSetData(testMongoCollection, bson.M{}, bson.M{}, dataSetDataBeforeRemoveData)
							Expect(mongoSession.DeleteOtherDataSetData(ctx, dataSet)).To(Succeed())
							ValidateDataSetData(testMongoCollection, bson.M{}, bson.M{"deletedTime": 0}, dataSetDataAfterRemoveData)
						})
					})

					Context("DestroyDataForUserByID", func() {
						var deleteUserID string
						var deleteDeviceID string
						var deleteDataSet *upload.Upload
						var deleteDataSetData []data.Datum

						BeforeEach(func() {
							Expect(mongoSession.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
							deleteUserID = userTest.RandomID()
							deleteDeviceID = dataTest.NewDeviceID()
							deleteDataSet = NewDataSet(deleteUserID, deleteDeviceID)
							deleteDataSet.CreatedTime = pointer.FromString("2016-09-01T11:00:00Z")
							Expect(testMongoCollection.Insert(deleteDataSet)).To(Succeed())
							deleteDataSetData = NewDataSetData(deleteDeviceID)
							Expect(mongoSession.CreateDataSetData(ctx, deleteDataSet, deleteDataSetData)).To(Succeed())
						})

						It("succeeds if it successfully destroys all data for user by id", func() {
							Expect(mongoSession.DestroyDataForUserByID(ctx, deleteUserID)).To(Succeed())
						})

						It("returns an error if the user id is missing", func() {
							Expect(mongoSession.DestroyDataForUserByID(ctx, "")).To(MatchError("user id is missing"))
						})

						It("returns an error if the session is closed", func() {
							mongoSession.Close()
							Expect(mongoSession.DestroyDataForUserByID(ctx, deleteUserID)).To(MatchError("session closed"))
						})

						It("has the correct stored data sets", func() {
							ValidateDataSet(testMongoCollection, bson.M{}, bson.M{}, dataSet, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo, deleteDataSet)
							Expect(mongoSession.DestroyDataForUserByID(ctx, deleteUserID)).To(Succeed())
							ValidateDataSet(testMongoCollection, bson.M{}, bson.M{}, dataSet, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo)
						})

						It("has the correct stored data set data", func() {
							dataSetDataAfterRemoveData := append(append(append(dataSetData, dataSetExistingOtherData...), dataSetExistingOneData...), dataSetExistingTwoData...)
							dataSetDataBeforeRemoveData := append(dataSetDataAfterRemoveData, deleteDataSetData...)
							ValidateDataSetData(testMongoCollection, bson.M{}, bson.M{}, dataSetDataBeforeRemoveData)
							Expect(mongoSession.DestroyDataForUserByID(ctx, deleteUserID)).To(Succeed())
							ValidateDataSetData(testMongoCollection, bson.M{}, bson.M{}, dataSetDataAfterRemoveData)
						})
					})
				})
			})
		})
	})
})
