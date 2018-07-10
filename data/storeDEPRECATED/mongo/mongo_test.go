package mongo_test

// import (
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"

// 	"sync"
// 	"time"

// 	mgo "gopkg.in/mgo.v2"
// 	"gopkg.in/mgo.v2/bson"

// 	"github.com/tidepool-org/platform/data"
// 	"github.com/tidepool-org/platform/data/storeDEPRECATED"
// 	"github.com/tidepool-org/platform/data/storeDEPRECATED/mongo"
// 	"github.com/tidepool-org/platform/data/types"
// 	"github.com/tidepool-org/platform/data/types/upload"
// 	"github.com/tidepool-org/platform/id"
// 	"github.com/tidepool-org/platform/log/null"
// 	"github.com/tidepool-org/platform/page"
// 	"github.com/tidepool-org/platform/pointer"
// 	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
// 	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
// )

// var _sampleTimeMutex sync.Mutex
// var _sampleTimeOnce sync.Once
// var _sampleTime time.Time

// func SampleTime() time.Time {
// 	_sampleTimeMutex.Lock()
// 	defer _sampleTimeMutex.Unlock()

// 	_sampleTimeOnce.Do(func() {
// 		_sampleTime, _ = time.Parse(time.RFC3339, "2016-08-30T23:59:50-07:00")
// 	})

// 	_sampleTime = _sampleTime.Add(time.Second)
// 	return _sampleTime
// }

// func NewDataSet(userID string, deviceID string) *upload.Upload {
// 	dataSet := upload.New()
// 	Expect(dataSet).ToNot(BeNil())

// 	dataSet.Deduplicator = &data.DeduplicatorDescriptor{Name: "test-deduplicator"}
// 	dataSet.UserID = userID

// 	dataSet.ClockDriftOffset = pointer.FromInt(0)
// 	dataSet.ConversionOffset = pointer.FromInt(0)
// 	dataSet.DeviceID = pointer.FromString(deviceID)
// 	dataSet.DeviceTime = pointer.FromString(SampleTime().Format("2006-01-02T15:04:05"))
// 	dataSet.Time = pointer.FromString(SampleTime().UTC().Format(time.RFC3339))
// 	dataSet.TimeZoneOffset = pointer.FromInt(-420)

// 	dataSet.ComputerTime = pointer.FromString(SampleTime().Format("2006-01-02T15:04:05"))
// 	dataSet.DeviceManufacturers = pointer.FromStringArray([]string{"Tesla"})
// 	dataSet.DeviceModel = pointer.FromString("1234")
// 	dataSet.DeviceSerialNumber = pointer.FromString("567890")
// 	dataSet.DeviceTags = pointer.FromStringArray([]string{upload.DeviceTagInsulinPump})
// 	dataSet.TimeProcessing = pointer.FromString(upload.TimeProcessingUTCBootstrapping)
// 	dataSet.TimeZoneName = pointer.FromString("US/Pacific")
// 	dataSet.Version = pointer.FromString("0.260.1")

// 	return dataSet
// }

// func NewDataSetData(deviceID string) []data.Datum {
// 	dataSetData := []data.Datum{}
// 	for count := 0; count < 3; count++ {
// 		baseDatum := &types.Base{}
// 		baseDatum.New()

// 		baseDatum.Deduplicator = &data.DeduplicatorDescriptor{Hash: id.New()}
// 		baseDatum.Type = "test"

// 		baseDatum.ClockDriftOffset = pointer.FromInt(0)
// 		baseDatum.ConversionOffset = pointer.FromInt(0)
// 		baseDatum.DeviceID = pointer.FromString(deviceID)
// 		baseDatum.DeviceTime = pointer.FromString(SampleTime().Format("2006-01-02T15:04:05"))
// 		baseDatum.Time = pointer.FromString(SampleTime().UTC().Format(time.RFC3339))
// 		baseDatum.TimeZoneOffset = pointer.FromInt(-420)

// 		dataSetData = append(dataSetData, baseDatum)
// 	}
// 	return dataSetData
// }

// func CloneDataSetData(dataSetData []data.Datum) []data.Datum {
// 	clonedDataSetData := []data.Datum{}
// 	for _, dataSetDatum := range dataSetData {
// 		if baseDatum, ok := dataSetDatum.(*types.Base); ok {
// 			clonedBaseDatum := &types.Base{}
// 			clonedBaseDatum.Active = baseDatum.Active
// 			clonedBaseDatum.ArchivedDataSetID = baseDatum.ArchivedDataSetID
// 			clonedBaseDatum.ArchivedTime = baseDatum.ArchivedTime
// 			clonedBaseDatum.CreatedTime = baseDatum.CreatedTime
// 			clonedBaseDatum.CreatedUserID = baseDatum.CreatedUserID
// 			clonedBaseDatum.Deduplicator = baseDatum.Deduplicator
// 			clonedBaseDatum.DeletedTime = baseDatum.DeletedTime
// 			clonedBaseDatum.DeletedUserID = baseDatum.DeletedUserID
// 			clonedBaseDatum.GUID = baseDatum.GUID
// 			clonedBaseDatum.ID = baseDatum.ID
// 			clonedBaseDatum.ModifiedTime = baseDatum.ModifiedTime
// 			clonedBaseDatum.ModifiedUserID = baseDatum.ModifiedUserID
// 			clonedBaseDatum.SchemaVersion = baseDatum.SchemaVersion
// 			clonedBaseDatum.Type = baseDatum.Type
// 			clonedBaseDatum.UploadID = baseDatum.UploadID
// 			clonedBaseDatum.UserID = baseDatum.UserID
// 			clonedBaseDatum.Version = baseDatum.Version
// 			clonedBaseDatum.Annotations = baseDatum.Annotations
// 			clonedBaseDatum.ClockDriftOffset = baseDatum.ClockDriftOffset
// 			clonedBaseDatum.ConversionOffset = baseDatum.ConversionOffset
// 			clonedBaseDatum.DeviceID = baseDatum.DeviceID
// 			clonedBaseDatum.DeviceTime = baseDatum.DeviceTime
// 			clonedBaseDatum.Payload = baseDatum.Payload
// 			clonedBaseDatum.Source = baseDatum.Source
// 			clonedBaseDatum.Time = baseDatum.Time
// 			clonedBaseDatum.TimeZoneOffset = baseDatum.TimeZoneOffset

