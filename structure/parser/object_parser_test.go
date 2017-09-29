package parser_test

// import (
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"

// 	"time"

// 	"github.com/tidepool-org/platform/errors"
// 	"github.com/tidepool-org/platform/structure"
// 	structureParser "github.com/tidepool-org/platform/structure/parser"
// 	testStructure "github.com/tidepool-org/platform/structure/test"
// 	"github.com/tidepool-org/platform/test"
// )

// var _ = Describe("Object", func() {
// 	var base *testStructure.Base

// 	BeforeEach(func() {
// 		base = testStructure.NewBase()
// 	})

// 	AfterEach(func() {
// 		base.Expectations()
// 	})

// 	Context("NewObject", func() {
// 		It("returns successfully", func() {
// 			Expect(structureParser.NewObject(nil)).ToNot(BeNil())
// 		})
// 	})

// 	Context("NewObjectParser", func() {
// 		It("returns successfully", func() {
// 			Expect(structureParser.NewObjectParser(base, nil)).ToNot(BeNil())
// 		})
// 	})

// 	Context("with new parser with nil object", func() {
// 		var parser *structureParser.Object

// 		BeforeEach(func() {
// 			parser = structureParser.NewObjectParser(base, nil)
// 			Expect(parser).ToNot(BeNil())
// 		})

// 		Context("Exists", func() {
// 			It("returns false", func() {
// 				Expect(parser.Exists()).To(BeFalse())
// 			})
// 		})

// 		Context("References", func() {
// 			It("returns nil references", func() {
// 				Expect(parser.References()).To(BeNil())
// 			})
// 		})

// 		It("Bool returns nil", func() {
// 			Expect(parser.Bool("0")).To(BeNil())
// 		})

// 		It("Float64 returns nil", func() {
// 			Expect(parser.Float64("2")).To(BeNil())
// 		})

// 		It("Int returns nil", func() {
// 			Expect(parser.Int("3")).To(BeNil())
// 		})

// 		It("String returns nil", func() {
// 			Expect(parser.String("4")).To(BeNil())
// 		})

// 		It("StringArray returns nil", func() {
// 			Expect(parser.StringArray("5")).To(BeNil())
// 		})

// 		It("Time returns nil", func() {
// 			Expect(parser.Time("6", time.RFC3339)).To(BeNil())
// 		})

// 		It("Object returns nil", func() {
// 			Expect(parser.Object("7")).To(BeNil())
// 		})

// 		It("Array returns nil", func() {
// 			Expect(parser.Array("8")).To(BeNil())
// 		})

// 		It("NotParsed does not report an error", func() {
// 			base.ErrorsOutputs = []*errors.Errors{nil}
// 			parser.NotParsed()
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		Context("WithSource", func() {
// 			It("returns new parser", func() {
// 				base.WithSourceOutputs = []structure.Base{testStructure.NewBase()}
// 				withSource := testStructure.NewSource()
// 				result := parser.WithSource(withSource)
// 				Expect(result).ToNot(BeNil())
// 				Expect(result).ToNot(Equal(parser))
// 				Expect(base.WithSourceInputs).To(Equal([]structure.Source{withSource}))
// 			})
// 		})

// 		Context("WithMeta", func() {
// 			It("returns new parser", func() {
// 				base.WithMetaOutputs = []structure.Base{testStructure.NewBase()}
// 				withMeta := test.NewText(1, 128)
// 				result := parser.WithMeta(withMeta)
// 				Expect(result).ToNot(BeNil())
// 				Expect(result).ToNot(Equal(parser))
// 				Expect(base.WithMetaInputs).To(Equal([]interface{}{withMeta}))
// 			})
// 		})

// 		Context("WithReferenceObjectParser", func() {
// 			It("returns new parser", func() {
// 				base.WithReferenceOutputs = []structure.Base{testStructure.NewBase()}
// 				withReference := testStructure.NewReference()
// 				result := parser.WithReferenceObjectParser(withReference)
// 				Expect(result).ToNot(BeNil())
// 				Expect(result).ToNot(Equal(parser))
// 				Expect(result.Exists()).To(BeFalse())
// 				Expect(base.WithReferenceInputs).To(Equal([]string{withReference}))
// 			})
// 		})

// 		Context("WithReferenceArrayParser", func() {
// 			It("returns new parser", func() {
// 				base.WithReferenceOutputs = []structure.Base{testStructure.NewBase()}
// 				withReference := testStructure.NewReference()
// 				result := parser.WithReferenceArrayParser(withReference)
// 				Expect(result).ToNot(BeNil())
// 				Expect(result).ToNot(Equal(parser))
// 				Expect(result.Exists()).To(BeFalse())
// 				Expect(base.WithReferenceInputs).To(Equal([]string{withReference}))
// 			})
// 		})
// 	})

// 	Context("with new parser with valid, empty object", func() {
// 		var parser *structureParser.Object

// 		BeforeEach(func() {
// 			parser = structureParser.NewObjectParser(base, &map[string]interface{}{})
// 			Expect(parser).ToNot(BeNil())
// 		})

// 		Context("Exists", func() {
// 			It("returns true", func() {
// 				Expect(parser.Exists()).To(BeTrue())
// 			})
// 		})

// 		Context("References", func() {
// 			It("returns empty references", func() {
// 				Expect(parser.References()).To(BeEmpty())
// 			})
// 		})

// 		It("NotParsed does not report an error", func() {
// 			base.ErrorsOutputs = []*errors.Errors{nil}
// 			parser.NotParsed()
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})
// 	})

