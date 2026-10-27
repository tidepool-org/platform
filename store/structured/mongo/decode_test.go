package mongo_test

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Decode", func() {
	Context("DecodeDocument", func() {
		It("decodes documents nested in arrays and maps as plain values at every depth", func() {
			tm := test.RandomTime().UTC()
			bytes, err := bson.Marshal(bson.M{
				"a": bson.A{bson.M{"b": bson.A{bson.M{"c": primitive.NewDateTimeFromTime(tm)}}}},
				"d": bson.M{"e": bson.A{int32(1), bson.M{"f": "g"}}},
				"h": primitive.NewObjectID(),
			})
			Expect(err).ToNot(HaveOccurred())
			document, err := storeStructuredMongo.DecodeDocument(bytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(document["a"]).To(BeAssignableToTypeOf([]any{}))
			inner := document["a"].([]any)[0]
			Expect(inner).To(BeAssignableToTypeOf(map[string]any{}))
			innermost := inner.(map[string]any)["b"].([]any)[0]
			Expect(innermost).To(BeAssignableToTypeOf(map[string]any{}))
			Expect(innermost.(map[string]any)["c"]).To(Equal(tm))
			Expect(document["d"].(map[string]any)["e"].([]any)[1]).To(BeAssignableToTypeOf(map[string]any{}))
			Expect(document["h"]).To(BeAssignableToTypeOf(primitive.ObjectID{}))
		})

		It("marshals to JSON as a client sees the document", func() {
			tm := test.RandomTime().UTC()
			id := primitive.NewObjectID()
			bytes, err := bson.Marshal(bson.M{"_id": id, "time": primitive.NewDateTimeFromTime(tm), "a": bson.A{bson.M{"b": int32(1)}}})
			Expect(err).ToNot(HaveOccurred())
			document, err := storeStructuredMongo.DecodeDocument(bytes)
			Expect(err).ToNot(HaveOccurred())
			body, err := json.Marshal(document)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(body)).To(Equal(`{"_id":"` + id.Hex() + `","a":[{"b":1}],"time":"` + tm.Format(time.RFC3339Nano) + `"}`))
		})

		It("decodes a datetime in UTC whatever zone it was marshalled in, truncated to the millisecond", func() {
			local := time.Date(2026, 10, 20, 10, 0, 0, 123456789, time.FixedZone("plus-two", 2*60*60))
			bytes, err := bson.Marshal(bson.M{"time": local})
			Expect(err).ToNot(HaveOccurred())
			document, err := storeStructuredMongo.DecodeDocument(bytes)
			Expect(err).ToNot(HaveOccurred())
			decoded := document["time"].(time.Time)
			Expect(decoded.Location()).To(Equal(time.UTC))
			Expect(decoded).To(Equal(time.Date(2026, 10, 20, 8, 0, 0, 123000000, time.UTC)))
			Expect(decoded.Format(time.RFC3339Nano)).To(Equal("2026-10-20T08:00:00.123Z"))
		})

		It("returns an error for malformed BSON", func() {
			document, err := storeStructuredMongo.DecodeDocument([]byte{1, 2, 3})
			Expect(err).To(HaveOccurred())
			Expect(document).To(BeNil())
		})
	})

	Context("UnmarshalDocument", func() {
		It("decodes into a struct with plain values where the struct leaves the type open", func() {
			type target struct {
				Name string         `bson:"name"`
				When time.Time      `bson:"when"`
				Undo map[string]any `bson:"undo"`
			}
			tm := test.RandomTime().UTC()
			bytes, err := bson.Marshal(bson.M{"name": "n", "when": primitive.NewDateTimeFromTime(tm), "undo": bson.M{"a": bson.A{"x"}, "t": primitive.NewDateTimeFromTime(tm)}})
			Expect(err).ToNot(HaveOccurred())
			result := target{}
			Expect(storeStructuredMongo.UnmarshalDocument(bytes, &result)).To(Succeed())
			Expect(result.Name).To(Equal("n"))
			Expect(result.When).To(Equal(tm))
			Expect(result.Undo).To(Equal(map[string]any{"a": []any{"x"}, "t": tm}))
		})
	})
})
