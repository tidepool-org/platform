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
// 	baseMongo "github.com/tidepool-org/platform/store/mongo"
// 	testMongo "github.com/tidepool-org/platform/test/mongo"
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

// func NewDataset(userID string, deviceID string) *upload.Upload {
// 	dataset := upload.New()
// 	Expect(dataset).ToNot(BeNil())

// 	dataset.Deduplicator = &data.DeduplicatorDescriptor{Name: "test-deduplicator"}
// 	dataset.UserID = userID

// 	dataset.ClockDriftOffset = pointer.Int(0)
// 	dataset.ConversionOffset = pointer.Int(0)
// 	dataset.DeviceID = pointer.String(deviceID)
// 	dataset.DeviceTime = pointer.String(SampleTime().Format("2006-01-02T15:04:05"))
// 	dataset.Time = pointer.String(SampleTime().UTC().Format(time.RFC3339))
// 	dataset.TimeZoneOffset = pointer.Int(-420)

// 	dataset.ComputerTime = pointer.String(SampleTime().Format("2006-01-02T15:04:05"))
// 	dataset.DeviceManufacturers = pointer.StringArray([]string{"Tesla"})
// 	dataset.DeviceModel = pointer.String("1234")
// 	dataset.DeviceSerialNumber = pointer.String("567890")
// 	dataset.DeviceTags = pointer.StringArray([]string{upload.DeviceTagInsulinPump})
// 	dataset.TimeProcessing = pointer.String(upload.TimeProcessingUTCBootstrapping)
// 	dataset.TimeZoneName = pointer.String("US/Pacific")
// 	dataset.Version = pointer.String("0.260.1")

// 	return dataset
// }

// func NewDatasetData(deviceID string) []data.Datum {
// 	datasetData := []data.Datum{}
// 	for count := 0; count < 3; count++ {
// 		baseDatum := &types.Base{}
// 		baseDatum.New()

// 		baseDatum.Deduplicator = &data.DeduplicatorDescriptor{Hash: id.New()}
// 		baseDatum.Type = "test"

// 		baseDatum.ClockDriftOffset = pointer.Int(0)
// 		baseDatum.ConversionOffset = pointer.Int(0)
// 		baseDatum.DeviceID = pointer.String(deviceID)
// 		baseDatum.DeviceTime = pointer.String(SampleTime().Format("2006-01-02T15:04:05"))
// 		baseDatum.Time = pointer.String(SampleTime().UTC().Format(time.RFC3339))
// 		baseDatum.TimeZoneOffset = pointer.Int(-420)

// 		datasetData = append(datasetData, baseDatum)
// 	}
// 	return datasetData
// }

// func CloneDatasetData(datasetData []data.Datum) []data.Datum {
// 	clonedDatasetData := []data.Datum{}
// 	for _, datasetDatum := range datasetData {
// 		if baseDatum, ok := datasetDatum.(*types.Base); ok {
// 			clonedBaseDatum := &types.Base{}
// 			clonedBaseDatum.Active = baseDatum.Active
// 			clonedBaseDatum.ArchivedDatasetID = baseDatum.ArchivedDatasetID
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

// 			clonedDatasetData = append(clonedDatasetData, clonedBaseDatum)
// 		}
// 	}
// 	return clonedDatasetData
// }

// func ValidateDataset(testMongoCollection *mgo.Collection, query bson.M, filter bson.M, expectedDatasets ...*upload.Upload) {
// 	query["type"] = "upload"
// 	filter["_id"] = 0
// 	var actualDatasets []*upload.Upload
// 	Expect(testMongoCollection.Find(query).Select(filter).All(&actualDatasets)).To(Succeed())
// 	Expect(actualDatasets).To(ConsistOf(DatasetsAsInterface(expectedDatasets)...))
// }

// func DatasetsAsInterface(datasets []*upload.Upload) []interface{} {
// 	var datasetsAsInterface []interface{}
// 	for _, dataset := range datasets {
// 		datasetsAsInterface = append(datasetsAsInterface, dataset)
// 	}
// 	return datasetsAsInterface
// }

// func ValidateDatasetData(testMongoCollection *mgo.Collection, query bson.M, filter bson.M, expectedDatasetData []data.Datum) {
// 	query["type"] = bson.M{"$ne": "upload"}
// 	filter["_id"] = 0
// 	var actualDatasetData []interface{}
// 	Expect(testMongoCollection.Find(query).Select(filter).All(&actualDatasetData)).To(Succeed())
// 	Expect(actualDatasetData).To(ConsistOf(DatasetDataAsInterface(expectedDatasetData)...))
// }

// func DatasetDataAsInterface(datasetData []data.Datum) []interface{} {
// 	var datasetDataAsInterface []interface{}
// 	for _, datasetDatum := range datasetData {
// 		datasetDataAsInterface = append(datasetDataAsInterface, DatasetDatumAsInterface(datasetDatum))
// 	}
// 	return datasetDataAsInterface
// }

// func DatasetDatumAsInterface(datasetDatum data.Datum) interface{} {
// 	bytes, err := bson.Marshal(datasetDatum)
// 	Expect(err).ToNot(HaveOccurred())
// 	Expect(bytes).ToNot(BeNil())
// 	var datasetDatumAsInterface interface{}
// 	Expect(bson.Unmarshal(bytes, &datasetDatumAsInterface)).To(Succeed())
// 	return datasetDatumAsInterface
// }

// var _ = Describe("Mongo", func() {
// 	var mongoConfig *baseMongo.Config
// 	var mongoStore *mongo.Store
// 	var mongoSession storeDEPRECATED.DataSession

