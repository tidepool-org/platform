package parser_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/parser"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/service"
)

var _ = Describe("StandardObject", func() {
	var standardContext *context.Standard

	BeforeEach(func() {
		var err error
		standardContext, err = context.NewStandard(null.NewLogger())
		Expect(err).ToNot(HaveOccurred())
		Expect(standardContext).ToNot(BeNil())
	})

	It("NewStandardObject returns an error if context is nil", func() {
		standard, err := parser.NewStandardObject(nil, &map[string]interface{}{}, parser.IgnoreNotParsed)
		Expect(err).To(MatchError("context is missing"))
		Expect(standard).To(BeNil())
	})

	Context("new standard object with nil object", func() {
		var standardObject *parser.StandardObject

		BeforeEach(func() {
			var err error
			standardObject, err = parser.NewStandardObject(standardContext, nil, parser.IgnoreNotParsed)
			Expect(err).ToNot(HaveOccurred())
		})

		It("exists", func() {
			Expect(standardObject).ToNot(BeNil())
		})

		It("does not have a contained object", func() {
			Expect(standardObject.Object()).To(BeNil())
		})

		It("Logger returns a logger", func() {
			Expect(standardObject.Logger()).ToNot(BeNil())
		})

		It("SetMeta sets the meta on the context", func() {
			meta := "metametameta"
			standardObject.SetMeta(meta)
			Expect(standardContext.Meta()).To(BeIdenticalTo(meta))
		})

		It("AppendError appends an error on the context", func() {
			standardObject.AppendError("append-error", &service.Error{})
			Expect(standardContext.Errors()).To(HaveLen(1))
		})

		It("ParseBoolean returns nil", func() {
			Expect(standardObject.ParseBoolean("0")).To(BeNil())
		})

		It("ParseInteger returns nil", func() {
			Expect(standardObject.ParseInteger("1")).To(BeNil())
		})

		It("ParseFloat returns nil", func() {
			Expect(standardObject.ParseFloat("2")).To(BeNil())
		})

		It("ParseString returns nil", func() {
			Expect(standardObject.ParseString("3")).To(BeNil())
		})

		It("ParseStringArray returns nil", func() {
			Expect(standardObject.ParseStringArray("4")).To(BeNil())
		})

		It("ParseObject returns nil", func() {
			Expect(standardObject.ParseObject("5")).To(BeNil())
		})

		It("ParseObjectArray returns nil", func() {
			Expect(standardObject.ParseObjectArray("6")).To(BeNil())
		})

		It("ParseInterface returns nil", func() {
			Expect(standardObject.ParseInterface("7")).To(BeNil())
		})

		It("ParseInterfaceArray returns nil", func() {
			Expect(standardObject.ParseInterfaceArray("8")).To(BeNil())
		})

		It("ProcessNotParsed does not add an error", func() {
			standardObject.ProcessNotParsed()
			Expect(standardContext.Errors()).To(BeEmpty())
		})

		It("NewChildObjectParser returns an object parser with a nil object", func() {
			objectParser := standardObject.NewChildObjectParser("0")
			Expect(objectParser).ToNot(BeNil())
			Expect(objectParser.Object()).To(BeNil())
		})

		It("NewChildArrayParser returns an array parser with a nil array", func() {
			arrayParser := standardObject.NewChildArrayParser("0")
			Expect(arrayParser).ToNot(BeNil())
			Expect(arrayParser.Array()).To(BeNil())
		})
	})

	Context("new standard object with valid, empty object", func() {
		var standardObject *parser.StandardObject

		BeforeEach(func() {
			var err error
			standardObject, err = parser.NewStandardObject(standardContext, &map[string]interface{}{}, parser.IgnoreNotParsed)
			Expect(err).ToNot(HaveOccurred())
		})

		It("exists", func() {
			Expect(standardObject).ToNot(BeNil())
		})

		It("has a contained Object", func() {
			Expect(standardObject.Object()).ToNot(BeNil())
		})

		It("ProcessNotParsed does not add an error", func() {
			standardObject.ProcessNotParsed()
			Expect(standardContext.Errors()).To(BeEmpty())
		})
	})

	Context("parsing elements with", func() {
		var standardObject *parser.StandardObject

		Context("ParseBoolean", func() {
			BeforeEach(func() {
				standardObject, _ = parser.NewStandardObject(standardContext, &map[string]interface{}{
					"zero": "not a boolean",
					"one":  true,
				}, parser.IgnoreNotParsed)
			})

			It("with key not found in the object returns nil", func() {
				Expect(standardObject.ParseBoolean("unknown")).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with key with value with different type returns nil and appends an ErrorTypeNotBoolean", func() {
				Expect(standardObject.ParseBoolean("zero")).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-boolean"))
			})

			It("with key with value with boolean type returns value", func() {
				value := standardObject.ParseBoolean("one")
				Expect(value).ToNot(BeNil())
				Expect(*value).To(BeTrue())
				Expect(standardContext.Errors()).To(BeEmpty())
			})
		})

		Context("ParseInteger", func() {
			BeforeEach(func() {
				standardObject, _ = parser.NewStandardObject(standardContext, &map[string]interface{}{
					"zero":  false,
					"one":   3,
					"two":   4.0,
					"three": 5.67,
				}, parser.IgnoreNotParsed)
			})

			It("with key not found in the object returns nil", func() {
				Expect(standardObject.ParseInteger("unknown")).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with key with value with different type returns nil and appends an ErrorTypeNotInteger", func() {
				Expect(standardObject.ParseInteger("zero")).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-integer"))
			})

			It("with key with value with integer type returns value", func() {
				value := standardObject.ParseInteger("one")
				Expect(value).ToNot(BeNil())
				Expect(*value).To(BeNumerically("==", 3))
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with key with value with float type and whole number returns value", func() {
				value := standardObject.ParseInteger("two")
				Expect(value).ToNot(BeNil())
				Expect(*value).To(BeNumerically("==", 4))
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with key with value with float type and not whole number returns nil and appends an ErrorTypeNotInteger", func() {
				Expect(standardObject.ParseInteger("three")).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-integer"))
			})
		})

		Context("ParseFloat", func() {
			BeforeEach(func() {
				standardObject, _ = parser.NewStandardObject(standardContext, &map[string]interface{}{
					"zero":  false,
					"one":   3,
					"two":   4.0,
					"three": 5.67,
				}, parser.IgnoreNotParsed)
			})

			It("with key not found in the object returns nil", func() {
				Expect(standardObject.ParseFloat("unknown")).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with key with value with different type returns nil and appends an ErrorTypeNotFloat", func() {
				Expect(standardObject.ParseFloat("zero")).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-float"))
			})

			It("with key with value with integer type returns value", func() {
				value := standardObject.ParseFloat("one")
				Expect(value).ToNot(BeNil())
				Expect(*value).To(BeNumerically("==", 3.))
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with key with value with float type and whole number returns value", func() {
				value := standardObject.ParseFloat("two")
				Expect(value).ToNot(BeNil())
				Expect(*value).To(BeNumerically("==", 4.0))
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with key with value with float type and not whole number returns value", func() {
				value := standardObject.ParseFloat("three")
				Expect(value).ToNot(BeNil())
				Expect(*value).To(BeNumerically("==", 5.67))
				Expect(standardContext.Errors()).To(BeEmpty())
			})
		})

		Context("ParseString", func() {
			BeforeEach(func() {
				standardObject, _ = parser.NewStandardObject(standardContext, &map[string]interface{}{
					"zero": false,
					"one":  "this is a string",
				}, parser.IgnoreNotParsed)
			})

			It("with key not found in the object returns nil", func() {
				Expect(standardObject.ParseString("unknown")).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with key with value with different type returns nil and appends an ErrorTypeNotString", func() {
				Expect(standardObject.ParseString("zero")).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-string"))
			})

			It("with key with value with string type returns value", func() {
				value := standardObject.ParseString("one")
				Expect(value).ToNot(BeNil())
				Expect(*value).To(Equal("this is a string"))
				Expect(standardContext.Errors()).To(BeEmpty())
			})
		})

		Context("ParseStringArray", func() {
			BeforeEach(func() {
				standardObject, _ = parser.NewStandardObject(standardContext, &map[string]interface{}{
					"zero": false,
					"one": []string{
						"one",
						"two",
					},
					"two": []interface{}{
						"three",
						"four",
					},
					"three": []interface{}{
						"five",
						6,
					},
				}, parser.IgnoreNotParsed)
			})

			It("with key not found in the object returns nil", func() {
				Expect(standardObject.ParseStringArray("unknown")).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with key with value with different type returns nil and appends an ErrorTypeNotArray", func() {
				Expect(standardObject.ParseStringArray("zero")).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-array"))
			})

			It("with key with value with string array type returns value", func() {
				value := standardObject.ParseStringArray("one")
				Expect(value).ToNot(BeNil())
				Expect(*value).To(Equal([]string{"one", "two"}))
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with key with value with interface array and contains all string type returns value", func() {
				value := standardObject.ParseStringArray("two")
				Expect(value).ToNot(BeNil())
				Expect(*value).To(Equal([]string{"three", "four"}))
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with key with value with interface array and does not contains all string type returns partial value and error", func() {
				value := standardObject.ParseStringArray("three")
				Expect(value).ToNot(BeNil())
				Expect(*value).To(Equal([]string{"five", ""}))
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-string"))
			})
		})

		Context("ParseTime", func() {
			BeforeEach(func() {
				standardObject, _ = parser.NewStandardObject(standardContext, &map[string]interface{}{
					"zero": false,
					"one":  "this is a string",
					"two":  "2017-12-15T12:34:56Z",
				}, parser.IgnoreNotParsed)
			})

			It("with key not found in the object returns nil", func() {
				Expect(standardObject.ParseTime("unknown", time.RFC3339)).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with key with value with different type returns nil and appends an ErrorTypeNotTime", func() {
				Expect(standardObject.ParseTime("zero", time.RFC3339)).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-time"))
			})

			It("with index parameter with not parseable table returns nil and appends an ErrorValueTimeNotParsable", func() {
				Expect(standardObject.ParseTime("one", time.RFC3339)).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("value-not-parsable"))
			})

			It("with index parameter with time type returns value", func() {
				value := standardObject.ParseTime("two", time.RFC3339)
				Expect(value).ToNot(BeNil())
				Expect(value.Format(time.RFC3339)).To(Equal("2017-12-15T12:34:56Z"))
				Expect(standardContext.Errors()).To(BeEmpty())
			})
		})

		Context("ParseObject", func() {
			BeforeEach(func() {
				standardObject, _ = parser.NewStandardObject(standardContext, &map[string]interface{}{
					"zero": false,
					"one": map[string]interface{}{
						"1": "2",
					},
				}, parser.IgnoreNotParsed)
			})

			It("with key not found in the object returns nil", func() {
				Expect(standardObject.ParseObject("unknown")).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with key with value with different type returns nil and appends an ErrorTypeNotObject", func() {
				Expect(standardObject.ParseObject("zero")).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-object"))
			})

			It("with key with value with object type returns value", func() {
				value := standardObject.ParseObject("one")
				Expect(value).ToNot(BeNil())
				Expect(*value).To(Equal(map[string]interface{}{"1": "2"}))
				Expect(standardContext.Errors()).To(BeEmpty())
			})
		})

		Context("ParseObjectArray", func() {
			BeforeEach(func() {
				standardObject, _ = parser.NewStandardObject(standardContext, &map[string]interface{}{
					"zero": false,
					"one": []map[string]interface{}{
						{
							"1": "2",
						},
						{
							"3": "4",
						},
					},
					"two": []interface{}{
						map[string]interface{}{
							"5": "6",
						},
						map[string]interface{}{
							"7": "8",
						},
					},
					"three": []interface{}{
						map[string]interface{}{
							"9": "0",
						},
						"not",
					},
				}, parser.IgnoreNotParsed)
			})

			It("with key not found in the object returns nil", func() {
				Expect(standardObject.ParseObjectArray("unknown")).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with key with value with different type returns nil and appends an ErrorTypeNotArray", func() {
				Expect(standardObject.ParseObjectArray("zero")).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-array"))
			})

			It("with key with value with object array type returns value", func() {
				value := standardObject.ParseObjectArray("one")
				Expect(value).ToNot(BeNil())
				Expect(*value).To(Equal([]map[string]interface{}{{"1": "2"}, {"3": "4"}}))
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with key with value with interface array and contains all object type returns value", func() {
				value := standardObject.ParseObjectArray("two")
				Expect(value).ToNot(BeNil())
				Expect(*value).To(Equal([]map[string]interface{}{{"5": "6"}, {"7": "8"}}))
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with key with value with interface array and does not contains all object type returns partial value and error", func() {
				value := standardObject.ParseObjectArray("three")
				Expect(value).ToNot(BeNil())
				Expect(*value).To(Equal([]map[string]interface{}{{"9": "0"}, nil}))
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-object"))
			})
		})

		Context("ParseInterface", func() {
			BeforeEach(func() {
				standardObject, _ = parser.NewStandardObject(standardContext, &map[string]interface{}{
					"zero": false,
					"one":  "zombie",
				}, parser.IgnoreNotParsed)
			})

			It("with key not found in the object returns nil", func() {
				Expect(standardObject.ParseInterface("unknown")).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with key with value with object type returns value", func() {
				value := standardObject.ParseInterface("zero")
				Expect(value).ToNot(BeNil())
				Expect(*value).To(BeFalse())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with key with value with object type returns value", func() {
				value := standardObject.ParseInterface("one")
				Expect(value).ToNot(BeNil())
				Expect(*value).To(Equal("zombie"))
				Expect(standardContext.Errors()).To(BeEmpty())
			})
		})

		Context("ParseInterfaceArray", func() {
			BeforeEach(func() {
				standardObject, _ = parser.NewStandardObject(standardContext, &map[string]interface{}{
					"zero": false,
					"one": []interface{}{
						"1",
						false,
					},
				}, parser.IgnoreNotParsed)
			})

			It("with key not found in the object returns nil", func() {
				Expect(standardObject.ParseInterfaceArray("unknown")).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with key with value with different type returns nil and appends an ErrorTypeNotArray", func() {
				Expect(standardObject.ParseInterfaceArray("zero")).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-array"))
			})

			It("with key with value with object array type returns value", func() {
				value := standardObject.ParseInterfaceArray("one")
				Expect(value).ToNot(BeNil())
				Expect(*value).To(Equal([]interface{}{"1", false}))
				Expect(standardContext.Errors()).To(BeEmpty())
			})
		})

		Context("ProcessNotParsed", func() {
			Context("with ParsedPolicy as IgnoreNotParsed", func() {
				BeforeEach(func() {
					standardObject, _ = parser.NewStandardObject(standardContext, &map[string]interface{}{
						"one":   1,
						"two":   "two",
						"three": 3,
					}, parser.IgnoreNotParsed)
				})

				It("without anything parsed has no errors", func() {
					standardObject.ProcessNotParsed()
					Expect(standardContext.Errors()).To(BeEmpty())
				})
			})

			Context("with ParsedPolicy as WarnLoggerNotParsed", func() {
				BeforeEach(func() {
					standardObject, _ = parser.NewStandardObject(standardContext, &map[string]interface{}{
						"one":   1,
						"two":   "two",
						"three": 3,
					}, parser.WarnLoggerNotParsed)
				})

				It("without anything parsed has no errors", func() {
					standardObject.ProcessNotParsed()
					Expect(standardContext.Errors()).To(BeEmpty())
				})
			})

			Context("with ParsedPolicy as AppendErrorNotParsed", func() {
				BeforeEach(func() {
					standardObject, _ = parser.NewStandardObject(standardContext, &map[string]interface{}{
						"one":   1,
						"two":   "two",
						"three": 3,
					}, parser.AppendErrorNotParsed)
				})

				It("without anything parsed appends all unparsed as errors", func() {
					standardObject.ProcessNotParsed()
					Expect(standardContext.Errors()).To(HaveLen(3))
					Expect(standardContext.Errors()[0].Code).To(Equal("not-parsed"))
					Expect(standardContext.Errors()[1].Code).To(Equal("not-parsed"))
					Expect(standardContext.Errors()[2].Code).To(Equal("not-parsed"))
				})

				It("with some items parsed appends all unparsed as errors", func() {
					standardObject.ParseString("two")
					standardObject.ProcessNotParsed()
					Expect(standardContext.Errors()).To(HaveLen(2))
					Expect(standardContext.Errors()[0].Code).To(Equal("not-parsed"))
					Expect(standardContext.Errors()[1].Code).To(Equal("not-parsed"))
				})

				It("with all items parsed has no errors", func() {
					standardObject.ParseInteger("one")
					standardObject.ParseString("two")
					standardObject.ParseInteger("three")
					standardObject.ProcessNotParsed()
					Expect(standardContext.Errors()).To(BeEmpty())
				})
			})
		})

		Context("NewChildObjectParser", func() {
			BeforeEach(func() {
				standardObject, _ = parser.NewStandardObject(standardContext, &map[string]interface{}{
					"zero": false,
					"one": map[string]interface{}{
						"1": "2",
					},
				}, parser.IgnoreNotParsed)
			})

			It("with key not found in the object returns nil", func() {
				objectParser := standardObject.NewChildObjectParser("unknown")
				Expect(objectParser).ToNot(BeNil())
				Expect(objectParser.Object()).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with key with value with different type returns nil and appends an ErrorTypeNotObject", func() {
				objectParser := standardObject.NewChildObjectParser("zero")
				Expect(objectParser).ToNot(BeNil())
				Expect(objectParser.Object()).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-object"))
			})

			It("with key with value with object type returns value", func() {
				objectParser := standardObject.NewChildObjectParser("one")
				Expect(objectParser).ToNot(BeNil())
				Expect(objectParser.Logger()).ToNot(BeNil())
				Expect(objectParser.Object()).ToNot(BeNil())
				Expect(*objectParser.Object()).To(Equal(map[string]interface{}{"1": "2"}))
				Expect(standardContext.Errors()).To(BeEmpty())
			})
		})

		Context("NewChildArrayParser", func() {
			BeforeEach(func() {
				standardObject, _ = parser.NewStandardObject(standardContext, &map[string]interface{}{
					"zero": false,
					"one": []interface{}{
						"1",
						false,
					},
				}, parser.IgnoreNotParsed)
			})

			It("with key not found in the object returns nil", func() {
				arrayParser := standardObject.NewChildArrayParser("unknown")
				Expect(arrayParser).ToNot(BeNil())
				Expect(arrayParser.Array()).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with key with value with different type returns nil and appends an ErrorTypeNotObject", func() {
				arrayParser := standardObject.NewChildArrayParser("zero")
				Expect(arrayParser).ToNot(BeNil())
				Expect(arrayParser.Array()).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-array"))
			})

			It("with key with value with object type returns value", func() {
				arrayParser := standardObject.NewChildArrayParser("one")
				Expect(arrayParser).ToNot(BeNil())
				Expect(arrayParser.Logger()).ToNot(BeNil())
				Expect(arrayParser.Array()).ToNot(BeNil())
				Expect(*arrayParser.Array()).To(Equal([]interface{}{"1", false}))
				Expect(standardContext.Errors()).To(BeEmpty())
			})
		})
	})
})