// 			clonedDataSetData = append(clonedDataSetData, clonedBaseDatum)
// 		}
// 	}
// 	return clonedDataSetData
// }

// func ValidateDataSet(testMongoCollection *mgo.Collection, query bson.M, filter bson.M, expectedDataSets ...*upload.Upload) {
// 	query["type"] = "upload"
// 	filter["_id"] = 0
// 	var actualDataSets []*upload.Upload
// 	Expect(testMongoCollection.Find(query).Select(filter).All(&actualDataSets)).To(Succeed())
// 	Expect(actualDataSets).To(ConsistOf(DataSetsAsInterface(expectedDataSets)...))
// }

// func DataSetsAsInterface(dataSets []*upload.Upload) []interface{} {
// 	var dataSetsAsInterface []interface{}
// 	for _, dataSet := range dataSets {
// 		dataSetsAsInterface = append(dataSetsAsInterface, dataSet)
// 	}
// 	return dataSetsAsInterface
// }

// func ValidateDataSetData(testMongoCollection *mgo.Collection, query bson.M, filter bson.M, expectedDataSetData []data.Datum) {
// 	query["type"] = bson.M{"$ne": "upload"}
// 	filter["_id"] = 0
// 	var actualDataSetData []interface{}
// 	Expect(testMongoCollection.Find(query).Select(filter).All(&actualDataSetData)).To(Succeed())
// 	Expect(actualDataSetData).To(ConsistOf(DataSetDataAsInterface(expectedDataSetData)...))
// }

// func DataSetDataAsInterface(dataSetData []data.Datum) []interface{} {
// 	var dataSetDataAsInterface []interface{}
// 	for _, dataSetDatum := range dataSetData {
// 		dataSetDataAsInterface = append(dataSetDataAsInterface, DataSetDatumAsInterface(dataSetDatum))
// 	}
// 	return dataSetDataAsInterface
// }

// func DataSetDatumAsInterface(dataSetDatum data.Datum) interface{} {
// 	bytes, err := bson.Marshal(dataSetDatum)
// 	Expect(err).ToNot(HaveOccurred())
// 	Expect(bytes).ToNot(BeNil())
// 	var dataSetDatumAsInterface interface{}
// 	Expect(bson.Unmarshal(bytes, &dataSetDatumAsInterface)).To(Succeed())
// 	return dataSetDatumAsInterface
// }

// var _ = Describe("Mongo", func() {
// 	var mongoConfig *storeStructuredMongo.Config
// 	var mongoStore *mongo.Store
// 	var mongoSession storeDEPRECATED.DataSession

// 	BeforeEach(func() {
// 		mongoConfig = storeStructuredMongoTest.NewConfig()
// 	})

// 	AfterEach(func() {
// 		if mongoSession != nil {
// 			mongoSession.Close()
// 		}
// 		if mongoStore != nil {
// 			mongoStore.Close()
// 		}
// 	})

// 	Context("New", func() {
// 		It("returns an error if unsuccessful", func() {
// 			var err error
// 			mongoStore, err = mongo.NewStore(nil, nil)
// 			Expect(err).To(HaveOccurred())
// 			Expect(mongoStore).To(BeNil())
// 		})

// 		It("returns a new store and no error if successful", func() {
// 			var err error
// 			mongoStore, err = mongo.NewStore(mongoConfig, null.NewLogger())
// 			Expect(err).ToNot(HaveOccurred())
// 			Expect(mongoStore).ToNot(BeNil())
// 		})
// 	})

// 	Context("with a new store", func() {
// 		BeforeEach(func() {
// 			var err error
// 			mongoStore, err = mongo.NewStore(mongoConfig, null.NewLogger())
// 			Expect(err).ToNot(HaveOccurred())
// 			Expect(mongoStore).ToNot(BeNil())
// 		})

// 		Context("NewDataSession", func() {
// 			It("returns a new session", func() {
// 				mongoSession = mongoStore.NewDataSession()
// 				Expect(mongoSession).ToNot(BeNil())
// 			})
// 		})

// 		Context("with a new session", func() {
// 			BeforeEach(func() {
// 				mongoSession = mongoStore.NewDataSession()
// 				Expect(mongoSession).ToNot(BeNil())
// 			})

// 			Context("with persisted data", func() {
// 				var testMongoSession *mgo.Session
// 				var testMongoCollection *mgo.Collection
// 				var userID string
// 				var deviceID string
// 				var dataSetExistingOther *upload.Upload
// 				var dataSetExistingOne *upload.Upload
// 				var dataSetExistingTwo *upload.Upload
// 				var dataSet *upload.Upload

// 				BeforeEach(func() {
// 					testMongoSession = storeStructuredMongoTest.Session().Copy()
// 					testMongoCollection = testMongoSession.DB(mongoConfig.Database).C(mongoConfig.CollectionPrefix + "deviceData")
// 					userID = id.New()
// 					deviceID = id.New()
// 					dataSetExistingOther = NewDataSet(id.New(), id.New())
// 					dataSetExistingOther.CreatedTime = "2016-09-01T12:00:00Z"
// 					Expect(testMongoCollection.Insert(dataSetExistingOther)).To(Succeed())
// 					dataSetExistingOne = NewDataSet(userID, deviceID)
// 					dataSetExistingOne.CreatedTime = "2016-09-01T12:30:00Z"
// 					Expect(testMongoCollection.Insert(dataSetExistingOne)).To(Succeed())
// 					dataSetExistingTwo = NewDataSet(userID, deviceID)
// 					dataSetExistingTwo.CreatedTime = "2016-09-01T10:00:00Z"
// 					Expect(testMongoCollection.Insert(dataSetExistingTwo)).To(Succeed())
// 					dataSet = NewDataSet(userID, deviceID)
// 				})

// 				AfterEach(func() {
// 					if testMongoSession != nil {
// 						testMongoSession.Close()
// 					}
// 				})

// 				Context("GetDataSetsForUserByID", func() {
// 					var filter *storeDEPRECATED.Filter
// 					var pagination *page.Pagination

// 					BeforeEach(func() {
// 						dataSet.CreatedTime = "2016-09-01T11:00:00Z"
// 						Expect(testMongoCollection.Insert(dataSet)).To(Succeed())
// 						filter = storeDEPRECATED.NewFilter()
// 						pagination = page.NewPagination()
// 					})

