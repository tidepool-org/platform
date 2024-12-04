package prescription_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/prescription"
	"github.com/tidepool-org/platform/prescription/test"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/structure/validator"
	test2 "github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Revision", func() {
	Describe("New revision", func() {
		var revision *prescription.Revision
		var create *prescription.RevisionCreate
		var revisionID int

		Context("With random revision create", func() {
			BeforeEach(func() {
				create = test.RandomRevisionCreate()
				revisionID = test2.RandomIntFromRange(0, 100)
				revision = prescription.NewRevision(revisionID, create)
			})

			It("sets the revision id correctly", func() {
				Expect(revision.RevisionID).To(Equal(revisionID))
			})

			Context("integrity hash", func() {
				It("is not nil", func() {
					Expect(revision.IntegrityHash).ToNot(BeNil())
				})

				It("algorithm is set to JCSSHA512", func() {
					Expect(revision.IntegrityHash.Algorithm).To(Equal("JCSSHA512"))
				})

				It("hash is not empty", func() {
					Expect(revision.IntegrityHash.Hash).ToNot(BeEmpty())
				})

				It("hash equals to the create hash", func() {
					Expect(revision.IntegrityHash.Hash).To(Equal(create.RevisionHash))
				})
			})

			It("creates non-nil attributes", func() {
				Expect(revision.Attributes).ToNot(BeNil())
			})

			It("sets account type correctly", func() {
				Expect(revision.Attributes.AccountType).To(Equal(create.AccountType))
			})

			It("sets caregiver first name correctly", func() {
				Expect(revision.Attributes.CaregiverFirstName).To(Equal(create.CaregiverFirstName))
			})

			It("sets caregiver last name correctly", func() {
				Expect(revision.Attributes.CaregiverLastName).To(Equal(create.CaregiverLastName))
			})

			It("sets the first name correctly", func() {
				Expect(revision.Attributes.FirstName).To(Equal(create.FirstName))
			})

			It("sets the last name correctly", func() {
				Expect(revision.Attributes.LastName).To(Equal(create.LastName))
			})

			It("sets the birthday correctly", func() {
				Expect(revision.Attributes.Birthday).To(Equal(create.Birthday))
			})

			It("sets the mrn correctly", func() {
				Expect(revision.Attributes.MRN).To(Equal(create.MRN))
			})

			It("sets the email correctly", func() {
				Expect(revision.Attributes.Email).To(Equal(create.Email))
			})

			It("sets the sex correctly", func() {
				Expect(revision.Attributes.Sex).To(Equal(create.Sex))
			})

			It("sets the weight correctly", func() {
				Expect(revision.Attributes.Weight).ToNot(BeNil())
				Expect(revision.Attributes.Weight.Units).To(Equal("kg"))
				Expect(revision.Attributes.Weight.Value).To(Equal(create.Weight.Value))
			})

			It("sets the year of diagnosis correctly", func() {
				Expect(revision.Attributes.YearOfDiagnosis).To(Equal(create.YearOfDiagnosis))
			})

			It("sets the phone number correctly", func() {
				Expect(revision.Attributes.PhoneNumber).To(Equal(create.PhoneNumber))
			})

			It("sets the initial settings correctly", func() {
				Expect(revision.Attributes.InitialSettings).To(Equal(create.InitialSettings))
			})

			It("sets the calculator correctly", func() {
				Expect(revision.Attributes.Calculator).To(Equal(create.Calculator))
			})

			It("sets the training correctly", func() {
				Expect(revision.Attributes.Training).To(Equal(create.Training))
			})

			It("sets the therapy settings correctly", func() {
				Expect(revision.Attributes.TherapySettings).To(Equal(create.TherapySettings))
			})

			It("sets the prescriber terms accepted correctly", func() {
				Expect(revision.Attributes.PrescriberTermsAccepted).To(Equal(create.PrescriberTermsAccepted))
			})

			It("sets the state correctly", func() {
				Expect(revision.Attributes.State).To(Equal(create.State))
			})

			It("sets the modified time correctly", func() {
				Expect(revision.Attributes.CreatedTime).To(BeTemporally("~", time.Now(), 10*time.Millisecond))
			})

			It("sets the modified userID correctly", func() {
				Expect(revision.Attributes.CreatedUserID).To(Equal(create.ClinicianID))
			})
		})
	})

	Describe("Revision", func() {
		var revision *prescription.Revision
		var validate structure.Validator

		BeforeEach(func() {
			revision = test.RandomRevision()
			validate = validator.New(logTest.NewLogger())
			Expect(validate.Validate(revision)).ToNot(HaveOccurred())
		})

		Context("Validate", func() {
			BeforeEach(func() {
				validate = validator.New(logTest.NewLogger())
			})

			It("fails when revision id is negative", func() {
				revision.RevisionID = -5
				Expect(validate.Validate(revision)).To(HaveOccurred())
			})

			It("fails when the integrity hash is not set", func() {
				revision.IntegrityHash = nil
				Expect(validate.Validate(revision)).To(HaveOccurred())
			})

			It("fails when the integrity hash algorithm is invalid", func() {
				revision.IntegrityHash.Algorithm = "invalid"
				Expect(validate.Validate(revision)).To(HaveOccurred())
			})

			It("fails when the integrity hash value is empty", func() {
				revision.IntegrityHash.Hash = ""
				Expect(validate.Validate(revision)).To(HaveOccurred())
			})

			It("fails when attributes are invalid", func() {
				revision.Attributes.State = "invalid"
				Expect(validate.Validate(revision)).To(HaveOccurred())
			})
		})
	})

	Describe("Attributes", func() {
		var attr *prescription.Attributes
		var validate structure.Validator

		BeforeEach(func() {
			attr = test.RandomAttribtues()
			validate = validator.New(logTest.NewLogger())
			Expect(validate.Validate(attr)).ToNot(HaveOccurred())
		})

		Describe("Validate", func() {
			When("state is 'submitted'", func() {
				BeforeEach(func() {
					validate = validator.New(logTest.NewLogger())
					attr.State = prescription.StateSubmitted
				})

				It("fails with empty account type", func() {
					attr.AccountType = pointer.FromString("")
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("fails with empty caregiver first name when account type is 'caregiver'", func() {
					attr.AccountType = pointer.FromString(prescription.AccountTypeCaregiver)
					attr.CaregiverFirstName = pointer.FromString("")
					attr.CaregiverLastName = pointer.FromString("Doe")
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("fails with empty caregiver last name when account type is 'caregiver'", func() {
					attr.AccountType = pointer.FromString(prescription.AccountTypeCaregiver)
					attr.CaregiverFirstName = pointer.FromString("Jane")
					attr.CaregiverLastName = pointer.FromString("")
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("doesn't fail with empty caregiver names when account type is patient", func() {
					attr.AccountType = pointer.FromString(prescription.AccountTypePatient)
					attr.CaregiverFirstName = pointer.FromString("")
					attr.CaregiverLastName = pointer.FromString("")
					Expect(validate.Validate(attr)).To(Not(HaveOccurred()))
				})

				It("doesn't fail with nil caregiver first name when account type is patient", func() {
					attr.AccountType = pointer.FromString(prescription.AccountTypePatient)
					attr.CaregiverFirstName = nil
					attr.CaregiverLastName = pointer.FromString("")
					Expect(validate.Validate(attr)).To(Not(HaveOccurred()))
				})

				It("doesn't fail with nil caregiver last name when account type is patient", func() {
					attr.AccountType = pointer.FromString(prescription.AccountTypePatient)
					attr.CaregiverFirstName = pointer.FromString("")
					attr.CaregiverLastName = nil
					Expect(validate.Validate(attr)).To(Not(HaveOccurred()))
				})

				It("doesn't fail with nil caregiver names when account type is patient", func() {
					attr.AccountType = pointer.FromString(prescription.AccountTypePatient)
					attr.CaregiverFirstName = nil
					attr.CaregiverLastName = nil
					Expect(validate.Validate(attr)).To(Not(HaveOccurred()))
				})

				It("fails with non-empty caregiver names when account type is patient", func() {
					attr.AccountType = pointer.FromString(prescription.AccountTypePatient)
					attr.CaregiverFirstName = pointer.FromString("Jane")
					attr.CaregiverLastName = pointer.FromString("Doe")
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("fails with empty first name", func() {
					attr.FirstName = pointer.FromString("")
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("fails with empty last name", func() {
					attr.LastName = pointer.FromString("")
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("fails with empty birthday", func() {
					attr.Birthday = pointer.FromString("")
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("fails with invalid birthday", func() {
					attr.Birthday = pointer.FromString("20222-03-10")
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("doesn't fail with empty MRN", func() {
					attr.MRN = pointer.FromString("")
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
				})

				It("fails with empty sex", func() {
					attr.Sex = pointer.FromString("")
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("fails with invalid sex", func() {
					attr.Sex = pointer.FromString("invalid-option")
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("doesn't fail with valid sex", func() {
					attr.Sex = pointer.FromString(prescription.SexMale)
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
					attr.Sex = pointer.FromString(prescription.SexFemale)
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
					attr.Sex = pointer.FromString(prescription.SexUndisclosed)
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
				})

				It("doesn't fail without weight", func() {
					attr.Weight = nil
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
				})

				It("doesn't fail valid weight", func() {
					attr.Weight = &prescription.Weight{
						Value: pointer.FromFloat64(50.0),
						Units: "kg",
					}
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
				})

				It("fails with invalid weight", func() {
					attr.Weight = &prescription.Weight{
						Value: pointer.FromFloat64(-50.0),
						Units: "kg",
					}
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("doesn't fail with a empty year of diagnosis", func() {
					attr.YearOfDiagnosis = nil
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
				})

				It("doesn't fail with a valid year of diagnosis", func() {
					attr.YearOfDiagnosis = pointer.FromInt(1999)
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
				})

				It("fails with an invalid year of diagnosis", func() {
					attr.YearOfDiagnosis = pointer.FromInt(1857)
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("doesn't fail when phone number is not set", func() {
					attr.PhoneNumber = nil
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
				})

				It("fails with invalid initial settings", func() {
					attr.InitialSettings.BasalRateMaximum.Value = pointer.FromFloat64(10000.0)
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("fails with empty initial settings when therapy settings is initial", func() {
					attr.InitialSettings = nil
					attr.TherapySettings = pointer.FromString(prescription.TherapySettingInitial)
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("fails with empty initial settings when therapy settings is 'transfer pump settings'", func() {
					attr.InitialSettings = nil
					attr.TherapySettings = pointer.FromString(prescription.TherapySettingTransferPumpSettings)
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("doesn't fail with empty calculator", func() {
					attr.Calculator = nil
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
				})

				It("fails with invalid calculator values", func() {
					attr.Calculator.Method = pointer.FromString(prescription.CalculatorMethodWeight)
					attr.Calculator.Weight = pointer.FromFloat64(-1.0)
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("doesn't fail with valid training", func() {
					attr.Training = pointer.FromString(prescription.TrainingInModule)
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
					attr.Training = pointer.FromString(prescription.TrainingInPerson)
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
				})

				It("doesn't fail when training is not set", func() {
					attr.Training = nil
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
				})

				It("fails with invalid training", func() {
					attr.Training = pointer.FromString("invalid-value")
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("doesn't fail with valid therapy settings", func() {
					attr.TherapySettings = pointer.FromString(prescription.TherapySettingInitial)
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
					attr.TherapySettings = pointer.FromString(prescription.TherapySettingTransferPumpSettings)
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
				})

				It("fails with invalid therapy settings", func() {
					attr.TherapySettings = pointer.FromString("invalid-value")
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("fails when state is 'submitted' and prescriber terms are not accepted", func() {
					attr.PrescriberTermsAccepted = pointer.FromBool(false)
					attr.State = prescription.StateSubmitted
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("doesn't fail when state is 'submitted' and prescriber terms accepted is true", func() {
					attr.PrescriberTermsAccepted = pointer.FromBool(true)
					attr.State = prescription.StateSubmitted
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
				})

				It("doesn't fail with valid states", func() {
					attr.State = prescription.StateSubmitted
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
					attr.State = prescription.StateDraft
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
					attr.State = prescription.StatePending
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
				})

				It("fails when state is 'claimed'", func() {
					attr.State = prescription.StateClaimed
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("fails when state is 'inactive'", func() {
					attr.State = prescription.StateInactive
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("fails when state is 'active'", func() {
					attr.State = prescription.StateActive
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("fails when state is 'expired'", func() {
					attr.State = prescription.StateExpired
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("fails with invalid modified user id", func() {
					attr.CreatedUserID = "invalid"
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("doesn't fail with nil attributes when state is 'draft' or 'expired'", func() {
					now := time.Now()
					attr = &prescription.Attributes{
						DataAttributes: prescription.DataAttributes{
							AccountType:             nil,
							CaregiverFirstName:      nil,
							CaregiverLastName:       nil,
							FirstName:               nil,
							LastName:                nil,
							Birthday:                nil,
							MRN:                     nil,
							Email:                   nil,
							Sex:                     nil,
							Weight:                  nil,
							YearOfDiagnosis:         nil,
							PhoneNumber:             nil,
							InitialSettings:         nil,
							Calculator:              nil,
							Training:                nil,
							TherapySettings:         nil,
							PrescriberTermsAccepted: nil,
						},
						CreationAttributes: prescription.CreationAttributes{
							CreatedTime:   now,
							CreatedUserID: userTest.RandomID(),
						},
					}
					attr.State = prescription.StateDraft
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
					attr.State = prescription.StatePending
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
				})
			})
		})
	})

	Describe("Weight", func() {
		var weight *prescription.Weight
		var validate structure.Validator

		BeforeEach(func() {
			weight = test.RandomWeight()
			validate = validator.New(logTest.NewLogger())
		})

		Describe("Validate", func() {
			It("fails with 'lb' as units", func() {
				weight.Units = "lb"
				Expect(validate.Validate(weight)).To(HaveOccurred())
			})

			It("fails if value is negative number", func() {
				weight.Value = pointer.FromFloat64(-75.0)
				Expect(validate.Validate(weight)).To(HaveOccurred())
			})

			It("fails if value is 0", func() {
				weight.Value = pointer.FromFloat64(0)
				Expect(validate.Validate(weight)).To(HaveOccurred())
			})

			It("doesn't fail if value is nil", func() {
				weight.Value = nil
				Expect(validate.Validate(weight)).ToNot(HaveOccurred())
			})
		})
	})
})
