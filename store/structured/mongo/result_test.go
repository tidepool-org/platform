package mongo_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
)

var _ = Describe("Result", func() {
	Describe("BSONToMap", func() {
		It("returns nil for nil input", func() {
			Expect(storeStructuredMongo.BSONToMap(nil)).To(BeNil())
		})

		It("handles empty bson.M", func() {
			Expect(storeStructuredMongo.BSONToMap(bson.M{})).To(Equal(map[string]any{}))
		})

		It("converts bson.M to map[string]any", func() {
			result := storeStructuredMongo.BSONToMap(bson.M{"string": "value", "number": 42})
			Expect(result).To(Equal(map[string]any{"string": "value", "number": 42}))
		})

		It("handles deeply nested structures", func() {
			result := storeStructuredMongo.BSONToMap(bson.M{"zero": bson.A{"nested", "array"}, "one": bson.M{"nested": "object"}})
			Expect(result).To(MatchAllKeys(Keys{
				"zero": Equal([]any{"nested", "array"}),
				"one":  Equal(map[string]any{"nested": "object"}),
			}))
		})
	})

	Describe("BSONToArray", func() {
		It("returns nil for nil input", func() {
			Expect(storeStructuredMongo.BSONToArray(nil)).To(BeNil())
		})

		It("handles empty bson.A", func() {
			Expect(storeStructuredMongo.BSONToArray(bson.A{})).To(Equal([]any{}))
		})

		It("converts bson.A to []any", func() {
			result := storeStructuredMongo.BSONToArray(bson.A{"value", 42})
			Expect(result).To(Equal([]any{"value", 42}))
		})

		It("handles deeply nested structures", func() {
			result := storeStructuredMongo.BSONToArray(bson.A{bson.A{"nested", "array"}, bson.M{"nested": "object"}})
			Expect(result).To(MatchAllElementsWithIndex(IndexIdentity, Elements{
				"0": Equal([]any{"nested", "array"}),
				"1": Equal(map[string]any{"nested": "object"}),
			}))
		})
	})

	Describe("BSONToAny", func() {
		It("returns nil for nil input", func() {
			Expect(storeStructuredMongo.BSONToAny(nil)).To(BeNil())
		})

		It("returns primitive types unchanged", func() {
			Expect(storeStructuredMongo.BSONToAny(true)).To(Equal(true))
			Expect(storeStructuredMongo.BSONToAny(42)).To(Equal(42))
			Expect(storeStructuredMongo.BSONToAny(42.345)).To(Equal(42.345))
			Expect(storeStructuredMongo.BSONToAny("string")).To(Equal("string"))
			Expect(storeStructuredMongo.BSONToAny(map[string]string{"string": "value", "number": "42"})).To(Equal(map[string]string{"string": "value", "number": "42"}))
			Expect(storeStructuredMongo.BSONToAny([]string{"value", "42"})).To(Equal([]string{"value", "42"}))
		})

		It("converts bson.M", func() {
			result := storeStructuredMongo.BSONToAny(bson.M{"zero": bson.A{"nested", "array"}, "one": bson.M{"nested": "object"}})
			Expect(result).To(MatchAllKeys(Keys{
				"zero": Equal([]any{"nested", "array"}),
				"one":  Equal(map[string]any{"nested": "object"}),
			}))
		})

		It("converts bson.A", func() {
			result := storeStructuredMongo.BSONToAny(bson.A{bson.A{"nested", "array"}, bson.M{"nested": "object"}})
			Expect(result).To(MatchAllElementsWithIndex(IndexIdentity, Elements{
				"0": Equal([]any{"nested", "array"}),
				"1": Equal(map[string]any{"nested": "object"}),
			}))
		})
	})

	Describe("CloseCursor", func() {
		var lgr *logTest.Logger
		var ctx context.Context
		var repository *storeStructuredMongo.Repository

		BeforeEach(func() {
			lgr = logTest.NewLogger()
			ctx = log.NewContextWithLogger(context.Background(), lgr)
			repository = storeStructuredMongoTest.GetSuiteStore().GetRepository(storeStructuredMongoTest.NewCollectionPrefix())
			_, err := repository.InsertMany(ctx, []any{bson.M{"value": 1}, bson.M{"value": 2}})
			Expect(err).ToNot(HaveOccurred())
		})

		openCursor := func() *mongo.Cursor {
			cursor, err := repository.Find(ctx, bson.M{}, options.Find().SetBatchSize(1))
			Expect(err).ToNot(HaveOccurred())
			Expect(cursor.ID()).ToNot(BeZero())
			return cursor
		}

		It("does not panic when the cursor is nil", func() {
			Expect(func() { storeStructuredMongo.CloseCursor(ctx, nil) }).ToNot(Panic())
		})

		It("does not panic when both the context and the cursor are nil", func() {
			Expect(func() { storeStructuredMongo.CloseCursor(context.Context(nil), nil) }).ToNot(Panic())
		})

		It("closes the cursor when the context is provided", func() {
			cursor := openCursor()
			storeStructuredMongo.CloseCursor(ctx, cursor)
			Expect(cursor.ID()).To(BeZero())
		})

		It("closes the cursor when the context is nil", func() {
			cursor := openCursor()
			storeStructuredMongo.CloseCursor(context.Context(nil), cursor)
			Expect(cursor.ID()).To(BeZero())
		})
	})
})
