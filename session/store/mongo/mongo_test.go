package mongo_test

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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
	createdTime := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now())
	expiresAt := test.RandomTimeFromRange(time.Now(), test.RandomTimeMaximum())
	thyme := test.RandomTimeFromRange(createdTime, time.Now())
	return bson.M{
		"_id":       test.RandomStringFromRangeAndCharset(32, 32, test.CharsetAlphaNumeric),
		"duration":  int32(86400),
		"createdAt": createdTime.Unix(),
		"expiresAt": expiresAt.Unix(),
		"time":      thyme.Unix(),
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

func ValidateSessions(testMongoCollection *mongo.Collection, selector bson.M, expectedSessions []interface{}) {
	var actualSessions []bson.M
	cursor, err := testMongoCollection.Find(context.Background(), selector)
	Expect(err).ToNot(HaveOccurred())
	Expect(cursor).ToNot(BeNil())
	Expect(cursor.All(context.Background(), &actualSessions)).To(Succeed())
	Expect(actualSessions).To(ConsistOf(expectedSessions...))
}

var _ = Describe("Mongo", func() {
	var mongoConfig *storeStructuredMongo.Config
	var mongoStore *sessionStoreMongo.Store
	var mongoRepository sessionStore.TokenRepository

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
			mongoStore, err = sessionStoreMongo.NewStore(nil)
			Expect(err).To(HaveOccurred())
			Expect(mongoStore).To(BeNil())
		})

		It("returns a new store and no error if successful", func() {
			var err error
			mongoStore, err = sessionStoreMongo.NewStore(mongoConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			mongoStore, err = sessionStoreMongo.NewStore(mongoConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})

		Context("NewSessionsSession", func() {
			It("returns a new session", func() {
				mongoRepository = mongoStore.NewTokenRepository()
				Expect(mongoRepository).ToNot(BeNil())
			})
		})

		Context("with a new session", func() {
			BeforeEach(func() {
				mongoRepository = mongoStore.NewTokenRepository()
				Expect(mongoRepository).ToNot(BeNil())
			})

			Context("with persisted data", func() {
				var collection *mongo.Collection
				var sessions []interface{}
				var ctx context.Context

				BeforeEach(func() {
					collection = mongoStore.GetCollection("tokens")
					sessions = append(NewServerSessions(), NewUserSessions(user.NewID())...)
					ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
				})

				JustBeforeEach(func() {
					_, err := collection.InsertMany(context.Background(), sessions)
					Expect(err).ToNot(HaveOccurred())
				})

				Context("DestroySessionsForUserByID", func() {
					var destroyUserID string
					var destroySessions []interface{}

					BeforeEach(func() {
						destroyUserID = user.NewID()
						destroySessions = NewUserSessions(destroyUserID)
					})

					JustBeforeEach(func() {
						_, err := collection.InsertMany(context.Background(), destroySessions)
						Expect(err).ToNot(HaveOccurred())
					})

					It("succeeds if it successfully removes sessions", func() {
						Expect(mongoRepository.DestroySessionsForUserByID(ctx, destroyUserID)).To(Succeed())
					})

					It("returns an error if the context is missing", func() {
						Expect(mongoRepository.DestroySessionsForUserByID(nil, destroyUserID)).To(MatchError("context is missing"))
					})

					It("returns an error if the user id is missing", func() {
						Expect(mongoRepository.DestroySessionsForUserByID(ctx, "")).To(MatchError("user id is missing"))
					})

					It("has the correct stored sessions", func() {
						ValidateSessions(collection, bson.M{}, append(sessions, destroySessions...))
						Expect(mongoRepository.DestroySessionsForUserByID(ctx, destroyUserID)).To(Succeed())
						ValidateSessions(collection, bson.M{}, sessions)
					})
				})
			})
		})
	})
})
