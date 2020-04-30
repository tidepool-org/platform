package mongo_test

import (
	"context"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"

	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/store/structured/mongoofficial"

	logNull "github.com/tidepool-org/platform/log/null"

	"github.com/tidepool-org/platform/user"

	"syreclabs.com/go/faker"

	userTest "github.com/tidepool-org/platform/user/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/prescription"
	prescriptionStore "github.com/tidepool-org/platform/prescription/store"
	prescriptionStoreMongo "github.com/tidepool-org/platform/prescription/store/mongo"
	"github.com/tidepool-org/platform/prescription/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

var _ = Describe("PrescriptionRepository", func() {
	var mongoConfig *storeStructuredMongo.Config
	var store *prescriptionStoreMongo.PrescriptionStore
	var logger *logTest.Logger
	var repository prescriptionStore.PrescriptionRepository

	BeforeEach(func() {
		logger = logTest.NewLogger()
		mongoConfig = storeStructuredMongoTest.NewConfig()
	})

	AfterEach(func() {
		if store != nil {
			store.Terminate(context.Background())
		}
	})

	Context("with a new store", func() {
		var collection *mongo.Collection

		BeforeEach(func() {
			err := fx.New(
				fx.NopLogger,
				fx.Supply(mongoConfig),
				fx.Provide(logNull.NewLogger),
				fx.Provide(mongoofficial.NewStore),
				fx.Provide(prescriptionStoreMongo.NewStore),
				fx.Invoke(func(str prescriptionStore.Store) {
					store = str.(*prescriptionStoreMongo.PrescriptionStore)
				}),
			).Start(context.Background())
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())

			clientOptions := options.Client().ApplyURI(mongoConfig.AsConnectionString())
			client, err := mongo.Connect(context.Background(), clientOptions)
			Expect(err).ToNot(HaveOccurred())
			Expect(client).ToNot(BeNil())

			collection = client.Database(mongoConfig.Database).Collection(mongoConfig.CollectionPrefix + "prescriptions")
		})

		Context("CreateIndexes", func() {
			It("returns successfully", func() {
				ctx := context.Background()
				Expect(store.CreateIndexes(ctx)).To(Succeed())

				cur, err := collection.Indexes().List(ctx)
				Expect(err).ToNot(HaveOccurred())
				Expect(cur).ToNot(BeNil())

				indexes := make([]bson.M, 0)
				err = cur.All(nil, &indexes)
				Expect(err).ToNot(HaveOccurred())
				Expect(indexes).To(ConsistOf(
					MatchKeys(IgnoreExtras, Keys{
						"key": HaveKey("_id"),
					}),
					MatchKeys(IgnoreExtras, Keys{
						"key":        HaveKey("patientId"),
						"name":       Equal("GetByPatientId"),
						"background": Equal(true),
					}),
					MatchKeys(IgnoreExtras, Keys{
						"key":        HaveKey("prescriberId"),
						"name":       Equal("GetByPrescriberId"),
						"background": Equal(true),
					}),
					MatchKeys(IgnoreExtras, Keys{
						"key":        HaveKey("createdUserId"),
						"name":       Equal("GetByCreatedUserId"),
						"background": Equal(true),
					}),
					MatchKeys(IgnoreExtras, Keys{
						"key":        HaveKey("accessCode"),
						"name":       Equal("GetByUniqueAccessCode"),
						"unique":     Equal(true),
						"sparse":     Equal(true),
						"background": Equal(true),
					}),
					MatchKeys(IgnoreExtras, Keys{
						"key":        HaveKey("latestRevision.attributes.email"),
						"name":       Equal("GetByPatientEmail"),
						"background": Equal(true),
					}),
					MatchKeys(IgnoreExtras, Keys{
						"key": MatchKeys(IgnoreExtras, Keys{
							"_id":                        BeEquivalentTo(1),
							"revisionHistory.revisionId": BeEquivalentTo(1),
						}),
						"name":       Equal("UniqueRevisionId"),
						"unique":     Equal(true),
						"background": Equal(true),
					}),
				))
			})
		})

		Context("GetPrescriptionRepository", func() {
			It("returns a repository", func() {
				repository = store.GetPrescriptionRepository()
				Expect(repository).ToNot(BeNil())
			})
		})

		Context("with a new repository", func() {
			var ctx context.Context

			BeforeEach(func() {
				Expect(store.CreateIndexes(context.Background())).To(Succeed())
				repository = store.GetPrescriptionRepository()
				ctx = log.NewContextWithLogger(context.Background(), logger)
			})

			Context("CreatePrescription", func() {
				var userID = ""
				var revisionCreate *prescription.RevisionCreate = nil

				BeforeEach(func() {
					userID = authTest.RandomUserID()
					revisionCreate = test.RandomRevisionCreate()
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					result, err := repository.CreatePrescription(ctx, userID, revisionCreate)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the userID is missing", func() {
					userID = ""
					result, err := repository.CreatePrescription(ctx, userID, revisionCreate)
					errorsTest.ExpectEqual(err, errors.New("userID is missing"))
					Expect(result).To(BeNil())
				})

				It("returns the created prescription on success", func() {
					result, err := repository.CreatePrescription(ctx, userID, revisionCreate)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).ToNot(BeNil())
				})
			})

			Context("ListPrescriptions", func() {
				It("returns an error when the context is missing", func() {
					ctx = nil
					result, err := repository.ListPrescriptions(ctx, nil, nil)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error with invalid filter", func() {
					patient := userTest.RandomUser()
					patient.Roles = &[]string{}
					filter, err := prescription.NewFilter(patient)
					Expect(err).ToNot(HaveOccurred())
					filter.PatientID = userTest.RandomID()

					result, err := repository.ListPrescriptions(ctx, filter, nil)
					errorsTest.ExpectEqual(err, errors.New("filter is invalid"))
					Expect(result).To(BeNil())
				})

				Context("With pre-existing data", func() {
					count := 5
					var clinician *user.User
					var prescriptions prescription.Prescriptions
					var ids []primitive.ObjectID

					BeforeEach(func() {
						_, err := collection.DeleteMany(nil, bson.M{})
						Expect(err).ToNot(HaveOccurred())

						clinician = userTest.RandomUser()
						clinician.Roles = &[]string{user.RoleClinic}

						prescriptions = test.RandomPrescriptions(count)
						ids = make([]primitive.ObjectID, count)
						for i := 0; i < count; i++ {
							p := prescriptions[i]
							p.PatientID = ""
							p.State = prescription.StateSubmitted
							p.CreatedUserID = *clinician.UserID

							_, err := collection.InsertOne(nil, p)
							Expect(err).ToNot(HaveOccurred())

							ids[i] = p.ID
						}
					})

					AfterEach(func() {
						_, err := collection.DeleteMany(nil, bson.M{"id": bson.M{"$in": ids}})
						Expect(err).ToNot(HaveOccurred())
					})

					It("returns the correct prescriptions when prescriber id matches the clinician id", func() {
						_, err := collection.UpdateMany(nil, bson.M{}, bson.M{"$set": bson.M{"createdUserId": userTest.RandomID(), "prescriberUserId": userTest.RandomID()}})
						Expect(err).ToNot(HaveOccurred())

						expectedIDs := ids[1:3]
						_, err = collection.UpdateMany(nil, bson.M{"_id": bson.M{"$in": expectedIDs}}, bson.M{"$set": bson.M{"prescriberId": *clinician.UserID}})
						Expect(err).ToNot(HaveOccurred())

						filter, err := prescription.NewFilter(clinician)
						Expect(err).ToNot(HaveOccurred())
						result, err := repository.ListPrescriptions(ctx, filter, nil)
						Expect(err).ToNot(HaveOccurred())
						ExpectPrescriptionIdsToMatch(result, expectedIDs)
					})

					It("returns the correct prescriptions when created user id matches the clinician id", func() {
						_, err := collection.UpdateMany(nil, bson.M{}, bson.M{"$set": bson.M{"createdUserId": userTest.RandomID(), "prescriberUserId": userTest.RandomID()}})
						Expect(err).ToNot(HaveOccurred())

						expectedIDs := ids[1:3]
						_, err = collection.UpdateMany(nil, bson.M{"_id": bson.M{"$in": expectedIDs}}, bson.M{"$set": bson.M{"createdUserId": *clinician.UserID}})
						Expect(err).ToNot(HaveOccurred())

						filter, err := prescription.NewFilter(clinician)
						Expect(err).ToNot(HaveOccurred())
						result, err := repository.ListPrescriptions(ctx, filter, nil)
						Expect(err).ToNot(HaveOccurred())
						ExpectPrescriptionIdsToMatch(result, expectedIDs)
					})

					It("returns the correct prescriptions given a prescription state", func() {
						expectedPrescription := prescriptions[faker.RandomInt(0, count-1)]
						expectedIDs := []primitive.ObjectID{expectedPrescription.ID}
						expectedState := prescription.StateDraft

						_, err := collection.UpdateMany(nil, bson.M{"_id": bson.M{"$in": expectedIDs}}, bson.M{"$set": bson.M{"state": expectedState}})
						Expect(err).ToNot(HaveOccurred())

						filter, err := prescription.NewFilter(clinician)
						Expect(err).ToNot(HaveOccurred())
						filter.State = expectedState
						result, err := repository.ListPrescriptions(ctx, filter, nil)
						Expect(err).ToNot(HaveOccurred())
						ExpectPrescriptionIdsToMatch(result, expectedIDs)
					})

					It("returns the correct prescriptions given a prescription id", func() {
						expectedPrescription := prescriptions[faker.RandomInt(0, count-1)]
						expectedIDs := []primitive.ObjectID{expectedPrescription.ID}

						filter, err := prescription.NewFilter(clinician)
						Expect(err).ToNot(HaveOccurred())
						filter.ID = expectedPrescription.ID.Hex()

						result, err := repository.ListPrescriptions(ctx, filter, nil)
						Expect(err).ToNot(HaveOccurred())
						ExpectPrescriptionIdsToMatch(result, expectedIDs)
					})

					It("does not return deleted prescriptions", func() {
						indexToDelete := faker.RandomInt(0, count-1)
						prescriptionToDelete := prescriptions[indexToDelete]
						expectedIDs := append(ids[:indexToDelete], ids[indexToDelete+1:]...)

						_, err := collection.UpdateMany(nil, bson.M{"_id": prescriptionToDelete.ID}, bson.M{"$set": bson.M{"deletedTime": time.Now()}})
						Expect(err).ToNot(HaveOccurred())

						filter, err := prescription.NewFilter(clinician)
						Expect(err).ToNot(HaveOccurred())

						result, err := repository.ListPrescriptions(ctx, filter, nil)
						Expect(err).ToNot(HaveOccurred())
						ExpectPrescriptionIdsToMatch(result, expectedIDs)
					})

					It("returns the correct prescriptions given a patient id", func() {
						expectedIDs := ids[1:3]
						patientID := userTest.RandomID()

						_, err := collection.UpdateMany(nil, bson.M{"_id": bson.M{"$in": expectedIDs}}, bson.M{"$set": bson.M{"patientId": patientID}})
						Expect(err).ToNot(HaveOccurred())

						filter, err := prescription.NewFilter(clinician)
						Expect(err).ToNot(HaveOccurred())
						filter.PatientID = patientID

						result, err := repository.ListPrescriptions(ctx, filter, nil)
						Expect(err).ToNot(HaveOccurred())
						ExpectPrescriptionIdsToMatch(result, expectedIDs)
					})

					It("returns the correct prescriptions given a patient email", func() {
						expectedIDs := ids[1:3]
						patientEmail := faker.Internet().Email()

						_, err := collection.UpdateMany(nil, bson.M{"_id": bson.M{"$in": expectedIDs}}, bson.M{"$set": bson.M{"latestRevision.attributes.email": patientEmail}})
						Expect(err).ToNot(HaveOccurred())

						filter, err := prescription.NewFilter(clinician)
						Expect(err).ToNot(HaveOccurred())
						filter.PatientEmail = patientEmail

						result, err := repository.ListPrescriptions(ctx, filter, nil)
						Expect(err).ToNot(HaveOccurred())
						ExpectPrescriptionIdsToMatch(result, expectedIDs)
					})

					It("returns the correct prescriptions given a created start date", func() {
						sort.SliceStable(prescriptions, func(i int, j int) bool {
							return prescriptions[i].CreatedTime.Before(prescriptions[j].CreatedTime)
						})

						expectedIDs := ids[3:5]

						filter, err := prescription.NewFilter(clinician)
						Expect(err).ToNot(HaveOccurred())
						filter.CreatedTimeStart = &prescriptions[2].CreatedTime

						result, err := repository.ListPrescriptions(ctx, filter, nil)
						Expect(err).ToNot(HaveOccurred())
						ExpectPrescriptionIdsToMatch(result, expectedIDs)
					})

					It("returns the correct prescriptions given a created end date", func() {
						sort.SliceStable(prescriptions, func(i int, j int) bool {
							return prescriptions[i].CreatedTime.Before(prescriptions[j].CreatedTime)
						})

						expectedIDs := ids[0:2]

						filter, err := prescription.NewFilter(clinician)
						Expect(err).ToNot(HaveOccurred())
						filter.CreatedTimeEnd = &prescriptions[2].CreatedTime

						result, err := repository.ListPrescriptions(ctx, filter, nil)
						Expect(err).ToNot(HaveOccurred())
						ExpectPrescriptionIdsToMatch(result, expectedIDs)
					})

					It("returns the correct prescriptions given a modified start date", func() {
						sort.SliceStable(prescriptions, func(i int, j int) bool {
							return prescriptions[i].LatestRevision.Attributes.ModifiedTime.Before(prescriptions[j].LatestRevision.Attributes.ModifiedTime)
						})

						expectedIDs := ids[3:5]

						filter, err := prescription.NewFilter(clinician)
						Expect(err).ToNot(HaveOccurred())
						filter.ModifiedTimeStart = &prescriptions[2].LatestRevision.Attributes.ModifiedTime

						result, err := repository.ListPrescriptions(ctx, filter, nil)
						Expect(err).ToNot(HaveOccurred())
						ExpectPrescriptionIdsToMatch(result, expectedIDs)
					})

					It("returns the correct prescriptions given a modified end date", func() {
						sort.SliceStable(prescriptions, func(i int, j int) bool {
							return prescriptions[i].LatestRevision.Attributes.ModifiedTime.Before(prescriptions[j].LatestRevision.Attributes.ModifiedTime)
						})

						expectedIDs := ids[0:2]

						filter, err := prescription.NewFilter(clinician)
						Expect(err).ToNot(HaveOccurred())
						filter.ModifiedTimeEnd = &prescriptions[2].LatestRevision.Attributes.ModifiedTime

						result, err := repository.ListPrescriptions(ctx, filter, nil)
						Expect(err).ToNot(HaveOccurred())
						ExpectPrescriptionIdsToMatch(result, expectedIDs)
					})

					It("returns the correct patient prescriptions", func() {
						expectedIDs := ids[1:3]
						patient := userTest.RandomUser()
						patientID := patient.UserID

						_, err := collection.UpdateMany(nil, bson.M{"_id": bson.M{"$in": expectedIDs}}, bson.M{"$set": bson.M{"patientId": patientID}})
						Expect(err).ToNot(HaveOccurred())

						filter, err := prescription.NewFilter(patient)
						Expect(err).ToNot(HaveOccurred())

						result, err := repository.ListPrescriptions(ctx, filter, nil)
						Expect(err).ToNot(HaveOccurred())
						ExpectPrescriptionIdsToMatch(result, expectedIDs)
					})
				})
			})

			Context("DeletePrescription", func() {
				It("returns an error when the context is missing", func() {
					ctx = nil
					result, err := repository.DeletePrescription(ctx, "", "")
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeFalse())
				})

				It("returns an error when the clinician id is empty", func() {
					result, err := repository.DeletePrescription(ctx, "", "")
					errorsTest.ExpectEqual(err, errors.New("clinician id is missing"))
					Expect(result).To(BeFalse())
				})

				Context("With pre-existing data", func() {
					count := 5
					var prescriptions prescription.Prescriptions
					var prescr *prescription.Prescription
					var ids []primitive.ObjectID

					BeforeEach(func() {
						prescriptions = test.RandomPrescriptions(count)
						prescr = prescriptions[faker.RandomInt(0, count-1)]
						ids = make([]primitive.ObjectID, count)
						for i := 0; i < count; i++ {
							p := prescriptions[i]
							p.PatientID = ""
							p.State = faker.RandomChoice([]string{prescription.StateDraft, prescription.StatePending})
							p.DeletedTime = nil
							p.DeletedUserID = ""

							_, err := collection.InsertOne(nil, p)
							Expect(err).ToNot(HaveOccurred())
							ids[i] = p.ID
						}
					})

					AfterEach(func() {
						changeInfo, err := collection.DeleteMany(nil, bson.M{"_id": bson.M{"$in": ids}})
						Expect(err).ToNot(HaveOccurred())
						Expect(changeInfo.DeletedCount).To(Equal(int64(count)))
					})

					It("deletes the correct prescriptions given a prescriber id", func() {
						success, err := repository.DeletePrescription(ctx, prescr.PrescriberUserID, prescr.ID.Hex())
						Expect(err).ToNot(HaveOccurred())
						Expect(success).To(BeTrue())

						deletedSelector := bson.M{
							"_id":           prescr.ID,
							"deletedTime":   bson.M{"$ne": nil},
							"deletedUserId": prescr.PrescriberUserID,
						}

						found, err := collection.CountDocuments(nil, deletedSelector)
						Expect(err).ToNot(HaveOccurred())
						Expect(found).To(Equal(int64(1)))
					})

					It("deletes the correct prescriptions given a created user id", func() {
						success, err := repository.DeletePrescription(ctx, prescr.CreatedUserID, prescr.ID.Hex())
						Expect(err).ToNot(HaveOccurred())
						Expect(success).To(BeTrue())

						deletedSelector := bson.M{
							"_id":           prescr.ID,
							"deletedTime":   bson.M{"$ne": nil},
							"deletedUserId": prescr.CreatedUserID,
						}
						found, err := collection.CountDocuments(nil, deletedSelector)
						Expect(err).ToNot(HaveOccurred())
						Expect(found).To(Equal(int64(1)))
					})

					It("does not delete a prescription which is already deleted", func() {
						success, err := repository.DeletePrescription(ctx, prescr.CreatedUserID, prescr.ID.Hex())
						Expect(err).ToNot(HaveOccurred())
						Expect(success).To(BeTrue())

						success, err = repository.DeletePrescription(ctx, prescr.CreatedUserID, prescr.ID.Hex())
						Expect(err).ToNot(HaveOccurred())
						Expect(success).To(BeFalse())
					})

					It("does not delete a prescription which is submitted", func() {
						_, err := collection.UpdateOne(nil, bson.M{"_id": prescr.ID}, bson.M{"$set": bson.M{"state": prescription.StateSubmitted}})
						Expect(err).ToNot(HaveOccurred())

						success, err := repository.DeletePrescription(ctx, prescr.CreatedUserID, prescr.ID.Hex())
						Expect(err).ToNot(HaveOccurred())
						Expect(success).To(BeFalse())
					})

					It("does not delete a prescription which is reviewed", func() {
						_, err := collection.UpdateOne(nil, bson.M{"_id": prescr.ID}, bson.M{"$set": bson.M{"state": prescription.StateReviewed}})
						Expect(err).ToNot(HaveOccurred())

						success, err := repository.DeletePrescription(ctx, prescr.CreatedUserID, prescr.ID.Hex())
						Expect(err).ToNot(HaveOccurred())
						Expect(success).To(BeFalse())
					})

					It("does not delete a prescription which is active", func() {
						_, err := collection.UpdateOne(nil, bson.M{"_id": prescr.ID}, bson.M{"$set": bson.M{"state": prescription.StateActive}})
						Expect(err).ToNot(HaveOccurred())

						success, err := repository.DeletePrescription(ctx, prescr.CreatedUserID, prescr.ID.Hex())
						Expect(err).ToNot(HaveOccurred())
						Expect(success).To(BeFalse())
					})

					It("does not delete a prescription which is inactive", func() {
						_, err := collection.UpdateOne(nil, bson.M{"_id": prescr.ID}, bson.M{"$set": bson.M{"state": prescription.StateInactive}})
						Expect(err).ToNot(HaveOccurred())

						success, err := repository.DeletePrescription(ctx, prescr.CreatedUserID, prescr.ID.Hex())
						Expect(err).ToNot(HaveOccurred())
						Expect(success).To(BeFalse())
					})

					It("does not delete a prescription which is expired", func() {
						_, err := collection.UpdateOne(nil, bson.M{"_id": prescr.ID}, bson.M{"$set": bson.M{"state": prescription.StateExpired}})
						Expect(err).ToNot(HaveOccurred())

						success, err := repository.DeletePrescription(ctx, prescr.CreatedUserID, prescr.ID.Hex())
						Expect(err).ToNot(HaveOccurred())
						Expect(success).To(BeFalse())
					})
				})
			})

			Context("AddRevision", func() {
				var usr *user.User
				var prescr *prescription.Prescription
				var prescrID string
				var create *prescription.RevisionCreate

				BeforeEach(func() {
					usr = userTest.RandomUser()
					usr.Roles = &[]string{user.RoleClinic}
					prescr = test.RandomPrescription()
					prescr.State = prescription.StateDraft
					prescr.CreatedUserID = *usr.UserID
					prescrID = prescr.ID.Hex()
					create = test.RandomRevisionCreate()
					create.State = prescription.StatePending
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					result, err := repository.AddRevision(ctx, usr, prescrID, create)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the user is nil", func() {
					result, err := repository.AddRevision(ctx, nil, prescrID, create)
					errorsTest.ExpectEqual(err, errors.New("user is missing"))
					Expect(result).To(BeNil())
				})

				Context("With pre-existing data", func() {
					BeforeEach(func() {
						prescr.ID = primitive.NewObjectID()
						prescrID = prescr.ID.Hex()
						_, err := collection.InsertOne(nil, prescr)
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						_, err := collection.DeleteMany(nil, bson.M{"_id": prescrID})
						Expect(err).ToNot(HaveOccurred())
					})

					It("returns nil if the prescription doesn't exist", func() {
						randomID := primitive.NewObjectID().Hex()
						result, err := repository.AddRevision(ctx, usr, randomID, create)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).To(BeNil())
					})

					It("returns the result on success", func() {
						result, err := repository.AddRevision(ctx, usr, prescrID, create)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
					})

					It("adds a revision to the list of revisions", func() {
						result, err := repository.AddRevision(ctx, usr, prescrID, create)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
						Expect(result.RevisionHistory).To(HaveLen(2))
					})

					It("does not prepend the new revision to the revision history array", func() {
						result, err := repository.AddRevision(ctx, usr, prescrID, create)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
						Expect(result.RevisionHistory[0].RevisionID).To(Equal(0))
					})

					It("appends the latest revision to the newly created revision", func() {
						result, err := repository.AddRevision(ctx, usr, prescrID, create)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
						Expect(result.RevisionHistory[1].RevisionID).To(Equal(1))
					})

					It("sets the revision attributes correctly", func() {
						result, err := repository.AddRevision(ctx, usr, prescrID, create)
						update := prescription.NewPrescriptionAddRevisionUpdate(usr, prescr, create)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())

						result.LatestRevision.Attributes.ModifiedTime = update.Revision.Attributes.ModifiedTime
						Expect(*result.LatestRevision.Attributes).To(Equal(*update.Revision.Attributes))
					})
				})
			})

			Context("ClaimPrescription", func() {
				var usr *user.User
				var prescr *prescription.Prescription
				var claim *prescription.Claim

				BeforeEach(func() {
					usr = userTest.RandomUser()
					prescr = test.RandomPrescription()
					prescr.State = prescription.StateSubmitted
					claim = prescription.NewPrescriptionClaim()
					claim.AccessCode = prescr.AccessCode
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					result, err := repository.ClaimPrescription(ctx, usr, claim)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the user is nil", func() {
					result, err := repository.ClaimPrescription(ctx, nil, claim)
					errorsTest.ExpectEqual(err, errors.New("user is missing"))
					Expect(result).To(BeNil())
				})

				Context("With pre-existing data", func() {
					BeforeEach(func() {
						prescr.ID = primitive.NewObjectID()
						_, err := collection.InsertOne(nil, prescr)
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						_, err := collection.DeleteMany(nil, bson.M{"_id": prescr.ID})
						Expect(err).ToNot(HaveOccurred())
					})

					It("returns nil if the access code is incorrect", func() {
						claim.AccessCode = "XXXXXX"
						result, err := repository.ClaimPrescription(ctx, usr, claim)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).To(BeNil())
					})

					It("returns the prescription on success", func() {
						result, err := repository.ClaimPrescription(ctx, usr, claim)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
					})

					It("resets the access code", func() {
						result, err := repository.ClaimPrescription(ctx, usr, claim)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
						Expect(result.AccessCode).To(BeEmpty())
					})

					It("sets the state of the prescription to reviewed", func() {
						result, err := repository.ClaimPrescription(ctx, usr, claim)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
						Expect(result.State).To(Equal(prescription.StateReviewed))
					})

					It("sets the patient id", func() {
						result, err := repository.ClaimPrescription(ctx, usr, claim)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
						Expect(result.PatientID).To(Equal(*usr.UserID))
					})
				})
			})

			Context("UpdateState", func() {
				var usr *user.User
				var prescr *prescription.Prescription
				var prescrID string
				var stateUpdate *prescription.StateUpdate

				BeforeEach(func() {
					usr = userTest.RandomUser()
					prescr = test.RandomClaimedPrescription()
					prescr.PatientID = *usr.UserID
					prescrID = prescr.ID.Hex()
					stateUpdate = prescription.NewStateUpdate()
					stateUpdate.State = prescription.StateActive
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					result, err := repository.UpdatePrescriptionState(ctx, usr, prescrID, stateUpdate)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the user is nil", func() {
					result, err := repository.UpdatePrescriptionState(ctx, nil, prescrID, stateUpdate)
					errorsTest.ExpectEqual(err, errors.New("user is missing"))
					Expect(result).To(BeNil())
				})

				Context("With pre-existing data", func() {
					BeforeEach(func() {
						prescr.ID = primitive.NewObjectID()
						prescrID = prescr.ID.Hex()
						_, err := collection.InsertOne(nil, prescr)
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						_, err := collection.DeleteMany(nil, bson.M{"_id": prescr.ID})
						Expect(err).ToNot(HaveOccurred())
					})

					It("returns the prescription on success", func() {
						result, err := repository.UpdatePrescriptionState(ctx, usr, prescrID, stateUpdate)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
					})

					It("returns nil when trying to activate an already active prescription", func() {
						result, err := repository.UpdatePrescriptionState(ctx, usr, prescrID, stateUpdate)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())

						result, err = repository.UpdatePrescriptionState(ctx, usr, prescrID, stateUpdate)
						errorsTest.ExpectEqual(err, errors.New("the prescription update is invalid"))
					})
				})
			})
		})
	})
})

func ExpectPrescriptionIdsToMatch(actual prescription.Prescriptions, expected []primitive.ObjectID) {
	Expect(actual).To(HaveLen(len(expected)))

	for i := 0; i < len(expected); i++ {
		Expect(actual[i].ID).To(Equal(expected[i]))
	}
}
