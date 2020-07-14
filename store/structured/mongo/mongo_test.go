package mongo_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"

	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
)

type ProtoLifecycle struct{}

func (t ProtoLifecycle) Append(x fx.Hook) {}

var _ = Describe("Mongo", func() {
	Context("Store", func() {
		var params storeStructuredMongo.Params
		var store *storeStructuredMongo.Store
		var repository *storeStructuredMongo.Repository

		BeforeEach(func() {
			params = storeStructuredMongo.Params{
				DatabaseConfig: storeStructuredMongoTest.NewConfig(),
			}
		})

		AfterEach(func() {
			if store != nil {
				err := store.Terminate(context.Background())
				Expect(err).ToNot(HaveOccurred())
			}
		})

		Context("NewStore", func() {
			It("returns an error if the config is missing", func() {
				var err error
				store, err = storeStructuredMongo.NewStore(storeStructuredMongo.Params{})
				Expect(err).To(MatchError("database config is empty"))
				Expect(store).To(BeNil())
			})
		})

		Context("Initialize", func() {
			It("returns an error if the config is invalid", func() {
				params.DatabaseConfig.Addresses = nil
				var err error
				store, err = storeStructuredMongo.NewStore(params)
				Expect(err).To(MatchError("connection options are invalid; error parsing uri: must have at least 1 host"))
			})

			It("returns an error if the addresses are not reachable", func() {
				params.DatabaseConfig.Addresses = []string{"127.0.0.0", "127.0.0.0"}
				var err error
				store, err = storeStructuredMongo.NewStore(params)
				Expect(err).To(MatchError("server selection error: server selection timeout, current topology: { Type: Unknown, Servers: [{ Addr: 127.0.0.0:27017, Type: Unknown, State: Connected, Average RTT: 0 }, ] }"))
			})

			It("returns an error if the username or password is invalid", func() {
				params.DatabaseConfig.Username = pointer.FromString("username")
				params.DatabaseConfig.Password = pointer.FromString("password")
				var err error
				store, err = storeStructuredMongo.NewStore(params)
				Expect(err).To(MatchError("connection() : auth error: sasl conversation error: unable to authenticate using mechanism \"SCRAM-SHA-1\": (AuthenticationFailed) Authentication failed."))
			})

			It("returns an error if TLS is specified on a server that does not support it", func() {
				params.DatabaseConfig.TLS = true
				var err error
				store, err = storeStructuredMongo.NewStore(params)
				Expect(err).To(MatchError("server selection error: server selection timeout, current topology: { Type: Unknown, Servers: [{ Addr: localhost:27017, Type: Unknown, State: Connected, Average RTT: 0, Last error: connection() : EOF }, ] }"))
			})

			It("returns no error if successful", func() {
				var err error
				store, err = storeStructuredMongo.NewStore(params)
				Expect(err).ToNot(HaveOccurred())
				Expect(store).ToNot(BeNil())
			})
		})

		Context("with an uninitialized store", func() {
			BeforeEach(func() {
				var err error
				params.Lifecycle = ProtoLifecycle{}
				store, err = storeStructuredMongo.NewStore(params)
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns the appropriate status when uninitialized", func() {
				status := store.Status(context.Background())
				Expect(status).ToNot(BeNil())
				Expect(status.Ping).To(Equal("FAILED"))
			})

			It("returns a nil collection", func() {
				repository = store.GetRepository("")
				Expect(repository).To(BeNil())
			})

		})

		Context("with an initialized store", func() {
			BeforeEach(func() {
				var err error
				store, err = storeStructuredMongo.NewStore(params)
				Expect(err).ToNot(HaveOccurred())
				Expect(store).ToNot(BeNil())
			})

			Context("Status", func() {
				It("returns the appropriate status when initialized", func() {
					status := store.Status(context.Background())
					Expect(status).ToNot(BeNil())
					Expect(status.Ping).To(Equal("OK"))
				})
			})

			Context("GetRepository", func() {
				It("returns a new collection if no collection specified", func() {
					repository = store.GetRepository("")
					Expect(repository).ToNot(BeNil())
				})

				It("returns successfully", func() {
					repository = store.GetRepository("test")
					Expect(repository).ToNot(BeNil())
				})
			})

			Context("with a new collection", func() {
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

				Context("C", func() {
					It("returns successfully", func() {
						Expect(repository).ToNot(BeNil())
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
