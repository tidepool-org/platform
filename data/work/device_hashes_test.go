package work_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataWork "github.com/tidepool-org/platform/data/work"
	dataWorkTest "github.com/tidepool-org/platform/data/work/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("device_hashes", func() {
	It("DeviceIDLengthMaximum is expected", func() {
		Expect(dataWork.DeviceIDLengthMaximum).To(Equal(100))
	})

	It("DeviceHashLengthMaximum is expected", func() {
		Expect(dataWork.DeviceHashLengthMaximum).To(Equal(100))
	})

	It("DeviceHashesLengthMaximum is expected", func() {
		Expect(dataWork.DeviceHashesLengthMaximum).To(Equal(100))
	})

	Context("DeviceHashes", func() {
		Context("ParseDeviceHashes", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataWork.ParseDeviceHashes(structureParser.NewObject(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataWorkTest.RandomDeviceHashes()
				object := dataWorkTest.NewObjectFromDeviceHashes(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(logTest.NewLogger(), &object)
				Expect(dataWork.ParseDeviceHashes(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataWork.DeviceHashes)) {
				datum := dataWorkTest.RandomDeviceHashes()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataWorkTest.NewObjectFromDeviceHashes(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataWorkTest.NewObjectFromDeviceHashes(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataWork.DeviceHashes) {},
			),
			Entry("empty",
				func(datum *dataWork.DeviceHashes) {
					*datum = dataWork.DeviceHashes{}
				},
			),
		)

		Context("ParseDeviceHashes", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataWork.ParseDeviceHashes(structureParser.NewObject(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataWorkTest.RandomDeviceHashes()
				object := dataWorkTest.NewObjectFromDeviceHashes(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(logTest.NewLogger(), &object)
				Expect(dataWork.ParseDeviceHashes(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *dataWork.DeviceHashes), expectedErrors ...error) {
					expectedDatum := dataWorkTest.RandomDeviceHashes()
					object := dataWorkTest.NewObjectFromDeviceHashes(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &dataWork.DeviceHashes{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *dataWork.DeviceHashes) {},
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataWork.DeviceHashes), expectedErrors ...error) {
					datum := dataWorkTest.RandomDeviceHashes()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataWork.DeviceHashes) {},
				),
				Entry("empty",
					func(datum *dataWork.DeviceHashes) { *datum = dataWork.DeviceHashes{} },
				),
				Entry("single valid",
					func(datum *dataWork.DeviceHashes) {
						*datum = dataWork.DeviceHashes{}
						(*datum)[dataWorkTest.RandomDeviceID()] = dataWorkTest.RandomDeviceHash()
					},
				),
				Entry("multiple valid",
					func(datum *dataWork.DeviceHashes) {
						*datum = *dataWorkTest.RandomDeviceHashes()
					},
				),
				Entry("length in range (upper)",
					func(datum *dataWork.DeviceHashes) {
						*datum = dataWork.DeviceHashes{}
						for range dataWork.DeviceHashesLengthMaximum {
							(*datum)[dataWorkTest.RandomDeviceID()] = dataWorkTest.RandomDeviceHash()
						}
					},
				),
				Entry("length out of range (upper)",
					func(datum *dataWork.DeviceHashes) {
						*datum = dataWork.DeviceHashes{}
						for range dataWork.DeviceHashesLengthMaximum + 1 {
							(*datum)[dataWorkTest.RandomDeviceID()] = dataWorkTest.RandomDeviceHash()
						}
					},
					structureValidator.ErrorLengthNotLessThanOrEqualTo(dataWork.DeviceHashesLengthMaximum+1, dataWork.DeviceHashesLengthMaximum),
				),
				Entry("device id empty",
					func(datum *dataWork.DeviceHashes) {
						*datum = dataWork.DeviceHashes{}
						(*datum)[""] = dataWorkTest.RandomDeviceHash()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/#"),
				),
				Entry("device id length in range (upper)",
					func(datum *dataWork.DeviceHashes) {
						*datum = dataWork.DeviceHashes{}
						(*datum)[test.RandomStringFromRange(dataWork.DeviceIDLengthMaximum, dataWork.DeviceIDLengthMaximum)] = dataWorkTest.RandomDeviceHash()
					},
				),
				Entry("device id length out of range (upper)",
					func(datum *dataWork.DeviceHashes) {
						*datum = dataWork.DeviceHashes{}
						(*datum)[strings.Repeat("X", dataWork.DeviceIDLengthMaximum+1)] = dataWorkTest.RandomDeviceHash()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(dataWork.DeviceIDLengthMaximum+1, dataWork.DeviceIDLengthMaximum), "/"+strings.Repeat("X", dataWork.DeviceIDLengthMaximum+1)+"#"),
				),
				Entry("device hash empty",
					func(datum *dataWork.DeviceHashes) {
						*datum = dataWork.DeviceHashes{}
						(*datum)["empty"] = ""
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/empty"),
				),
				Entry("device hash length in range (upper)",
					func(datum *dataWork.DeviceHashes) {
						*datum = dataWork.DeviceHashes{}
						(*datum)[dataWorkTest.RandomDeviceID()] = test.RandomStringFromRange(dataWork.DeviceHashLengthMaximum, dataWork.DeviceHashLengthMaximum)
					},
				),
				Entry("device hash length out of range (upper)",
					func(datum *dataWork.DeviceHashes) {
						*datum = dataWork.DeviceHashes{}
						(*datum)["length"] = test.RandomStringFromRange(dataWork.DeviceHashLengthMaximum+1, dataWork.DeviceHashLengthMaximum+1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(dataWork.DeviceHashLengthMaximum+1, dataWork.DeviceHashLengthMaximum), "/length"),
				),
				Entry("multiple errors",
					func(datum *dataWork.DeviceHashes) {
						*datum = dataWork.DeviceHashes{}
						for range dataWork.DeviceHashesLengthMaximum {
							(*datum)[dataWorkTest.RandomDeviceID()] = dataWorkTest.RandomDeviceHash()
						}
						(*datum)[""] = ""
					},
					structureValidator.ErrorLengthNotLessThanOrEqualTo(dataWork.DeviceHashesLengthMaximum+1, dataWork.DeviceHashesLengthMaximum),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/#"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/"),
				),
			)
		})
	})

	Context("DeviceHashesMetadata", func() {
		Context("MetadataKeyDeviceHashes", func() {
			It("returns expected value", func() {
				Expect(dataWork.MetadataKeyDeviceHashes).To(Equal("deviceHashes"))
			})
		})

		Context("DeviceHashesMetadata", func() {
			DescribeTable("serializes the datum as expected",
				func(mutator func(datum *dataWork.DeviceHashesMetadata)) {
					datum := dataWorkTest.RandomDeviceHashesMetadata(test.AllowOptional())
					mutator(datum)
					test.ExpectSerializedObjectJSON(datum, dataWorkTest.NewObjectFromDeviceHashesMetadata(datum, test.ObjectFormatJSON))
					test.ExpectSerializedObjectBSON(datum, dataWorkTest.NewObjectFromDeviceHashesMetadata(datum, test.ObjectFormatBSON))
				},
				Entry("succeeds",
					func(datum *dataWork.DeviceHashesMetadata) {},
				),
				Entry("empty",
					func(datum *dataWork.DeviceHashesMetadata) {
						*datum = dataWork.DeviceHashesMetadata{}
					},
				),
				Entry("all",
					func(datum *dataWork.DeviceHashesMetadata) {
						datum.DeviceHashes = dataWorkTest.RandomDeviceHashes()
					},
				),
			)

			Context("Parse", func() {
				DescribeTable("parses the datum",
					func(mutator func(object map[string]any, expectedDatum *dataWork.DeviceHashesMetadata), expectedErrors ...error) {
						expectedDatum := dataWorkTest.RandomDeviceHashesMetadata(test.AllowOptional())
						object := dataWorkTest.NewObjectFromDeviceHashesMetadata(expectedDatum, test.ObjectFormatJSON)
						mutator(object, expectedDatum)
						result := &dataWork.DeviceHashesMetadata{}
						errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
						Expect(result).To(Equal(expectedDatum))
					},
					Entry("succeeds",
						func(object map[string]any, expectedDatum *dataWork.DeviceHashesMetadata) {},
					),
					Entry("empty",
						func(object map[string]any, expectedDatum *dataWork.DeviceHashesMetadata) {
							clear(object)
							*expectedDatum = dataWork.DeviceHashesMetadata{}
						},
					),
					Entry("multiple errors",
						func(object map[string]any, expectedDatum *dataWork.DeviceHashesMetadata) {
							object["deviceHashes"] = true
							expectedDatum.DeviceHashes = nil
						},
						errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/deviceHashes"),
					),
				)
			})

			Context("Validate", func() {
				DescribeTable("validates the datum",
					func(mutator func(datum *dataWork.DeviceHashesMetadata), expectedErrors ...error) {
						datum := dataWorkTest.RandomDeviceHashesMetadata(test.AllowOptional())
						mutator(datum)
						errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
					},
					Entry("succeeds",
						func(datum *dataWork.DeviceHashesMetadata) {},
					),
					Entry("device hashes missing",
						func(datum *dataWork.DeviceHashesMetadata) {
							datum.DeviceHashes = nil
						},
					),
					Entry("device hashes invalid",
						func(datum *dataWork.DeviceHashesMetadata) {
							deviceHashes := dataWork.DeviceHashes{}
							deviceHashes[""] = dataWorkTest.RandomDeviceHash()
							datum.DeviceHashes = &deviceHashes
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/deviceHashes/#"),
					),
					Entry("device hashes valid",
						func(datum *dataWork.DeviceHashesMetadata) {
							datum.DeviceHashes = dataWorkTest.RandomDeviceHashes()
						},
					),
					Entry("multiple errors",
						func(datum *dataWork.DeviceHashesMetadata) {
							deviceHashes := dataWork.DeviceHashes{}
							deviceHashes[""] = dataWorkTest.RandomDeviceHash()
							datum.DeviceHashes = &deviceHashes
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/deviceHashes/#"),
					),
				)
			})
		})
	})
})
