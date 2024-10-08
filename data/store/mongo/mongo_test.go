package mongo_test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/alerts"
	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types"
	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	"github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/data/types/insulin"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/data/types/upload"
	dataTypesUploadTest "github.com/tidepool-org/platform/data/types/upload/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
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
	dataSet.DeviceID = pointer.FromAny(deviceID)
	dataSet.Location.GPS.Origin.Time = nil
	dataSet.ModifiedTime = nil
	dataSet.ModifiedUserID = nil
	dataSet.Origin.Time = nil
	dataSet.UserID = pointer.FromAny(userID)
	return dataSet
}

func NewDataSetData(deviceID string) data.Data {
	requiredRecords := test.RandomIntFromRange(6, 8)
	typ := test.RandomChoice([]string{"cbg", "smbg", "basal", "bolus"})
	t := time.Now().UTC().AddDate(0, 0, -10)
	var dataSetData = make([]data.Datum, requiredRecords)
	for count := 0; count < requiredRecords; count++ {
		datum := dataTypesTest.RandomBase()
		datum.Type = typ
		datum.Time = pointer.FromAny(t.Add(time.Duration(count) * time.Hour))
		datum.Active = false
		datum.ArchivedDataSetID = nil
		datum.ArchivedTime = nil
		datum.CreatedTime = nil
		datum.CreatedUserID = nil
		datum.DeletedTime = nil
		datum.DeletedUserID = nil
		datum.DeviceID = pointer.FromAny(deviceID)
		datum.ModifiedTime = nil
		datum.ModifiedUserID = nil
		dataSetData[count] = datum
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
	actualDataSets := getDataSets(collection, query, filter)
	clearModifiedTimes(expectedDataSets...)
	clearModifiedTimes(actualDataSets...)
	Expect(actualDataSets).To(ConsistOf(DataSetsAsInterface(expectedDataSets)...))
}

func ValidateDataSetWithModifiedThreshold(collection *mongo.Collection, query bson.M, filter bson.M, modifiedTimeThreshold time.Duration, expectedDataSets ...*upload.Upload) {
	actualDataSets := getDataSets(collection, query, filter)
	// Check the modified times manually
	// Double Loop / O(M*N) but the number of entries is small so don't care.
	for _, actual := range actualDataSets {
		for _, expected := range expectedDataSets {
			if *expected.ID == *actual.ID {
				Expect(expected.ModifiedTime).ToNot(BeNil())
				Expect(actual.ModifiedTime).ToNot(BeNil())
				Expect(*expected.ModifiedTime).Should(BeTemporally("~", *actual.ModifiedTime, modifiedTimeThreshold))
			}
		}
	}

	// clear modified times like the regular ValidateDataSet function. The
	// normal Expect compares the bson.M representation. Because the
	// modifiedTimes between the actual and expected may be different by a
	// few milliseconds because the time it is set in the Repo and the time
	// it is actually compared are not necessarily at the exact same time
	// (hence the need to use the time threshold above to check modifiedTimes).
	clearModifiedTimes(expectedDataSets...)
	clearModifiedTimes(actualDataSets...)
	Expect(actualDataSets).To(ConsistOf(DataSetsAsInterface(expectedDataSets)...))
}

// clearModifiedTimes sets all the supplied data's ModifiedTime to nil.
func clearModifiedTimes(dataSets ...*upload.Upload) {
	for _, dataSet := range dataSets {
		dataSet.SetModifiedTime(nil)
	}
}

func getDataSets(collection *mongo.Collection, query bson.M, filter bson.M) []*upload.Upload {
	query["type"] = "upload"
	filter["_id"] = 0
	var actualDataSets []*upload.Upload
	opts := options.Find().SetProjection(filter)
	cursor, err := collection.Find(context.Background(), query, opts)
	Expect(err).ToNot(HaveOccurred())
	Expect(cursor).ToNot(BeNil())
	Expect(cursor.All(context.Background(), &actualDataSets)).To(Succeed())
	return actualDataSets
}

func DataSetsAsInterface(dataSets []*upload.Upload) []interface{} {
	var dataSetsAsInterface []interface{}
	for _, dataSet := range dataSets {
		dataSetsAsInterface = append(dataSetsAsInterface, dataSet)
	}
	return dataSetsAsInterface
}

func ValidateDataSetData(collection *mongo.Collection, query bson.M, filter bson.M, expectedDataSetData data.Data) {
	actualDataSetData := getDataSetData(collection, query, filter)
	// delete/remove "modifiedTime" from comparison - this is because even if
	// it is omitted from projection, the actual struct may have had
	// its .ModifiedTime property set in a Repository's method.
	for _, datum := range actualDataSetData {
		delete(datum, "modifiedTime")
	}
	Expect(actualDataSetData).To(ConsistOf(DataSetDataAsInterface(expectedDataSetData)...))
}

func ValidateDataSetDataWithModifiedThreshold(collection *mongo.Collection, query bson.M, filter bson.M, modifiedTimeThreshold time.Duration, expectedDataSetData data.Data) {
	actualDataSetData := getDataSetData(collection, query, filter)

	// The main comparison between datasets does a json comparison between
	// each object in a slice. However, this does a deep equal and certain
	// times may not be 100% the same due to when it was updated in the repo
	// vs when it was defined in a before step, thus the need to compare time
	// thresholds.
	actualTimes := make([]time.Time, 0, len(actualDataSetData))
	for _, actual := range actualDataSetData {
		modifiedTimeRaw, ok := actual["modifiedTime"]
		if !ok {
			continue
		}
		modifiedTime, ok := modifiedTimeRaw.(primitive.DateTime)
		if !ok {
			continue
		}
		actualTimes = append(actualTimes, modifiedTime.Time())
	}
	expectedTimeMatchers := make([]interface{}, 0, len(expectedDataSetData))
	for _, expected := range expectedDataSetData {
		baseDatum, ok := expected.(*types.Base)
		Expect(ok).To(BeTrue())

		if baseDatum.ModifiedTime == nil {
			continue
		}
		expectedTimeMatchers = append(expectedTimeMatchers, BeTemporally("~", *baseDatum.ModifiedTime, modifiedTimeThreshold))
	}
	Expect(actualTimes).To(ConsistOf(expectedTimeMatchers))

	// delete/remove "modifiedTime" from comparison - this is because even if
	// it is omitted from projection, the actual struct may have had
	// its .ModifiedTime property set in a Repository's method.
	for _, datum := range actualDataSetData {
		delete(datum, "modifiedTime")
	}
	Expect(actualDataSetData).To(ConsistOf(DataSetDataAsInterface(expectedDataSetData)...))
}

func getDataSetData(collection *mongo.Collection, query bson.M, filter bson.M) []bson.M {
	query["type"] = bson.M{"$ne": "upload"}
	filter["_id"] = 0
	filter["revision"] = 0
	var actualDataSetData []bson.M
	opts := options.Find().SetProjection(filter)
	cursor, err := collection.Find(context.Background(), query, opts)
	Expect(err).ToNot(HaveOccurred())
	Expect(cursor).ToNot(BeNil())
	Expect(cursor.All(context.Background(), &actualDataSetData)).To(Succeed())
	return actualDataSetData
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
	// We don't want to check the modifiedTime as from the time it's called to
	// the time it's checked the time will (likely) be different. Instead, we
	// compare them and make sure they're within a time.Duration threshold of
	// each other outside of this function.
	delete(dataSetDatumAsInterface, "modifiedTime")
	return dataSetDatumAsInterface
}

var _ = Describe("Mongo", Label("mongodb", "slow", "integration"), func() {
	var repository dataStore.DataRepository
	var summaryRepository dataStore.SummaryRepository
	var alertsRepository alerts.Repository
	var logger = logTest.NewLogger()
	var store *dataStoreMongo.Store

	BeforeEach(func() {
		store = GetSuiteStore()
	})

	Context("New", func() {
		It("returns an error if unsuccessful", func() {
			testStore, err := dataStoreMongo.NewStore(nil)
			Expect(err).To(HaveOccurred())
			Expect(testStore).To(BeNil())
		})

		It("returns a new store and no error if successful", func() {
			testStore, err := dataStoreMongo.NewStore(storeStructuredMongoTest.NewConfig())
			Expect(err).ToNot(HaveOccurred())
			Expect(testStore).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		var collection *mongo.Collection
		var dataSetCollection *mongo.Collection
		var summaryCollection *mongo.Collection
		var alertsCollection *mongo.Collection
		var collectionsOnce sync.Once

		BeforeEach(func() {
			collectionsOnce.Do(func() {
				collection = store.GetCollection("deviceData")
				dataSetCollection = store.GetCollection("deviceDataSets")
				summaryCollection = store.GetCollection("summary")
				alertsCollection = store.GetCollection("alerts")
			})
		})

		AfterEach(func() {
			var err error
			ctx := context.Background()
			all := bson.D{}
			_, err = collection.DeleteMany(ctx, all)
			Expect(err).To(Succeed())
			_, err = dataSetCollection.DeleteMany(ctx, all)
			Expect(err).To(Succeed())
			_, err = summaryCollection.DeleteMany(ctx, all)
			Expect(err).To(Succeed())
			_, err = alertsCollection.DeleteMany(ctx, all)
			Expect(err).To(Succeed())
		})

		Context("EnsureIndexes", func() {
			It("deviceData indexes return successfully", func() {
				cursor, err := collection.Indexes().List(context.Background())
				Expect(err).ToNot(HaveOccurred())
				Expect(cursor).ToNot(BeNil())
				var indexes []storeStructuredMongoTest.MongoIndex
				err = cursor.All(context.Background(), &indexes)
				Expect(err).ToNot(HaveOccurred())

				lowerTimeIndex, err := time.Parse(time.RFC3339, dataStoreMongo.LowerTimeIndexRaw)
				Expect(err).ToNot(HaveOccurred())
				Expect(indexes).To(ConsistOf(
					MatchFields(IgnoreExtras, Fields{
						"Key": Equal(storeStructuredMongoTest.MakeKeySlice("_id")),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":  Equal(storeStructuredMongoTest.MakeKeySlice("_userId", "_active", "type", "-time")),
						"Name": Equal("UserIdTypeWeighted_v2"),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":  Equal(storeStructuredMongoTest.MakeKeySlice("_userId", "_active", "type", "time", "modifiedTime")),
						"Name": Equal("UserIdActiveTypeTimeModifiedTime"),
						"PartialFilterExpression": Equal(bson.D{
							{Key: "time", Value: bson.D{{Key: "$gt", Value: primitive.NewDateTimeFromTime(lowerTimeIndex)}}},
						}),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":  Equal(storeStructuredMongoTest.MakeKeySlice("_userId", "origin.id", "-deletedTime", "_active")),
						"Name": Equal("UserIdOriginId"),
						"PartialFilterExpression": Equal(bson.D{
							{Key: "origin.id", Value: bson.D{{Key: "$exists", Value: true}}},
						}),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":  Equal(storeStructuredMongoTest.MakeKeySlice("uploadId", "type", "-deletedTime", "_active")),
						"Name": Equal("UploadId"),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":  Equal(storeStructuredMongoTest.MakeKeySlice("_userId", "deviceId", "type", "_active", "_deduplicator.hash")),
						"Name": Equal("DeduplicatorHash"),
						"PartialFilterExpression": Equal(bson.D{
							{Key: "_active", Value: true},
							{Key: "_deduplicator.hash", Value: bson.D{{Key: "$exists", Value: true}}},
							{Key: "deviceId", Value: bson.D{{Key: "$exists", Value: true}}},
						}),
					}),
				))
			})

			It("deviceDataSets indexes return successfully", func() {
				cursor, err := dataSetCollection.Indexes().List(context.Background())
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
						"Key":  Equal(storeStructuredMongoTest.MakeKeySlice("_userId", "_active", "type", "-time")),
						"Name": Equal("UserIdTypeWeighted_v2"),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":  Equal(storeStructuredMongoTest.MakeKeySlice("origin.id", "type", "-deletedTime", "_active")),
						"Name": Equal("OriginId"),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":  Equal(storeStructuredMongoTest.MakeKeySlice("uploadId", "type", "-deletedTime", "_active")),
						"Name": Equal("UploadId"),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":    Equal(storeStructuredMongoTest.MakeKeySlice("uploadId")),
						"Unique": Equal(true),
						"Name":   Equal("UniqueUploadId"),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":  Equal(storeStructuredMongoTest.MakeKeySlice("_userId", "client.name", "type", "-createdTime")),
						"Name": Equal("ListUserDataSets"),
						"PartialFilterExpression": Equal(bson.D{
							{Key: "_active", Value: true},
						}),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":  Equal(storeStructuredMongoTest.MakeKeySlice("_userId", "deviceId", "type", "-createdTime")),
						"Name": Equal("ListUserDataSetsDeviceId"),
						"PartialFilterExpression": Equal(bson.D{
							{Key: "_active", Value: true},
						}),
					}),
				))
			})

			It("summary indexes return successfully", func() {
				cursor, err := summaryCollection.Indexes().List(context.Background())
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
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("userId", "type")),
						"Background": Equal(false),
						"Unique":     Equal(true),
						"Name":       Equal("UserIDTypeUnique"),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("type", "dates.outdatedSince", "config.schemaVersion", "dates.lastUpdatedDate")),
						"Background": Equal(false),
						"Name":       Equal("OutdatedSinceSchemaLastUpdated"),
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

		Context("NewSummaryRepository", func() {
			It("returns a new repository", func() {
				summaryRepository = store.NewSummaryRepository()
				Expect(summaryRepository).ToNot(BeNil())
			})
		})

		Context("NewAlertsRepository", func() {
			It("returns a new repository", func() {
				alertsRepository = store.NewAlertsRepository()
				Expect(alertsRepository).ToNot(BeNil())
			})
		})

		Context("with a new repository", func() {
			BeforeEach(func() {
				repository = store.NewDataRepository()
				summaryRepository = store.NewSummaryRepository()
				alertsRepository = store.NewAlertsRepository()
				Expect(repository).ToNot(BeNil())
				Expect(summaryRepository).ToNot(BeNil())
				Expect(alertsRepository).ToNot(BeNil())
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
					createdTimeOther, _ := time.Parse(time.RFC3339, "2016-09-01T12:00:00Z")
					collection = store.GetCollection("deviceData")
					dataSetCollection = store.GetCollection("deviceDataSets")
					dataSetExistingOther = NewDataSet(userTest.RandomID(), dataTest.NewDeviceID())
					dataSetExistingOther.CreatedTime = pointer.FromTime(createdTimeOther)
					dataSetExistingOther.ModifiedTime = pointer.FromTime(createdTimeOther)
					_, err := dataSetCollection.InsertOne(context.Background(), dataSetExistingOther)
					Expect(err).ToNot(HaveOccurred())
					dataSetExistingOne = NewDataSet(userID, deviceID)
					createdTimeOne, _ := time.Parse(time.RFC3339, "2016-09-01T12:30:00Z")
					dataSetExistingOne.CreatedTime = pointer.FromTime(createdTimeOne)
					dataSetExistingOne.ModifiedTime = pointer.FromTime(createdTimeOne)
					_, err = dataSetCollection.InsertOne(context.Background(), dataSetExistingOne)
					Expect(err).ToNot(HaveOccurred())
					dataSetExistingTwo = NewDataSet(userID, deviceID)
					createdTimeTwo, _ := time.Parse(time.RFC3339, "2016-09-01T10:00:00Z")
					dataSetExistingTwo.CreatedTime = pointer.FromTime(createdTimeTwo)
					dataSetExistingTwo.ModifiedTime = pointer.FromTime(createdTimeTwo)
					_, err = dataSetCollection.InsertOne(context.Background(), dataSetExistingTwo)
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
						createdTime, _ := time.Parse(time.RFC3339, "2016-09-01T11:00:00Z")
						dataSet.CreatedTime = pointer.FromTime(createdTime)
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
							_, err := dataSetCollection.InsertOne(context.Background(), dataSet)
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
								createdTime, _ := time.Parse(time.RFC3339, "2016-09-01T13:00:00Z")
								dataSet.DeletedTime = pointer.FromTime(createdTime)
								result := dataSetCollection.FindOneAndReplace(context.Background(), bson.M{"id": dataSet.ID}, dataSet)
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
						createdTime, _ := time.Parse(time.RFC3339, "2016-09-01T11:00:00Z")
						dataSet.CreatedTime = pointer.FromTime(createdTime)
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
							_, err := dataSetCollection.InsertOne(context.Background(), dataSet)
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

						It("sets the created time and modified time", func() {
							Expect(repository.CreateDataSet(ctx, dataSet)).To(Succeed())
							Expect(dataSet.CreatedTime).ToNot(BeNil())
							Expect(dataSet.ModifiedTime).ToNot(BeNil())
							Expect(*dataSet.CreatedTime).To(Equal(*dataSet.ModifiedTime))
							Expect(dataSet.CreatedUserID).To(BeNil())
							Expect(dataSet.ByUser).To(BeNil())

							// Make sure the values are set in the db as well.
							var result *upload.Upload
							err := dataSetCollection.FindOne(context.Background(), bson.M{"uploadId": dataSet.UploadID}).Decode(&result)
							Expect(err).ToNot(HaveOccurred())
							Expect(*result.CreatedTime).To(Equal(*dataSet.CreatedTime))
							Expect(*result.ModifiedTime).To(Equal(*dataSet.ModifiedTime))
							Expect(*result.CreatedTime).To(Equal(*result.ModifiedTime))

						})

						It("has the correct stored data sets", func() {
							ValidateDataSet(dataSetCollection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, bson.M{}, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo)
							Expect(repository.CreateDataSet(ctx, dataSet)).To(Succeed())
							ValidateDataSet(dataSetCollection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, bson.M{}, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo, dataSet)
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
							createdTime := time.Now().UTC().Truncate(time.Millisecond)
							dataSet.CreatedTime = pointer.FromTime(createdTime)
							dataSet.ModifiedTime = pointer.FromTime(createdTime)
							// Insert in BOTH collections to mimic the
							// migration where dataSet will be in deviceData
							// and deviceDataSets. This is because while
							// migration happens an update to a dataset will
							// only succeed if it is still in the old deviceData collection.
							_, err := dataSetCollection.InsertOne(context.Background(), dataSet)
							Expect(err).ToNot(HaveOccurred())
							_, err = collection.InsertOne(context.Background(), dataSet)
							Expect(err).ToNot(HaveOccurred())
							id = *dataSet.UploadID
						})

						AfterEach(func() {
							logger.AssertDebug("DataSetRepository.UpdateDataSet", log.Fields{"id": id, "update": update})
						})

						Context("with updates", func() {
							// TODO

							It("returns nil when the id does not exist", func() {
								id = dataTest.RandomSetID()
								Expect(repository.UpdateDataSet(ctx, id, update)).To(BeNil())
							})

							It("updates modified time when updated", func() {
								newTime, err := time.Parse(time.RFC3339, "2022-01-01T11:00:00Z")
								Expect(err).ToNot(HaveOccurred())
								dataSet.Time = pointer.FromTime(newTime)
								dataSet.SetModifiedTime(pointer.FromTime(time.Now().UTC().Truncate(time.Millisecond)))
								update.Time = pointer.FromTime(newTime)
								_, err = repository.UpdateDataSet(ctx, id, update)
								Expect(err).ToNot(HaveOccurred())
								ValidateDataSetWithModifiedThreshold(dataSetCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, time.Second, dataSet)
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
							ValidateDataSet(dataSetCollection, bson.M{}, bson.M{}, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo, dataSet)
							// All newly created data now includes the modifiedTime as well.
							ValidateDataSet(dataSetCollection, bson.M{"modifiedTime": bson.M{"$exists": true}}, bson.M{}, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo, dataSet)
							update = data.NewDataSetUpdate()
							update.State = pointer.FromString("closed")
							result, err := repository.UpdateDataSet(ctx, id, update)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).ToNot(BeNil())
							Expect(result.State).ToNot(BeNil())
							Expect(*result.State).To(Equal("closed"))
							ValidateDataSet(dataSetCollection, bson.M{}, bson.M{}, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo, result)
							ValidateDataSet(dataSetCollection, bson.M{"modifiedTime": bson.M{"$exists": true}}, bson.M{}, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo, result)

							logger.AssertDebug("DataSetRepository.UpdateDataSet", log.Fields{"id": id, "update": update})
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
						// Insert in BOTH collections to mimic the
						// migration where dataSet will be in deviceData
						// and deviceDataSets. This is because while
						// migration happens an update to a dataset will
						// only succeed if it is still in the old deviceData collection.
						_, err := collection.InsertOne(context.Background(), dataSet)
						Expect(err).ToNot(HaveOccurred())
						_, err = dataSetCollection.InsertOne(context.Background(), dataSet)
						Expect(err).ToNot(HaveOccurred())
						Expect(repository.CreateDataSetData(ctx, dataSetExistingOther, dataSetExistingOtherData)).To(Succeed())
						Expect(repository.CreateDataSetData(ctx, dataSetExistingOne, dataSetExistingOneData)).To(Succeed())
						Expect(repository.CreateDataSetData(ctx, dataSetExistingTwo, dataSetExistingTwoData)).To(Succeed())
					}

					BeforeEach(func() {
						createdTime, _ := time.Parse(time.RFC3339, "2016-09-01T11:00:00Z")
						dataSet.CreatedTime = pointer.FromTime(createdTime)
						dataSet.ModifiedTime = pointer.FromTime(createdTime)
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

							It("sets the deleted and modified time on the data set", func() {
								Expect(repository.DeleteDataSet(ctx, dataSet)).To(Succeed())
								Expect(dataSet.DeletedTime).ToNot(BeNil())
								Expect(dataSet.ModifiedTime).ToNot(BeNil())
								Expect(dataSet.DeletedUserID).To(BeNil())
								Expect(*dataSet.ModifiedTime).Should(BeTemporally("~", time.Now(), time.Second))
							})

							It("has the correct stored data sets", func() {
								ValidateDataSet(dataSetCollection, bson.M{"deletedTime": bson.M{"$exists": true}, "deletedUserId": bson.M{"$exists": false}}, bson.M{})
								Expect(repository.DeleteDataSet(ctx, dataSet)).To(Succeed())
								ValidateDataSet(dataSetCollection, bson.M{"deletedTime": bson.M{"$exists": true}, "deletedUserId": bson.M{"$exists": false}}, bson.M{}, dataSet)
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

							It("sets the created and modified time on the data set data", func() {
								Expect(repository.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
								for _, dataSetDatum := range dataSetData {
									baseDatum, ok := dataSetDatum.(*types.Base)
									Expect(ok).To(BeTrue())
									Expect(baseDatum).ToNot(BeNil())
									Expect(baseDatum.CreatedTime).ToNot(BeNil())
									Expect(baseDatum.ModifiedTime).ToNot(BeNil())
									Expect(*baseDatum.CreatedTime).To(Equal(*baseDatum.ModifiedTime))
									Expect(baseDatum.CreatedUserID).To(BeNil())
								}
							})

							It("sets the modified time on the data set data", func() {
								Expect(repository.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
								ValidateDataSetDataWithModifiedThreshold(collection, bson.M{"uploadId": dataSet.UploadID}, bson.M{"archivedTime": 0}, time.Second, dataSetData)
							})

							It("has the correct stored data set data", func() {
								dataSetBeforeCreateData := append(append(dataSetExistingOtherData, dataSetExistingOneData...), dataSetExistingTwoData...)
								ValidateDataSetData(collection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, bson.M{}, dataSetBeforeCreateData)
								Expect(repository.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
								ValidateDataSetData(collection, bson.M{"createdTime": bson.M{"$exists": true}, "createdUserId": bson.M{"$exists": false}}, bson.M{}, append(dataSetBeforeCreateData, dataSetData...))
							})
						})
					})

					Context("GetDataRange", func() {
						It("returns an error if context is missing", func() {
							status := &data.UserDataStatus{
								FirstData:       *dataSetData[0].GetTime(),
								LastData:        *dataSetData[len(dataSetData)-1].GetTime(),
								LastUpdated:     time.Time{},
								NextLastUpdated: time.Now(),
							}
							_, err := repository.GetDataRange(nil,
								*dataSet.UserID,
								[]string{dataSetData[0].GetType()},
								status)
							Expect(err).To(HaveOccurred())
							Expect(err).To(MatchError("context is missing"))
						})

						It("returns an error if the userId is empty", func() {
							status := &data.UserDataStatus{
								FirstData:       *dataSetData[0].GetTime(),
								LastData:        *dataSetData[len(dataSetData)-1].GetTime(),
								LastUpdated:     time.Time{},
								NextLastUpdated: time.Now(),
							}
							_, err := repository.GetDataRange(ctx,
								"",
								[]string{dataSetData[0].GetType()},
								status)
							Expect(err).To(HaveOccurred())
							Expect(err).To(MatchError("userId is empty"))
						})

						It("returns an error if the typ is empty", func() {
							status := &data.UserDataStatus{
								FirstData:       *dataSetData[0].GetTime(),
								LastData:        *dataSetData[len(dataSetData)-1].GetTime(),
								LastUpdated:     time.Time{},
								NextLastUpdated: time.Now(),
							}
							_, err := repository.GetDataRange(ctx,
								*dataSet.UserID,
								[]string{},
								status)
							Expect(err).To(HaveOccurred())
							Expect(err).To(MatchError("typ is empty"))
						})

						It("returns error if the data times are inverted", func() {
							status := &data.UserDataStatus{
								FirstData:       *dataSetData[len(dataSetData)-1].GetTime(),
								LastData:        *dataSetData[0].GetTime(),
								LastUpdated:     time.Time{},
								NextLastUpdated: time.Now(),
							}
							_, err := repository.GetDataRange(ctx,
								*dataSet.UserID,
								[]string{continuous.Type},
								status)
							Expect(err).To(HaveOccurred())
							Expect(err).To(MatchError(MatchRegexp("^FirstData.*after LastData")))
						})

						It("returns error if the LastUpdated times are inverted", func() {
							status := &data.UserDataStatus{
								FirstData:       *dataSetData[0].GetTime(),
								LastData:        *dataSetData[len(dataSetData)-1].GetTime(),
								LastUpdated:     time.Now(),
								NextLastUpdated: time.Now().Add(-time.Hour),
							}
							_, err := repository.GetDataRange(ctx,
								*dataSet.UserID,
								[]string{continuous.Type},
								status)
							Expect(err).To(HaveOccurred())
							Expect(err).To(MatchError(MatchRegexp("^LastUpdated.*after NextLastUpdated")))
						})

						Context("with database access", func() {
							BeforeEach(func() {
								for i := 0; i < len(dataSetData); i++ {
									dataSetData[i].SetType(continuous.Type)
									dataSetData[i].SetActive(true)
								}
								Expect(repository.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
							})

							It("correctly returns data within range", func() {
								var userData []*glucoseDatum.Glucose
								status := &data.UserDataStatus{
									FirstData:       *dataSetData[0].GetTime(),
									LastData:        *dataSetData[len(dataSetData)-1].GetTime(),
									LastUpdated:     time.Time{},
									NextLastUpdated: time.Now(),
								}
								cursor, err := repository.GetDataRange(ctx,
									*dataSet.UserID,
									[]string{dataSetData[0].GetType()},
									status,
								)
								Expect(err).ToNot(HaveOccurred())

								err = cursor.All(ctx, &userData)
								Expect(err).ToNot(HaveOccurred())
								Expect(cursor).ToNot(BeNil())

								Expect(userData).To(HaveLen(len(dataSetData) - 1))

								// query is $gt, we expect to miss the first record
								Expect(*userData[0].GetTime()).To(Equal(dataSetData[1].GetTime().Truncate(time.Millisecond)))

								// query is $lte, we expect to get the last record requested
								Expect(*userData[len(userData)-1].GetTime()).To(Equal(dataSetData[len(dataSetData)-1].GetTime().Truncate(time.Millisecond)))
							})

							It("correctly misses data outside range", func() {
								var userData []*glucoseDatum.Glucose
								status := &data.UserDataStatus{
									FirstData:       dataSetData[0].GetTime().AddDate(-1, 0, 0),
									LastData:        dataSetData[len(dataSetData)-1].GetTime().AddDate(-1, 0, 0),
									LastUpdated:     time.Time{},
									NextLastUpdated: time.Now(),
								}
								cursor, err := repository.GetDataRange(ctx,
									*dataSet.UserID,
									[]string{dataSetData[0].GetType()},
									status,
								)
								Expect(err).ToNot(HaveOccurred())
								Expect(cursor).ToNot(BeNil())

								err = cursor.All(ctx, &userData)
								Expect(err).ToNot(HaveOccurred())

								Expect(userData).To(HaveLen(0))
							})

							It("correctly misses data of wrong type", func() {
								var userData []*glucoseDatum.Glucose
								status := &data.UserDataStatus{
									FirstData:       *dataSetData[0].GetTime(),
									LastData:        *dataSetData[len(dataSetData)-1].GetTime(),
									LastUpdated:     time.Time{},
									NextLastUpdated: time.Now(),
								}
								cursor, err := repository.GetDataRange(ctx,
									*dataSet.UserID,
									[]string{selfmonitored.Type},
									status,
								)
								Expect(err).ToNot(HaveOccurred())
								Expect(cursor).ToNot(BeNil())

								err = cursor.All(ctx, &userData)
								Expect(err).ToNot(HaveOccurred())

								Expect(userData).To(HaveLen(0))
							})
						})
					})

					Context("GetLastUpdatedForUser", func() {
						It("returns an error if context is missing", func() {
							_, err := repository.GetLastUpdatedForUser(nil, *dataSet.UserID, []string{dataSetData[2].GetType()}, time.Time{})
							Expect(err).To(HaveOccurred())
							Expect(err).To(MatchError("context is missing"))
						})

						It("returns an error if userId is empty", func() {
							_, err := repository.GetLastUpdatedForUser(ctx, "", []string{dataSetData[2].GetType()}, time.Time{})
							Expect(err).To(HaveOccurred())
							Expect(err).To(MatchError("userId is empty"))
						})

						It("returns an error if typ is empty", func() {
							_, err := repository.GetLastUpdatedForUser(ctx, *dataSet.UserID, []string{}, time.Time{})
							Expect(err).To(HaveOccurred())
							Expect(err).To(MatchError("typ is empty"))
						})

						It("returns an error if typ is upload", func() {
							_, err := repository.GetLastUpdatedForUser(ctx, *dataSet.UserID, []string{upload.Type}, time.Time{})
							Expect(err).To(HaveOccurred())
							Expect(err).To(MatchError(fmt.Errorf("unexpected type: %v", upload.Type)))
						})

						Context("with database access", func() {
							var createdTime time.Time
							BeforeEach(func() {
								createdTime = time.Now().UTC().Truncate(time.Millisecond)

								dataSetData[0].SetType(selfmonitored.Type)
								dataSetData[0].SetActive(false)
								dataSetData[0].SetCreatedTime(&createdTime)

								for i := 1; i < len(dataSetData); i++ {
									dataSetData[i].SetType(continuous.Type)
									dataSetData[i].SetActive(true)
									dataSetData[i].SetCreatedTime(&createdTime)
								}

								Expect(repository.CreateDataSetData(ctx, dataSet, dataSetData)).To(Succeed())
							})

							It("correctly finds the LastUpload and LastData for a matching set", func() {
								status, err := repository.GetLastUpdatedForUser(ctx, *dataSet.UserID, []string{continuous.Type}, time.Time{})
								Expect(err).ToNot(HaveOccurred())
								Expect(status.LastData).To(Equal(dataSetData[len(dataSetData)-1].GetTime().Truncate(time.Millisecond)))
								Expect(status.LastUpload).To(BeTemporally("~", createdTime, time.Second))
							})

							It("correctly does not find the LastUpload and LastData for an inactive type", func() {
								status, err := repository.GetLastUpdatedForUser(ctx, *dataSet.UserID, []string{selfmonitored.Type}, time.Time{})
								Expect(err).ToNot(HaveOccurred())
								Expect(status).To(BeNil())
							})

							It("correctly does not find the LastUpload and LastData for an unused type", func() {
								status, err := repository.GetLastUpdatedForUser(ctx, *dataSet.UserID, []string{bolus.Type}, time.Time{})
								Expect(err).ToNot(HaveOccurred())
								Expect(status).To(BeNil())
							})

						})
					})

					Context("DistinctUserIDs", func() {
						It("returns an error if context is missing", func() {
							userIds, err := repository.DistinctUserIDs(nil, []string{continuous.Type})
							Expect(userIds).To(BeNil())
							Expect(err).To(HaveOccurred())
							Expect(err).To(MatchError("context is missing"))
						})

						It("returns an error if typ is empty", func() {
							userIds, err := repository.DistinctUserIDs(ctx, []string{})
							Expect(userIds).To(BeNil())
							Expect(err).To(HaveOccurred())
							Expect(err).To(MatchError("typ is empty"))
						})

						It("returns an error if typ is upload", func() {
							userIds, err := repository.DistinctUserIDs(ctx, []string{upload.Type})
							Expect(userIds).To(BeNil())
							Expect(err).To(HaveOccurred())
							Expect(err).To(MatchError(fmt.Errorf("unexpected type: %v", upload.Type)))
						})

						Context("with database access", func() {
							var userIdOne string
							var userIdTwo string
							var userIdThree string

							BeforeEach(func() {
								userIdOne = userTest.RandomID()
								dataSetOne := NewDataSet(userIdOne, deviceID)
								dataSetDataOne := NewDataSetData(deviceID)

								userIdTwo = userTest.RandomID()
								dataSetTwo := NewDataSet(userIdTwo, deviceID)
								dataSetDataTwo := NewDataSetData(deviceID)

								userIdThree = userTest.RandomID()
								dataSetThree := NewDataSet(userIdThree, deviceID)
								dataSetDataThree := NewDataSetData(deviceID)

								dataSetDataOne[0].SetType(selfmonitored.Type)
								dataSetDataOne[0].SetActive(false)
								for i := 1; i < len(dataSetDataOne); i++ {
									dataSetDataOne[i].SetType(continuous.Type)
									dataSetDataOne[i].SetActive(true)
								}

								dataSetDataTwo[0].SetType(selfmonitored.Type)
								dataSetDataTwo[0].SetActive(true)
								for i := 1; i < len(dataSetDataTwo); i++ {
									dataSetDataTwo[i].SetType(continuous.Type)
									dataSetDataTwo[i].SetActive(true)
								}

								for i := 0; i < len(dataSetDataThree); i++ {
									dataSetDataThree[i].SetType(insulin.Type)
									dataSetDataThree[i].SetActive(true)
								}

								Expect(repository.CreateDataSetData(ctx, dataSetOne, dataSetDataOne)).To(Succeed())
								Expect(repository.CreateDataSetData(ctx, dataSetTwo, dataSetDataTwo)).To(Succeed())
								Expect(repository.CreateDataSetData(ctx, dataSetThree, dataSetDataThree)).To(Succeed())
							})

							It("correctly identifies distinct users", func() {
								userIds, err := repository.DistinctUserIDs(ctx, []string{continuous.Type})
								Expect(userIds).ToNot(BeNil())
								Expect(err).ToNot(HaveOccurred())
								Expect(userIds).To(HaveLen(2))
								Expect(userIds).To(ConsistOf([]string{userIdOne, userIdTwo}))
							})

							It("correctly identifies distinct users with inactive data", func() {
								userIds, err := repository.DistinctUserIDs(ctx, []string{selfmonitored.Type})
								Expect(userIds).ToNot(BeNil())
								Expect(err).ToNot(HaveOccurred())
								Expect(userIds).To(HaveLen(1))
								Expect(userIds[0]).To(Equal(userIdTwo))
							})

							It("correctly identifies distinct users with different-type data", func() {
								userIds, err := repository.DistinctUserIDs(ctx, []string{bolus.Type})
								Expect(userIds).ToNot(BeNil())
								Expect(err).ToNot(HaveOccurred())
								Expect(userIds).To(HaveLen(0))
							})

							It("correctly identifies distinct users with multiple types", func() {
								userIds, err := repository.DistinctUserIDs(ctx, []string{selfmonitored.Type, insulin.Type})
								Expect(userIds).ToNot(BeNil())
								Expect(err).ToNot(HaveOccurred())
								Expect(userIds).To(HaveLen(2))
								Expect(userIds).To(ConsistOf([]string{userIdThree, userIdTwo}))
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
										ValidateDataSet(dataSetCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSet)
									})

									It("succeeds and has the correct stored active data set data", func() {
										Expect(repository.ActivateDataSetData(ctx, dataSet, selectors)).To(Succeed())
										selectedDataSetData.SetActive(true)
										ValidateDataSetData(collection, bson.M{"_active": true}, bson.M{"modifiedTime": 0}, selectedDataSetData)
										ValidateDataSet(dataSetCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSet)
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
									err := repository.ArchiveDataSetData(ctx, dataSet, selectors)
									Expect(err).To(MatchError(dataStoreMongo.ErrSelectorsInvalid))
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
									err := repository.ArchiveDataSetData(ctx, dataSet, selectors)

									Expect(err).To(MatchError(dataStoreMongo.ErrSelectorsInvalid))
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
									err := repository.ArchiveDataSetData(ctx, dataSet, selectors)
									Expect(err).To(MatchError(dataStoreMongo.ErrSelectorsInvalid))
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
										ValidateDataSet(dataSetCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSet)
									})

									It("succeeds and has the correct stored active data set data", func() {
										Expect(repository.ArchiveDataSetData(ctx, dataSet, selectors)).To(Succeed())
										selectedDataSetData.SetActive(false)
										ValidateDataSetData(collection, bson.M{"_active": false, "uploadId": dataSet.UploadID}, bson.M{"archivedTime": 0, "modifiedTime": 0}, selectedDataSetData)
										ValidateDataSet(dataSetCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSet)
									})

									It("succeeds and updates .modifiedTime", func() {
										Expect(repository.ArchiveDataSetData(ctx, dataSet, selectors)).To(Succeed())
										selectedDataSetData.SetActive(false)
										selectedDataSetData.SetModifiedTime(pointer.FromTime(time.Now().UTC().Truncate(time.Millisecond)))
										ValidateDataSetDataWithModifiedThreshold(collection, bson.M{"_active": false, "uploadId": dataSet.UploadID}, bson.M{"archivedTime": 0}, time.Second, selectedDataSetData)
										ValidateDataSet(dataSetCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSet)
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
									err := repository.ArchiveDataSetData(ctx, dataSet, selectors)
									Expect(err).To(MatchError(dataStoreMongo.ErrSelectorsInvalid))
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
									err := repository.ArchiveDataSetData(ctx, dataSet, selectors)
									Expect(err).To(MatchError(dataStoreMongo.ErrSelectorsInvalid))
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
									err := repository.ArchiveDataSetData(ctx, dataSet, selectors)
									Expect(err).To(MatchError(dataStoreMongo.ErrSelectorsInvalid))
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
										ValidateDataSet(dataSetCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSet)
									})

									It("succeeds and has the correct stored active data set data", func() {
										Expect(repository.DeleteDataSetData(ctx, dataSet, selectors)).To(Succeed())
										selectedDataSetData.SetActive(false)
										ValidateDataSetData(collection, bson.M{"deletedTime": bson.M{"$exists": true}}, bson.M{"archivedTime": 0, "deletedTime": 0, "modifiedTime": 0}, selectedDataSetData)
										ValidateDataSet(dataSetCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSet)
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
									err := repository.ArchiveDataSetData(ctx, dataSet, selectors)
									Expect(err).To(MatchError(dataStoreMongo.ErrSelectorsInvalid))
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
									err := repository.ArchiveDataSetData(ctx, dataSet, selectors)
									Expect(err).To(MatchError(dataStoreMongo.ErrSelectorsInvalid))
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
									err := repository.ArchiveDataSetData(ctx, dataSet, selectors)
									Expect(err).To(MatchError(dataStoreMongo.ErrSelectorsInvalid))
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
										ValidateDataSet(dataSetCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSet)
									})

									It("succeeds and has the correct stored active data set data", func() {
										Expect(repository.DestroyDeletedDataSetData(ctx, dataSet, selectors)).To(Succeed())
										ValidateDataSetData(collection, bson.M{"deletedTime": bson.M{"$exists": true}}, bson.M{"archivedTime": 0, "deletedTime": 0, "modifiedTime": 0}, unselectedDataSetData)
										ValidateDataSet(dataSetCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSet)
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
									err := repository.ArchiveDataSetData(ctx, dataSet, selectors)
									Expect(err).To(MatchError(dataStoreMongo.ErrSelectorsInvalid))
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
									err := repository.DestroyDeletedDataSetData(ctx, dataSet, selectors)
									Expect(err).To(MatchError(dataStoreMongo.ErrSelectorsInvalid))
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
									err := repository.ArchiveDataSetData(ctx, dataSet, selectors)
									Expect(err).To(MatchError(dataStoreMongo.ErrSelectorsInvalid))
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
										ValidateDataSet(dataSetCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSet)
									})

									It("succeeds and has the correct stored active data set data", func() {
										Expect(repository.DestroyDataSetData(ctx, dataSet, selectors)).To(Succeed())
										ValidateDataSetData(collection, bson.M{"uploadId": dataSet.UploadID}, bson.M{"modifiedTime": 0}, unselectedDataSetData)
										ValidateDataSet(dataSetCollection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, dataSet)
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
									err := repository.ArchiveDataSetData(ctx, dataSet, selectors)
									Expect(err).To(MatchError(dataStoreMongo.ErrSelectorsInvalid))
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
									err := repository.ArchiveDataSetData(ctx, dataSet, selectors)
									Expect(err).To(MatchError(dataStoreMongo.ErrSelectorsInvalid))
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
									err := repository.ArchiveDataSetData(ctx, dataSet, selectors)
									Expect(err).To(MatchError(dataStoreMongo.ErrSelectorsInvalid))
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
								ValidateDataSet(dataSetCollection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true}, bson.M{}, dataSetExistingOne)
								Expect(repository.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(Succeed())
								ValidateDataSet(dataSetCollection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true}, bson.M{}, dataSetExistingOne)
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
								ValidateDataSet(dataSetCollection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true}, bson.M{}, dataSetExistingOne)
								Expect(repository.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)).To(Succeed())
								ValidateDataSet(dataSetCollection, bson.M{"uploadId": dataSetExistingOne.UploadID, "_active": true}, bson.M{}, dataSetExistingOne)
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
								ValidateDataSet(dataSetCollection, bson.M{"uploadId": dataSetExistingTwo.UploadID, "_active": true}, bson.M{}, dataSetExistingTwo)
								Expect(repository.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSetExistingOne)).To(Succeed())
								ValidateDataSet(dataSetCollection, bson.M{"uploadId": dataSetExistingTwo.UploadID, "_active": true}, bson.M{}, dataSetExistingTwo)
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
								ValidateDataSet(dataSetCollection, bson.M{}, bson.M{}, dataSet, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo)
								ValidateDataSet(dataSetCollection, bson.M{"deletedTime": bson.M{"$exists": false}, "deletedUserId": bson.M{"$exists": false}}, bson.M{}, dataSet, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo)
								Expect(repository.DeleteOtherDataSetData(ctx, dataSet)).To(Succeed())
								Expect(dataSetCollection.CountDocuments(ctx, bson.M{"type": "upload"})).To(Equal(int64(4)))
								ValidateDataSet(dataSetCollection, bson.M{"deletedTime": bson.M{"$exists": true}, "deletedUserId": bson.M{"$exists": false}}, bson.M{"deletedTime": 0}, dataSetExistingTwo, dataSetExistingOne)
								ValidateDataSet(dataSetCollection, bson.M{"deletedTime": bson.M{"$exists": false}, "deletedUserId": bson.M{"$exists": false}}, bson.M{}, dataSet, dataSetExistingOther)
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
								createdTime, _ := time.Parse(time.RFC3339, "2016-09-01T11:00:00Z")
								destroyDataSet.CreatedTime = pointer.FromTime(createdTime)
								destroyDataSet.ModifiedTime = pointer.FromTime(createdTime)
								// Insert in BOTH collections to mimic the
								// migration where dataSet will be in deviceData
								// and deviceDataSets. This is because while
								// migration happens an update to a dataset will
								// only succeed if it is still in the old deviceData collection.
								_, err := collection.InsertOne(context.Background(), destroyDataSet)
								Expect(err).ToNot(HaveOccurred())
								_, err = dataSetCollection.InsertOne(context.Background(), destroyDataSet)
								Expect(err).ToNot(HaveOccurred())
								destroyDataSetData = NewDataSetData(destroyDeviceID)
								Expect(repository.CreateDataSetData(ctx, destroyDataSet, destroyDataSetData)).To(Succeed())
							})

							It("succeeds if it successfully destroys all data for user by id", func() {
								Expect(repository.DestroyDataForUserByID(ctx, destroyUserID)).To(Succeed())
							})

							It("has the correct stored data sets", func() {
								ValidateDataSet(dataSetCollection, bson.M{}, bson.M{}, dataSet, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo, destroyDataSet)
								Expect(repository.DestroyDataForUserByID(ctx, destroyUserID)).To(Succeed())
								ValidateDataSet(dataSetCollection, bson.M{}, bson.M{}, dataSet, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo)
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

		Context("GetAlertableData", func() {
			testDataSet := func(userID, deviceID string) *upload.Upload {
				GinkgoHelper()
				collection = store.GetCollection("deviceData")
				dataSetCollection = store.GetCollection("deviceDataSets")
				upload := NewDataSet(userID, deviceID)
				creationTime := time.Now().Add(-24 * time.Hour)
				upload.Active = true
				upload.CreatedTime = pointer.FromTime(creationTime)
				upload.ModifiedTime = pointer.FromTime(creationTime)
				return upload
			}

			testDataSetData := func(upload *upload.Upload) []data.Datum {
				requiredRecords := test.RandomIntFromRange(6, 8)
				recordInterval := 5 * time.Minute
				t := time.Now().UTC().Add(-time.Duration(requiredRecords) * recordInterval)
				var dataSetData = make([]data.Datum, 0, 2*requiredRecords)
				for count := 0; count < requiredRecords; count++ {
					datum := dataTypesTest.RandomBase()
					datum.Type = "cbg"
					datumTime := t.Add(time.Duration(count) * recordInterval)
					datum.Time = pointer.FromAny(datumTime)
					datum.Active = true
					datum.DeviceID = pointer.FromAny(*upload.DeviceID)
					datum.UploadID = pointer.FromAny(*upload.UploadID)
					dataSetData = append(dataSetData, datum)

					datum = dataTypesTest.RandomBase()
					datum.Type = "dosingDecision"
					datum.Time = pointer.FromAny(datumTime)
					datum.Active = true
					datum.DeviceID = pointer.FromAny(*upload.DeviceID)
					datum.UploadID = pointer.FromAny(*upload.UploadID)
					dataSetData = append(dataSetData, datum)
				}
				return dataSetData
			}

			It("retrieves both cbg and dosing data", func() {
				ctx := context.Background()
				ctx = log.NewContextWithLogger(ctx, logTest.NewLogger())
				repository = store.NewDataRepository()
				Expect(repository).ToNot(BeNil())
				testUserID := userTest.RandomID()
				testDeviceID := dataTest.NewDeviceID()
				testSet := testDataSet(testUserID, testDeviceID)
				Expect(repository.CreateDataSet(ctx, testSet)).To(Succeed())
				testSetData := testDataSetData(testSet)
				Expect(repository.CreateDataSetData(ctx, testSet, testSetData)).To(Succeed())

				params := dataStore.AlertableParams{
					Start:    time.Now().Add(-time.Hour),
					UserID:   testUserID,
					UploadID: *testSet.UploadID,
				}
				resp, err := repository.GetAlertableData(ctx, params)

				Expect(err).To(Succeed())
				Expect(resp).ToNot(BeNil())
				Expect(resp.Glucose).ToNot(BeEmpty())
				Expect(resp.DosingDecisions).ToNot(BeEmpty())
			})

		})

		Context("alerts", func() {
			BeforeEach(func() {
				alertsRepository = store.NewAlertsRepository()
				Expect(alertsRepository).ToNot(BeNil())
			})

			prep := func(upsertDoc bool) (context.Context, *alerts.Config, bson.M) {
				cfg := &alerts.Config{
					FollowedUserID: "followed-user-id",
					UserID:         "user-id",
				}
				ctx := context.Background()
				filter := bson.M{}
				if upsertDoc {
					Expect(alertsRepository.Upsert(ctx, cfg)).
						To(Succeed())
					filter["userId"] = cfg.UserID
					filter["followedUserId"] = cfg.FollowedUserID
				}

				return ctx, cfg, filter
			}

			Describe("Upsert", func() {
				Context("when no document exists", func() {
					It("creates a new document", func() {
						ctx, cfg, filter := prep(false)

						Expect(alertsRepository.Upsert(ctx, cfg)).To(Succeed())

						res := store.GetCollection("alerts").FindOne(ctx, filter)
						Expect(res.Err()).To(Succeed())
					})
				})

				It("updates the existing document", func() {
					ctx, cfg, filter := prep(true)

					cfg.Low = &alerts.LowAlert{Base: alerts.Base{Enabled: true}}
					err := alertsRepository.Upsert(ctx, cfg)
					Expect(err).To(Succeed())

					doc := &alerts.Config{}
					res := store.GetCollection("alerts").FindOne(ctx, filter)
					Expect(res.Err()).To(Succeed())
					Expect(res.Decode(doc)).To(Succeed())
					Expect(doc.Low).ToNot(BeNil())
					Expect(doc.Low.Base.Enabled).To(Equal(true))
				})

			})

			Describe("Get", func() {
				Context("when no document exists", func() {
					It("returns an error", func() {
						ctx, cfg, _ := prep(false)

						_, err := alertsRepository.Get(ctx, cfg)
						Expect(err).To(MatchError(mongo.ErrNoDocuments))
					})
				})

				It("retrieves the correct document", func() {
					ctx, cfg, _ := prep(true)
					other := &alerts.Config{
						UserID:         "879d5cb2-f70d-4b05-8d38-fb6d88ef2ea9",
						FollowedUserID: "d2ee01db-3458-42ac-95d2-ac2fc571a21d",
						Alerts: alerts.Alerts{
							High: &alerts.HighAlert{
								Base: alerts.Base{Enabled: true},
							},
						}}
					Expect(alertsRepository.Upsert(ctx, other)).To(Succeed())
					cfg.Low = &alerts.LowAlert{Base: alerts.Base{Enabled: true}}
					err := alertsRepository.Upsert(ctx, cfg)
					Expect(err).To(Succeed())

					got, err := alertsRepository.Get(ctx, cfg)
					Expect(err).To(Succeed())
					Expect(got).ToNot(BeNil())
					Expect(got.Low).ToNot(BeNil())
					Expect(got.Low.Enabled).To(Equal(true))
					Expect(got.UserID).To(Equal(cfg.UserID))
					Expect(got.FollowedUserID).To(Equal(cfg.FollowedUserID))
				})
			})

			Describe("Delete", func() {
				It("deletes the document", func() {
					ctx, cfg, filter := prep(true)

					err := alertsRepository.Delete(ctx, cfg)
					Expect(err).To(Succeed())

					res := store.GetCollection("alerts").FindOne(ctx, filter)
					Expect(res.Err()).To(MatchError(mongo.ErrNoDocuments))
				})
			})
		})
	})
})
