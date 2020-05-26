package mongo_test

import (
	mgo "github.com/globalsign/mgo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	logTest "github.com/tidepool-org/platform/log/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	taskStore "github.com/tidepool-org/platform/task/store"
	taskStoreMongo "github.com/tidepool-org/platform/task/store/mongo"
)

var _ = Describe("Mongo", func() {
	var cfg *storeStructuredMongo.Config
	var str *taskStoreMongo.Store
	var ssn taskStore.TaskSession

	BeforeEach(func() {
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
		It("returns an error if unsuccessful", func() {
			var err error
			str, err = taskStoreMongo.NewStore(nil, nil)
			Expect(err).To(HaveOccurred())
			Expect(str).To(BeNil())
		})

		It("returns a new store and no error if successful", func() {
			var err error
			str, err = taskStoreMongo.NewStore(cfg, logTest.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		var mgoSession *mgo.Session
		var mgoCollection *mgo.Collection

		BeforeEach(func() {
			var err error
			str, err = taskStoreMongo.NewStore(cfg, logTest.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
			mgoSession = storeStructuredMongoTest.Session().Copy()
			mgoCollection = mgoSession.DB(cfg.Database).C(cfg.CollectionPrefix + "tasks")
		})

		AfterEach(func() {
			if mgoSession != nil {
				mgoSession.Close()
			}
		})

		Context("EnsureIndexes", func() {
			It("returns successfully", func() {
				Expect(str.EnsureIndexes()).To(Succeed())
				indexes, err := mgoCollection.Indexes()
				Expect(err).ToNot(HaveOccurred())
				Expect(indexes).To(ConsistOf(
					MatchFields(IgnoreExtras, Fields{"Key": ConsistOf("_id")}),
					MatchFields(IgnoreExtras, Fields{"Key": ConsistOf("id"), "Background": Equal(true), "Unique": Equal(true)}),
					MatchFields(IgnoreExtras, Fields{"Key": ConsistOf("name"), "Background": Equal(true), "Unique": Equal(true), "Sparse": Equal(true)}),
					MatchFields(IgnoreExtras, Fields{"Key": ConsistOf("priority"), "Background": Equal(true)}),
					MatchFields(IgnoreExtras, Fields{"Key": ConsistOf("availableTime"), "Background": Equal(true)}),
					MatchFields(IgnoreExtras, Fields{"Key": ConsistOf("expirationTime"), "Background": Equal(true)}),
					MatchFields(IgnoreExtras, Fields{"Key": ConsistOf("state"), "Background": Equal(true)}),
				))
			})
		})

		Context("NewTaskSession", func() {
			It("returns a new session", func() {
				ssn = str.NewTaskSession()
				Expect(ssn).ToNot(BeNil())
			})
		})
	})
})
