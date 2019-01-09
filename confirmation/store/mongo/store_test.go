package mongo_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/confirmation/store"
	"github.com/tidepool-org/platform/confirmation/store/mongo"
	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
	netTest "github.com/tidepool-org/platform/net/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/test"
)

func NewConfirmation(userID string, typ string) bson.M {
	createdTime := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now())
	modifiedTime := test.RandomTimeFromRange(createdTime, time.Now())
	return bson.M{
		"created":   createdTime.Format(time.RFC3339Nano),
		"creator":   bson.M{},
		"creatorId": "",
		"email":     netTest.RandomEmail(),
		"modified":  modifiedTime.Format(time.RFC3339Nano),
		"status":    "completed",
		"type":      typ,
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

func ValidateConfirmations(mgoCollection *mgo.Collection, selector bson.M, expected []interface{}) {
	var actual []interface{}
	Expect(mgoCollection.Find(selector).Select(bson.M{"_id": 0}).All(&actual)).To(Succeed())
	Expect(actual).To(ConsistOf(expected))
}

var _ = Describe("Store", func() {
	var ctx context.Context
	var cfg *storeStructuredMongo.Config
	var str *mongo.Store
	var ssn store.ConfirmationSession

	BeforeEach(func() {
		ctx = log.NewContextWithLogger(context.Background(), logNull.NewLogger())
		cfg = storeStructuredMongoTest.NewConfig()
	})

	AfterEach(func() {
		if ssn != nil {
			ssn.Close()
		}
		if str != nil {
			str.Close()
		}
	})

	Context("NewStore", func() {
		It("returns an error if unsuccessful", func() {
			var err error
			str, err = mongo.NewStore(nil, nil)
			Expect(err).To(HaveOccurred())
			Expect(str).To(BeNil())
		})

		It("returns a new store and no error if successful", func() {
			var err error
			str, err = mongo.NewStore(cfg, logNull.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		var mgoSession *mgo.Session
		var mgoCollection *mgo.Collection

		BeforeEach(func() {
			var err error
			str, err = mongo.NewStore(cfg, logNull.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
			mgoSession = storeStructuredMongoTest.Session().Copy()
			mgoCollection = mgoSession.DB(cfg.Database).C(cfg.CollectionPrefix + "confirmations")
		})

		AfterEach(func() {
			if mgoSession != nil {
				mgoSession.Close()
			}
		})

		Context("EnsureIndexes", func() {
			It("returns successfully", func() {
				Expect(str.EnsureIndexes()).To(Succeed())
				indexes, err := mgoCollection.Indexes()
				Expect(err).ToNot(HaveOccurred())
				Expect(indexes).To(ConsistOf(
					MatchFields(IgnoreExtras, Fields{"Key": ConsistOf("_id")}),
					MatchFields(IgnoreExtras, Fields{"Key": ConsistOf("email")}),
					MatchFields(IgnoreExtras, Fields{"Key": ConsistOf("status")}),
					MatchFields(IgnoreExtras, Fields{"Key": ConsistOf("type")}),
					MatchFields(IgnoreExtras, Fields{"Key": ConsistOf("userId")}),
				))
			})
		})

		Context("NewConfirmationSession", func() {
			It("returns a new confirmation session", func() {
				ssn = str.NewConfirmationSession()
				Expect(ssn).ToNot(BeNil())
			})
		})

		Context("with a new confirmation session", func() {
			BeforeEach(func() {
				ssn = str.NewConfirmationSession()
				Expect(ssn).ToNot(BeNil())
			})

			Context("with persisted data", func() {
				var confirmations []interface{}

				BeforeEach(func() {
					confirmations = NewConfirmations(test.RandomStringFromRangeAndCharset(10, 10, test.CharsetHexidecimalLowercase), test.RandomStringFromRangeAndCharset(10, 10, test.CharsetHexidecimalLowercase))
					Expect(mgoCollection.Insert(confirmations...)).To(Succeed())
				})

				Context("DeleteUserConfirmations", func() {
					var userID string
					var userConfirmations []interface{}

					BeforeEach(func() {
						userID = test.RandomStringFromRangeAndCharset(10, 10, test.CharsetHexidecimalLowercase)
						userConfirmations = NewConfirmations(userID, test.RandomStringFromRangeAndCharset(10, 10, test.CharsetHexidecimalLowercase))
						Expect(mgoCollection.Insert(userConfirmations...)).To(Succeed())
					})

					It("returns an error if the context is missing", func() {
						Expect(ssn.DeleteUserConfirmations(nil, userID)).To(MatchError("context is missing"))
					})

					It("returns an error if the user id is missing", func() {
						Expect(ssn.DeleteUserConfirmations(ctx, "")).To(MatchError("user id is missing"))
					})

					It("returns an error if the session is closed", func() {
						ssn.Close()
						Expect(ssn.DeleteUserConfirmations(ctx, userID)).To(MatchError("session closed"))
					})

					It("succeeds if it successfully removes confirmations", func() {
						Expect(ssn.DeleteUserConfirmations(ctx, userID)).To(Succeed())
					})

					It("has the correct stored confirmations", func() {
						ValidateConfirmations(mgoCollection, bson.M{}, append(confirmations, userConfirmations...))
						Expect(ssn.DeleteUserConfirmations(ctx, userID)).To(Succeed())
						ValidateConfirmations(mgoCollection, bson.M{}, confirmations)
					})
				})
			})
		})
	})
})
