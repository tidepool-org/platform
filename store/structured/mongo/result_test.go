package mongo_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/store/structured/mongo"
)

var _ = Describe("Result", func() {
	Describe("BSONToMap", func() {
		It("returns nil for nil input", func() {
			Expect(mongo.BSONToMap(nil)).To(BeNil())
		})

		It("handles empty bson.M", func() {
			Expect(mongo.BSONToMap(bson.M{})).To(Equal(map[string]any{}))
		})

		It("converts bson.M to map[string]any", func() {
			result := mongo.BSONToMap(bson.M{"string": "value", "number": 42})
			Expect(result).To(Equal(map[string]any{"string": "value", "number": 42}))
		})

		It("handles deeply nested structures", func() {
			result := mongo.BSONToMap(bson.M{"zero": bson.A{"nested", "array"}, "one": bson.M{"nested": "object"}})
			Expect(result).To(MatchAllKeys(Keys{
				"zero": Equal([]any{"nested", "array"}),
				"one":  Equal(map[string]any{"nested": "object"}),
			}))
		})
	})

	Describe("BSONToArray", func() {
		It("returns nil for nil input", func() {
			Expect(mongo.BSONToArray(nil)).To(BeNil())
		})

		It("handles empty bson.A", func() {
			Expect(mongo.BSONToArray(bson.A{})).To(Equal([]any{}))
		})

		It("converts bson.A to []any", func() {
			result := mongo.BSONToArray(bson.A{"value", 42})
			Expect(result).To(Equal([]any{"value", 42}))
		})

		It("handles deeply nested structures", func() {
			result := mongo.BSONToArray(bson.A{bson.A{"nested", "array"}, bson.M{"nested": "object"}})
			Expect(result).To(MatchAllElementsWithIndex(IndexIdentity, Elements{
				"0": Equal([]any{"nested", "array"}),
				"1": Equal(map[string]any{"nested": "object"}),
			}))
		})
	})

	Describe("BSONToAny", func() {
		It("returns nil for nil input", func() {
			Expect(mongo.BSONToAny(nil)).To(BeNil())
		})

		It("returns primitive types unchanged", func() {
			Expect(mongo.BSONToAny(true)).To(Equal(true))
			Expect(mongo.BSONToAny(42)).To(Equal(42))
			Expect(mongo.BSONToAny(42.345)).To(Equal(42.345))
			Expect(mongo.BSONToAny("string")).To(Equal("string"))
			Expect(mongo.BSONToAny(map[string]string{"string": "value", "number": "42"})).To(Equal(map[string]string{"string": "value", "number": "42"}))
			Expect(mongo.BSONToAny([]string{"value", "42"})).To(Equal([]string{"value", "42"}))
		})

		It("converts bson.M", func() {
			result := mongo.BSONToAny(bson.M{"zero": bson.A{"nested", "array"}, "one": bson.M{"nested": "object"}})
			Expect(result).To(MatchAllKeys(Keys{
				"zero": Equal([]any{"nested", "array"}),
				"one":  Equal(map[string]any{"nested": "object"}),
			}))
		})

		It("converts bson.A", func() {
			result := mongo.BSONToAny(bson.A{bson.A{"nested", "array"}, bson.M{"nested": "object"}})
			Expect(result).To(MatchAllElementsWithIndex(IndexIdentity, Elements{
				"0": Equal([]any{"nested", "array"}),
				"1": Equal(map[string]any{"nested": "object"}),
			}))
		})
	})
})