// 					It("succeeds if it successfully finds the user data sets", func() {
// 						Expect(mongoSession.GetDataSetsForUserByID(userID, filter, pagination)).To(ConsistOf([]*upload.Upload{dataSetExistingOne, dataSet, dataSetExistingTwo}))
// 					})

// 					It("succeeds if the filter is not specified", func() {
// 						Expect(mongoSession.GetDataSetsForUserByID(userID, nil, pagination)).To(ConsistOf([]*upload.Upload{dataSetExistingOne, dataSet, dataSetExistingTwo}))
// 					})

// 					It("succeeds if the pagination is not specified", func() {
// 						Expect(mongoSession.GetDataSetsForUserByID(userID, filter, nil)).To(ConsistOf([]*upload.Upload{dataSetExistingOne, dataSet, dataSetExistingTwo}))
// 					})

// 					It("succeeds if the pagination size is not default", func() {
// 						pagination.Size = 2
// 						Expect(mongoSession.GetDataSetsForUserByID(userID, filter, pagination)).To(ConsistOf([]*upload.Upload{dataSetExistingOne, dataSet}))
// 					})

// 					It("succeeds if the pagination page and size is not default", func() {
// 						pagination.Page = 1
// 						pagination.Size = 2
// 						Expect(mongoSession.GetDataSetsForUserByID(userID, filter, pagination)).To(ConsistOf([]*upload.Upload{dataSetExistingTwo}))
// 					})

// 					It("succeeds if it successfully does not find another user data sets", func() {
// 						resultDataSets, err := mongoSession.GetDataSetsForUserByID(id.New(), filter, pagination)
// 						Expect(err).ToNot(HaveOccurred())
// 						Expect(resultDataSets).ToNot(BeNil())
// 						Expect(resultDataSets).To(BeEmpty())
// 					})

// 					It("returns an error if the user id is missing", func() {
// 						resultDataSets, err := mongoSession.GetDataSetsForUserByID("", filter, pagination)
// 						Expect(err).To(MatchError("user id is missing"))
// 						Expect(resultDataSets).To(BeNil())
// 					})

// 					It("returns an error if the pagination page is less than minimum", func() {
// 						pagination.Page = -1
// 						resultDataSets, err := mongoSession.GetDataSetsForUserByID(userID, filter, pagination)
// 						Expect(err).To(MatchError("pagination is invalid; page is invalid"))
// 						Expect(resultDataSets).To(BeNil())
// 					})

// 					It("returns an error if the pagination size is less than minimum", func() {
// 						pagination.Size = 0
// 						resultDataSets, err := mongoSession.GetDataSetsForUserByID(userID, filter, pagination)
// 						Expect(err).To(MatchError("pagination is invalid; size is invalid"))
// 						Expect(resultDataSets).To(BeNil())
// 					})

// 					It("returns an error if the pagination size is greater than maximum", func() {
// 						pagination.Size = 101
// 						resultDataSets, err := mongoSession.GetDataSetsForUserByID(userID, filter, pagination)
// 						Expect(err).To(MatchError("pagination is invalid; size is invalid"))
// 						Expect(resultDataSets).To(BeNil())
// 					})

// 					It("returns an error if the session is closed", func() {
// 						mongoSession.Close()
// 						resultDataSets, err := mongoSession.GetDataSetsForUserByID(userID, filter, pagination)
// 						Expect(err).To(MatchError("session closed"))
// 						Expect(resultDataSets).To(BeNil())
// 					})

// 					Context("with deleted data set", func() {
// 						BeforeEach(func() {
// 							dataSet.DeletedTime = "2016-09-01T13:00:00Z"
// 							Expect(testMongoCollection.Update(bson.M{"id": dataSet.ID}, dataSet)).To(Succeed())
// 						})

// 						It("succeeds if it successfully finds the non-deleted user data sets", func() {
// 							Expect(mongoSession.GetDataSetsForUserByID(userID, filter, pagination)).To(ConsistOf([]*upload.Upload{dataSetExistingOne, dataSetExistingTwo}))
// 						})

// 						It("succeeds if it successfully finds all the user data sets", func() {
// 							filter.Deleted = true
// 							Expect(mongoSession.GetDataSetsForUserByID(userID, filter, pagination)).To(ConsistOf([]*upload.Upload{dataSetExistingOne, dataSet, dataSetExistingTwo}))
// 						})
// 					})
// 				})

// 				Context("GetDataSetByID", func() {
// 					BeforeEach(func() {
// 						dataSet.CreatedTime = "2016-09-01T11:00:00Z"
// 						Expect(testMongoCollection.Insert(dataSet)).To(Succeed())
// 					})

// 					It("succeeds if it successfully finds the data set", func() {
// 						Expect(mongoSession.GetDataSetByID(dataSet.UploadID)).To(Equal(dataSet))
// 					})

// 					It("returns an error if the data set id is missing", func() {
// 						resultDataSet, err := mongoSession.GetDataSetByID("")
// 						Expect(err).To(MatchError("data set id is missing"))
// 						Expect(resultDataSet).To(BeNil())
// 					})

// 					It("returns an error if the session is closed", func() {
// 						mongoSession.Close()
// 						resultDataSet, err := mongoSession.GetDataSetByID(dataSet.UploadID)
// 						Expect(err).To(MatchError("session closed"))
// 						Expect(resultDataSet).To(BeNil())
// 					})

// 					It("returns no data set successfully if the data set cannot be found", func() {
// 						resultDataSet, err := mongoSession.GetDataSetByID("not-found")
// 						Expect(err).ToNot(HaveOccurred())
// 						Expect(resultDataSet).To(BeNil())
// 					})
// 				})

// 				Context("CreateDataSet", func() {
// 					It("succeeds if it successfully creates the data set", func() {
// 						Expect(mongoSession.CreateDataSet(dataSet)).To(Succeed())
// 					})

// 					It("returns an error if the data set is missing", func() {
// 						Expect(mongoSession.CreateDataSet(nil)).To(MatchError("data set is missing"))
// 					})

// 					It("returns an error if the user id is missing", func() {
// 						dataSet.UserID = ""
// 						Expect(mongoSession.CreateDataSet(dataSet)).To(MatchError("data set user id is missing"))
// 					})

