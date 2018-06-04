package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
)

var _ = Describe("Mongo", func() {
	var lgr log.Logger
	var cfg *storeStructuredMongo.Config
	var str *storeStructuredMongo.Store
	var ssn *storeStructuredMongo.Session

	BeforeEach(func() {
		lgr = logNull.NewLogger()
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

	Context("New", func() {
		It("returns an error if the config is missing", func() {
			var err error
			str, err = storeStructuredMongo.NewStore(nil, lgr)
			Expect(err).To(MatchError("config is missing"))
			Expect(str).To(BeNil())
		})

		It("returns an error if the config is invalid", func() {
			cfg.Addresses = nil
			var err error
			str, err = storeStructuredMongo.NewStore(cfg, lgr)
			Expect(err).To(MatchError("config is invalid; addresses is missing"))
			Expect(str).To(BeNil())
		})

		It("returns an error if the logger is missing", func() {
			var err error
			str, err = storeStructuredMongo.NewStore(cfg, nil)
			Expect(err).To(MatchError("logger is missing"))
			Expect(str).To(BeNil())
		})

		It("returns an error if the addresses are not reachable", func() {
			cfg.Addresses = []string{"127.0.0.0", "127.0.0.0"}
			var err error
			str, err = storeStructuredMongo.NewStore(cfg, lgr)
			Expect(err).To(MatchError("unable to dial database; no reachable servers"))
			Expect(str).To(BeNil())
		})

		It("returns an error if the username or password is invalid", func() {
			cfg.Username = pointer.FromString("username")
			cfg.Password = pointer.FromString("password")
			var err error
			str, err = storeStructuredMongo.NewStore(cfg, lgr)
			Expect(err).To(MatchError("unable to dial database; server returned error on SASL authentication step: Authentication failed."))
			Expect(str).To(BeNil())
		})

		It("returns an error if TLS is specified on a server that does not support it", func() {
			cfg.TLS = true
			var err error
			str, err = storeStructuredMongo.NewStore(cfg, lgr)
			Expect(err).To(MatchError("unable to dial database; no reachable servers"))
			Expect(str).To(BeNil())
		})

		It("returns no error if successful", func() {
			var err error
			str, err = storeStructuredMongo.NewStore(cfg, lgr)
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			str, err = storeStructuredMongo.NewStore(cfg, lgr)
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
		})

		Context("IsClosed/Close", func() {
			It("returns false if it is not closed", func() {
				Expect(str.IsClosed()).To(BeFalse())
			})

			It("returns true if it is closed", func() {
				str.Close()
				Expect(str.IsClosed()).To(BeTrue())
			})
		})

		Context("Status", func() {
			It("returns the appropriate status when not closed", func() {
				status := str.Status()
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
				str.Close()
				Expect(str.IsClosed()).To(BeTrue())
				status := str.Status()
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
				ssn = str.NewSession("")
				Expect(ssn).ToNot(BeNil())
			})

			It("returns successfully", func() {
				ssn = str.NewSession("test")
				Expect(ssn).ToNot(BeNil())
			})
		})

		Context("with a new session", func() {
			BeforeEach(func() {
				ssn = str.NewSession("test")
				Expect(ssn).ToNot(BeNil())
			})

			Context("IsClosed/Close", func() {
				It("returns false if it is not closed", func() {
					Expect(ssn.IsClosed()).To(BeFalse())
				})

				It("returns true if it is closed", func() {
					ssn.Close()
					Expect(ssn.IsClosed()).To(BeTrue())
				})
			})

			Context("EnsureAllIndexes", func() {
				It("returns an error if the index is invalid", func() {
					Expect(ssn.EnsureAllIndexes([]mgo.Index{{}})).To(MatchError("unable to ensure index with key []; invalid index key: no fields provided"))
				})

				It("returns successfully with nil indexes", func() {
					Expect(ssn.EnsureAllIndexes(nil)).To(Succeed())
				})

				It("returns successfully with empty indexes", func() {
					Expect(ssn.EnsureAllIndexes([]mgo.Index{})).To(Succeed())
				})

				It("returns successfully with multiple indexes", func() {
					Expect(ssn.EnsureAllIndexes([]mgo.Index{
						{Key: []string{"one"}, Unique: true, Background: true},
						{Key: []string{"two"}, Background: true},
						{Key: []string{"three"}},
					})).To(Succeed())
				})
			})

			Context("C", func() {
				It("returns successfully", func() {
					Expect(ssn.C()).ToNot(BeNil())
				})

				It("returns nil if the session is closed", func() {
					ssn.Close()
					Expect(ssn.C()).To(BeNil())
				})
			})

			DescribeTable("ConstructUpdate",
				func(set bson.M, unset bson.M, expected bson.M) {
					Expect(ssn.ConstructUpdate(set, unset)).To(Equal(expected))
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
