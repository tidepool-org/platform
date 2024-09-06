package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/mock/gomock"

	"github.com/tidepool-org/platform/clinics"

	"syreclabs.com/go/faker"

	prescriptionService "github.com/tidepool-org/platform/prescription/service"
	serviceTest "github.com/tidepool-org/platform/prescription/service/test"

	"github.com/ant0ine/go-json-rest/rest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"go.mongodb.org/mongo-driver/bson/primitive"

	clinic "github.com/tidepool-org/clinic/client"

	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/prescription"
	"github.com/tidepool-org/platform/prescription/api"
	prescriptionTest "github.com/tidepool-org/platform/prescription/test"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"

	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	testRest "github.com/tidepool-org/platform/test/rest"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("V1", func() {
	var ctrl *gomock.Controller
	var clinicsClient *clinics.MockClient
	var deviceSettingsValidator prescriptionService.DeviceSettingsValidator
	var prescriptionService *prescriptionTest.PrescriptionAccessor

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		clinicsClient = clinics.NewMockClient(ctrl)
		deviceSettingsValidator = serviceTest.NewNoopSettingsValidator()
		prescriptionService = prescriptionTest.NewPrescriptionAccessor()
	})

	AfterEach(func() {
		prescriptionService.Expectations()
		ctrl.Finish()
	})

	Context("NewRouter", func() {
		It("returns successfully", func() {
			Expect(api.NewRouter(api.Params{
				ClinicsClient:           clinicsClient,
				DeviceSettingsValidator: deviceSettingsValidator,
				PrescriptionService:     prescriptionService,
			})).ToNot(BeNil())
		})
	})

	Context("with new router", func() {
		var router service.Router

		BeforeEach(func() {
			router = api.NewRouter(api.Params{
				ClinicsClient:           clinicsClient,
				DeviceSettingsValidator: deviceSettingsValidator,
				PrescriptionService:     prescriptionService,
			})
			Expect(router).ToNot(BeNil())
		})

		Context("Routes", func() {
			It("returns the expected routes", func() {
				Expect(router.Routes()).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodPost), "PathExp": Equal("/v1/clinics/:clinicId/prescriptions")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodPost), "PathExp": Equal("/v1/clinics/:clinicId/prescriptions/:prescriptionId/revisions")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/clinics/:clinicId/prescriptions")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/clinics/:clinicId/prescriptions/:prescriptionId")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodDelete), "PathExp": Equal("/v1/clinics/:clinicId/prescriptions/:prescriptionId")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/patients/:userId/prescriptions")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodPost), "PathExp": Equal("/v1/patients/:userId/prescriptions")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/patients/:userId/prescriptions/:prescriptionId")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodPatch), "PathExp": Equal("/v1/patients/:userId/prescriptions/:prescriptionId")})),
				))
			})
		})

		Context("with response and request", func() {
			var res *testRest.ResponseWriter
			var req *rest.Request
			var ctx context.Context
			var handlerFunc rest.HandlerFunc
			var details request.AuthDetails

			BeforeEach(func() {
				res = testRest.NewResponseWriter()
				res.HeaderOutput = &http.Header{}
				req = testRest.NewRequest()
				ctx = log.NewContextWithLogger(req.Context(), logTest.NewLogger())
				req.Request = req.WithContext(ctx)
			})

			JustBeforeEach(func() {
				app, err := rest.MakeRouter(router.Routes()...)
				Expect(err).ToNot(HaveOccurred())
				Expect(app).ToNot(BeNil())
				handlerFunc = app.AppFunc()
			})

			AfterEach(func() {
				res.AssertOutputsEmpty()
			})

			Context("with patient and clinician", func() {
				var userID string
				var clinicID string
				var clinician *clinic.Clinician

				BeforeEach(func() {
					userID = userTest.RandomID()
					clinicianID := clinic.TidepoolUserId(userID)
					clinician = &clinic.Clinician{
						Id:    &clinicianID,
						Roles: clinic.ClinicianRoles{"PRESCRIBER"},
					}
					clinicID = faker.Number().Hexadecimal(24)
				})

				When("signed in", func() {
					asService := false

					JustBeforeEach(func() {
						if asService {
							details = request.NewAuthDetails(request.MethodServiceSecret, "", authTest.NewServiceSecret())
						} else {
							details = request.NewAuthDetails(request.MethodSessionToken, userID, "")
						}
						req.Request = req.WithContext(request.NewContextWithAuthDetails(req.Context(), details))
					})

					JustAfterEach(func() {
						asService = false
					})

					Context("create prescription", func() {
						var create *prescription.RevisionCreate
						var prescr *prescription.Prescription

						BeforeEach(func() {
							req.Method = http.MethodPost
							req.URL.Path = fmt.Sprintf("/v1/clinics/%v/prescriptions", clinicID)

							create = prescriptionTest.RandomRevisionCreate()
							create.CreatedUserID = userID
							create.ClinicianID = userID
							prescriptionTest.ResetRevisionCreateHash(create)

							prescr = prescriptionTest.RandomPrescription()
							body, err := json.Marshal(create)
							Expect(err).ToNot(HaveOccurred())

							req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						Context("as patient", func() {
							BeforeEach(func() {
								clinicsClient.EXPECT().GetClinician(gomock.Any(), clinicID, userID).Return(nil, nil)
							})

							It("returns forbidden status code", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})

						Context("as clinician", func() {
							BeforeEach(func() {
								clinicsClient.EXPECT().GetClinician(gomock.Any(), clinicID, userID).Return(clinician, nil)
							})

							It("returns created status code", func() {
								prescriptionService.CreatePrescriptionOutputs = []prescriptionTest.CreatePrescriptionOutput{{Prescription: prescr, Error: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusCreated}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})

							Context("with missing required attribute when state is 'submitted'", func() {
								BeforeEach(func() {
									create.FirstName = nil
									integrityHash := prescription.MustGenerateIntegrityHash(prescription.NewIntegrityAttributesFromRevisionCreate(*create))
									create.RevisionHash = integrityHash.Hash
									body, err := json.Marshal(create)
									Expect(err).ToNot(HaveOccurred())

									req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
								})

								It("returns bad request with validation error", func() {
									expectedErrorBody := "{\"code\":\"value-not-exists\",\"title\":\"value does not exist\",\"detail\":\"value does not exist\",\"source\":{\"pointer\":\"/firstName\"}}\n"
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									Expect(string(res.WriteInputs[0])).To(Equal(expectedErrorBody))
								})

							})
						})

						Context("as service", func() {
							BeforeEach(func() {
								asService = true
							})

							It("returns unauthorized status code", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})
					})

					Context("list clinic prescriptions", func() {
						var prescrs []*prescription.Prescription

						BeforeEach(func() {
							req.Method = http.MethodGet
							req.URL.Path = fmt.Sprintf("/v1/clinics/%v/prescriptions", clinicID)

							prescrs = []*prescription.Prescription{prescriptionTest.RandomPrescription()}
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						Context("as clinician", func() {
							BeforeEach(func() {
								clinicsClient.EXPECT().GetClinician(gomock.Any(), clinicID, userID).Return(clinician, nil)
							})

							It("filters the prescriptions for the current clinic", func() {
								prescriptionService.ListPrescriptionOutputs = []prescriptionTest.ListPrescriptionsOutput{{Prescriptions: prescrs, Err: nil}}
								handlerFunc(res, req)
								Expect(prescriptionService.ListPrescriptionsInputs).To(HaveLen(1))
								Expect(prescriptionService.ListPrescriptionsInputs[0].Filter.PatientUserID).To(BeEmpty())
								Expect(prescriptionService.ListPrescriptionsInputs[0].Filter.ClinicID).To(Equal(clinicID))
							})

							It("returns ok status code", func() {
								prescriptionService.ListPrescriptionOutputs = []prescriptionTest.ListPrescriptionsOutput{{Prescriptions: prescrs, Err: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})

						Context("as patient", func() {
							BeforeEach(func() {
								clinicsClient.EXPECT().GetClinician(gomock.Any(), clinicID, userID).Return(nil, nil)
							})

							It("returns ok status code", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})

						Context("as service", func() {
							BeforeEach(func() {
								asService = true
							})

							It("returns ok status code", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})
					})

					Context("list user prescriptions", func() {
						var prescrs []*prescription.Prescription

						Context("for currently authenticated user", func() {
							BeforeEach(func() {
								req.URL.Path = fmt.Sprintf("/v1/patients/%v/prescriptions", userID)
								req.Method = http.MethodGet
								prescrs = []*prescription.Prescription{prescriptionTest.RandomPrescription()}
								res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
							})

							It("filters the prescriptions with the currently signed in patient user id", func() {
								prescriptionService.ListPrescriptionOutputs = []prescriptionTest.ListPrescriptionsOutput{{Prescriptions: prescrs, Err: nil}}
								handlerFunc(res, req)
								Expect(prescriptionService.ListPrescriptionsInputs).To(HaveLen(1))
								Expect(prescriptionService.ListPrescriptionsInputs[0].Filter.PatientUserID).To(Equal(userID))
								Expect(prescriptionService.ListPrescriptionsInputs[0].Filter.ClinicID).To(BeEmpty())
							})

							It("returns ok status code", func() {
								prescriptionService.ListPrescriptionOutputs = []prescriptionTest.ListPrescriptionsOutput{{Prescriptions: prescrs, Err: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})

						Context("for a different user", func() {
							BeforeEach(func() {
								req.URL.Path = fmt.Sprintf("/v1/patients/%v/prescriptions", userTest.RandomID())
								req.Method = http.MethodGet
								res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
							})

							It("returns ok status code", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})

						Context("as service request patient prescriptions", func() {
							BeforeEach(func() {
								asService = true
								req.URL.Path = fmt.Sprintf("/v1/patients/%v/prescriptions", userID)
								req.Method = http.MethodGet
								prescrs = []*prescription.Prescription{prescriptionTest.RandomPrescription()}
								res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
							})

							It("filters the prescriptions with the given patient user id", func() {
								prescriptionService.ListPrescriptionOutputs = []prescriptionTest.ListPrescriptionsOutput{{Prescriptions: prescrs, Err: nil}}
								handlerFunc(res, req)
								Expect(prescriptionService.ListPrescriptionsInputs).To(HaveLen(1))
								Expect(prescriptionService.ListPrescriptionsInputs[0].Filter.PatientUserID).To(Equal(userID))
								Expect(prescriptionService.ListPrescriptionsInputs[0].Filter.ClinicID).To(BeEmpty())
							})

							It("returns ok status code", func() {
								prescriptionService.ListPrescriptionOutputs = []prescriptionTest.ListPrescriptionsOutput{{Prescriptions: prescrs, Err: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})
					})

					Context("get patient prescription by id", func() {
						var prescr *prescription.Prescription

						BeforeEach(func() {
							prescr = prescriptionTest.RandomPrescription()
						})

						Context("for the currently authenticated user", func() {
							BeforeEach(func() {
								prescrs := []*prescription.Prescription{prescr}

								req.Method = http.MethodGet
								req.URL.Path = fmt.Sprintf("/v1/patients/%v/prescriptions/%v", userID, prescr.ID)

								prescriptionService.ListPrescriptionOutputs = []prescriptionTest.ListPrescriptionsOutput{{Prescriptions: prescrs, Err: nil}}
								res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
							})

							It("returns ok status code", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})

						Context("for a different user", func() {
							BeforeEach(func() {
								req.Method = http.MethodGet
								req.URL.Path = fmt.Sprintf("/v1/patients/%v/prescriptions/%v", userTest.RandomID(), prescr.ID)
								res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
							})

							It("returns forbidden status code", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})
					})

					Context("get clinic prescription by id", func() {
						var prescr *prescription.Prescription
						var prescrs []*prescription.Prescription

						BeforeEach(func() {
							prescr = prescriptionTest.RandomPrescription()
							prescrs = []*prescription.Prescription{prescr}

							req.Method = http.MethodGet
							req.URL.Path = fmt.Sprintf("/v1/clinics/%v/prescriptions/%v", clinicID, prescr.ID)

							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						Context("as clinician", func() {
							BeforeEach(func() {
								clinicsClient.EXPECT().GetClinician(gomock.Any(), clinicID, userID).Return(clinician, nil)
								prescriptionService.ListPrescriptionOutputs = []prescriptionTest.ListPrescriptionsOutput{{Prescriptions: prescrs, Err: nil}}
							})

							It("returns ok status code", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})

						Context("as a regular user", func() {
							BeforeEach(func() {
								clinicsClient.EXPECT().GetClinician(gomock.Any(), clinicID, userID).Return(nil, nil)
							})

							It("returns forbidden status code", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})
					})

					Context("activate prescription", func() {
						var prescr *prescription.Prescription
						var update *prescription.StateUpdate

						BeforeEach(func() {
							prescr = prescriptionTest.RandomPrescription()
							update = &prescription.StateUpdate{State: prescription.StateActive}
							body, err := json.Marshal(update)
							Expect(err).ToNot(HaveOccurred())

							req.Method = http.MethodPatch
							req.URL.Path = fmt.Sprintf("/v1/patients/%v/prescriptions/%v", userID, prescr.ID)
							req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

							prescriptionService.UpdatePrescriptionStateOutputs = []prescriptionTest.UpdatePrescriptionStateOutput{{Prescr: prescr, Err: nil}}
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						Context("as patient", func() {
							It("returns ok status code", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})

						Context("as service", func() {
							BeforeEach(func() {
								asService = true
							})

							It("returns forbidden status code", func() {
								prescriptionService.UpdatePrescriptionStateOutputs = []prescriptionTest.UpdatePrescriptionStateOutput{}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})
					})

					Context("delete prescription", func() {
						var id primitive.ObjectID

						BeforeEach(func() {
							id = prescriptionTest.RandomPrescription().ID

							req.Method = http.MethodDelete
							req.URL.Path = fmt.Sprintf("/v1/clinics/%v/prescriptions/%v", clinicID, id)

							prescriptionService.DeletePrescriptionOutputs = []prescriptionTest.DeletePrescriptionOutput{{Success: true, Err: nil}}
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						Context("as patient", func() {
							BeforeEach(func() {
								clinicsClient.EXPECT().GetClinician(gomock.Any(), clinicID, userID).Return(nil, nil)
							})

							It("returns forbidden status code", func() {
								prescriptionService.DeletePrescriptionOutputs = []prescriptionTest.DeletePrescriptionOutput{}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})

						Context("as clinician", func() {
							BeforeEach(func() {
								clinicsClient.EXPECT().GetClinician(gomock.Any(), clinicID, userID).Return(clinician, nil)
							})

							It("returns ok status code", func() {
								res.WriteOutputs = []testRest.WriteOutput{}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(0))
							})
						})

						Context("as service", func() {
							BeforeEach(func() {
								asService = true
							})

							It("returns forbidden status code", func() {
								prescriptionService.DeletePrescriptionOutputs = []prescriptionTest.DeletePrescriptionOutput{}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})
					})

					Context("add revision", func() {
						var create *prescription.RevisionCreate
						var prescr *prescription.Prescription

						BeforeEach(func() {
							create = prescriptionTest.RandomRevisionCreate()
							create.CreatedUserID = userID
							create.ClinicianID = userID
							prescriptionTest.ResetRevisionCreateHash(create)

							prescr = prescriptionTest.RandomPrescription()
							body, err := json.Marshal(create)
							Expect(err).ToNot(HaveOccurred())

							req.Method = http.MethodPost
							req.URL.Path = fmt.Sprintf("/v1/clinics/%v/prescriptions/%v/revisions", clinicID, prescr.ID.Hex())

							req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						Context("as patient", func() {
							BeforeEach(func() {
								clinicsClient.EXPECT().GetClinician(gomock.Any(), clinicID, userID).Return(nil, nil)
							})

							It("returns forbidden status code", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})

						Context("as clinician", func() {
							BeforeEach(func() {
								clinicsClient.EXPECT().GetClinician(gomock.Any(), clinicID, userID).Return(clinician, nil)
							})

							It("returns ok status code", func() {
								prescriptionService.AddRevisionOutputs = []prescriptionTest.AddRevisionOutput{{Prescr: prescr, Err: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})

						Context("as service", func() {
							BeforeEach(func() {
								asService = true
							})

							It("returns forbidden status code", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})
					})

					Context("claim prescription", func() {
						var claim *prescription.Claim
						var prescr *prescription.Prescription

						BeforeEach(func() {
							prescr = prescriptionTest.RandomPrescription()
							claim = &prescription.Claim{
								AccessCode: prescription.GenerateAccessCode(),
								Birthday:   *prescr.LatestRevision.Attributes.Birthday,
							}
							body, err := json.Marshal(claim)
							Expect(err).ToNot(HaveOccurred())

							req.Method = http.MethodPost
							req.URL.Path = fmt.Sprintf("/v1/patients/%v/prescriptions", userID)

							req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
							prescriptionService.ClaimPrescriptionOutputs = []prescriptionTest.ClaimPrescriptionOutput{}
						})

						Context("as patient", func() {
							It("returns ok status code", func() {
								prescriptionService.ClaimPrescriptionOutputs = []prescriptionTest.ClaimPrescriptionOutput{{Prescr: prescr, Err: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})

						Context("as service", func() {
							BeforeEach(func() {
								asService = true
							})

							It("returns forbidden status code", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})
					})
				})
			})
		})
	})
})