// 	Context("Bool", func() {
// 		var parser *structureParser.Object

// 		BeforeEach(func() {
// 			parser = structureParser.NewObjectParser(base, &map[string]interface{}{
// 				"zero": "not a boolean",
// 				"one":  true,
// 			})
// 			Expect(parser).ToNot(BeNil())
// 		})

// 		It("with key not found in the object returns nil", func() {
// 			Expect(parser.Bool("unknown")).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with key with different type returns nil and reports an ErrorTypeNotBool", func() {
// 			withReference := testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			Expect(parser.Bool("zero")).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTypeNotBool("not a boolean")}))
// 		})

// 		It("with key with boolean type returns value", func() {
// 			value := parser.Bool("one")
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(BeTrue())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})
// 	})

// 	Context("Float64", func() {
// 		var parser *structureParser.Object

// 		BeforeEach(func() {
// 			parser = structureParser.NewObjectParser(base, &map[string]interface{}{
// 				"zero":  false,
// 				"one":   3,
// 				"two":   4.0,
// 				"three": 5.67,
// 			})
// 			Expect(parser).ToNot(BeNil())
// 		})

// 		It("with key not found in the object returns nil", func() {
// 			Expect(parser.Float64("unknown")).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with key with different type returns nil and reports an ErrorTypeNotFloat64", func() {
// 			withReference := testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			Expect(parser.Float64("zero")).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTypeNotFloat64(false)}))
// 		})

// 		It("with key with integer type returns value", func() {
// 			value := parser.Float64("one")
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(BeNumerically("==", 3.))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with key with float type and whole number returns value", func() {
// 			value := parser.Float64("two")
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(BeNumerically("==", 4.0))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with key with float type and not whole number returns value", func() {
// 			value := parser.Float64("three")
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(BeNumerically("==", 5.67))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})
// 	})

// 	Context("Int", func() {
// 		var parser *structureParser.Object

// 		BeforeEach(func() {
// 			parser = structureParser.NewObjectParser(base, &map[string]interface{}{
// 				"zero":  false,
// 				"one":   3,
// 				"two":   4.0,
// 				"three": 5.67,
// 			})
// 			Expect(parser).ToNot(BeNil())
// 		})

// 		It("with key not found in the object returns nil", func() {
// 			Expect(parser.Int("unknown")).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with key with different type returns nil and reports an ErrorTypeNotInt", func() {
// 			withReference := testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			Expect(parser.Int("zero")).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTypeNotInt(false)}))
// 		})

// 		It("with key with integer type returns value", func() {
// 			value := parser.Int("one")
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(BeNumerically("==", 3))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with key with float type and whole number returns value", func() {
// 			value := parser.Int("two")
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(BeNumerically("==", 4))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with key with float type and not whole number returns nil and reports an ErrorTypeNotInt", func() {
// 			withReference := testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			Expect(parser.Int("three")).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTypeNotInt(5.67)}))
// 		})
// 	})

// 	Context("String", func() {
// 		var parser *structureParser.Object

// 		BeforeEach(func() {
// 			parser = structureParser.NewObjectParser(base, &map[string]interface{}{
// 				"zero": false,
// 				"one":  "this is a string",
// 			})
// 			Expect(parser).ToNot(BeNil())
// 		})

