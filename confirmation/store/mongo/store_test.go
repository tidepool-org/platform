package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"context"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/confirmation/store"
	"github.com/tidepool-org/platform/confirmation/store/mongo"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
	storeMongo "github.com/tidepool-org/platform/store/mongo"
	testInternet "github.com/tidepool-org/platform/test/internet"
	testMongo "github.com/tidepool-org/platform/test/mongo"
)

func NewConfirmation(userID string, typ string) bson.M {
	now := time.Now().UTC()
	return bson.M{
		"created":   now.Format(time.RFC3339),
		"creator":   bson.M{},
		"creatorId": "",
		"email":     testInternet.NewEmail(),
		"modified":  now.Format(time.RFC3339),
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
	var cfg *storeMongo.Config
	var str *mongo.Store
	var ssn store.ConfirmationSession

	BeforeEach(func() {
		ctx = log.NewContextWithLogger(context.Background(), logNull.NewLogger())
		cfg = testMongo.NewConfig()
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
			mgoSession = testMongo.Session().Copy()
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

			Context("EnsureIndexes", func() {
				It("returns successfully", func() {
					Expect(ssn.EnsureIndexes()).To(Succeed())
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

			Context("with persisted data", func() {
				var confirmations []interface{}

				BeforeEach(func() {
					confirmations = NewConfirmations(id.New(), id.New())
					Expect(mgoCollection.Insert(confirmations...)).To(Succeed())
				})

				Context("DeleteUserConfirmations", func() {
					var userID string
					var userConfirmations []interface{}

					BeforeEach(func() {
						userID = id.New()
						userConfirmations = NewConfirmations(userID, id.New())
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
