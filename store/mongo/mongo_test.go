package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	mgo "gopkg.in/mgo.v2"

	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/store/mongo"
	testMongo "github.com/tidepool-org/platform/test/mongo"
)

type TestAgent struct {
	TestIsServer bool
	TestUserID   string
}

func (t *TestAgent) IsServer() bool {
	return t.TestIsServer
}

func (t *TestAgent) UserID() string {
	return t.TestUserID
}

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
			mongoConfig.Addresses = nil
			var err error
			mongoStore, err = mongo.New(logger, mongoConfig)
			Expect(err).To(MatchError("mongo: config is invalid; mongo: addresses is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if the addresses are not reachable", func() {
			mongoConfig.Addresses = []string{"127.0.0.0", "127.0.0.0"}
			var err error
			mongoStore, err = mongo.New(logger, mongoConfig)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(HavePrefix("mongo: unable to dial database; "))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if the username or password is invalid", func() {
			mongoConfig.Username = pointer.String("username")
			mongoConfig.Password = pointer.String("password")
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
			It("returns a new session if no logger specified", func() {
				mongoSession = mongoStore.NewSession(nil, "test")
				Expect(mongoSession).ToNot(BeNil())
				Expect(mongoSession.Logger()).ToNot(BeNil())
			})

			It("returns a new session if no collection specified", func() {
				mongoSession = mongoStore.NewSession(logger, "")
				Expect(mongoSession).ToNot(BeNil())
				Expect(mongoSession.Logger()).ToNot(BeNil())
			})

			It("returns successfully", func() {
				mongoSession = mongoStore.NewSession(logger, "test")
				Expect(mongoSession).ToNot(BeNil())
				Expect(mongoSession.Logger()).ToNot(BeNil())
			})
		})

		Context("with a new session", func() {
			BeforeEach(func() {
				mongoSession = mongoStore.NewSession(null.NewLogger(), "test")
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

			Context("SetAgent", func() {
				It("successfully sets the agent", func() {
					mongoSession.SetAgent(&TestAgent{false, id.New()})
				})

				It("successfully sets the agent if nil", func() {
					mongoSession.SetAgent(nil)
				})
			})

			Context("AgentUserID", func() {
				It("returns an empty string if the agent is not set", func() {
					Expect(mongoSession.AgentUserID()).To(BeEmpty())
				})

				It("returns an empty string if the agent is nil", func() {
					mongoSession.SetAgent(nil)
					Expect(mongoSession.AgentUserID()).To(BeEmpty())
				})

				It("returns an empty string if the agent is server", func() {
					mongoSession.SetAgent(&TestAgent{true, id.New()})
					Expect(mongoSession.AgentUserID()).To(BeEmpty())
				})

				It("returns the agent user id if the agent is set", func() {
					agentUserID := id.New()
					mongoSession.SetAgent(&TestAgent{false, agentUserID})
					Expect(mongoSession.AgentUserID()).To(Equal(agentUserID))
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
			Context("Timestamp", func() {
				It("returns a new timestamp in RFC3339 format", func() {
					parsedTimestamp, err := time.Parse(time.RFC3339, mongoSession.Timestamp())
					Expect(err).ToNot(HaveOccurred())
					Expect(parsedTimestamp).ToNot(BeNil())
				})
			})
		})
	})
})
