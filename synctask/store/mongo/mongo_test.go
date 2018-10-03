package mongo_test

// import (
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"

// 	"time"

// 	mgo "gopkg.in/mgo.v2"
// 	"gopkg.in/mgo.v2/bson"

// 	"github.com/tidepool-org/platform/id"
// 	"github.com/tidepool-org/platform/log/null"
// 	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
// 	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
// 	"github.com/tidepool-org/platform/synctask/store"
// 	"github.com/tidepool-org/platform/synctask/store/mongo"
// )

// func NewSyncTask(userID string) bson.M {
// 	return bson.M{
// 		"_createdTime":  time.Now().UTC().Format(time.RFC3339),
// 		"_modifiedTime": time.Now().UTC().Format(time.RFC3339),
// 		"_storage": bson.M{
// 			"bucket":     "shovel",
// 			"encryption": "none",
// 			"key":        "1234567890",
// 			"region":     "world",
// 			"type":       "aws/s3",
// 		},
// 		"_userId": userID,
// 		"status":  "success",
// 	}
// }

// func NewSyncTasks(userID string) []interface{} {
// 	syncTasks := []interface{}{}
// 	for count := 0; count < 3; count++ {
// 		syncTasks = append(syncTasks, NewSyncTask(userID))
// 	}
// 	return syncTasks
// }

// func ValidateSyncTasks(testMongoCollection *mgo.Collection, selector bson.M, expectedSyncTasks []interface{}) {
// 	var actualSyncTasks []interface{}
// 	Expect(testMongoCollection.Find(selector).Select(bson.M{"_id": 0}).All(&actualSyncTasks)).To(Succeed())
// 	Expect(actualSyncTasks).To(ConsistOf(expectedSyncTasks))
// }

// var _ = Describe("Mongo", func() {
// 	var mongoConfig *storeStructuredMongo.Config
// 	var mongoStore *mongo.Store
// 	var mongoSession store.SyncTaskSession

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

// 		Context("NewSyncTaskSession", func() {
// 			It("returns a new session", func() {
// 				mongoSession = mongoStore.NewSyncTaskSession()
// 				Expect(mongoSession).ToNot(BeNil())
// 			})
// 		})

// 		Context("with a new session", func() {
// 			BeforeEach(func() {
// 				mongoSession = mongoStore.NewSyncTaskSession()
// 				Expect(mongoSession).ToNot(BeNil())
// 			})

// 			Context("with persisted data", func() {
// 				var testMongoSession *mgo.Session
// 				var testMongoCollection *mgo.Collection
// 				var syncTasks []interface{}

// 				BeforeEach(func() {
// 					testMongoSession = storeStructuredMongoTest.Session().Copy()
// 					testMongoCollection = testMongoSession.DB(mongoConfig.Database).C(mongoConfig.CollectionPrefix + "syncTasks")
// 					syncTasks = NewSyncTasks(id.New())
// 				})

// 				JustBeforeEach(func() {
// 					Expect(testMongoCollection.Insert(syncTasks...)).To(Succeed())
// 				})

// 				AfterEach(func() {
// 					if testMongoSession != nil {
// 						testMongoSession.Close()
// 					}
// 				})

// 				Context("DestroySyncTasksForUserByID", func() {
// 					var destroyUserID string
// 					var destroySyncTasks []interface{}

// 					BeforeEach(func() {
// 						destroyUserID = id.New()
// 						destroySyncTasks = NewSyncTasks(destroyUserID)
// 					})

// 					JustBeforeEach(func() {
// 						Expect(testMongoCollection.Insert(destroySyncTasks...)).To(Succeed())
// 					})

// 					It("succeeds if it successfully removes sync tasks", func() {
// 						Expect(mongoSession.DestroySyncTasksForUserByID(destroyUserID)).To(Succeed())
// 					})

// 					It("returns an error if the user id is missing", func() {
// 						Expect(mongoSession.DestroySyncTasksForUserByID("")).To(MatchError("user id is missing"))
// 					})

// 					It("returns an error if the session is closed", func() {
// 						mongoSession.Close()
// 						Expect(mongoSession.DestroySyncTasksForUserByID(destroyUserID)).To(MatchError("session closed"))
// 					})

// 					It("has the correct stored sync tasks", func() {
// 						ValidateSyncTasks(testMongoCollection, bson.M{}, append(syncTasks, destroySyncTasks...))
// 						Expect(mongoSession.DestroySyncTasksForUserByID(destroyUserID)).To(Succeed())
// 						ValidateSyncTasks(testMongoCollection, bson.M{}, syncTasks)
// 					})
// 				})
// 			})
// 		})
// 	})
// })
