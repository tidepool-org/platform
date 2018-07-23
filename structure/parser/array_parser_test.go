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

var _ = Describe("Array", func() {
	var base *structureBase.Base

	BeforeEach(func() {
		base = structureBase.New().WithSource(structure.NewPointerSource())
	})

	Context("NewArray", func() {
		It("returns successfully", func() {
			Expect(structureParser.NewArray(nil)).ToNot(BeNil())
		})
	})

	Context("NewArrayParser", func() {
		It("returns successfully", func() {
			Expect(structureParser.NewArrayParser(base, nil)).ToNot(BeNil())
		})
	})

	Context("with new parser with nil array", func() {
		var parser *structureParser.Array

		BeforeEach(func() {
			parser = structureParser.NewArrayParser(base, nil)
			Expect(parser).ToNot(BeNil())
		})

		Context("Origin", func() {
			It("returns OriginExternal if default", func() {
				Expect(parser.Origin()).To(Equal(structure.OriginExternal))
			})

			It("returns set origin", func() {
				Expect(parser.WithOrigin(structure.OriginInternal).Origin()).To(Equal(structure.OriginInternal))
			})
		})

		Context("HasSource", func() {
			It("returns false if no source set", func() {
				Expect(parser.WithSource(nil).HasSource()).To(BeFalse())
			})

			It("returns true if source set", func() {
				Expect(parser.WithSource(testStructure.NewSource()).HasSource()).To(BeTrue())
			})
		})

		Context("Source", func() {
			It("returns set source", func() {
				src := testStructure.NewSource()
				Expect(parser.WithSource(src).Source()).To(Equal(src))
			})
		})

		Context("HasMeta", func() {
			It("returns false if no meta set", func() {
				Expect(parser.WithMeta(nil).HasMeta()).To(BeFalse())
			})

			It("returns true if meta set", func() {
				Expect(parser.WithMeta(testErrors.NewMeta()).HasMeta()).To(BeTrue())
			})
		})

		Context("Meta", func() {
			It("returns default meta", func() {
				Expect(parser.Meta()).To(BeNil())
			})

			It("returns set meta", func() {
				meta := testErrors.NewMeta()
				Expect(parser.WithMeta(meta).Meta()).To(Equal(meta))
			})
		})

		Context("HasError", func() {
			It("returns false if no errors reported", func() {
				Expect(parser.HasError()).To(BeFalse())
			})

			It("returns true if any errors reported", func() {
				base.ReportError(testErrors.RandomError())
				Expect(parser.HasError()).To(BeTrue())
			})
		})

		Context("Error", func() {
			It("returns the error from the base", func() {
				err := testErrors.RandomError()
				base.ReportError(err)
				Expect(parser.Error()).To(Equal(errors.Normalize(err)))
			})
		})

		Context("ReportError", func() {
			It("reports the error to the base", func() {
				err := testErrors.RandomError()
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
			var arrayParsable *testStructure.ArrayParsable

			BeforeEach(func() {
				arrayParsable = testStructure.NewArrayParsable()
			})

			AfterEach(func() {
				arrayParsable.Expectations()
			})

			It("invokes parse and returns current errors", func() {
				err := testErrors.RandomError()
				base.ReportError(err)
				Expect(parser.Parse(arrayParsable)).To(Equal(errors.Normalize(err)))
				Expect(arrayParsable.ParseInputs).To(Equal([]structure.ArrayParser{parser}))
			})
		})

		Context("References", func() {
			It("returns nil references", func() {
				Expect(parser.References()).To(BeNil())
			})
		})

		Context("ReferenceExists", func() {
			It("returns false", func() {
				Expect(parser.ReferenceExists(0)).To(BeFalse())
			})
		})

		It("Bool returns nil", func() {
			Expect(parser.Bool(1)).To(BeNil())
		})

		It("Float64 returns nil", func() {
			Expect(parser.Float64(2)).To(BeNil())
		})

		It("Int returns nil", func() {
			Expect(parser.Int(3)).To(BeNil())
		})

		It("String returns nil", func() {
			Expect(parser.String(4)).To(BeNil())
		})

		It("StringArray returns nil", func() {
			Expect(parser.StringArray(5)).To(BeNil())
		})

		It("Time returns nil", func() {
			Expect(parser.Time(6, time.RFC3339)).To(BeNil())
		})

		It("Object returns nil", func() {
			Expect(parser.Object(7)).To(BeNil())
		})

		It("Array returns nil", func() {
			Expect(parser.Array(8)).To(BeNil())
		})

		It("Interface returns nil", func() {
			Expect(parser.Interface(9)).To(BeNil())
		})

		It("NotParsed only returns existing errors", func() {
			err := testErrors.RandomError()
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
			It("returns new parser", func() {
				result := parser.WithReferenceObjectParser(0)
				Expect(result).ToNot(BeNil())
				Expect(result).ToNot(Equal(parser))
				Expect(result.Exists()).To(BeFalse())
			})
		})

		Context("WithReferenceArrayParser", func() {
			It("with source returns new parser", func() {
				src := testStructure.NewSource()
				src.WithReferenceOutputs = []structure.Source{testStructure.NewSource()}
				resultWithSource := parser.WithSource(src)
				resultWithReference := parser.WithReferenceArrayParser(0)
				Expect(resultWithReference).ToNot(BeNil())
				Expect(resultWithReference).ToNot(Equal(resultWithSource))
				Expect(resultWithReference.Exists()).To(BeFalse())
			})
		})
	})

	Context("with new parser with valid, empty array", func() {
		var parser *structureParser.Array

		BeforeEach(func() {
			parser = structureParser.NewArrayParser(base, &[]interface{}{})
			Expect(parser).ToNot(BeNil())
		})

		Context("Error", func() {
			It("returns the error from the base", func() {
				err := testErrors.RandomError()
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
			var arrayParsable *testStructure.ArrayParsable

			BeforeEach(func() {
				arrayParsable = testStructure.NewArrayParsable()
			})

			AfterEach(func() {
				arrayParsable.Expectations()
			})

			It("invokes parse and returns current errors", func() {
				err := testErrors.RandomError()
				base.ReportError(err)
				Expect(parser.Parse(arrayParsable)).To(Equal(errors.Normalize(err)))
				Expect(arrayParsable.ParseInputs).To(Equal([]structure.ArrayParser{parser}))
			})
		})

		Context("References", func() {
			It("returns empty references", func() {
				Expect(parser.References()).To(BeEmpty())
			})
		})
	})

	Context("References", func() {
		var parser *structureParser.Array

		BeforeEach(func() {
			parser = structureParser.NewArrayParser(base, &[]interface{}{
				"exists",
				"exists",
			})
			Expect(parser).ToNot(BeNil())
		})

		It("returns the array of references", func() {
			Expect(parser.References()).To(ConsistOf(0, 1))
		})
	})

	Context("ReferenceExists", func() {
		var parser *structureParser.Array

		BeforeEach(func() {
			parser = structureParser.NewArrayParser(base, &[]interface{}{
				"exists",
			})
			Expect(parser).ToNot(BeNil())
		})

		It("returns true if reference exists", func() {
			Expect(parser.ReferenceExists(0)).To(BeTrue())
		})

		It("returns false if reference does not exist", func() {
			Expect(parser.ReferenceExists(1)).To(BeFalse())
		})
	})

	Context("Bool", func() {
		var parser *structureParser.Array

		BeforeEach(func() {
			parser = structureParser.NewArrayParser(base, &[]interface{}{
				"not a boolean",
				true,
			})
			Expect(parser).ToNot(BeNil())
		})

		It("with index parameter less that the first index in the array returns nil", func() {
			Expect(parser.Bool(-1)).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter greater than the last index in the array returns nil", func() {
			Expect(parser.Bool(len(parser.References()))).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter with different type returns nil and reports an ErrorTypeNotBool", func() {
			Expect(parser.Bool(0)).To(BeNil())
			Expect(base.Error()).To(HaveOccurred())
			testErrors.ExpectEqual(base.Error(), testErrors.WithPointerSource(structureParser.ErrorTypeNotBool("not a boolean"), "/0"))
		})

		It("with index parameter with boolean type returns value", func() {
			value := parser.Bool(1)
			Expect(value).ToNot(BeNil())
			Expect(*value).To(BeTrue())
			Expect(base.Error()).ToNot(HaveOccurred())
		})
	})

	Context("Float64", func() {
		var parser *structureParser.Array

		BeforeEach(func() {
			parser = structureParser.NewArrayParser(base, &[]interface{}{
				false,
				3,
				4.0,
				5.67,
			})
			Expect(parser).ToNot(BeNil())
		})

		It("with index parameter less that the first index in the array returns nil", func() {
			Expect(parser.Float64(-1)).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter greater than the last index in the array returns nil", func() {
			Expect(parser.Float64(len(parser.References()))).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter with different type returns nil and reports an ErrorTypeNotFloat64", func() {
			Expect(parser.Float64(0)).To(BeNil())
			Expect(base.Error()).To(HaveOccurred())
			testErrors.ExpectEqual(base.Error(), testErrors.WithPointerSource(structureParser.ErrorTypeNotFloat64(false), "/0"))
		})

		It("with index parameter with integer type returns value", func() {
			value := parser.Float64(1)
			Expect(value).ToNot(BeNil())
			Expect(*value).To(BeNumerically("==", 3.))
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter with float type and whole number returns value", func() {
			value := parser.Float64(2)
			Expect(value).ToNot(BeNil())
			Expect(*value).To(BeNumerically("==", 4.0))
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter with float type and not whole number returns value", func() {
			value := parser.Float64(3)
			Expect(value).ToNot(BeNil())
			Expect(*value).To(BeNumerically("==", 5.67))
			Expect(base.Error()).ToNot(HaveOccurred())
		})
	})

	Context("Int", func() {
		var parser *structureParser.Array

		BeforeEach(func() {
			parser = structureParser.NewArrayParser(base, &[]interface{}{
				false,
				3,
				4.0,
				5.67,
			})
			Expect(parser).ToNot(BeNil())
		})

		It("with index parameter less that the first index in the array returns nil", func() {
			Expect(parser.Int(-1)).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter greater than the last index in the array returns nil", func() {
			Expect(parser.Int(len(parser.References()))).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter with different type returns nil and reports an ErrorTypeNotInt", func() {
			Expect(parser.Int(0)).To(BeNil())
			Expect(base.Error()).To(HaveOccurred())
			testErrors.ExpectEqual(base.Error(), testErrors.WithPointerSource(structureParser.ErrorTypeNotInt(false), "/0"))
		})

		It("with index parameter with integer type returns value", func() {
			value := parser.Int(1)
			Expect(value).ToNot(BeNil())
			Expect(*value).To(BeNumerically("==", 3))
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter with float type and whole number returns value", func() {
			value := parser.Int(2)
			Expect(value).ToNot(BeNil())
			Expect(*value).To(BeNumerically("==", 4))
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter with float type and not whole number returns nil and reports an ErrorTypeNotInt", func() {
			Expect(parser.Int(3)).To(BeNil())
			Expect(base.Error()).To(HaveOccurred())
			testErrors.ExpectEqual(base.Error(), testErrors.WithPointerSource(structureParser.ErrorTypeNotInt(5.67), "/3"))
		})
	})

	Context("String", func() {
		var parser *structureParser.Array

		BeforeEach(func() {
			parser = structureParser.NewArrayParser(base, &[]interface{}{
				false,
				"this is a string",
			})
			Expect(parser).ToNot(BeNil())
		})

		It("with index parameter less that the first index in the array returns nil", func() {
			Expect(parser.String(-1)).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter greater than the last index in the array returns nil", func() {
			Expect(parser.String(len(parser.References()))).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter with different type returns nil and reports an ErrorTypeNotString", func() {
			Expect(parser.String(0)).To(BeNil())
			Expect(base.Error()).To(HaveOccurred())
			testErrors.ExpectEqual(base.Error(), testErrors.WithPointerSource(structureParser.ErrorTypeNotString(false), "/0"))
		})

		It("with index parameter with string type returns value", func() {
			value := parser.String(1)
			Expect(value).ToNot(BeNil())
			Expect(*value).To(Equal("this is a string"))
			Expect(base.Error()).ToNot(HaveOccurred())
		})
	})

	Context("StringArray", func() {
		var parser *structureParser.Array

		BeforeEach(func() {
			parser = structureParser.NewArrayParser(base, &[]interface{}{
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
			})
			Expect(parser).ToNot(BeNil())
		})

		It("with index parameter less that the first index in the array returns nil", func() {
			Expect(parser.StringArray(-1)).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter greater than the last index in the array returns nil", func() {
			Expect(parser.StringArray(len(parser.References()))).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter with different type returns nil and reports an ErrorTypeNotArray", func() {
			Expect(parser.StringArray(0)).To(BeNil())
			Expect(base.Error()).To(HaveOccurred())
			testErrors.ExpectEqual(base.Error(), testErrors.WithPointerSource(structureParser.ErrorTypeNotArray(false), "/0"))
		})

		It("with index parameter with string array type returns value", func() {
			value := parser.StringArray(1)
			Expect(value).ToNot(BeNil())
			Expect(*value).To(Equal([]string{"one", "two"}))
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter with interface array and contains all string type returns value", func() {
			value := parser.StringArray(2)
			Expect(value).ToNot(BeNil())
			Expect(*value).To(Equal([]string{"three", "four"}))
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter with interface array and does not contains all string type returns partial value and ErrorTypeNotString", func() {
			value := parser.StringArray(3)
			Expect(value).ToNot(BeNil())
			Expect(*value).To(Equal([]string{"five", ""}))
			Expect(base.Error()).To(HaveOccurred())
			testErrors.ExpectEqual(base.Error(), testErrors.WithPointerSource(structureParser.ErrorTypeNotString(6), "/3/1"))
		})
	})

	Context("Time", func() {
		var now time.Time
		var parser *structureParser.Array

		BeforeEach(func() {
			now = time.Now().Truncate(time.Second)
			parser = structureParser.NewArrayParser(base, &[]interface{}{
				false,
				"abc",
				now.Format(time.RFC3339),
			})
			Expect(parser).ToNot(BeNil())
		})

		It("with index parameter less that the first index in the array returns nil", func() {
			Expect(parser.Time(-1, time.RFC3339)).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter greater than the last index in the array returns nil", func() {
			Expect(parser.Time(len(parser.References()), time.RFC3339)).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter with different type returns nil and reports an ErrorTypeNotTime", func() {
			Expect(parser.Time(0, time.RFC3339)).To(BeNil())
			Expect(base.Error()).To(HaveOccurred())
			testErrors.ExpectEqual(base.Error(), testErrors.WithPointerSource(structureParser.ErrorTypeNotTime(false), "/0"))
		})

		It("with index parameter with different type returns nil and reports an ErrorValueTimeNotParsable", func() {
			Expect(parser.Time(1, time.RFC3339)).To(BeNil())
			Expect(base.Error()).To(HaveOccurred())
			testErrors.ExpectEqual(base.Error(), testErrors.WithPointerSource(structureParser.ErrorValueTimeNotParsable("abc", time.RFC3339), "/1"))
		})

		It("with index parameter with string type returns value", func() {
			value := parser.Time(2, time.RFC3339)
			Expect(value).ToNot(BeNil())
			Expect(*value).To(BeTemporally("==", now))
			Expect(base.Error()).ToNot(HaveOccurred())
		})
	})

	Context("Object", func() {
		var parser *structureParser.Array

		BeforeEach(func() {
			parser = structureParser.NewArrayParser(base, &[]interface{}{
				false,
				map[string]interface{}{
					"1": "2",
				},
			})
			Expect(parser).ToNot(BeNil())
		})

		It("with index parameter less that the first index in the array returns nil", func() {
			Expect(parser.Object(-1)).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter greater than the last index in the array returns nil", func() {
			Expect(parser.Object(len(parser.References()))).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter with different type returns nil and reports an ErrorTypeNotObject", func() {
			Expect(parser.Object(0)).To(BeNil())
			Expect(base.Error()).To(HaveOccurred())
			testErrors.ExpectEqual(base.Error(), testErrors.WithPointerSource(structureParser.ErrorTypeNotObject(false), "/0"))
		})

		It("with index parameter with object type returns value", func() {
			value := parser.Object(1)
			Expect(value).ToNot(BeNil())
			Expect(*value).To(Equal(map[string]interface{}{"1": "2"}))
			Expect(base.Error()).ToNot(HaveOccurred())
		})
	})

	Context("Array", func() {
		var parser *structureParser.Array

		BeforeEach(func() {
			parser = structureParser.NewArrayParser(base, &[]interface{}{
				false,
				[]interface{}{
					"1",
					false,
				},
			})
			Expect(parser).ToNot(BeNil())
		})

		It("with index parameter less that the first index in the array returns nil", func() {
			Expect(parser.Array(-1)).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter greater than the last index in the array returns nil", func() {
			Expect(parser.Array(len(parser.References()))).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter with different type returns nil and reports an ErrorTypeNotArray", func() {
			Expect(parser.Array(0)).To(BeNil())
			Expect(base.Error()).To(HaveOccurred())
			testErrors.ExpectEqual(base.Error(), testErrors.WithPointerSource(structureParser.ErrorTypeNotArray(false), "/0"))
		})

		It("with index parameter with object array type returns value", func() {
			value := parser.Array(1)
			Expect(value).ToNot(BeNil())
			Expect(*value).To(Equal([]interface{}{"1", false}))
			Expect(base.Error()).ToNot(HaveOccurred())
		})
	})

	Context("Interface", func() {
		var parser *structureParser.Array

		BeforeEach(func() {
			parser = structureParser.NewArrayParser(base, &[]interface{}{
				false,
				map[string]interface{}{
					"1": "2",
				},
				[]interface{}{
					"1",
					false,
				},
				nil,
			})
			Expect(parser).ToNot(BeNil())
		})

		It("with index parameter less that the first index in the array returns nil", func() {
			Expect(parser.Interface(-1)).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter greater than the last index in the array returns nil", func() {
			Expect(parser.Interface(len(parser.References()))).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter with primitive type returns value", func() {
			value := parser.Interface(0)
			Expect(value).ToNot(BeNil())
			Expect(*value).To(BeFalse())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter with object type returns value", func() {
			value := parser.Interface(1)
			Expect(value).ToNot(BeNil())
			Expect(*value).To(Equal(map[string]interface{}{"1": "2"}))
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter with array type returns value", func() {
			value := parser.Interface(2)
			Expect(value).ToNot(BeNil())
			Expect(*value).To(Equal([]interface{}{"1", false}))
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter with nil return type", func() {
			Expect(parser.Interface(3)).To(BeNil())
			Expect(base.Error()).ToNot(HaveOccurred())
		})
	})

	Context("NotParsed", func() {
		var parser *structureParser.Array

		BeforeEach(func() {
			parser = structureParser.NewArrayParser(base, &[]interface{}{
				1,
				"two",
				3,
			})
			Expect(parser).ToNot(BeNil())
		})

		It("without anything parsed reports all unparsed as errors", func() {
			parser.NotParsed()
			Expect(base.Error()).To(HaveOccurred())
			testErrors.ExpectEqual(base.Error(), errors.Append(
				testErrors.WithPointerSource(structureParser.ErrorNotParsed(), "/0"),
				testErrors.WithPointerSource(structureParser.ErrorNotParsed(), "/1"),
				testErrors.WithPointerSource(structureParser.ErrorNotParsed(), "/2"),
			))
		})

		It("with some items parsed reports all unparsed as errors", func() {
			parser.String(1)
			parser.NotParsed()
			Expect(base.Error()).To(HaveOccurred())
			testErrors.ExpectEqual(base.Error(), errors.Append(
				testErrors.WithPointerSource(structureParser.ErrorNotParsed(), "/0"),
				testErrors.WithPointerSource(structureParser.ErrorNotParsed(), "/2"),
			))
		})

		It("with all items parsed has no errors", func() {
			parser.Int(0)
			parser.String(1)
			parser.Int(2)
			parser.NotParsed()
			Expect(base.Error()).ToNot(HaveOccurred())
		})
	})

	Context("WithOrigin", func() {
		var parser *structureParser.Array

		BeforeEach(func() {
			parser = structureParser.NewArrayParser(base, nil)
			Expect(parser).ToNot(BeNil())
		})

		It("returns a new parser with origin", func() {
			result := parser.WithOrigin(structure.OriginInternal)
			Expect(result).ToNot(BeNil())
			Expect(result).ToNot(BeIdenticalTo(parser))
			Expect(result.Error()).ToNot(HaveOccurred())
			Expect(result.Origin()).To(Equal(structure.OriginInternal))
			Expect(parser.Origin()).To(Equal(structure.OriginExternal))
		})
	})

	Context("WithSource", func() {
		var parser *structureParser.Array

		BeforeEach(func() {
			parser = structureParser.NewArrayParser(base, nil)
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
		var parser *structureParser.Array

		BeforeEach(func() {
			parser = structureParser.NewArrayParser(base, nil)
			Expect(parser).ToNot(BeNil())
		})

		It("returns new parser", func() {
			result := parser.WithMeta(testErrors.NewMeta())
			Expect(result).ToNot(BeNil())
			Expect(result).ToNot(Equal(parser))
		})
	})

	Context("WithReferenceObjectParser", func() {
		var parser *structureParser.Array

		BeforeEach(func() {
			parser = structureParser.NewArrayParser(base, &[]interface{}{
				false,
				map[string]interface{}{
					"1": "2",
				},
			})
			Expect(parser).ToNot(BeNil())
		})

		It("with index parameter less that the first index in the array returns nil", func() {
			objectParser := parser.WithReferenceObjectParser(-1)
			Expect(objectParser).ToNot(BeNil())
			Expect(objectParser.Exists()).To(BeFalse())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter greater than the last index in the array returns nil", func() {
			objectParser := parser.WithReferenceObjectParser(len(parser.References()))
			Expect(objectParser).ToNot(BeNil())
			Expect(objectParser.Exists()).To(BeFalse())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter with different type returns nil and reports an ErrorTypeNotObject", func() {
			objectParser := parser.WithReferenceObjectParser(0)
			Expect(objectParser).ToNot(BeNil())
			Expect(objectParser.Exists()).To(BeFalse())
			Expect(base.Error()).To(HaveOccurred())
			testErrors.ExpectEqual(base.Error(), testErrors.WithPointerSource(structureParser.ErrorTypeNotObject(false), "/0"))
		})

		It("with index parameter with object type returns value", func() {
			objectParser := parser.WithReferenceObjectParser(1)
			Expect(objectParser).ToNot(BeNil())
			Expect(objectParser.Exists()).To(BeTrue())
			Expect(objectParser.References()).To(Equal([]string{"1"}))
			Expect(base.Error()).ToNot(HaveOccurred())
		})
	})

	Context("WithReferenceArrayParser", func() {
		var parser *structureParser.Array

		BeforeEach(func() {
			parser = structureParser.NewArrayParser(base, &[]interface{}{
				false,
				[]interface{}{
					"1",
					false,
				},
			})
			Expect(parser).ToNot(BeNil())
		})

		It("with index parameter less that the first index in the array returns nil", func() {
			arrayParser := parser.WithReferenceArrayParser(-1)
			Expect(arrayParser).ToNot(BeNil())
			Expect(arrayParser.Exists()).To(BeFalse())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter greater than the last index in the array returns nil", func() {
			arrayParser := parser.WithReferenceArrayParser(len(parser.References()))
			Expect(arrayParser).ToNot(BeNil())
			Expect(arrayParser.Exists()).To(BeFalse())
			Expect(base.Error()).ToNot(HaveOccurred())
		})

		It("with index parameter with different type returns nil and reports an ErrorTypeNotArray", func() {
			arrayParser := parser.WithReferenceArrayParser(0)
			Expect(arrayParser).ToNot(BeNil())
			Expect(arrayParser.Exists()).To(BeFalse())
			Expect(base.Error()).To(HaveOccurred())
			testErrors.ExpectEqual(base.Error(), testErrors.WithPointerSource(structureParser.ErrorTypeNotArray(false), "/0"))
		})

		It("with index parameter with object type returns value", func() {
			arrayParser := parser.WithReferenceArrayParser(1)
			Expect(arrayParser).ToNot(BeNil())
			Expect(arrayParser.Exists()).To(BeTrue())
			Expect(arrayParser.References()).To(Equal([]int{0, 1}))
			Expect(base.Error()).ToNot(HaveOccurred())
		})
	})
})
