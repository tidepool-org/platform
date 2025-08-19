package consent_test

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/log"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"syreclabs.com/go/faker"

	"github.com/tidepool-org/platform/consent"
	consentTest "github.com/tidepool-org/platform/consent/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/structure/validator"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Consent", func() {
	var logger log.Logger

	BeforeEach(func() {
		logger = logTest.NewLogger()
	})

	Context("Record", func() {
		Describe("NewRecord", func() {
			var ctx context.Context
			var userID string
			var recordCreate *consent.RecordCreate

			BeforeEach(func() {
				ctx = context.Background()
				userID = userTest.RandomUserID()
				recordCreate = consentTest.RandomRecordCreate()
				logger = logTest.NewLogger()
			})

			Context("with valid inputs", func() {
				It("creates a record successfully", func() {
					record, err := consent.NewRecord(ctx, userID, recordCreate)
					Expect(err).ToNot(HaveOccurred())
					Expect(record).ToNot(BeNil())
				})

				It("sets the user ID correctly", func() {
					record, _ := consent.NewRecord(ctx, userID, recordCreate)
					Expect(record.UserID).To(Equal(userID))
				})

				It("generates a non-empty ID", func() {
					record, _ := consent.NewRecord(ctx, userID, recordCreate)
					Expect(record.ID).ToNot(BeEmpty())
				})

				It("sets status to active", func() {
					record, _ := consent.NewRecord(ctx, userID, recordCreate)
					Expect(record.Status).To(Equal(consent.RecordStatusActive))
				})

				It("sets grant time to creation time", func() {
					record, _ := consent.NewRecord(ctx, userID, recordCreate)
					Expect(record.GrantTime).To(Equal(recordCreate.CreatedTime))
				})

				It("sets created time to creation time", func() {
					record, _ := consent.NewRecord(ctx, userID, recordCreate)
					Expect(record.CreatedTime).To(Equal(recordCreate.CreatedTime))
				})

				It("sets modified time to now", func() {
					before := time.Now()
					record, _ := consent.NewRecord(ctx, userID, recordCreate)
					after := time.Now()
					Expect(record.ModifiedTime).To(BeTemporally(">=", before))
					Expect(record.ModifiedTime).To(BeTemporally("<=", after))
				})

				It("copies all fields from create", func() {
					record, _ := consent.NewRecord(ctx, userID, recordCreate)
					Expect(record.AgeGroup).To(Equal(recordCreate.AgeGroup))
					Expect(record.OwnerName).To(Equal(recordCreate.OwnerName))
					Expect(record.ParentGuardianName).To(Equal(recordCreate.ParentGuardianName))
					Expect(record.GrantorType).To(Equal(recordCreate.GrantorType))
					Expect(record.Type).To(Equal(recordCreate.Type))
					Expect(record.Version).To(Equal(recordCreate.Version))
					Expect(record.Metadata).To(Equal(recordCreate.Metadata))
				})

				It("does not set revocation time", func() {
					record, _ := consent.NewRecord(ctx, userID, recordCreate)
					Expect(record.RevocationTime).To(BeNil())
				})
			})

			Context("with invalid inputs", func() {
				It("fails when user ID is empty", func() {
					record, err := consent.NewRecord(ctx, "", recordCreate)
					Expect(err).To(HaveOccurred())
					Expect(record).To(BeNil())
					Expect(err.Error()).To(ContainSubstring("user id is missing"))
				})

				It("fails when create is nil", func() {
					record, err := consent.NewRecord(ctx, userID, nil)
					Expect(err).To(HaveOccurred())
					Expect(record).To(BeNil())
					Expect(err.Error()).To(ContainSubstring("create is missing"))
				})

				It("fails when create is invalid", func() {
					recordCreate.OwnerName = "" // Invalid: empty owner name
					record, err := consent.NewRecord(ctx, userID, recordCreate)
					Expect(err).To(HaveOccurred())
					Expect(record).To(BeNil())
					Expect(err.Error()).To(ContainSubstring("create is invalid"))
				})
			})
		})

		Describe("Validate", func() {
			var record *consent.Record
			var validate structure.Validator

			BeforeEach(func() {
				ctx := context.Background()
				userID := userTest.RandomUserID()
				recordCreate := consentTest.RandomRecordCreate()
				var err error
				record, err = consent.NewRecord(ctx, userID, recordCreate)
				Expect(err).ToNot(HaveOccurred())

				validate = validator.New(logger)
			})

			Context("with valid record", func() {
				It("passes validation", func() {
					record.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})
			})

			Context("with invalid fields", func() {
				It("fails when ID is empty", func() {
					record.ID = ""
					record.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when UserID is empty", func() {
					record.UserID = ""
					record.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when UserID is invalid", func() {
					record.UserID = "invalid-user-id"
					record.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when status is invalid", func() {
					record.Status = "invalid-status"
					record.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when age group is invalid", func() {
					record.AgeGroup = "invalid-age-group"
					record.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when owner name is empty", func() {
					record.OwnerName = ""
					record.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when owner name is too long", func() {
					record.OwnerName = faker.Lorem().Characters(257)
					record.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when grantor type is invalid", func() {
					record.GrantorType = "invalid-grantor"
					record.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when type is empty", func() {
					record.Type = ""
					record.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when version is zero", func() {
					record.Version = 0
					record.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when version is negative", func() {
					record.Version = -1
					record.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when grant time is zero", func() {
					record.GrantTime = time.Time{}
					record.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when grant time is in the future", func() {
					record.GrantTime = time.Now().Add(time.Hour)
					record.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when revocation time is in the future", func() {
					futureTime := time.Now().Add(time.Hour)
					record.RevocationTime = &futureTime
					record.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when created time is in the future", func() {
					record.CreatedTime = time.Now().Add(time.Hour)
					record.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when modified time is in the future", func() {
					record.ModifiedTime = time.Now().Add(time.Hour)
					record.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})
			})

			Context("age group specific validations", func() {
				Context("when age group is under 13", func() {
					BeforeEach(func() {
						record.AgeGroup = consent.AgeGroupUnderThirteen
					})

					It("requires parent guardian name", func() {
						record.ParentGuardianName = nil
						record.Validate(validate)
						Expect(validate.Error()).To(HaveOccurred())
					})

					It("requires grantor type to be parent/guardian", func() {
						record.GrantorType = consent.GrantorTypeOwner
						record.Validate(validate)
						Expect(validate.Error()).To(HaveOccurred())
					})

					It("passes when parent guardian name exists and grantor type is correct", func() {
						record.ParentGuardianName = pointer.FromString("Parent Name")
						record.GrantorType = consent.GrantorTypeParentGuardian
						record.Validate(validate)
						Expect(validate.Error()).ToNot(HaveOccurred())
					})
				})

				Context("when age group is 13-17", func() {
					BeforeEach(func() {
						record.AgeGroup = consent.AgeGroupThirteenSeventeen
					})

					It("requires parent guardian name", func() {
						record.ParentGuardianName = nil
						record.Validate(validate)
						Expect(validate.Error()).To(HaveOccurred())
					})

					It("requires grantor type to be parent/guardian", func() {
						record.GrantorType = consent.GrantorTypeOwner
						record.Validate(validate)
						Expect(validate.Error()).To(HaveOccurred())
					})
				})

				Context("when age group is 18 or over", func() {
					BeforeEach(func() {
						record.AgeGroup = consent.AgeGroupEighteenOrOver
					})

					It("requires parent guardian name to not exist", func() {
						record.ParentGuardianName = pointer.FromString("Parent Name")
						record.Validate(validate)
						Expect(validate.Error()).To(HaveOccurred())
					})

					It("passes when parent guardian name is nil and grantor type is owner", func() {
						record.ParentGuardianName = nil
						record.GrantorType = consent.GrantorTypeOwner
						record.Validate(validate)
						Expect(validate.Error()).ToNot(HaveOccurred())
					})
				})
			})
		})

		Describe("ToUpdate", func() {
			It("creates a record update from record", func() {
				ctx := context.Background()
				userID := userTest.RandomUserID()
				recordCreate := consentTest.RandomRecordCreate()
				record, err := consent.NewRecord(ctx, userID, recordCreate)
				Expect(err).ToNot(HaveOccurred())

				update := record.ToUpdate()
				Expect(update).ToNot(BeNil())
				Expect(update.ID).To(Equal(record.ID))
				Expect(update.UserID).To(Equal(record.UserID))
				Expect(update.Metadata).To(Equal(record.Metadata))
			})
		})
	})

	Context("RecordCreate", func() {
		Describe("NewRecordCreate", func() {
			It("creates a new record create with current time", func() {
				before := time.Now()
				create := consent.NewRecordCreate()
				after := time.Now()

				Expect(create).ToNot(BeNil())
				Expect(create.CreatedTime).To(BeTemporally(">=", before))
				Expect(create.CreatedTime).To(BeTemporally("<=", after))
			})
		})

		Describe("Validate", func() {
			var recordCreate *consent.RecordCreate
			var validate structure.Validator

			BeforeEach(func() {
				recordCreate = consentTest.RandomRecordCreate()
				validate = validator.New(logger)
			})

			Context("with valid data", func() {
				It("passes validation", func() {
					recordCreate.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})
			})

			Context("age group validations", func() {
				Context("when age group is 18 or over", func() {
					BeforeEach(func() {
						recordCreate.AgeGroup = consent.AgeGroupEighteenOrOver
						recordCreate.GrantorType = consent.GrantorTypeOwner
						recordCreate.ParentGuardianName = nil
					})

					It("requires grantor type to be owner", func() {
						recordCreate.GrantorType = consent.GrantorTypeParentGuardian
						recordCreate.Validate(validate)
						Expect(validate.Error()).To(HaveOccurred())
					})

					It("requires parent guardian name to not exist", func() {
						recordCreate.ParentGuardianName = pointer.FromString("Parent Name")
						recordCreate.Validate(validate)
						Expect(validate.Error()).To(HaveOccurred())
					})

					It("passes with correct settings", func() {
						recordCreate.Validate(validate)
						Expect(validate.Error()).ToNot(HaveOccurred())
					})
				})

				Context("when age group is under 18", func() {
					BeforeEach(func() {
						recordCreate.AgeGroup = consent.AgeGroupUnderThirteen
						recordCreate.GrantorType = consent.GrantorTypeParentGuardian
						recordCreate.ParentGuardianName = pointer.FromString("Parent Name")
					})

					It("requires parent guardian name to exist", func() {
						recordCreate.ParentGuardianName = nil
						recordCreate.Validate(validate)
						Expect(validate.Error()).To(HaveOccurred())
					})

					It("requires grantor type to be parent/guardian", func() {
						recordCreate.GrantorType = consent.GrantorTypeOwner
						recordCreate.Validate(validate)
						Expect(validate.Error()).To(HaveOccurred())
					})

					It("passes with correct settings", func() {
						recordCreate.Validate(validate)
						Expect(validate.Error()).ToNot(HaveOccurred())
					})
				})
			})

			Context("field validations", func() {
				It("fails when age group is invalid", func() {
					recordCreate.AgeGroup = "invalid"
					recordCreate.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when created time is zero", func() {
					recordCreate.CreatedTime = time.Time{}
					recordCreate.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when created time is in future", func() {
					recordCreate.CreatedTime = time.Now().Add(time.Hour)
					recordCreate.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when grantor type is missing", func() {
					recordCreate.GrantorType = ""
					recordCreate.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when grantor type is invalid", func() {
					recordCreate.GrantorType = "invalid"
					recordCreate.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when owner name is missing", func() {
					recordCreate.OwnerName = ""
					recordCreate.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when owner name is too long", func() {
					recordCreate.OwnerName = faker.Lorem().Characters(257)
					recordCreate.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when type is missing", func() {
					recordCreate.Type = ""
					recordCreate.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when version is missing", func() {
					recordCreate.Version = 0
					recordCreate.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when version is negative", func() {
					recordCreate.Version = -1
					recordCreate.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})
			})
		})
	})

	Context("RecordFilter", func() {
		Describe("NewRecordFilter", func() {
			It("creates a filter with latest set to true", func() {
				filter := consent.NewRecordFilter()
				Expect(filter).ToNot(BeNil())
				Expect(filter.Latest).ToNot(BeNil())
				Expect(*filter.Latest).To(BeTrue())
			})
		})

		Describe("Validate", func() {
			var filter *consent.RecordFilter
			var validate structure.Validator

			BeforeEach(func() {
				filter = consent.NewRecordFilter()
				validate = validator.New(logger)
			})

			Context("with valid filter", func() {
				It("passes validation", func() {
					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})
			})

			Context("with invalid fields", func() {
				It("fails when ID is empty string", func() {
					filter.ID = pointer.FromString("")
					filter.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when status is invalid", func() {
					invalidStatus := consent.RecordStatus("invalid")
					filter.Status = &invalidStatus
					filter.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when version is zero", func() {
					filter.Version = pointer.FromInt(0)
					filter.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when version is negative", func() {
					filter.Version = pointer.FromInt(-1)
					filter.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})
			})

			Context("with valid optional fields", func() {
				It("passes when ID is valid", func() {
					filter.ID = pointer.FromString(primitive.NewObjectID().Hex())
					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})

				It("passes when status is valid", func() {
					filter.Status = pointer.FromAny(consent.RecordStatusActive)
					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})

				It("passes when type is valid", func() {
					filter.Type = pointer.FromString("big_data_donation_project")
					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})

				It("passes when version is positive", func() {
					filter.Version = pointer.FromInt(1)
					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})
			})
		})
	})

	Context("RecordMetadata", func() {
		Describe("ValidateBigDataDonationProject", func() {
			var metadata *consent.RecordMetadata
			var validate structure.Validator

			BeforeEach(func() {
				metadata = &consent.RecordMetadata{}
				validate = validator.New(logger)
			})

			Context("with valid organizations", func() {
				It("passes validation", func() {
					metadata.SupportedOrganizations = []consent.BigDataDonationProjectOrganization{
						consent.BigDataDonationProjectOrganizationsADCES,
						consent.BigDataDonationProjectOrganizationsBeyondType1,
					}
					metadata.ValidateBigDataDonationProject(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})
			})

			Context("with invalid organizations", func() {
				It("fails when organization is invalid", func() {
					metadata.SupportedOrganizations = []consent.BigDataDonationProjectOrganization{
						"Invalid Organization",
					}
					metadata.ValidateBigDataDonationProject(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})
			})

			Context("with empty organizations", func() {
				It("passes validation", func() {
					metadata.SupportedOrganizations = []consent.BigDataDonationProjectOrganization{}
					metadata.ValidateBigDataDonationProject(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})
			})
		})

		Describe("Validator", func() {
			var metadata *consent.RecordMetadata

			BeforeEach(func() {
				metadata = &consent.RecordMetadata{}
			})

			It("returns big data donation validator for correct type", func() {
				validatorFn := metadata.Validator("big_data_donation_project")
				Expect(validatorFn).ToNot(BeNil())
				// We can't easily test the function equality, but we can test it doesn't panic
				validate := validator.New(logger)
				validatorFn(validate)
			})

			It("returns default validator for unknown type", func() {
				validatorFn := metadata.Validator("unknown_type")
				Expect(validatorFn).ToNot(BeNil())
				// Default validator should not cause errors
				validate := validator.New(logger)
				validatorFn(validate)
				Expect(validate.Error()).ToNot(HaveOccurred())
			})

			Context("when metadata is nil", func() {
				BeforeEach(func() {
					metadata = nil
				})

				It("returns default validator", func() {
					validatorFn := metadata.Validator("big_data_donation_project")
					Expect(validatorFn).ToNot(BeNil())
					validate := validator.New(logger)
					validatorFn(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})
			})
		})
	})

	Context("RecordUpdate", func() {
		Describe("ToRecord", func() {
			It("converts update back to record", func() {
				ctx := context.Background()
				userID := userTest.RandomUserID()
				recordCreate := consentTest.RandomRecordCreate()
				record, err := consent.NewRecord(ctx, userID, recordCreate)
				Expect(err).ToNot(HaveOccurred())

				update := record.ToUpdate()
				convertedRecord := update.ToRecord()

				Expect(convertedRecord).ToNot(BeNil())
				Expect(convertedRecord.ID).To(Equal(record.ID))
				Expect(convertedRecord.UserID).To(Equal(record.UserID))
				Expect(convertedRecord.Status).To(Equal(record.Status))
			})
		})

		Describe("Validate", func() {
			var update *consent.RecordUpdate
			var validate structure.Validator

			BeforeEach(func() {
				ctx := context.Background()
				userID := userTest.RandomUserID()
				recordCreate := consentTest.RandomRecordCreate()
				record, err := consent.NewRecord(ctx, userID, recordCreate)
				Expect(err).ToNot(HaveOccurred())

				update = record.ToUpdate()
				validate = validator.New(logger)
			})

			It("passes validation with valid update", func() {
				update.Validate(validate)
				Expect(validate.Error()).ToNot(HaveOccurred())
			})
		})
	})

	Context("RecordRevoke", func() {
		Describe("NewRecordRevoke", func() {
			It("creates a revoke with current time", func() {
				before := time.Now()
				revoke := consent.NewRecordRevoke()
				after := time.Now()

				Expect(revoke).ToNot(BeNil())
				Expect(revoke.RevocationTime).To(BeTemporally(">=", before))
				Expect(revoke.RevocationTime).To(BeTemporally("<=", after))
			})
		})

		Describe("Validate", func() {
			var revoke *consent.RecordRevoke
			var validate structure.Validator

			BeforeEach(func() {
				revoke = consent.NewRecordRevoke()
				revoke.ID = consent.NewRecordID()
				validate = validator.New(logger)
			})

			Context("with valid revoke", func() {
				It("passes validation", func() {
					revoke.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})
			})

			Context("with invalid fields", func() {
				It("fails when ID is empty", func() {
					revoke.ID = ""
					revoke.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})

				It("fails when revocation time is zero", func() {
					revoke.RevocationTime = time.Time{}
					revoke.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})
			})
		})
	})

	Describe("NewRecordID", func() {
		It("generates a non-empty ID", func() {
			id := consent.NewRecordID()
			Expect(id).ToNot(BeEmpty())
		})

		It("generates unique IDs", func() {
			id1 := consent.NewRecordID()
			id2 := consent.NewRecordID()
			Expect(id1).ToNot(Equal(id2))
		})
	})

	Describe("RecordStatuses", func() {
		It("returns all valid statuses", func() {
			statuses := consent.RecordStatuses()
			Expect(statuses).To(ContainElement(consent.RecordStatusActive))
			Expect(statuses).To(ContainElement(consent.RecordStatusRevoked))
			Expect(len(statuses)).To(Equal(2))
		})
	})

	Describe("AgeGroups", func() {
		It("returns all valid age groups", func() {
			ageGroups := consent.AgeGroups()
			Expect(ageGroups).To(ContainElement(consent.AgeGroupUnderThirteen))
			Expect(ageGroups).To(ContainElement(consent.AgeGroupThirteenSeventeen))
			Expect(ageGroups).To(ContainElement(consent.AgeGroupEighteenOrOver))
			Expect(len(ageGroups)).To(Equal(3))
		})
	})

	Describe("GrantorTypes", func() {
		It("returns all valid grantor types", func() {
			grantorTypes := consent.GrantorTypes()
			Expect(grantorTypes).To(ContainElement(consent.GrantorTypeOwner))
			Expect(grantorTypes).To(ContainElement(consent.GrantorTypeParentGuardian))
			Expect(len(grantorTypes)).To(Equal(2))
		})
	})

	Describe("BigDataDonationProjectOrganizations", func() {
		It("returns all valid organizations", func() {
			orgs := consent.BigDataDonationProjectOrganizations()
			Expect(orgs).To(ContainElement(consent.BigDataDonationProjectOrganizationsADCES))
			Expect(orgs).To(ContainElement(consent.BigDataDonationProjectOrganizationsBeyondType1))
			Expect(orgs).To(ContainElement(consent.BigDataDonationProjectOrganizationsChildrenWithDiabetes))
			Expect(orgs).To(ContainElement(consent.BigDataDonationProjectOrganizationsTheDiabetesLink))
			Expect(orgs).To(ContainElement(consent.BigDataDonationProjectOrganizationsDYF))
			Expect(orgs).To(ContainElement(consent.BigDataDonationProjectOrganizationsDiabetesSisters))
			Expect(orgs).To(ContainElement(consent.BigDataDonationProjectOrganizationsTheDiaTribeFoundation))
			Expect(orgs).To(ContainElement(consent.BigDataDonationProjectOrganizationsBreakthroughT1D))
			Expect(len(orgs)).To(Equal(8))
		})
	})

	Describe("NewRecordStatus", func() {
		It("returns nil when value is nil", func() {
			status := consent.NewRecordStatus(nil)
			Expect(status).To(BeNil())
		})

		It("returns pointer to RecordStatus when value is provided", func() {
			value := "active"
			status := consent.NewRecordStatus(&value)
			Expect(status).ToNot(BeNil())
			Expect(*status).To(Equal(consent.RecordStatus("active")))
		})
	})

	Describe("RecordStatus constants", func() {
		It("has correct values", func() {
			Expect(string(consent.RecordStatusActive)).To(Equal("active"))
			Expect(string(consent.RecordStatusRevoked)).To(Equal("revoked"))
		})
	})

	Describe("GrantorType constants", func() {
		It("has correct values", func() {
			Expect(consent.GrantorTypeOwner).To(Equal("owner"))
			Expect(consent.GrantorTypeParentGuardian).To(Equal("parent/guardian"))
		})
	})

	Describe("AgeGroup constants", func() {
		It("has correct values", func() {
			Expect(string(consent.AgeGroupUnderThirteen)).To(Equal("<13"))
			Expect(string(consent.AgeGroupThirteenSeventeen)).To(Equal("13-17"))
			Expect(string(consent.AgeGroupEighteenOrOver)).To(Equal(">=18"))
		})
	})

	Describe("BigDataDonationProjectOrganization constants", func() {
		It("has correct values", func() {
			Expect(string(consent.BigDataDonationProjectOrganizationsADCES)).To(Equal("ADCES Foundation"))
			Expect(string(consent.BigDataDonationProjectOrganizationsBeyondType1)).To(Equal("Beyond Type 1"))
			Expect(string(consent.BigDataDonationProjectOrganizationsChildrenWithDiabetes)).To(Equal("Children With Diabetes"))
			Expect(string(consent.BigDataDonationProjectOrganizationsTheDiabetesLink)).To(Equal("The Diabetes Link"))
			Expect(string(consent.BigDataDonationProjectOrganizationsDYF)).To(Equal("Diabetes Youth Families (DYF)"))
			Expect(string(consent.BigDataDonationProjectOrganizationsDiabetesSisters)).To(Equal("DiabetesSisters"))
			Expect(string(consent.BigDataDonationProjectOrganizationsTheDiaTribeFoundation)).To(Equal("The diaTribe Foundation"))
			Expect(string(consent.BigDataDonationProjectOrganizationsBreakthroughT1D)).To(Equal("Breakthrough T1D"))
		})
	})

	Context("Integration Tests", func() {
		Describe("Complete workflow", func() {
			var ctx context.Context
			var userID string

			BeforeEach(func() {
				ctx = context.Background()
				userID = userTest.RandomUserID()
			})

			Context("Adult user consent", func() {
				It("creates, updates, and revokes consent successfully", func() {
					// Create consent for adult user
					recordCreate := consent.NewRecordCreate()
					recordCreate.AgeGroup = consent.AgeGroupEighteenOrOver
					recordCreate.OwnerName = faker.Name().Name()
					recordCreate.GrantorType = consent.GrantorTypeOwner
					recordCreate.Type = "big_data_donation_project"
					recordCreate.Version = 1
					recordCreate.Metadata = &consent.RecordMetadata{
						SupportedOrganizations: []consent.BigDataDonationProjectOrganization{
							consent.BigDataDonationProjectOrganizationsADCES,
						},
					}

					// Validate create
					validate := validator.New(logger)
					recordCreate.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())

					// Create record
					record, err := consent.NewRecord(ctx, userID, recordCreate)
					Expect(err).ToNot(HaveOccurred())
					Expect(record.Status).To(Equal(consent.RecordStatusActive))

					// Validate record
					validate = validator.New(logger)
					record.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())

					// Create update
					update := record.ToUpdate()
					update.Metadata.SupportedOrganizations = append(
						update.Metadata.SupportedOrganizations,
						consent.BigDataDonationProjectOrganizationsBeyondType1,
					)

					// Validate update
					validate = validator.New(logger)
					update.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})
			})

			Context("Minor user consent", func() {
				It("requires parent/guardian information", func() {
					// Create consent for minor user
					recordCreate := consent.NewRecordCreate()
					recordCreate.AgeGroup = consent.AgeGroupUnderThirteen
					recordCreate.OwnerName = faker.Name().Name()
					recordCreate.ParentGuardianName = pointer.FromString(faker.Name().Name())
					recordCreate.GrantorType = consent.GrantorTypeParentGuardian
					recordCreate.Type = "big_data_donation_project"
					recordCreate.Version = 1
					recordCreate.Metadata = &consent.RecordMetadata{
						SupportedOrganizations: []consent.BigDataDonationProjectOrganization{
							consent.BigDataDonationProjectOrganizationsChildrenWithDiabetes,
						},
					}

					// Validate create
					validate := validator.New(logger)
					recordCreate.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())

					// Create record
					record, err := consent.NewRecord(ctx, userID, recordCreate)
					Expect(err).ToNot(HaveOccurred())
					Expect(record.ParentGuardianName).ToNot(BeNil())
					Expect(record.GrantorType).To(Equal(consent.GrantorTypeParentGuardian))

					// Validate record
					validate = validator.New(logger)
					record.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})

				It("fails validation when parent/guardian info is missing", func() {
					// Create consent for minor without parent/guardian info
					recordCreate := consent.NewRecordCreate()
					recordCreate.AgeGroup = consent.AgeGroupUnderThirteen
					recordCreate.OwnerName = faker.Name().Name()
					recordCreate.GrantorType = consent.GrantorTypeOwner // Wrong for minor
					recordCreate.Type = "big_data_donation_project"
					recordCreate.Version = 1

					// Validate create - should fail
					validate := validator.New(logger)
					recordCreate.Validate(validate)
					Expect(validate.Error()).To(HaveOccurred())
				})
			})

			Context("Filtering scenarios", func() {
				It("creates and validates various filters", func() {
					// Basic filter
					filter := consent.NewRecordFilter()
					validate := validator.New(logger)
					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())

					// Filter with specific status
					filter.Status = pointer.FromAny(consent.RecordStatusActive)
					validate = validator.New(logger)
					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())

					// Filter with specific type
					filter.Type = pointer.FromString("big_data_donation_project")
					validate = validator.New(logger)
					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())

					// Filter with specific version
					filter.Version = pointer.FromInt(1)
					validate = validator.New(logger)
					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())

					// Filter with ID
					filter.ID = pointer.FromString(consent.NewRecordID())
					validate = validator.New(logger)
					filter.Validate(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})
			})
		})

		Describe("Edge cases and error conditions", func() {
			Context("Metadata validation edge cases", func() {
				It("handles nil metadata gracefully", func() {
					var metadata *consent.RecordMetadata
					validatorFn := metadata.Validator("big_data_donation_project")
					Expect(validatorFn).ToNot(BeNil())

					validate := validator.New(logger)
					validatorFn(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})

				It("validates empty supported organizations", func() {
					metadata := &consent.RecordMetadata{
						SupportedOrganizations: []consent.BigDataDonationProjectOrganization{},
					}
					validate := validator.New(logger)
					metadata.ValidateBigDataDonationProject(validate)
					Expect(validate.Error()).ToNot(HaveOccurred())
				})
			})
		})
	})
})