// 	BeforeEach(func() {
// 		mongoConfig = &baseMongo.Config{
// 			Addresses:        []string{testMongo.Address()},
// 			Database:         testMongo.Database(),
// 			CollectionPrefix: testMongo.NewCollectionPrefix(),
// 			Timeout:          5 * time.Second,
// 		}
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
// 				var datasetExistingOther *upload.Upload
// 				var datasetExistingOne *upload.Upload
// 				var datasetExistingTwo *upload.Upload
// 				var dataset *upload.Upload

// 				BeforeEach(func() {
// 					testMongoSession = testMongo.Session().Copy()
// 					testMongoCollection = testMongoSession.DB(mongoConfig.Database).C(mongoConfig.CollectionPrefix + "deviceData")
// 					userID = id.New()
// 					deviceID = id.New()
// 					datasetExistingOther = NewDataset(id.New(), id.New())
// 					datasetExistingOther.CreatedTime = "2016-09-01T12:00:00Z"
// 					Expect(testMongoCollection.Insert(datasetExistingOther)).To(Succeed())
// 					datasetExistingOne = NewDataset(userID, deviceID)
// 					datasetExistingOne.CreatedTime = "2016-09-01T12:30:00Z"
// 					Expect(testMongoCollection.Insert(datasetExistingOne)).To(Succeed())
// 					datasetExistingTwo = NewDataset(userID, deviceID)
// 					datasetExistingTwo.CreatedTime = "2016-09-01T10:00:00Z"
// 					Expect(testMongoCollection.Insert(datasetExistingTwo)).To(Succeed())
// 					dataset = NewDataset(userID, deviceID)
// 				})

// 				AfterEach(func() {
// 					if testMongoSession != nil {
// 						testMongoSession.Close()
// 					}
// 				})

// 				Context("GetDatasetsForUserByID", func() {
// 					var filter *storeDEPRECATED.Filter
// 					var pagination *page.Pagination

// 					BeforeEach(func() {
// 						dataset.CreatedTime = "2016-09-01T11:00:00Z"
// 						Expect(testMongoCollection.Insert(dataset)).To(Succeed())
// 						filter = storeDEPRECATED.NewFilter()
// 						pagination = page.NewPagination()
// 					})

// 					It("succeeds if it successfully finds the user datasets", func() {
// 						Expect(mongoSession.GetDatasetsForUserByID(userID, filter, pagination)).To(ConsistOf([]*upload.Upload{datasetExistingOne, dataset, datasetExistingTwo}))
// 					})

// 					It("succeeds if the filter is not specified", func() {
// 						Expect(mongoSession.GetDatasetsForUserByID(userID, nil, pagination)).To(ConsistOf([]*upload.Upload{datasetExistingOne, dataset, datasetExistingTwo}))
// 					})

// 					It("succeeds if the pagination is not specified", func() {
// 						Expect(mongoSession.GetDatasetsForUserByID(userID, filter, nil)).To(ConsistOf([]*upload.Upload{datasetExistingOne, dataset, datasetExistingTwo}))
// 					})

// 					It("succeeds if the pagination size is not default", func() {
// 						pagination.Size = 2
// 						Expect(mongoSession.GetDatasetsForUserByID(userID, filter, pagination)).To(ConsistOf([]*upload.Upload{datasetExistingOne, dataset}))
// 					})

// 					It("succeeds if the pagination page and size is not default", func() {
// 						pagination.Page = 1
// 						pagination.Size = 2
// 						Expect(mongoSession.GetDatasetsForUserByID(userID, filter, pagination)).To(ConsistOf([]*upload.Upload{datasetExistingTwo}))
// 					})

// 					It("succeeds if it successfully does not find another user datasets", func() {
// 						resultDatasets, err := mongoSession.GetDatasetsForUserByID(id.New(), filter, pagination)
// 						Expect(err).ToNot(HaveOccurred())
// 						Expect(resultDatasets).ToNot(BeNil())
// 						Expect(resultDatasets).To(BeEmpty())
// 					})

// 					It("returns an error if the user id is missing", func() {
// 						resultDatasets, err := mongoSession.GetDatasetsForUserByID("", filter, pagination)
// 						Expect(err).To(MatchError("user id is missing"))
// 						Expect(resultDatasets).To(BeNil())
// 					})

// 					It("returns an error if the pagination page is less than minimum", func() {
// 						pagination.Page = -1
// 						resultDatasets, err := mongoSession.GetDatasetsForUserByID(userID, filter, pagination)
// 						Expect(err).To(MatchError("pagination is invalid; page is invalid"))
// 						Expect(resultDatasets).To(BeNil())
// 					})

// 					It("returns an error if the pagination size is less than minimum", func() {
// 						pagination.Size = 0
// 						resultDatasets, err := mongoSession.GetDatasetsForUserByID(userID, filter, pagination)
// 						Expect(err).To(MatchError("pagination is invalid; size is invalid"))
// 						Expect(resultDatasets).To(BeNil())
// 					})

// 					It("returns an error if the pagination size is greater than maximum", func() {
// 						pagination.Size = 101
// 						resultDatasets, err := mongoSession.GetDatasetsForUserByID(userID, filter, pagination)
// 						Expect(err).To(MatchError("pagination is invalid; size is invalid"))
// 						Expect(resultDatasets).To(BeNil())
// 					})

// 					It("returns an error if the session is closed", func() {
// 						mongoSession.Close()
// 						resultDatasets, err := mongoSession.GetDatasetsForUserByID(userID, filter, pagination)
// 						Expect(err).To(MatchError("session closed"))
// 						Expect(resultDatasets).To(BeNil())
// 					})

