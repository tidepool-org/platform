package work_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataWork "github.com/tidepool-org/platform/data/work"
	dataWorkTest "github.com/tidepool-org/platform/data/work/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("ingestion_offset", func() {
	Context("IngestionOffsetMetadata", func() {
		Context("MetadataKeyDataIngestionOffset", func() {
			It("returns expected value", func() {
				Expect(dataWork.MetadataKeyIngestionOffset).To(Equal("ingestionOffset"))
			})
		})

		Context("IngestionOffsetMetadata", func() {
			DescribeTable("serializes the datum as expected",
				func(mutator func(datum *dataWork.IngestionOffsetMetadata)) {
					datum := dataWorkTest.RandomIngestionOffsetMetadata(test.AllowOptional())
					mutator(datum)
					test.ExpectSerializedObjectJSON(datum, dataWorkTest.NewObjectFromIngestionOffsetMetadata(datum, test.ObjectFormatJSON))
					test.ExpectSerializedObjectBSON(datum, dataWorkTest.NewObjectFromIngestionOffsetMetadata(datum, test.ObjectFormatBSON))
				},
				Entry("succeeds",
					func(datum *dataWork.IngestionOffsetMetadata) {},
				),
				Entry("empty",
					func(datum *dataWork.IngestionOffsetMetadata) {
						*datum = dataWork.IngestionOffsetMetadata{}
					},
				),
				Entry("all",
					func(datum *dataWork.IngestionOffsetMetadata) {
						datum.IngestionOffset = pointer.From(dataWorkTest.RandomIngestionOffset())
					},
				),
			)

			Context("Parse", func() {
				DescribeTable("parses the datum",
					func(mutator func(object map[string]any, expectedDatum *dataWork.IngestionOffsetMetadata), expectedErrors ...error) {
						expectedDatum := dataWorkTest.RandomIngestionOffsetMetadata(test.AllowOptional())
						object := dataWorkTest.NewObjectFromIngestionOffsetMetadata(expectedDatum, test.ObjectFormatJSON)
						mutator(object, expectedDatum)
						result := &dataWork.IngestionOffsetMetadata{}
						errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
						Expect(result).To(Equal(expectedDatum))
					},
					Entry("succeeds",
						func(object map[string]any, expectedDatum *dataWork.IngestionOffsetMetadata) {},
					),
					Entry("empty",
						func(object map[string]any, expectedDatum *dataWork.IngestionOffsetMetadata) {
							clear(object)
							*expectedDatum = dataWork.IngestionOffsetMetadata{}
						},
					),
					Entry("multiple errors",
						func(object map[string]any, expectedDatum *dataWork.IngestionOffsetMetadata) {
							object["ingestionOffset"] = true
							expectedDatum.IngestionOffset = nil
						},
						errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/ingestionOffset"),
					),
				)
			})

			Context("Validate", func() {
				DescribeTable("validates the datum",
					func(mutator func(datum *dataWork.IngestionOffsetMetadata), expectedErrors ...error) {
						datum := dataWorkTest.RandomIngestionOffsetMetadata(test.AllowOptional())
						mutator(datum)
						errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
					},
					Entry("succeeds",
						func(datum *dataWork.IngestionOffsetMetadata) {},
					),
					Entry("ingestion offset missing",
						func(datum *dataWork.IngestionOffsetMetadata) {
							datum.IngestionOffset = nil
						},
					),
					Entry("ingestion offset invalid",
						func(datum *dataWork.IngestionOffsetMetadata) {
							datum.IngestionOffset = pointer.From(-1)
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/ingestionOffset"),
					),
					Entry("ingestion offset valid",
						func(datum *dataWork.IngestionOffsetMetadata) {
							datum.IngestionOffset = pointer.From(dataWorkTest.RandomIngestionOffset())
						},
					),
					Entry("multiple errors",
						func(datum *dataWork.IngestionOffsetMetadata) {
							datum.IngestionOffset = pointer.From(-1)
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/ingestionOffset"),
					),
				)
			})
		})
	})
})
