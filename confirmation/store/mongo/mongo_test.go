package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"context"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/confirmation/store"
	"github.com/tidepool-org/platform/confirmation/store/mongo"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
	baseMongo "github.com/tidepool-org/platform/store/mongo"
	testMongo "github.com/tidepool-org/platform/test/mongo"
)

func NewConfirmation(userID string, confirmationType string) bson.M {
	return bson.M{
		"created":   time.Now().UTC().Format(time.RFC3339),
		"creator":   bson.M{},
		"creatorId": "",
		"email":     id.New(),
		"modified":  time.Now().UTC().Format(time.RFC3339),
		"status":    "completed",
		"type":      confirmationType,
		"userId":    userID,
	}
}

func NewConfirmations(userID string, otherID string) []interface{} {
	confirmations := []interface{}{}
	for count := 0; count < 3; count++ {
		confirmations = append(confirmations, NewConfirmation(userID, "signup_confirmation"))
		confirmations = append(confirmations, NewConfirmation(userID, "password_reset"))
		confirmation := NewConfirmation(userID, "careteam_invitation")
		confirmation["creatorId"] = otherID
		confirmations = append(confirmations, confirmation)
		confirmation = NewConfirmation(otherID, "careteam_invitation")
		confirmation["creatorId"] = userID
		confirmations = append(confirmations, confirmation)
	}
	return confirmations
}

func ValidateConfirmations(testMongoCollection *mgo.Collection, selector bson.M, expectedConfirmations []interface{}) {
	var actualConfirmations []interface{}
	Expect(testMongoCollection.Find(selector).Select(bson.M{"_id": 0}).All(&actualConfirmations)).To(Succeed())
	Expect(actualConfirmations).To(ConsistOf(expectedConfirmations))
}

var _ = Describe("Mongo", func() {
	var ctx context.Context
	var mongoConfig *baseMongo.Config
	var mongoStore *mongo.Store
	var mongoSession store.ConfirmationsSession

	BeforeEach(func() {
		ctx = log.NewContextWithLogger(context.Background(), logNull.NewLogger())
		mongoConfig = &baseMongo.Config{
			Addresses:        []string{testMongo.Address()},
			Database:         testMongo.Database(),
			CollectionPrefix: testMongo.NewCollectionPrefix(),
			Timeout:          5 * time.Second,
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
			mongoStore, err = mongo.New(mongoConfig, logNull.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			mongoStore, err = mongo.New(mongoConfig, logNull.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})

		Context("NewConfirmationsSession", func() {
			It("returns a new session", func() {
				mongoSession = mongoStore.NewConfirmationsSession()
				Expect(mongoSession).ToNot(BeNil())
			})
		})

		Context("with a new session", func() {
			BeforeEach(func() {
				mongoSession = mongoStore.NewConfirmationsSession()
				Expect(mongoSession).ToNot(BeNil())
			})

			Context("with persisted data", func() {
				var testMongoSession *mgo.Session
				var testMongoCollection *mgo.Collection
				var confirmations []interface{}

				BeforeEach(func() {
					testMongoSession = testMongo.Session().Copy()
					testMongoCollection = testMongoSession.DB(mongoConfig.Database).C(mongoConfig.CollectionPrefix + "confirmations")
					confirmations = NewConfirmations(id.New(), id.New())
				})

				JustBeforeEach(func() {
					Expect(testMongoCollection.Insert(confirmations...)).To(Succeed())
				})

				AfterEach(func() {
					if testMongoSession != nil {
						testMongoSession.Close()
					}
				})

				Context("DestroyConfirmationsForUserByID", func() {
					var destroyUserID string
					var destroyOtherID string
					var destroyConfirmations []interface{}

					BeforeEach(func() {
						destroyUserID = id.New()
						destroyOtherID = id.New()
						destroyConfirmations = NewConfirmations(destroyUserID, destroyOtherID)
					})

					JustBeforeEach(func() {
						Expect(testMongoCollection.Insert(destroyConfirmations...)).To(Succeed())
					})

					It("succeeds if it successfully removes confirmations", func() {
						Expect(mongoSession.DestroyConfirmationsForUserByID(ctx, destroyUserID)).To(Succeed())
					})

					It("returns an error if the user id is missing", func() {
						Expect(mongoSession.DestroyConfirmationsForUserByID(ctx, "")).To(MatchError("user id is missing"))
					})

					It("returns an error if the session is closed", func() {
						mongoSession.Close()
						Expect(mongoSession.DestroyConfirmationsForUserByID(ctx, destroyUserID)).To(MatchError("session closed"))
					})

					It("has the correct stored confirmations", func() {
						ValidateConfirmations(testMongoCollection, bson.M{}, append(confirmations, destroyConfirmations...))
						Expect(mongoSession.DestroyConfirmationsForUserByID(ctx, destroyUserID)).To(Succeed())
						ValidateConfirmations(testMongoCollection, bson.M{}, confirmations)
					})
				})
			})
		})
	})
})