// 					Context("with deleted dataset", func() {
// 						BeforeEach(func() {
// 							dataset.DeletedTime = "2016-09-01T13:00:00Z"
// 							Expect(testMongoCollection.Update(bson.M{"id": dataset.ID}, dataset)).To(Succeed())
// 						})

// 						It("succeeds if it successfully finds the non-deleted user datasets", func() {
// 							Expect(mongoSession.GetDatasetsForUserByID(userID, filter, pagination)).To(ConsistOf([]*upload.Upload{datasetExistingOne, datasetExistingTwo}))
// 						})

// 						It("succeeds if it successfully finds all the user datasets", func() {
// 							filter.Deleted = true
// 							Expect(mongoSession.GetDatasetsForUserByID(userID, filter, pagination)).To(ConsistOf([]*upload.Upload{datasetExistingOne, dataset, datasetExistingTwo}))
// 						})
// 					})
// 				})

// 				Context("GetDatasetByID", func() {
// 					BeforeEach(func() {
// 						dataset.CreatedTime = "2016-09-01T11:00:00Z"
// 						Expect(testMongoCollection.Insert(dataset)).To(Succeed())
// 					})

// 					It("succeeds if it successfully finds the dataset", func() {
// 						Expect(mongoSession.GetDatasetByID(dataset.UploadID)).To(Equal(dataset))
// 					})

// 					It("returns an error if the dataset id is missing", func() {
// 						resultDataset, err := mongoSession.GetDatasetByID("")
// 						Expect(err).To(MatchError("dataset id is missing"))
// 						Expect(resultDataset).To(BeNil())
// 					})

// 					It("returns an error if the session is closed", func() {
// 						mongoSession.Close()
// 						resultDataset, err := mongoSession.GetDatasetByID(dataset.UploadID)
// 						Expect(err).To(MatchError("session closed"))
// 						Expect(resultDataset).To(BeNil())
// 					})

// 					It("returns no dataset successfully if the dataset cannot be found", func() {
// 						resultDataset, err := mongoSession.GetDatasetByID("not-found")
// 						Expect(err).ToNot(HaveOccurred())
// 						Expect(resultDataset).To(BeNil())
// 					})
// 				})

// 				Context("CreateDataset", func() {
// 					It("succeeds if it successfully creates the dataset", func() {
// 						Expect(mongoSession.CreateDataset(dataset)).To(Succeed())
// 					})

// 					It("returns an error if the dataset is missing", func() {
// 						Expect(mongoSession.CreateDataset(nil)).To(MatchError("dataset is missing"))
// 					})

// 					It("returns an error if the user id is missing", func() {
// 						dataset.UserID = ""
// 						Expect(mongoSession.CreateDataset(dataset)).To(MatchError("dataset user id is missing"))
// 					})

// 					It("returns an error if the upload id is missing", func() {
// 						dataset.UploadID = ""
// 						Expect(mongoSession.CreateDataset(dataset)).To(MatchError("dataset upload id is missing"))
// 					})

// 					It("returns an error if the device id is missing (nil)", func() {
// 						dataset.DeviceID = nil
// 						Expect(mongoSession.CreateDataset(dataset)).To(MatchError("dataset device id is missing"))
// 					})

// 					It("returns an error if the device id is missing (empty)", func() {
// 						dataset.DeviceID = pointer.String("")
// 						Expect(mongoSession.CreateDataset(dataset)).To(MatchError("dataset device id is missing"))
// 					})

// 					It("returns an error if the session is closed", func() {
// 						mongoSession.Close()
// 						Expect(mongoSession.CreateDataset(dataset)).To(MatchError("session closed"))
// 					})

// 					It("returns an error if the dataset with the same id already exists", func() {
// 						Expect(mongoSession.CreateDataset(dataset)).To(Succeed())
// 						Expect(mongoSession.CreateDataset(dataset)).To(MatchError("unable to create dataset; dataset already exists"))
// 					})

// 					It("sets the created time", func() {
// 						Expect(mongoSession.CreateDataset(dataset)).To(Succeed())
// 						Expect(dataset.CreatedTime).ToNot(BeEmpty())
// 						Expect(dataset.CreatedUserID).To(BeEmpty())
// 						Expect(dataset.ByUser).To(BeNil())
// 					})

// 					It("has the correct stored datasets", func() {
// 						ValidateDataset(testMongoCollection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, bson.M{}, datasetExistingOther, datasetExistingOne, datasetExistingTwo)
// 						Expect(mongoSession.CreateDataset(dataset)).To(Succeed())
// 						ValidateDataset(testMongoCollection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, bson.M{}, datasetExistingOther, datasetExistingOne, datasetExistingTwo, dataset)
// 					})
// 				})

// 				Context("UpdateDataset", func() {
// 					BeforeEach(func() {
// 						dataset.CreatedTime = "2016-09-01T11:00:00Z"
// 						Expect(testMongoCollection.Insert(dataset)).To(Succeed())
// 					})

// 					Context("with state closed", func() {
// 						BeforeEach(func() {
// 							dataset.State = "closed"
// 						})

// 						It("succeeds if it successfully updates the dataset", func() {
// 							Expect(mongoSession.UpdateDataset(dataset)).To(Succeed())
// 						})

// 						It("returns an error if the dataset is missing", func() {
// 							Expect(mongoSession.UpdateDataset(nil)).To(MatchError("dataset is missing"))
// 						})

// 						It("returns an error if the user id is missing", func() {
// 							dataset.UserID = ""
// 							Expect(mongoSession.UpdateDataset(dataset)).To(MatchError("dataset user id is missing"))
// 						})

// 						It("returns an error if the upload id is missing", func() {
// 							dataset.UploadID = ""
// 							Expect(mongoSession.UpdateDataset(dataset)).To(MatchError("dataset upload id is missing"))
// 						})

