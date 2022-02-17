package mongo_test

import (
	"os"
	"time"

	mgo "github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
)

var _ = Describe("Mongo", func() {
	Context("Store", func() {
		var logger log.Logger
		var config *storeStructuredMongo.Config
		var store *storeStructuredMongo.Store
		var session *storeStructuredMongo.Session

		BeforeEach(func() {
			logger = logTest.NewLogger()
			config = storeStructuredMongoTest.NewConfig()
		})

		AfterEach(func() {
			if session != nil {
				session.Close()
			}
			if store != nil {
				store.Close()
			}
		})

		Context("NewStore", func() {
			It("returns an error if the config is missing", func() {
				var err error
				store, err = storeStructuredMongo.NewStore(nil, logger)
				Expect(err).To(MatchError("config is missing"))
				Expect(store).To(BeNil())
			})

			It("returns an error if the config is invalid", func() {
				config.SetAddresses(nil)
				var err error
				store, err = storeStructuredMongo.NewStore(config, logger)
				Expect(err).To(MatchError("config is invalid; addresses is missing"))
				Expect(store).To(BeNil())
			})

			It("returns an error if the logger is missing", func() {
				var err error
				store, err = storeStructuredMongo.NewStore(config, nil)
				Expect(err).To(MatchError("logger is missing"))
				Expect(store).To(BeNil())
			})

			It("returns no error if the server is not reachable and initialize session once it is", func() {
				config.SetAddresses([]string{"127.0.0.0"})
				config.WaitConnectionInterval = 1 * time.Second
				config.Timeout = 2 * time.Second
				var err error

				store, err = storeStructuredMongo.NewStore(config, logger)
				Expect(err).To(BeNil())
				Expect(store).ToNot(BeNil())
				Expect(store.Session()).To(BeNil())
				time.Sleep(3 * time.Second)
				Expect(store.Session()).To(BeNil())
				mongoAddress := "127.0.0.1:27017"
				if os.Getenv("TIDEPOOL_STORE_ADDRESSES") != "" {
					mongoAddress = os.Getenv("TIDEPOOL_STORE_ADDRESSES")
				}
				config.SetAddresses([]string{mongoAddress})
				store.WaitUntilStarted()
				Expect(store.Session()).ToNot(BeNil())
			})

			It("returns no error if successful", func() {
				var err error
				store, err = storeStructuredMongo.NewStore(config, logger)
				store.WaitUntilStarted()
				Expect(err).ToNot(HaveOccurred())
				Expect(store).ToNot(BeNil())
			})
		})

		Context("with a new store", func() {
			BeforeEach(func() {
				var err error
				store, err = storeStructuredMongo.NewStore(config, logger)
				store.WaitUntilStarted()
				Expect(err).ToNot(HaveOccurred())
				Expect(store).ToNot(BeNil())
			})

			Context("IsClosed/Close", func() {
				It("returns false if it is not closed", func() {
					Expect(store.IsClosed()).To(BeFalse())
				})

				It("returns true if it is closed", func() {
					store.Close()
					Expect(store.IsClosed()).To(BeTrue())
				})
			})

			Context("Status", func() {
				It("returns the appropriate status when not closed", func() {
					status := store.Status()
					Expect(status).ToNot(BeNil())
					mongoStatus, ok := status.(*storeStructuredMongo.Status)
					Expect(ok).To(BeTrue())
					Expect(mongoStatus).ToNot(BeNil())
					Expect(mongoStatus.State).To(Equal("OPEN"))
					Expect(mongoStatus.BuildInfo).ToNot(BeNil())
					Expect(mongoStatus.LiveServers).ToNot(BeEmpty())
					Expect(mongoStatus.Mode).To(Equal(mgo.Strong))
					Expect(mongoStatus.Safe).ToNot(BeNil())
					Expect(mongoStatus.Ping).To(Equal("OK"))
				})

				It("returns the appropriate status when closed", func() {
					store.Close()
					Expect(store.IsClosed()).To(BeTrue())
					status := store.Status()
					Expect(status).ToNot(BeNil())
					mongoStatus, ok := status.(*storeStructuredMongo.Status)
					Expect(ok).To(BeTrue())
					Expect(mongoStatus).ToNot(BeNil())
					Expect(mongoStatus.State).To(Equal("CLOSED"))
					Expect(mongoStatus.BuildInfo).To(BeNil())
					Expect(mongoStatus.LiveServers).To(BeEmpty())
					Expect(mongoStatus.Mode).To(Equal(mgo.Eventual))
					Expect(mongoStatus.Safe).To(BeNil())
					Expect(mongoStatus.Ping).To(Equal("FAILED"))
				})
			})

			Context("NewSession", func() {
				It("returns a new session if no collection specified", func() {
					session = store.NewSession("")
					Expect(session).ToNot(BeNil())
				})

				It("returns successfully", func() {
					session = store.NewSession("test")
					Expect(session).ToNot(BeNil())
				})
			})

			Context("with a new session", func() {
				BeforeEach(func() {
					session = store.NewSession("test")
					Expect(session).ToNot(BeNil())
				})

				Context("IsClosed/Close", func() {
					It("returns false if it is not closed", func() {
						Expect(session.IsClosed()).To(BeFalse())
					})

					It("returns true if it is closed", func() {
						session.Close()
						Expect(session.IsClosed()).To(BeTrue())
					})
				})

				Context("EnsureAllIndexes", func() {
					It("returns an error if the index is invalid", func() {
						Expect(session.EnsureAllIndexes([]mgo.Index{{}})).To(MatchError("unable to ensure index with key []; invalid index key: no fields provided"))
					})

					It("returns successfully with nil indexes", func() {
						Expect(session.EnsureAllIndexes(nil)).To(Succeed())
					})

					It("returns successfully with empty indexes", func() {
						Expect(session.EnsureAllIndexes([]mgo.Index{})).To(Succeed())
					})

					It("returns successfully with multiple indexes", func() {
						Expect(session.EnsureAllIndexes([]mgo.Index{
							{Key: []string{"one"}, Unique: true, Background: true},
							{Key: []string{"two"}, Background: true},
							{Key: []string{"three"}},
						})).To(Succeed())
					})
				})

				Context("C", func() {
					It("returns successfully", func() {
						Expect(session.C()).ToNot(BeNil())
					})

					It("returns nil if the session is closed", func() {
						session.Close()
						Expect(session.C()).To(BeNil())
					})
				})

				Context("ArchiveC", func() {
					It("returns successfully", func() {
						Expect(session.ArchiveC()).ToNot(BeNil())
					})

					It("returns nil if the session is closed", func() {
						session.Close()
						Expect(session.ArchiveC()).To(BeNil())
					})
				})

				DescribeTable("ConstructUpdate",
					func(set bson.M, unset bson.M, operators []map[string]bson.M, expected bson.M) {
						Expect(session.ConstructUpdate(set, unset, operators...)).To(Equal(expected))
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
