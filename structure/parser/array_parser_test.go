package parser_test

// import (
// 	"math/rand"
// 	"strconv"

// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"

// 	"time"

// 	"github.com/tidepool-org/platform/errors"
// 	"github.com/tidepool-org/platform/structure"
// 	structureParser "github.com/tidepool-org/platform/structure/parser"
// 	testStructure "github.com/tidepool-org/platform/structure/test"
// 	"github.com/tidepool-org/platform/test"
// )

// var _ = Describe("Array", func() {
// 	var base *testStructure.Base

// 	BeforeEach(func() {
// 		base = testStructure.NewBase()
// 	})

// 	AfterEach(func() {
// 		base.Expectations()
// 	})

// 	Context("NewArray", func() {
// 		It("returns successfully", func() {
// 			Expect(structureParser.NewArray(nil)).ToNot(BeNil())
// 		})
// 	})

// 	Context("NewArrayParser", func() {
// 		It("returns successfully", func() {
// 			Expect(structureParser.NewArrayParser(base, nil)).ToNot(BeNil())
// 		})
// 	})

// 	Context("with new parser with nil array", func() {
// 		var parser *structureParser.Array

// 		BeforeEach(func() {
// 			parser = structureParser.NewArrayParser(base, nil)
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
// 			Expect(parser.Bool(0)).To(BeNil())
// 		})

// 		It("Float64 returns nil", func() {
// 			Expect(parser.Float64(2)).To(BeNil())
// 		})

// 		It("Int returns nil", func() {
// 			Expect(parser.Int(3)).To(BeNil())
// 		})

// 		It("String returns nil", func() {
// 			Expect(parser.String(4)).To(BeNil())
// 		})

// 		It("StringArray returns nil", func() {
// 			Expect(parser.StringArray(5)).To(BeNil())
// 		})

// 		It("Time returns nil", func() {
// 			Expect(parser.Time(6, time.RFC3339)).To(BeNil())
// 		})

// 		It("Object returns nil", func() {
// 			Expect(parser.Object(7)).To(BeNil())
// 		})

// 		It("Array returns nil", func() {
// 			Expect(parser.Array(8)).To(BeNil())
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
// 				withReference := rand.Int()
// 				result := parser.WithReferenceObjectParser(withReference)
// 				Expect(result).ToNot(BeNil())
// 				Expect(result).ToNot(Equal(parser))
// 				Expect(result.Exists()).To(BeFalse())
// 				Expect(base.WithReferenceInputs).To(Equal([]string{strconv.Itoa(withReference)}))
// 			})
// 		})

// 		Context("WithReferenceArrayParser", func() {
// 			It("returns new parser", func() {
// 				base.WithReferenceOutputs = []structure.Base{testStructure.NewBase()}
// 				withReference := rand.Int()
// 				result := parser.WithReferenceArrayParser(withReference)
// 				Expect(result).ToNot(BeNil())
// 				Expect(result).ToNot(Equal(parser))
// 				Expect(result.Exists()).To(BeFalse())
// 				Expect(base.WithReferenceInputs).To(Equal([]string{strconv.Itoa(withReference)}))
// 			})
// 		})
// 	})

// 	Context("with new parser with valid, empty array", func() {
// 		var parser *structureParser.Array

// 		BeforeEach(func() {
// 			parser = structureParser.NewArrayParser(base, &[]interface{}{})
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
// 		var parser *structureParser.Array

// 		BeforeEach(func() {
// 			parser = structureParser.NewArrayParser(base, &[]interface{}{
// 				"not a boolean",
// 				true,
// 			})
// 			Expect(parser).ToNot(BeNil())
// 		})

// 		It("with index parameter less that the first index in the array returns nil", func() {
// 			Expect(parser.Bool(-1)).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter greater than the last index in the array returns nil", func() {
// 			Expect(parser.Bool(len(parser.References()))).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter with different type returns nil and reports an ErrorTypeNotBool", func() {
// 			withReference := testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			Expect(parser.Bool(0)).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTypeNotBool("not a boolean")}))
// 		})

// 		It("with index parameter with boolean type returns value", func() {
// 			value := parser.Bool(1)
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(BeTrue())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})
// 	})