// 						It("returns an error if the device id is missing (nil)", func() {
// 							dataset.DeviceID = nil
// 							Expect(mongoSession.UpdateDataset(dataset)).To(MatchError("dataset device id is missing"))
// 						})

// 						It("returns an error if the device id is missing (empty)", func() {
// 							dataset.DeviceID = pointer.String("")
// 							Expect(mongoSession.UpdateDataset(dataset)).To(MatchError("dataset device id is missing"))
// 						})

// 						It("returns an error if the session is closed", func() {
// 							mongoSession.Close()
// 							Expect(mongoSession.UpdateDataset(dataset)).To(MatchError("session closed"))
// 						})

// 						It("returns an error if the dataset with the same user id and upload id does not yet exist", func() {
// 							dataset.UploadID = id.New()
// 							Expect(mongoSession.UpdateDataset(dataset)).To(MatchError("unable to update dataset; not found"))
// 						})
// 					})

// 					It("sets the modified time", func() {
// 						dataset.State = "closed"
// 						Expect(mongoSession.UpdateDataset(dataset)).To(Succeed())
// 						Expect(dataset.ModifiedTime).ToNot(BeEmpty())
// 						Expect(dataset.ModifiedUserID).To(BeEmpty())
// 					})

// 					It("has the correct stored datasets", func() {
// 						ValidateDataset(testMongoCollection, bson.M{}, bson.M{}, datasetExistingOther, datasetExistingOne, datasetExistingTwo, dataset)
// 						ValidateDataset(testMongoCollection, bson.M{"modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, bson.M{})
// 						dataset.State = "closed"
// 						Expect(mongoSession.UpdateDataset(dataset)).To(Succeed())
// 						ValidateDataset(testMongoCollection, bson.M{}, bson.M{}, datasetExistingOther, datasetExistingOne, datasetExistingTwo, dataset)
// 						ValidateDataset(testMongoCollection, bson.M{"modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, bson.M{}, dataset)
// 					})
// 				})

// 				Context("with data", func() {
// 					var datasetExistingOtherData []data.Datum
// 					var datasetExistingOneData []data.Datum
// 					var datasetExistingTwoData []data.Datum
// 					var datasetData []data.Datum

// 					BeforeEach(func() {
// 						dataset.CreatedTime = "2016-09-01T11:00:00Z"
// 						Expect(testMongoCollection.Insert(dataset)).To(Succeed())
// 						datasetExistingOtherData = NewDatasetData(id.New())
// 						Expect(mongoSession.CreateDatasetData(datasetExistingOther, datasetExistingOtherData)).To(Succeed())
// 						datasetExistingOneData = NewDatasetData(deviceID)
// 						Expect(mongoSession.CreateDatasetData(datasetExistingOne, datasetExistingOneData)).To(Succeed())
// 						datasetExistingTwoData = NewDatasetData(deviceID)
// 						Expect(mongoSession.CreateDatasetData(datasetExistingTwo, datasetExistingTwoData)).To(Succeed())
// 						datasetData = NewDatasetData(deviceID)
// 					})

// 					Context("DeleteDataset", func() {
// 						BeforeEach(func() {
// 							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
// 						})

// 						It("succeeds if it successfully deletes the dataset", func() {
// 							Expect(mongoSession.DeleteDataset(dataset)).To(Succeed())
// 						})

// 						It("returns an error if the dataset is missing", func() {
// 							Expect(mongoSession.DeleteDataset(nil)).To(MatchError("dataset is missing"))
// 						})

// 						It("returns an error if the user id is missing", func() {
// 							dataset.UserID = ""
// 							Expect(mongoSession.DeleteDataset(dataset)).To(MatchError("dataset user id is missing"))
// 						})

// 						It("returns an error if the upload id is missing", func() {
// 							dataset.UploadID = ""
// 							Expect(mongoSession.DeleteDataset(dataset)).To(MatchError("dataset upload id is missing"))
// 						})

// 						It("returns an error if the device id is missing (nil)", func() {
// 							dataset.DeviceID = nil
// 							Expect(mongoSession.DeleteDataset(dataset)).To(MatchError("dataset device id is missing"))
// 						})

// 						It("returns an error if the device id is missing (empty)", func() {
// 							dataset.DeviceID = pointer.String("")
// 							Expect(mongoSession.DeleteDataset(dataset)).To(MatchError("dataset device id is missing"))
// 						})

// 						It("returns an error if the session is closed", func() {
// 							mongoSession.Close()
// 							Expect(mongoSession.DeleteDataset(dataset)).To(MatchError("session closed"))
// 						})

// 						It("sets the deleted time on the dataset", func() {
// 							Expect(mongoSession.DeleteDataset(dataset)).To(Succeed())
// 							Expect(dataset.DeletedTime).ToNot(BeEmpty())
// 							Expect(dataset.DeletedUserID).To(BeEmpty())
// 						})

// 						It("has the correct stored datasets", func() {
// 							ValidateDataset(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": true}, "deletedUserId": bson.M{"$exists": false}}, bson.M{})
// 							Expect(mongoSession.DeleteDataset(dataset)).To(Succeed())
// 							ValidateDataset(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": true}, "deletedUserId": bson.M{"$exists": false}}, bson.M{}, dataset)
// 						})

// 						It("has the correct stored dataset data", func() {
// 							ValidateDatasetData(testMongoCollection, bson.M{"uploadId": dataset.UploadID}, bson.M{}, datasetData)
// 							Expect(mongoSession.DeleteDataset(dataset)).To(Succeed())
// 							ValidateDatasetData(testMongoCollection, bson.M{"uploadId": dataset.UploadID}, bson.M{}, []data.Datum{})
// 						})
// 					})

