package prescription_test

import (
	"time"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/structure/validator"
	test2 "github.com/tidepool-org/platform/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/prescription"
	"github.com/tidepool-org/platform/prescription/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Revision", func() {
	Describe("New revision", func() {
		var revision *prescription.Revision
		var create *prescription.RevisionCreate
		var userID string
		var revisionID int

		BeforeEach(func() {
			userID = userTest.RandomID()
		})

		Context("With random revision create", func() {
			BeforeEach(func() {
				create = test.RandomRevisionCreate()
				revisionID = test2.RandomIntFromRange(0, 100)
				revision = prescription.NewRevision(userID, revisionID, create)
			})

			It("sets the revision id correctly", func() {
				Expect(revision.RevisionID).To(Equal(revisionID))
			})

			It("sets the signature to nil", func() {
				Expect(revision.Signature).To(BeNil())
			})

			It("creates non-nil attributes", func() {
				Expect(revision.Attributes).ToNot(BeNil())
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
				Expect(revision.Attributes.CreatedTime).To(BeTemporally("~", time.Now()))
			})

			It("sets the modified userID correctly", func() {
				Expect(revision.Attributes.CreatedUserID).To(Equal(userID))
			})
		})
	})

	Describe("Revision", func() {
		var revision *prescription.Revision
		var validate structure.Validator

		BeforeEach(func() {
			revision = test.RandomRevision()
			validate = validator.New()
			Expect(validate.Validate(revision)).ToNot(HaveOccurred())
		})

		Context("Validate", func() {
			BeforeEach(func() {
				validate = validator.New()
			})

			It("fails when revision id is negative", func() {
				revision.RevisionID = -5
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
			validate = validator.New()
			Expect(validate.Validate(attr)).ToNot(HaveOccurred())
		})

		Describe("Validate", func() {
			When("state is 'submitted'", func() {
				BeforeEach(func() {
					validate = validator.New()
					attr.State = prescription.StateSubmitted
				})

				It("fails with empty first name", func() {
					attr.FirstName = ""
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("fails with empty last name", func() {
					attr.LastName = ""
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("fails with empty birthday", func() {
					attr.Birthday = ""
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("fails with invalid birthday", func() {
					attr.Birthday = "20222-03-10"
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("doesn't fail with empty MRN", func() {
					attr.MRN = ""
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
				})

				It("fails with empty sex", func() {
					attr.Sex = ""
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("fails with invalid sex", func() {
					attr.Sex = "invalid-option"
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("doesn't fail with valid sex", func() {
					attr.Sex = prescription.SexMale
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
					attr.Sex = prescription.SexFemale
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
					attr.Sex = prescription.SexUndisclosed
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

				It("doesn't fail with a valid year of diagnosis", func() {
					attr.YearOfDiagnosis = 1999
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
				})

				It("fails with an invalid year of diagnosis", func() {
					attr.YearOfDiagnosis = 1857
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("fails with invalid initial settings", func() {
					attr.InitialSettings.BasalRateMaximum.Value = pointer.FromFloat64(10000.0)
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("fails with empty initial settings when therapy settings is initial", func() {
					attr.InitialSettings = nil
					attr.TherapySettings = prescription.TherapySettingInitial
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("fails with empty initial settings when therapy settings is 'transfer pump settings'", func() {
					attr.InitialSettings = nil
					attr.TherapySettings = prescription.TherapySettingTransferPumpSettings
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("doesn't fail with valid training", func() {
					attr.Training = prescription.TrainingInModule
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
					attr.Training = prescription.TrainingInPerson
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
				})

				It("fails with invalid training", func() {
					attr.Training = "invalid-value"
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("doesn't fail with valid therapy settings", func() {
					attr.TherapySettings = prescription.TherapySettingInitial
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
					attr.TherapySettings = prescription.TherapySettingTransferPumpSettings
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
				})

				It("fails with invalid therapy settings", func() {
					attr.TherapySettings = "invalid-value"
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("fails when state is 'submitted' and prescriber terms are not accepted", func() {
					attr.PrescriberTermsAccepted = false
					attr.State = prescription.StateSubmitted
					Expect(validate.Validate(attr)).To(HaveOccurred())
				})

				It("doesn't fail when state is 'submitted' and prescriber terms accepted is true", func() {
					attr.PrescriberTermsAccepted = true
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

				It("fails when state is 'reviewed'", func() {
					attr.State = prescription.StateReviewed
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

				It("doesn't fail empty attributes and when state is draft or expired", func() {
					now := time.Now()
					attr = &prescription.Attributes{
						FirstName:               "",
						LastName:                "",
						Birthday:                "",
						MRN:                     "",
						Email:                   "",
						Sex:                     "",
						Weight:                  nil,
						YearOfDiagnosis:         0,
						PhoneNumber:             nil,
						InitialSettings:         nil,
						Training:                "",
						TherapySettings:         "",
						PrescriberTermsAccepted: false,
						State:                   "",
						CreatedTime:             now,
						CreatedUserID:           userTest.RandomID(),
					}
					attr.State = prescription.StateDraft
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
					attr.State = prescription.StatePending
					Expect(validate.Validate(attr)).ToNot(HaveOccurred())
				})
			})
		})
	})

	Describe("Initial Settings", func() {
		var settings *prescription.InitialSettings
		var validate structure.Validator

		BeforeEach(func() {
			settings = test.RandomInitialSettings()
			validate = validator.New()
			Expect(validate.Validate(settings)).ToNot(HaveOccurred())
		})

		Describe("Validate", func() {
			BeforeEach(func() {
				validate = validator.New()
			})

			It("fails with empty basal rate schedule", func() {
				settings.BasalRateSchedule = nil
				settings.ValidateAllRequired(validate)
				Expect(validate.Error()).To(HaveOccurred())
			})

			It("fails with empty blood glucose target schedule", func() {
				settings.BloodGlucoseTargetSchedule = nil
				settings.ValidateAllRequired(validate)
				Expect(validate.Error()).To(HaveOccurred())
			})

			It("fails with empty carbohydrate ratio schedule", func() {
				settings.CarbohydrateRatioSchedule = nil
				settings.ValidateAllRequired(validate)
				Expect(validate.Error()).To(HaveOccurred())
			})

			It("fails fail with empty insulin sensitivity schedule", func() {
				settings.InsulinSensitivitySchedule = nil
				settings.ValidateAllRequired(validate)
				Expect(validate.Error()).To(HaveOccurred())
			})

			It("fails with empty basal rate maximum", func() {
				settings.BasalRateMaximum = nil
				settings.ValidateAllRequired(validate)
				Expect(validate.Error()).To(HaveOccurred())
			})

			It("fails with empty bolus amount maximum", func() {
				settings.BolusAmountMaximum = nil
				settings.ValidateAllRequired(validate)
				Expect(validate.Error()).To(HaveOccurred())
			})

			It("fails with empty pump type", func() {
				settings.PumpID = nil
				settings.ValidateAllRequired(validate)
				Expect(validate.Error()).To(HaveOccurred())
			})

			It("fails with empty cgm type", func() {
				settings.CgmID = nil
				settings.ValidateAllRequired(validate)
				Expect(validate.Error()).To(HaveOccurred())
			})
		})
	})

	Describe("Weight", func() {
		var weight *prescription.Weight
		var validate structure.Validator

		BeforeEach(func() {
			weight = test.RandomWeight()
			validate = validator.New()
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
