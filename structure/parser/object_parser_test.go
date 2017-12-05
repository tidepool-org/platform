package parser_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	"github.com/tidepool-org/platform/errors"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	testStructure "github.com/tidepool-org/platform/structure/test"
)

var _ = Describe("Object", func() {
	var base *structureBase.Base

	BeforeEach(func() {
		base = structureBase.New()
	})

	Context("NewObject", func() {
		It("returns successfully", func() {
			Expect(structureParser.NewObject(nil)).ToNot(BeNil())
		})
	})

	Context("NewObjectParser", func() {
		It("returns successfully", func() {
			Expect(structureParser.NewObjectParser(base, nil)).ToNot(BeNil())
		})
	})

	Context("with new parser with nil object", func() {
		var parser *structureParser.Object

		BeforeEach(func() {
			parser = structureParser.NewObjectParser(base, nil)
			Expect(parser).ToNot(BeNil())
		})

		Context("Error", func() {
			It("returns the error from the base", func() {
				err := testErrors.NewError()
				base.ReportError(err)
				Expect(parser.Error()).To(Equal(errors.Normalize(err)))
			})
		})

		Context("ReportError", func() {
			It("reports the error to the base", func() {
				err := testErrors.NewError()
				parser.ReportError(err)
				Expect(base.Error()).To(Equal(errors.Normalize(err)))
			})
		})

		Context("Exists", func() {
			It("returns false", func() {
				Expect(parser.Exists()).To(BeFalse())
			})
		})

		Context("Parse", func() {
			var objectParsable *testStructure.ObjectParsable

			BeforeEach(func() {
				objectParsable = testStructure.NewObjectParsable()
			})

			AfterEach(func() {
				objectParsable.Expectations()
			})

			It("invokes parse and returns current errors", func() {
				err := testErrors.NewError()
				base.ReportError(err)
				Expect(parser.Parse(objectParsable)).To(Equal(errors.Normalize(err)))
				Expect(objectParsable.ParseInputs).To(Equal([]structure.ObjectParser{parser}))
			})
		})

		Context("References", func() {
			It("returns nil references", func() {
				Expect(parser.References()).To(BeNil())
			})
		})

		Context("ReferenceExists", func() {
			It("returns false", func() {
				Expect(parser.ReferenceExists("0")).To(BeFalse())
			})
		})

		It("Bool returns nil", func() {
			Expect(parser.Bool("1")).To(BeNil())
		})

		It("Float64 returns nil", func() {
			Expect(parser.Float64("2")).To(BeNil())
		})

		It("Int returns nil", func() {
			Expect(parser.Int("3")).To(BeNil())
		})

		It("String returns nil", func() {
			Expect(parser.String("4")).To(BeNil())
		})

		It("StringArray returns nil", func() {
			Expect(parser.StringArray("5")).To(BeNil())
		})

		It("Time returns nil", func() {
			Expect(parser.Time("6", time.RFC3339)).To(BeNil())
		})

		It("Object returns nil", func() {
			Expect(parser.Object("7")).To(BeNil())
		})

		It("Array returns nil", func() {
			Expect(parser.Array("8")).To(BeNil())
		})

		It("Interface returns nil", func() {
			Expect(parser.Interface("9")).To(BeNil())
		})

		It("NotParsed only returns existing errors", func() {
			err := testErrors.NewError()
			base.ReportError(err)
			Expect(parser.NotParsed()).To(Equal(errors.Normalize(err)))
		})

		Context("WithSource", func() {
			It("returns new parser", func() {
				src := testStructure.NewSource()
				result := parser.WithSource(src)
				Expect(result).ToNot(BeNil())
				Expect(result).ToNot(Equal(parser))
			})
		})

		Context("WithMeta", func() {
			It("returns new parser", func() {
				result := parser.WithMeta(testErrors.NewMeta())
				Expect(result).ToNot(BeNil())
				Expect(result).ToNot(Equal(parser))
			})
		})

		Context("WithReferenceObjectParser", func() {
			It("without source returns new parser", func() {
				reference := testStructure.NewReference()
				result := parser.WithReferenceObjectParser(reference)
				Expect(result).ToNot(BeNil())
				Expect(result).ToNot(BeIdenticalTo(parser))
				Expect(result).To(Equal(parser))
				Expect(result.Exists()).To(BeFalse())
			})

			It("with source returns new parser", func() {
				src := testStructure.NewSource()
				src.WithReferenceOutputs = []structure.Source{testStructure.NewSource()}
				reference := testStructure.NewReference()
				resultWithSource := parser.WithSource(src)
				resultWithReference := parser.WithReferenceObjectParser(reference)
				Expect(resultWithReference).ToNot(BeNil())
				Expect(resultWithReference).ToNot(Equal(resultWithSource))
				Expect(resultWithReference.Exists()).To(BeFalse())
			})
		})

		Context("WithReferenceArrayParser", func() {
			It("returns new parser", func() {
				reference := testStructure.NewReference()
				result := parser.WithReferenceArrayParser(reference)
				Expect(result).ToNot(BeNil())
				Expect(result).ToNot(Equal(parser))
				Expect(result.Exists()).To(BeFalse())
			})
		})
	})

	Context("with new parser with valid, empty object", func() {
		var parser *structureParser.Object

		BeforeEach(func() {
			parser = structureParser.NewObjectParser(base, &map[string]interface{}{})
			Expect(parser).ToNot(BeNil())
		})

		Context("Error", func() {
			It("returns the error from the base", func() {
				err := testErrors.NewError()
				base.ReportError(err)
				Expect(parser.Error()).To(Equal(errors.Normalize(err)))
			})
		})

		Context("Exists", func() {
			It("returns true", func() {
				Expect(parser.Exists()).To(BeTrue())
			})
		})

		Context("Parse", func() {
			var objectParsable *testStructure.ObjectParsable

			BeforeEach(func() {
				objectParsable = testStructure.NewObjectParsable()
			})

			AfterEach(func() {
				objectParsable.Expectations()
			})

			It("invokes parse and returns current errors", func() {
				err := testErrors.NewError()
				base.ReportError(err)
				Expect(parser.Parse(objectParsable)).To(Equal(errors.Normalize(err)))
				Expect(objectParsable.ParseInputs).To(Equal([]structure.ObjectParser{parser}))
			})
		})

		Context("References", func() {
			It("returns empty references", func() {
				Expect(parser.References()).To(BeEmpty())
			})
		})
	})

	Context("References", func() {
		var parser *structureParser.Object

		BeforeEach(func() {
			parser = structureParser.NewObjectParser(base, &map[string]interface{}{
				"zero": "exists",
				"one":  "exists",
			})
			Expect(parser).ToNot(BeNil())
		})

		It("returns the array of references", func() {
			Expect(parser.References()).To(ConsistOf("zero", "one"))
		})
	})

	Context("ReferenceExists", func() {
		var parser *structureParser.Object

		BeforeEach(func() {
			parser = structureParser.NewObjectParser(base, &map[string]interface{}{
				"zero": "exists",
			})
			Expect(parser).ToNot(BeNil())
		})

		It("returns true if reference exists", func() {
			Expect(parser.ReferenceExists("zero")).To(BeTrue())
		})

		It("returns false if reference does not exist", func() {
			Expect(parser.ReferenceExists("one")).To(BeFalse())
		})
	})

	Context("Bool", func() {
		var parser *structureParser.Object

		BeforeEach(func() {
			parser = structureParser.NewObjectParser(base, &map[string]interface{}{
				"zero": "not a boolean",
				"one":  true,
			})
			Expect(parser).ToNot(BeNil())
		})

		It("with key not found in the object returns nil", func() {
			Expect(parser.Bool("unknown")).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with key with different type returns nil and reports an ErrorTypeNotBool", func() {
			Expect(parser.Bool("zero")).To(BeNil())
			Expect(base.Error()).To(HaveOccurred())
			Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureParser.ErrorTypeNotBool("not a boolean"))))
		})

		It("with key with boolean type returns value", func() {
			value := parser.Bool("one")
			Expect(value).ToNot(BeNil())
			Expect(*value).To(BeTrue())
			Expect(base.Error()).ToNot(HaveOccurred())
		})
	})

	Context("Float64", func() {
		var parser *structureParser.Object

		BeforeEach(func() {
			parser = structureParser.NewObjectParser(base, &map[string]interface{}{
				"zero":  false,
				"one":   3,
				"two":   4.0,
				"three": 5.67,
			})
			Expect(parser).ToNot(BeNil())
		})

		It("with key not found in the object returns nil", func() {
			Expect(parser.Float64("unknown")).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with key with different type returns nil and reports an ErrorCodeTypeNotFloat64", func() {
			Expect(parser.Float64("zero")).To(BeNil())
			Expect(base.Error()).To(HaveOccurred())
			Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureParser.ErrorTypeNotFloat64(false))))
		})

		It("with key with integer type returns value", func() {
			value := parser.Float64("one")
			Expect(value).ToNot(BeNil())
			Expect(*value).To(BeNumerically("==", 3.))
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with key with float type and whole number returns value", func() {
			value := parser.Float64("two")
			Expect(value).ToNot(BeNil())
			Expect(*value).To(BeNumerically("==", 4.0))
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with key with float type and not whole number returns value", func() {
			value := parser.Float64("three")
			Expect(value).ToNot(BeNil())
			Expect(*value).To(BeNumerically("==", 5.67))
			Expect(base.Error()).ToNot(HaveOccurred())
		})
	})

	Context("Int", func() {
		var parser *structureParser.Object

		BeforeEach(func() {
			parser = structureParser.NewObjectParser(base, &map[string]interface{}{
				"zero":  false,
				"one":   3,
				"two":   4.0,
				"three": 5.67,
			})
			Expect(parser).ToNot(BeNil())
		})

		It("with key not found in the object returns nil", func() {
			Expect(parser.Int("unknown")).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with key with different type returns nil and reports an ErrorCodeTypeNotInt", func() {
			Expect(parser.Int("zero")).To(BeNil())
			Expect(base.Error()).To(HaveOccurred())
			Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureParser.ErrorTypeNotInt(false))))
		})

		It("with key with integer type returns value", func() {
			value := parser.Int("one")
			Expect(value).ToNot(BeNil())
			Expect(*value).To(BeNumerically("==", 3))
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with key with float type and whole number returns value", func() {
			value := parser.Int("two")
			Expect(value).ToNot(BeNil())
			Expect(*value).To(BeNumerically("==", 4))
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with key with float type and not whole number returns nil and reports an ErrorCodeTypeNotInt", func() {
			Expect(parser.Int("three")).To(BeNil())
			Expect(base.Error()).To(HaveOccurred())
			Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureParser.ErrorTypeNotInt(5.67))))
		})
	})

	Context("String", func() {
		var parser *structureParser.Object

		BeforeEach(func() {
			parser = structureParser.NewObjectParser(base, &map[string]interface{}{
				"zero": false,
				"one":  "this is a string",
			})
			Expect(parser).ToNot(BeNil())
		})

		It("with key not found in the object returns nil", func() {
			Expect(parser.String("unknown")).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with key with different type returns nil and reports an ErrorCodeTypeNotString", func() {
			Expect(parser.String("zero")).To(BeNil())
			Expect(base.Error()).To(HaveOccurred())
			Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureParser.ErrorTypeNotString(false))))
		})

		It("with key with string type returns value", func() {
			value := parser.String("one")
			Expect(value).ToNot(BeNil())
			Expect(*value).To(Equal("this is a string"))
			Expect(base.Error()).ToNot(HaveOccurred())
		})
	})

	Context("StringArray", func() {
		var parser *structureParser.Object

		BeforeEach(func() {
			parser = structureParser.NewObjectParser(base, &map[string]interface{}{
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
			})
			Expect(parser).ToNot(BeNil())
		})

		It("with key not found in the object returns nil", func() {
			Expect(parser.StringArray("unknown")).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with key with different type returns nil and reports an ErrorCodeTypeNotArray", func() {
			Expect(parser.StringArray("zero")).To(BeNil())
			Expect(base.Error()).To(HaveOccurred())
			Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureParser.ErrorTypeNotArray(false))))
		})

		It("with key with string array type returns value", func() {
			value := parser.StringArray("one")
			Expect(value).ToNot(BeNil())
			Expect(*value).To(Equal([]string{"one", "two"}))
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with key with interface array and contains all string type returns value", func() {
			value := parser.StringArray("two")
			Expect(value).ToNot(BeNil())
			Expect(*value).To(Equal([]string{"three", "four"}))
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with key with interface array and does not contains all string type returns partial value and ErrorCodeTypeNotString", func() {
			value := parser.StringArray("three")
			Expect(value).ToNot(BeNil())
			Expect(*value).To(Equal([]string{"five", ""}))
			Expect(base.Error()).To(HaveOccurred())
			Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureParser.ErrorTypeNotString(6))))
		})
	})

	Context("Time", func() {
		var now time.Time
		var parser *structureParser.Object

		BeforeEach(func() {
			now = time.Now().Truncate(time.Second)
			parser = structureParser.NewObjectParser(base, &map[string]interface{}{
				"zero": false,
				"one":  "abc",
				"two":  now.Format(time.RFC3339),
			})
			Expect(parser).ToNot(BeNil())
		})

		It("with key not found in the object returns nil", func() {
			Expect(parser.Time("unknown", time.RFC3339)).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with key with different type returns nil and reports an ErrorCodeTypeNotTime", func() {
			Expect(parser.Time("zero", time.RFC3339)).To(BeNil())
			Expect(base.Error()).To(HaveOccurred())
			Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureParser.ErrorTypeNotTime(false))))
		})

		It("with key with different type returns nil and reports an ErrorCodeTimeNotParsable", func() {
			Expect(parser.Time("one", time.RFC3339)).To(BeNil())
			Expect(base.Error()).To(HaveOccurred())
			Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureParser.ErrorTimeNotParsable("abc", time.RFC3339))))
		})

		It("with key with string type returns value", func() {
			value := parser.Time("two", time.RFC3339)
			Expect(value).ToNot(BeNil())
			Expect(*value).To(BeTemporally("==", now))
			Expect(base.Error()).ToNot(HaveOccurred())
		})
	})

	Context("Object", func() {
		var parser *structureParser.Object

		BeforeEach(func() {
			parser = structureParser.NewObjectParser(base, &map[string]interface{}{
				"zero": false,
				"one": map[string]interface{}{
					"1": "2",
				},
			})
			Expect(parser).ToNot(BeNil())
		})

		It("with key not found in the object returns nil", func() {
			Expect(parser.Object("unknown")).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with key with different type returns nil and reports an ErrorCodeTypeNotObject", func() {
			Expect(parser.Object("zero")).To(BeNil())
			Expect(base.Error()).To(HaveOccurred())
			Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureParser.ErrorTypeNotObject(false))))
		})

		It("with key with object type returns value", func() {
			value := parser.Object("one")
			Expect(value).ToNot(BeNil())
			Expect(*value).To(Equal(map[string]interface{}{"1": "2"}))
			Expect(base.Error()).ToNot(HaveOccurred())
		})
	})

	Context("Array", func() {
		var parser *structureParser.Object

		BeforeEach(func() {
			parser = structureParser.NewObjectParser(base, &map[string]interface{}{
				"zero": false,
				"one": []interface{}{
					"1",
					false,
				},
			})
			Expect(parser).ToNot(BeNil())
		})

		It("with key not found in the object returns nil", func() {
			Expect(parser.Array("unknown")).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with key with different type returns nil and reports an ErrorCodeTypeNotArray", func() {
			Expect(parser.Array("zero")).To(BeNil())
			Expect(base.Error()).To(HaveOccurred())
			Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureParser.ErrorTypeNotArray(false))))
		})

		It("with key with array type returns value", func() {
			value := parser.Array("one")
			Expect(value).ToNot(BeNil())
			Expect(*value).To(Equal([]interface{}{"1", false}))
			Expect(base.Error()).ToNot(HaveOccurred())
		})
	})

	Context("Interface", func() {
		var parser *structureParser.Object

		BeforeEach(func() {
			parser = structureParser.NewObjectParser(base, &map[string]interface{}{
				"zero": false,
				"one": map[string]interface{}{
					"1": "2",
				},
				"two": []interface{}{
					"1",
					false,
				},
			})
			Expect(parser).ToNot(BeNil())
		})

		It("with key not found in the object returns nil", func() {
			Expect(parser.Interface("unknown")).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with key with primitive type returns value", func() {
			value := parser.Interface("zero")
			Expect(value).ToNot(BeNil())
			Expect(*value).To(BeFalse())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with key with object type returns value", func() {
			value := parser.Interface("one")
			Expect(value).ToNot(BeNil())
			Expect(*value).To(Equal(map[string]interface{}{"1": "2"}))
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with key with array type returns value", func() {
			value := parser.Interface("two")
			Expect(value).ToNot(BeNil())
			Expect(*value).To(Equal([]interface{}{"1", false}))
			Expect(base.Error()).ToNot(HaveOccurred())
		})
	})

	Context("NotParsed", func() {
		var parser *structureParser.Object

		BeforeEach(func() {
			parser = structureParser.NewObjectParser(base, &map[string]interface{}{
				"zero": 1,
				"one":  "two",
				"two":  3,
			})
			Expect(parser).ToNot(BeNil())
		})

		It("without anything parsed reports all unparsed as errors", func() {
			parser.NotParsed()
			Expect(base.Error()).To(HaveOccurred())
			Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(errors.Append(
				structureParser.ErrorNotParsed(),
				structureParser.ErrorNotParsed(),
				structureParser.ErrorNotParsed(),
			))))
		})

		It("with some items parsed reports all unparsed as errors", func() {
			parser.String("one")
			parser.NotParsed()
			Expect(base.Error()).To(HaveOccurred())
			Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(errors.Append(
				structureParser.ErrorNotParsed(),
				structureParser.ErrorNotParsed(),
			))))
		})

		It("with all items parsed has no errors", func() {
			parser.Int("zero")
			parser.String("one")
			parser.Int("two")
			parser.NotParsed()
			Expect(base.Error()).ToNot(HaveOccurred())
		})
	})

	Context("WithSource", func() {
		var parser *structureParser.Object

		BeforeEach(func() {
			parser = structureParser.NewObjectParser(base, nil)
			Expect(parser).ToNot(BeNil())
		})

		It("returns new parser", func() {
			src := testStructure.NewSource()
			result := parser.WithSource(src)
			Expect(result).ToNot(BeNil())
			Expect(result).ToNot(Equal(parser))
		})
	})

	Context("WithMeta", func() {
		var parser *structureParser.Object

		BeforeEach(func() {
			parser = structureParser.NewObjectParser(base, nil)
			Expect(parser).ToNot(BeNil())
		})

		It("returns new parser", func() {
			result := parser.WithMeta(testErrors.NewMeta())
			Expect(result).ToNot(BeNil())
			Expect(result).ToNot(Equal(parser))
		})
	})

	Context("WithReferenceObjectParser", func() {
		var parser *structureParser.Object

		BeforeEach(func() {
			parser = structureParser.NewObjectParser(base, &map[string]interface{}{
				"zero": false,
				"one": map[string]interface{}{
					"1": "2",
				},
			})
			Expect(parser).ToNot(BeNil())
		})

		It("with key not found in the object returns nil", func() {
			objectParser := parser.WithReferenceObjectParser("unknown")
			Expect(objectParser).ToNot(BeNil())
			Expect(objectParser.Exists()).To(BeFalse())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with key with different type returns nil and reports an ErrorCodeTypeNotObject", func() {
			objectParser := parser.WithReferenceObjectParser("zero")
			Expect(objectParser).ToNot(BeNil())
			Expect(objectParser.Exists()).To(BeFalse())
			Expect(base.Error()).To(HaveOccurred())
			Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureParser.ErrorTypeNotObject(false))))
		})

		It("with key with object type returns value", func() {
			objectParser := parser.WithReferenceObjectParser("one")
			Expect(objectParser).ToNot(BeNil())
			Expect(objectParser.Exists()).To(BeTrue())
			Expect(objectParser.References()).To(Equal([]string{"1"}))
			Expect(base.Error()).ToNot(HaveOccurred())
		})
	})

	Context("WithReferenceArrayParser", func() {
		var parser *structureParser.Object

		BeforeEach(func() {
			parser = structureParser.NewObjectParser(base, &map[string]interface{}{
				"zero": false,
				"one": []interface{}{
					"1",
					false,
				},
			})
			Expect(parser).ToNot(BeNil())
		})

		It("with key not found in the object returns nil", func() {
			arrayParser := parser.WithReferenceArrayParser("unknown")
			Expect(arrayParser).ToNot(BeNil())
			Expect(arrayParser.Exists()).To(BeFalse())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with key with different type returns nil and reports an ErrorCodeTypeNotArray", func() {
			arrayParser := parser.WithReferenceArrayParser("zero")
			Expect(arrayParser).ToNot(BeNil())
			Expect(arrayParser.Exists()).To(BeFalse())
			Expect(base.Error()).To(HaveOccurred())
			Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureParser.ErrorTypeNotArray(false))))
		})

		It("with key with object type returns value", func() {
			arrayParser := parser.WithReferenceArrayParser("one")
			Expect(arrayParser).ToNot(BeNil())
			Expect(arrayParser.Exists()).To(BeTrue())
			Expect(arrayParser.References()).To(Equal([]int{0, 1}))
			Expect(base.Error()).ToNot(HaveOccurred())
		})
	})
})