// 					It("returns an error if the upload id is missing", func() {
// 						dataSet.UploadID = ""
// 						Expect(mongoSession.CreateDataSet(dataSet)).To(MatchError("data set upload id is missing"))
// 					})

// 					It("returns an error if the device id is missing (nil)", func() {
// 						dataSet.DeviceID = nil
// 						Expect(mongoSession.CreateDataSet(dataSet)).To(MatchError("data set device id is missing"))
// 					})

// 					It("returns an error if the device id is missing (empty)", func() {
// 						dataSet.DeviceID = pointer.FromString("")
// 						Expect(mongoSession.CreateDataSet(dataSet)).To(MatchError("data set device id is missing"))
// 					})

// 					It("returns an error if the session is closed", func() {
// 						mongoSession.Close()
// 						Expect(mongoSession.CreateDataSet(dataSet)).To(MatchError("session closed"))
// 					})

// 					It("returns an error if the data set with the same id already exists", func() {
// 						Expect(mongoSession.CreateDataSet(dataSet)).To(Succeed())
// 						Expect(mongoSession.CreateDataSet(dataSet)).To(MatchError("unable to create data set; data set already exists"))
// 					})

// 					It("sets the created time", func() {
// 						Expect(mongoSession.CreateDataSet(dataSet)).To(Succeed())
// 						Expect(dataSet.CreatedTime).ToNot(BeEmpty())
// 						Expect(dataSet.CreatedUserID).To(BeEmpty())
// 						Expect(dataSet.ByUser).To(BeNil())
// 					})

// 					It("has the correct stored data sets", func() {
// 						ValidateDataSet(testMongoCollection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, bson.M{}, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo)
// 						Expect(mongoSession.CreateDataSet(dataSet)).To(Succeed())
// 						ValidateDataSet(testMongoCollection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, bson.M{}, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo, dataSet)
// 					})
// 				})

// 				Context("UpdateDataSet", func() {
// 					BeforeEach(func() {
// 						dataSet.CreatedTime = "2016-09-01T11:00:00Z"
// 						Expect(testMongoCollection.Insert(dataSet)).To(Succeed())
// 					})

// 					Context("with state closed", func() {
// 						BeforeEach(func() {
// 							dataSet.State = "closed"
// 						})

// 						It("succeeds if it successfully updates the data set", func() {
// 							Expect(mongoSession.UpdateDataSet(dataSet)).To(Succeed())
// 						})

// 						It("returns an error if the data set is missing", func() {
// 							Expect(mongoSession.UpdateDataSet(nil)).To(MatchError("data set is missing"))
// 						})

// 						It("returns an error if the user id is missing", func() {
// 							dataSet.UserID = ""
// 							Expect(mongoSession.UpdateDataSet(dataSet)).To(MatchError("data set user id is missing"))
// 						})

// 						It("returns an error if the upload id is missing", func() {
// 							dataSet.UploadID = ""
// 							Expect(mongoSession.UpdateDataSet(dataSet)).To(MatchError("data set upload id is missing"))
// 						})

// 						It("returns an error if the device id is missing (nil)", func() {
// 							dataSet.DeviceID = nil
// 							Expect(mongoSession.UpdateDataSet(dataSet)).To(MatchError("data set device id is missing"))
// 						})

// 						It("returns an error if the device id is missing (empty)", func() {
// 							dataSet.DeviceID = pointer.FromString("")
// 							Expect(mongoSession.UpdateDataSet(dataSet)).To(MatchError("data set device id is missing"))
// 						})

// 						It("returns an error if the session is closed", func() {
// 							mongoSession.Close()
// 							Expect(mongoSession.UpdateDataSet(dataSet)).To(MatchError("session closed"))
// 						})

// 						It("returns an error if the data set with the same user id and upload id does not yet exist", func() {
// 							dataSet.UploadID = id.New()
// 							Expect(mongoSession.UpdateDataSet(dataSet)).To(MatchError("unable to update data set; not found"))
// 						})
// 					})

// 					It("sets the modified time", func() {
// 						dataSet.State = "closed"
// 						Expect(mongoSession.UpdateDataSet(dataSet)).To(Succeed())
// 						Expect(dataSet.ModifiedTime).ToNot(BeEmpty())
// 						Expect(dataSet.ModifiedUserID).To(BeEmpty())
// 					})

// 					It("has the correct stored data sets", func() {
// 						ValidateDataSet(testMongoCollection, bson.M{}, bson.M{}, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo, dataSet)
// 						ValidateDataSet(testMongoCollection, bson.M{"modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, bson.M{})
// 						dataSet.State = "closed"
// 						Expect(mongoSession.UpdateDataSet(dataSet)).To(Succeed())
// 						ValidateDataSet(testMongoCollection, bson.M{}, bson.M{}, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo, dataSet)
// 						ValidateDataSet(testMongoCollection, bson.M{"modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, bson.M{}, dataSet)
// 					})
// 				})

// 				Context("with data", func() {
// 					var dataSetExistingOtherData []data.Datum
// 					var dataSetExistingOneData []data.Datum
// 					var dataSetExistingTwoData []data.Datum
// 					var dataSetData []data.Datum

// 					BeforeEach(func() {
// 						dataSet.CreatedTime = "2016-09-01T11:00:00Z"
// 						Expect(testMongoCollection.Insert(dataSet)).To(Succeed())
// 						dataSetExistingOtherData = NewDataSetData(id.New())
// 						Expect(mongoSession.CreateDataSetData(dataSetExistingOther, dataSetExistingOtherData)).To(Succeed())
// 						dataSetExistingOneData = NewDataSetData(deviceID)
// 						Expect(mongoSession.CreateDataSetData(dataSetExistingOne, dataSetExistingOneData)).To(Succeed())
// 						dataSetExistingTwoData = NewDataSetData(deviceID)
// 						Expect(mongoSession.CreateDataSetData(dataSetExistingTwo, dataSetExistingTwoData)).To(Succeed())
// 						dataSetData = NewDataSetData(deviceID)
// 					})

// 					Context("DeleteDataSet", func() {
// 						BeforeEach(func() {
// 							Expect(mongoSession.CreateDataSetData(dataSet, dataSetData)).To(Succeed())
// 						})

// 						It("succeeds if it successfully deletes the data set", func() {
// 							Expect(mongoSession.DeleteDataSet(dataSet)).To(Succeed())
// 						})

