package mongo_test

import (
	"context"
	"math/rand"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/data/types/upload"
	dataTypesUploadTest "github.com/tidepool-org/platform/data/types/upload/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/test"
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

func NewDataSetData(deviceID string) data.Data {
	dataSetData := data.Data{}
	for count := 0; count < test.RandomIntFromRange(4, 6); count++ {
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

func CloneDataSetData(dataSetData data.Data) data.Data {
	clonedDataSetData := data.Data{}
	for _, dataSetDatum := range dataSetData {
		if datum, ok := dataSetDatum.(*types.Base); ok {
			clonedDataSetData = append(clonedDataSetData, dataTypesTest.CloneBase(datum))
		}
	}
	return clonedDataSetData
}

func ValidateDataSet(collection *mongo.Collection, query bson.M, filter bson.M, expectedDataSets ...*upload.Upload) {
	query["type"] = "upload"
	filter["_id"] = 0
	var actualDataSets []*upload.Upload
	opts := options.Find().SetProjection(filter)
	cursor, err := collection.Find(context.Background(), query, opts)
	Expect(err).ToNot(HaveOccurred())
	Expect(cursor).ToNot(BeNil())
	Expect(cursor.All(context.Background(), &actualDataSets)).To(Succeed())
	Expect(actualDataSets).To(ConsistOf(DataSetsAsInterface(expectedDataSets)...))
}

func DataSetsAsInterface(dataSets []*upload.Upload) []interface{} {
	var dataSetsAsInterface []interface{}
	for _, dataSet := range dataSets {
		dataSetsAsInterface = append(dataSetsAsInterface, dataSet)
	}
	return dataSetsAsInterface
}

func ValidateDataSetData(collection *mongo.Collection, query bson.M, filter bson.M, expectedDataSetData data.Data) {
	query["type"] = bson.M{"$ne": "upload"}
	filter["_id"] = 0
	filter["revision"] = 0
	var actualDataSetData []bson.M
	opts := options.Find().SetProjection(filter)
	cursor, err := collection.Find(context.Background(), query, opts)
	Expect(err).ToNot(HaveOccurred())
	Expect(cursor).ToNot(BeNil())
	Expect(cursor.All(context.Background(), &actualDataSetData)).To(Succeed())
	Expect(actualDataSetData).To(ConsistOf(DataSetDataAsInterface(expectedDataSetData)...))
}

func DataSetDataAsInterface(dataSetData data.Data) []interface{} {
	var dataSetDataAsInterface []interface{}
	for _, dataSetDatum := range dataSetData {
		dataSetDataAsInterface = append(dataSetDataAsInterface, DataSetDatumAsInterface(dataSetDatum))
	}
	return dataSetDataAsInterface
}

func DataSetDatumAsInterface(dataSetDatum data.Datum) interface{} {
	bites, err := bson.Marshal(dataSetDatum)
	Expect(err).ToNot(HaveOccurred())
	Expect(bites).ToNot(BeNil())
	var dataSetDatumAsInterface bson.M
	Expect(bson.Unmarshal(bites, &dataSetDatumAsInterface)).To(Succeed())
	return dataSetDatumAsInterface
}

var _ = Describe("Mongo", func() {
	var logger *logTest.Logger
	var config *storeStructuredMongo.Config
	var store *dataStoreMongo.Store
	var repository dataStore.DataRepository

	BeforeEach(func() {
		logger = logTest.NewLogger()
		config = storeStructuredMongoTest.NewConfig()
	})

	AfterEach(func() {
		if store != nil {
			store.Terminate(context.Background())
		}
	})

	Context("New", func() {
		It("returns an error if unsuccessful", func() {
			var err error
			store, err = dataStoreMongo.NewStore(nil)
			Expect(err).To(HaveOccurred())
			Expect(store).To(BeNil())
		})

		It("returns a new store and no error if successful", func() {
			var err error
			store, err = dataStoreMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		var collection *mongo.Collection

		BeforeEach(func() {
			var err error
			store, err = dataStoreMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
			collection = store.GetCollection("deviceData")
			Expect(store.EnsureIndexes()).To(Succeed())
		})

		AfterEach(func() {
			if collection != nil {
				collection.Database().Drop(context.Background())
			}
		})

		Context("EnsureIndexes", func() {
			It("returns successfully", func() {
				cursor, err := collection.Indexes().List(context.Background())
				Expect(err).ToNot(HaveOccurred())
				Expect(cursor).ToNot(BeNil())
				var indexes []storeStructuredMongoTest.MongoIndex
				err = cursor.All(context.Background(), &indexes)
				Expect(err).ToNot(HaveOccurred())

				Expect(indexes).To(ConsistOf(
					MatchFields(IgnoreExtras, Fields{
						"Key": Equal(storeStructuredMongoTest.MakeKeySlice("_id")),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("_userId", "_active", "type", "-time")),
						"Background": Equal(true),
						"Name":       Equal("UserIdTypeWeighted_v2"),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("origin.id", "type", "-deletedTime", "_active")),
						"Background": Equal(true),
						"Name":       Equal("OriginId"),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":                     Equal(storeStructuredMongoTest.MakeKeySlice("uploadId")),
						"Background":              Equal(true),
						"Unique":                  Equal(true),
						"Name":                    Equal("UniqueUploadId"),
						"PartialFilterExpression": Equal(bson.D{{Key: "type", Value: "upload"}}),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("uploadId", "type", "-deletedTime", "_active")),
						"Background": Equal(true),
						"Name":       Equal("UploadId"),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("_userId", "deviceId", "type", "_active", "_deduplicator.hash")),
						"Background": Equal(true),
						"Name":       Equal("DeduplicatorHash"),
						"PartialFilterExpression": Equal(bson.D{
							{Key: "_active", Value: true},
							{Key: "_deduplicator.hash", Value: bson.D{{Key: "$exists", Value: true}}},
							{Key: "deviceId", Value: bson.D{{Key: "$exists", Value: true}}},
						}),
					}),
				))
			})
		})

		Context("NewDataRepository", func() {
			It("returns a new repository", func() {
				repository = store.NewDataRepository()
				Expect(repository).ToNot(BeNil())
			})
		})

		Context("with a new repository", func() {
			BeforeEach(func() {
				repository = store.NewDataRepository()
				Expect(repository).ToNot(BeNil())
			})

			AfterEach(func() {
				if repository != nil {
					collection.DeleteMany(context.Background(), bson.D{})
				}
			})

			Context("with persisted data sets", func() {
				var ctx context.Context
				var userID string
				var deviceID string
				var dataSet *upload.Upload
				var dataSetExistingOther *upload.Upload
				var dataSetExistingOne *upload.Upload
				var dataSetExistingTwo *upload.Upload

				preparePersistedDataSets := func() {
					collection = store.GetCollection("deviceData")
					dataSetExistingOther = NewDataSet(userTest.RandomID(), dataTest.NewDeviceID())
					dataSetExistingOther.CreatedTime = pointer.FromString("2016-09-01T12:00:00Z")
					_, err := collection.InsertOne(context.Background(), dataSetExistingOther)
					Expect(err).ToNot(HaveOccurred())
					dataSetExistingOne = NewDataSet(userID, deviceID)
					dataSetExistingOne.CreatedTime = pointer.FromString("2016-09-01T12:30:00Z")
					_, err = collection.InsertOne(context.Background(), dataSetExistingOne)
					Expect(err).ToNot(HaveOccurred())
					dataSetExistingTwo = NewDataSet(userID, deviceID)
					dataSetExistingTwo.CreatedTime = pointer.FromString("2016-09-01T10:00:00Z")
					_, err = collection.InsertOne(context.Background(), dataSetExistingTwo)
					Expect(err).ToNot(HaveOccurred())
				}

				BeforeEach(func() {
					ctx = log.NewContextWithLogger(context.Background(), logger)
					userID = userTest.RandomID()
					deviceID = dataTest.NewDeviceID()
					dataSet = NewDataSet(userID, deviceID)
				})

				Context("GetDataSetsForUserByID", func() {
					var filter *dataStore.Filter
					var pagination *page.Pagination

					BeforeEach(func() {
						dataSet.CreatedTime = pointer.FromString("2016-09-01T11:00:00Z")
						filter = dataStore.NewFilter()
						pagination = page.NewPagination()
					})

					It("returns an error if the user id is missing", func() {
						resultDataSets, err := repository.GetDataSetsForUserByID(ctx, "", filter, pagination)
						Expect(err).To(MatchError("user id is missing"))
						Expect(resultDataSets).To(BeNil())
					})

					It("returns an error if the pagination page is less than minimum", func() {
						pagination.Page = -1
						resultDataSets, err := repository.GetDataSetsForUserByID(ctx, userID, filter, pagination)
						Expect(err).To(MatchError("pagination is invalid; value -1 is not greater than or equal to 0"))
						Expect(resultDataSets).To(BeNil())
					})

					It("returns an error if the pagination size is less than minimum", func() {
						pagination.Size = 0
						resultDataSets, err := repository.GetDataSetsForUserByID(ctx, userID, filter, pagination)
						Expect(err).To(MatchError("pagination is invalid; value 0 is not between 1 and 1000"))
						Expect(resultDataSets).To(BeNil())
					})

					It("returns an error if the pagination size is greater than maximum", func() {
						pagination.Size = 1001
						resultDataSets, err := repository.GetDataSetsForUserByID(ctx, userID, filter, pagination)
						Expect(err).To(MatchError("pagination is invalid; value 1001 is not between 1 and 1000"))
						Expect(resultDataSets).To(BeNil())
					})

					Context("with database access", func() {
						BeforeEach(func() {
							preparePersistedDataSets()
							_, err := collection.InsertOne(context.Background(), dataSet)
							Expect(err).ToNot(HaveOccurred())
						})

						It("succeeds if it successfully finds the user data sets", func() {
							Expect(repository.GetDataSetsForUserByID(ctx, userID, filter, pagination)).To(ConsistOf([]*upload.Upload{dataSetExistingOne, dataSet, dataSetExistingTwo}))
						})

						It("succeeds if the filter is not specified", func() {
							Expect(repository.GetDataSetsForUserByID(ctx, userID, nil, pagination)).To(ConsistOf([]*upload.Upload{dataSetExistingOne, dataSet, dataSetExistingTwo}))
						})

						It("succeeds if the pagination is not specified", func() {
							Expect(repository.GetDataSetsForUserByID(ctx, userID, filter, nil)).To(ConsistOf([]*upload.Upload{dataSetExistingOne, dataSet, dataSetExistingTwo}))
						})

						It("succeeds if the pagination size is not default", func() {
							pagination.Size = 2
							Expect(repository.GetDataSetsForUserByID(ctx, userID, filter, pagination)).To(ConsistOf([]*upload.Upload{dataSetExistingOne, dataSet}))
						})

						It("succeeds if the pagination page and size is not default", func() {
							pagination.Page = 1
							pagination.Size = 2
							Expect(repository.GetDataSetsForUserByID(ctx, userID, filter, pagination)).To(ConsistOf([]*upload.Upload{dataSetExistingTwo}))
						})

						It("succeeds if it successfully does not find another user data sets", func() {
							resultDataSets, err := repository.GetDataSetsForUserByID(ctx, userTest.RandomID(), filter, pagination)
							Expect(err).ToNot(HaveOccurred())
							Expect(resultDataSets).ToNot(BeNil())
							Expect(resultDataSets).To(BeEmpty())
						})

						Context("with deleted data set", func() {
							BeforeEach(func() {
								dataSet.DeletedTime = pointer.FromString("2016-09-01T13:00:00Z")
								result := collection.FindOneAndReplace(context.Background(), bson.M{"id": dataSet.ID}, dataSet)
								Expect(result.Err()).ToNot(HaveOccurred())
							})

							It("succeeds if it successfully finds the non-deleted user data sets", func() {
								Expect(repository.GetDataSetsForUserByID(ctx, userID, filter, pagination)).To(ConsistOf([]*upload.Upload{dataSetExistingOne, dataSetExistingTwo}))
							})

							It("succeeds if it successfully finds all the user data sets", func() {
								filter.Deleted = true
								Expect(repository.GetDataSetsForUserByID(ctx, userID, filter, pagination)).To(ConsistOf([]*upload.Upload{dataSetExistingOne, dataSet, dataSetExistingTwo}))
							})
						})
					})
				})

				Context("GetDataSetByID", func() {
					BeforeEach(func() {
						dataSet.CreatedTime = pointer.FromString("2016-09-01T11:00:00Z")
						err := repository.EnsureIndexes()
						Expect(err).ToNot(HaveOccurred())
					})

					It("returns an error if the data set id is missing", func() {
						resultDataSet, err := repository.GetDataSetByID(ctx, "")
						Expect(err).To(MatchError("data set id is missing"))
						Expect(resultDataSet).To(BeNil())
					})

					Context("with database access", func() {
						BeforeEach(func() {
							preparePersistedDataSets()
							_, err := collection.InsertOne(context.Background(), dataSet)
							Expect(err).ToNot(HaveOccurred())
						})

						It("succeeds if it successfully finds the data set", func() {
							Expect(repository.GetDataSetByID(ctx, *dataSet.UploadID)).To(Equal(dataSet))
						})

						It("returns no data set successfully if the data set cannot be found", func() {
							resultDataSet, err := repository.GetDataSetByID(ctx, "not-found")
							Expect(err).ToNot(HaveOccurred())
							Expect(resultDataSet).To(BeNil())
						})
					})
				})

				Context("CreateDataSet", func() {
					It("returns an error if the data set is missing", func() {
						Expect(repository.CreateDataSet(ctx, nil)).To(MatchError("data set is missing"))
					})

					It("returns an error if the user id is missing", func() {
						dataSet.UserID = nil
						Expect(repository.CreateDataSet(ctx, dataSet)).To(MatchError("data set user id is missing"))
					})

					It("returns an error if the user id is empty", func() {
						dataSet.UserID = pointer.FromString("")
						Expect(repository.CreateDataSet(ctx, dataSet)).To(MatchError("data set user id is empty"))
					})

					It("returns an error if the upload id is missing", func() {
						dataSet.UploadID = nil
						Expect(repository.CreateDataSet(ctx, dataSet)).To(MatchError("data set upload id is missing"))
					})

					It("returns an error if the upload id is empty", func() {
						dataSet.UploadID = pointer.FromString("")
						Expect(repository.CreateDataSet(ctx, dataSet)).To(MatchError("data set upload id is empty"))
					})

					Context("with database access", func() {
						BeforeEach(func() {
							preparePersistedDataSets()
						})

						It("succeeds if it successfully creates the data set", func() {
							Expect(repository.CreateDataSet(ctx, dataSet)).To(Succeed())
						})

						It("returns an error if the data set with the same id already exists", func() {
							Expect(repository.CreateDataSet(ctx, dataSet)).To(Succeed())
							Expect(repository.CreateDataSet(ctx, dataSet)).To(MatchError("unable to create data set; data set already exists"))
						})

						It("returns an error if the data set with the same uploadId (but different userId) already exists", func() {
							dataSet.UserID = pointer.FromString("differentUser")
							Expect(repository.CreateDataSet(ctx, dataSet)).To(Succeed())
							Expect(repository.CreateDataSet(ctx, dataSet)).To(MatchError("unable to create data set; data set already exists"))
							dataSet.UserID = pointer.FromString("")
						})

						It("sets the created time", func() {
							Expect(repository.CreateDataSet(ctx, dataSet)).To(Succeed())
							Expect(dataSet.CreatedTime).ToNot(BeNil())
							Expect(*dataSet.CreatedTime).ToNot(BeEmpty())
							Expect(dataSet.CreatedUserID).To(BeNil())
							Expect(dataSet.ByUser).To(BeNil())
						})

						It("has the correct stored data sets", func() {
							ValidateDataSet(collection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, bson.M{}, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo)
							Expect(repository.CreateDataSet(ctx, dataSet)).To(Succeed())
							ValidateDataSet(collection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, bson.M{}, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo, dataSet)
						})
					})
				})

				Context("UpdateDataSet", func() {
					var id string
					var update *data.DataSetUpdate

					BeforeEach(func() {
						id = data.NewSetID()
						update = data.NewDataSetUpdate()
					})

					It("returns an error if the context is missing", func() {
						result, err := repository.UpdateDataSet(nil, id, update)
						Expect(err).To(MatchError("context is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error if the id is missing", func() {
						id = ""
						result, err := repository.UpdateDataSet(ctx, id, update)
						Expect(err).To(MatchError("id is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error if the id is invalid", func() {
						id = "invalid"
						result, err := repository.UpdateDataSet(ctx, id, update)
						Expect(err).To(MatchError("id is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error if the update is missing", func() {
						result, err := repository.UpdateDataSet(ctx, id, nil)
						Expect(err).To(MatchError("update is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error if the update is invalid", func() {
						update.DeviceID = pointer.FromString("")
						result, err := repository.UpdateDataSet(ctx, id, update)
						Expect(err).To(MatchError("update is invalid; value is empty"))
						Expect(result).To(BeNil())
					})

					Context("with database access", func() {
						BeforeEach(func() {
							preparePersistedDataSets()
							dataSet.State = pointer.FromString("open")
							_, err := collection.InsertOne(context.Background(), dataSet)
							Expect(err).ToNot(HaveOccurred())
							id = *dataSet.UploadID
						})

						AfterEach(func() {
							logger.AssertDebug("UpdateDataSet", log.Fields{"id": id, "update": update})
						})

						Context("with updates", func() {
							// TODO

							It("returns nil when the id does not exist", func() {
								id = dataTest.RandomSetID()
								Expect(repository.UpdateDataSet(ctx, id, update)).To(BeNil())
							})
						})

						Context("without updates", func() {
							BeforeEach(func() {
								update = data.NewDataSetUpdate()
							})

							// TODO

							It("returns nil when the id does not exist", func() {
								id = dataTest.RandomSetID()
								Expect(repository.UpdateDataSet(ctx, id, update)).To(BeNil())
							})
						})

						It("has the correct stored data sets", func() {
							ValidateDataSet(collection, bson.M{}, bson.M{}, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo, dataSet)
							ValidateDataSet(collection, bson.M{"modifiedTime": bson.M{"$exists": true}}, bson.M{})
							update = data.NewDataSetUpdate()
							update.State = pointer.FromString("closed")
							result, err := repository.UpdateDataSet(ctx, id, update)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).ToNot(BeNil())
							Expect(result.State).ToNot(BeNil())
							Expect(*result.State).To(Equal("closed"))
							ValidateDataSet(collection, bson.M{}, bson.M{}, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo, result)
							ValidateDataSet(collection, bson.M{"modifiedTime": bson.M{"$exists": true}}, bson.M{}, result)
						})
					})
				})

				Context("with persisted data set data", func() {
					var dataSetExistingOtherData data.Data
					var dataSetExistingOneData data.Data
					var dataSetExistingTwoData data.Data
					var dataSetData data.Data

					preparePersistedDataSetsData := func() {
						preparePersistedDataSets()
						_, err := collection.InsertOne(context.Background(), dataSet)
						Expect(err).ToNot(HaveOccurred())
						Expect(repository.CreateDataSetData(ctx, dataSetExistingOther, dataSetExistingOtherData)).To(Succeed())
						Expect(repository.CreateDataSetData(ctx, dataSetExistingOne, dataSetExistingOneData)).To(Succeed())
						Expect(repository.CreateDataSetData(ctx, dataSetExistingTwo, dataSetExistingTwoData)).To(Succeed())
					}

					BeforeEach(func() {
						dataSet.CreatedTime = pointer.FromString("2016-09-01T11:00:00Z")
						dataSetExistingOtherData = NewDataSetData(dataTest.NewDeviceID())
						dataSetExistingOneData = NewDataSetData(deviceID)
						dataSetExistingTwoData = NewDataSetData(deviceID)
						dataSetData = NewDataSetData(deviceID)
					})

					Context("DeleteDataSet", func() {
						It("returns an error if the data set is missing", func() {
							Expect(repository.DeleteDataSet(ctx, nil)).To(MatchError("data set is missing"))
						})

						It("returns an error if the user id is missing", func() {
							dataSet.UserID = nil
							Expect(repository.DeleteDataSet(ctx, dataSet)).To(MatchError("data set user id is missing"))
						})

						It("returns an error if the user id is empty", func() {
							dataSet.UserID = pointer.FromString("")
							Expect(repository.DeleteDataSet(ctx, dataSet)).To(MatchError("data set user id is empty"))
						})

						It("returns an error if the upload id is missing", func() {
							dataSet.UploadID = nil
							Expect(repository.DeleteDataSet(ctx, dataSet)).To(MatchError("data set upload id is missing"))
						})

						It("returns an error if the upload id is empty", func() {
							dataSet.UploadID = pointer.FromString("")
							Expect(repository.DeleteDataSet(ctx, dataSet)).To(MatchError("data set upload id is empty"))
						})

						Context("with database access", func() {
							BeforeEach(func() {
								preparePersistedDataSetsData()
								Expect(repository.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
							})

							It("succeeds if it successfully deletes the data set", func() {
								Expect(repository.DeleteDataSet(ctx, dataSet)).To(Succeed())
							})

							It("sets the deleted time on the data set", func() {
								Expect(repository.DeleteDataSet(ctx, dataSet)).To(Succeed())
								Expect(dataSet.DeletedTime).ToNot(BeNil())
								Expect(*dataSet.DeletedTime).ToNot(BeEmpty())
								Expect(dataSet.DeletedUserID).To(BeNil())
							})

							It("has the correct stored data sets", func() {
								ValidateDataSet(collection, bson.M{"deletedTime": bson.M{"$exists": true}, "deletedUserId": bson.M{"$exists": false}}, bson.M{})
								Expect(repository.DeleteDataSet(ctx, dataSet)).To(Succeed())
								ValidateDataSet(collection, bson.M{"deletedTime": bson.M{"$exists": true}, "deletedUserId": bson.M{"$exists": false}}, bson.M{}, dataSet)
							})

							It("has the correct stored data set data", func() {
								ValidateDataSetData(collection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSetData)
								Expect(repository.DeleteDataSet(ctx, dataSet)).To(Succeed())
								ValidateDataSetData(collection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, data.Data{})
							})
						})
					})

					Context("CreateDataSetData", func() {
						It("returns an error if the data set is missing", func() {
							Expect(repository.CreateDataSetData(ctx, nil, dataSetData)).To(MatchError("data set is missing"))
						})

						It("returns an error if the data set data is missing", func() {
							Expect(repository.CreateDataSetData(ctx, dataSet, nil)).To(MatchError("data set data is missing"))
						})

						It("returns an error if the user id is missing", func() {
							dataSet.UserID = nil
							Expect(repository.CreateDataSetData(ctx, dataSet, dataSetData)).To(MatchError("data set user id is missing"))
						})

						It("returns an error if the user id is empty", func() {
							dataSet.UserID = pointer.FromString("")
							Expect(repository.CreateDataSetData(ctx, dataSet, dataSetData)).To(MatchError("data set user id is empty"))
						})

						It("returns an error if the upload id is missing", func() {
							dataSet.UploadID = nil
							Expect(repository.CreateDataSetData(ctx, dataSet, dataSetData)).To(MatchError("data set upload id is missing"))
						})

						It("returns an error if the upload id is empty", func() {
							dataSet.UploadID = pointer.FromString("")
							Expect(repository.CreateDataSetData(ctx, dataSet, dataSetData)).To(MatchError("data set upload id is empty"))
						})

						Context("with database access", func() {
							BeforeEach(func() {
								preparePersistedDataSetsData()
							})

							It("succeeds if it successfully creates the data set data", func() {
								Expect(repository.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
							})

							It("succeeds if data set data is empty", func() {
								dataSetData = data.Data{}
								Expect(repository.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
							})

							It("sets the user id and upload id on the data set data to match the data set", func() {
								Expect(repository.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
								for _, dataSetDatum := range dataSetData {
									baseDatum, ok := dataSetDatum.(*types.Base)
									Expect(ok).To(BeTrue())
									Expect(baseDatum).ToNot(BeNil())
									Expect(baseDatum.UserID).To(Equal(dataSet.UserID))
									Expect(baseDatum.UploadID).To(Equal(dataSet.UploadID))
								}
							})

							It("leaves the data set data not active", func() {
								Expect(repository.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
								for _, dataSetDatum := range dataSetData {
									baseDatum, ok := dataSetDatum.(*types.Base)
									Expect(ok).To(BeTrue())
									Expect(baseDatum).ToNot(BeNil())
									Expect(baseDatum.Active).To(BeFalse())
								}
							})

							It("sets the created time on the data set data", func() {
								Expect(repository.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
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
								ValidateDataSetData(collection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, bson.M{}, dataSetBeforeCreateData)
								Expect(repository.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
								ValidateDataSetData(collection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, bson.M{}, append(dataSetBeforeCreateData, dataSetData...))
							})
						})
					})

					Context("with selected data set data", func() {
						var selectors *data.Selectors
						var selectedDataSetData data.Data
						var unselectedDataSetData data.Data

						BeforeEach(func() {
							selectors = dataTest.RandomSelectors()
							selectedDataSetData = data.Data{}
							unselectedDataSetData = data.Data{}
							selectedCount := test.RandomIntFromRange(1, len(dataSetData)-1)
							for index, dataSetDataIndex := range rand.Perm(len(dataSetData)) {
								if index < selectedCount {
									selectedDataSetData = append(selectedDataSetData, dataSetData[dataSetDataIndex])
								} else {
									unselectedDataSetData = append(unselectedDataSetData, dataSetData[dataSetDataIndex])
								}
							}
						})

						Context("ActivateDataSetData", func() {
							commonAssertions := func() {
								It("returns an error when the context is missing", func() {
									Expect(repository.ActivateDataSetData(nil, dataSet, selectors)).To(MatchError("context is missing"))
								})

								It("returns an error when the data set is missing", func() {
									Expect(repository.ActivateDataSetData(ctx, nil, selectors)).To(MatchError("data set is missing"))
								})

								It("returns an error when the user id is missing", func() {
									dataSet.UserID = nil
									Expect(repository.ActivateDataSetData(ctx, dataSet, selectors)).To(MatchError("data set user id is missing"))
								})

								It("returns an error when the user id is empty", func() {
									dataSet.UserID = pointer.FromString("")
									Expect(repository.ActivateDataSetData(ctx, dataSet, selectors)).To(MatchError("data set user id is empty"))
								})

								It("returns an error when the upload id is missing", func() {
									dataSet.UploadID = nil
									Expect(repository.ActivateDataSetData(ctx, dataSet, selectors)).To(MatchError("data set upload id is missing"))
								})

								It("returns an error when the upload id is empty", func() {
									dataSet.UploadID = pointer.FromString("")
									Expect(repository.ActivateDataSetData(ctx, dataSet, selectors)).To(MatchError("data set upload id is empty"))
								})
							}

							selectorAssertions := func() {
								Context("with database access", func() {
									BeforeEach(func() {
										preparePersistedDataSetsData()
										Expect(repository.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
										ValidateDataSetData(collection, bson.M{"_active": true}, bson.M{}, data.Data{})
										ValidateDataSet(collection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSet)
									})

									It("succeeds and has the correct stored active data set data", func() {
										Expect(repository.ActivateDataSetData(ctx, dataSet, selectors)).To(Succeed())
										selectedDataSetData.SetActive(true)
										ValidateDataSetData(collection, bson.M{"_active": true}, bson.M{"modifiedTime": 0}, selectedDataSetData)
										ValidateDataSet(collection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSet)
									})

									It("succeeds with no changes when the data set user id is different", func() {
										dataSet.UserID = pointer.FromString(userTest.RandomID())
										Expect(repository.ActivateDataSetData(ctx, dataSet, selectors)).To(Succeed())
										ValidateDataSetData(collection, bson.M{"_active": true}, bson.M{}, data.Data{})
									})

									It("succeeds with no changes when the data set upload id is different", func() {
										dataSet.UploadID = pointer.FromString(dataTest.RandomSetID())
										Expect(repository.ActivateDataSetData(ctx, dataSet, selectors)).To(Succeed())
										ValidateDataSetData(collection, bson.M{"_active": true}, bson.M{}, data.Data{})
									})
								})
							}

							Context("with selectors missing", func() {
								BeforeEach(func() {
									selectors = nil
									selectedDataSetData = dataSetData
									unselectedDataSetData = data.Data{}
								})

								commonAssertions()
								selectorAssertions()
							})

							Context("with selectors empty", func() {
								BeforeEach(func() {
									selectors = data.NewSelectors()
								})

								commonAssertions()

								It("returns an error", func() {
									errorsTest.ExpectEqual(repository.ActivateDataSetData(ctx, dataSet, selectors), errors.New("selectors is invalid"))
								})
							})

							Context("with selectors by id", func() {
								BeforeEach(func() {
									for _, datum := range selectedDataSetData {
										*selectors = append(*selectors, &data.Selector{ID: datum.(*types.Base).ID})
									}
								})

								commonAssertions()
								selectorAssertions()
							})

							Context("with selectors by origin id", func() {
								BeforeEach(func() {
									for _, datum := range selectedDataSetData {
										*selectors = append(*selectors, &data.Selector{Origin: &data.SelectorOrigin{ID: datum.(*types.Base).Origin.ID}})
									}
								})

								commonAssertions()
								selectorAssertions()
							})

							Context("with selectors by both id and origin id", func() {
								BeforeEach(func() {
									for _, datum := range selectedDataSetData {
										*selectors = append(*selectors, &data.Selector{ID: datum.(*types.Base).ID, Origin: &data.SelectorOrigin{ID: datum.(*types.Base).Origin.ID}})
									}
								})

								commonAssertions()

								It("returns an error", func() {
									errorsTest.ExpectEqual(repository.ActivateDataSetData(ctx, dataSet, selectors), errors.New("selectors is invalid"))
								})
							})

							Context("with selectors with neither id nor origin id", func() {
								BeforeEach(func() {
									for range selectedDataSetData {
										*selectors = append(*selectors, &data.Selector{})
									}
								})

								commonAssertions()

								It("returns an error", func() {
									errorsTest.ExpectEqual(repository.ActivateDataSetData(ctx, dataSet, selectors), errors.New("selectors is invalid"))
								})
							})
						})

						Context("ArchiveDataSetData", func() {
							commonAssertions := func() {
								It("returns an error when the context is missing", func() {
									Expect(repository.ArchiveDataSetData(nil, dataSet, selectors)).To(MatchError("context is missing"))
								})

								It("returns an error when the data set is missing", func() {
									Expect(repository.ArchiveDataSetData(ctx, nil, selectors)).To(MatchError("data set is missing"))
								})

								It("returns an error when the user id is missing", func() {
									dataSet.UserID = nil
									Expect(repository.ArchiveDataSetData(ctx, dataSet, selectors)).To(MatchError("data set user id is missing"))
								})

								It("returns an error when the user id is empty", func() {
									dataSet.UserID = pointer.FromString("")
									Expect(repository.ArchiveDataSetData(ctx, dataSet, selectors)).To(MatchError("data set user id is empty"))
								})

								It("returns an error when the upload id is missing", func() {
									dataSet.UploadID = nil
									Expect(repository.ArchiveDataSetData(ctx, dataSet, selectors)).To(MatchError("data set upload id is missing"))
								})

								It("returns an error when the upload id is empty", func() {
									dataSet.UploadID = pointer.FromString("")
									Expect(repository.ArchiveDataSetData(ctx, dataSet, selectors)).To(MatchError("data set upload id is empty"))
								})
							}

							selectorAssertions := func() {
								Context("with database access", func() {
									BeforeEach(func() {
										preparePersistedDataSetsData()
										Expect(repository.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
										Expect(repository.ActivateDataSetData(ctx, dataSet, nil)).To(Succeed())
										dataSetData.SetActive(true)
										ValidateDataSetData(collection, bson.M{"_active": false, "uploadId": dataSet.UploadID}, bson.M{"modifiedTime": 0}, data.Data{})
										ValidateDataSet(collection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSet)
									})

									It("succeeds and has the correct stored active data set data", func() {
										Expect(repository.ArchiveDataSetData(ctx, dataSet, selectors)).To(Succeed())
										selectedDataSetData.SetActive(false)
										ValidateDataSetData(collection, bson.M{"_active": false, "uploadId": dataSet.UploadID}, bson.M{"archivedTime": 0, "modifiedTime": 0}, selectedDataSetData)
										ValidateDataSet(collection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSet)
									})

									It("succeeds with no changes when the data set user id is different", func() {
										dataSet.UserID = pointer.FromString(userTest.RandomID())
										Expect(repository.ArchiveDataSetData(ctx, dataSet, selectors)).To(Succeed())
										ValidateDataSetData(collection, bson.M{"_active": false, "uploadId": dataSet.UploadID}, bson.M{"modifiedTime": 0}, data.Data{})
									})

									It("succeeds with no changes when the data set upload id is different", func() {
										dataSetUploadID := dataSet.UploadID
										dataSet.UploadID = pointer.FromString(dataTest.RandomSetID())
										Expect(repository.ArchiveDataSetData(ctx, dataSet, selectors)).To(Succeed())
										ValidateDataSetData(collection, bson.M{"_active": false, "uploadId": dataSetUploadID}, bson.M{"modifiedTime": 0}, data.Data{})
									})
								})
							}

							Context("with selectors missing", func() {
								BeforeEach(func() {
									selectors = nil
									selectedDataSetData = dataSetData
									unselectedDataSetData = data.Data{}
								})

								commonAssertions()
								selectorAssertions()
							})

							Context("with selectors empty", func() {
								BeforeEach(func() {
									selectors = data.NewSelectors()
								})

								commonAssertions()

								It("returns an error", func() {
									errorsTest.ExpectEqual(repository.ArchiveDataSetData(ctx, dataSet, selectors), errors.New("selectors is invalid"))
								})
							})

							Context("with selectors by id", func() {
								BeforeEach(func() {
									for _, datum := range selectedDataSetData {
										*selectors = append(*selectors, &data.Selector{ID: datum.(*types.Base).ID})
									}
								})

								commonAssertions()
								selectorAssertions()
							})

							Context("with selectors by origin id", func() {
								BeforeEach(func() {
									for _, datum := range selectedDataSetData {
										*selectors = append(*selectors, &data.Selector{Origin: &data.SelectorOrigin{ID: datum.(*types.Base).Origin.ID}})
									}
								})

								commonAssertions()
								selectorAssertions()
							})

							Context("with selectors by both id and origin id", func() {
								BeforeEach(func() {
									for _, datum := range selectedDataSetData {
										*selectors = append(*selectors, &data.Selector{ID: datum.(*types.Base).ID, Origin: &data.SelectorOrigin{ID: datum.(*types.Base).Origin.ID}})
									}
								})

								commonAssertions()

								It("returns an error", func() {
									errorsTest.ExpectEqual(repository.ArchiveDataSetData(ctx, dataSet, selectors), errors.New("selectors is invalid"))
								})
							})

							Context("with selectors with neither id nor origin id", func() {
								BeforeEach(func() {
									for range selectedDataSetData {
										*selectors = append(*selectors, &data.Selector{})
									}
								})

								commonAssertions()

								It("returns an error", func() {
									errorsTest.ExpectEqual(repository.ArchiveDataSetData(ctx, dataSet, selectors), errors.New("selectors is invalid"))
								})
							})
						})

						Context("DeleteDataSetData", func() {
							commonAssertions := func() {
								It("returns an error when the context is missing", func() {
									Expect(repository.DeleteDataSetData(nil, dataSet, selectors)).To(MatchError("context is missing"))
								})

								It("returns an error when the data set is missing", func() {
									Expect(repository.DeleteDataSetData(ctx, nil, selectors)).To(MatchError("data set is missing"))
								})

								It("returns an error when the user id is missing", func() {
									dataSet.UserID = nil
									Expect(repository.DeleteDataSetData(ctx, dataSet, selectors)).To(MatchError("data set user id is missing"))
								})

								It("returns an error when the user id is empty", func() {
									dataSet.UserID = pointer.FromString("")
									Expect(repository.DeleteDataSetData(ctx, dataSet, selectors)).To(MatchError("data set user id is empty"))
								})

								It("returns an error when the upload id is missing", func() {
									dataSet.UploadID = nil
									Expect(repository.DeleteDataSetData(ctx, dataSet, selectors)).To(MatchError("data set upload id is missing"))
								})

								It("returns an error when the upload id is empty", func() {
									dataSet.UploadID = pointer.FromString("")
									Expect(repository.DeleteDataSetData(ctx, dataSet, selectors)).To(MatchError("data set upload id is empty"))
								})
							}

							selectorAssertions := func() {
								Context("with database access", func() {
									BeforeEach(func() {
										preparePersistedDataSetsData()
										Expect(repository.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
										ValidateDataSetData(collection, bson.M{"deletedTime": bson.M{"$exists": true}}, bson.M{"modifiedTime": 0}, data.Data{})
										ValidateDataSet(collection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSet)
									})

									It("succeeds and has the correct stored active data set data", func() {
										Expect(repository.DeleteDataSetData(ctx, dataSet, selectors)).To(Succeed())
										selectedDataSetData.SetActive(false)
										ValidateDataSetData(collection, bson.M{"deletedTime": bson.M{"$exists": true}}, bson.M{"archivedTime": 0, "deletedTime": 0, "modifiedTime": 0}, selectedDataSetData)
										ValidateDataSet(collection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSet)
									})

									It("succeeds with no changes when the data set user id is different", func() {
										dataSet.UserID = pointer.FromString(userTest.RandomID())
										Expect(repository.DeleteDataSetData(ctx, dataSet, selectors)).To(Succeed())
										ValidateDataSetData(collection, bson.M{"deletedTime": bson.M{"$exists": true}}, bson.M{"modifiedTime": 0}, data.Data{})
									})

									It("succeeds with no changes when the data set upload id is different", func() {
										dataSet.UploadID = pointer.FromString(dataTest.RandomSetID())
										Expect(repository.DeleteDataSetData(ctx, dataSet, selectors)).To(Succeed())
										ValidateDataSetData(collection, bson.M{"deletedTime": bson.M{"$exists": true}}, bson.M{"modifiedTime": 0}, data.Data{})
									})
								})
							}

							Context("with selectors missing", func() {
								BeforeEach(func() {
									selectors = nil
									selectedDataSetData = dataSetData
									unselectedDataSetData = data.Data{}
								})

								commonAssertions()
								selectorAssertions()
							})

							Context("with selectors empty", func() {
								BeforeEach(func() {
									selectors = data.NewSelectors()
								})

								commonAssertions()

								It("returns an error", func() {
									errorsTest.ExpectEqual(repository.DeleteDataSetData(ctx, dataSet, selectors), errors.New("selectors is invalid"))
								})
							})

							Context("with selectors by id", func() {
								BeforeEach(func() {
									for _, datum := range selectedDataSetData {
										*selectors = append(*selectors, &data.Selector{ID: datum.(*types.Base).ID})
									}
								})

								commonAssertions()
								selectorAssertions()
							})

							Context("with selectors by origin id", func() {
								BeforeEach(func() {
									for _, datum := range selectedDataSetData {
										*selectors = append(*selectors, &data.Selector{Origin: &data.SelectorOrigin{ID: datum.(*types.Base).Origin.ID}})
									}
								})

								commonAssertions()
								selectorAssertions()
							})

							Context("with selectors by both id and origin id", func() {
								BeforeEach(func() {
									for _, datum := range selectedDataSetData {
										*selectors = append(*selectors, &data.Selector{ID: datum.(*types.Base).ID, Origin: &data.SelectorOrigin{ID: datum.(*types.Base).Origin.ID}})
									}
								})

								commonAssertions()

								It("returns an error", func() {
									errorsTest.ExpectEqual(repository.DeleteDataSetData(ctx, dataSet, selectors), errors.New("selectors is invalid"))
								})
							})

							Context("with selectors with neither id nor origin id", func() {
								BeforeEach(func() {
									for range selectedDataSetData {
										*selectors = append(*selectors, &data.Selector{})
									}
								})

								commonAssertions()

								It("returns an error", func() {
									errorsTest.ExpectEqual(repository.DeleteDataSetData(ctx, dataSet, selectors), errors.New("selectors is invalid"))
								})
							})
						})

						Context("DestroyDeletedDataSetData", func() {
							commonAssertions := func() {
								It("returns an error when the context is missing", func() {
									Expect(repository.DestroyDeletedDataSetData(nil, dataSet, selectors)).To(MatchError("context is missing"))
								})

								It("returns an error when the data set is missing", func() {
									Expect(repository.DestroyDeletedDataSetData(ctx, nil, selectors)).To(MatchError("data set is missing"))
								})

								It("returns an error when the user id is missing", func() {
									dataSet.UserID = nil
									Expect(repository.DestroyDeletedDataSetData(ctx, dataSet, selectors)).To(MatchError("data set user id is missing"))
								})

								It("returns an error when the user id is empty", func() {
									dataSet.UserID = pointer.FromString("")
									Expect(repository.DestroyDeletedDataSetData(ctx, dataSet, selectors)).To(MatchError("data set user id is empty"))
								})

								It("returns an error when the upload id is missing", func() {
									dataSet.UploadID = nil
									Expect(repository.DestroyDeletedDataSetData(ctx, dataSet, selectors)).To(MatchError("data set upload id is missing"))
								})

								It("returns an error when the upload id is empty", func() {
									dataSet.UploadID = pointer.FromString("")
									Expect(repository.DestroyDeletedDataSetData(ctx, dataSet, selectors)).To(MatchError("data set upload id is empty"))
								})
							}

							selectorAssertions := func() {
								Context("with database access", func() {
									BeforeEach(func() {
										preparePersistedDataSetsData()
										Expect(repository.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
										Expect(repository.DeleteDataSetData(ctx, dataSet, nil)).To(Succeed())
										ValidateDataSetData(collection, bson.M{"deletedTime": bson.M{"$exists": true}}, bson.M{"archivedTime": 0, "deletedTime": 0, "modifiedTime": 0}, dataSetData)
										ValidateDataSet(collection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSet)
									})

									It("succeeds and has the correct stored active data set data", func() {
										Expect(repository.DestroyDeletedDataSetData(ctx, dataSet, selectors)).To(Succeed())
										ValidateDataSetData(collection, bson.M{"deletedTime": bson.M{"$exists": true}}, bson.M{"archivedTime": 0, "deletedTime": 0, "modifiedTime": 0}, unselectedDataSetData)
										ValidateDataSet(collection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSet)
									})

									It("succeeds with no changes when the data set user id is different", func() {
										dataSet.UserID = pointer.FromString(userTest.RandomID())
										Expect(repository.DestroyDeletedDataSetData(ctx, dataSet, selectors)).To(Succeed())
										ValidateDataSetData(collection, bson.M{"deletedTime": bson.M{"$exists": true}}, bson.M{"archivedTime": 0, "deletedTime": 0, "modifiedTime": 0}, dataSetData)
									})

									It("succeeds with no changes when the data set upload id is different", func() {
										dataSet.UploadID = pointer.FromString(dataTest.RandomSetID())
										Expect(repository.DestroyDeletedDataSetData(ctx, dataSet, selectors)).To(Succeed())
										ValidateDataSetData(collection, bson.M{"deletedTime": bson.M{"$exists": true}}, bson.M{"archivedTime": 0, "deletedTime": 0, "modifiedTime": 0}, dataSetData)
									})
								})
							}

							Context("with selectors missing", func() {
								BeforeEach(func() {
									selectors = nil
									selectedDataSetData = dataSetData
									unselectedDataSetData = data.Data{}
								})

								commonAssertions()
								selectorAssertions()
							})

							Context("with selectors empty", func() {
								BeforeEach(func() {
									selectors = data.NewSelectors()
								})

								commonAssertions()

								It("returns an error", func() {
									errorsTest.ExpectEqual(repository.DestroyDeletedDataSetData(ctx, dataSet, selectors), errors.New("selectors is invalid"))
								})
							})

							Context("with selectors by id", func() {
								BeforeEach(func() {
									for _, datum := range selectedDataSetData {
										*selectors = append(*selectors, &data.Selector{ID: datum.(*types.Base).ID})
									}
								})

								commonAssertions()
								selectorAssertions()
							})

							Context("with selectors by origin id", func() {
								BeforeEach(func() {
									for _, datum := range selectedDataSetData {
										*selectors = append(*selectors, &data.Selector{Origin: &data.SelectorOrigin{ID: datum.(*types.Base).Origin.ID}})
									}
								})

								commonAssertions()
								selectorAssertions()
							})

							Context("with selectors by both id and origin id", func() {
								BeforeEach(func() {
									for _, datum := range selectedDataSetData {
										*selectors = append(*selectors, &data.Selector{ID: datum.(*types.Base).ID, Origin: &data.SelectorOrigin{ID: datum.(*types.Base).Origin.ID}})
									}
								})

								commonAssertions()

								It("returns an error", func() {
									errorsTest.ExpectEqual(repository.DestroyDeletedDataSetData(ctx, dataSet, selectors), errors.New("selectors is invalid"))
								})
							})

							Context("with selectors with neither id nor origin id", func() {
								BeforeEach(func() {
									for range selectedDataSetData {
										*selectors = append(*selectors, &data.Selector{})
									}
								})

								commonAssertions()

								It("returns an error", func() {
									errorsTest.ExpectEqual(repository.DestroyDeletedDataSetData(ctx, dataSet, selectors), errors.New("selectors is invalid"))
								})
							})
						})

						Context("DestroyDataSetData", func() {
							commonAssertions := func() {
								It("returns an error when the context is missing", func() {
									Expect(repository.DestroyDataSetData(nil, dataSet, selectors)).To(MatchError("context is missing"))
								})

								It("returns an error when the data set is missing", func() {
									Expect(repository.DestroyDataSetData(ctx, nil, selectors)).To(MatchError("data set is missing"))
								})

								It("returns an error when the user id is missing", func() {
									dataSet.UserID = nil
									Expect(repository.DestroyDataSetData(ctx, dataSet, selectors)).To(MatchError("data set user id is missing"))
								})

								It("returns an error when the user id is empty", func() {
									dataSet.UserID = pointer.FromString("")
									Expect(repository.DestroyDataSetData(ctx, dataSet, selectors)).To(MatchError("data set user id is empty"))
								})

								It("returns an error when the upload id is missing", func() {
									dataSet.UploadID = nil
									Expect(repository.DestroyDataSetData(ctx, dataSet, selectors)).To(MatchError("data set upload id is missing"))
								})

								It("returns an error when the upload id is empty", func() {
									dataSet.UploadID = pointer.FromString("")
									Expect(repository.DestroyDataSetData(ctx, dataSet, selectors)).To(MatchError("data set upload id is empty"))
								})
							}

							selectorAssertions := func() {
								Context("with database access", func() {
									BeforeEach(func() {
										preparePersistedDataSetsData()
										Expect(repository.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
										ValidateDataSetData(collection, bson.M{"uploadId": dataSet.UploadID}, bson.M{"modifiedTime": 0}, dataSetData)
										ValidateDataSet(collection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSet)
									})

									It("succeeds and has the correct stored active data set data", func() {
										Expect(repository.DestroyDataSetData(ctx, dataSet, selectors)).To(Succeed())
										ValidateDataSetData(collection, bson.M{"uploadId": dataSet.UploadID}, bson.M{"modifiedTime": 0}, unselectedDataSetData)
										ValidateDataSet(collection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSet)
									})

									It("succeeds with no changes when the data set user id is different", func() {
										dataSet.UserID = pointer.FromString(userTest.RandomID())
										Expect(repository.DestroyDataSetData(ctx, dataSet, selectors)).To(Succeed())
										ValidateDataSetData(collection, bson.M{"uploadId": dataSet.UploadID}, bson.M{"modifiedTime": 0}, dataSetData)
									})

									It("succeeds with no changes when the data set upload id is different", func() {
										dataSetUploadID := dataSet.UploadID
										dataSet.UploadID = pointer.FromString(dataTest.RandomSetID())
										Expect(repository.DestroyDataSetData(ctx, dataSet, selectors)).To(Succeed())
										ValidateDataSetData(collection, bson.M{"uploadId": dataSetUploadID}, bson.M{"modifiedTime": 0}, dataSetData)
									})
								})
							}

							Context("with selectors missing", func() {
								BeforeEach(func() {
									selectors = nil
									selectedDataSetData = dataSetData
									unselectedDataSetData = data.Data{}
								})

								commonAssertions()
								selectorAssertions()
							})

							Context("with selectors empty", func() {
								BeforeEach(func() {
									selectors = data.NewSelectors()
								})

								commonAssertions()

								It("returns an error", func() {
									errorsTest.ExpectEqual(repository.DestroyDataSetData(ctx, dataSet, selectors), errors.New("selectors is invalid"))
								})
							})

							Context("with selectors by id", func() {
								BeforeEach(func() {
									for _, datum := range selectedDataSetData {
										*selectors = append(*selectors, &data.Selector{ID: datum.(*types.Base).ID})
									}
								})

								commonAssertions()
								selectorAssertions()
							})

							Context("with selectors by origin id", func() {
								BeforeEach(func() {
									for _, datum := range selectedDataSetData {
										*selectors = append(*selectors, &data.Selector{Origin: &data.SelectorOrigin{ID: datum.(*types.Base).Origin.ID}})
									}
								})

								commonAssertions()
								selectorAssertions()
							})

							Context("with selectors by both id and origin id", func() {
								BeforeEach(func() {
									for _, datum := range selectedDataSetData {
										*selectors = append(*selectors, &data.Selector{ID: datum.(*types.Base).ID, Origin: &data.SelectorOrigin{ID: datum.(*types.Base).Origin.ID}})
									}
								})

								commonAssertions()

								It("returns an error", func() {
									errorsTest.ExpectEqual(repository.DestroyDataSetData(ctx, dataSet, selectors), errors.New("selectors is invalid"))
								})
							})

							Context("with selectors with neither id nor origin id", func() {
								BeforeEach(func() {
									for range selectedDataSetData {
										*selectors = append(*selectors, &data.Selector{})
									}
								})

								commonAssertions()

								It("returns an error", func() {
									errorsTest.ExpectEqual(repository.DestroyDataSetData(ctx, dataSet, selectors), errors.New("selectors is invalid"))
								})
							})
						})
					})

					Context("ArchiveDeviceDataUsingHashesFromDataSet", func() {
						It("returns an error if the data set is missing", func() {
							Expect(repository.ArchiveDeviceDataUsingHashesFromDataSet(ctx, nil)).To(MatchError("data set is missing"))
						})

						It("returns an error if the user id is missing", func() {
							dataSet.UserID = nil
							Expect(repository.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("data set user id is missing"))
						})

						It("returns an error if the user id is empty", func() {
							dataSet.UserID = pointer.FromString("")
							Expect(repository.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("data set user id is empty"))
						})

						It("returns an error if the upload id is missing", func() {
							dataSet.UploadID = nil
							Expect(repository.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("data set upload id is missing"))
						})

						It("returns an error if the upload id is empty", func() {
							dataSet.UploadID = pointer.FromString("")
							Expect(repository.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("data set upload id is empty"))
						})

						It("returns an error if the device id is missing (nil)", func() {
							dataSet.DeviceID = nil
							Expect(repository.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("data set device id is missing"))
						})

						It("returns an error if the device id is missing (empty)", func() {
							dataSet.DeviceID = pointer.FromString("")
							Expect(repository.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("data set device id is missing"))
						})

						Context("with database access", func() {
							var dataSetExistingOneDataCloned data.Data

							BeforeEach(func() {
								preparePersistedDataSetsData()
								dataSetExistingOneDataCloned = CloneDataSetData(dataSetData)
								Expect(repository.CreateDataSetData(ctx, dataSetExistingOne, dataSetExistingOneDataCloned)).To(Succeed())
								Expect(repository.ActivateDataSetData(ctx, dataSetExistingOne, nil)).To(Succeed())
								Expect(repository.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
								for _, dataSetDatum := range append(dataSetExistingOneData, dataSetExistingOneDataCloned...) {
									dataSetDatum.SetActive(true)
								}
							})

							It("succeeds if it successfully archives device data using hashes from data set", func() {
								Expect(repository.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(Succeed())
							})

							It("has the correct stored data sets", func() {
								ValidateDataSet(collection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true}, bson.M{}, dataSetExistingOne)
								Expect(repository.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(Succeed())
								ValidateDataSet(collection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true}, bson.M{}, dataSetExistingOne)
							})

							It("has the correct stored archived data set data", func() {
								ValidateDataSetData(collection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": false}, bson.M{}, data.Data{})
								ValidateDataSetData(collection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true}, bson.M{"modifiedTime": 0}, append(dataSetExistingOneData, dataSetExistingOneDataCloned...))
								Expect(repository.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(Succeed())
								for _, dataSetDatum := range dataSetExistingOneDataCloned {
									dataSetDatum.SetActive(false)
								}
								ValidateDataSetData(collection,
									bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}},
									bson.M{"modifiedTime": 0},
									dataSetExistingOneData)
								ValidateDataSetData(collection,
									bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": false, "archivedTime": bson.M{"$exists": true}, "archivedDatasetId": dataSet.UploadID, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}},
									bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0},
									dataSetExistingOneDataCloned)
								ValidateDataSetData(collection,
									bson.M{"uploadId": dataSet.UploadID, "_active": false, "archivedTime": bson.M{"$exists": false}, "archivedDatasetId": bson.M{"$exists": false}},
									bson.M{},
									dataSetData)
							})
						})
					})

					Context("UnarchiveDeviceDataUsingHashesFromDataSet", func() {
						It("returns an error if the data set is missing", func() {
							Expect(repository.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, nil)).To(MatchError("data set is missing"))
						})

						It("returns an error if the user id is missing", func() {
							dataSet.UserID = nil
							Expect(repository.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("data set user id is missing"))
						})

						It("returns an error if the user id is empty", func() {
							dataSet.UserID = pointer.FromString("")
							Expect(repository.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("data set user id is empty"))
						})

						It("returns an error if the upload id is missing", func() {
							dataSet.UploadID = nil
							Expect(repository.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("data set upload id is missing"))
						})

						It("returns an error if the upload id is empty", func() {
							dataSet.UploadID = pointer.FromString("")
							Expect(repository.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("data set upload id is empty"))
						})

						It("returns an error if the device id is missing (nil)", func() {
							dataSet.DeviceID = nil
							Expect(repository.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("data set device id is missing"))
						})

						It("returns an error if the device id is missing (empty)", func() {
							dataSet.DeviceID = pointer.FromString("")
							Expect(repository.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(MatchError("data set device id is missing"))
						})

						Context("with database access", func() {
							var dataSetExistingTwoDataCloned data.Data
							var dataSetExistingOneDataCloned data.Data

							BeforeEach(func() {
								preparePersistedDataSetsData()
								dataSetExistingTwoDataCloned = CloneDataSetData(dataSetData)
								dataSetExistingOneDataCloned = CloneDataSetData(dataSetData)
								Expect(repository.CreateDataSetData(ctx, dataSetExistingTwo, dataSetExistingTwoDataCloned)).To(Succeed())
								Expect(repository.ActivateDataSetData(ctx, dataSetExistingTwo, nil)).To(Succeed())
								Expect(repository.CreateDataSetData(ctx, dataSetExistingOne, dataSetExistingOneDataCloned)).To(Succeed())
								Expect(repository.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSetExistingOne)).To(Succeed())
								Expect(repository.ActivateDataSetData(ctx, dataSetExistingOne, nil)).To(Succeed())
								Expect(repository.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
								Expect(repository.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(Succeed())
								Expect(repository.ActivateDataSetData(ctx, dataSet, nil)).To(Succeed())
							})

							It("succeeds if it successfully unarchives device data using hashes from data set", func() {
								Expect(repository.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(Succeed())
							})

							It("has the correct stored data sets", func() {
								ValidateDataSet(collection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true}, bson.M{}, dataSetExistingOne)
								Expect(repository.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(Succeed())
								ValidateDataSet(collection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true}, bson.M{}, dataSetExistingOne)
							})

							It("has the correct stored unarchived data set data", func() {
								for _, dataSetDatum := range append(dataSetData, dataSetExistingOneData...) {
									dataSetDatum.SetActive(true)
								}
								ValidateDataSetData(collection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": false}, bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0}, dataSetExistingOneDataCloned)
								ValidateDataSetData(collection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true}, bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0}, dataSetExistingOneData)
								ValidateDataSetData(collection, bson.M{"uploadId": dataSet.UploadID, "_active": true}, bson.M{"modifiedTime": 0}, dataSetData)
								Expect(repository.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(Succeed())
								for _, dataSetDatum := range dataSetExistingOneDataCloned {
									dataSetDatum.SetActive(true)
								}
								ValidateDataSetData(collection,
									bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true, "archivedTime": bson.M{"$exists": false}, "archivedDatasetId": bson.M{"$exists": false}, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}},
									bson.M{"modifiedTime": 0},
									append(dataSetExistingOneData, dataSetExistingOneDataCloned...))
								ValidateDataSetData(collection,
									bson.M{"uploadId": dataSet.UploadID, "_active": true, "archivedTime": bson.M{"$exists": false}, "archivedDatasetId": bson.M{"$exists": false}},
									bson.M{"modifiedTime": 0},
									dataSetData)
							})

							It("has the correct stored data sets if an intermediary is unarchived", func() {
								ValidateDataSet(collection, bson.M{"uploadId": dataSetExistingTwo.UploadID, "_active": true}, bson.M{}, dataSetExistingTwo)
								Expect(repository.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSetExistingOne)).To(Succeed())
								ValidateDataSet(collection, bson.M{"uploadId": dataSetExistingTwo.UploadID, "_active": true}, bson.M{}, dataSetExistingTwo)
							})

							It("has the correct stored unarchived data set data if an intermediary is unarchived", func() {
								for _, dataSetDatum := range append(dataSetExistingOneData, dataSetExistingTwoData...) {
									dataSetDatum.SetActive(true)
								}
								ValidateDataSetData(collection, bson.M{"uploadId": dataSetExistingTwo.UploadID, "_active": false}, bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0}, dataSetExistingTwoDataCloned)
								ValidateDataSetData(collection, bson.M{"uploadId": dataSetExistingTwo.UploadID, "_active": true}, bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0}, dataSetExistingTwoData)
								ValidateDataSetData(collection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true}, bson.M{"modifiedTime": 0}, dataSetExistingOneData)
								Expect(repository.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSetExistingOne)).To(Succeed())
								ValidateDataSetData(collection,
									bson.M{"uploadId": dataSetExistingTwo.UploadID, "_active": true, "archivedTime": bson.M{"$exists": false}, "archivedDatasetId": bson.M{"$exists": false}},
									bson.M{"modifiedTime": 0},
									dataSetExistingTwoData)
								ValidateDataSetData(collection,
									bson.M{"uploadId": dataSetExistingTwo.UploadID, "_active": false, "archivedTime": bson.M{"$exists": true}, "archivedDatasetId": dataSet.UploadID, "modifiedTime": bson.M{"$exists": true}, "modifiedUserId": bson.M{"$exists": false}},
									bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0},
									dataSetExistingTwoDataCloned)
								ValidateDataSetData(collection,
									bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true, "archivedTime": bson.M{"$exists": false}, "archivedDatasetId": bson.M{"$exists": false}},
									bson.M{"modifiedTime": 0},
									dataSetExistingOneData)
								ValidateDataSetData(collection,
									bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": false, "archivedTime": bson.M{"$exists": true}, "archivedDatasetId": dataSet.UploadID},
									bson.M{"archivedTime": 0, "archivedDatasetId": 0, "modifiedTime": 0},
									dataSetExistingOneDataCloned)
							})
						})
					})

					Context("DeleteOtherDataSetData", func() {
						It("returns an error if the data set is missing", func() {
							Expect(repository.DeleteOtherDataSetData(ctx, nil)).To(MatchError("data set is missing"))
						})

						It("returns an error if the user id is missing", func() {
							dataSet.UserID = nil
							Expect(repository.DeleteOtherDataSetData(ctx, dataSet)).To(MatchError("data set user id is missing"))
						})

						It("returns an error if the user id is empty", func() {
							dataSet.UserID = pointer.FromString("")
							Expect(repository.DeleteOtherDataSetData(ctx, dataSet)).To(MatchError("data set user id is empty"))
						})

						It("returns an error if the upload id is missing", func() {
							dataSet.UploadID = nil
							Expect(repository.DeleteOtherDataSetData(ctx, dataSet)).To(MatchError("data set upload id is missing"))
						})

						It("returns an error if the upload id is empty", func() {
							dataSet.UploadID = pointer.FromString("")
							Expect(repository.DeleteOtherDataSetData(ctx, dataSet)).To(MatchError("data set upload id is empty"))
						})

						It("returns an error if the device id is missing (nil)", func() {
							dataSet.DeviceID = nil
							Expect(repository.DeleteOtherDataSetData(ctx, dataSet)).To(MatchError("data set device id is missing"))
						})

						It("returns an error if the device id is missing (empty)", func() {
							dataSet.DeviceID = pointer.FromString("")
							Expect(repository.DeleteOtherDataSetData(ctx, dataSet)).To(MatchError("data set device id is missing"))
						})

						Context("with database access", func() {
							BeforeEach(func() {
								preparePersistedDataSetsData()
								Expect(repository.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
							})

							It("succeeds if it successfully deletes all other data set data", func() {
								Expect(repository.DeleteOtherDataSetData(ctx, dataSet)).To(Succeed())
							})

							It("has the correct stored active data set", func() {
								ValidateDataSet(collection, bson.M{}, bson.M{}, dataSet, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo)
								ValidateDataSet(collection, bson.M{"deletedTime": bson.M{"$exists": false}, "deletedUserId": bson.M{"$exists": false}}, bson.M{}, dataSet, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo)
								Expect(repository.DeleteOtherDataSetData(ctx, dataSet)).To(Succeed())
								Expect(collection.CountDocuments(ctx, bson.M{"type": "upload"})).To(Equal(int64(4)))
								ValidateDataSet(collection, bson.M{"deletedTime": bson.M{"$exists": true}, "deletedUserId": bson.M{"$exists": false}}, bson.M{"deletedTime": 0}, dataSetExistingTwo, dataSetExistingOne)
								ValidateDataSet(collection, bson.M{"deletedTime": bson.M{"$exists": false}, "deletedUserId": bson.M{"$exists": false}}, bson.M{}, dataSet, dataSetExistingOther)
							})

							It("has the correct stored active data set data", func() {
								dataSetDataAfterRemoveData := append(dataSetData, dataSetExistingOtherData...)
								dataSetDataBeforeRemoveData := append(append(dataSetDataAfterRemoveData, dataSetExistingOneData...), dataSetExistingTwoData...)
								ValidateDataSetData(collection, bson.M{}, bson.M{}, dataSetDataBeforeRemoveData)
								Expect(repository.DeleteOtherDataSetData(ctx, dataSet)).To(Succeed())
								ValidateDataSetData(collection, bson.M{}, bson.M{"deletedTime": 0}, dataSetDataAfterRemoveData)
							})
						})
					})

					Context("DestroyDataForUserByID", func() {
						var destroyUserID string

						BeforeEach(func() {
							destroyUserID = userTest.RandomID()
						})

						It("returns an error if the user id is missing", func() {
							Expect(repository.DestroyDataForUserByID(ctx, "")).To(MatchError("user id is missing"))
						})

						Context("with database access", func() {
							var destroyDeviceID string
							var destroyDataSet *upload.Upload
							var destroyDataSetData data.Data

							BeforeEach(func() {
								preparePersistedDataSetsData()
								Expect(repository.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
								destroyDeviceID = dataTest.NewDeviceID()
								destroyDataSet = NewDataSet(destroyUserID, destroyDeviceID)
								destroyDataSet.CreatedTime = pointer.FromString("2016-09-01T11:00:00Z")
								_, err := collection.InsertOne(context.Background(), destroyDataSet)
								Expect(err).ToNot(HaveOccurred())
								destroyDataSetData = NewDataSetData(destroyDeviceID)
								Expect(repository.CreateDataSetData(ctx, destroyDataSet, destroyDataSetData)).To(Succeed())
							})

							It("succeeds if it successfully destroys all data for user by id", func() {
								Expect(repository.DestroyDataForUserByID(ctx, destroyUserID)).To(Succeed())
							})

							It("has the correct stored data sets", func() {
								ValidateDataSet(collection, bson.M{}, bson.M{}, dataSet, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo, destroyDataSet)
								Expect(repository.DestroyDataForUserByID(ctx, destroyUserID)).To(Succeed())
								ValidateDataSet(collection, bson.M{}, bson.M{}, dataSet, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo)
							})

							It("has the correct stored data set data", func() {
								dataSetDataAfterRemoveData := append(append(append(dataSetData, dataSetExistingOtherData...), dataSetExistingOneData...), dataSetExistingTwoData...)
								dataSetDataBeforeRemoveData := append(dataSetDataAfterRemoveData, destroyDataSetData...)
								ValidateDataSetData(collection, bson.M{}, bson.M{}, dataSetDataBeforeRemoveData)
								Expect(repository.DestroyDataForUserByID(ctx, destroyUserID)).To(Succeed())
								ValidateDataSetData(collection, bson.M{}, bson.M{}, dataSetDataAfterRemoveData)
							})
						})
					})
				})
			})
		})
	})
})