// 		It("with key not found in the object returns nil", func() {
// 			Expect(parser.String("unknown")).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with key with different type returns nil and reports an ErrorTypeNotString", func() {
// 			withReference := testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			Expect(parser.String("zero")).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTypeNotString(false)}))
// 		})

// 		It("with key with string type returns value", func() {
// 			value := parser.String("one")
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(Equal("this is a string"))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})
// 	})

// 	Context("StringArray", func() {
// 		var parser *structureParser.Object

// 		BeforeEach(func() {
// 			parser = structureParser.NewObjectParser(base, &map[string]interface{}{
// 				"zero": false,
// 				"one": []string{
// 					"one",
// 					"two",
// 				},
// 				"two": []interface{}{
// 					"three",
// 					"four",
// 				},
// 				"three": []interface{}{
// 					"five",
// 					6,
// 				},
// 			})
// 			Expect(parser).ToNot(BeNil())
// 		})

// 		It("with key not found in the object returns nil", func() {
// 			Expect(parser.StringArray("unknown")).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with key with different type returns nil and reports an ErrorTypeNotArray", func() {
// 			withReference := testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			Expect(parser.StringArray("zero")).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTypeNotArray(false)}))
// 		})

// 		It("with key with string array type returns value", func() {
// 			value := parser.StringArray("one")
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(Equal([]string{"one", "two"}))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with key with interface array and contains all string type returns value", func() {
// 			withReference := testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			value := parser.StringArray("two")
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(Equal([]string{"three", "four"}))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with key with interface array and does not contains all string type returns partial value and error", func() {
// 			withReference := testStructure.NewBase()
// 			withReference2 := testStructure.NewBase()
// 			withReference.WithReferenceOutputs = []structure.Base{withReference2}
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			value := parser.StringArray("three")
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(Equal([]string{"five", ""}))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference2.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTypeNotString(6)}))
// 		})
// 	})

// 	Context("Time", func() {
// 		var now time.Time
// 		var parser *structureParser.Object

// 		BeforeEach(func() {
// 			now = time.Now().Truncate(time.Second)
// 			parser = structureParser.NewObjectParser(base, &map[string]interface{}{
// 				"zero": false,
// 				"one":  "abc",
// 				"two":  now.Format(time.RFC3339),
// 			})
// 			Expect(parser).ToNot(BeNil())
// 		})

// 		It("with key not found in the object returns nil", func() {
// 			Expect(parser.Time("unknown", time.RFC3339)).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with key with different type returns nil and reports an ErrorTypeNotTime", func() {
// 			withReference := testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			Expect(parser.Time("zero", time.RFC3339)).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTypeNotTime(false)}))
// 		})

// 		It("with key with different type returns nil and reports an ErrorTimeNotParsable", func() {
// 			withReference := testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			Expect(parser.Time("one", time.RFC3339)).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTimeNotParsable("abc", time.RFC3339)}))
// 		})

// 		It("with key with string type returns value", func() {
// 			value := parser.Time("two", time.RFC3339)
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(Equal(now))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})
// 	})

// 	Context("Object", func() {
// 		var parser *structureParser.Object

// 		BeforeEach(func() {
// 			parser = structureParser.NewObjectParser(base, &map[string]interface{}{
// 				"zero": false,
// 				"one": map[string]interface{}{
// 					"1": "2",
// 				},
// 			})
// 			Expect(parser).ToNot(BeNil())
// 		})

// 		It("with key not found in the object returns nil", func() {
// 			Expect(parser.Object("unknown")).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with key with different type returns nil and reports an ErrorTypeNotObject", func() {
// 			withReference := testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			Expect(parser.Object("zero")).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTypeNotObject(false)}))
// 		})

// 		It("with key with object type returns value", func() {
// 			value := parser.Object("one")
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(Equal(map[string]interface{}{"1": "2"}))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})
// 	})

// 	Context("Array", func() {
// 		var parser *structureParser.Object

// 		BeforeEach(func() {
// 			parser = structureParser.NewObjectParser(base, &map[string]interface{}{
// 				"zero": false,
// 				"one": []interface{}{
// 					"1",
// 					false,
// 				},
// 			})
// 			Expect(parser).ToNot(BeNil())
// 		})

// 		It("with key not found in the object returns nil", func() {
// 			Expect(parser.Array("unknown")).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with key with different type returns nil and reports an ErrorTypeNotArray", func() {
// 			withReference := testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			Expect(parser.Array("zero")).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTypeNotArray(false)}))
// 		})

