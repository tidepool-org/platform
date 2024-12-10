package bolus_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/bolus"
	dataTypesBolusTest "github.com/tidepool-org/platform/data/types/bolus/test"
	dataTypesInsulinTest "github.com/tidepool-org/platform/data/types/insulin/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const ExpectedTimeFormat = time.RFC3339Nano

var _ = Describe("Bolus", func() {
	It("Type is expected", func() {
		Expect(bolus.Type).To(Equal("bolus"))
	})

	Context("New", func() {
		It("creates a new datum with all values initialized", func() {
			subType := dataTypesTest.NewType()
			datum := bolus.New(subType)
			Expect(datum.Type).To(Equal("bolus"))
			Expect(datum.SubType).To(Equal(subType))
			Expect(datum.InsulinFormulation).To(BeNil())
			Expect(datum.DeliveryContext).To(BeNil())
		})
	})

	Context("with new datum", func() {
		var subType string
		var datum bolus.Bolus

		BeforeEach(func() {
			subType = dataTypesTest.NewType()
			datum = bolus.New(subType)
		})

		Context("Meta", func() {
			It("returns the meta with delivery type", func() {
				Expect(datum.Meta()).To(Equal(&bolus.Meta{Type: "bolus", SubType: subType}))
			})
		})
	})

	Context("Bolus", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *bolus.Bolus), expectedErrors ...error) {
					datum := dataTypesBolusTest.RandomBolus()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *bolus.Bolus) {},
				),
				Entry("type missing",
					func(datum *bolus.Bolus) { datum.Type = "" },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/type"),
				),
				Entry("type invalid",
					func(datum *bolus.Bolus) { datum.Type = "invalid" },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", "bolus"), "/type"),
				),
				Entry("type bolus",
					func(datum *bolus.Bolus) { datum.Type = "bolus" },
				),
				Entry("sub type missing",
					func(datum *bolus.Bolus) { datum.SubType = "" },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/subType"),
				),
				Entry("sub type valid",
					func(datum *bolus.Bolus) { datum.SubType = dataTypesTest.NewType() },
				),
				Entry("delivery context missing",
					func(datum *bolus.Bolus) { datum.DeliveryContext = nil },
				),
				Entry("delivery context invalid",
					func(datum *bolus.Bolus) { datum.DeliveryContext = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", bolus.DeliveryContexts()), "/deliveryContext"),
				),
				Entry("delivery context valid",
					func(datum *bolus.Bolus) { datum.DeliveryContext = pointer.FromString(bolus.DeliveryContextAlgorithm) },
				),
				Entry("insulin formulation missing",
					func(datum *bolus.Bolus) { datum.InsulinFormulation = nil },
				),
				Entry("insulin formulation invalid",
					func(datum *bolus.Bolus) {
						datum.InsulinFormulation.Compounds = nil
						datum.InsulinFormulation.Name = nil
						datum.InsulinFormulation.Simple = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/insulinFormulation/simple"),
				),
				Entry("insulin formulation valid",
					func(datum *bolus.Bolus) { datum.InsulinFormulation = dataTypesInsulinTest.RandomFormulation(3) },
				),
				Entry("multiple errors",
					func(datum *bolus.Bolus) {
						datum.Type = "invalid"
						datum.SubType = ""
						datum.InsulinFormulation.Compounds = nil
						datum.InsulinFormulation.Name = nil
						datum.InsulinFormulation.Simple = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", "bolus"), "/type"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/subType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/insulinFormulation/simple"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *bolus.Bolus)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesBolusTest.RandomBolus()
						mutator(datum)
						expectedDatum := dataTypesBolusTest.CloneBolus(datum)
						normalizer := dataNormalizer.New(logTest.NewLogger())
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *bolus.Bolus) {},
				),
				Entry("does not modify the datum; type missing",
					func(datum *bolus.Bolus) { datum.Type = "" },
				),
				Entry("does not modify the datum; sub type missing",
					func(datum *bolus.Bolus) { datum.SubType = "" },
				),
				Entry("does not modify the datum; insulin formulation missing",
					func(datum *bolus.Bolus) { datum.InsulinFormulation = nil },
				),
			)
		})

		Context("IdentityFields", func() {
			var datumBolus *bolus.Bolus
			var datum data.Datum

			BeforeEach(func() {
				datumBolus = dataTypesBolusTest.RandomBolus()
				datum = datumBolus
			})

			It("returns error if user id is missing", func() {
				datumBolus.UserID = nil
				identityFields, err := datum.IdentityFields(types.IdentityFieldsVersion)
				Expect(err).To(MatchError("user id is missing"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if user id is empty", func() {
				datumBolus.UserID = pointer.FromString("")
				identityFields, err := datum.IdentityFields(types.IdentityFieldsVersion)
				Expect(err).To(MatchError("user id is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if sub type is empty", func() {
				datumBolus.SubType = ""
				identityFields, err := datum.IdentityFields(types.IdentityFieldsVersion)
				Expect(err).To(MatchError("sub type is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns the expected identity fields", func() {
				identityFields, err := datum.IdentityFields(types.IdentityFieldsVersion)
				Expect(err).ToNot(HaveOccurred())
				Expect(identityFields).To(Equal([]string{*datumBolus.UserID, *datumBolus.DeviceID, (*datumBolus.Time).Format(ExpectedTimeFormat), datumBolus.Type, datumBolus.SubType}))
			})
		})

		Context("Legacy IdentityFields", func() {
			var datum *bolus.Bolus

			BeforeEach(func() {
				datum = dataTypesBolusTest.RandomBolus()
			})

			It("returns error if sub type is empty", func() {
				datum.SubType = ""
				identityFields, err := datum.IdentityFields(types.LegacyIdentityFieldsVersion)
				Expect(err).To(MatchError("sub type is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns the expected legacy identity fields", func() {
				datum.DeviceID = pointer.FromString("some-device")
				t, err := time.Parse(types.TimeFormat, "2023-05-13T15:51:58Z")
				Expect(err).ToNot(HaveOccurred())
				datum.Time = pointer.FromTime(t)
				datum.SubType = "some-sub-type"
				legacyIdentityFields, err := datum.IdentityFields(types.LegacyIdentityFieldsVersion)
				Expect(err).ToNot(HaveOccurred())
				Expect(legacyIdentityFields).To(Equal([]string{"bolus", "some-sub-type", "some-device", "2023-05-13T15:51:58.000Z"}))
			})
		})
	})
})
