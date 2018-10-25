package mongo_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	sessionStore "github.com/tidepool-org/platform/session/store"
	sessionStoreMongo "github.com/tidepool-org/platform/session/store/mongo"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/user"
)

func NewBaseSession() bson.M {
	now := time.Now()
	return bson.M{
		"_id":       test.RandomStringFromRangeAndCharset(32, 32, test.CharsetAlphaNumeric),
		"duration":  86400,
		"expiresAt": now.Add(86400 * time.Second).Unix(),
		"createdAt": now.Unix(),
		"time":      now.Unix(),
	}
}

func NewServerSession() bson.M {
	session := NewBaseSession()
	session["isServer"] = true
	session["serverId"] = test.RandomString()
	return session
}

func NewUserSession(userID string) bson.M {
	session := NewBaseSession()
	session["isServer"] = false
	session["userId"] = userID
	return session
}

func NewServerSessions() []interface{} {
	sessions := []interface{}{}
	sessions = append(sessions, NewServerSession(), NewServerSession(), NewServerSession())
	return sessions
}

func NewUserSessions(userID string) []interface{} {
	sessions := []interface{}{}
	sessions = append(sessions, NewUserSession(userID), NewUserSession(userID), NewUserSession(userID))
	return sessions
}

func ValidateSessions(testMongoCollection *mgo.Collection, selector bson.M, expectedSessions []interface{}) {
	var actualSessions []interface{}
	Expect(testMongoCollection.Find(selector).All(&actualSessions)).To(Succeed())
	Expect(actualSessions).To(ConsistOf(expectedSessions...))
}

var _ = Describe("Mongo", func() {
	var mongoConfig *storeStructuredMongo.Config
	var mongoStore *sessionStoreMongo.Store
	var mongoSession sessionStore.SessionsSession

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
			mongoStore, err = sessionStoreMongo.NewStore(nil, nil)
			Expect(err).To(HaveOccurred())
			Expect(mongoStore).To(BeNil())
		})

		It("returns a new store and no error if successful", func() {
			var err error
			mongoStore, err = sessionStoreMongo.NewStore(mongoConfig, logTest.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			mongoStore, err = sessionStoreMongo.NewStore(mongoConfig, logTest.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})

		Context("NewSessionsSession", func() {
			It("returns a new session", func() {
				mongoSession = mongoStore.NewSessionsSession()
				Expect(mongoSession).ToNot(BeNil())
			})
		})

		Context("with a new session", func() {
			BeforeEach(func() {
				mongoSession = mongoStore.NewSessionsSession()
				Expect(mongoSession).ToNot(BeNil())
			})

			Context("with persisted data", func() {
				var testMongoSession *mgo.Session
				var testMongoCollection *mgo.Collection
				var sessions []interface{}
				var ctx context.Context

				BeforeEach(func() {
					testMongoSession = storeStructuredMongoTest.Session().Copy()
					testMongoCollection = testMongoSession.DB(mongoConfig.Database).C(mongoConfig.CollectionPrefix + "tokens")
					sessions = append(NewServerSessions(), NewUserSessions(user.NewID())...)
					ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
				})

				JustBeforeEach(func() {
					Expect(testMongoCollection.Insert(sessions...)).To(Succeed())
				})

				AfterEach(func() {
					if testMongoSession != nil {
						testMongoSession.Close()
					}
				})

				Context("DestroySessionsForUserByID", func() {
					var destroyUserID string
					var destroySessions []interface{}

					BeforeEach(func() {
						destroyUserID = user.NewID()
						destroySessions = NewUserSessions(destroyUserID)
					})

					JustBeforeEach(func() {
						Expect(testMongoCollection.Insert(destroySessions...)).To(Succeed())
					})

					It("succeeds if it successfully removes sessions", func() {
						Expect(mongoSession.DestroySessionsForUserByID(ctx, destroyUserID)).To(Succeed())
					})

					It("returns an error if the context is missing", func() {
						Expect(mongoSession.DestroySessionsForUserByID(nil, destroyUserID)).To(MatchError("context is missing"))
					})

					It("returns an error if the user id is missing", func() {
						Expect(mongoSession.DestroySessionsForUserByID(ctx, "")).To(MatchError("user id is missing"))
					})

					It("returns an error if the session is closed", func() {
						mongoSession.Close()
						Expect(mongoSession.DestroySessionsForUserByID(ctx, destroyUserID)).To(MatchError("session closed"))
					})

					It("has the correct stored sessions", func() {
						ValidateSessions(testMongoCollection, bson.M{}, append(sessions, destroySessions...))
						Expect(mongoSession.DestroySessionsForUserByID(ctx, destroyUserID)).To(Succeed())
						ValidateSessions(testMongoCollection, bson.M{}, sessions)
					})
				})
			})
		})
	})
})
