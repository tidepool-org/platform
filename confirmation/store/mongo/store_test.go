package mongo_test

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/tidepool-org/platform/confirmation/store"
	storeMongo "github.com/tidepool-org/platform/confirmation/store/mongo"
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

func ValidateConfirmations(collection *mongo.Collection, selector bson.M, expected []interface{}) {
	var actual []bson.M
	opts := options.Find().SetProjection(bson.M{"_id": 0})
	cursor, err := collection.Find(context.Background(), selector, opts)
	Expect(cursor).ToNot(BeNil())
	Expect(err).To(BeNil())
	Expect(cursor.All(context.Background(), &actual)).To(Succeed())
	Expect(actual).To(ConsistOf(expected))
}

var _ = Describe("Store", func() {
	var ctx context.Context
	var cfg *storeStructuredMongo.Config
	var str *storeMongo.Store
	var coll store.ConfirmationRepository

	BeforeEach(func() {
		ctx = log.NewContextWithLogger(context.Background(), logNull.NewLogger())
		cfg = storeStructuredMongoTest.NewConfig()
	})

	AfterEach(func() {
		if str != nil {
			str.Terminate(nil)
		}
	})

	Context("NewStore", func() {
		It("returns an error if unsuccessful", func() {
			var err error
			params := storeStructuredMongo.Params{DatabaseConfig: nil}
			str, err = storeMongo.NewStore(params)
			Expect(err).To(HaveOccurred())
			Expect(str).To(BeNil())
		})

		It("returns a new store and no error if successful", func() {
			var err error
			params := storeStructuredMongo.Params{DatabaseConfig: cfg}
			str, err = storeMongo.NewStore(params)
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		var collection *mongo.Collection

		BeforeEach(func() {
			var err error
			params := storeStructuredMongo.Params{DatabaseConfig: cfg}
			str, err = storeMongo.NewStore(params)
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
			collection = str.GetCollection("confirmations")
		})

		Context("EnsureIndexes", func() {
			It("returns successfully", func() {
				Expect(str.EnsureIndexes()).To(Succeed())
				cursor, err := collection.Indexes().List(context.Background())
				Expect(err).ToNot(HaveOccurred())
				Expect(cursor).ToNot(BeNil())
				var indexes []storeStructuredMongoTest.MongoIndex
				err = cursor.All(context.Background(), &indexes)
				Expect(err).ToNot(HaveOccurred())

				Expect(indexes).To(ConsistOf(
					MatchFields(IgnoreExtras, Fields{
						"Key": Equal(storeStructuredMongoTest.MakeKeySlice("_id")),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("email")),
						"Background": Equal(true),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("status")),
						"Background": Equal(true),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("type")),
						"Background": Equal(true),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("userId")),
						"Background": Equal(true),
					}),
				))
			})
		})

		Context("NewConfirmationSession", func() {
			It("returns a new confirmation session", func() {
				coll = str.NewConfirmationRepository()
				Expect(coll).ToNot(BeNil())
			})
		})

		Context("with a new confirmation session", func() {
			BeforeEach(func() {
				coll = str.NewConfirmationRepository()
				Expect(coll).ToNot(BeNil())
			})

			Context("with persisted data", func() {
				var confirmations []interface{}

				BeforeEach(func() {
					confirmations = NewConfirmations(test.RandomStringFromRangeAndCharset(10, 10, test.CharsetHexidecimalLowercase), test.RandomStringFromRangeAndCharset(10, 10, test.CharsetHexidecimalLowercase))
					_, err := collection.InsertMany(context.Background(), confirmations)
					Expect(err).ToNot(HaveOccurred())
				})

				Context("DeleteUserConfirmations", func() {
					var userID string
					var userConfirmations []interface{}

					BeforeEach(func() {
						userID = test.RandomStringFromRangeAndCharset(10, 10, test.CharsetHexidecimalLowercase)
						userConfirmations = NewConfirmations(userID, test.RandomStringFromRangeAndCharset(10, 10, test.CharsetHexidecimalLowercase))
						_, err := collection.InsertMany(context.Background(), userConfirmations)
						Expect(err).ToNot(HaveOccurred())
					})

					It("returns an error if the context is missing", func() {
						Expect(coll.DeleteUserConfirmations(nil, userID)).To(MatchError("context is missing"))
					})

					It("returns an error if the user id is missing", func() {
						Expect(coll.DeleteUserConfirmations(ctx, "")).To(MatchError("user id is missing"))
					})

					It("succeeds if it successfully removes confirmations", func() {
						Expect(coll.DeleteUserConfirmations(ctx, userID)).To(Succeed())
					})

					It("has the correct stored confirmations", func() {
						ValidateConfirmations(collection, bson.M{}, append(confirmations, userConfirmations...))
						Expect(coll.DeleteUserConfirmations(ctx, userID)).To(Succeed())
						ValidateConfirmations(collection, bson.M{}, confirmations)
					})
				})
			})
		})
	})
})
