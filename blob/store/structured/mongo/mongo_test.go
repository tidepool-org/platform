package mongo_test

import (
	"context"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	blobStoreStructuredMongo "github.com/tidepool-org/platform/blob/store/structured/mongo"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
)

var _ = Describe("Mongo", func() {
	Context("NewStore", func() {
		var createStore *blobStoreStructuredMongo.Store

		AfterEach(func() {
			if createStore != nil {
				createStore.Terminate(context.Background())
			}
		})

		It("returns an error when unsuccessful", func() {
			createStore, err := blobStoreStructuredMongo.NewStore(nil)
			errorsTest.ExpectEqual(err, errors.New("database config is empty"))
			Expect(createStore).To(BeNil())
		})

		It("returns a new store and no error when successful", func() {
			config := storeStructuredMongoTest.NewConfig()
			createStore, err := blobStoreStructuredMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(createStore).ToNot(BeNil())
		})
	})

	Context("EnsureIndexes", func() {
		var store *blobStoreStructuredMongo.Store
		var deviceLogsCollection *mongo.Collection
		var blobsCollection *mongo.Collection

		BeforeEach(func() {
			store = GetSuiteStore()
			deviceLogsCollection = store.GetCollection("deviceLogs")
			blobsCollection = store.GetCollection("blobs")
		})

		AfterEach(func() {
			ctx := context.Background()
			if deviceLogsCollection != nil {
				deviceLogsCollection.DeleteMany(ctx, bson.D{})
			}
			if blobsCollection != nil {
				blobsCollection.DeleteMany(ctx, bson.D{})
			}
		})

		It("deviceLogs returns successfully", func() {
			//Expect(store.EnsureIndexes()).To(Succeed())
			cursor, err := deviceLogsCollection.Indexes().List(context.Background())
			Expect(err).ToNot(HaveOccurred())
			Expect(cursor).ToNot(BeNil())
			var indexes []storeStructuredMongoTest.MongoIndex
			err = cursor.All(context.Background(), &indexes)
			Expect(err).ToNot(HaveOccurred())

			Expect(indexes).To(ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					"Key": Equal(storeStructuredMongoTest.MakeKeySlice("_id")),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Key":    Equal(storeStructuredMongoTest.MakeKeySlice("id")),
					"Unique": Equal(true),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Key": Equal(storeStructuredMongoTest.MakeKeySlice("userId", "startAtTime")),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Key": Equal(storeStructuredMongoTest.MakeKeySlice("userId", "endAtTime")),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Key": Equal(storeStructuredMongoTest.MakeKeySlice("startAtTime")),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Key": Equal(storeStructuredMongoTest.MakeKeySlice("endAtTime")),
				}),
			))
		})

		It("blobs returns successfully", func() {
			//Expect(store.EnsureIndexes()).To(Succeed())
			cursor, err := blobsCollection.Indexes().List(context.Background())
			Expect(err).ToNot(HaveOccurred())
			Expect(cursor).ToNot(BeNil())
			var indexes []storeStructuredMongoTest.MongoIndex
			err = cursor.All(context.Background(), &indexes)
			Expect(err).ToNot(HaveOccurred())

			Expect(indexes).To(ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					"Key": Equal(storeStructuredMongoTest.MakeKeySlice("_id")),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("id")),
					"Background": Equal(true),
					"Unique":     Equal(true),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("userId")),
					"Background": Equal(true),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("mediaType")),
					"Background": Equal(true),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("status")),
					"Background": Equal(true),
				}),
			))
		})
	})
})

var suiteStore *blobStoreStructuredMongo.Store
var suiteStoreOnce sync.Once

func GetSuiteStore() *blobStoreStructuredMongo.Store {
	GinkgoHelper()
	suiteStoreOnce.Do(func() {
		base := storeStructuredMongoTest.GetSuiteStore()
		suiteStore = blobStoreStructuredMongo.NewStoreFromBase(base)
		Expect(suiteStore.EnsureIndexes()).To(Succeed())
	})
	return suiteStore
}
