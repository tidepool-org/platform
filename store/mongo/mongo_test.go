package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/store/mongo"
	testMongo "github.com/tidepool-org/platform/test/mongo"
)

var _ = Describe("Mongo", func() {
	var logger log.Logger
	var mongoConfig *mongo.Config
	var mongoStore *mongo.Store
	var mongoSession *mongo.Session

	BeforeEach(func() {
		logger = null.NewLogger()
		mongoConfig = &mongo.Config{
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
		It("returns an error if the config is missing", func() {
			var err error
			mongoStore, err = mongo.NewStore(nil, logger)
			Expect(err).To(MatchError("config is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if the logger is missing", func() {
			var err error
			mongoStore, err = mongo.NewStore(mongoConfig, nil)
			Expect(err).To(MatchError("logger is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if the config is invalid", func() {
			mongoConfig.Addresses = nil
			var err error
			mongoStore, err = mongo.NewStore(mongoConfig, logger)
			Expect(err).To(MatchError("config is invalid; addresses is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if the addresses are not reachable", func() {
			mongoConfig.Addresses = []string{"127.0.0.0", "127.0.0.0"}
			var err error
			mongoStore, err = mongo.NewStore(mongoConfig, logger)
			Expect(err).To(MatchError("unable to dial database; no reachable servers"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if the username or password is invalid", func() {
			mongoConfig.Username = pointer.FromString("username")
			mongoConfig.Password = pointer.FromString("password")
			var err error
			mongoStore, err = mongo.NewStore(mongoConfig, logger)
			Expect(err).To(MatchError("unable to dial database; server returned error on SASL authentication step: Authentication failed."))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if TLS is specified on a server that does not support it", func() {
			mongoConfig.TLS = true
			var err error
			mongoStore, err = mongo.NewStore(mongoConfig, logger)
			Expect(err).To(MatchError("unable to dial database; no reachable servers"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns no error if successful", func() {
			var err error
			mongoStore, err = mongo.NewStore(mongoConfig, logger)
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			mongoStore, err = mongo.NewStore(mongoConfig, logger)
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})

		Context("IsClosed/Close", func() {
			It("returns false if it is not closed", func() {
				Expect(mongoStore.IsClosed()).To(BeFalse())
			})

			It("returns true if it is closed", func() {
				mongoStore.Close()
				Expect(mongoStore.IsClosed()).To(BeTrue())
			})
		})

		Context("Status", func() {
			It("returns the appropriate status when not closed", func() {
				status := mongoStore.Status()
				Expect(status).ToNot(BeNil())
				mongoStatus, ok := status.(*mongo.Status)
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
				mongoStore.Close()
				Expect(mongoStore.IsClosed()).To(BeTrue())
				status := mongoStore.Status()
				Expect(status).ToNot(BeNil())
				mongoStatus, ok := status.(*mongo.Status)
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
				mongoSession = mongoStore.NewSession("")
				Expect(mongoSession).ToNot(BeNil())
			})

			It("returns successfully", func() {
				mongoSession = mongoStore.NewSession("test")
				Expect(mongoSession).ToNot(BeNil())
			})
		})

		Context("with a new session", func() {
			BeforeEach(func() {
				mongoSession = mongoStore.NewSession("test")
				Expect(mongoSession).ToNot(BeNil())
			})

			Context("IsClosed/Close", func() {
				It("returns false if it is not closed", func() {
					Expect(mongoSession.IsClosed()).To(BeFalse())
				})

				It("returns true if it is closed", func() {
					mongoSession.Close()
					Expect(mongoSession.IsClosed()).To(BeTrue())
				})
			})

			Context("EnsureIndexes", func() {
				It("returns successfully", func() {
					Expect(mongoSession.EnsureIndexes()).To(Succeed())
				})
			})

			Context("EnsureAllIndexes", func() {
				It("returns an error if the index is invalid", func() {
					Expect(mongoSession.EnsureAllIndexes([]mgo.Index{{}})).To(MatchError("unable to ensure index with key []; invalid index key: no fields provided"))
				})

				It("returns successfully with nil indexes", func() {
					Expect(mongoSession.EnsureAllIndexes(nil)).To(Succeed())
				})

				It("returns successfully with empty indexes", func() {
					Expect(mongoSession.EnsureAllIndexes([]mgo.Index{})).To(Succeed())
				})

				It("returns successfully with multiple indexes", func() {
					Expect(mongoSession.EnsureAllIndexes([]mgo.Index{
						{Key: []string{"one"}, Unique: true, Background: true},
						{Key: []string{"two"}, Background: true},
						{Key: []string{"three"}},
					})).To(Succeed())
				})
			})

			Context("C", func() {
				It("returns successfully", func() {
					Expect(mongoSession.C()).ToNot(BeNil())
				})

				It("returns nil if the session is closed", func() {
					mongoSession.Close()
					Expect(mongoSession.C()).To(BeNil())
				})
			})

			DescribeTable("ConstructUpdate",
				func(set bson.M, unset bson.M, expected bson.M) {
					Expect(mongoSession.ConstructUpdate(set, unset)).To(Equal(expected))
				},
				Entry("where set is nil and unset is nil", nil, nil, nil),
				Entry("where set is empty and unset is nil", bson.M{}, nil, nil),
				Entry("where set is nil and unset is empty", nil, bson.M{}, nil),
				Entry("where set is empty and unset is empty", bson.M{}, bson.M{}, nil),
				Entry("where set is present and unset is nil", bson.M{"one": "alpha", "two": true}, nil, bson.M{"$set": bson.M{"one": "alpha", "two": true}}),
				Entry("where set is present and unset is empty", bson.M{"one": "alpha", "two": true}, bson.M{}, bson.M{"$set": bson.M{"one": "alpha", "two": true}}),
				Entry("where set is nil and unset is present", nil, bson.M{"three": "charlie", "four": false}, bson.M{"$unset": bson.M{"three": "charlie", "four": false}}),
				Entry("where set is empty and unset is present", bson.M{}, bson.M{"three": "charlie", "four": false}, bson.M{"$unset": bson.M{"three": "charlie", "four": false}}),
				Entry("where set is empty and unset is present", bson.M{"one": "alpha", "two": true}, bson.M{"three": "charlie", "four": false}, bson.M{"$set": bson.M{"one": "alpha", "two": true}, "$unset": bson.M{"three": "charlie", "four": false}}),
			)
		})
	})
})
