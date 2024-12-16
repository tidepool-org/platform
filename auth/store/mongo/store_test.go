package mongo_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/auth/store"
	"github.com/tidepool-org/platform/auth/store/mongo"
	"github.com/tidepool-org/platform/devicetokens"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
)

var _ = Describe("Store", func() {
	var config *storeStructuredMongo.Config
	var str *mongo.Store

	BeforeEach(func() {
		config = storeStructuredMongoTest.NewConfig()
	})

	AfterEach(func() {
		if str != nil {
			str.Terminate(context.Background())
		}
	})

	Context("New", func() {
		It("returns an error if unsuccessful", func() {
			var err error
			str, err = mongo.NewStore(nil)
			Expect(err).To(HaveOccurred())
			Expect(str).To(BeNil())
		})

		It("returns successfully", func() {
			var err error
			str, err = mongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			str, err = mongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
		})

		// TODO: EnsureIndexes

		Context("NewProviderSessionRepository", func() {
			var repository store.ProviderSessionRepository

			It("returns successfully", func() {
				repository = str.NewProviderSessionRepository()
				Expect(repository).ToNot(BeNil())
			})
		})

		Context("NewRestrictedTokenRepository", func() {
			var repository store.RestrictedTokenRepository

			It("returns successfully", func() {
				repository = str.NewRestrictedTokenRepository()
				Expect(repository).ToNot(BeNil())
			})
		})

		Context("NewDeviceTokenRepository", func() {
			var repository store.DeviceTokenRepository
			var err error

			It("returns successfully", func() {
				repository = str.NewDeviceTokenRepository()
				Expect(repository).ToNot(BeNil())
			})

			Context("device tokens", func() {
				BeforeEach(func() {
					repository = str.NewDeviceTokenRepository()
					Expect(repository).ToNot(BeNil())
					_, err = str.GetCollection("deviceTokens").DeleteMany(context.Background(), bson.D{})
					Expect(err).To(Succeed())
				})

				prep := func(upsertDoc bool) (context.Context, *devicetokens.Document, bson.M) {
					doc := &devicetokens.Document{
						UserID:   "user-id",
						TokenKey: "foo",
					}
					ctx := context.Background()
					filter := bson.M{}
					if upsertDoc {
						Expect(repository.Upsert(ctx, doc)).
							To(Succeed())
						filter["userId"] = doc.UserID
						filter["tokenKey"] = doc.TokenKey
					}

					return ctx, doc, filter
				}

				Describe("Upsert", func() {
					Context("when no document exists", func() {
						It("creates a new document", func() {
							ctx, doc, filter := prep(false)

							Expect(repository.Upsert(ctx, doc)).To(Succeed())

							res := str.GetCollection("deviceTokens").FindOne(ctx, filter)
							Expect(res.Err()).To(Succeed())
							newDoc := &devicetokens.Document{}
							err := res.Decode(newDoc)
							Expect(err).ToNot(HaveOccurred())
							Expect(newDoc.UserID).To(Equal(doc.UserID))
							Expect(newDoc.TokenKey).To(Equal(doc.TokenKey))
						})
					})

					It("requires UserID and TokenID", func() {
						ctx, doc, _ := prep(false)

						doc.UserID = ""
						err := repository.Upsert(ctx, doc)
						Expect(err).To(MatchError("UserID is empty"))

						doc.UserID = "user-id"
						doc.TokenKey = ""
						err = repository.Upsert(ctx, doc)
						Expect(err).To(MatchError("TokenKey is empty"))
					})

					It("updates the existing document, instead of creating a duplicate", func() {
						ctx, doc, filter := prep(true)

						err := repository.Upsert(ctx, doc)
						Expect(err).To(Succeed())

						cur, err := str.GetCollection("deviceTokens").Find(ctx, filter)
						Expect(err).To(Succeed())
						Expect(cur.RemainingBatchLength()).To(Equal(1))
						for cur.Next(ctx) {
							newDoc := &devicetokens.Document{}
							err = cur.Decode(newDoc)
							Expect(err).To(Succeed())
							Expect(newDoc.UserID).To(Equal("user-id"))
							Expect(newDoc.TokenKey).To(Equal("foo"))
						}
					})
				})
			})
		})
	})
})
