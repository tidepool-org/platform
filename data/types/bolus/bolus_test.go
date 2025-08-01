package bolus_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypes "github.com/tidepool-org/platform/data/types"
	dataTypesBolus "github.com/tidepool-org/platform/data/types/bolus"
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
		Expect(dataTypesBolus.Type).To(Equal("bolus"))
	})

	It("DeliveryContextAlgorithm is expected", func() {
		Expect(dataTypesBolus.DeliveryContextAlgorithm).To(Equal("algorithm"))
	})

	It("DeliveryContextDevice is expected", func() {
		Expect(dataTypesBolus.DeliveryContextDevice).To(Equal("device"))
	})

	It("DeliveryContextOneButton is expected", func() {
		Expect(dataTypesBolus.DeliveryContextOneButton).To(Equal("oneButton"))
	})

	It("DeliveryContextRemote is expected", func() {
		Expect(dataTypesBolus.DeliveryContextRemote).To(Equal("remote"))
	})

	It("DeliveryContextUndetermined is expected", func() {
		Expect(dataTypesBolus.DeliveryContextUndetermined).To(Equal("undetermined"))
	})

	It("DeliveryContextWatch is expected", func() {
		Expect(dataTypesBolus.DeliveryContextWatch).To(Equal("watch"))
	})

	It("DeliveryContexts returns expected", func() {
		Expect(dataTypesBolus.DeliveryContexts()).To(ConsistOf([]string{
			dataTypesBolus.DeliveryContextAlgorithm,
			dataTypesBolus.DeliveryContextDevice,
			dataTypesBolus.DeliveryContextOneButton,
			dataTypesBolus.DeliveryContextRemote,
			dataTypesBolus.DeliveryContextUndetermined,
			dataTypesBolus.DeliveryContextWatch,
		}))
	})

	Context("New", func() {
		It("creates a new datum with all values initialized", func() {
			subType := dataTypesTest.NewType()
			datum := dataTypesBolus.New(subType)
			Expect(datum.Type).To(Equal("bolus"))
			Expect(datum.SubType).To(Equal(subType))
			Expect(datum.DeliveryContext).To(BeNil())
			Expect(datum.InsulinFormulation).To(BeNil())
		})
	})

	Context("with new datum", func() {
		var subType string
		var datum dataTypesBolus.Bolus

		BeforeEach(func() {
			subType = dataTypesTest.NewType()
			datum = dataTypesBolus.New(subType)
		})

		Context("Meta", func() {
			It("returns the meta with delivery type", func() {
				Expect(datum.Meta()).To(Equal(&dataTypesBolus.Meta{Type: "bolus", SubType: subType}))
			})
		})
	})

	Context("Bolus", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesBolus.Bolus), expectedErrors ...error) {
					datum := dataTypesBolusTest.RandomBolus()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesBolus.Bolus) {},
				),
				Entry("type missing",
					func(datum *dataTypesBolus.Bolus) { datum.Type = "" },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/type"),
				),
				Entry("type invalid",
					func(datum *dataTypesBolus.Bolus) { datum.Type = "invalid" },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", "bolus"), "/type"),
				),
				Entry("type bolus",
					func(datum *dataTypesBolus.Bolus) { datum.Type = "bolus" },
				),
				Entry("sub type missing",
					func(datum *dataTypesBolus.Bolus) { datum.SubType = "" },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/subType"),
				),
				Entry("sub type valid",
					func(datum *dataTypesBolus.Bolus) { datum.SubType = dataTypesTest.NewType() },
				),
				Entry("delivery context missing",
					func(datum *dataTypesBolus.Bolus) { datum.DeliveryContext = nil },
				),
				Entry("delivery context invalid",
					func(datum *dataTypesBolus.Bolus) { datum.DeliveryContext = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dataTypesBolus.DeliveryContexts()), "/deliveryContext"),
				),
				Entry("delivery context algorithm",
					func(datum *dataTypesBolus.Bolus) {
						datum.DeliveryContext = pointer.FromString(dataTypesBolus.DeliveryContextAlgorithm)
					},
				),
				Entry("delivery context device",
					func(datum *dataTypesBolus.Bolus) {
						datum.DeliveryContext = pointer.FromString(dataTypesBolus.DeliveryContextDevice)
					},
				),
				Entry("delivery context one button",
					func(datum *dataTypesBolus.Bolus) {
						datum.DeliveryContext = pointer.FromString(dataTypesBolus.DeliveryContextOneButton)
					},
				),
				Entry("delivery context remote",
					func(datum *dataTypesBolus.Bolus) {
						datum.DeliveryContext = pointer.FromString(dataTypesBolus.DeliveryContextRemote)
					},
				),
				Entry("delivery context undetermined",
					func(datum *dataTypesBolus.Bolus) {
						datum.DeliveryContext = pointer.FromString(dataTypesBolus.DeliveryContextUndetermined)
					},
				),
				Entry("delivery context watch",
					func(datum *dataTypesBolus.Bolus) {
						datum.DeliveryContext = pointer.FromString(dataTypesBolus.DeliveryContextWatch)
					},
				),
				Entry("insulin formulation missing",
					func(datum *dataTypesBolus.Bolus) { datum.InsulinFormulation = nil },
				),
				Entry("insulin formulation invalid",
					func(datum *dataTypesBolus.Bolus) {
						datum.InsulinFormulation.Compounds = nil
						datum.InsulinFormulation.Name = nil
						datum.InsulinFormulation.Simple = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/insulinFormulation/simple"),
				),
				Entry("insulin formulation valid",
					func(datum *dataTypesBolus.Bolus) {
						datum.InsulinFormulation = dataTypesInsulinTest.RandomFormulation(3)
					},
				),
				Entry("multiple errors",
					func(datum *dataTypesBolus.Bolus) {
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
				func(mutator func(datum *dataTypesBolus.Bolus)) {
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
					func(datum *dataTypesBolus.Bolus) {},
				),
				Entry("does not modify the datum; type missing",
					func(datum *dataTypesBolus.Bolus) { datum.Type = "" },
				),
				Entry("does not modify the datum; sub type missing",
					func(datum *dataTypesBolus.Bolus) { datum.SubType = "" },
				),
				Entry("does not modify the datum; insulin formulation missing",
					func(datum *dataTypesBolus.Bolus) { datum.InsulinFormulation = nil },
				),
			)
		})

		Context("IdentityFields", func() {
			var datumBolus *dataTypesBolus.Bolus
			var datum data.Datum
			var version string

			BeforeEach(func() {
				datumBolus = dataTypesBolusTest.RandomBolus()
				datum = datumBolus
			})

			identityFieldsAssertions := func() {
				It("returns error if user id is missing", func() {
					datumBolus.UserID = nil
					identityFields, err := datum.IdentityFields(version)
					Expect(err).To(MatchError("user id is missing"))
					Expect(identityFields).To(BeEmpty())
				})

				It("returns error if user id is empty", func() {
					datumBolus.UserID = pointer.FromString("")
					identityFields, err := datum.IdentityFields(version)
					Expect(err).To(MatchError("user id is empty"))
					Expect(identityFields).To(BeEmpty())
				})

				It("returns error if sub type is empty", func() {
					datumBolus.SubType = ""
					identityFields, err := datum.IdentityFields(version)
					Expect(err).To(MatchError("sub type is empty"))
					Expect(identityFields).To(BeEmpty())
				})
			}

			When("version is IdentityFieldsVersionDefault", func() {
				BeforeEach(func() {
					version = dataTypes.IdentityFieldsVersionDeviceID
				})

				identityFieldsAssertions()

				It("returns the expected identity fields", func() {
					identityFields, err := datum.IdentityFields(version)
					Expect(err).ToNot(HaveOccurred())
					Expect(identityFields).To(Equal([]string{*datumBolus.UserID, *datumBolus.DeviceID, (*datumBolus.Time).Format(ExpectedTimeFormat), datumBolus.Type, datumBolus.SubType}))
				})
			})

			When("version is IdentityFieldsVersionDataSetID", func() {
				BeforeEach(func() {
					version = dataTypes.IdentityFieldsVersionDataSetID
				})

				identityFieldsAssertions()

				It("returns the expected identity fields", func() {
					identityFields, err := datum.IdentityFields(version)
					Expect(err).ToNot(HaveOccurred())
					Expect(identityFields).To(Equal([]string{*datumBolus.UserID, *datumBolus.UploadID, (*datumBolus.Time).Format(ExpectedTimeFormat), datumBolus.Type, datumBolus.SubType}))
				})
			})

			When("version is invalid", func() {
				It("returns an error", func() {
					identityFields, err := datum.IdentityFields("invalid")
					Expect(err).To(MatchError("version is invalid"))
					Expect(identityFields).To(BeEmpty())
				})
			})
		})
	})
})
