package timechange_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"time"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypesDeviceTimechange "github.com/tidepool-org/platform/data/types/device/timechange"
	dataTypesDeviceTimechangeTest "github.com/tidepool-org/platform/data/types/device/timechange/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	timeZone "github.com/tidepool-org/platform/time/zone"
	timeZoneTest "github.com/tidepool-org/platform/time/zone/test"
)

var _ = Describe("Info", func() {
	It("InfoTimeFormat is expected", func() {
		Expect(dataTypesDeviceTimechange.InfoTimeFormat).To(Equal("2006-01-02T15:04:05"))
	})

	Context("ParseInfo", func() {
		// TODO
	})

	Context("NewInfo", func() {
		It("is successful", func() {
			Expect(dataTypesDeviceTimechange.NewInfo()).To(Equal(&dataTypesDeviceTimechange.Info{}))
		})
	})

	Context("Info", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesDeviceTimechange.Info), expectedErrors ...error) {
					datum := dataTypesDeviceTimechangeTest.RandomInfo()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesDeviceTimechange.Info) {},
				),
				Entry("time missing",
					func(datum *dataTypesDeviceTimechange.Info) { datum.Time = nil },
				),
				Entry("time invalid",
					func(datum *dataTypesDeviceTimechange.Info) { datum.Time = pointer.FromTime(time.Time{}) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/time"),
				),
				Entry("time valid",
					func(datum *dataTypesDeviceTimechange.Info) { datum.Time = pointer.FromTime(test.RandomTime()) },
				),
				Entry("time zone name missing",
					func(datum *dataTypesDeviceTimechange.Info) { datum.TimeZoneName = nil },
				),
				Entry("time zone name empty",
					func(datum *dataTypesDeviceTimechange.Info) { datum.TimeZoneName = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/timeZoneName"),
				),
				Entry("time zone name invalid",
					func(datum *dataTypesDeviceTimechange.Info) { datum.TimeZoneName = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(timeZone.ErrorValueStringAsNameNotValid("invalid"), "/timeZoneName"),
				),
				Entry("time zone name valid",
					func(datum *dataTypesDeviceTimechange.Info) { pointer.FromString(timeZoneTest.RandomName()) },
				),
				Entry("time and time zone name missing",
					func(datum *dataTypesDeviceTimechange.Info) {
						datum.Time = nil
						datum.TimeZoneName = nil
					},
					structureValidator.ErrorValuesNotExistForAny("time", "timeZoneName"),
				),
				Entry("multiple errors",
					func(datum *dataTypesDeviceTimechange.Info) {
						datum.Time = pointer.FromTime(time.Time{})
						datum.TimeZoneName = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/time"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/timeZoneName"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *dataTypesDeviceTimechange.Info)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesDeviceTimechangeTest.RandomInfo()
						mutator(datum)
						expectedDatum := dataTypesDeviceTimechangeTest.CloneInfo(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *dataTypesDeviceTimechange.Info) {},
				),
				Entry("does not modify the datum; time missing",
					func(datum *dataTypesDeviceTimechange.Info) { datum.Time = nil },
				),
				Entry("does not modify the datum; time zone name missing",
					func(datum *dataTypesDeviceTimechange.Info) { datum.TimeZoneName = nil },
				),
			)
		})
	})
})
