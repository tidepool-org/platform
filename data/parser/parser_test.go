package parser_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/parser"
	testData "github.com/tidepool-org/platform/data/test"
)

var _ = Describe("Parser", func() {
	Context("ParseDatum", func() {
		var testDatum *testData.Datum
		var testObjectParser *TestObjectParser
		var testFactory *TestFactory

		BeforeEach(func() {
			testDatum = testData.NewDatum()
			testDatum.ParseOutputs = []error{nil}
			testObjectParser = &TestObjectParser{}
			testObjectParser.ObjectOutputs = []*map[string]interface{}{{}}
			testFactory = &TestFactory{}
			testFactory.InitOutputs = []InitOutput{{testDatum, nil}}
		})

		It("successfully returns a datum", func() {
			datum, err := parser.ParseDatum(testObjectParser, testFactory)
			Expect(err).ToNot(HaveOccurred())
			Expect(datum).ToNot(BeNil())
			Expect(*datum).To(Equal(testDatum))
			Expect(testFactory.InitInputs).To(HaveLen(1))
			Expect(testDatum.ParseInputs).To(ConsistOf(testObjectParser))
		})

		It("returns an error if the parser is missing", func() {
			datum, err := parser.ParseDatum(nil, testFactory)
			Expect(err).To(MatchError("parser is missing"))
			Expect(datum).To(BeNil())
		})

		It("returns an error if the factory is missing", func() {
			datum, err := parser.ParseDatum(testObjectParser, nil)
			Expect(err).To(MatchError("factory is missing"))
			Expect(datum).To(BeNil())
		})

		It("successfully returns a nil datum if the parser object is nil", func() {
			testObjectParser.ObjectOutputs = []*map[string]interface{}{nil}
			Expect(parser.ParseDatum(testObjectParser, testFactory)).To(BeNil())
		})

		It("returns an error if the factory init returns an error", func() {
			testFactory.InitOutputs = []InitOutput{{nil, errors.New("init returns error")}}
			datum, err := parser.ParseDatum(testObjectParser, testFactory)
			Expect(err).To(MatchError("init returns error"))
			Expect(datum).To(BeNil())
			Expect(testFactory.InitInputs).To(HaveLen(1))
		})

		It("successfully returns nil if the factory init returns nil", func() {
			testFactory.InitOutputs = []InitOutput{{nil, nil}}
			Expect(parser.ParseDatum(testObjectParser, testFactory)).To(BeNil())
			Expect(testFactory.InitInputs).To(HaveLen(1))
		})

		It("returns an error if datum parse returns an error", func() {
			testDatum.ParseOutputs = []error{errors.New("parse returns error")}
			datum, err := parser.ParseDatum(testObjectParser, testFactory)
			Expect(err).To(MatchError("parse returns error"))
			Expect(datum).To(BeNil())
			Expect(testFactory.InitInputs).To(HaveLen(1))
			Expect(testDatum.ParseInputs).To(ConsistOf(testObjectParser))
		})
	})

	Context("ParseArray", func() {
		var testDatum1 data.Datum
		var testDatum2 data.Datum
		var testArrayParser *TestArrayParser

		BeforeEach(func() {
			testDatum1 = testData.NewDatum()
			testDatum2 = testData.NewDatum()
			testArrayParser = &TestArrayParser{}
			testArrayParser.ArrayOutputs = []*[]interface{}{{testDatum1, testDatum2}}
			testArrayParser.ParseDatumOutputs = []*data.Datum{&testDatum1, &testDatum2}
		})

		It("successfully returns a datum array", func() {
			datumArray, err := parser.ParseDatumArray(testArrayParser)
			Expect(err).ToNot(HaveOccurred())
			Expect(datumArray).ToNot(BeNil())
			Expect(*datumArray).To(ConsistOf(testDatum1, testDatum2))
			Expect(testArrayParser.ParseDatumInputs).To(ConsistOf(0, 1))
		})

		It("returns an error if the parser is missing", func() {
			datumArray, err := parser.ParseDatumArray(nil)
			Expect(err).To(MatchError("parser is missing"))
			Expect(datumArray).To(BeNil())
		})

		It("successfully returns a nil datum array if the parser array is nil", func() {
			testArrayParser.ArrayOutputs = []*[]interface{}{nil}
			Expect(parser.ParseDatumArray(testArrayParser)).To(BeNil())
		})

		It("successfully returns a empty datum array if the parser array is empty", func() {
			testArrayParser.ArrayOutputs = []*[]interface{}{{}}
			datumArray, err := parser.ParseDatumArray(testArrayParser)
			Expect(err).ToNot(HaveOccurred())
			Expect(datumArray).ToNot(BeNil())
			Expect(*datumArray).To(BeEmpty())
			Expect(testArrayParser.ParseDatumInputs).To(BeNil())
		})

		It("successfully returns a partial datum array if any datum pointer is nil", func() {
			testArrayParser.ParseDatumOutputs = []*data.Datum{nil, &testDatum2}
			datumArray, err := parser.ParseDatumArray(testArrayParser)
			Expect(err).ToNot(HaveOccurred())
			Expect(datumArray).ToNot(BeNil())
			Expect(*datumArray).To(ConsistOf(testDatum2))
			Expect(testArrayParser.ParseDatumInputs).To(ConsistOf(0, 1))
		})

		It("successfully returns a partial datum array if any datum is nil", func() {
			testDatum1 = nil
			datumArray, err := parser.ParseDatumArray(testArrayParser)
			Expect(err).ToNot(HaveOccurred())
			Expect(datumArray).ToNot(BeNil())
			Expect(*datumArray).To(ConsistOf(testDatum2))
			Expect(testArrayParser.ParseDatumInputs).To(ConsistOf(0, 1))
		})
	})
})
