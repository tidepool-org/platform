package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/notification/store"
	"github.com/tidepool-org/platform/notification/store/mongo"
	baseMongo "github.com/tidepool-org/platform/store/mongo"
	testMongo "github.com/tidepool-org/platform/test/mongo"
)

func NewNotification(userID string, notificationType string) bson.M {
	return bson.M{
		"created":   time.Now().UTC().Format(time.RFC3339),
		"creator":   bson.M{},
		"creatorId": "",
		"email":     app.NewID(),
		"modified":  time.Now().UTC().Format(time.RFC3339),
		"status":    "completed",
		"type":      notificationType,
		"userId":    userID,
	}
}

func NewNotifications(userID string, otherID string) []interface{} {
	notifications := []interface{}{}
	for count := 0; count < 3; count++ {
		notifications = append(notifications, NewNotification(userID, "signup_confirmation"))
		notifications = append(notifications, NewNotification(userID, "password_reset"))
		notification := NewNotification(userID, "careteam_invitation")
		notification["creatorId"] = otherID
		notifications = append(notifications, notification)
		notification = NewNotification(otherID, "careteam_invitation")
		notification["creatorId"] = userID
		notifications = append(notifications, notification)
	}
	return notifications
}

func ValidateNotifications(testMongoCollection *mgo.Collection, selector bson.M, expectedNotifications []interface{}) {
	var actualNotifications []interface{}
	Expect(testMongoCollection.Find(selector).Select(bson.M{"_id": 0}).All(&actualNotifications)).To(Succeed())
	Expect(actualNotifications).To(ConsistOf(expectedNotifications))
}

var _ = Describe("Mongo", func() {
	var mongoConfig *baseMongo.Config
	var mongoStore *mongo.Store
	var mongoSession store.Session

	BeforeEach(func() {
		mongoConfig = &baseMongo.Config{
			Addresses:  []string{testMongo.Address()},
			Database:   testMongo.Database(),
			Collection: testMongo.NewCollectionName(),
			Timeout:    5 * time.Second,
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
				var notifications []interface{}

				BeforeEach(func() {
					testMongoSession = testMongo.Session().Copy()
					testMongoCollection = testMongoSession.DB(mongoConfig.Database).C(mongoConfig.Collection)
					notifications = NewNotifications(app.NewID(), app.NewID())
				})

				JustBeforeEach(func() {
					Expect(testMongoCollection.Insert(notifications...)).To(Succeed())
				})

				AfterEach(func() {
					if testMongoSession != nil {
						testMongoSession.Close()
					}
				})

				Context("DestroyNotificationsForUserByID", func() {
					var destroyUserID string
					var destroyOtherID string
					var destroyNotifications []interface{}

					BeforeEach(func() {
						destroyUserID = app.NewID()
						destroyOtherID = app.NewID()
						destroyNotifications = NewNotifications(destroyUserID, destroyOtherID)
					})

					JustBeforeEach(func() {
						Expect(testMongoCollection.Insert(destroyNotifications...)).To(Succeed())
					})

					It("succeeds if it successfully removes notifications", func() {
						Expect(mongoSession.DestroyNotificationsForUserByID(destroyUserID)).To(Succeed())
					})

					It("returns an error if the user id is missing", func() {
						Expect(mongoSession.DestroyNotificationsForUserByID("")).To(MatchError("mongo: user id is missing"))
					})

					It("returns an error if the session is closed", func() {
						mongoSession.Close()
						Expect(mongoSession.DestroyNotificationsForUserByID(destroyUserID)).To(MatchError("mongo: session closed"))
					})

					It("has the correct stored notifications", func() {
						ValidateNotifications(testMongoCollection, bson.M{}, append(notifications, destroyNotifications...))
						Expect(mongoSession.DestroyNotificationsForUserByID(destroyUserID)).To(Succeed())
						ValidateNotifications(testMongoCollection, bson.M{}, notifications)
					})
				})
			})
		})
	})
})