// 	Context("Float64", func() {
// 		var parser *structureParser.Array

// 		BeforeEach(func() {
// 			parser = structureParser.NewArrayParser(base, &[]interface{}{
// 				false,
// 				3,
// 				4.0,
// 				5.67,
// 			})
// 			Expect(parser).ToNot(BeNil())
// 		})

// 		It("with index parameter less that the first index in the array returns nil", func() {
// 			Expect(parser.Float64(-1)).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter greater than the last index in the array returns nil", func() {
// 			Expect(parser.Float64(len(parser.References()))).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter with different type returns nil and reports an ErrorTypeNotFloat64", func() {
// 			withReference := testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			Expect(parser.Float64(0)).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTypeNotFloat64(false)}))
// 		})

// 		It("with index parameter with integer type returns value", func() {
// 			value := parser.Float64(1)
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(BeNumerically("==", 3.))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter with float type and whole number returns value", func() {
// 			value := parser.Float64(2)
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(BeNumerically("==", 4.0))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter with float type and not whole number returns value", func() {
// 			value := parser.Float64(3)
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(BeNumerically("==", 5.67))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})
// 	})

// 	Context("Int", func() {
// 		var parser *structureParser.Array

// 		BeforeEach(func() {
// 			parser = structureParser.NewArrayParser(base, &[]interface{}{
// 				false,
// 				3,
// 				4.0,
// 				5.67,
// 			})
// 			Expect(parser).ToNot(BeNil())
// 		})

// 		It("with index parameter less that the first index in the array returns nil", func() {
// 			Expect(parser.Int(-1)).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter greater than the last index in the array returns nil", func() {
// 			Expect(parser.Int(len(parser.References()))).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter with different type returns nil and reports an ErrorTypeNotInt", func() {
// 			withReference := testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			Expect(parser.Int(0)).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTypeNotInt(false)}))
// 		})

// 		It("with index parameter with integer type returns value", func() {
// 			value := parser.Int(1)
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(BeNumerically("==", 3))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter with float type and whole number returns value", func() {
// 			value := parser.Int(2)
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(BeNumerically("==", 4))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter with float type and not whole number returns nil and reports an ErrorTypeNotInt", func() {
// 			withReference := testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			Expect(parser.Int(3)).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTypeNotInt(5.67)}))
// 		})
// 	})

// 	Context("String", func() {
// 		var parser *structureParser.Array

// 		BeforeEach(func() {
// 			parser = structureParser.NewArrayParser(base, &[]interface{}{
// 				false,
// 				"this is a string",
// 			})
// 			Expect(parser).ToNot(BeNil())
// 		})

// 		It("with index parameter less that the first index in the array returns nil", func() {
// 			Expect(parser.String(-1)).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter greater than the last index in the array returns nil", func() {
// 			Expect(parser.String(len(parser.References()))).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter with different type returns nil and reports an ErrorTypeNotString", func() {
// 			withReference := testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			Expect(parser.String(0)).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTypeNotString(false)}))
// 		})

// 		It("with index parameter with string type returns value", func() {
// 			value := parser.String(1)
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(Equal("this is a string"))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})
// 	})

// 	Context("StringArray", func() {
// 		var parser *structureParser.Array

// 		BeforeEach(func() {
// 			parser = structureParser.NewArrayParser(base, &[]interface{}{
// 				false,
// 				[]string{
// 					"one",
// 					"two",
// 				},
// 				[]interface{}{
// 					"three",
// 					"four",
// 				},
// 				[]interface{}{
// 					"five",
// 					6,
// 				},
// 			})
// 			Expect(parser).ToNot(BeNil())
// 		})

// 		It("with index parameter less that the first index in the array returns nil", func() {
// 			Expect(parser.StringArray(-1)).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter greater than the last index in the array returns nil", func() {
// 			Expect(parser.StringArray(len(parser.References()))).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter with different type returns nil and reports an ErrorTypeNotArray", func() {
// 			withReference := testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			Expect(parser.StringArray(0)).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTypeNotArray(false)}))
// 		})

