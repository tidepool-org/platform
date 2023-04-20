package mongo_test

import (
	"context"
	"math/rand"
	"time"

	"github.com/tidepool-org/platform/data/summary"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	dataTypesBloodGlucoseTest "github.com/tidepool-org/platform/data/types/blood/glucose/test"
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

func NewLegacyDataSet(userID string, deviceID string) *dataTypesUploadTest.LegacyUpload {
	dataSet := dataTypesUploadTest.RandomLegacyUpload()
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
	requiredRecords := test.RandomIntFromRange(4, 6)
	var dataSetData = make([]data.Datum, requiredRecords)
	for count := 0; count < requiredRecords; count++ {
		datum := dataTypesTest.RandomBase()
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
		dataSetData[count] = datum
	}
	return dataSetData
}

func NewContinuous(units *string) *continuous.Continuous {
	datum := continuous.New()
	datum.Glucose = *dataTypesBloodGlucoseTest.NewGlucose(units)
	datum.Type = "cbg"

	return datum
}

func NewDataSetCGMData(deviceID string, startTime time.Time, days int) data.Data {
	requiredRecords := days * 288
	var dataSetData = make([]data.Datum, requiredRecords)
	var datumTime time.Time
	unit := "mmol/L"

	// generate 2 weeks of data
	for count := 0; count < requiredRecords; count++ {
		datumTime = startTime.Add(time.Duration(-count) * time.Minute * 5)

		datum := NewContinuous(&unit)
		datum.Active = true
		datum.ArchivedDataSetID = nil
		datum.ArchivedTime = nil
		datum.CreatedTime = nil
		datum.CreatedUserID = nil
		datum.DeletedTime = nil
		datum.DeletedUserID = nil
		datum.DeviceID = pointer.FromString(deviceID)
		datum.ModifiedTime = nil
		datum.ModifiedUserID = nil
		datum.Time = pointer.FromTime(datumTime)

		datum.Glucose.Value = pointer.FromFloat64(1 + (25-1)*rand.Float64())

		dataSetData[requiredRecords-count-1] = datum
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
	// each object in a slice. However this does a deep equal and certain
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
	// the time it's checked the time will (likely) be different. Instead we
	// compare them and make sure they're within a time.Duration threshold of
	// each other outside of this function.
	delete(dataSetDatumAsInterface, "modifiedTime")
	return dataSetDatumAsInterface
}

var _ = Describe("Mongo", func() {
	var logger *logTest.Logger
	var config *storeStructuredMongo.Config
	var store *dataStoreMongo.Store
	var repository dataStore.DataRepository
	var summaryRepository dataStore.SummaryRepository

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
		var summaryCollection *mongo.Collection

		BeforeEach(func() {
			var err error
			store, err = dataStoreMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
			collection = store.GetCollection("deviceData")
			summaryCollection = store.GetCollection("summary")
			Expect(store.EnsureIndexes()).To(Succeed())
		})

		AfterEach(func() {
			if collection != nil {
				collection.Database().Drop(context.Background())
				summaryCollection.Database().Drop(context.Background())
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
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("_userId", "_active", "type", "-modifiedTime")),
						"Background": Equal(true),
						"Name":       Equal("ModifiedTime"),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("origin.id", "type", "-deletedTime", "_active")),
						"Background": Equal(true),
						"Name":       Equal("OriginId"),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":                     Equal(storeStructuredMongoTest.MakeKeySlice("uploadId")),
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

			It("returns successfully", func() {
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
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("userId")),
						"Background": Equal(true),
						"Unique":     Equal(true),
						"Name":       Equal("UserID"),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("lastUpdatedDate")),
						"Background": Equal(true),
						"Name":       Equal("LastUpdatedDate"),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("outdatedSince")),
						"Background": Equal(true),
						"Name":       Equal("OutdatedSince"),
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

		Context("with a new repository", func() {
			BeforeEach(func() {
				repository = store.NewDataRepository()
				summaryRepository = store.NewSummaryRepository()
				Expect(repository).ToNot(BeNil())
				Expect(summaryRepository).ToNot(BeNil())
			})

			AfterEach(func() {
				if repository != nil {
					collection.DeleteMany(context.Background(), bson.D{})
					summaryCollection.DeleteMany(context.Background(), bson.D{})
				}
			})

			Context("Summary", func() {
				var ctx context.Context
				var userID string
				var otherUserID string
				var deviceID string
				var randomSummary *summary.Summary
				var anotherRandomSummary *summary.Summary
				var dataSetCGM *upload.Upload
				var dataSetCGMData data.Data
				var dataSetLastUpdated time.Time
				var dataSetFirstData time.Time
				var otherDataSetCGM *upload.Upload
				var otherDataSetCGMData data.Data
				var otherDataSetLastUpdated time.Time
				var err error

				BeforeEach(func() {
					// generate all these once, they don't need to change
					ctx = log.NewContextWithLogger(context.Background(), logger)
					userID = userTest.RandomID()
					otherUserID = userTest.RandomID()
					deviceID = dataTest.NewDeviceID()

					randomSummary = dataTest.RandomSummary()
					randomSummary.UserID = userID

					anotherRandomSummary = dataTest.RandomSummary()
					anotherRandomSummary.UserID = userTest.RandomID()
				})

				Context("With cgm data", func() {
					BeforeEach(func() {
						dataSetCGM = NewDataSet(userID, deviceID)
						dataSetCGM.CreatedTime = pointer.FromTime(time.Now().UTC().AddDate(0, -3, 0))

						dataSetLastUpdated = time.Now().UTC().AddDate(0, -3, 0).Truncate(time.Millisecond)
						dataSetCGMData = NewDataSetCGMData(deviceID, dataSetLastUpdated, 3)

						_, err = collection.InsertOne(context.Background(), dataSetCGM)
						Expect(err).ToNot(HaveOccurred())

						Expect(repository.CreateDataSetData(ctx, dataSetCGM, dataSetCGMData)).To(Succeed())
					})

					Context("GetCGMDataRange", func() {
						It("returns right count for the requested range", func() {
							var cgmRecords []*continuous.Continuous
							dataSetFirstData = dataSetLastUpdated.AddDate(0, 0, -2)
							cgmRecords, err = repository.GetCGMDataRange(ctx, userID, dataSetFirstData, dataSetLastUpdated)

							Expect(err).ToNot(HaveOccurred())
							Expect(len(cgmRecords)).To(Equal(576))
						})

						It("returns right data for the requested range", func() {
							var cgmRecords []*continuous.Continuous
							dataSetFirstData = dataSetLastUpdated.AddDate(0, 0, -3)
							cgmRecords, err = repository.GetCGMDataRange(ctx, userID, dataSetFirstData, dataSetLastUpdated)

							Expect(err).ToNot(HaveOccurred())
							Expect(len(cgmRecords)).To(Equal(864))
							for i, cgmDatum := range cgmRecords {
								Expect(*cgmDatum.Time).To(Equal(dataSetFirstData.Add(time.Duration(i+1) * 5 * time.Minute).Truncate(time.Millisecond)))
							}
						})
					})

					Context("GetLastUpdatedForUser", func() {
						It("returns right lastUpdated for user", func() {
							var userLastUpdated *summary.UserLastUpdated
							userLastUpdated, err = repository.GetLastUpdatedForUser(ctx, userID)

							Expect(err).ToNot(HaveOccurred())
							Expect(userLastUpdated.LastData).To(Equal(dataSetLastUpdated))
							Expect(userLastUpdated.LastUpload.After(dataSetLastUpdated)).To(BeTrue())
						})

						It("returns right lastUpdated for user with no data", func() {
							var userLastUpdated *summary.UserLastUpdated
							userLastUpdated, err = repository.GetLastUpdatedForUser(ctx, "deadbeef")

							Expect(err).ToNot(HaveOccurred())
							Expect(userLastUpdated).To(BeNil())
						})

						It("returns right lastUpdated for user with far future data", func() {
							dataSetLastUpdatedFuture := time.Now().UTC().AddDate(0, 0, 4).Truncate(time.Millisecond)
							dataSetCGMFuture := NewDataSet(userID, deviceID)
							dataSetCGMFuture.CreatedTime = pointer.FromTime(time.Now().UTC().AddDate(0, 0, 4))
							dataSetCGMDataFuture := NewDataSetCGMData(deviceID, dataSetLastUpdatedFuture, 1)

							_, err = collection.InsertOne(context.Background(), dataSetCGMFuture)
							Expect(err).ToNot(HaveOccurred())
							Expect(repository.CreateDataSetData(ctx, dataSetCGMFuture, dataSetCGMDataFuture)).To(Succeed())

							var userLastUpdated *summary.UserLastUpdated
							userLastUpdated, err = repository.GetLastUpdatedForUser(ctx, userID)

							Expect(err).ToNot(HaveOccurred())
							Expect(userLastUpdated.LastData).To(Equal(dataSetLastUpdated))
							Expect(userLastUpdated.LastUpload.After(dataSetLastUpdated)).To(BeTrue())
						})
					})

					Context("DistinctCGMUserIDs", func() {
						It("returns correct count and IDs of distinct users", func() {
							resultUserIDs, err := repository.DistinctCGMUserIDs(ctx)
							Expect(err).ToNot(HaveOccurred())
							Expect(len(resultUserIDs)).To(Equal(1))
							Expect(resultUserIDs).To(ConsistOf([1]string{userID}))
						})

						It("returns correct count and IDs of distinct users after change", func() {
							otherDataSetCGM = NewDataSet(otherUserID, deviceID)
							otherDataSetCGM.CreatedTime = pointer.FromTime(time.Now().UTC().AddDate(0, -6, 0))

							otherDataSetLastUpdated = time.Now().UTC().AddDate(0, -6, 0)
							otherDataSetCGMData = NewDataSetCGMData(deviceID, otherDataSetLastUpdated, 1)

							_, err = collection.InsertOne(context.Background(), otherDataSetCGM)
							Expect(err).ToNot(HaveOccurred())
							Expect(repository.CreateDataSetData(ctx, otherDataSetCGM, otherDataSetCGMData)).To(Succeed())

							resultUserIDs, err := repository.DistinctCGMUserIDs(ctx)
							Expect(err).ToNot(HaveOccurred())
							Expect(len(resultUserIDs)).To(Equal(2))
							Expect(resultUserIDs).To(ConsistOf([2]string{userID, otherUserID}))
						})
					})
				})

				Context("UpdateSummary", func() {
					It("returns error if context is empty", func() {
						_, err = summaryRepository.UpdateSummary(nil, randomSummary)
						Expect(err).To(MatchError("context is missing"))
					})

					It("test empty summary is correctly handled", func() {
						_, err = summaryRepository.UpdateSummary(ctx, nil)
						Expect(err).To(MatchError("summary object is missing"))
					})

					It("test empty UserID is correctly handled", func() {
						randomSummary.UserID = ""
						_, err = summaryRepository.UpdateSummary(ctx, randomSummary)
						Expect(err).To(MatchError("summary missing UserID"))
					})

					It("test that summary can be written", func() {
						_, err = summaryRepository.UpdateSummary(ctx, randomSummary)
						Expect(err).ToNot(HaveOccurred())
					})

					It("test that summary changes when written", func() {
						var firstSummary *summary.Summary
						var newSummary *summary.Summary
						// make keys match, and remove some days to ensure they also get removed
						anotherRandomSummary.UserID = randomSummary.UserID
						anotherRandomSummary.DailyStats = anotherRandomSummary.DailyStats[0:0]

						_, err = summaryRepository.UpdateSummary(ctx, randomSummary)
						Expect(err).ToNot(HaveOccurred())

						firstSummary, err = summaryRepository.GetSummary(ctx, randomSummary.UserID)
						Expect(err).ToNot(HaveOccurred())
						randomSummary.ID = firstSummary.ID
						Expect(firstSummary).To(Equal(randomSummary))

						_, err = summaryRepository.UpdateSummary(ctx, anotherRandomSummary)
						Expect(err).ToNot(HaveOccurred())

						newSummary, err = summaryRepository.GetSummary(ctx, randomSummary.UserID)
						Expect(err).ToNot(HaveOccurred())
						anotherRandomSummary.ID = firstSummary.ID
						Expect(newSummary).To(Equal(anotherRandomSummary))
						Expect(firstSummary).ToNot(Equal(newSummary))
					})

					It("ensure that nil summary fields are correctly removed from the db", func() {
						var newSummary *summary.Summary
						var userSummary *summary.Summary
						userSummary, err = summaryRepository.GetSummary(ctx, userID)
						Expect(err).ToNot(HaveOccurred())
						Expect(userSummary).To(BeNil())

						randomSummary.Periods["14d"].GlucoseManagementIndicator = pointer.FromFloat64(7.5)
						Expect(randomSummary.Periods["14d"].GlucoseManagementIndicator).ToNot(BeNil())

						_, err = summaryRepository.UpdateSummary(ctx, randomSummary)
						Expect(err).ToNot(HaveOccurred())

						newSummary, err = summaryRepository.GetSummary(ctx, userID)
						Expect(err).ToNot(HaveOccurred())
						Expect(newSummary.Periods["14d"].GlucoseManagementIndicator).ToNot(BeNil())

						randomSummary.Periods["14d"].GlucoseManagementIndicator = nil
						Expect(randomSummary.Periods["14d"].GlucoseManagementIndicator).To(BeNil())

						_, err = summaryRepository.UpdateSummary(ctx, randomSummary)
						Expect(err).ToNot(HaveOccurred())

						newSummary, err = summaryRepository.GetSummary(ctx, userID)
						Expect(err).ToNot(HaveOccurred())
						Expect(newSummary.Periods["14d"].GlucoseManagementIndicator).To(BeNil())
					})
				})

				Context("GetSummary", func() {
					It("returns error if context is empty", func() {
						_, err = summaryRepository.GetSummary(nil, userID)
						Expect(err).To(MatchError("context is missing"))
					})

					It("returns error if UserID is empty", func() {
						_, err = summaryRepository.GetSummary(ctx, "")
						Expect(err).To(MatchError("summary UserID is missing"))
					})

					It("returns an error if getsummary cannot retrieve record", func() {
						userSummary, err := summaryRepository.GetSummary(ctx, userID)
						Expect(err).ToNot(HaveOccurred())
						Expect(userSummary).To(BeNil())

						_, err = summaryRepository.UpdateSummary(ctx, randomSummary)
						Expect(err).ToNot(HaveOccurred())

						newSummary, err := summaryRepository.GetSummary(ctx, userID)
						Expect(err).ToNot(HaveOccurred())
						// copy id from inserted summary for easy equality
						randomSummary.ID = newSummary.ID
						Expect(newSummary).To(Equal(randomSummary))
					})
				})

				Context("CreateSummaries", func() {
					It("returns error if context is empty", func() {
						summaries := []*summary.Summary{randomSummary}
						count, err := summaryRepository.CreateSummaries(nil, summaries)

						Expect(count).To(Equal(0))
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("context is missing"))
					})

					It("returns error if summaries is empty", func() {
						summaries := []*summary.Summary{}
						count, err := summaryRepository.CreateSummaries(ctx, summaries)

						Expect(count).To(Equal(0))
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("summaries for create missing"))
					})

					It("creates one summary correctly", func() {
						summaries := []*summary.Summary{randomSummary}
						count, err := summaryRepository.CreateSummaries(ctx, summaries)

						Expect(err).ToNot(HaveOccurred())
						Expect(count).To(Equal(1))

						newSummary, err := summaryRepository.GetSummary(ctx, randomSummary.UserID)
						Expect(err).ToNot(HaveOccurred())

						// copy id from inserted summary for easy equality
						randomSummary.ID = newSummary.ID
						Expect(newSummary).To(Equal(randomSummary))
					})

					It("creates multiple summaries correctly", func() {
						summaries := []*summary.Summary{randomSummary, anotherRandomSummary}
						count, err := summaryRepository.CreateSummaries(ctx, summaries)

						Expect(err).ToNot(HaveOccurred())
						Expect(count).To(Equal(2))

						for _, insertedSummary := range summaries {
							newSummary, err := summaryRepository.GetSummary(ctx, insertedSummary.UserID)
							Expect(err).ToNot(HaveOccurred())

							// copy id from inserted summary for easy equality
							insertedSummary.ID = newSummary.ID
							Expect(newSummary).To(Equal(insertedSummary))
						}
					})
				})

				Context("SetOutdated", func() {
					It("returns error if context is empty", func() {
						_, err := summaryRepository.SetOutdated(nil, randomSummary.UserID)

						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("context is missing"))
					})

					It("returns error if id is empty", func() {
						_, err := summaryRepository.SetOutdated(ctx, "")

						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("user id is missing"))
					})

					It("returns and correctly sets outdated", func() {
						summaries := []*summary.Summary{randomSummary}
						_, err := summaryRepository.CreateSummaries(ctx, summaries)
						Expect(err).ToNot(HaveOccurred())

						timestamp, err := summaryRepository.SetOutdated(ctx, randomSummary.UserID)
						Expect(err).ToNot(HaveOccurred())
						summary, err := summaryRepository.GetSummary(ctx, randomSummary.UserID)

						Expect(err).ToNot(HaveOccurred())
						Expect(timestamp).To(Equal(summary.OutdatedSince))
					})

					It("returns and correctly upserts an outdated summary if none existed", func() {
						timestamp, err := summaryRepository.SetOutdated(ctx, randomSummary.UserID)
						Expect(err).ToNot(HaveOccurred())
						summary, err := summaryRepository.GetSummary(ctx, randomSummary.UserID)

						Expect(err).ToNot(HaveOccurred())
						Expect(timestamp).To(Equal(summary.OutdatedSince))
						Expect(randomSummary.UserID).To(Equal(summary.UserID))
					})

					It("returns and correctly leaves outdated unchanged if already set", func() {
						summaries := []*summary.Summary{randomSummary}
						_, err := summaryRepository.CreateSummaries(ctx, summaries)
						Expect(err).ToNot(HaveOccurred())

						timestampOne, err := summaryRepository.SetOutdated(ctx, randomSummary.UserID)
						Expect(err).ToNot(HaveOccurred())

						timestampTwo, err := summaryRepository.SetOutdated(ctx, randomSummary.UserID)

						Expect(err).ToNot(HaveOccurred())
						Expect(timestampOne).To(Equal(timestampTwo))
					})
				})

				Context("GetOutdatedUserIDs", func() {
					var pagination *page.Pagination

					It("returns error if context is empty", func() {
						pagination = page.NewPagination()
						_, err := summaryRepository.GetOutdatedUserIDs(nil, pagination)

						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("context is missing"))
					})

					It("returns error if pagination is empty", func() {
						_, err := summaryRepository.GetOutdatedUserIDs(ctx, nil)

						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("pagination is missing"))
					})

					It("returns and correctly gets outdated summaries", func() {
						pagination = page.NewPagination()

						summaries := []*summary.Summary{randomSummary, anotherRandomSummary}
						_, err := summaryRepository.CreateSummaries(ctx, summaries)
						Expect(err).ToNot(HaveOccurred())

						_, err = summaryRepository.SetOutdated(ctx, randomSummary.UserID)
						Expect(err).ToNot(HaveOccurred())

						userIDs, err := summaryRepository.GetOutdatedUserIDs(ctx, pagination)

						Expect(err).ToNot(HaveOccurred())
						Expect(userIDs).To(ConsistOf([1]string{randomSummary.UserID}))
					})
				})

				Context("Test full update summary flow", func() {
					It("ensure an outdated record is no longer outdated after update", func() {
						userSummary, err := summaryRepository.GetSummary(ctx, randomSummary.UserID)
						Expect(err).ToNot(HaveOccurred())
						Expect(userSummary).To(BeNil())

						_, err = summaryRepository.UpdateSummary(ctx, randomSummary)
						Expect(err).ToNot(HaveOccurred())

						newSummary, err := summaryRepository.GetSummary(ctx, randomSummary.UserID)
						Expect(err).ToNot(HaveOccurred())
						randomSummary.ID = newSummary.ID
						Expect(newSummary).To(Equal(randomSummary))

						outdatedTime, err := summaryRepository.SetOutdated(ctx, randomSummary.UserID)
						Expect(err).ToNot(HaveOccurred())
						Expect(outdatedTime).ToNot(BeNil())

						outdatedSummary, err := summaryRepository.GetSummary(ctx, randomSummary.UserID)
						Expect(err).ToNot(HaveOccurred())
						Expect(outdatedSummary.OutdatedSince).To(Equal(outdatedTime))

						outdatedSummary.OutdatedSince = nil
						_, err = summaryRepository.UpdateSummary(ctx, outdatedSummary)
						Expect(err).ToNot(HaveOccurred())

						finalSummary, err := summaryRepository.GetSummary(ctx, randomSummary.UserID)
						Expect(err).ToNot(HaveOccurred())
						Expect(finalSummary.OutdatedSince).To(BeNil())
					})
				})

				Context("DistinctSummaryIDs", func() {
					It("returns error if context is empty", func() {
						_, err := summaryRepository.DistinctSummaryIDs(nil)

						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("context is missing"))
					})

					It("returns correct count and IDs of one summary", func() {
						summaries := []*summary.Summary{randomSummary}
						_, err := summaryRepository.CreateSummaries(ctx, summaries)
						Expect(err).ToNot(HaveOccurred())

						resultUserIDs, err := summaryRepository.DistinctSummaryIDs(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(resultUserIDs).To(ConsistOf([1]string{randomSummary.UserID}))
					})

					It("returns correct count and IDs of multiple summaries", func() {
						summaries := []*summary.Summary{randomSummary, anotherRandomSummary}
						_, err := summaryRepository.CreateSummaries(ctx, summaries)
						Expect(err).ToNot(HaveOccurred())

						resultUserIDs, err := summaryRepository.DistinctSummaryIDs(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(resultUserIDs).To(ConsistOf([2]string{randomSummary.UserID, anotherRandomSummary.UserID}))

					})

					It("returns correct count and IDs of summaries with 0 summaries", func() {
						resultUserIDs, err := summaryRepository.DistinctSummaryIDs(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(resultUserIDs).To(BeEmpty())
					})
				})
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
					dataSetExistingOther = NewDataSet(userTest.RandomID(), dataTest.NewDeviceID())
					dataSetExistingOther.CreatedTime = pointer.FromTime(createdTimeOther)
					dataSetExistingOther.ModifiedTime = pointer.FromTime(createdTimeOther)
					_, err := collection.InsertOne(context.Background(), dataSetExistingOther)
					Expect(err).ToNot(HaveOccurred())
					dataSetExistingOne = NewDataSet(userID, deviceID)
					createdTimeOne, _ := time.Parse(time.RFC3339, "2016-09-01T12:30:00Z")
					dataSetExistingOne.CreatedTime = pointer.FromTime(createdTimeOne)
					dataSetExistingOne.ModifiedTime = pointer.FromTime(createdTimeOne)
					_, err = collection.InsertOne(context.Background(), dataSetExistingOne)
					Expect(err).ToNot(HaveOccurred())
					dataSetExistingTwo = NewDataSet(userID, deviceID)
					createdTimeTwo, _ := time.Parse(time.RFC3339, "2016-09-01T10:00:00Z")
					dataSetExistingTwo.CreatedTime = pointer.FromTime(createdTimeTwo)
					dataSetExistingTwo.ModifiedTime = pointer.FromTime(createdTimeTwo)
					_, err = collection.InsertOne(context.Background(), dataSetExistingTwo)
					Expect(err).ToNot(HaveOccurred())
				}

				BeforeEach(func() {
					ctx = log.NewContextWithLogger(context.Background(), logger)
					userID = userTest.RandomID()
					deviceID = dataTest.NewDeviceID()
					dataSet = NewDataSet(userID, deviceID)
				})

				Context("DateUnMarshal", func() {
					var legacyUpload *dataTypesUploadTest.LegacyUpload
					var result *upload.Upload
					var createdTime time.Time
					var modifiedTime time.Time
					var deletedTime time.Time
					var recordTime time.Time

					BeforeEach(func() {
						legacyUpload = NewLegacyDataSet(userID, deviceID)
						recordTime = test.PastNearTime()
						createdTime = test.PastNearTime().AddDate(0, 0, 1)
						modifiedTime = test.PastNearTime().AddDate(0, 0, 2)
						deletedTime = test.PastNearTime().AddDate(0, 0, 3)
					})

					It("ensure string legacy dates are unmarshalled correctly", func() {
						legacyUpload.Time = legacyUpload.CreatedTime
						legacyUpload.Time = pointer.FromString(recordTime.Format(time.RFC3339Nano))
						legacyUpload.CreatedTime = pointer.FromString(createdTime.Format(time.RFC3339Nano))
						legacyUpload.ModifiedTime = pointer.FromString(modifiedTime.Format(time.RFC3339Nano))
						legacyUpload.DeletedTime = pointer.FromString(deletedTime.Format(time.RFC3339Nano))

						_, err := collection.InsertOne(context.Background(), legacyUpload)
						Expect(err).ToNot(HaveOccurred())

						err = collection.FindOne(context.Background(), bson.M{"_userId": userID}).Decode(&result)
						Expect(err).ToNot(HaveOccurred())

						Expect(*result.CreatedTime).To(Equal(createdTime))
						Expect(*result.ModifiedTime).To(Equal(modifiedTime))
						Expect(*result.DeletedTime).To(Equal(deletedTime))
						Expect(*result.Time).To(Equal(recordTime))
					})
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
								createdTime, _ := time.Parse(time.RFC3339, "2016-09-01T13:00:00Z")
								dataSet.DeletedTime = pointer.FromTime(createdTime)
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

						It("sets the created time and modified time", func() {
							Expect(repository.CreateDataSet(ctx, dataSet)).To(Succeed())
							Expect(dataSet.CreatedTime).ToNot(BeNil())
							Expect(dataSet.ModifiedTime).ToNot(BeNil())
							Expect(*dataSet.CreatedTime).To(Equal(*dataSet.ModifiedTime))
							Expect(dataSet.CreatedUserID).To(BeNil())
							Expect(dataSet.ByUser).To(BeNil())

							// Make sure the values are set in the db as well.
							var result *upload.Upload
							err := collection.FindOne(context.Background(), bson.M{"uploadId": dataSet.UploadID}).Decode(&result)
							Expect(err).ToNot(HaveOccurred())
							Expect(*result.CreatedTime).To(Equal(*dataSet.CreatedTime))
							Expect(*result.ModifiedTime).To(Equal(*dataSet.ModifiedTime))
							Expect(*result.CreatedTime).To(Equal(*result.ModifiedTime))

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
							createdTime := time.Now().UTC().Truncate(time.Millisecond)
							dataSet.CreatedTime = pointer.FromTime(createdTime)
							dataSet.ModifiedTime = pointer.FromTime(createdTime)
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

							It("updates modified time when updated", func() {
								newTime, err := time.Parse(time.RFC3339, "2022-01-01T11:00:00Z")
								Expect(err).ToNot(HaveOccurred())
								dataSet.Time = pointer.FromTime(newTime)
								dataSet.SetModifiedTime(pointer.FromTime(time.Now().UTC().Truncate(time.Millisecond)))
								update.Time = pointer.FromTime(newTime)
								_, err = repository.UpdateDataSet(ctx, id, update)
								Expect(err).ToNot(HaveOccurred())
								ValidateDataSetWithModifiedThreshold(collection, bson.M{"uploadId": dataSet.UploadID}, bson.M{}, time.Second, dataSet)
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
							// All newly created data now includes the modifiedTime as well.
							ValidateDataSet(collection, bson.M{"modifiedTime": bson.M{"$exists": true}}, bson.M{}, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo, dataSet)
							update = data.NewDataSetUpdate()
							update.State = pointer.FromString("closed")
							result, err := repository.UpdateDataSet(ctx, id, update)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).ToNot(BeNil())
							Expect(result.State).ToNot(BeNil())
							Expect(*result.State).To(Equal("closed"))
							ValidateDataSet(collection, bson.M{}, bson.M{}, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo, result)
							ValidateDataSet(collection, bson.M{"modifiedTime": bson.M{"$exists": true}}, bson.M{}, dataSetExistingOther, dataSetExistingOne, dataSetExistingTwo, result)
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

									It("succeeds and updates .modifiedTime", func() {
										Expect(repository.ArchiveDataSetData(ctx, dataSet, selectors)).To(Succeed())
										selectedDataSetData.SetActive(false)
										selectedDataSetData.SetModifiedTime(pointer.FromTime(time.Now().UTC().Truncate(time.Millisecond)))
										ValidateDataSetDataWithModifiedThreshold(collection, bson.M{"_active": false, "uploadId": dataSet.UploadID}, bson.M{"archivedTime": 0}, time.Second, selectedDataSetData)
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
								createdTime, _ := time.Parse(time.RFC3339, "2016-09-01T11:00:00Z")
								destroyDataSet.CreatedTime = pointer.FromTime(createdTime)
								destroyDataSet.ModifiedTime = pointer.FromTime(createdTime)
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
