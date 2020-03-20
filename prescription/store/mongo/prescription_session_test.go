package mongo_test

import (
	"context"

	"github.com/globalsign/mgo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/prescription"
	prescriptionStore "github.com/tidepool-org/platform/prescription/store"
	prescriptionStoreMongo "github.com/tidepool-org/platform/prescription/store/mongo"
	"github.com/tidepool-org/platform/prescription/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
)

var _ = Describe("PrescriptionSession", func() {
	var config *storeStructuredMongo.Config
	var logger *logTest.Logger
	var store *prescriptionStoreMongo.Store
	var session prescriptionStore.PrescriptionSession

	BeforeEach(func() {
		config = storeStructuredMongoTest.NewConfig()
		logger = logTest.NewLogger()
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
		It("returns an error when unsuccessful", func() {
			var err error
			store, err = prescriptionStoreMongo.NewStore(nil, logger)
			errorsTest.ExpectEqual(err, errors.New("config is missing"))
			Expect(store).To(BeNil())
		})

		It("returns a new store and no error when successful", func() {
			var err error
			store, err = prescriptionStoreMongo.NewStore(config, logger)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		var mgoSession *mgo.Session
		var mgoCollection *mgo.Collection

		BeforeEach(func() {
			var err error
			store, err = prescriptionStoreMongo.NewStore(config, logger)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
			mgoSession = storeStructuredMongoTest.Session().Copy()
			mgoCollection = mgoSession.DB(config.Database).C(config.CollectionPrefix + "prescriptions")
		})

		AfterEach(func() {
			if mgoSession != nil {
				mgoSession.Close()
			}
		})

		Context("EnsureIndexes", func() {
			It("returns successfully", func() {
				Expect(store.EnsureIndexes()).To(Succeed())
				indexes, err := mgoCollection.Indexes()
				Expect(err).ToNot(HaveOccurred())
				Expect(indexes).To(ContainElement(
					MatchFields(IgnoreExtras, Fields{"Key": ConsistOf("id")}),
				))
			})
		})

		Context("NewSession", func() {
			It("returns a new session", func() {
				session = store.NewPrescriptionSession()
				Expect(session).ToNot(BeNil())
			})
		})

		Context("with a new session", func() {
			var ctx context.Context

			BeforeEach(func() {
				Expect(store.EnsureIndexes()).To(Succeed())
				session = store.NewPrescriptionSession()
				ctx = log.NewContextWithLogger(context.Background(), logger)
			})

			Context("CreatePrescription", func() {
				var userID = ""
				var revisionCreate *prescription.RevisionCreate = nil

				BeforeEach(func() {
					userID = authTest.RandomUserID()
					revisionCreate = test.RandomRevisionCreate()
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					result, err := session.CreatePrescription(ctx, userID, revisionCreate)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the userID is missing", func() {
					userID = ""
					result, err := session.CreatePrescription(ctx, userID, revisionCreate)
					errorsTest.ExpectEqual(err, errors.New("userID is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the session is closed", func() {
					session.Close()
					result, err := session.CreatePrescription(ctx, userID, revisionCreate)
					errorsTest.ExpectEqual(err, errors.New("session closed"))
					Expect(result).To(BeNil())
				})

				It("returns the created prescription on success", func() {
					result, err := session.CreatePrescription(ctx, userID, revisionCreate)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).ToNot(BeNil())
				})
			})
		})
	})
})
