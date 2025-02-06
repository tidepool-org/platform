package mongo_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
)

var _ = Describe("Mongo", Label("mongodb", "slow", "integration"), func() {
	Context("Store", func() {
		var config *storeStructuredMongo.Config
		var store *storeStructuredMongo.Store
		var repository *storeStructuredMongo.Repository
		// unusedIPs for testing error paths. These IPs are defined in
		// TEST-NET-3, for documentation and examples, so in theory they
		// should be unused. Details at:
		// https://datatracker.ietf.org/doc/html/rfc5737
		unusedIPs := []string{
			"203.0.113.2", "203.0.113.3",
		}

		Context("NewStore", func() {
			BeforeEach(func() {
				config = storeStructuredMongoTest.NewConfig()
			})

			AfterEach(func() {
				if store != nil {
					err := store.Terminate(context.Background())
					Expect(err).ToNot(HaveOccurred())
				}
			})

			It("returns an error if the config is missing", func() {
				var err error
				store, err = storeStructuredMongo.NewStore(nil)
				Expect(err).To(MatchError("database config is empty"))
				Expect(store).To(BeNil())
			})

			It("returns an error if the config is invalid", func() {
				config.Addresses = nil
				var err error
				store, err = storeStructuredMongo.NewStore(config)
				Expect(err).To(MatchError("connection options are invalid; error parsing uri: must have at least 1 host"))
				Expect(store).To(BeNil())
			})

			It("returns an error if the addresses are not reachable", func() {
				var err error
				clientOptions := options.Client()
				clientOptions.SetTimeout(time.Millisecond)
				clientOptions.SetSocketTimeout(time.Millisecond)
				clientOptions.SetServerSelectionTimeout(time.Millisecond)
				clientOptions.SetConnectTimeout(time.Millisecond)
				config.Addresses = unusedIPs
				store, err = storeStructuredMongo.NewStoreFromClient(config, clientOptions)
				Expect(store).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())
				// We can't compare the exact error here, since different OSes display slightly different errors
				err = store.Ping(context.Background())
				Expect(err).To(HaveOccurred())
			})

			It("returns the correct status if the addresses are not reachable", func() {
				var err error
				clientOptions := options.Client()
				clientOptions.SetTimeout(time.Millisecond)
				clientOptions.SetSocketTimeout(time.Millisecond)
				clientOptions.SetServerSelectionTimeout(time.Millisecond)
				clientOptions.SetConnectTimeout(time.Millisecond)
				config.Addresses = unusedIPs
				store, err = storeStructuredMongo.NewStoreFromClient(config, clientOptions)
				Expect(store).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())
				status := store.Status(context.Background())
				Expect(status).ToNot(BeNil())
				Expect(status.Ping).To(Equal("FAILED"))
			})

			It("returns an error if the username or password is invalid", func() {
				config.Username = pointer.FromString("username")
				config.Password = pointer.FromString("password")
				var err error
				store, err = storeStructuredMongo.NewStore(config)
				Expect(store).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())
				err = store.Ping(context.Background())
				Expect(err).To(MatchError(`connection() error occurred during connection handshake: auth error: sasl conversation error: unable to authenticate using mechanism "SCRAM-SHA-1": (AuthenticationFailed) Authentication failed.`))
			})

			It("returns no error if successful", func() {
				var err error
				store, err = storeStructuredMongo.NewStore(config)
				Expect(err).ToNot(HaveOccurred())
				Expect(store).ToNot(BeNil())
			})
		})

		Context("with a new store", func() {
			BeforeEach(func() {
				store = storeStructuredMongoTest.GetSuiteStore()
			})

			Context("Status", func() {
				It("returns the appropriate status when initialized", func() {
					status := store.Status(context.Background())
					Expect(status).ToNot(BeNil())
					Expect(status.Ping).To(Equal("OK"))
				})
			})

			Context("GetRepository", func() {
				It("returns a new repository if no repository specified", func() {
					repository = store.GetRepository("")
					Expect(repository).ToNot(BeNil())
				})

				It("returns successfully", func() {
					repository = store.GetRepository("test")
					Expect(repository).ToNot(BeNil())
				})
			})

			Context("with a new repository", func() {
				BeforeEach(func() {
					repository = store.GetRepository("test")
					Expect(repository).ToNot(BeNil())
				})

				Context("CreateAllIndexes", func() {
					It("returns an error if the index is invalid", func() {
						Expect(repository.CreateAllIndexes(context.Background(), []mongo.IndexModel{{}})).To(MatchError("unable to create indexes; index model keys cannot be nil"))
					})

					It("returns successfully with nil indexes", func() {
						Expect(repository.CreateAllIndexes(context.Background(), nil)).To(Succeed())
					})

					It("returns successfully with empty indexes", func() {
						Expect(repository.CreateAllIndexes(context.Background(), []mongo.IndexModel{})).To(Succeed())
					})

					It("returns successfully with multiple indexes", func() {
						Expect(repository.CreateAllIndexes(context.Background(), []mongo.IndexModel{
							{
								Keys: bson.D{{Key: "one", Value: 1}},
								Options: options.Index().
									SetUnique(true).
									SetBackground(true),
							},
							{
								Keys: bson.D{{Key: "two", Value: 1}},
								Options: options.Index().
									SetBackground(true),
							},
							{
								Keys: bson.D{{Key: "three", Value: 1}},
							},
						})).To(Succeed())
					})
				})

				DescribeTable("ConstructUpdate",
					func(set bson.M, unset bson.M, operators []map[string]bson.M, expected bson.M) {
						Expect(repository.ConstructUpdate(set, unset, operators...)).To(Equal(expected))
					},
					Entry("where set is nil and unset is nil", nil, nil, nil, nil),
					Entry("where set is empty and unset is nil", bson.M{}, nil, nil, nil),
					Entry("where set is nil and unset is empty", nil, bson.M{}, nil, nil),
					Entry("where set is empty and unset is empty", bson.M{}, bson.M{}, nil, nil),
					Entry("where set is present and unset is nil", bson.M{"one": "alpha", "two": true}, nil, nil, bson.M{"$set": bson.M{"one": "alpha", "two": true}, "$inc": bson.M{"revision": 1}}),
					Entry("where set is present and unset is empty", bson.M{"one": "alpha", "two": true}, bson.M{}, nil, bson.M{"$set": bson.M{"one": "alpha", "two": true}, "$inc": bson.M{"revision": 1}}),
					Entry("where set is nil and unset is present", nil, bson.M{"three": "charlie", "four": false}, nil, bson.M{"$unset": bson.M{"three": "charlie", "four": false}, "$inc": bson.M{"revision": 1}}),
					Entry("where set is empty and unset is present", bson.M{}, bson.M{"three": "charlie", "four": false}, nil, bson.M{"$unset": bson.M{"three": "charlie", "four": false}, "$inc": bson.M{"revision": 1}}),
					Entry("where set is present and unset is present", bson.M{"one": "alpha", "two": true}, bson.M{"three": "charlie", "four": false}, nil, bson.M{"$set": bson.M{"one": "alpha", "two": true}, "$unset": bson.M{"three": "charlie", "four": false}, "$inc": bson.M{"revision": 1}}),
					Entry("where operators are present", bson.M{"one": "alpha", "two": true}, bson.M{"three": "charlie", "four": false}, []map[string]bson.M{{"$inc": {"alpha": -1}}, {"$unset": {"four": true}}, {"$set": {"two": false}}}, bson.M{"$set": bson.M{"one": "alpha", "two": false}, "$unset": bson.M{"three": "charlie", "four": true}, "$inc": bson.M{"alpha": -1, "revision": 1}}),
					Entry("where operators are all empty", nil, nil, []map[string]bson.M{{"one": bson.M{}, "two": bson.M{}}}, nil),
					Entry("where operators are partially empty", nil, nil, []map[string]bson.M{{"one": bson.M{}, "two": bson.M{"three": "four"}}}, bson.M{"two": bson.M{"three": "four"}, "$inc": bson.M{"revision": 1}}),
				)
			})
		})
	})

	Context("ModifyQuery", func() {
		It("returns nil if the query is nil", func() {
			Expect(storeStructuredMongo.ModifyQuery(nil,
				func(query bson.M) bson.M {
					query["alpha"] = "bravo"
					return query
				})).To(BeNil())
		})

		It("calls the modifiers that add fields", func() {
			Expect(storeStructuredMongo.ModifyQuery(bson.M{},
				func(query bson.M) bson.M {
					query["alpha"] = "bravo"
					return query
				},
				func(query bson.M) bson.M {
					query["charlie"] = "delta"
					return query
				},
			)).To(Equal(bson.M{"alpha": "bravo", "charlie": "delta"}))
		})

		It("calls the modifiers that set fields", func() {
			Expect(storeStructuredMongo.ModifyQuery(bson.M{"alpha": "bravo"},
				func(query bson.M) bson.M {
					return bson.M{"charlie": "delta"}
				},
			)).To(Equal(bson.M{"charlie": "delta"}))
		})

		It("calls the modifiers that removes fields", func() {
			Expect(storeStructuredMongo.ModifyQuery(bson.M{"alpha": "bravo", "charlie": "delta"},
				func(query bson.M) bson.M {
					delete(query, "charlie")
					return query
				},
			)).To(Equal(bson.M{"alpha": "bravo"}))
		})
	})

	Context("NotDeleted", func() {
		It("returns nil if the query is nil", func() {
			Expect(storeStructuredMongo.NotDeleted(nil)).To(BeNil())
		})

		It("adds the deleted time field to an empty query", func() {
			Expect(storeStructuredMongo.NotDeleted(bson.M{})).To(Equal(bson.M{"deletedTime": bson.M{"$exists": false}}))
		})

		It("adds the deleted time field to a non-empty query", func() {
			Expect(storeStructuredMongo.NotDeleted(bson.M{"alpha": "bravo"})).To(Equal(bson.M{"alpha": "bravo", "deletedTime": bson.M{"$exists": false}}))
		})
	})
})
