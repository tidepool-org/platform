package store_test

import (
	"os"
	"time"

	. "github.com/tidepool-org/platform/store"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Store", func() {

	const (
		test_collection  = "my_test_collection"
		test_mgo_url     = "mongodb://localhost/store_test"
		test_mgo_timeout = "5"
		test_mgo_db_name = "store_test"
	)

	BeforeSuite(func() {
		os.Setenv(MONGO_STORE_URL, test_mgo_url)
		os.Setenv(MONGO_STORE_TIMEOUT, test_mgo_timeout)
		os.Setenv(MONGO_STORE_DB_NAME, test_mgo_db_name)
	})

	Context("When created", func() {
		It("should be assingable to the interface", func() {
			var testStore Store
			testStore = NewMongoStore(test_collection)
			Expect(testStore).To(Not(BeNil()))
		})
		It("should set the collection name", func() {
			mgo := NewMongoStore(test_collection)
			Expect(mgo.CollectionName).To(Equal(test_collection))
		})
		It("should set the db name", func() {
			mgo := NewMongoStore(test_collection)
			Expect(mgo.DbName).To(Equal(test_mgo_db_name))
		})
	})

	Context("When used", func() {

		type SaveMe struct {
			Date  time.Time
			Id    string
			Stuff []string
		}

		var (
			testStore MongoStore

			saveMeOne = SaveMe{
				Date:  time.Now(),
				Id:    "one-12345-89-asfde",
				Stuff: []string{"1", "2", "miss", "a", "few", "99", "100"},
			}

			saveMeTwo = SaveMe{
				Date:  time.Now(),
				Id:    "two-12345-89-asfde",
				Stuff: []string{"100", "99", "miss", "a", "few", "2", "1"},
			}
		)

		BeforeEach(func() {
			testStore = NewMongoStore(test_collection)
			if err := testStore.CleanUp(); err != nil {
				Fail("Failed mongo store test setup", err.Error)
			}
		})

		It("should be able to save", func() {
			Expect(testStore.Save(saveMeOne)).To(BeNil())
		})
		It("should be able to update", func() {
			Expect(testStore.Save(saveMeOne)).To(BeNil())
			updated = copy(saveMeOne)
			updated.Stuff = []string{"just", "1"}
			Expect(testStore.Update(updated.Id, updated)).To(BeNil())

		})
		It("should be able to delete", func() {
			Expect(testStore.Save(saveMeOne)).To(BeNil())
			Expect(testStore.Delete(saveMeOne.Id)).To(BeNil())
		})
		It("should be able to get one", func() {
			var found SaveMe
			Expect(testStore.Save(saveMeOne)).To(BeNil())
			Expect(testStore.Read(saveMeOne.Id, found)).To(BeNil())
			Expect(found).To(Equal(saveMeOne))
		})
		It("should be able to get all", func() {
			var found []SaveMe
			Expect(testStore.Save(saveMeOne)).To(BeNil())
			Expect(testStore.Save(saveMeTwo)).To(BeNil())
			Expect(testStore.ReadAll(saveMeOne.Id, found)).To(BeNil())
			Expect(len(found)).To(Equal(2))
			Expect(found[0]).To(Equal(saveMeOne))
			Expect(found[1]).To(Equal(saveMeTwo))
		})
	})
})
