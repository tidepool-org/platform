package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	mgo "gopkg.in/mgo.v2"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store/mongo"
	testMongo "github.com/tidepool-org/platform/test/mongo"
)

var _ = Describe("Mongo", func() {
	var logger log.Logger
	var mongoConfig *mongo.Config
	var mongoStore *mongo.Store
	var mongoSession *mongo.Session

	BeforeEach(func() {
		logger = log.NewNull()
		mongoConfig = &mongo.Config{
			Addresses:  testMongo.Address(),
			Database:   testMongo.Database(),
			Collection: testMongo.NewCollectionName(),
			Timeout:    app.DurationAsPointer(5 * time.Second),
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
		It("returns no error if successful", func() {
			var err error
			mongoStore, err = mongo.New(logger, mongoConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})

		It("returns an error if the logger is missing", func() {
			var err error
			mongoStore, err = mongo.New(nil, mongoConfig)
			Expect(err).To(MatchError("mongo: logger is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if the config is missing", func() {
			var err error
			mongoStore, err = mongo.New(logger, nil)
			Expect(err).To(MatchError("mongo: config is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if the config is invalid", func() {
			mongoConfig.Addresses = ""
			var err error
			mongoStore, err = mongo.New(logger, mongoConfig)
			Expect(err).To(MatchError("mongo: config is invalid; mongo: addresses is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if the addresses are not reachable", func() {
			mongoConfig.Addresses = "127.0.0.0, 127.0.0.0"
			var err error
			mongoStore, err = mongo.New(logger, mongoConfig)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(HavePrefix("mongo: unable to dial database; "))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if the username or password is invalid", func() {
			mongoConfig.Username = app.StringAsPointer("username")
			mongoConfig.Password = app.StringAsPointer("password")
			var err error
			mongoStore, err = mongo.New(logger, mongoConfig)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(HavePrefix("mongo: unable to dial database; "))
			Expect(mongoStore).To(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			mongoStore, err = mongo.New(logger, mongoConfig)
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

		Context("GetStatus", func() {
			It("returns the appropriate status when not closed", func() {
				status := mongoStore.GetStatus()
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
				status := mongoStore.GetStatus()
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
			It("returns no error if successful", func() {
				var err error
				mongoSession, err = mongoStore.NewSession(log.NewNull())
				Expect(err).ToNot(HaveOccurred())
				Expect(mongoSession).ToNot(BeNil())
			})

			It("returns an error if the logger is missing", func() {
				var err error
				mongoSession, err = mongoStore.NewSession(nil)
				Expect(err).To(MatchError("mongo: logger is missing"))
				Expect(mongoSession).To(BeNil())
			})

			It("returns an error if the store is closed", func() {
				mongoStore.Close()
				Expect(mongoStore.IsClosed()).To(BeTrue())
				var err error
				mongoSession, err = mongoStore.NewSession(log.NewNull())
				Expect(err).To(MatchError("mongo: store closed"))
				Expect(mongoSession).To(BeNil())
			})
		})

		Context("with a new session", func() {
			BeforeEach(func() {
				var err error
				mongoSession, err = mongoStore.NewSession(log.NewNull())
				Expect(err).ToNot(HaveOccurred())
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

			Context("Logger", func() {
				It("returns successfully", func() {
					Expect(mongoSession.Logger()).ToNot(BeNil())
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
		})
	})
})
