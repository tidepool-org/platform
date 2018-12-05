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

var _ = Describe("StandardArray", func() {
	var standardContext *context.Standard

	BeforeEach(func() {
		var err error
		standardContext, err = context.NewStandard(null.NewLogger())
		Expect(err).ToNot(HaveOccurred())
		Expect(standardContext).ToNot(BeNil())
	})

	It("NewStandardArray returns an error if context is nil", func() {
		standard, err := parser.NewStandardArray(nil, &[]interface{}{}, parser.IgnoreNotParsed)
		Expect(err).To(MatchError("context is missing"))
		Expect(standard).To(BeNil())
	})

	Context("new standard array with nil array", func() {
		var standardArray *parser.StandardArray

		BeforeEach(func() {
			var err error
			standardArray, err = parser.NewStandardArray(standardContext, nil, parser.IgnoreNotParsed)
			Expect(err).ToNot(HaveOccurred())
		})

		It("exists", func() {
			Expect(standardArray).ToNot(BeNil())
		})

		It("does not have a contained array", func() {
			Expect(standardArray.Array()).To(BeNil())
		})

		It("Logger returns a logger", func() {
			Expect(standardArray.Logger()).ToNot(BeNil())
		})

		It("SetMeta sets the meta on the context", func() {
			meta := "metametameta"
			standardArray.SetMeta(meta)
			Expect(standardContext.Meta()).To(BeIdenticalTo(meta))
		})

		It("AppendError appends an error on the context", func() {
			standardArray.AppendError("append-error", &service.Error{})
			Expect(standardContext.Errors()).To(HaveLen(1))
		})

		It("ParseBoolean returns nil", func() {
			Expect(standardArray.ParseBoolean(0)).To(BeNil())
		})

		It("ParseInteger returns nil", func() {
			Expect(standardArray.ParseInteger(1)).To(BeNil())
		})

		It("ParseFloat returns nil", func() {
			Expect(standardArray.ParseFloat(2)).To(BeNil())
		})

		It("ParseString returns nil", func() {
			Expect(standardArray.ParseString(3)).To(BeNil())
		})

		It("ParseStringArray returns nil", func() {
			Expect(standardArray.ParseStringArray(4)).To(BeNil())
		})

		It("ParseObject returns nil", func() {
			Expect(standardArray.ParseObject(5)).To(BeNil())
		})

		It("ParseObjectArray returns nil", func() {
			Expect(standardArray.ParseObjectArray(6)).To(BeNil())
		})

		It("ParseInterface returns nil", func() {
			Expect(standardArray.ParseInterface(7)).To(BeNil())
		})

		It("ParseInterfaceArray returns nil", func() {
			Expect(standardArray.ParseInterfaceArray(8)).To(BeNil())
		})

		It("ProcessNotParsed does not add an error", func() {
			standardArray.ProcessNotParsed()
			Expect(standardContext.Errors()).To(BeEmpty())
		})

		It("NewChildObjectParser returns an object parser with a nil object", func() {
			objectParser := standardArray.NewChildObjectParser(7)
			Expect(objectParser).ToNot(BeNil())
			Expect(objectParser.Object()).To(BeNil())
		})

		It("NewChildArrayParser returns an array parser with a nil array", func() {
			arrayParser := standardArray.NewChildArrayParser(8)
			Expect(arrayParser).ToNot(BeNil())
			Expect(arrayParser.Array()).To(BeNil())
		})
	})

	Context("new standard array with valid, empty array", func() {
		var standardArray *parser.StandardArray

		BeforeEach(func() {
			var err error
			standardArray, err = parser.NewStandardArray(standardContext, &[]interface{}{}, parser.IgnoreNotParsed)
			Expect(err).ToNot(HaveOccurred())
		})

		It("exists", func() {
			Expect(standardArray).ToNot(BeNil())
		})

		It("has a contained Array", func() {
			Expect(standardArray.Array()).ToNot(BeNil())
		})

		It("ProcessNotParsed does not add an error", func() {
			standardArray.ProcessNotParsed()
			Expect(standardContext.Errors()).To(BeEmpty())
		})
	})

	Context("parsing elements with", func() {
		var standardArray *parser.StandardArray

		Context("ParseBoolean", func() {
			BeforeEach(func() {
				standardArray, _ = parser.NewStandardArray(standardContext, &[]interface{}{
					"not a boolean",
					true,
				}, parser.IgnoreNotParsed)
			})

			It("with index parameter less that the first index in the array returns nil", func() {
				Expect(standardArray.ParseBoolean(-1)).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter greater than the last index in the array returns nil", func() {
				Expect(standardArray.ParseBoolean(len(*standardArray.Array()))).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter with different type returns nil and appends an ErrorTypeNotBoolean", func() {
				Expect(standardArray.ParseBoolean(0)).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-boolean"))
			})

			It("with index parameter with boolean type returns value", func() {
				value := standardArray.ParseBoolean(1)
				Expect(value).ToNot(BeNil())
				Expect(*value).To(BeTrue())
				Expect(standardContext.Errors()).To(BeEmpty())
			})
		})

		Context("ParseInteger", func() {
			BeforeEach(func() {
				standardArray, _ = parser.NewStandardArray(standardContext, &[]interface{}{
					false,
					3,
					4.0,
					5.67,
				}, parser.IgnoreNotParsed)
			})

			It("with index parameter less that the first index in the array returns nil", func() {
				Expect(standardArray.ParseInteger(-1)).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter greater than the last index in the array returns nil", func() {
				Expect(standardArray.ParseInteger(len(*standardArray.Array()))).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter with different type returns nil and appends an ErrorTypeNotInteger", func() {
				Expect(standardArray.ParseInteger(0)).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-integer"))
			})

			It("with index parameter with integer type returns value", func() {
				value := standardArray.ParseInteger(1)
				Expect(value).ToNot(BeNil())
				Expect(*value).To(BeNumerically("==", 3))
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter with float type and whole number returns value", func() {
				value := standardArray.ParseInteger(2)
				Expect(value).ToNot(BeNil())
				Expect(*value).To(BeNumerically("==", 4))
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter with float type and not whole number returns nil and appends an ErrorTypeNotInteger", func() {
				Expect(standardArray.ParseInteger(3)).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-integer"))
			})
		})

		Context("ParseFloat", func() {
			BeforeEach(func() {
				standardArray, _ = parser.NewStandardArray(standardContext, &[]interface{}{
					false,
					3,
					4.0,
					5.67,
				}, parser.IgnoreNotParsed)
			})

			It("with index parameter less that the first index in the array returns nil", func() {
				Expect(standardArray.ParseFloat(-1)).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter greater than the last index in the array returns nil", func() {
				Expect(standardArray.ParseFloat(len(*standardArray.Array()))).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter with different type returns nil and appends an ErrorTypeNotFloat", func() {
				Expect(standardArray.ParseFloat(0)).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-float"))
			})

			It("with index parameter with integer type returns value", func() {
				value := standardArray.ParseFloat(1)
				Expect(value).ToNot(BeNil())
				Expect(*value).To(BeNumerically("==", 3.))
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter with float type and whole number returns value", func() {
				value := standardArray.ParseFloat(2)
				Expect(value).ToNot(BeNil())
				Expect(*value).To(BeNumerically("==", 4.0))
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter with float type and not whole number returns value", func() {
				value := standardArray.ParseFloat(3)
				Expect(value).ToNot(BeNil())
				Expect(*value).To(BeNumerically("==", 5.67))
				Expect(standardContext.Errors()).To(BeEmpty())
			})
		})

		Context("ParseString", func() {
			BeforeEach(func() {
				standardArray, _ = parser.NewStandardArray(standardContext, &[]interface{}{
					false,
					"this is a string",
				}, parser.IgnoreNotParsed)
			})

			It("with index parameter less that the first index in the array returns nil", func() {
				Expect(standardArray.ParseString(-1)).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter greater than the last index in the array returns nil", func() {
				Expect(standardArray.ParseString(len(*standardArray.Array()))).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter with different type returns nil and appends an ErrorTypeNotString", func() {
				Expect(standardArray.ParseString(0)).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-string"))
			})

			It("with index parameter with string type returns value", func() {
				value := standardArray.ParseString(1)
				Expect(value).ToNot(BeNil())
				Expect(*value).To(Equal("this is a string"))
				Expect(standardContext.Errors()).To(BeEmpty())
			})
		})

		Context("ParseStringArray", func() {
			BeforeEach(func() {
				standardArray, _ = parser.NewStandardArray(standardContext, &[]interface{}{
					false,
					[]string{
						"one",
						"two",
					},
					[]interface{}{
						"three",
						"four",
					},
					[]interface{}{
						"five",
						6,
					},
				}, parser.IgnoreNotParsed)
			})

			It("with index parameter less that the first index in the array returns nil", func() {
				Expect(standardArray.ParseStringArray(-1)).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter greater than the last index in the array returns nil", func() {
				Expect(standardArray.ParseStringArray(len(*standardArray.Array()))).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter with different type returns nil and appends an ErrorTypeNotArray", func() {
				Expect(standardArray.ParseStringArray(0)).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-array"))
			})

			It("with index parameter with string array type returns value", func() {
				value := standardArray.ParseStringArray(1)
				Expect(value).ToNot(BeNil())
				Expect(*value).To(Equal([]string{"one", "two"}))
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter with interface array and contains all string type returns value", func() {
				value := standardArray.ParseStringArray(2)
				Expect(value).ToNot(BeNil())
				Expect(*value).To(Equal([]string{"three", "four"}))
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter with interface array and does not contains all string type returns partial value and error", func() {
				value := standardArray.ParseStringArray(3)
				Expect(value).ToNot(BeNil())
				Expect(*value).To(Equal([]string{"five", ""}))
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-string"))
			})
		})

		Context("ParseTime", func() {
			BeforeEach(func() {
				standardArray, _ = parser.NewStandardArray(standardContext, &[]interface{}{
					false,
					"this is a string",
					"2017-12-15T12:34:56Z",
				}, parser.IgnoreNotParsed)
			})

			It("with index parameter less that the first index in the array returns nil", func() {
				Expect(standardArray.ParseTime(-1, time.RFC3339)).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter greater than the last index in the array returns nil", func() {
				Expect(standardArray.ParseTime(len(*standardArray.Array()), time.RFC3339)).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter with different type returns nil and appends an ErrorTypeNotTime", func() {
				Expect(standardArray.ParseTime(0, time.RFC3339)).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-time"))
			})

			It("with index parameter with not parseable table returns nil and appends an ErrorValueTimeNotParsable", func() {
				Expect(standardArray.ParseTime(1, time.RFC3339)).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("value-not-parsable"))
			})

			It("with index parameter with time type returns value", func() {
				value := standardArray.ParseTime(2, time.RFC3339)
				Expect(value).ToNot(BeNil())
				Expect(value.Format(time.RFC3339)).To(Equal("2017-12-15T12:34:56Z"))
				Expect(standardContext.Errors()).To(BeEmpty())
			})
		})

		Context("ParseObject", func() {
			BeforeEach(func() {
				standardArray, _ = parser.NewStandardArray(standardContext, &[]interface{}{
					false,
					map[string]interface{}{
						"1": "2",
					},
				}, parser.IgnoreNotParsed)
			})

			It("with index parameter less that the first index in the array returns nil", func() {
				Expect(standardArray.ParseObject(-1)).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter greater than the last index in the array returns nil", func() {
				Expect(standardArray.ParseObject(len(*standardArray.Array()))).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter with different type returns nil and appends an ErrorTypeNotObject", func() {
				Expect(standardArray.ParseObject(0)).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-object"))
			})

			It("with index parameter with object type returns value", func() {
				value := standardArray.ParseObject(1)
				Expect(value).ToNot(BeNil())
				Expect(*value).To(Equal(map[string]interface{}{"1": "2"}))
				Expect(standardContext.Errors()).To(BeEmpty())
			})
		})

		Context("ParseObjectArray", func() {
			BeforeEach(func() {
				standardArray, _ = parser.NewStandardArray(standardContext, &[]interface{}{
					false,
					[]map[string]interface{}{
						{
							"1": "2",
						},
						{
							"3": "4",
						},
					},
					[]interface{}{
						map[string]interface{}{
							"5": "6",
						},
						map[string]interface{}{
							"7": "8",
						},
					},
					[]interface{}{
						map[string]interface{}{
							"9": "0",
						},
						"not",
					},
				}, parser.IgnoreNotParsed)
			})

			It("with index parameter less that the first index in the array returns nil", func() {
				Expect(standardArray.ParseObjectArray(-1)).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter greater than the last index in the array returns nil", func() {
				Expect(standardArray.ParseObjectArray(len(*standardArray.Array()))).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter with different type returns nil and appends an ErrorTypeNotArray", func() {
				Expect(standardArray.ParseObjectArray(0)).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-array"))
			})

			It("with index parameter with object array type returns value", func() {
				value := standardArray.ParseObjectArray(1)
				Expect(value).ToNot(BeNil())
				Expect(*value).To(Equal([]map[string]interface{}{{"1": "2"}, {"3": "4"}}))
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter with interface array and contains all object type returns value", func() {
				value := standardArray.ParseObjectArray(2)
				Expect(value).ToNot(BeNil())
				Expect(*value).To(Equal([]map[string]interface{}{{"5": "6"}, {"7": "8"}}))
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter with interface array and does not contains all object type returns partial value and error", func() {
				value := standardArray.ParseObjectArray(3)
				Expect(value).ToNot(BeNil())
				Expect(*value).To(Equal([]map[string]interface{}{{"9": "0"}, nil}))
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-object"))
			})
		})

		Context("ParseInterface", func() {
			BeforeEach(func() {
				standardArray, _ = parser.NewStandardArray(standardContext, &[]interface{}{
					false,
					"zombie",
				}, parser.IgnoreNotParsed)
			})

			It("with index parameter less that the first index in the array returns nil", func() {
				Expect(standardArray.ParseInterface(-1)).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter greater than the last index in the array returns nil", func() {
				Expect(standardArray.ParseInterface(len(*standardArray.Array()))).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter with object type returns value", func() {
				value := standardArray.ParseInterface(0)
				Expect(value).ToNot(BeNil())
				Expect(*value).To(BeFalse())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter with object type returns value", func() {
				value := standardArray.ParseInterface(1)
				Expect(value).ToNot(BeNil())
				Expect(*value).To(Equal("zombie"))
				Expect(standardContext.Errors()).To(BeEmpty())
			})
		})

		Context("ParseInterfaceArray", func() {
			BeforeEach(func() {
				standardArray, _ = parser.NewStandardArray(standardContext, &[]interface{}{
					false,
					[]interface{}{
						"1",
						false,
					},
				}, parser.IgnoreNotParsed)
			})

			It("with index parameter less that the first index in the array returns nil", func() {
				Expect(standardArray.ParseInterfaceArray(-1)).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter greater than the last index in the array returns nil", func() {
				Expect(standardArray.ParseInterfaceArray(len(*standardArray.Array()))).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter with different type returns nil and appends an ErrorTypeNotArray", func() {
				Expect(standardArray.ParseInterfaceArray(0)).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-array"))
			})

			It("with index parameter with object array type returns value", func() {
				value := standardArray.ParseInterfaceArray(1)
				Expect(value).ToNot(BeNil())
				Expect(*value).To(Equal([]interface{}{"1", false}))
				Expect(standardContext.Errors()).To(BeEmpty())
			})
		})

		Context("ProcessNotParsed", func() {
			Context("with ParsedPolicy as IgnoreNotParsed", func() {
				BeforeEach(func() {
					standardArray, _ = parser.NewStandardArray(standardContext, &[]interface{}{
						1,
						"two",
						3,
					}, parser.IgnoreNotParsed)
				})

				It("without anything parsed has no errors", func() {
					standardArray.ProcessNotParsed()
					Expect(standardContext.Errors()).To(BeEmpty())
				})
			})

			Context("with ParsedPolicy as WarnLoggerNotParsed", func() {
				BeforeEach(func() {
					standardArray, _ = parser.NewStandardArray(standardContext, &[]interface{}{
						1,
						"two",
						3,
					}, parser.WarnLoggerNotParsed)
				})

				It("without anything parsed has no errors", func() {
					standardArray.ProcessNotParsed()
					Expect(standardContext.Errors()).To(BeEmpty())
				})
			})

			Context("with ParsedPolicy as AppendErrorNotParsed", func() {
				BeforeEach(func() {
					standardArray, _ = parser.NewStandardArray(standardContext, &[]interface{}{
						1,
						"two",
						3,
					}, parser.AppendErrorNotParsed)
				})

				It("without anything parsed appends all unparsed as errors", func() {
					standardArray.ProcessNotParsed()
					Expect(standardContext.Errors()).To(HaveLen(3))
					Expect(standardContext.Errors()[0].Code).To(Equal("not-parsed"))
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/0"))
					Expect(standardContext.Errors()[1].Code).To(Equal("not-parsed"))
					Expect(standardContext.Errors()[1].Source.Pointer).To(Equal("/1"))
					Expect(standardContext.Errors()[2].Code).To(Equal("not-parsed"))
					Expect(standardContext.Errors()[2].Source.Pointer).To(Equal("/2"))
				})

				It("with some items parsed appends all unparsed as errors", func() {
					standardArray.ParseString(1)
					standardArray.ProcessNotParsed()
					Expect(standardContext.Errors()).To(HaveLen(2))
					Expect(standardContext.Errors()[0].Code).To(Equal("not-parsed"))
					Expect(standardContext.Errors()[0].Source.Pointer).To(Equal("/0"))
					Expect(standardContext.Errors()[1].Code).To(Equal("not-parsed"))
					Expect(standardContext.Errors()[1].Source.Pointer).To(Equal("/2"))
				})

				It("with all items parsed has no errors", func() {
					standardArray.ParseInteger(0)
					standardArray.ParseString(1)
					standardArray.ParseInteger(2)
					standardArray.ProcessNotParsed()
					Expect(standardContext.Errors()).To(BeEmpty())
				})
			})
		})

		Context("NewChildObjectParser", func() {
			BeforeEach(func() {
				standardArray, _ = parser.NewStandardArray(standardContext, &[]interface{}{
					false,
					map[string]interface{}{
						"1": "2",
					},
				}, parser.IgnoreNotParsed)
			})

			It("with index parameter less that the first index in the array returns nil", func() {
				objectParser := standardArray.NewChildObjectParser(-1)
				Expect(objectParser).ToNot(BeNil())
				Expect(objectParser.Object()).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter greater than the last index in the array returns nil", func() {
				objectParser := standardArray.NewChildObjectParser(len(*standardArray.Array()))
				Expect(objectParser).ToNot(BeNil())
				Expect(objectParser.Object()).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter with different type returns nil and appends an ErrorTypeNotObject", func() {
				objectParser := standardArray.NewChildObjectParser(0)
				Expect(objectParser).ToNot(BeNil())
				Expect(objectParser.Object()).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-object"))
			})

			It("with index parameter with object type returns value", func() {
				objectParser := standardArray.NewChildObjectParser(1)
				Expect(objectParser).ToNot(BeNil())
				Expect(objectParser.Logger()).ToNot(BeNil())
				Expect(objectParser.Object()).ToNot(BeNil())
				Expect(*objectParser.Object()).To(Equal(map[string]interface{}{"1": "2"}))
				Expect(standardContext.Errors()).To(BeEmpty())
			})
		})

		Context("NewChildArrayParser", func() {
			BeforeEach(func() {
				standardArray, _ = parser.NewStandardArray(standardContext, &[]interface{}{
					false,
					[]interface{}{
						"1",
						false,
					},
				}, parser.IgnoreNotParsed)
			})

			It("with index parameter less that the first index in the array returns nil", func() {
				arrayParser := standardArray.NewChildArrayParser(-1)
				Expect(arrayParser).ToNot(BeNil())
				Expect(arrayParser.Array()).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter greater than the last index in the array returns nil", func() {
				arrayParser := standardArray.NewChildArrayParser(len(*standardArray.Array()))
				Expect(arrayParser).ToNot(BeNil())
				Expect(arrayParser.Array()).To(BeNil())
				Expect(standardContext.Errors()).To(BeEmpty())
			})

			It("with index parameter with different type returns nil and appends an ErrorTypeNotObject", func() {
				arrayParser := standardArray.NewChildArrayParser(0)
				Expect(arrayParser).ToNot(BeNil())
				Expect(arrayParser.Array()).To(BeNil())
				Expect(standardContext.Errors()).To(HaveLen(1))
				Expect(standardContext.Errors()[0].Code).To(Equal("type-not-array"))
			})

			It("with index parameter with object type returns value", func() {
				arrayParser := standardArray.NewChildArrayParser(1)
				Expect(arrayParser).ToNot(BeNil())
				Expect(arrayParser.Logger()).ToNot(BeNil())
				Expect(arrayParser.Array()).ToNot(BeNil())
				Expect(*arrayParser.Array()).To(Equal([]interface{}{"1", false}))
				Expect(standardContext.Errors()).To(BeEmpty())
			})
		})
	})
})