// 		It("with index parameter with string array type returns value", func() {
// 			value := parser.StringArray(1)
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(Equal([]string{"one", "two"}))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter with interface array and contains all string type returns value", func() {
// 			withReference := testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			value := parser.StringArray(2)
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(Equal([]string{"three", "four"}))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter with interface array and does not contains all string type returns partial value and error", func() {
// 			withReference := testStructure.NewBase()
// 			withReference2 := testStructure.NewBase()
// 			withReference.WithReferenceOutputs = []structure.Base{withReference2}
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			value := parser.StringArray(3)
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(Equal([]string{"five", ""}))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference2.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTypeNotString(6)}))
// 		})
// 	})

// 	Context("Time", func() {
// 		var now time.Time
// 		var parser *structureParser.Array

// 		BeforeEach(func() {
// 			now = time.Now().Truncate(time.Second)
// 			parser = structureParser.NewArrayParser(base, &[]interface{}{
// 				false,
// 				"abc",
// 				now.Format(time.RFC3339),
// 			})
// 			Expect(parser).ToNot(BeNil())
// 		})

// 		It("with index parameter less that the first index in the array returns nil", func() {
// 			Expect(parser.Time(-1, time.RFC3339)).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter greater than the last index in the array returns nil", func() {
// 			Expect(parser.Time(len(parser.References()), time.RFC3339)).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter with different type returns nil and reports an ErrorTypeNotTime", func() {
// 			withReference := testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			Expect(parser.Time(0, time.RFC3339)).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTypeNotTime(false)}))
// 		})

// 		It("with index parameter with different type returns nil and reports an ErrorTimeNotParsable", func() {
// 			withReference := testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			Expect(parser.Time(1, time.RFC3339)).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTimeNotParsable("abc", time.RFC3339)}))
// 		})

// 		It("with index parameter with string type returns value", func() {
// 			value := parser.Time(2, time.RFC3339)
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(Equal(now))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})
// 	})

// 	Context("Object", func() {
// 		var parser *structureParser.Array

// 		BeforeEach(func() {
// 			parser = structureParser.NewArrayParser(base, &[]interface{}{
// 				false,
// 				map[string]interface{}{
// 					"1": "2",
// 				},
// 			})
// 			Expect(parser).ToNot(BeNil())
// 		})

// 		It("with index parameter less that the first index in the array returns nil", func() {
// 			Expect(parser.Object(-1)).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter greater than the last index in the array returns nil", func() {
// 			Expect(parser.Object(len(parser.References()))).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter with different type returns nil and reports an ErrorTypeNotObject", func() {
// 			withReference := testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			Expect(parser.Object(0)).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTypeNotObject(false)}))
// 		})

// 		It("with index parameter with object type returns value", func() {
// 			value := parser.Object(1)
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(Equal(map[string]interface{}{"1": "2"}))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})
// 	})

// 	Context("Array", func() {
// 		var parser *structureParser.Array

// 		BeforeEach(func() {
// 			parser = structureParser.NewArrayParser(base, &[]interface{}{
// 				false,
// 				[]interface{}{
// 					"1",
// 					false,
// 				},
// 			})
// 			Expect(parser).ToNot(BeNil())
// 		})

// 		It("with index parameter less that the first index in the array returns nil", func() {
// 			Expect(parser.Array(-1)).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter greater than the last index in the array returns nil", func() {
// 			Expect(parser.Array(len(parser.References()))).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter with different type returns nil and reports an ErrorTypeNotArray", func() {
// 			withReference := testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			Expect(parser.Array(0)).To(BeNil())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTypeNotArray(false)}))
// 		})

// 		It("with index parameter with object array type returns value", func() {
// 			value := parser.Array(1)
// 			Expect(value).ToNot(BeNil())
// 			Expect(*value).To(Equal([]interface{}{"1", false}))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})
// 	})

// 	Context("NotParsed", func() {
// 		var parser *structureParser.Array