// 						It("returns an error if the data set is missing", func() {
// 							Expect(mongoSession.DeleteDataSet(nil)).To(MatchError("data set is missing"))
// 						})

// 						It("returns an error if the user id is missing", func() {
// 							dataSet.UserID = ""
// 							Expect(mongoSession.DeleteDataSet(dataSet)).To(MatchError("data set user id is missing"))
// 						})

// 						It("returns an error if the upload id is missing", func() {
// 							dataSet.UploadID = ""
// 							Expect(mongoSession.DeleteDataSet(dataSet)).To(MatchError("data set upload id is missing"))
// 						})

// 						It("returns an error if the device id is missing (nil)", func() {
// 							dataSet.DeviceID = nil
// 							Expect(mongoSession.DeleteDataSet(dataSet)).To(MatchError("data set device id is missing"))
// 						})

// 						It("returns an error if the device id is missing (empty)", func() {
// 							dataSet.DeviceID = pointer.FromString("")
// 							Expect(mongoSession.DeleteDataSet(dataSet)).To(MatchError("data set device id is missing"))
// 						})

// 						It("returns an error if the session is closed", func() {
// 							mongoSession.Close()
// 							Expect(mongoSession.DeleteDataSet(dataSet)).To(MatchError("session closed"))
// 						})

// 						It("sets the deleted time on the data set", func() {
// 							Expect(mongoSession.DeleteDataSet(dataSet)).To(Succeed())
// 							Expect(dataSet.DeletedTime).ToNot(BeEmpty())
// 							Expect(dataSet.DeletedUserID).To(BeEmpty())
// 						})

// 						It("has the correct stored data sets", func() {
// 							ValidateDataSet(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": true}, "deletedUserId": bson.M{"$exists": false}}, bson.M{})
// 							Expect(mongoSession.DeleteDataSet(dataSet)).To(Succeed())
// 							ValidateDataSet(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": true}, "deletedUserId": bson.M{"$exists": false}}, bson.M{}, dataSet)
// 						})

// 						It("has the correct stored data set data", func() {
// 							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSetData)
// 							Expect(mongoSession.DeleteDataSet(dataSet)).To(Succeed())
// 							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, []data.Datum{})
// 						})
// 					})

// 					Context("CreateDataSetData", func() {
// 						It("succeeds if it successfully creates the data set data", func() {
// 							Expect(mongoSession.CreateDataSetData(dataSet, dataSetData)).To(Succeed())
// 						})

// 						It("returns an error if the data set is missing", func() {
// 							Expect(mongoSession.CreateDataSetData(nil, dataSetData)).To(MatchError("data set is missing"))
// 						})

// 						It("returns an error if the data set data is missing", func() {
// 							Expect(mongoSession.CreateDataSetData(dataSet, nil)).To(MatchError("data set data is missing"))
// 						})

// 						It("returns an error if the user id is missing", func() {
// 							dataSet.UserID = ""
// 							Expect(mongoSession.CreateDataSetData(dataSet, dataSetData)).To(MatchError("data set user id is missing"))
// 						})

// 						It("returns an error if the upload id is missing", func() {
// 							dataSet.UploadID = ""
// 							Expect(mongoSession.CreateDataSetData(dataSet, dataSetData)).To(MatchError("data set upload id is missing"))
// 						})

// 						It("returns an error if the device id is missing (nil)", func() {
// 							dataSet.DeviceID = nil
// 							Expect(mongoSession.CreateDataSetData(dataSet, dataSetData)).To(MatchError("data set device id is missing"))
// 						})

// 						It("returns an error if the device id is missing (empty)", func() {
// 							dataSet.DeviceID = pointer.FromString("")
// 							Expect(mongoSession.CreateDataSetData(dataSet, dataSetData)).To(MatchError("data set device id is missing"))
// 						})

// 						It("returns an error if the session is closed", func() {
// 							mongoSession.Close()
// 							Expect(mongoSession.CreateDataSetData(dataSet, dataSetData)).To(MatchError("session closed"))
// 						})

// 						It("sets the user id and upload id on the data set data to match the data set", func() {
// 							Expect(mongoSession.CreateDataSetData(dataSet, dataSetData)).To(Succeed())
// 							for _, dataSetDatum := range dataSetData {
// 								baseDatum, ok := dataSetDatum.(*types.Base)
// 								Expect(ok).To(BeTrue())
// 								Expect(baseDatum).ToNot(BeNil())
// 								Expect(baseDatum.UserID).To(Equal(dataSet.UserID))
// 								Expect(baseDatum.UploadID).To(Equal(dataSet.UploadID))
// 							}
// 						})

// 						It("leaves the data set data not active", func() {
// 							Expect(mongoSession.CreateDataSetData(dataSet, dataSetData)).To(Succeed())
// 							for _, dataSetDatum := range dataSetData {
// 								baseDatum, ok := dataSetDatum.(*types.Base)
// 								Expect(ok).To(BeTrue())
// 								Expect(baseDatum).ToNot(BeNil())
// 								Expect(baseDatum.Active).To(BeFalse())
// 							}
// 						})

// 						It("sets the created time on the data set data", func() {
// 							Expect(mongoSession.CreateDataSetData(dataSet, dataSetData)).To(Succeed())
// 							for _, dataSetDatum := range dataSetData {
// 								baseDatum, ok := dataSetDatum.(*types.Base)
// 								Expect(ok).To(BeTrue())
// 								Expect(baseDatum).ToNot(BeNil())
// 								Expect(baseDatum.CreatedTime).ToNot(BeEmpty())
// 								Expect(baseDatum.CreatedUserID).To(BeEmpty())
// 							}
// 						})

// 						It("has the correct stored data set data", func() {
// 							dataSetBeforeCreateData := append(append(dataSetExistingOtherData, dataSetExistingOneData...), dataSetExistingTwoData...)
// 							ValidateDataSetData(testMongoCollection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, bson.M{}, dataSetBeforeCreateData)
// 							Expect(mongoSession.CreateDataSetData(dataSet, dataSetData)).To(Succeed())
// 							ValidateDataSetData(testMongoCollection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, bson.M{}, append(dataSetBeforeCreateData, dataSetData...))
// 						})
// 					})

// 					Context("ActivateDataSetData", func() {
// 						BeforeEach(func() {
// 							Expect(mongoSession.CreateDataSetData(dataSet, dataSetData)).To(Succeed())
// 						})

