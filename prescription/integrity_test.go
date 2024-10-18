package prescription_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/tidepool-org/platform/data/blood/glucose"

	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/prescription"
	"github.com/tidepool-org/platform/prescription/test"
)

var haveDataAttributes = MatchAllFields(Fields{
	"AccountType":             Ignore(),
	"CaregiverFirstName":      Ignore(),
	"CaregiverLastName":       Ignore(),
	"FirstName":               Ignore(),
	"LastName":                Ignore(),
	"Birthday":                Ignore(),
	"MRN":                     Ignore(),
	"Email":                   Ignore(),
	"Sex":                     Ignore(),
	"Weight":                  Ignore(),
	"YearOfDiagnosis":         Ignore(),
	"PhoneNumber":             Ignore(),
	"InitialSettings":         Ignore(),
	"Calculator":              Ignore(),
	"Training":                Ignore(),
	"TherapySettings":         Ignore(),
	"PrescriberTermsAccepted": Ignore(),
	"State":                   Ignore(),
})

var _ = Describe("Integrity hash", func() {
	Context("NewIntegrityAttributesFromRevisionCreate", func() {
		var revisionCreate *prescription.RevisionCreate

		BeforeEach(func() {
			revisionCreate = test.RandomRevisionCreate()
		})

		It("sets all integrity attributes correctly", func() {
			attrs := prescription.NewIntegrityAttributesFromRevisionCreate(*revisionCreate)
			Expect(attrs).To(haveDataAttributes)
		})
	})

	Context("NewIntegrityAttributesFromRevision", func() {
		var revision *prescription.Revision

		BeforeEach(func() {
			revision = test.RandomRevision()
		})

		It("sets all integrity attributes correctly", func() {
			attrs := prescription.NewIntegrityAttributesFromRevision(*revision)
			Expect(attrs).To(haveDataAttributes)
		})
	})

	Context("MustGenerateIntegrityHash", func() {
		var revision *prescription.Revision
		var attrs prescription.DataAttributes

		BeforeEach(func() {
			revision = test.RandomRevision()
		})

		It("generates hash with the expected algorithm", func() {
			attrs = prescription.NewIntegrityAttributesFromRevision(*revision)
			hash := prescription.MustGenerateIntegrityHash(attrs)
			Expect(hash.Algorithm).To(Equal("JCSSHA512"))
		})

		It("generates hash with length of 128", func() {
			attrs = prescription.NewIntegrityAttributesFromRevision(*revision)
			hash := prescription.MustGenerateIntegrityHash(attrs)
			Expect(hash.Hash).To(HaveLen(128))
		})

		It("generates a different hash for different attributes", func() {
			attrs = prescription.NewIntegrityAttributesFromRevision(*revision)
			first := prescription.MustGenerateIntegrityHash(attrs)
			revision = test.RandomRevision()
			attrs = prescription.NewIntegrityAttributesFromRevision(*revision)
			second := prescription.MustGenerateIntegrityHash(attrs)
			Expect(first.Hash).ToNot(Equal(second.Hash))
		})

		It("generates correct hash for the fixture", func() {
			attrs = test.IntegrityAttributes
			hash := prescription.MustGenerateIntegrityHash(attrs)
			Expect(hash.Algorithm).To(Equal("JCSSHA512"))
			Expect(hash.Hash).To(Equal(test.ExpectedHash))
		})

		Context("Integrity Attributes", func() {
			It("are up to date with the integrity test fixture", func() {
				// If this test fails, the struct keys here must be updated,
				// as well as the fixture and the expected hash in test/integrity.go
				attrs = test.IntegrityAttributes
				Expect(attrs).To(haveDataAttributes)
				Expect(*attrs.Weight).To(MatchAllFields(Fields{
					"Value": Ignore(),
					"Units": Ignore(),
				}))
				Expect(*attrs.Calculator).To(MatchAllFields(Fields{
					"Method":                        Ignore(),
					"RecommendedBasalRate":          Ignore(),
					"RecommendedCarbohydrateRatio":  Ignore(),
					"RecommendedInsulinSensitivity": Ignore(),
					"TotalDailyDose":                Ignore(),
					"TotalDailyDoseScaleFactor":     Ignore(),
					"Weight":                        Ignore(),
					"WeightUnits":                   Ignore(),
				}))
				Expect(*attrs.PhoneNumber).To(MatchAllFields(Fields{
					"CountryCode": Ignore(),
					"Number":      Ignore(),
				}))
				Expect(*attrs.InitialSettings).To(MatchAllFields(Fields{
					"BloodGlucoseUnits":                  Ignore(),
					"BasalRateSchedule":                  Ignore(),
					"BloodGlucoseTargetPhysicalActivity": Ignore(),
					"BloodGlucoseTargetPreprandial":      Ignore(),
					"BloodGlucoseTargetSchedule":         Ignore(),
					"CarbohydrateRatioSchedule":          Ignore(),
					"GlucoseSafetyLimit":                 Ignore(),
					"InsulinModel":                       Ignore(),
					"InsulinSensitivitySchedule":         Ignore(),
					"BasalRateMaximum":                   Ignore(),
					"BolusAmountMaximum":                 Ignore(),
					"PumpID":                             Ignore(),
					"CgmID":                              Ignore(),
				}))
			})
		})
	})

	Context("Prescription settings constants", func() {
		It("don't change", func() {
			// When this test fails it means that some of the constants used in prescriptions have changed,
			// which will cause prescription integrity checks to fail. Those constant shouldn't be changed
			// in normal circumstances.
			Expect(prescription.StateActive).To(Equal("active"))
			Expect(prescription.StateClaimed).To(Equal("claimed"))
			Expect(prescription.StateSubmitted).To(Equal("submitted"))
			Expect(prescription.StateDraft).To(Equal("draft"))
			Expect(prescription.StateExpired).To(Equal("expired"))
			Expect(prescription.StateInactive).To(Equal("inactive"))
			Expect(prescription.StatePending).To(Equal("pending"))
			Expect(prescription.AccountTypePatient).To(Equal("patient"))
			Expect(prescription.AccountTypeCaregiver).To(Equal("caregiver"))
			Expect(prescription.CalculatorMethodTotalDailyDose).To(Equal("totalDailyDose"))
			Expect(prescription.CalculatorMethodTotalDailyDoseAndWeight).To(Equal("totalDailyDoseAndWeight"))
			Expect(prescription.CalculatorMethodWeight).To(Equal("weight"))
			Expect(prescription.SexFemale).To(Equal("female"))
			Expect(prescription.SexMale).To(Equal("male"))
			Expect(prescription.SexUndisclosed).To(Equal("undisclosed"))
			Expect(prescription.TherapySettingInitial).To(Equal("initial"))
			Expect(prescription.TherapySettingTransferPumpSettings).To(Equal("transferPumpSettings"))
			Expect(prescription.TrainingInModule).To(Equal("inModule"))
			Expect(prescription.TrainingInPerson).To(Equal("inPerson"))
			Expect(prescription.UnitKg).To(Equal("kg"))
			Expect(prescription.UnitLbs).To(Equal("lbs"))
			Expect(pump.BolusAmountMaximumUnitsUnits).To(Equal("Units"))
			Expect(pump.BasalRateMaximumUnitsUnitsPerHour).To(Equal("Units/hour"))
			Expect(glucose.MgdL).To(Equal("mg/dL"))
		})
	})
})
