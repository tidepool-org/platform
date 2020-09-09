package mongo_test

import (
	"context"
	"time"

	mgo "github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	. "github.com/onsi/ginkgo"
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

func ValidateSyncTasks(testMongoCollection *mgo.Collection, selector bson.M, expectedSyncTasks []interface{}) {
	var actualSyncTasks []interface{}
	Expect(testMongoCollection.Find(selector).Select(bson.M{"_id": 0}).All(&actualSyncTasks)).To(Succeed())
	Expect(actualSyncTasks).To(ConsistOf(expectedSyncTasks))
}

var _ = Describe("Mongo", func() {
	var mongoConfig *storeStructuredMongo.Config
	var mongoStore *synctaskStoreMongo.Store
	var mongoSession synctaskStore.SyncTaskSession

	BeforeEach(func() {
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
			mongoStore, err = synctaskStoreMongo.NewStore(nil, nil)
			Expect(err).To(HaveOccurred())
			Expect(mongoStore).To(BeNil())
		})

		It("returns a new store and no error if successful", func() {
			var err error
			mongoStore, err = synctaskStoreMongo.NewStore(mongoConfig, logTest.NewLogger())
			mongoStore.WaitUntilStarted()
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			mongoStore, err = synctaskStoreMongo.NewStore(mongoConfig, logTest.NewLogger())
			mongoStore.WaitUntilStarted()
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})

		Context("NewSyncTaskSession", func() {
			It("returns a new session", func() {
				mongoSession = mongoStore.NewSyncTaskSession()
				Expect(mongoSession).ToNot(BeNil())
			})
		})

		Context("with a new session", func() {
			BeforeEach(func() {
				mongoSession = mongoStore.NewSyncTaskSession()
				Expect(mongoSession).ToNot(BeNil())
			})

			Context("with persisted data", func() {
				var testMongoSession *mgo.Session
				var testMongoCollection *mgo.Collection
				var syncTasks []interface{}
				var ctx context.Context

				BeforeEach(func() {
					testMongoSession = storeStructuredMongoTest.Session().Copy()
					testMongoCollection = testMongoSession.DB(mongoConfig.Database).C(mongoConfig.CollectionPrefix + "syncTasks")
					syncTasks = NewSyncTasks(user.NewID())
					ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
				})

				JustBeforeEach(func() {
					Expect(testMongoCollection.Insert(syncTasks...)).To(Succeed())
				})

				AfterEach(func() {
					if testMongoSession != nil {
						testMongoSession.Close()
					}
				})

				Context("DestroySyncTasksForUserByID", func() {
					var destroyUserID string
					var destroySyncTasks []interface{}

					BeforeEach(func() {
						destroyUserID = user.NewID()
						destroySyncTasks = NewSyncTasks(destroyUserID)
					})

					JustBeforeEach(func() {
						Expect(testMongoCollection.Insert(destroySyncTasks...)).To(Succeed())
					})

					It("succeeds if it successfully removes sync tasks", func() {
						Expect(mongoSession.DestroySyncTasksForUserByID(ctx, destroyUserID)).To(Succeed())
					})

					It("returns an error if the context is missing", func() {
						Expect(mongoSession.DestroySyncTasksForUserByID(nil, destroyUserID)).To(MatchError("context is missing"))
					})

					It("returns an error if the user id is missing", func() {
						Expect(mongoSession.DestroySyncTasksForUserByID(ctx, "")).To(MatchError("user id is missing"))
					})

					It("returns an error if the session is closed", func() {
						mongoSession.Close()
						Expect(mongoSession.DestroySyncTasksForUserByID(ctx, destroyUserID)).To(MatchError("session closed"))
					})

					It("has the correct stored sync tasks", func() {
						ValidateSyncTasks(testMongoCollection, bson.M{}, append(syncTasks, destroySyncTasks...))
						Expect(mongoSession.DestroySyncTasksForUserByID(ctx, destroyUserID)).To(Succeed())
						ValidateSyncTasks(testMongoCollection, bson.M{}, syncTasks)
					})
				})
			})
		})
	})
})
