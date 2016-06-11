package store_test

import (
	. "github.com/onsi/ginkgo"
	// . "github.com/onsi/gomega"
)

var _ = PDescribe("Store", func() {

	// const (
	// 	testCollection = "my_test_collection"
	// )

	// Context("When created", func() {

	// 	var (
	// 		mongoConfig store.MongoConfig
	// 	)

	// 	BeforeEach(func() {
	// 		config.FromJSON(&mongoConfig, "mongo.json")
	// 	})

	// 	It("should be assingable to the interface", func() {
	// 		var testStore store.Store
	// 		testStore = store.NewMongoStore(testCollection)
	// 		Expect(testStore).To(Not(BeNil()))
	// 	})
	// 	It("should set the collection name", func() {
	// 		mgo := store.NewMongoStore(testCollection)
	// 		Expect(mgo.CollectionName).To(Equal(testCollection))
	// 	})
	// })

	// Context("When used", func() {

	// 	type SaveMe struct {
	// 		Timestamp string
	// 		UserID    string
	// 		ID        string
	// 		Stuff     []string
	// 	}

	// 	var (
	// 		testStore *store.MongoStore
	// 	)

	// 	BeforeEach(func() {
	// 		testStore = store.NewMongoStore(testCollection)
	// 		testStore.Cleanup()
	// 	})

	// 	It("should be able to save", func() {
	// 		saveMe := SaveMe{
	// 			Timestamp: time.Now().UTC().String(),
	// 			UserID:    "99",
	// 			ID:        "one-12345-89-asfde",
	// 			Stuff:     []string{"1", "2", "miss", "a", "few", "99", "100"},
	// 		}
	// 		Expect(testStore.Save(saveMe)).To(BeNil())
	// 	})
	// 	It("should be able to update", func() {
	// 		saveMe := SaveMe{
	// 			Timestamp: time.Now().UTC().String(),
	// 			UserID:    "99",
	// 			ID:        "one-12345-89-asfde",
	// 			Stuff:     []string{"1", "2", "miss", "a", "few", "99", "100"},
	// 		}

	// 		Expect(testStore.Save(saveMe)).To(BeNil())
	// 		var updated SaveMe
	// 		updated = saveMe
	// 		updated.Stuff = []string{"just", "1"}

	// 		selector := map[string]interface{}{"id": updated.ID}

	// 		Expect(testStore.Update(selector, updated)).To(BeNil())

	// 	})
	// 	It("should be able to delete", func() {
	// 		saveMe := SaveMe{
	// 			Timestamp: time.Now().UTC().String(),
	// 			ID:        "one-12345-89-asfde",
	// 			UserID:    "99",
	// 			Stuff:     []string{"1", "2", "miss", "a", "few", "99", "100"},
	// 		}

	// 		Expect(testStore.Save(saveMe)).To(BeNil())
	// 		Expect(testStore.Delete(store.Field{Name: "id", Value: saveMe.ID})).To(BeNil())
	// 	})
	// 	It("should be able to get one", func() {
	// 		saveMe := SaveMe{
	// 			Timestamp: time.Now().UTC().String(),
	// 			ID:        "one-12345-89-asfde",
	// 			UserID:    "99",
	// 			Stuff:     []string{"1", "2", "miss", "a", "few", "99", "100"},
	// 		}

	// 		var found SaveMe

	// 		Expect(testStore.Save(saveMe)).To(BeNil())
	// 		Expect(testStore.Read(store.Field{Name: "id", Value: saveMe.ID}, store.Filter{}, &found)).To(BeNil())
	// 		Expect(found).To(Equal(saveMe))
	// 	})
	// 	It("should be able to get all", func() {

	// 		one := SaveMe{
	// 			Timestamp: time.Now().UTC().String(),
	// 			ID:        "one-12345-89-asfde",
	// 			UserID:    "99",
	// 			Stuff:     []string{"1", "2", "miss", "a", "few", "99", "100"},
	// 		}
	// 		two := SaveMe{
	// 			Timestamp: time.Now().UTC().String(),
	// 			ID:        "two-9876-54-asfde",
	// 			UserID:    "99",
	// 			Stuff:     []string{"100", "99", "miss", "a", "few", "2", "1"},
	// 		}
	// 		var found []SaveMe

	// 		process := func(iter store.Iterator) []SaveMe {
	// 			var chunk SaveMe
	// 			var all = []SaveMe{}

	// 			for iter.Next(&chunk) {
	// 				all = append(all, chunk)
	// 			}
	// 			return all
	// 		}

	// 		Expect(testStore.Save(one)).To(BeNil())
	// 		Expect(testStore.Save(two)).To(BeNil())

	// 		iter := testStore.ReadAll(store.Field{Name: "userid", Value: one.UserID}, store.Query{}, []string{}, store.Filter{})
	// 		found = process(iter)

	// 		Expect(len(found)).To(Equal(2))
	// 		Expect(found[0]).To(Equal(one))
	// 		Expect(found[1]).To(Equal(two))
	// 	})
	// })
})
