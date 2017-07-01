package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/log"
	baseMongo "github.com/tidepool-org/platform/store/mongo"
	"github.com/tidepool-org/platform/task/store"
	"github.com/tidepool-org/platform/task/store/mongo"
	testMongo "github.com/tidepool-org/platform/test/mongo"
)

func NewTask(userID string) bson.M {
	return bson.M{
		"_createdTime":  time.Now().UTC().Format(time.RFC3339),
		"_modifiedTime": time.Now().UTC().Format(time.RFC3339),
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

func NewTasks(userID string) []interface{} {
	tasks := []interface{}{}
	for count := 0; count < 3; count++ {
		tasks = append(tasks, NewTask(userID))
	}
	return tasks
}

func ValidateTasks(testMongoCollection *mgo.Collection, selector bson.M, expectedTasks []interface{}) {
	var actualTasks []interface{}
	Expect(testMongoCollection.Find(selector).Select(bson.M{"_id": 0}).All(&actualTasks)).To(Succeed())
	Expect(actualTasks).To(ConsistOf(expectedTasks))
}

var _ = Describe("Mongo", func() {
	var mongoConfig *baseMongo.Config
	var mongoStore *mongo.Store
	var mongoSession store.Session

	BeforeEach(func() {
		mongoConfig = &baseMongo.Config{
			Addresses:  testMongo.Address(),
			Database:   testMongo.Database(),
			Collection: testMongo.NewCollectionName(),
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
		It("returns an error if unsuccessful", func() {
			var err error
			mongoStore, err = mongo.New(nil, nil)
			Expect(err).To(HaveOccurred())
			Expect(mongoStore).To(BeNil())
		})

		It("returns a new store and no error if successful", func() {
			var err error
			mongoStore, err = mongo.New(log.NewNull(), mongoConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			mongoStore, err = mongo.New(log.NewNull(), mongoConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})

		Context("NewSession", func() {
			It("returns a new session if no logger specified", func() {
				mongoSession = mongoStore.NewSession(nil)
				Expect(mongoSession).ToNot(BeNil())
				Expect(mongoSession.Logger()).ToNot(BeNil())
			})

			It("returns a new session if logger specified", func() {
				logger := log.NewNull()
				mongoSession = mongoStore.NewSession(logger)
				Expect(mongoSession).ToNot(BeNil())
				Expect(mongoSession.Logger()).To(Equal(logger))
			})
		})

		Context("with a new session", func() {
			BeforeEach(func() {
				mongoSession = mongoStore.NewSession(log.NewNull())
				Expect(mongoSession).ToNot(BeNil())
			})

			Context("with persisted data", func() {
				var testMongoSession *mgo.Session
				var testMongoCollection *mgo.Collection
				var tasks []interface{}

				BeforeEach(func() {
					testMongoSession = testMongo.Session().Copy()
					testMongoCollection = testMongoSession.DB(mongoConfig.Database).C(mongoConfig.Collection)
					tasks = NewTasks(app.NewID())
				})

				JustBeforeEach(func() {
					Expect(testMongoCollection.Insert(tasks...)).To(Succeed())
				})

				AfterEach(func() {
					if testMongoSession != nil {
						testMongoSession.Close()
					}
				})

				Context("DestroyTasksForUserByID", func() {
					var destroyUserID string
					var destroyTasks []interface{}

					BeforeEach(func() {
						destroyUserID = app.NewID()
						destroyTasks = NewTasks(destroyUserID)
					})

					JustBeforeEach(func() {
						Expect(testMongoCollection.Insert(destroyTasks...)).To(Succeed())
					})

					It("succeeds if it successfully removes tasks", func() {
						Expect(mongoSession.DestroyTasksForUserByID(destroyUserID)).To(Succeed())
					})

					It("returns an error if the user id is missing", func() {
						Expect(mongoSession.DestroyTasksForUserByID("")).To(MatchError("mongo: user id is missing"))
					})

					It("returns an error if the session is closed", func() {
						mongoSession.Close()
						Expect(mongoSession.DestroyTasksForUserByID(destroyUserID)).To(MatchError("mongo: session closed"))
					})

					It("has the correct stored tasks", func() {
						ValidateTasks(testMongoCollection, bson.M{}, append(tasks, destroyTasks...))
						Expect(mongoSession.DestroyTasksForUserByID(destroyUserID)).To(Succeed())
						ValidateTasks(testMongoCollection, bson.M{}, tasks)
					})
				})
			})
		})
	})
})