// 						It("succeeds if it successfully activates the data set", func() {
// 							Expect(mongoSession.ActivateDataSetData(dataSet)).To(Succeed())
// 						})

// 						It("returns an error if the data set is missing", func() {
// 							Expect(mongoSession.ActivateDataSetData(nil)).To(MatchError("data set is missing"))
// 						})

// 						It("returns an error if the user id is missing", func() {
// 							dataSet.UserID = ""
// 							Expect(mongoSession.ActivateDataSetData(dataSet)).To(MatchError("data set user id is missing"))
// 						})

// 						It("returns an error if the upload id is missing", func() {
// 							dataSet.UploadID = ""
// 							Expect(mongoSession.ActivateDataSetData(dataSet)).To(MatchError("data set upload id is missing"))
// 						})

// 						It("returns an error if the device id is missing (nil)", func() {
// 							dataSet.DeviceID = nil
// 							Expect(mongoSession.ActivateDataSetData(dataSet)).To(MatchError("data set device id is missing"))
// 						})

// 						It("returns an error if the device id is missing (empty)", func() {
// 							dataSet.DeviceID = pointer.FromString("")
// 							Expect(mongoSession.ActivateDataSetData(dataSet)).To(MatchError("data set device id is missing"))
// 						})

// 						It("returns an error if the session is closed", func() {
// 							mongoSession.Close()
// 							Expect(mongoSession.ActivateDataSetData(dataSet)).To(MatchError("session closed"))
// 						})

// 						It("sets the active on the data set", func() {
// 							Expect(mongoSession.ActivateDataSetData(dataSet)).To(Succeed())
// 							Expect(dataSet.Active).To(BeTrue())
// 						})

// 						It("sets the modified time on the data set", func() {
// 							Expect(mongoSession.ActivateDataSetData(dataSet)).To(Succeed())
// 							Expect(dataSet.ModifiedTime).ToNot(BeEmpty())
// 							Expect(dataSet.ModifiedUserID).To(BeEmpty())
// 						})

// 						It("has the correct stored active data set", func() {
// 							ValidateDataSet(testMongoCollection, bson.M{"_active": true, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, bson.M{})
// 							Expect(mongoSession.ActivateDataSetData(dataSet)).To(Succeed())
// 							ValidateDataSet(testMongoCollection, bson.M{"_active": true, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, bson.M{}, dataSet)
// 						})

// 						It("has the correct stored active data set data", func() {
// 							ValidateDataSetData(testMongoCollection, bson.M{"_active": true, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, bson.M{}, []data.Datum{})
// 							Expect(mongoSession.ActivateDataSetData(dataSet)).To(Succeed())
// 							for _, dataSetDatum := range dataSetData {
// 								dataSetDatum.SetActive(true)
// 							}
// 							ValidateDataSetData(testMongoCollection, bson.M{"_active": true, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, bson.M{"modifiedTime": 0}, dataSetData)
// 						})
// 					})

// 					Context("ArchiveDeviceDataUsingHashesFromDataSet", func() {
// 						var dataSetExistingOneDataCloned []data.Datum

// 						BeforeEach(func() {
// 							dataSetExistingOneDataCloned = CloneDataSetData(dataSetData)
// 							Expect(mongoSession.CreateDataSetData(dataSetExistingOne, dataSetExistingOneDataCloned)).To(Succeed())
// 							Expect(mongoSession.ActivateDataSetData(dataSetExistingOne)).To(Succeed())
// 							Expect(mongoSession.CreateDataSetData(dataSet, dataSetData)).To(Succeed())
// 							for _, dataSetDatum := range append(dataSetExistingOneData, dataSetExistingOneDataCloned...) {
// 								dataSetDatum.SetActive(true)
// 							}
// 						})

// 						It("succeeds if it successfully archives device data using hashes from data set", func() {
// 							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataSet(dataSet)).To(Succeed())
// 						})

// 						It("returns an error if the data set is missing", func() {
// 							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataSet(nil)).To(MatchError("data set is missing"))
// 						})

// 						It("returns an error if the user id is missing", func() {
// 							dataSet.UserID = ""
// 							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataSet(dataSet)).To(MatchError("data set user id is missing"))
// 						})

// 						It("returns an error if the upload id is missing", func() {
// 							dataSet.UploadID = ""
// 							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataSet(dataSet)).To(MatchError("data set upload id is missing"))
// 						})

// 						It("returns an error if the device id is missing (nil)", func() {
// 							dataSet.DeviceID = nil
// 							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataSet(dataSet)).To(MatchError("data set device id is missing"))
// 						})

// 						It("returns an error if the device id is missing (empty)", func() {
// 							dataSet.DeviceID = pointer.FromString("")
// 							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataSet(dataSet)).To(MatchError("data set device id is missing"))
// 						})

// 						It("returns an error if the session is closed", func() {
// 							mongoSession.Close()
// 							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataSet(dataSet)).To(MatchError("session closed"))
// 						})

// 						It("has the correct stored data sets", func() {
// 							ValidateDataSet(testMongoCollection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true}, bson.M{}, dataSetExistingOne)
// 							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataSet(dataSet)).To(Succeed())
// 							ValidateDataSet(testMongoCollection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true}, bson.M{}, dataSetExistingOne)
// 						})

// 						It("has the correct stored archived data set data", func() {
// 							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": false}, bson.M{}, []data.Datum{})
// 							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true}, bson.M{"modifiedTime": 0}, append(dataSetExistingOneData, dataSetExistingOneDataCloned...))
// 							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataSet(dataSet)).To(Succeed())
// 							for _, dataSetDatum := range dataSetExistingOneDataCloned {
// 								dataSetDatum.SetActive(false)
// 							}
// 							ValidateDataSetData(testMongoCollection,
// 								bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}},
// 								bson.M{"modifiedTime": 0},
// 								dataSetExistingOneData)
// 							ValidateDataSetData(testMongoCollection,
// 								bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": false, "archivedTime": bson.M{"$exists": true}, "archivedDatasetId": dataSet.UploadID, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}},
// 								bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0},
// 								dataSetExistingOneDataCloned)
// 							ValidateDataSetData(testMongoCollection,
// 								bson.M{"uploadId": dataSet.UploadID, "_active": false, "archivedTime": bson.M{"$exists": false}, "archivedDatasetId": bson.M{"$exists": false}},
// 								bson.M{},
// 								dataSetData)
// 						})
// 					})

