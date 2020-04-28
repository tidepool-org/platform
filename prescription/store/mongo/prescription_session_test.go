package mongo_test

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.uber.org/fx/fxtest"

	logNull "github.com/tidepool-org/platform/log/null"

	"github.com/tidepool-org/platform/user"

	"github.com/globalsign/mgo/bson"
	"syreclabs.com/go/faker"

	userTest "github.com/tidepool-org/platform/user/test"

	"github.com/globalsign/mgo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	authTest "github.com/tidepool-org/platform/auth/test"
	configTest "github.com/tidepool-org/platform/config/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/prescription"
	prescriptionStore "github.com/tidepool-org/platform/prescription/store"
	prescriptionStoreMongo "github.com/tidepool-org/platform/prescription/store/mongo"
	"github.com/tidepool-org/platform/prescription/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
)

var _ = Describe("PrescriptionSession", func() {
	var mongoConfig *storeStructuredMongo.Config
	var store *prescriptionStoreMongo.PrescriptionStore
	var configReporter *configTest.Reporter
	var logger *logTest.Logger
	var session prescriptionStore.PrescriptionSession

	BeforeEach(func() {
		logger = logTest.NewLogger()
		mongoConfig = storeStructuredMongoTest.NewConfig()
		prescriptionStoreConfig := map[string]interface{}{
			"addresses":         strings.Join(mongoConfig.Addresses, ","),
			"collection_prefix": mongoConfig.CollectionPrefix,
			"database":          mongoConfig.Database,
			"tls":               fmt.Sprintf("%v", mongoConfig.TLS),
			"timeout":           fmt.Sprintf("%v", int(mongoConfig.Timeout.Seconds())),
		}
		serviceConfig := map[string]interface{}{
			"prescription": map[string]interface{}{
				"store": prescriptionStoreConfig,
			},
		}

		configReporter = configTest.NewReporter()
		configReporter.Config = serviceConfig
	})

	AfterEach(func() {
		if session != nil {
			session.Close()
		}
		if store != nil {
			store.Close()
		}
	})

	Context("with a new store", func() {
		var mgoSession *mgo.Session
		var mgoCollection *mgo.Collection

		BeforeEach(func() {
			prescrStr, err := prescriptionStoreMongo.NewStore(prescriptionStoreMongo.Params{
				ConfigReporter: configReporter,
				Logger:         logNull.NewLogger(),
				Lifestyle:      fxtest.NewLifecycle(GinkgoT()),
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(prescrStr).ToNot(BeNil())

			store = prescrStr.(*prescriptionStoreMongo.PrescriptionStore)
			err = store.Initialize()
			Expect(err).ToNot(HaveOccurred())

			mgoSession = storeStructuredMongoTest.Session().Copy()
			mgoCollection = mgoSession.DB(mongoConfig.Database).C(mongoConfig.CollectionPrefix + "prescriptions")
		})

		AfterEach(func() {
			if mgoSession != nil {
				mgoSession.Close()
			}
		})

		Context("EnsureIndexes", func() {
			It("returns successfully", func() {
				Expect(store.EnsureIndexes()).To(Succeed())
				indexes, err := mgoCollection.Indexes()
				Expect(err).ToNot(HaveOccurred())
				Expect(indexes).To(ContainElement(
					MatchFields(IgnoreExtras, Fields{"Key": ConsistOf("_id")}),
				))
			})
		})

		Context("NewSession", func() {
			It("returns a new session", func() {
				session = store.NewPrescriptionSession()
				Expect(session).ToNot(BeNil())
			})
		})

		Context("with a new session", func() {
			var ctx context.Context

			BeforeEach(func() {
				Expect(store.EnsureIndexes()).To(Succeed())
				session = store.NewPrescriptionSession()
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
					result, err := session.CreatePrescription(ctx, userID, revisionCreate)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the userID is missing", func() {
					userID = ""
					result, err := session.CreatePrescription(ctx, userID, revisionCreate)
					errorsTest.ExpectEqual(err, errors.New("userID is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the session is closed", func() {
					session.Close()
					result, err := session.CreatePrescription(ctx, userID, revisionCreate)
					errorsTest.ExpectEqual(err, errors.New("session closed"))
					Expect(result).To(BeNil())
				})

				It("returns the created prescription on success", func() {
					result, err := session.CreatePrescription(ctx, userID, revisionCreate)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).ToNot(BeNil())
				})
			})

			Context("ListPrescriptions", func() {
				It("returns an error when the context is missing", func() {
					ctx = nil
					result, err := session.ListPrescriptions(ctx, nil, nil)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the session is closed", func() {
					session.Close()
					result, err := session.ListPrescriptions(ctx, nil, nil)
					errorsTest.ExpectEqual(err, errors.New("session closed"))
					Expect(result).To(BeNil())
				})

				It("returns an error with invalid filter", func() {
					patient := userTest.RandomUser()
					patient.Roles = &[]string{}
					filter, err := prescription.NewFilter(patient)
					Expect(err).ToNot(HaveOccurred())
					filter.PatientID = userTest.RandomID()

					result, err := session.ListPrescriptions(ctx, filter, nil)
					errorsTest.ExpectEqual(err, errors.New("filter is invalid"))
					Expect(result).To(BeNil())
				})

				Context("With pre-existing data", func() {
					count := 5
					var clinician *user.User
					var prescriptions prescription.Prescriptions
					var ids []bson.ObjectId

					BeforeEach(func() {
						_, err := mgoCollection.RemoveAll(bson.M{})
						Expect(err).ToNot(HaveOccurred())

						clinician = userTest.RandomUser()
						clinician.Roles = &[]string{user.RoleClinic}

						prescriptions = test.RandomPrescriptions(count)
						ids = make([]bson.ObjectId, count)
						for i := 0; i < count; i++ {
							p := prescriptions[i]
							p.PatientID = ""
							p.State = prescription.StateSubmitted
							p.CreatedUserID = *clinician.UserID

							err := mgoCollection.Insert(p)
							Expect(err).ToNot(HaveOccurred())

							ids[i] = p.ID
						}
					})

					AfterEach(func() {
						_, err := mgoCollection.RemoveAll(bson.M{"id": bson.M{"$in": ids}})
						Expect(err).ToNot(HaveOccurred())
					})

					It("returns the correct prescriptions when prescriber id matches the clinician id", func() {
						_, err := mgoCollection.UpdateAll(bson.M{}, bson.M{"$set": bson.M{"createdUserId": userTest.RandomID(), "prescriberUserId": userTest.RandomID()}})
						Expect(err).ToNot(HaveOccurred())

						expectedIDs := ids[1:3]
						_, err = mgoCollection.UpdateAll(bson.M{"_id": bson.M{"$in": expectedIDs}}, bson.M{"$set": bson.M{"prescriberId": *clinician.UserID}})
						Expect(err).ToNot(HaveOccurred())

						filter, err := prescription.NewFilter(clinician)
						Expect(err).ToNot(HaveOccurred())
						result, err := session.ListPrescriptions(ctx, filter, nil)
						Expect(err).ToNot(HaveOccurred())
						ExpectPrescriptionIdsToMatch(result, expectedIDs)
					})

					It("returns the correct prescriptions when created user id matches the clinician id", func() {
						_, err := mgoCollection.UpdateAll(bson.M{}, bson.M{"$set": bson.M{"createdUserId": userTest.RandomID(), "prescriberUserId": userTest.RandomID()}})
						Expect(err).ToNot(HaveOccurred())

						expectedIDs := ids[1:3]
						_, err = mgoCollection.UpdateAll(bson.M{"_id": bson.M{"$in": expectedIDs}}, bson.M{"$set": bson.M{"createdUserId": *clinician.UserID}})
						Expect(err).ToNot(HaveOccurred())

						filter, err := prescription.NewFilter(clinician)
						Expect(err).ToNot(HaveOccurred())
						result, err := session.ListPrescriptions(ctx, filter, nil)
						Expect(err).ToNot(HaveOccurred())
						ExpectPrescriptionIdsToMatch(result, expectedIDs)
					})

					It("returns the correct prescriptions given a prescription state", func() {
						expectedPrescription := prescriptions[faker.RandomInt(0, count-1)]
						expectedIDs := []bson.ObjectId{expectedPrescription.ID}
						expectedState := prescription.StateDraft

						_, err := mgoCollection.UpdateAll(bson.M{"_id": bson.M{"$in": expectedIDs}}, bson.M{"$set": bson.M{"state": expectedState}})
						Expect(err).ToNot(HaveOccurred())

						filter, err := prescription.NewFilter(clinician)
						Expect(err).ToNot(HaveOccurred())
						filter.State = expectedState
						result, err := session.ListPrescriptions(ctx, filter, nil)
						Expect(err).ToNot(HaveOccurred())
						ExpectPrescriptionIdsToMatch(result, expectedIDs)
					})

					It("returns the correct prescriptions given a prescription id", func() {
						expectedPrescription := prescriptions[faker.RandomInt(0, count-1)]
						expectedIDs := []bson.ObjectId{expectedPrescription.ID}

						filter, err := prescription.NewFilter(clinician)
						Expect(err).ToNot(HaveOccurred())
						filter.ID = expectedPrescription.ID.Hex()

						result, err := session.ListPrescriptions(ctx, filter, nil)
						Expect(err).ToNot(HaveOccurred())
						ExpectPrescriptionIdsToMatch(result, expectedIDs)
					})

					It("does not return deleted prescriptions", func() {
						indexToDelete := faker.RandomInt(0, count-1)
						prescriptionToDelete := prescriptions[indexToDelete]
						expectedIDs := append(ids[:indexToDelete], ids[indexToDelete+1:]...)

						_, err := mgoCollection.UpdateAll(bson.M{"_id": prescriptionToDelete.ID}, bson.M{"$set": bson.M{"deletedTime": time.Now()}})
						Expect(err).ToNot(HaveOccurred())

						filter, err := prescription.NewFilter(clinician)
						Expect(err).ToNot(HaveOccurred())

						result, err := session.ListPrescriptions(ctx, filter, nil)
						Expect(err).ToNot(HaveOccurred())
						ExpectPrescriptionIdsToMatch(result, expectedIDs)
					})

					It("returns the correct prescriptions given a patient id", func() {
						expectedIDs := ids[1:3]
						patientID := userTest.RandomID()

						_, err := mgoCollection.UpdateAll(bson.M{"_id": bson.M{"$in": expectedIDs}}, bson.M{"$set": bson.M{"patientId": patientID}})
						Expect(err).ToNot(HaveOccurred())

						filter, err := prescription.NewFilter(clinician)
						Expect(err).ToNot(HaveOccurred())
						filter.PatientID = patientID

						result, err := session.ListPrescriptions(ctx, filter, nil)
						Expect(err).ToNot(HaveOccurred())
						ExpectPrescriptionIdsToMatch(result, expectedIDs)
					})

					It("returns the correct prescriptions given a patient email", func() {
						expectedIDs := ids[1:3]
						patientEmail := faker.Internet().Email()

						_, err := mgoCollection.UpdateAll(bson.M{"_id": bson.M{"$in": expectedIDs}}, bson.M{"$set": bson.M{"latestRevision.attributes.email": patientEmail}})
						Expect(err).ToNot(HaveOccurred())

						filter, err := prescription.NewFilter(clinician)
						Expect(err).ToNot(HaveOccurred())
						filter.PatientEmail = patientEmail

						result, err := session.ListPrescriptions(ctx, filter, nil)
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

						result, err := session.ListPrescriptions(ctx, filter, nil)
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

						result, err := session.ListPrescriptions(ctx, filter, nil)
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

						result, err := session.ListPrescriptions(ctx, filter, nil)
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

						result, err := session.ListPrescriptions(ctx, filter, nil)
						Expect(err).ToNot(HaveOccurred())
						ExpectPrescriptionIdsToMatch(result, expectedIDs)
					})

					It("returns the correct patient prescriptions", func() {
						expectedIDs := ids[1:3]
						patient := userTest.RandomUser()
						patientID := patient.UserID

						_, err := mgoCollection.UpdateAll(bson.M{"_id": bson.M{"$in": expectedIDs}}, bson.M{"$set": bson.M{"patientId": patientID}})
						Expect(err).ToNot(HaveOccurred())

						filter, err := prescription.NewFilter(patient)
						Expect(err).ToNot(HaveOccurred())

						result, err := session.ListPrescriptions(ctx, filter, nil)
						Expect(err).ToNot(HaveOccurred())
						ExpectPrescriptionIdsToMatch(result, expectedIDs)
					})
				})
			})

			Context("DeletePrescription", func() {
				It("returns an error when the context is missing", func() {
					ctx = nil
					result, err := session.DeletePrescription(ctx, "", "")
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeFalse())
				})

				It("returns an error when the session is closed", func() {
					session.Close()
					result, err := session.DeletePrescription(ctx, "", "")
					errorsTest.ExpectEqual(err, errors.New("session closed"))
					Expect(result).To(BeFalse())
				})

				It("returns an error when the clinician id is empty", func() {
					result, err := session.DeletePrescription(ctx, "", "")
					errorsTest.ExpectEqual(err, errors.New("clinician id is missing"))
					Expect(result).To(BeFalse())
				})

				It("returns an error when the prescription id is empty", func() {
					result, err := session.DeletePrescription(ctx, userTest.RandomID(), "")
					errorsTest.ExpectEqual(err, errors.New("prescription id is missing"))
					Expect(result).To(BeFalse())
				})

				Context("With pre-existing data", func() {
					count := 5
					var prescriptions prescription.Prescriptions
					var prescr *prescription.Prescription
					var ids []bson.ObjectId

					BeforeEach(func() {
						prescriptions = test.RandomPrescriptions(count)
						prescr = prescriptions[faker.RandomInt(0, count-1)]
						ids = make([]bson.ObjectId, count)
						for i := 0; i < count; i++ {
							p := prescriptions[i]
							p.PatientID = ""
							p.State = faker.RandomChoice([]string{prescription.StateDraft, prescription.StatePending})
							p.DeletedTime = nil
							p.DeletedUserID = ""

							err := mgoCollection.Insert(p)
							Expect(err).ToNot(HaveOccurred())
							ids[i] = p.ID
						}
					})

					AfterEach(func() {
						changeInfo, err := mgoCollection.RemoveAll(bson.M{"_id": bson.M{"$in": ids}})
						Expect(err).ToNot(HaveOccurred())
						Expect(changeInfo.Removed).To(Equal(count))
					})

					It("deletes the correct prescriptions given a prescriber id", func() {
						success, err := session.DeletePrescription(ctx, prescr.PrescriberUserID, prescr.ID.Hex())
						Expect(err).ToNot(HaveOccurred())
						Expect(success).To(BeTrue())

						deletedSelector := bson.M{
							"_id":           prescr.ID,
							"deletedTime":   bson.M{"$ne": nil},
							"deletedUserId": prescr.PrescriberUserID,
						}
						found, err := mgoCollection.Find(deletedSelector).Count()
						Expect(err).ToNot(HaveOccurred())
						Expect(found).To(Equal(1))
					})

					It("deletes the correct prescriptions given a created user id", func() {
						success, err := session.DeletePrescription(ctx, prescr.CreatedUserID, prescr.ID.Hex())
						Expect(err).ToNot(HaveOccurred())
						Expect(success).To(BeTrue())

						deletedSelector := bson.M{
							"_id":           prescr.ID,
							"deletedTime":   bson.M{"$ne": nil},
							"deletedUserId": prescr.CreatedUserID,
						}
						found, err := mgoCollection.Find(deletedSelector).Count()
						Expect(err).ToNot(HaveOccurred())
						Expect(found).To(Equal(1))
					})

					It("does not delete a prescription which is already deleted", func() {
						success, err := session.DeletePrescription(ctx, prescr.CreatedUserID, prescr.ID.Hex())
						Expect(err).ToNot(HaveOccurred())
						Expect(success).To(BeTrue())

						success, err = session.DeletePrescription(ctx, prescr.CreatedUserID, prescr.ID.Hex())
						Expect(err).ToNot(HaveOccurred())
						Expect(success).To(BeFalse())
					})

					It("does not delete a prescription which is submitted", func() {
						err := mgoCollection.Update(bson.M{"_id": prescr.ID}, bson.M{"$set": bson.M{"state": prescription.StateSubmitted}})
						Expect(err).ToNot(HaveOccurred())

						success, err := session.DeletePrescription(ctx, prescr.CreatedUserID, prescr.ID.Hex())
						Expect(err).ToNot(HaveOccurred())
						Expect(success).To(BeFalse())
					})

					It("does not delete a prescription which is reviewed", func() {
						err := mgoCollection.Update(bson.M{"_id": prescr.ID}, bson.M{"$set": bson.M{"state": prescription.StateReviewed}})
						Expect(err).ToNot(HaveOccurred())

						success, err := session.DeletePrescription(ctx, prescr.CreatedUserID, prescr.ID.Hex())
						Expect(err).ToNot(HaveOccurred())
						Expect(success).To(BeFalse())
					})

					It("does not delete a prescription which is active", func() {
						err := mgoCollection.Update(bson.M{"_id": prescr.ID}, bson.M{"$set": bson.M{"state": prescription.StateActive}})
						Expect(err).ToNot(HaveOccurred())

						success, err := session.DeletePrescription(ctx, prescr.CreatedUserID, prescr.ID.Hex())
						Expect(err).ToNot(HaveOccurred())
						Expect(success).To(BeFalse())
					})

					It("does not delete a prescription which is inactive", func() {
						err := mgoCollection.Update(bson.M{"_id": prescr.ID}, bson.M{"$set": bson.M{"state": prescription.StateInactive}})
						Expect(err).ToNot(HaveOccurred())

						success, err := session.DeletePrescription(ctx, prescr.CreatedUserID, prescr.ID.Hex())
						Expect(err).ToNot(HaveOccurred())
						Expect(success).To(BeFalse())
					})

					It("does not delete a prescription which is expired", func() {
						err := mgoCollection.Update(bson.M{"_id": prescr.ID}, bson.M{"$set": bson.M{"state": prescription.StateExpired}})
						Expect(err).ToNot(HaveOccurred())

						success, err := session.DeletePrescription(ctx, prescr.CreatedUserID, prescr.ID.Hex())
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
					result, err := session.AddRevision(ctx, usr, prescrID, create)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the session is closed", func() {
					session.Close()
					result, err := session.AddRevision(ctx, usr, prescrID, create)
					errorsTest.ExpectEqual(err, errors.New("session closed"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the user is nil", func() {
					result, err := session.AddRevision(ctx, nil, prescrID, create)
					errorsTest.ExpectEqual(err, errors.New("user is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the prescription id is empty", func() {
					result, err := session.AddRevision(ctx, usr, "", create)
					errorsTest.ExpectEqual(err, errors.New("prescription id is missing"))
					Expect(result).To(BeNil())
				})

				Context("With pre-existing data", func() {
					BeforeEach(func() {
						prescr.ID = bson.NewObjectId()
						prescrID = prescr.ID.Hex()
						err := mgoCollection.Insert(prescr)
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						_, err := mgoCollection.RemoveAll(bson.M{"_id": prescrID})
						Expect(err).ToNot(HaveOccurred())
					})

					It("returns nil if the prescription doesn't exist", func() {
						result, err := session.AddRevision(ctx, usr, bson.NewObjectId().Hex(), create)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).To(BeNil())
					})

					It("returns the result on success", func() {
						result, err := session.AddRevision(ctx, usr, prescrID, create)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
					})

					It("adds a revision to the list of revisions", func() {
						result, err := session.AddRevision(ctx, usr, prescrID, create)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
						Expect(result.RevisionHistory).To(HaveLen(2))
					})

					It("does not prepend the new revision to the revision history array", func() {
						result, err := session.AddRevision(ctx, usr, prescrID, create)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
						Expect(result.RevisionHistory[0].RevisionID).To(Equal(0))
					})

					It("appends the latest revision to the newly created revision", func() {
						result, err := session.AddRevision(ctx, usr, prescrID, create)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
						Expect(result.RevisionHistory[1].RevisionID).To(Equal(1))
					})

					It("sets the revision attributes correctly", func() {
						result, err := session.AddRevision(ctx, usr, prescrID, create)
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
				var prescrID string
				var claim *prescription.Claim

				BeforeEach(func() {
					usr = userTest.RandomUser()
					prescr = test.RandomPrescription()
					prescr.State = prescription.StateSubmitted
					prescrID = prescr.ID.Hex()
					claim = prescription.NewPrescriptionClaim()
					claim.AccessCode = prescr.AccessCode
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					result, err := session.ClaimPrescription(ctx, usr, claim)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the session is closed", func() {
					session.Close()
					result, err := session.ClaimPrescription(ctx, usr, claim)
					errorsTest.ExpectEqual(err, errors.New("session closed"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the user is nil", func() {
					result, err := session.ClaimPrescription(ctx, nil, claim)
					errorsTest.ExpectEqual(err, errors.New("user is missing"))
					Expect(result).To(BeNil())
				})

				Context("With pre-existing data", func() {
					BeforeEach(func() {
						prescr.ID = bson.NewObjectId()
						prescrID = prescr.ID.Hex()
						err := mgoCollection.Insert(prescr)
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						_, err := mgoCollection.RemoveAll(bson.M{"_id": prescrID})
						Expect(err).ToNot(HaveOccurred())
					})

					It("returns nil if the access code is incorrect", func() {
						claim.AccessCode = "XXXXXX"
						result, err := session.ClaimPrescription(ctx, usr, claim)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).To(BeNil())
					})

					It("returns the prescription on success", func() {
						result, err := session.ClaimPrescription(ctx, usr, claim)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
					})

					It("resets the access code", func() {
						result, err := session.ClaimPrescription(ctx, usr, claim)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
						Expect(result.AccessCode).To(BeEmpty())
					})

					It("sets the state of the prescription to reviewed", func() {
						result, err := session.ClaimPrescription(ctx, usr, claim)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
						Expect(result.State).To(Equal(prescription.StateReviewed))
					})

					It("sets the patient id", func() {
						result, err := session.ClaimPrescription(ctx, usr, claim)
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
					result, err := session.UpdatePrescriptionState(ctx, usr, prescrID, stateUpdate)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the session is closed", func() {
					session.Close()
					result, err := session.UpdatePrescriptionState(ctx, usr, prescrID, stateUpdate)
					errorsTest.ExpectEqual(err, errors.New("session closed"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the user is nil", func() {
					result, err := session.UpdatePrescriptionState(ctx, nil, prescrID, stateUpdate)
					errorsTest.ExpectEqual(err, errors.New("user is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the prescription id is empty", func() {
					result, err := session.UpdatePrescriptionState(ctx, usr, "", stateUpdate)
					errorsTest.ExpectEqual(err, errors.New("prescription id is missing"))
					Expect(result).To(BeNil())
				})

				Context("With pre-existing data", func() {
					BeforeEach(func() {
						prescr.ID = bson.NewObjectId()
						prescrID = prescr.ID.Hex()
						err := mgoCollection.Insert(prescr)
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						_, err := mgoCollection.RemoveAll(bson.M{"_id": prescrID})
						Expect(err).ToNot(HaveOccurred())
					})

					It("returns the prescription on success", func() {
						result, err := session.UpdatePrescriptionState(ctx, usr, prescrID, stateUpdate)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
					})

					It("returns an error when trying to activate an already active prescription", func() {
						result, err := session.UpdatePrescriptionState(ctx, usr, prescrID, stateUpdate)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())

						result, err = session.UpdatePrescriptionState(ctx, usr, prescrID, stateUpdate)
						errorsTest.ExpectEqual(err, errors.New("the prescription update is invalid"))
					})
				})
			})
		})
	})
})

func ExpectPrescriptionIdsToMatch(actual prescription.Prescriptions, expected []bson.ObjectId) {
	Expect(actual).To(HaveLen(len(expected)))

	for i := 0; i < len(expected); i++ {
		Expect(actual[i].ID).To(Equal(expected[i]))
	}
}
