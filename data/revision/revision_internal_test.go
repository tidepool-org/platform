package revision

import (
	"encoding/json"
	"reflect"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types"
	dataTypesBasalScheduled "github.com/tidepool-org/platform/data/types/basal/scheduled"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/errors"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	"github.com/tidepool-org/platform/test"
)

func fromJSON(text string) any {
	var value any
	Expect(json.Unmarshal([]byte(text), &value)).To(Succeed())
	return value
}

func mapFromJSON(text string) map[string]any {
	return fromJSON(text).(map[string]any)
}

var _ = Describe("Revision internals", func() {
	Context("OmittedKeys", func() {
		It("returns the bson names of the fields tagged to be omitted, then the extra keys", func() {
			type tagged struct {
				A string `json:"a" bson:"a" revision:"omit"`
				B string `json:"b" bson:"bb,omitempty" revision:"omit"`
				C string `json:"c" bson:"c"`
				D string `json:"-" bson:"_d" revision:"omit"`
			}
			Expect(omittedKeys(reflect.TypeOf(tagged{}), "_id")).To(Equal([]string{"a", "bb", "_d", "_id"}))
		})

		It("returns an error for a revision tag with another value", func() {
			type tagged struct {
				A string `json:"a" bson:"a" revision:"keep"`
			}
			keys, err := omittedKeys(reflect.TypeOf(tagged{}))
			Expect(err).To(MatchError(`field revision.tagged.A has revision tag "keep", only "omit" is allowed`))
			Expect(keys).To(BeNil())
		})

		It("returns an error for an omitted field without a bson name", func() {
			type tagged struct {
				A string `json:"a" bson:"-" revision:"omit"`
			}
			keys, err := omittedKeys(reflect.TypeOf(tagged{}))
			Expect(err).To(MatchError("field revision.tagged.A has no bson name"))
			Expect(keys).To(BeNil())
		})

		It("returns an error for a type that is not a struct", func() {
			keys, err := omittedKeys(reflect.TypeOf(""))
			Expect(err).To(MatchError("type string is not a struct"))
			Expect(keys).To(BeNil())
		})
	})

	Context("server-owned keys", func() {
		var serverOwnedKeys []string

		BeforeEach(func() {
			var err error
			serverOwnedKeys, err = omittedKeys(reflect.TypeOf(types.Base{}), "_id")
			Expect(err).ToNot(HaveOccurred())
		})

		// Every server-owned key lives on Base and is tagged revision:"omit". A field tagged json:"-" is server-only;
		// a field the external Parse leaves unparsed is server-set but client-visible; id is parsed but cleared and
		// regenerated; _id has no field. A new server-set field on Base fails here until it carries the tag.
		It("are exactly the keys Base owns on the server", func() {
			jsonToBSON := map[string]string{}
			expected := []string{"_id", "id"}
			baseType := reflect.TypeOf(types.Base{})
			for index := range baseType.NumField() {
				field := baseType.Field(index)
				jsonName, _, _ := strings.Cut(field.Tag.Get("json"), ",")
				bsonName, _, _ := strings.Cut(field.Tag.Get("bson"), ",")
				if jsonName == "-" {
					expected = append(expected, bsonName)
				} else {
					jsonToBSON[jsonName] = bsonName
				}
			}

			datum := dataTypesTest.RandomBase()
			datum.Revisions = dataTest.RandomRevisions()
			object := dataTypesTest.NewObjectFromBase(datum, test.ObjectFormatJSON)
			parser := structureParser.NewObject(logTest.NewLogger(), &object)
			datum.Parse(parser)
			for _, err := range errors.ToArray(parser.NotParsed()) {
				jsonName := strings.TrimPrefix(errors.AsSource(err).Pointer(), "/")
				if jsonName == "type" {
					continue // read by the factory, not by Base.Parse
				}
				Expect(jsonToBSON).To(HaveKey(jsonName))
				expected = append(expected, jsonToBSON[jsonName])
			}

			Expect(serverOwnedKeys).To(ConsistOf(expected))
		})

		It("name every server-set key by its BSON name", func() {
			Expect(serverOwnedKeys).To(ConsistOf(
				"_id", "_userId", "_active", "_version", "_deduplicator",
				"provenance", "revisions", "id", "uploadId", "guid",
				"createdTime", "createdUserId", "modifiedTime", "modifiedUserId",
				"deletedTime", "deletedUserId", "archivedTime", "archivedDatasetId",
			))
		})

		It("are reported by IsServerOwnedKey, and nothing else is", func() {
			for _, key := range serverOwnedKeys {
				Expect(isServerOwnedKey(key)).To(BeTrue(), key)
			}
			for _, key := range []string{"time", "rate", "origin", "payload", "type", "deviceId", "_other"} {
				Expect(isServerOwnedKey(key)).To(BeFalse(), key)
			}
		})
	})

	Context("Project", func() {
		It("removes every server-owned key and nothing else", func() {
			datum := dataTypesBasalScheduled.New()
			datum.Base = *dataTypesTest.RandomBase()
			datum.Active = true
			datum.Provenance = dataTest.RandomProvenance()
			datum.Revisions = dataTest.RandomRevisions()
			datum.VersionInternal = 3
			datum.Rate = pointer.FromFloat64(1.25)
			document, err := ToMap(datum)
			Expect(err).ToNot(HaveOccurred())
			serverOwnedKeys, err := omittedKeys(reflect.TypeOf(types.Base{}), "_id")
			Expect(err).ToNot(HaveOccurred())
			// _id is assigned by Mongo, so the typed datum has no such key
			for _, key := range serverOwnedKeys {
				if key != "_id" {
					Expect(document).To(HaveKey(key), key)
				}
			}
			projection := Project(document)
			expected := map[string]any{}
			for key, value := range document {
				if !isServerOwnedKey(key) {
					expected[key] = value
				}
			}
			Expect(projection).To(Equal(expected))
			Expect(projection).To(HaveLen(len(document) - len(serverOwnedKeys) + 1))
		})

	})

	Context("Merge", func() {
		DescribeTable("applies the RFC 7386 appendix A examples",
			func(target string, patch string, expected string) {
				if expectedValue := fromJSON(expected); expectedValue == nil {
					Expect(merge(fromJSON(target), fromJSON(patch))).To(BeNil())
				} else {
					Expect(merge(fromJSON(target), fromJSON(patch))).To(Equal(expectedValue))
				}
			},
			Entry("1", `{"a":"b"}`, `{"a":"c"}`, `{"a":"c"}`),
			Entry("2", `{"a":"b"}`, `{"b":"c"}`, `{"a":"b","b":"c"}`),
			Entry("3", `{"a":"b"}`, `{"a":null}`, `{}`),
			Entry("4", `{"a":"b","b":"c"}`, `{"a":null}`, `{"b":"c"}`),
			Entry("5", `{"a":["b"]}`, `{"a":"c"}`, `{"a":"c"}`),
			Entry("6", `{"a":"c"}`, `{"a":["b"]}`, `{"a":["b"]}`),
			Entry("7", `{"a":{"b":"c"}}`, `{"a":{"b":"d","c":null}}`, `{"a":{"b":"d"}}`),
			Entry("8", `{"a":[{"b":"c"}]}`, `{"a":[1]}`, `{"a":[1]}`),
			Entry("9", `["a","b"]`, `{"a":"b"}`, `{"a":"b"}`),
			Entry("10", `{"a":"foo"}`, `null`, `null`),
			Entry("11", `{"a":"foo"}`, `"bar"`, `"bar"`),
			Entry("12", `{"e":null}`, `{"a":1}`, `{"e":null,"a":1}`),
			Entry("13", `[1,2]`, `{"a":"b","c":null}`, `{"a":"b"}`),
			Entry("14", `{}`, `{"a":{"bb":{"ccc":null}}}`, `{"a":{"bb":{}}}`),
		)

		It("does not modify the target or the patch", func() {
			target := mapFromJSON(`{"a":{"b":"c"},"d":["e"]}`)
			patch := mapFromJSON(`{"a":{"b":"x"},"d":["y"],"f":{"g":"h"}}`)
			result := Apply(target, patch)
			result["a"].(map[string]any)["b"] = "mutated"
			result["d"].([]any)[0] = "mutated"
			result["f"].(map[string]any)["g"] = "mutated"
			Expect(target).To(Equal(mapFromJSON(`{"a":{"b":"c"},"d":["e"]}`)))
			Expect(patch).To(Equal(mapFromJSON(`{"a":{"b":"x"},"d":["y"],"f":{"g":"h"}}`)))
		})
	})

	Context("diff", func() {
		It("reports whether non-object values differ", func() {
			patch, changed := diff("a", "a")
			Expect(changed).To(BeFalse())
			Expect(patch).To(BeNil())
			patch, changed = diff("a", "b")
			Expect(changed).To(BeTrue())
			Expect(patch).To(Equal("b"))
			patch, changed = diff("a", nil)
			Expect(changed).To(BeTrue())
			Expect(patch).To(BeNil())
			patch, changed = diff("a", map[string]any{})
			Expect(changed).To(BeTrue())
			Expect(patch).To(Equal(map[string]any{}))
		})

	})

	Context("ChangedKeys", func() {
		It("splits the top-level keys of a patch into set and unset, sorted", func() {
			set, unset := changedKeys(map[string]any{"z": 1, "a": map[string]any{"x": nil}, "m": nil, "b": nil})
			Expect(set).To(Equal([]string{"a", "z"}))
			Expect(unset).To(Equal([]string{"b", "m"}))
		})

		It("returns nothing for an empty patch", func() {
			set, unset := changedKeys(map[string]any{})
			Expect(set).To(BeEmpty())
			Expect(unset).To(BeEmpty())
		})
	})

})
