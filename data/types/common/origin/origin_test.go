package origin_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/common/origin"
	testDataTypesCommonOrigin "github.com/tidepool-org/platform/data/types/common/origin/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Origin", func() {
	It("IDLengthMaximum is expected", func() {
		Expect(origin.IDLengthMaximum).To(Equal(100))
	})

	It("NameLengthMaximum is expected", func() {
		Expect(origin.NameLengthMaximum).To(Equal(100))
	})

	It("TimeFormat is expected", func() {
		Expect(origin.TimeFormat).To(Equal(time.RFC3339Nano))
	})

	It("TypeDevice is expected", func() {
		Expect(origin.TypeDevice).To(Equal("device"))
	})

	It("TypeManual is expected", func() {
		Expect(origin.TypeManual).To(Equal("manual"))
	})

	It("TypeService is expected", func() {
		Expect(origin.TypeService).To(Equal("service"))
	})

	It("VersionLengthMaximum is expected", func() {
		Expect(origin.VersionLengthMaximum).To(Equal(100))
	})

	It("Types returns expected", func() {
		Expect(origin.Types()).To(Equal([]string{"device", "manual", "service"}))
	})

	Context("ParseOrigin", func() {
		// TODO
	})

	Context("NewOrigin", func() {
		It("is successful", func() {
			Expect(origin.NewOrigin()).To(Equal(&origin.Origin{}))
		})
	})

	Context("Origin", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *origin.Origin), expectedErrors ...error) {
					datum := testDataTypesCommonOrigin.NewOrigin()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *origin.Origin) {},
				),
				Entry("id missing",
					func(datum *origin.Origin) { datum.ID = nil },
				),
				Entry("id empty",
					func(datum *origin.Origin) { datum.ID = pointer.FromString("") },
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/id"),
				),
				Entry("id length; in range (upper)",
					func(datum *origin.Origin) { datum.ID = pointer.FromString(test.NewText(100, 100)) },
				),
				Entry("id length; out of range (upper)",
					func(datum *origin.Origin) { datum.ID = pointer.FromString(test.NewText(101, 101)) },
					testErrors.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/id"),
				),
				Entry("name missing",
					func(datum *origin.Origin) { datum.Name = nil },
				),
				Entry("name empty",
					func(datum *origin.Origin) { datum.Name = pointer.FromString("") },
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
				),
				Entry("name length; in range (upper)",
					func(datum *origin.Origin) { datum.Name = pointer.FromString(test.NewText(100, 100)) },
				),
				Entry("name length; out of range (upper)",
					func(datum *origin.Origin) { datum.Name = pointer.FromString(test.NewText(101, 101)) },
					testErrors.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/name"),
				),
				Entry("payload missing",
					func(datum *origin.Origin) { datum.Payload = nil },
				),
				Entry("payload exists",
					func(datum *origin.Origin) { datum.Payload = dataTest.NewBlob() },
				),
				Entry("time missing",
					func(datum *origin.Origin) { datum.Time = nil },
				),
				Entry("time empty",
					func(datum *origin.Origin) { datum.Time = pointer.FromString("") },
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("", time.RFC3339Nano), "/time"),
				),
				Entry("time zero",
					func(datum *origin.Origin) { datum.Time = pointer.FromString(time.Time{}.Format(time.RFC3339Nano)) },
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/time"),
				),
				Entry("time not zero",
					func(datum *origin.Origin) {
						datum.Time = pointer.FromString(test.RandomTime().Format(time.RFC3339Nano))
					},
				),
				Entry("type missing",
					func(datum *origin.Origin) { datum.Type = nil },
				),
				Entry("type invalid",
					func(datum *origin.Origin) { datum.Type = pointer.FromString("invalid") },
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"device", "manual", "service"}), "/type"),
				),
				Entry("type device",
					func(datum *origin.Origin) { datum.Type = pointer.FromString("device") },
				),
				Entry("type manual",
					func(datum *origin.Origin) { datum.Type = pointer.FromString("manual") },
				),
				Entry("type service",
					func(datum *origin.Origin) { datum.Type = pointer.FromString("service") },
				),
				Entry("version missing",
					func(datum *origin.Origin) { datum.Version = nil },
				),
				Entry("version empty",
					func(datum *origin.Origin) { datum.Version = pointer.FromString("") },
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/version"),
				),
				Entry("version length; in range (upper)",
					func(datum *origin.Origin) { datum.Version = pointer.FromString(test.NewText(100, 100)) },
				),
				Entry("version length; out of range (upper)",
					func(datum *origin.Origin) { datum.Version = pointer.FromString(test.NewText(101, 101)) },
					testErrors.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/version"),
				),
				Entry("multiple errors",
					func(datum *origin.Origin) {
						datum.ID = pointer.FromString("")
						datum.Name = pointer.FromString("")
						datum.Time = pointer.FromString("")
						datum.Type = pointer.FromString("invalid")
						datum.Version = pointer.FromString("")
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/id"),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("", time.RFC3339Nano), "/time"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"device", "manual", "service"}), "/type"),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/version"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *origin.Origin)) {
					for _, origin := range structure.Origins() {
						datum := testDataTypesCommonOrigin.NewOrigin()
						mutator(datum)
						expectedDatum := testDataTypesCommonOrigin.CloneOrigin(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *origin.Origin) {},
				),
				Entry("does not modify the datum; id missing",
					func(datum *origin.Origin) { datum.ID = nil },
				),
				Entry("does not modify the datum; name missing",
					func(datum *origin.Origin) { datum.Name = nil },
				),
				Entry("does not modify the datum; payload missing",
					func(datum *origin.Origin) { datum.Payload = nil },
				),
				Entry("does not modify the datum; time missing",
					func(datum *origin.Origin) { datum.Time = nil },
				),
				Entry("does not modify the datum; type missing",
					func(datum *origin.Origin) { datum.Type = nil },
				),
				Entry("does not modify the datum; version missing",
					func(datum *origin.Origin) { datum.Version = nil },
				),
				Entry("does not modify the datum; all missing",
					func(datum *origin.Origin) { *datum = origin.Origin{} },
				),
			)
		})
	})
})
