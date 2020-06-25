package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"go.mongodb.org/mongo-driver/bson/primitive"

	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/prescription"
	"github.com/tidepool-org/platform/prescription/api"
	prescriptionTest "github.com/tidepool-org/platform/prescription/test"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/user"

	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	testRest "github.com/tidepool-org/platform/test/rest"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("V1", func() {
	var userClient *userTest.Client
	var prescriptionService *prescriptionTest.PrescriptionAccessor

	BeforeEach(func() {
		prescriptionService = prescriptionTest.NewPrescriptionAccessor()
		userClient = userTest.NewClient()
	})

	AfterEach(func() {
		prescriptionService.Expectations()
		userClient.AssertOutputsEmpty()
	})

	Context("NewRouter", func() {
		It("returns successfully", func() {
			Expect(api.NewRouter(api.Params{
				PrescriptionService: prescriptionService,
				UserClient:          userClient,
			})).ToNot(BeNil())
		})
	})

	Context("with new router", func() {
		var router service.Router

		BeforeEach(func() {
			router = api.NewRouter(api.Params{
				PrescriptionService: prescriptionService,
				UserClient:          userClient,
			})
			Expect(router).ToNot(BeNil())
		})

		Context("Routes", func() {
			It("returns the expected routes", func() {
				Expect(router.Routes()).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodPost), "PathExp": Equal("/v1/prescriptions")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/prescriptions")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/users/:userId/prescriptions")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodPost), "PathExp": Equal("/v1/prescriptions/claim")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/prescriptions/:prescriptionId")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodPatch), "PathExp": Equal("/v1/prescriptions/:prescriptionId")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodDelete), "PathExp": Equal("/v1/prescriptions/:prescriptionId")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodPost), "PathExp": Equal("/v1/prescriptions/:prescriptionId/revisions")})),
				))
			})
		})

		Context("with response and request", func() {
			var res *testRest.ResponseWriter
			var req *rest.Request
			var ctx context.Context
			var handlerFunc rest.HandlerFunc
			var details request.Details

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
				var patient *user.User
				var clinician *user.User

				BeforeEach(func() {
					patient = userTest.RandomUser()
					clinician = userTest.RandomUser()

					clinicianRoles := []string{user.RoleClinic}
					clinician.Roles = &clinicianRoles
				})

				When("signed in", func() {
					var currentUser *user.User
					asService := false

					JustBeforeEach(func() {
						if asService {
							details = request.NewDetails(request.MethodServiceSecret, "", authTest.NewServiceSecret())
						} else {
							details = request.NewDetails(request.MethodSessionToken, *currentUser.UserID, "")
						}
						if currentUser != nil {
							userClient.GetOutputs = []userTest.GetOutput{{User: currentUser, Error: nil}}
						}
						req.Request = req.WithContext(request.NewContextWithDetails(req.Context(), details))
					})

					JustAfterEach(func() {
						currentUser = nil
						asService = false
					})

					Context("create prescription", func() {
						var create *prescription.RevisionCreate
						var prescr *prescription.Prescription

						BeforeEach(func() {
							req.Method = http.MethodPost
							req.URL.Path = "/v1/prescriptions"

							create = prescriptionTest.RandomRevisionCreate()
							prescr = prescriptionTest.RandomPrescription()
							body, err := json.Marshal(create)
							Expect(err).ToNot(HaveOccurred())

							req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						Context("as patient", func() {
							BeforeEach(func() {
								currentUser = patient
							})

							It("returns forbidden status code", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})

						Context("as clinician", func() {
							BeforeEach(func() {
								currentUser = clinician
							})

							It("returns created status code", func() {
								prescriptionService.CreatePrescriptionOutputs = []prescriptionTest.CreatePrescriptionOutput{{Prescription: prescr, Error: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusCreated}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})

						Context("as service", func() {
							BeforeEach(func() {
								asService = true
							})

							It("returns unauthorized status code", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusUnauthorized}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})
					})

					Context("list current user prescriptions", func() {
						var prescrs []*prescription.Prescription

						BeforeEach(func() {
							req.Method = http.MethodGet
							req.URL.Path = "/v1/prescriptions"

							prescrs = []*prescription.Prescription{prescriptionTest.RandomPrescription()}
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						Context("as patient", func() {
							BeforeEach(func() {
								currentUser = patient
							})

							It("returns ok status code", func() {
								prescriptionService.ListPrescriptionOutputs = []prescriptionTest.ListPrescriptionsOutput{{Prescriptions: prescrs, Err: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})

						Context("as clinician", func() {
							BeforeEach(func() {
								currentUser = clinician
							})

							It("returns ok status code", func() {
								prescriptionService.ListPrescriptionOutputs = []prescriptionTest.ListPrescriptionsOutput{{Prescriptions: prescrs, Err: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})

						Context("as service", func() {
							BeforeEach(func() {
								asService = true
							})

							It("returns unauthorized status code", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusUnauthorized}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})
					})

					Context("list user prescriptions", func() {
						var prescrs []*prescription.Prescription

						BeforeEach(func() {
							req.Method = http.MethodGet
							prescrs = []*prescription.Prescription{prescriptionTest.RandomPrescription()}
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						JustBeforeEach(func() {
							req.URL.Path = fmt.Sprintf("/v1/users/%v/prescriptions", *currentUser.UserID)
						})

						Context("as patient", func() {
							BeforeEach(func() {
								currentUser = patient
							})

							It("returns ok status code", func() {
								prescriptionService.ListPrescriptionOutputs = []prescriptionTest.ListPrescriptionsOutput{{Prescriptions: prescrs, Err: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})

						Context("as clinician", func() {
							BeforeEach(func() {
								currentUser = clinician
							})

							It("returns ok status code", func() {
								prescriptionService.ListPrescriptionOutputs = []prescriptionTest.ListPrescriptionsOutput{{Prescriptions: prescrs, Err: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})

						Context("as service request patient prescriptions", func() {
							BeforeEach(func() {
								currentUser = patient
								asService = true
							})

							It("returns ok status code", func() {
								prescriptionService.ListPrescriptionOutputs = []prescriptionTest.ListPrescriptionsOutput{{Prescriptions: prescrs, Err: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})

						Context("as service requesting clinician prescriptions", func() {
							BeforeEach(func() {
								currentUser = clinician
								asService = true
							})

							It("returns ok status code", func() {
								prescriptionService.ListPrescriptionOutputs = []prescriptionTest.ListPrescriptionsOutput{{Prescriptions: prescrs, Err: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})
					})

					Context("get prescription by id", func() {
						var prescr *prescription.Prescription

						BeforeEach(func() {
							prescr = prescriptionTest.RandomPrescription()
							prescrs := []*prescription.Prescription{prescr}

							req.Method = http.MethodGet
							req.URL.Path = fmt.Sprintf("/v1/prescriptions/%v", prescr.ID)

							prescriptionService.ListPrescriptionOutputs = []prescriptionTest.ListPrescriptionsOutput{{Prescriptions: prescrs, Err: nil}}
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						Context("as patient", func() {
							BeforeEach(func() {
								currentUser = patient
							})

							It("returns ok status code", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})

						Context("as clinician", func() {
							BeforeEach(func() {
								currentUser = clinician
							})

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

							It("returns unauthorized code", func() {
								prescriptionService.ListPrescriptionOutputs = []prescriptionTest.ListPrescriptionsOutput{}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusUnauthorized}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})
					})

					Context("update prescription", func() {
						var prescr *prescription.Prescription
						var update *prescription.StateUpdate

						BeforeEach(func() {
							prescr = prescriptionTest.RandomPrescription()
							update = &prescription.StateUpdate{State: prescription.StateActive}
							body, err := json.Marshal(update)
							Expect(err).ToNot(HaveOccurred())

							req.Method = http.MethodPatch
							req.URL.Path = fmt.Sprintf("/v1/prescriptions/%v", prescr.ID)
							req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

							prescriptionService.UpdatePrescriptionStateOutputs = []prescriptionTest.UpdatePrescriptionStateOutput{{Prescr: prescr, Err: nil}}
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						Context("as patient", func() {
							BeforeEach(func() {
								currentUser = patient
							})

							It("returns ok status code", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})

						Context("as clinician", func() {
							BeforeEach(func() {
								currentUser = clinician
							})

							It("returns unauthorized status code", func() {
								prescriptionService.UpdatePrescriptionStateOutputs = []prescriptionTest.UpdatePrescriptionStateOutput{}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusUnauthorized}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})

						Context("as service", func() {
							BeforeEach(func() {
								asService = true
							})

							It("returns unauthorized status code", func() {
								prescriptionService.UpdatePrescriptionStateOutputs = []prescriptionTest.UpdatePrescriptionStateOutput{}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusUnauthorized}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})
					})

					Context("delete prescription", func() {
						var id primitive.ObjectID

						BeforeEach(func() {
							id = prescriptionTest.RandomPrescription().ID

							req.Method = http.MethodDelete
							req.URL.Path = fmt.Sprintf("/v1/prescriptions/%v", id)

							prescriptionService.DeletePrescriptionOutputs = []prescriptionTest.DeletePrescriptionOutput{{Success: true, Err: nil}}
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						Context("as patient", func() {
							BeforeEach(func() {
								currentUser = patient
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
								currentUser = clinician
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

							It("returns unauthorized status code", func() {
								prescriptionService.DeletePrescriptionOutputs = []prescriptionTest.DeletePrescriptionOutput{}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusUnauthorized}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})
					})

					Context("add revision", func() {
						var create *prescription.RevisionCreate
						var prescr *prescription.Prescription

						BeforeEach(func() {
							create = prescriptionTest.RandomRevisionCreate()
							prescr = prescriptionTest.RandomPrescription()
							body, err := json.Marshal(create)
							Expect(err).ToNot(HaveOccurred())

							req.Method = http.MethodPost
							req.URL.Path = fmt.Sprintf("/v1/prescriptions/%v/revisions", prescr.ID.Hex())

							req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						Context("as patient", func() {
							BeforeEach(func() {
								currentUser = patient
							})

							It("returns forbidden status code", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})

						Context("as clinician", func() {
							BeforeEach(func() {
								currentUser = clinician
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

							It("returns unauthorized status code", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusUnauthorized}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})
					})

					Context("claim prescription", func() {
						var claim *prescription.Claim
						var prescr *prescription.Prescription

						BeforeEach(func() {
							claim = &prescription.Claim{AccessCode: prescription.GenerateAccessCode()}
							prescr = prescriptionTest.RandomPrescription()
							body, err := json.Marshal(claim)
							Expect(err).ToNot(HaveOccurred())

							req.Method = http.MethodPost
							req.URL.Path = "/v1/prescriptions/claim"

							req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
							prescriptionService.ClaimPrescriptionOutputs = []prescriptionTest.ClaimPrescriptionOutput{}
						})

						Context("as patient", func() {
							BeforeEach(func() {
								currentUser = patient
							})

							It("returns ok status code", func() {
								prescriptionService.ClaimPrescriptionOutputs = []prescriptionTest.ClaimPrescriptionOutput{{Prescr: prescr, Err: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})

						Context("as clinician", func() {
							BeforeEach(func() {
								currentUser = clinician
							})

							It("returns forbidden status code", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})

						Context("as service", func() {
							BeforeEach(func() {
								asService = true
							})

							It("returns unauthorized status code", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusUnauthorized}))
								Expect(res.WriteInputs).To(HaveLen(1))
							})
						})
					})
				})
			})
		})
	})
})
