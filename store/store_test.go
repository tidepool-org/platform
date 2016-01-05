package store_test

import (
	"time"

	"github.com/tidepool-org/platform/config"
	. "github.com/tidepool-org/platform/store"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Store", func() {

	const (
		test_collection = "my_test_collection"
	)

	Context("When created", func() {

		var (
			mgoConfig MongoConfig
		)

		BeforeEach(func() {
			config.FromJson(&mgoConfig, "mongo.json")
		})

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
			Expect(mgo.Config.DbName).To(Equal(mgoConfig.DbName))
		})
	})

	Context("When used", func() {

		type SaveMe struct {
			Timestamp string
			UserId    string
			Id        string
			Stuff     []string
		}

		var (
			testStore *MongoStore
		)

		BeforeEach(func() {
			testStore = NewMongoStore(test_collection)
			testStore.Cleanup()
		})

		It("should be able to save", func() {
			saveMe := SaveMe{
				Timestamp: time.Now().UTC().String(),
				UserId:    "99",
				Id:        "one-12345-89-asfde",
				Stuff:     []string{"1", "2", "miss", "a", "few", "99", "100"},
			}
			Expect(testStore.Save(saveMe)).To(BeNil())
		})
		It("should be able to update", func() {
			saveMe := SaveMe{
				Timestamp: time.Now().UTC().String(),
				UserId:    "99",
				Id:        "one-12345-89-asfde",
				Stuff:     []string{"1", "2", "miss", "a", "few", "99", "100"},
			}

			Expect(testStore.Save(saveMe)).To(BeNil())
			var updated SaveMe
			updated = saveMe
			updated.Stuff = []string{"just", "1"}
			Expect(testStore.Update(StoreIdField{"id", updated.Id}, updated)).To(BeNil())

		})
		It("should be able to delete", func() {
			saveMe := SaveMe{
				Timestamp: time.Now().UTC().String(),
				Id:        "one-12345-89-asfde",
				UserId:    "99",
				Stuff:     []string{"1", "2", "miss", "a", "few", "99", "100"},
			}

			Expect(testStore.Save(saveMe)).To(BeNil())
			Expect(testStore.Delete(StoreIdField{"id", saveMe.Id})).To(BeNil())
		})
		It("should be able to get one", func() {
			saveMe := SaveMe{
				Timestamp: time.Now().UTC().String(),
				Id:        "one-12345-89-asfde",
				UserId:    "99",
				Stuff:     []string{"1", "2", "miss", "a", "few", "99", "100"},
			}

			var found SaveMe

			Expect(testStore.Save(saveMe)).To(BeNil())
			Expect(testStore.Read(StoreIdField{"id", saveMe.Id}, &found)).To(BeNil())
			Expect(found).To(Equal(saveMe))
		})
		It("should be able to get all", func() {

			one := SaveMe{
				Timestamp: time.Now().UTC().String(),
				Id:        "one-12345-89-asfde",
				UserId:    "99",
				Stuff:     []string{"1", "2", "miss", "a", "few", "99", "100"},
			}
			two := SaveMe{
				Timestamp: time.Now().UTC().String(),
				Id:        "two-9876-54-asfde",
				UserId:    "99",
				Stuff:     []string{"100", "99", "miss", "a", "few", "2", "1"},
			}
			var found []SaveMe
			Expect(testStore.Save(one)).To(BeNil())
			Expect(testStore.Save(two)).To(BeNil())
			Expect(testStore.ReadAll(StoreIdField{"userid", one.UserId}, &found)).To(BeNil())
			Expect(len(found)).To(Equal(2))
			Expect(found[0]).To(Equal(one))
			Expect(found[1]).To(Equal(two))
		})
	})
})
