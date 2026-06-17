package insulin_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/insulin"
	dataTypesInsulinTest "github.com/tidepool-org/platform/data/types/insulin/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &types.Meta{
		Type: "insulin",
	}
}

func NewInsulin() *insulin.Insulin {
	datum := insulin.New()
	datum.Base = *dataTypesTest.RandomBase()
	datum.Type = "insulin"
	datum.Dose = dataTypesInsulinTest.NewDose()
	datum.Formulation = dataTypesInsulinTest.RandomFormulation(3)
	datum.Site = pointer.FromString(test.RandomStringFromRange(1, 100))
	return datum
}

func CloneInsulin(datum *insulin.Insulin) *insulin.Insulin {
	if datum == nil {
		return nil
	}
	clone := insulin.New()
	clone.Base = *dataTypesTest.CloneBase(&datum.Base)
	clone.Dose = dataTypesInsulinTest.CloneDose(datum.Dose)
	clone.Formulation = dataTypesInsulinTest.CloneFormulation(datum.Formulation)
	clone.Site = pointer.CloneString(datum.Site)
	return clone
}

var _ = Describe("Insulin", func() {
	It("Type is expected", func() {
		Expect(insulin.Type).To(Equal("insulin"))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := insulin.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("insulin"))
			Expect(datum.Dose).To(BeNil())
			Expect(datum.Formulation).To(BeNil())
			Expect(datum.Site).To(BeNil())
		})
	})

	Context("Insulin", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *insulin.Insulin), expectedErrors ...error) {
					datum := NewInsulin()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *insulin.Insulin) {},
				),
				Entry("type missing",
					func(datum *insulin.Insulin) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
				),
				Entry("type invalid",
					func(datum *insulin.Insulin) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "insulin"), "/type", &types.Meta{Type: "invalidType"}),
				),
				Entry("type insulin",
					func(datum *insulin.Insulin) { datum.Type = "insulin" },
				),
				Entry("dose missing",
					func(datum *insulin.Insulin) { datum.Dose = nil },
				),
				Entry("dose invalid",
					func(datum *insulin.Insulin) { datum.Dose.Total = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/dose/total", NewMeta()),
				),
				Entry("dose valid",
					func(datum *insulin.Insulin) { datum.Dose = dataTypesInsulinTest.NewDose() },
				),
				Entry("formulation missing",
					func(datum *insulin.Insulin) { datum.Formulation = nil },
				),
				Entry("formulation invalid",
					func(datum *insulin.Insulin) {
						datum.Formulation.Name = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/formulation/name", NewMeta()),
				),
				Entry("formulation valid",
					func(datum *insulin.Insulin) { datum.Formulation = dataTypesInsulinTest.RandomFormulation(3) },
				),
				Entry("site missing",
					func(datum *insulin.Insulin) { datum.Site = nil },
				),
				Entry("site empty",
					func(datum *insulin.Insulin) { datum.Site = pointer.FromString("") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/site", NewMeta()),
				),
				Entry("site invalid",
					func(datum *insulin.Insulin) { datum.Site = pointer.FromString(test.RandomStringFromRange(101, 101)) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/site", NewMeta()),
				),
				Entry("site valid",
					func(datum *insulin.Insulin) { datum.Site = pointer.FromString(test.RandomStringFromRange(1, 100)) },
				),
				Entry("multiple errors",
					func(datum *insulin.Insulin) {
						datum.Type = "invalidType"
						datum.Dose.Total = nil
						datum.Formulation.Name = pointer.FromString("")
						datum.Site = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "insulin"), "/type", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/dose/total", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/formulation/name", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/site", &types.Meta{Type: "invalidType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *insulin.Insulin)) {
					for _, origin := range structure.Origins() {
						datum := NewInsulin()
						mutator(datum)
						expectedDatum := CloneInsulin(datum)
						normalizer := dataNormalizer.New(logTest.NewLogger())
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *insulin.Insulin) {},
				),
				Entry("does not modify the datum; dose nil",
					func(datum *insulin.Insulin) { datum.Dose = nil },
				),
				Entry("does not modify the datum; formulation nil",
					func(datum *insulin.Insulin) { datum.Formulation = nil },
				),
				Entry("does not modify the datum; site nil",
					func(datum *insulin.Insulin) { datum.Site = nil },
				),
			)
		})

		Context("Legacy IdentityFields", func() {
			It("returns the expected legacy identity fields", func() {
				datum := NewInsulin()
				datum.DeviceID = pointer.FromString("some-pump-device")
				t, err := time.Parse(types.TimeFormat, "2023-05-13T15:51:58Z")
				Expect(err).ToNot(HaveOccurred())
				datum.Time = pointer.FromTime(t)
				legacyIdentityFields, err := datum.IdentityFields(types.LegacyIdentityFieldsVersion)
				Expect(err).ToNot(HaveOccurred())
				Expect(legacyIdentityFields).To(Equal([]string{"insulin", "some-pump-device", "2023-05-13T15:51:58.000Z"}))
			})
		})
	})
})
