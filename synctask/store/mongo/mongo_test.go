package mongo_test

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	synctaskStore "github.com/tidepool-org/platform/synctask/store"
	synctaskStoreMongo "github.com/tidepool-org/platform/synctask/store/mongo"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/user"
)

func NewSyncTask(userID string) bson.M {
	createdTime := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now())
	modifiedTime := test.RandomTimeFromRange(createdTime, time.Now())
	return bson.M{
		"_createdTime":  createdTime.Format(time.RFC3339Nano),
		"_modifiedTime": modifiedTime.Format(time.RFC3339Nano),
		"_storage": bson.M{
			"bucket":     "shovel",
			"encryption": "none",
			"key":        "1234567890",
			"region":     "world",
			"type":       "aws/s3",
		},
		"_userId": userID,
		"status":  "success",
	}
}

func NewSyncTasks(userID string) []interface{} {
	syncTasks := []interface{}{}
	for count := 0; count < 3; count++ {
		syncTasks = append(syncTasks, NewSyncTask(userID))
	}
	return syncTasks
}

func ValidateSyncTasks(testMongoCollection *mongo.Collection, selector bson.M, expectedSyncTasks []interface{}) {
	var actualSyncTasks []bson.M
	opts := options.Find().SetProjection(bson.M{"_id": 0})
	cursor, err := testMongoCollection.Find(context.Background(), selector, opts)
	Expect(err).ToNot(HaveOccurred())
	Expect(cursor).ToNot(BeNil())
	Expect(cursor.All(context.Background(), &actualSyncTasks)).To(Succeed())
	Expect(actualSyncTasks).To(ConsistOf(expectedSyncTasks))
}

var _ = Describe("Mongo", func() {
	var mongoConfig *storeStructuredMongo.Config
	var mongoStore *synctaskStoreMongo.Store
	var mongoRepository synctaskStore.SyncTaskRepository

	BeforeEach(func() {
		mongoConfig = storeStructuredMongoTest.NewConfig()
	})

	AfterEach(func() {
		if mongoStore != nil {
			mongoStore.Terminate(context.Background())
		}
	})

	Context("New", func() {
		It("returns an error if unsuccessful", func() {
			var err error
			mongoStore, err = synctaskStoreMongo.NewStore(nil)
			Expect(err).To(HaveOccurred())
			Expect(mongoStore).To(BeNil())
		})

		It("returns a new store and no error if successful", func() {
			var err error
			mongoStore, err = synctaskStoreMongo.NewStore(mongoConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			mongoStore, err = synctaskStoreMongo.NewStore(mongoConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})

		Context("NewSyncTaskRepository", func() {
			It("returns a new repository", func() {
				mongoRepository = mongoStore.NewSyncTaskRepository()
				Expect(mongoRepository).ToNot(BeNil())
			})
		})

		Context("with a new repository", func() {
			BeforeEach(func() {
				mongoRepository = mongoStore.NewSyncTaskRepository()
				Expect(mongoRepository).ToNot(BeNil())
			})

			Context("with persisted data", func() {
				var testMongoCollection *mongo.Collection
				var syncTasks []interface{}
				var ctx context.Context

				BeforeEach(func() {
					testMongoCollection = mongoStore.GetCollection("syncTasks")
					syncTasks = NewSyncTasks(user.NewID())
					ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
				})

				JustBeforeEach(func() {
					_, err := testMongoCollection.InsertMany(context.Background(), syncTasks)
					Expect(err).ToNot(HaveOccurred())
				})

				Context("DestroySyncTasksForUserByID", func() {
					var destroyUserID string
					var destroySyncTasks []interface{}

					BeforeEach(func() {
						destroyUserID = user.NewID()
						destroySyncTasks = NewSyncTasks(destroyUserID)
					})

					JustBeforeEach(func() {
						_, err := testMongoCollection.InsertMany(context.Background(), destroySyncTasks)
						Expect(err).ToNot(HaveOccurred())
					})

					It("succeeds if it successfully removes sync tasks", func() {
						Expect(mongoRepository.DestroySyncTasksForUserByID(ctx, destroyUserID)).To(Succeed())
					})

					It("returns an error if the context is missing", func() {
						Expect(mongoRepository.DestroySyncTasksForUserByID(nil, destroyUserID)).To(MatchError("context is missing"))
					})

					It("returns an error if the user id is missing", func() {
						Expect(mongoRepository.DestroySyncTasksForUserByID(ctx, "")).To(MatchError("user id is missing"))
					})

					It("has the correct stored sync tasks", func() {
						ValidateSyncTasks(testMongoCollection, bson.M{}, append(syncTasks, destroySyncTasks...))
						Expect(mongoRepository.DestroySyncTasksForUserByID(ctx, destroyUserID)).To(Succeed())
						ValidateSyncTasks(testMongoCollection, bson.M{}, syncTasks)
					})
				})
			})
		})
	})
})