// 					Context("CreateDatasetData", func() {
// 						It("succeeds if it successfully creates the dataset data", func() {
// 							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
// 						})

// 						It("returns an error if the dataset is missing", func() {
// 							Expect(mongoSession.CreateDatasetData(nil, datasetData)).To(MatchError("dataset is missing"))
// 						})

// 						It("returns an error if the dataset data is missing", func() {
// 							Expect(mongoSession.CreateDatasetData(dataset, nil)).To(MatchError("dataset data is missing"))
// 						})

// 						It("returns an error if the user id is missing", func() {
// 							dataset.UserID = ""
// 							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(MatchError("dataset user id is missing"))
// 						})

// 						It("returns an error if the upload id is missing", func() {
// 							dataset.UploadID = ""
// 							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(MatchError("dataset upload id is missing"))
// 						})

// 						It("returns an error if the device id is missing (nil)", func() {
// 							dataset.DeviceID = nil
// 							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(MatchError("dataset device id is missing"))
// 						})

// 						It("returns an error if the device id is missing (empty)", func() {
// 							dataset.DeviceID = pointer.String("")
// 							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(MatchError("dataset device id is missing"))
// 						})

// 						It("returns an error if the session is closed", func() {
// 							mongoSession.Close()
// 							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(MatchError("session closed"))
// 						})

// 						It("sets the user id and upload id on the dataset data to match the dataset", func() {
// 							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
// 							for _, datasetDatum := range datasetData {
// 								baseDatum, ok := datasetDatum.(*types.Base)
// 								Expect(ok).To(BeTrue())
// 								Expect(baseDatum).ToNot(BeNil())
// 								Expect(baseDatum.UserID).To(Equal(dataset.UserID))
// 								Expect(baseDatum.UploadID).To(Equal(dataset.UploadID))
// 							}
// 						})

// 						It("leaves the dataset data not active", func() {
// 							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
// 							for _, datasetDatum := range datasetData {
// 								baseDatum, ok := datasetDatum.(*types.Base)
// 								Expect(ok).To(BeTrue())
// 								Expect(baseDatum).ToNot(BeNil())
// 								Expect(baseDatum.Active).To(BeFalse())
// 							}
// 						})

// 						It("sets the created time on the dataset data", func() {
// 							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
// 							for _, datasetDatum := range datasetData {
// 								baseDatum, ok := datasetDatum.(*types.Base)
// 								Expect(ok).To(BeTrue())
// 								Expect(baseDatum).ToNot(BeNil())
// 								Expect(baseDatum.CreatedTime).ToNot(BeEmpty())
// 								Expect(baseDatum.CreatedUserID).To(BeEmpty())
// 							}
// 						})

// 						It("has the correct stored dataset data", func() {
// 							datasetBeforeCreateData := append(append(datasetExistingOtherData, datasetExistingOneData...), datasetExistingTwoData...)
// 							ValidateDatasetData(testMongoCollection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, bson.M{}, datasetBeforeCreateData)
// 							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
// 							ValidateDatasetData(testMongoCollection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, bson.M{}, append(datasetBeforeCreateData, datasetData...))
// 						})
// 					})

// 					Context("ActivateDatasetData", func() {
// 						BeforeEach(func() {
// 							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
// 						})

// 						It("succeeds if it successfully activates the dataset", func() {
// 							Expect(mongoSession.ActivateDatasetData(dataset)).To(Succeed())
// 						})

// 						It("returns an error if the dataset is missing", func() {
// 							Expect(mongoSession.ActivateDatasetData(nil)).To(MatchError("dataset is missing"))
// 						})

// 						It("returns an error if the user id is missing", func() {
// 							dataset.UserID = ""
// 							Expect(mongoSession.ActivateDatasetData(dataset)).To(MatchError("dataset user id is missing"))
// 						})

// 						It("returns an error if the upload id is missing", func() {
// 							dataset.UploadID = ""
// 							Expect(mongoSession.ActivateDatasetData(dataset)).To(MatchError("dataset upload id is missing"))
// 						})

// 						It("returns an error if the device id is missing (nil)", func() {
// 							dataset.DeviceID = nil
// 							Expect(mongoSession.ActivateDatasetData(dataset)).To(MatchError("dataset device id is missing"))
// 						})

// 						It("returns an error if the device id is missing (empty)", func() {
// 							dataset.DeviceID = pointer.String("")
// 							Expect(mongoSession.ActivateDatasetData(dataset)).To(MatchError("dataset device id is missing"))
// 						})

// 						It("returns an error if the session is closed", func() {
// 							mongoSession.Close()
// 							Expect(mongoSession.ActivateDatasetData(dataset)).To(MatchError("session closed"))
// 						})

// 						It("sets the active on the dataset", func() {
// 							Expect(mongoSession.ActivateDatasetData(dataset)).To(Succeed())
// 							Expect(dataset.Active).To(BeTrue())
// 						})

// 						It("sets the modified time on the dataset", func() {
// 							Expect(mongoSession.ActivateDatasetData(dataset)).To(Succeed())
// 							Expect(dataset.ModifiedTime).ToNot(BeEmpty())
// 							Expect(dataset.ModifiedUserID).To(BeEmpty())
// 						})

// 						It("has the correct stored active dataset", func() {
// 							ValidateDataset(testMongoCollection, bson.M{"_active": true, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, bson.M{})
// 							Expect(mongoSession.ActivateDatasetData(dataset)).To(Succeed())
// 							ValidateDataset(testMongoCollection, bson.M{"_active": true, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, bson.M{}, dataset)
// 						})

