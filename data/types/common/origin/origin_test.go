package origin_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"time"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/common/origin"
	testDataTypesCommonOrigin "github.com/tidepool-org/platform/data/types/common/origin/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	testInternet "github.com/tidepool-org/platform/test/internet"
	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("Origin", func() {
	It("IDLengthMaximum is expected", func() {
		Expect(origin.IDLengthMaximum).To(Equal(100))
	})

	It("TimeFormat is expected", func() {
		Expect(origin.TimeFormat).To(Equal(time.RFC3339))
	})

	It("TypeDevice is expected", func() {
		Expect(origin.TypeDevice).To(Equal("device"))
	})

	It("TypeManual is expected", func() {
		Expect(origin.TypeManual).To(Equal("manual"))
	})

	It("VersionLengthMaximum is expected", func() {
		Expect(origin.VersionLengthMaximum).To(Equal(100))
	})

	It("Types returns expected", func() {
		Expect(origin.Types()).To(Equal([]string{"device", "manual"}))
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
					func(datum *origin.Origin) { datum.ID = pointer.String("") },
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/id"),
				),
				Entry("id length; in range (upper)",
					func(datum *origin.Origin) { datum.ID = pointer.String(test.NewText(100, 100)) },
				),
				Entry("id length; out of range (upper)",
					func(datum *origin.Origin) { datum.ID = pointer.String(test.NewText(101, 101)) },
					testErrors.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/id"),
				),
				Entry("name missing",
					func(datum *origin.Origin) { datum.Name = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/name"),
				),
				Entry("name empty",
					func(datum *origin.Origin) { datum.Name = pointer.String("") },
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
				),
				Entry("name invalid",
					func(datum *origin.Origin) { datum.Name = pointer.String("org") },
					testErrors.WithPointerSource(validate.ErrorValueStringAsReverseDomainNotValid("org"), "/name"),
				),
				Entry("name valid",
					func(datum *origin.Origin) { datum.Name = pointer.String(testInternet.NewReverseDomain()) },
				),
				Entry("time missing",
					func(datum *origin.Origin) { datum.Time = nil },
				),
				Entry("time zero",
					func(datum *origin.Origin) { datum.Time = pointer.Time(time.Time{}) },
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeZero(), "/time"),
				),
				Entry("time not zero",
					func(datum *origin.Origin) { datum.Time = pointer.Time(test.NewTime()) },
				),
				Entry("type missing",
					func(datum *origin.Origin) { datum.Type = nil },
				),
				Entry("type empty",
					func(datum *origin.Origin) { datum.Type = pointer.String("") },
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", []string{"device", "manual"}), "/type"),
				),
				Entry("type invalid",
					func(datum *origin.Origin) { datum.Type = pointer.String("invalid") },
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"device", "manual"}), "/type"),
				),
				Entry("type device",
					func(datum *origin.Origin) { datum.Type = pointer.String("device") },
				),
				Entry("type manual",
					func(datum *origin.Origin) { datum.Type = pointer.String("manual") },
				),
				Entry("version missing",
					func(datum *origin.Origin) { datum.Version = nil },
				),
				Entry("version empty",
					func(datum *origin.Origin) { datum.Version = pointer.String("") },
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/version"),
				),
				Entry("version length; in range (upper)",
					func(datum *origin.Origin) { datum.Version = pointer.String(test.NewText(100, 100)) },
				),
				Entry("version length; out of range (upper)",
					func(datum *origin.Origin) { datum.Version = pointer.String(test.NewText(101, 101)) },
					testErrors.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/version"),
				),
				Entry("multiple errors",
					func(datum *origin.Origin) {
						datum.ID = pointer.String("")
						datum.Name = nil
						datum.Time = pointer.Time(time.Time{})
						datum.Type = pointer.String("")
						datum.Version = pointer.String("")
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/id"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/name"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeZero(), "/time"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", []string{"device", "manual"}), "/type"),
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