// 		BeforeEach(func() {
// 			parser = structureParser.NewArrayParser(base, &[]interface{}{
// 				1,
// 				"two",
// 				3,
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
// 			parser.String(1)
// 			parser.NotParsed()
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{
// 				structureParser.ErrorNotParsed(),
// 				structureParser.ErrorNotParsed(),
// 			}))
// 		})

// 		It("with all items parsed has no errors", func() {
// 			base.ErrorsOutputs = []*errors.Errors{nil}
// 			parser.Int(0)
// 			parser.String(1)
// 			parser.Int(2)
// 			parser.NotParsed()
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})
// 	})

// 	Context("WithReferenceObjectParser", func() {
// 		var withReference *testStructure.Base
// 		var parser *structureParser.Array

// 		BeforeEach(func() {
// 			withReference = testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			parser = structureParser.NewArrayParser(base, &[]interface{}{
// 				false,
// 				map[string]interface{}{
// 					"1": "2",
// 				},
// 			})
// 			Expect(parser).ToNot(BeNil())
// 		})

// 		It("with index parameter less that the first index in the array returns nil", func() {
// 			objectParser := parser.WithReferenceObjectParser(-1)
// 			Expect(objectParser).ToNot(BeNil())
// 			Expect(objectParser.Exists()).To(BeFalse())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter greater than the last index in the array returns nil", func() {
// 			objectParser := parser.WithReferenceObjectParser(len(parser.References()))
// 			Expect(objectParser).ToNot(BeNil())
// 			Expect(objectParser.Exists()).To(BeFalse())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter with different type returns nil and reports an ErrorTypeNotObject", func() {
// 			base.WithReferenceOutputs = []structure.Base{withReference, withReference}
// 			objectParser := parser.WithReferenceObjectParser(0)
// 			Expect(objectParser).ToNot(BeNil())
// 			Expect(objectParser.Exists()).To(BeFalse())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTypeNotObject(false)}))
// 		})

// 		It("with index parameter with object type returns value", func() {
// 			objectParser := parser.WithReferenceObjectParser(1)
// 			Expect(objectParser).ToNot(BeNil())
// 			Expect(objectParser.Exists()).To(BeTrue())
// 			Expect(objectParser.References()).To(Equal([]string{"1"}))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})
// 	})

// 	Context("WithReferenceArrayParser", func() {
// 		var withReference *testStructure.Base
// 		var parser *structureParser.Array

// 		BeforeEach(func() {
// 			withReference = testStructure.NewBase()
// 			base.WithReferenceOutputs = []structure.Base{withReference}
// 			parser = structureParser.NewArrayParser(base, &[]interface{}{
// 				false,
// 				[]interface{}{
// 					"1",
// 					false,
// 				},
// 			})
// 			Expect(parser).ToNot(BeNil())
// 		})

// 		It("with index parameter less that the first index in the array returns nil", func() {
// 			arrayParser := parser.WithReferenceArrayParser(-1)
// 			Expect(arrayParser).ToNot(BeNil())
// 			Expect(arrayParser.Exists()).To(BeFalse())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter greater than the last index in the array returns nil", func() {
// 			arrayParser := parser.WithReferenceArrayParser(len(parser.References()))
// 			Expect(arrayParser).ToNot(BeNil())
// 			Expect(arrayParser.Exists()).To(BeFalse())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})

// 		It("with index parameter with different type returns nil and reports an ErrorTypeNotObject", func() {
// 			base.WithReferenceOutputs = []structure.Base{withReference, withReference}
// 			arrayParser := parser.WithReferenceArrayParser(0)
// 			Expect(arrayParser).ToNot(BeNil())
// 			Expect(arrayParser.Exists()).To(BeFalse())
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 			Expect(withReference.ReportErrorInputs).To(Equal([]*errors.Error{structureParser.ErrorTypeNotArray(false)}))
// 		})

// 		It("with index parameter with object type returns value", func() {
// 			arrayParser := parser.WithReferenceArrayParser(1)
// 			Expect(arrayParser).ToNot(BeNil())
// 			Expect(arrayParser.Exists()).To(BeTrue())
// 			Expect(arrayParser.References()).To(Equal([]int{0, 1}))
// 			Expect(base.ReportErrorInputs).To(BeEmpty())
// 		})
// 	})
// })
