package bolus_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"time"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/bolus"
	dataTypesBolusTest "github.com/tidepool-org/platform/data/types/bolus/test"
	dataTypesInsulinTest "github.com/tidepool-org/platform/data/types/insulin/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

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
					datum := dataTypesBolusTest.NewBolus()
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
					func(datum *bolus.Bolus) { datum.InsulinFormulation = dataTypesInsulinTest.NewFormulation(3) },
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
						datum := dataTypesBolusTest.NewBolus()
						mutator(datum)
						expectedDatum := dataTypesBolusTest.CloneBolus(datum)
						normalizer := dataNormalizer.New()
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
			var datum *bolus.Bolus

			BeforeEach(func() {
				datum = dataTypesBolusTest.NewBolus()
			})

			It("returns error if user id is missing", func() {
				datum.UserID = nil
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("user id is missing"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if user id is empty", func() {
				datum.UserID = pointer.FromString("")
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("user id is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if sub type is empty", func() {
				datum.SubType = ""
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("sub type is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns the expected identity fields", func() {
				identityFields, err := datum.IdentityFields()
				Expect(err).ToNot(HaveOccurred())
				Expect(identityFields).To(Equal([]string{*datum.UserID, *datum.DeviceID, (*datum.Time).Format(time.RFC3339Nano), datum.Type, datum.SubType}))
			})
		})
	})
})