// 						It("has the correct stored active dataset data", func() {
// 							ValidateDatasetData(testMongoCollection, bson.M{"_active": true, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, bson.M{}, []data.Datum{})
// 							Expect(mongoSession.ActivateDatasetData(dataset)).To(Succeed())
// 							for _, datasetDatum := range datasetData {
// 								datasetDatum.SetActive(true)
// 							}
// 							ValidateDatasetData(testMongoCollection, bson.M{"_active": true, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}}, bson.M{"modifiedTime": 0}, datasetData)
// 						})
// 					})

// 					Context("ArchiveDeviceDataUsingHashesFromDataset", func() {
// 						var datasetExistingOneDataCloned []data.Datum

// 						BeforeEach(func() {
// 							datasetExistingOneDataCloned = CloneDatasetData(datasetData)
// 							Expect(mongoSession.CreateDatasetData(datasetExistingOne, datasetExistingOneDataCloned)).To(Succeed())
// 							Expect(mongoSession.ActivateDatasetData(datasetExistingOne)).To(Succeed())
// 							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
// 							for _, datasetDatum := range append(datasetExistingOneData, datasetExistingOneDataCloned...) {
// 								datasetDatum.SetActive(true)
// 							}
// 						})

// 						It("succeeds if it successfully archives device data using hashes from dataset", func() {
// 							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataset(dataset)).To(Succeed())
// 						})

// 						It("returns an error if the dataset is missing", func() {
// 							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataset(nil)).To(MatchError("dataset is missing"))
// 						})

// 						It("returns an error if the user id is missing", func() {
// 							dataset.UserID = ""
// 							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataset(dataset)).To(MatchError("dataset user id is missing"))
// 						})

// 						It("returns an error if the upload id is missing", func() {
// 							dataset.UploadID = ""
// 							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataset(dataset)).To(MatchError("dataset upload id is missing"))
// 						})

// 						It("returns an error if the device id is missing (nil)", func() {
// 							dataset.DeviceID = nil
// 							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataset(dataset)).To(MatchError("dataset device id is missing"))
// 						})

// 						It("returns an error if the device id is missing (empty)", func() {
// 							dataset.DeviceID = pointer.String("")
// 							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataset(dataset)).To(MatchError("dataset device id is missing"))
// 						})

// 						It("returns an error if the session is closed", func() {
// 							mongoSession.Close()
// 							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataset(dataset)).To(MatchError("session closed"))
// 						})

// 						It("has the correct stored datasets", func() {
// 							ValidateDataset(testMongoCollection, bson.M{"uploadId": datasetExistingOne.UploadID, "_active": true}, bson.M{}, datasetExistingOne)
// 							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataset(dataset)).To(Succeed())
// 							ValidateDataset(testMongoCollection, bson.M{"uploadId": datasetExistingOne.UploadID, "_active": true}, bson.M{}, datasetExistingOne)
// 						})

// 						It("has the correct stored archived dataset data", func() {
// 							ValidateDatasetData(testMongoCollection, bson.M{"uploadId": datasetExistingOne.UploadID, "_active": false}, bson.M{}, []data.Datum{})
// 							ValidateDatasetData(testMongoCollection, bson.M{"uploadId": datasetExistingOne.UploadID, "_active": true}, bson.M{"modifiedTime": 0}, append(datasetExistingOneData, datasetExistingOneDataCloned...))
// 							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataset(dataset)).To(Succeed())
// 							for _, datasetDatum := range datasetExistingOneDataCloned {
// 								datasetDatum.SetActive(false)
// 							}
// 							ValidateDatasetData(testMongoCollection,
// 								bson.M{"uploadId": datasetExistingOne.UploadID, "_active": true, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}},
// 								bson.M{"modifiedTime": 0},
// 								datasetExistingOneData)
// 							ValidateDatasetData(testMongoCollection,
// 								bson.M{"uploadId": datasetExistingOne.UploadID, "_active": false, "archivedTime": bson.M{"$exists": true}, "archivedDatasetId": dataset.UploadID, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}},
// 								bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0},
// 								datasetExistingOneDataCloned)
// 							ValidateDatasetData(testMongoCollection,
// 								bson.M{"uploadId": dataset.UploadID, "_active": false, "archivedTime": bson.M{"$exists": false}, "archivedDatasetId": bson.M{"$exists": false}},
// 								bson.M{},
// 								datasetData)
// 						})
// 					})

// 					Context("UnarchiveDeviceDataUsingHashesFromDataset", func() {
// 						var datasetExistingTwoDataCloned []data.Datum
// 						var datasetExistingOneDataCloned []data.Datum

// 						BeforeEach(func() {
// 							datasetExistingTwoDataCloned = CloneDatasetData(datasetData)
// 							datasetExistingOneDataCloned = CloneDatasetData(datasetData)
// 							Expect(mongoSession.CreateDatasetData(datasetExistingTwo, datasetExistingTwoDataCloned)).To(Succeed())
// 							Expect(mongoSession.ActivateDatasetData(datasetExistingTwo)).To(Succeed())
// 							Expect(mongoSession.CreateDatasetData(datasetExistingOne, datasetExistingOneDataCloned)).To(Succeed())
// 							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataset(datasetExistingOne)).To(Succeed())
// 							Expect(mongoSession.ActivateDatasetData(datasetExistingOne)).To(Succeed())
// 							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
// 							Expect(mongoSession.ArchiveDeviceDataUsingHashesFromDataset(dataset)).To(Succeed())
// 							Expect(mongoSession.ActivateDatasetData(dataset)).To(Succeed())
// 						})

// 						It("succeeds if it successfully unarchives device data using hashes from dataset", func() {
// 							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataset(dataset)).To(Succeed())
// 						})