// 		It("with key with object array type returns value", func() {
// 			value := parser.Array("one")
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(Equal([]interface{}{"1", false}))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})
// 	})

// 	Context("NotParsed", func() {
// 		var parser *structureParser.Object

// 		BeforeEach(func() {
// 			parser = structureParser.NewObjectParser(base, &map[string]interface{}{
// 				"zero": 1,
// 				"one":  "two",
// 				"two":  3,
// 			})
// 			Expect(parser).ToNot(BeNil())
// 		})

// 		It("without anything parsed reports all unparsed as errors", func() {
// 			withReference := testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference, withReference, withReference}
// 			base.ErrorsOutputs = []*errors.Errors{nil}
// 			parser.NotParsed()
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{
// 				structureParser.ErrorNotParsed(),
// 				structureParser.ErrorNotParsed(),
// 				structureParser.ErrorNotParsed(),
// 			}))
// 		})

// 		It("with some items parsed reports all unparsed as errors", func() {
// 			withReference := testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference, withReference}
// 			base.ErrorsOutputs = []*errors.Errors{nil}
// 			parser.String("one")
// 			parser.NotParsed()
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{
// 				structureParser.ErrorNotParsed(),
// 				structureParser.ErrorNotParsed(),
// 			}))
// 		})

// 		It("with all items parsed has no errors", func() {
// 			base.ErrorsOutputs = []*errors.Errors{nil}
// 			parser.Int("zero")
// 			parser.String("one")
// 			parser.Int("two")
// 			parser.NotParsed()
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})
// 	})

// 	Context("WithReferenceObjectParser", func() {
// 		var withReference *testStructure.Base
// 		var parser *structureParser.Object

// 		BeforeEach(func() {
// 			withReference = testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			parser = structureParser.NewObjectParser(base, &map[string]interface{}{
// 				"zero": false,
// 				"one": map[string]interface{}{
// 					"1": "2",
// 				},
// 			})
// 			Expect(parser).ToNot(BeNil())
// 		})

// 		It("with key not found in the object returns nil", func() {
// 			objectParser := parser.WithReferenceObjectParser("unknown")
// 			Expect(objectParser).ToNot(BeNil())
// 			Expect(objectParser.Exists()).To(BeFalse())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with key with different type returns nil and reports an ErrorTypeNotObject", func() {
// 			base.WithReferenceOutputs = []structure.Base{withReference, withReference}
// 			objectParser := parser.WithReferenceObjectParser("zero")
// 			Expect(objectParser).ToNot(BeNil())
// 			Expect(objectParser.Exists()).To(BeFalse())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTypeNotObject(false)}))
// 		})

// 		It("with key with object type returns value", func() {
// 			objectParser := parser.WithReferenceObjectParser("one")
// 			Expect(objectParser).ToNot(BeNil())
// 			Expect(objectParser.Exists()).To(BeTrue())
// 			Expect(objectParser.References()).To(Equal([]string{"1"}))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})
// 	})

// 	Context("WithReferenceArrayParser", func() {
// 		var withReference *testStructure.Base
// 		var parser *structureParser.Object

// 		BeforeEach(func() {
// 			withReference = testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			parser = structureParser.NewObjectParser(base, &map[string]interface{}{
// 				"zero": false,
// 				"one": []interface{}{
// 					"1",
// 					false,
// 				},
// 			})
// 			Expect(parser).ToNot(BeNil())
// 		})

// 		It("with key not found in the object returns nil", func() {
// 			arrayParser := parser.WithReferenceArrayParser("unknown")
// 			Expect(arrayParser).ToNot(BeNil())
// 			Expect(arrayParser.Exists()).To(BeFalse())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with key with different type returns nil and reports an ErrorTypeNotObject", func() {
// 			base.WithReferenceOutputs = []structure.Base{withReference, withReference}
// 			arrayParser := parser.WithReferenceArrayParser("zero")
// 			Expect(arrayParser).ToNot(BeNil())
// 			Expect(arrayParser.Exists()).To(BeFalse())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTypeNotArray(false)}))
// 		})

// 		It("with key with object type returns value", func() {
// 			arrayParser := parser.WithReferenceArrayParser("one")
// 			Expect(arrayParser).ToNot(BeNil())
// 			Expect(arrayParser.Exists()).To(BeTrue())
// 			Expect(arrayParser.References()).To(Equal([]int{0, 1}))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})
// 	})
// })