// 					Context("UnarchiveDeviceDataUsingHashesFromDataSet", func() {
// 						var dataSetExistingTwoDataCloned []data.Datum
// 						var dataSetExistingOneDataCloned []data.Datum

// 						BeforeEach(func() {
// 							dataSetExistingTwoDataCloned = CloneDataSetData(dataSetData)
// 							dataSetExistingOneDataCloned = CloneDataSetData(dataSetData)
// 							Expect(mongoSession.CreateDataSetData(dataSetExistingTwo, dataSetExistingTwoDataCloned)).To(Succeed())
// 							Expect(mongoSession.ActivateDataSetData(dataSetExistingTwo)).To(Succeed())
// 							Expect(mongoSession.CreateDataSetData(dataSetExistingOne, dataSetExistingOneDataCloned)).To(Succeed())
// 							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataSet(dataSetExistingOne)).To(Succeed())
// 							Expect(mongoSession.ActivateDataSetData(dataSetExistingOne)).To(Succeed())
// 							Expect(mongoSession.CreateDataSetData(dataSet, dataSetData)).To(Succeed())
// 							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataSet(dataSet)).To(Succeed())
// 							Expect(mongoSession.ActivateDataSetData(dataSet)).To(Succeed())
// 						})

// 						It("succeeds if it successfully unarchives device data using hashes from data set", func() {
// 							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataSet(dataSet)).To(Succeed())
// 						})

// 						It("returns an error if the data set is missing", func() {
// 							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataSet(nil)).To(MatchError("data set is missing"))
// 						})

// 						It("returns an error if the user id is missing", func() {
// 							dataSet.UserID = ""
// 							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataSet(dataSet)).To(MatchError("data set user id is missing"))
// 						})

// 						It("returns an error if the upload id is missing", func() {
// 							dataSet.UploadID = ""
// 							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataSet(dataSet)).To(MatchError("data set upload id is missing"))
// 						})

// 						It("returns an error if the device id is missing (nil)", func() {
// 							dataSet.DeviceID = nil
// 							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataSet(dataSet)).To(MatchError("data set device id is missing"))
// 						})

// 						It("returns an error if the device id is missing (empty)", func() {
// 							dataSet.DeviceID = pointer.FromString("")
// 							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataSet(dataSet)).To(MatchError("data set device id is missing"))
// 						})

// 						It("returns an error if the session is closed", func() {
// 							mongoSession.Close()
// 							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataSet(dataSet)).To(MatchError("session closed"))
// 						})

// 						It("has the correct stored data sets", func() {
// 							ValidateDataSet(testMongoCollection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true}, bson.M{}, dataSetExistingOne)
// 							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataSet(dataSet)).To(Succeed())
// 							ValidateDataSet(testMongoCollection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true}, bson.M{}, dataSetExistingOne)
// 						})

// 						It("has the correct stored unarchived data set data", func() {
// 							for _, dataSetDatum := range append(dataSetData, dataSetExistingOneData...) {
// 								dataSetDatum.SetActive(true)
// 							}
// 							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": false}, bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0}, dataSetExistingOneDataCloned)
// 							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true}, bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0}, dataSetExistingOneData)
// 							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSet.UploadID, "_active": true}, bson.M{"modifiedTime": 0}, dataSetData)
// 							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataSet(dataSet)).To(Succeed())
// 							for _, dataSetDatum := range dataSetExistingOneDataCloned {
// 								dataSetDatum.SetActive(true)
// 							}
// 							ValidateDataSetData(testMongoCollection,
// 								bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true, "archivedTime": bson.M{"$exists": false}, "archivedDatasetId": bson.M{"$exists": false}, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}},
// 								bson.M{"modifiedTime": 0},
// 								append(dataSetExistingOneData, dataSetExistingOneDataCloned...))
// 							ValidateDataSetData(testMongoCollection,
// 								bson.M{"uploadId": dataSet.UploadID, "_active": true, "archivedTime": bson.M{"$exists": false}, "archivedDatasetId": bson.M{"$exists": false}},
// 								bson.M{"modifiedTime": 0},
// 								dataSetData)
// 						})

// 						It("has the correct stored data sets if an intermediary is unarchived", func() {
// 							ValidateDataSet(testMongoCollection, bson.M{"uploadId": dataSetExistingTwo.UploadID, "_active": true}, bson.M{}, dataSetExistingTwo)
// 							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataSet(dataSetExistingOne)).To(Succeed())
// 							ValidateDataSet(testMongoCollection, bson.M{"uploadId": dataSetExistingTwo.UploadID, "_active": true}, bson.M{}, dataSetExistingTwo)
// 						})

// 						It("has the correct stored unarchived data set data if an intermediary is unarchived", func() {
// 							for _, dataSetDatum := range append(dataSetExistingOneData, dataSetExistingTwoData...) {
// 								dataSetDatum.SetActive(true)
// 							}
// 							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSetExistingTwo.UploadID, "_active": false}, bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0}, dataSetExistingTwoDataCloned)
// 							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSetExistingTwo.UploadID, "_active": true}, bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0}, dataSetExistingTwoData)
// 							ValidateDataSetData(testMongoCollection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true}, bson.M{"modifiedTime": 0}, dataSetExistingOneData)
// 							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataSet(dataSetExistingOne)).To(Succeed())
// 							ValidateDataSetData(testMongoCollection,
// 								bson.M{"uploadId": dataSetExistingTwo.UploadID, "_active": true, "archivedTime": bson.M{"$exists": false}, "archivedDatasetId": bson.M{"$exists": false}},
// 								bson.M{"modifiedTime": 0},
// 								dataSetExistingTwoData)
// 							ValidateDataSetData(testMongoCollection,
// 								bson.M{"uploadId": dataSetExistingTwo.UploadID, "_active": false, "archivedTime": bson.M{"$exists": true}, "archivedDatasetId": dataSet.UploadID, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}},
// 								bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0},
// 								dataSetExistingTwoDataCloned)
// 							ValidateDataSetData(testMongoCollection,
// 								bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true, "archivedTime": bson.M{"$exists": false}, "archivedDatasetId": bson.M{"$exists": false}},
// 								bson.M{"modifiedTime": 0},
// 								dataSetExistingOneData)
// 							ValidateDataSetData(testMongoCollection,
// 								bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": false, "archivedTime": bson.M{"$exists": true}, "archivedDatasetId": dataSet.UploadID},
// 								bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0},
// 								dataSetExistingOneDataCloned)
// 						})
// 					})