// 						It("returns an error if the dataset is missing", func() {
// 							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataset(nil)).To(MatchError("dataset is missing"))
// 						})

// 						It("returns an error if the user id is missing", func() {
// 							dataset.UserID = ""
// 							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataset(dataset)).To(MatchError("dataset user id is missing"))
// 						})

// 						It("returns an error if the upload id is missing", func() {
// 							dataset.UploadID = ""
// 							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataset(dataset)).To(MatchError("dataset upload id is missing"))
// 						})

// 						It("returns an error if the device id is missing (nil)", func() {
// 							dataset.DeviceID = nil
// 							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataset(dataset)).To(MatchError("dataset device id is missing"))
// 						})

// 						It("returns an error if the device id is missing (empty)", func() {
// 							dataset.DeviceID = pointer.String("")
// 							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataset(dataset)).To(MatchError("dataset device id is missing"))
// 						})

// 						It("returns an error if the session is closed", func() {
// 							mongoSession.Close()
// 							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataset(dataset)).To(MatchError("session closed"))
// 						})

// 						It("has the correct stored datasets", func() {
// 							ValidateDataset(testMongoCollection, bson.M{"uploadId": datasetExistingOne.UploadID, "_active": true}, bson.M{}, datasetExistingOne)
// 							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataset(dataset)).To(Succeed())
// 							ValidateDataset(testMongoCollection, bson.M{"uploadId": datasetExistingOne.UploadID, "_active": true}, bson.M{}, datasetExistingOne)
// 						})

// 						It("has the correct stored unarchived dataset data", func() {
// 							for _, datasetDatum := range append(datasetData, datasetExistingOneData...) {
// 								datasetDatum.SetActive(true)
// 							}
// 							ValidateDatasetData(testMongoCollection, bson.M{"uploadId": datasetExistingOne.UploadID, "_active": false}, bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0}, datasetExistingOneDataCloned)
// 							ValidateDatasetData(testMongoCollection, bson.M{"uploadId": datasetExistingOne.UploadID, "_active": true}, bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0}, datasetExistingOneData)
// 							ValidateDatasetData(testMongoCollection, bson.M{"uploadId": dataset.UploadID, "_active": true}, bson.M{"modifiedTime": 0}, datasetData)
// 							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataset(dataset)).To(Succeed())
// 							for _, datasetDatum := range datasetExistingOneDataCloned {
// 								datasetDatum.SetActive(true)
// 							}
// 							ValidateDatasetData(testMongoCollection,
// 								bson.M{"uploadId": datasetExistingOne.UploadID, "_active": true, "archivedTime": bson.M{"$exists": false}, "archivedDatasetId": bson.M{"$exists": false}, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}},
// 								bson.M{"modifiedTime": 0},
// 								append(datasetExistingOneData, datasetExistingOneDataCloned...))
// 							ValidateDatasetData(testMongoCollection,
// 								bson.M{"uploadId": dataset.UploadID, "_active": true, "archivedTime": bson.M{"$exists": false}, "archivedDatasetId": bson.M{"$exists": false}},
// 								bson.M{"modifiedTime": 0},
// 								datasetData)
// 						})

// 						It("has the correct stored datasets if an intermediary is unarchived", func() {
// 							ValidateDataset(testMongoCollection, bson.M{"uploadId": datasetExistingTwo.UploadID, "_active": true}, bson.M{}, datasetExistingTwo)
// 							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataset(datasetExistingOne)).To(Succeed())
// 							ValidateDataset(testMongoCollection, bson.M{"uploadId": datasetExistingTwo.UploadID, "_active": true}, bson.M{}, datasetExistingTwo)
// 						})

// 						It("has the correct stored unarchived dataset data if an intermediary is unarchived", func() {
// 							for _, datasetDatum := range append(datasetExistingOneData, datasetExistingTwoData...) {
// 								datasetDatum.SetActive(true)
// 							}
// 							ValidateDatasetData(testMongoCollection, bson.M{"uploadId": datasetExistingTwo.UploadID, "_active": false}, bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0}, datasetExistingTwoDataCloned)
// 							ValidateDatasetData(testMongoCollection, bson.M{"uploadId": datasetExistingTwo.UploadID, "_active": true}, bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0}, datasetExistingTwoData)
// 							ValidateDatasetData(testMongoCollection, bson.M{"uploadId": datasetExistingOne.UploadID, "_active": true}, bson.M{"modifiedTime": 0}, datasetExistingOneData)
// 							Expect(mongoSession.UnarchiveDeviceDataUsingHashesFromDataset(datasetExistingOne)).To(Succeed())
// 							ValidateDatasetData(testMongoCollection,
// 								bson.M{"uploadId": datasetExistingTwo.UploadID, "_active": true, "archivedTime": bson.M{"$exists": false}, "archivedDatasetId": bson.M{"$exists": false}},
// 								bson.M{"modifiedTime": 0},
// 								datasetExistingTwoData)
// 							ValidateDatasetData(testMongoCollection,
// 								bson.M{"uploadId": datasetExistingTwo.UploadID, "_active": false, "archivedTime": bson.M{"$exists": true}, "archivedDatasetId": dataset.UploadID, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}},
// 								bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0},
// 								datasetExistingTwoDataCloned)
// 							ValidateDatasetData(testMongoCollection,
// 								bson.M{"uploadId": datasetExistingOne.UploadID, "_active": true, "archivedTime": bson.M{"$exists": false}, "archivedDatasetId": bson.M{"$exists": false}},
// 								bson.M{"modifiedTime": 0},
// 								datasetExistingOneData)
// 							ValidateDatasetData(testMongoCollection,
// 								bson.M{"uploadId": datasetExistingOne.UploadID, "_active": false, "archivedTime": bson.M{"$exists": true}, "archivedDatasetId": dataset.UploadID},
// 								bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0},
// 								datasetExistingOneDataCloned)
// 						})
// 					})

// 					Context("DeleteOtherDatasetData", func() {
// 						BeforeEach(func() {
// 							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
// 						})

// 						It("succeeds if it successfully deletes all other dataset data", func() {
// 							Expect(mongoSession.DeleteOtherDatasetData(dataset)).To(Succeed())
// 						})

// 						It("returns an error if the dataset is missing", func() {
// 							Expect(mongoSession.DeleteOtherDatasetData(nil)).To(MatchError("dataset is missing"))
// 						})

// 						It("returns an error if the user id is missing", func() {
// 							dataset.UserID = ""
// 							Expect(mongoSession.DeleteOtherDatasetData(dataset)).To(MatchError("dataset user id is missing"))
// 						})

// 						It("returns an error if the upload id is missing", func() {
// 							dataset.UploadID = ""
// 							Expect(mongoSession.DeleteOtherDatasetData(dataset)).To(MatchError("dataset upload id is missing"))
// 						})

// 						It("returns an error if the device id is missing (nil)", func() {
// 							dataset.DeviceID = nil
// 							Expect(mongoSession.DeleteOtherDatasetData(dataset)).To(MatchError("dataset device id is missing"))
// 						})

// 						It("returns an error if the device id is missing (empty)", func() {
// 							dataset.DeviceID = pointer.String("")
// 							Expect(mongoSession.DeleteOtherDatasetData(dataset)).To(MatchError("dataset device id is missing"))
// 						})

// 						It("returns an error if the session is closed", func() {
// 							mongoSession.Close()
// 							Expect(mongoSession.DeleteOtherDatasetData(dataset)).To(MatchError("session closed"))
// 						})

// 						It("has the correct stored active dataset", func() {
// 							ValidateDataset(testMongoCollection, bson.M{}, bson.M{}, dataset, datasetExistingOther, datasetExistingOne, datasetExistingTwo)
// 							ValidateDataset(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": false}, "deletedUserId": bson.M{"$exists": false}}, bson.M{}, dataset, datasetExistingOther, datasetExistingOne, datasetExistingTwo)
// 							Expect(mongoSession.DeleteOtherDatasetData(dataset)).To(Succeed())
// 							Expect(testMongoCollection.Find(bson.M{"type": "upload"}).Count()).To(Equal(4))
// 							ValidateDataset(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": true}, "deletedUserId": bson.M{"$exists": false}}, bson.M{"deletedTime": 0}, datasetExistingTwo, datasetExistingOne)
// 							ValidateDataset(testMongoCollection, bson.M{"deletedTime": bson.M{"$exists": false}, "deletedUserId": bson.M{"$exists": false}}, bson.M{}, dataset, datasetExistingOther)
// 						})

// 						It("has the correct stored active dataset data", func() {
// 							datasetDataAfterRemoveData := append(datasetData, datasetExistingOtherData...)
// 							datasetDataBeforeRemoveData := append(append(datasetDataAfterRemoveData, datasetExistingOneData...), datasetExistingTwoData...)
// 							ValidateDatasetData(testMongoCollection, bson.M{}, bson.M{}, datasetDataBeforeRemoveData)
// 							Expect(mongoSession.DeleteOtherDatasetData(dataset)).To(Succeed())
// 							ValidateDatasetData(testMongoCollection, bson.M{}, bson.M{"deletedTime": 0}, datasetDataAfterRemoveData)
// 						})
// 					})

// 					Context("DestroyDataForUserByID", func() {
// 						var deleteUserID string
// 						var deleteDeviceID string
// 						var deleteDataset *upload.Upload
// 						var deleteDatasetData []data.Datum

// 						BeforeEach(func() {
// 							Expect(mongoSession.CreateDatasetData(dataset, datasetData)).To(Succeed())
// 							deleteUserID = id.New()
// 							deleteDeviceID = id.New()
// 							deleteDataset = NewDataset(deleteUserID, deleteDeviceID)
// 							deleteDataset.CreatedTime = "2016-09-01T11:00:00Z"
// 							Expect(testMongoCollection.Insert(deleteDataset)).To(Succeed())
// 							deleteDatasetData = NewDatasetData(deleteDeviceID)
// 							Expect(mongoSession.CreateDatasetData(deleteDataset, deleteDatasetData)).To(Succeed())
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

// 						It("has the correct stored datasets", func() {
// 							ValidateDataset(testMongoCollection, bson.M{}, bson.M{}, dataset, datasetExistingOther, datasetExistingOne, datasetExistingTwo, deleteDataset)
// 							Expect(mongoSession.DestroyDataForUserByID(deleteUserID)).To(Succeed())
// 							ValidateDataset(testMongoCollection, bson.M{}, bson.M{}, dataset, datasetExistingOther, datasetExistingOne, datasetExistingTwo)
// 						})

// 						It("has the correct stored dataset data", func() {
// 							datasetDataAfterRemoveData := append(append(append(datasetData, datasetExistingOtherData...), datasetExistingOneData...), datasetExistingTwoData...)
// 							datasetDataBeforeRemoveData := append(datasetDataAfterRemoveData, deleteDatasetData...)
// 							ValidateDatasetData(testMongoCollection, bson.M{}, bson.M{}, datasetDataBeforeRemoveData)
// 							Expect(mongoSession.DestroyDataForUserByID(deleteUserID)).To(Succeed())
// 							ValidateDatasetData(testMongoCollection, bson.M{}, bson.M{}, datasetDataAfterRemoveData)
// 						})
// 					})
// 				})
// 			})
// 		})
// 	})
// })