// 					Context("DeleteOtherDataSetData", func() {
// 						BeforeEach(func() {
// 							Expect(mongoSession.CreateDataSetData(dataSet, dataSetData)).To(Succeed())
// 						})

// 						It("succeeds if it successfully deletes all other data set data", func() {
// 							Expect(mongoSession.DeleteOtherDataSetData(dataSet)).To(Succeed())
// 						})

// 						It("returns an error if the data set is missing", func() {
// 							Expect(mongoSession.DeleteOtherDataSetData(nil)).To(MatchError("data set is missing"))
// 						})

// 						It("returns an error if the user id is missing", func() {
// 							dataSet.UserID = ""
// 							Expect(mongoSession.DeleteOtherDataSetData(dataSet)).To(MatchError("data set user id is missing"))
// 						})

// 						It("returns an error if the upload id is missing", func() {
// 							dataSet.UploadID = ""
// 							Expect(mongoSession.DeleteOtherDataSetData(dataSet)).To(MatchError("data set upload id is missing"))
// 						})

// 						It("returns an error if the device id is missing (nil)", func() {
// 							dataSet.DeviceID = nil
// 							Expect(mongoSession.DeleteOtherDataSetData(dataSet)).To(MatchError("data set device id is missing"))
// 						})

// 						It("returns an error if the device id is missing (empty)", func() {
// 							dataSet.DeviceID = pointer.FromString("")
// 							Expect(mongoSession.DeleteOtherDataSetData(dataSet)).To(MatchError("data set device id is missing"))
// 						})

// 						It("returns an error if the session is closed", func() {
// 							mongoSession.Close()
// 							Expect(mongoSession.DeleteOtherDataSetData(dataSet)).To(MatchError("session closed"))
// 						})

// 						It("has the correct stored active data set", func() {
// 							ValidateDataSet(testMongoCollection, bson.M{}, bson.M{}, dataSet, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo)
// 							ValidateDataSet(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": false}, "deletedUserId": bson.M{"$exists": false}}, bson.M{}, dataSet, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo)
// 							Expect(mongoSession.DeleteOtherDataSetData(dataSet)).To(Succeed())
// 							Expect(testMongoCollection.Find(bson.M{"type": "upload"}).Count()).To(Equal(4))
// 							ValidateDataSet(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": true}, "deletedUserId": bson.M{"$exists": false}}, bson.M{"deletedTime": 0}, dataSetExistingTwo, dataSetExistingOne)
// 							ValidateDataSet(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": false}, "deletedUserId": bson.M{"$exists": false}}, bson.M{}, dataSet, dataSetExistingOther)
// 						})

// 						It("has the correct stored active data set data", func() {
// 							dataSetDataAfterRemoveData := append(dataSetData, dataSetExistingOtherData...)
// 							dataSetDataBeforeRemoveData := append(append(dataSetDataAfterRemoveData, dataSetExistingOneData...), dataSetExistingTwoData...)
// 							ValidateDataSetData(testMongoCollection, bson.M{}, bson.M{}, dataSetDataBeforeRemoveData)
// 							Expect(mongoSession.DeleteOtherDataSetData(dataSet)).To(Succeed())
// 							ValidateDataSetData(testMongoCollection, bson.M{}, bson.M{"deletedTime": 0}, dataSetDataAfterRemoveData)
// 						})
// 					})

// 					Context("DestroyDataForUserByID", func() {
// 						var deleteUserID string
// 						var deleteDeviceID string
// 						var deleteDataSet *upload.Upload
// 						var deleteDataSetData []data.Datum

// 						BeforeEach(func() {
// 							Expect(mongoSession.CreateDataSetData(dataSet, dataSetData)).To(Succeed())
// 							deleteUserID = id.New()
// 							deleteDeviceID = id.New()
// 							deleteDataSet = NewDataSet(deleteUserID, deleteDeviceID)
// 							deleteDataSet.CreatedTime = "2016-09-01T11:00:00Z"
// 							Expect(testMongoCollection.Insert(deleteDataSet)).To(Succeed())
// 							deleteDataSetData = NewDataSetData(deleteDeviceID)
// 							Expect(mongoSession.CreateDataSetData(deleteDataSet, deleteDataSetData)).To(Succeed())
// 						})

// 						It("succeeds if it successfully destroys all data for user by id", func() {
// 							Expect(mongoSession.DestroyDataForUserByID(deleteUserID)).To(Succeed())
// 						})

// 						It("returns an error if the user id is missing", func() {
// 							Expect(mongoSession.DestroyDataForUserByID("")).To(MatchError("user id is missing"))
// 						})

// 						It("returns an error if the session is closed", func() {
// 							mongoSession.Close()
// 							Expect(mongoSession.DestroyDataForUserByID(deleteUserID)).To(MatchError("session closed"))
// 						})

// 						It("has the correct stored data sets", func() {
// 							ValidateDataSet(testMongoCollection, bson.M{}, bson.M{}, dataSet, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo, deleteDataSet)
// 							Expect(mongoSession.DestroyDataForUserByID(deleteUserID)).To(Succeed())
// 							ValidateDataSet(testMongoCollection, bson.M{}, bson.M{}, dataSet, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo)
// 						})

// 						It("has the correct stored data set data", func() {
// 							dataSetDataAfterRemoveData := append(append(append(dataSetData, dataSetExistingOtherData...), dataSetExistingOneData...), dataSetExistingTwoData...)
// 							dataSetDataBeforeRemoveData := append(dataSetDataAfterRemoveData, deleteDataSetData...)
// 							ValidateDataSetData(testMongoCollection, bson.M{}, bson.M{}, dataSetDataBeforeRemoveData)
// 							Expect(mongoSession.DestroyDataForUserByID(deleteUserID)).To(Succeed())
// 							ValidateDataSetData(testMongoCollection, bson.M{}, bson.M{}, dataSetDataAfterRemoveData)
// 						})
// 					})
// 				})
// 			})
// 		})
// 	})
// })
