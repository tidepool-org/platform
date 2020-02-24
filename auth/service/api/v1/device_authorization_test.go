package v1_test

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

	"github.com/tidepool-org/platform/auth"
	v1 "github.com/tidepool-org/platform/auth/service/api/v1"
	serviceTest "github.com/tidepool-org/platform/auth/service/test"
	"github.com/tidepool-org/platform/auth/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/request"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	test2 "github.com/tidepool-org/platform/test"
	testRest "github.com/tidepool-org/platform/test/rest"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("DeviceAuthorization", func() {
	var svc *serviceTest.Service

	BeforeEach(func() {
		svc = serviceTest.NewService()
	})

	AfterEach(func() {
		svc.Expectations()
	})

	Context("with new router", func() {
		var router *v1.Router

		BeforeEach(func() {
			var err error
			router, err = v1.NewRouter(svc)
			Expect(err).ToNot(HaveOccurred())
			Expect(router).ToNot(BeNil())
		})

		Context("Routes", func() {
			It("returns the expected routes", func() {
				Expect(router.Routes()).To(ContainElement(
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodPost), "PathExp": Equal("/v1/users/:userId/device_authorizations")})),
				))
				Expect(router.Routes()).To(ContainElement(
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/users/:userId/device_authorizations")})),
				))
				Expect(router.Routes()).To(ContainElement(
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/users/:userId/device_authorizations/:deviceAuthorizationId")})),
				))
				Expect(router.Routes()).To(ContainElement(
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodPost), "PathExp": Equal("/v1/device_authorizations")})),
				))
			})
		})

		Context("with response and request", func() {
			var res *testRest.ResponseWriter
			var req *rest.Request
			var ctx context.Context
			var handlerFunc rest.HandlerFunc

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

			Context("with user and device authorization id", func() {
				var userID string
				var deviceAuthorizationID string

				BeforeEach(func() {
					userID = userTest.RandomID()
					deviceAuthorizationID = test.RandomDeviceAuthorizationID()
					details := request.NewDetails(request.MethodSessionToken, userID, test.NewSessionToken())
					ctx = request.NewContextWithDetails(ctx, details)
					req.Request = req.WithContext(ctx)
				})

				Context("GetDeviceAuthorization", func() {
					BeforeEach(func() {
						req.Method = http.MethodGet
						req.URL.Path = fmt.Sprintf("/v1/users/%s/device_authorizations/%s", userID, deviceAuthorizationID)
					})

					Context("responds with JSON", func() {
						BeforeEach(func() {
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						AfterEach(func() {
							Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
						})

						When("the path does not contain a user id", func() {
							BeforeEach(func() {
								req.URL.Path = fmt.Sprintf("/v1/users//device_authorizations/%s", deviceAuthorizationID)
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterMissing("userId"), res.WriteInputs[0])
							})
						})

						When("requesting device authorizations owned by different user", func() {
							BeforeEach(func() {
								req.URL.Path = fmt.Sprintf("/v1/users/%s/device_authorizations/%s", userTest.RandomID(), deviceAuthorizationID)
							})

							It("responds with unauthorized and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusUnauthorized}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
							})
						})

						When("the device authorization exists", func() {
							var deviceAuthorization *auth.DeviceAuthorization

							BeforeEach(func() {
								deviceAuthorization = test.RandomDeviceAuthorization()
								output := test.GetUserDeviceAuthorizationOutput{
									Authorization: deviceAuthorization,
									Err:           nil,
								}
								svc.AuthClientImpl.GetUserDeviceAuthorizationOutputs = []test.GetUserDeviceAuthorizationOutput{
									output,
								}
							})

							It("responds with ok status and the device authorization in the body", func() {
								handlerFunc(res, req)
								Expect(svc.AuthClientImpl.GetUserDeviceAuthorizationInputs).To(Not(BeEmpty()))
								Expect(svc.AuthClientImpl.GetUserDeviceAuthorizationInputs[0].UserID).To(Equal(userID))
								Expect(svc.AuthClientImpl.GetUserDeviceAuthorizationInputs[0].ID).To(Equal(deviceAuthorizationID))
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(1))
								Expect(json.Marshal(deviceAuthorization)).To(MatchJSON(res.WriteInputs[0]))
							})
						})

						When("the device authorization doesn't exists", func() {
							BeforeEach(func() {
								output := test.GetUserDeviceAuthorizationOutput{Authorization: nil, Err: nil}
								svc.AuthClientImpl.GetUserDeviceAuthorizationOutputs = []test.GetUserDeviceAuthorizationOutput{
									output,
								}
							})

							It("responds with not found status and the expected error", func() {
								handlerFunc(res, req)
								Expect(svc.AuthClientImpl.GetUserDeviceAuthorizationInputs).To(Not(BeEmpty()))
								Expect(svc.AuthClientImpl.GetUserDeviceAuthorizationInputs[0].UserID).To(Equal(userID))
								Expect(svc.AuthClientImpl.GetUserDeviceAuthorizationInputs[0].ID).To(Equal(deviceAuthorizationID))
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusNotFound}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorResourceNotFound(), res.WriteInputs[0])
							})
						})
					})
				})

				Context("ListDeviceAuthorizations", func() {
					BeforeEach(func() {
						req.Method = http.MethodGet
						req.URL.Path = fmt.Sprintf("/v1/users/%s/device_authorizations", userID)
					})

					Context("responds with JSON", func() {
						BeforeEach(func() {
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						AfterEach(func() {
							Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
						})

						When("the path does not contain a user id", func() {
							BeforeEach(func() {
								req.URL.Path = fmt.Sprintf("/v1/users//device_authorizations")
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterMissing("userId"), res.WriteInputs[0])
							})
						})

						When("requesting device authorizations owned by different user", func() {
							BeforeEach(func() {
								req.URL.Path = fmt.Sprintf("/v1/users/%s/device_authorizations", userTest.RandomID())
							})

							It("responds with unauthorized and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusUnauthorized}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
							})
						})

						When("there are one or more device authorizations", func() {
							var deviceAuthorizations auth.DeviceAuthorizations

							BeforeEach(func() {
								count := test2.RandomIntFromRange(1, 10)
								deviceAuthorizations = make(auth.DeviceAuthorizations, count)
								for i := 0; i < count; i++ {
									deviceAuthorizations = append(deviceAuthorizations, test.RandomDeviceAuthorization())
								}
								output := test.ListUserDeviceAuthorizationsOutput{
									Authorizations: deviceAuthorizations,
									Err:            nil,
								}
								svc.AuthClientImpl.ListUserDeviceAuthorizationsOutputs = []test.ListUserDeviceAuthorizationsOutput{
									output,
								}
							})

							It("responds with ok status and the device authorization in the body", func() {
								handlerFunc(res, req)
								Expect(svc.AuthClientImpl.ListUserDeviceAuthorizationsInputs).To(Not(BeEmpty()))
								Expect(svc.AuthClientImpl.ListUserDeviceAuthorizationsInputs[0].UserID).To(Equal(userID))
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(1))
								Expect(json.Marshal(deviceAuthorizations)).To(MatchJSON(res.WriteInputs[0]))
							})
						})

						When("no device authorizations exist", func() {
							authorizations := make(auth.DeviceAuthorizations, 0)

							BeforeEach(func() {
								output := test.ListUserDeviceAuthorizationsOutput{Authorizations: authorizations, Err: nil}
								svc.AuthClientImpl.ListUserDeviceAuthorizationsOutputs = []test.ListUserDeviceAuthorizationsOutput{
									output,
								}
							})

							It("responds with an empty array", func() {
								handlerFunc(res, req)
								Expect(svc.AuthClientImpl.ListUserDeviceAuthorizationsInputs).To(Not(BeEmpty()))
								Expect(svc.AuthClientImpl.ListUserDeviceAuthorizationsInputs[0].UserID).To(Equal(userID))
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(1))
								Expect(json.Marshal(authorizations)).To(MatchJSON(res.WriteInputs[0]))
							})
						})
					})
				})

				Context("CreateDeviceAuthorization", func() {
					BeforeEach(func() {
						req.Method = http.MethodPost
						req.URL.Path = fmt.Sprintf("/v1/users/%s/device_authorizations", userID)
					})

					Context("responds with JSON", func() {
						BeforeEach(func() {
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						AfterEach(func() {
							Expect(*res.HeaderOutput).To(HaveKey("Content-Type"))
							Expect((*res.HeaderOutput)["Content-Type"]).To(Equal([]string{"application/json; charset=utf-8"}))
						})

						When("the path does not contain a user id", func() {
							BeforeEach(func() {
								req.URL.Path = fmt.Sprintf("/v1/users//device_authorizations")
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterMissing("userId"), res.WriteInputs[0])
							})
						})

						When("creating a device authorization for a different user", func() {
							BeforeEach(func() {
								req.URL.Path = fmt.Sprintf("/v1/users/%s/device_authorizations", userTest.RandomID())
							})

							It("responds with unauthorized and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusUnauthorized}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
							})
						})

						When("creating a new device authorization", func() {
							var create *auth.DeviceAuthorizationCreate
							var deviceAuthorization *auth.DeviceAuthorization

							BeforeEach(func() {
								create = test.RandomDeviceAuthorizationCreate()
								deviceAuthorization = test.RandomDeviceAuthorization()
								deviceAuthorization.UserID = userID
								deviceAuthorization.Status = auth.DeviceAuthorizationPending
								deviceAuthorization.VerificationCode = ""
								deviceAuthorization.DevicePushToken = create.DevicePushToken

								output := test.CreateUserDeviceAuthorizationOutput{Authorization: deviceAuthorization, Err: nil}
								svc.AuthClientImpl.CreateUserDeviceAuthorizationOutputs = []test.CreateUserDeviceAuthorizationOutput{
									output,
								}

								req.Body = ioutil.NopCloser(bytes.NewBuffer(test2.MarshalRequestBody(create)))
							})

							AfterEach(func() {
								Expect(svc.AuthClientImpl.CreateUserDeviceAuthorizationInputs).To(Not(BeEmpty()))
								Expect(svc.AuthClientImpl.CreateUserDeviceAuthorizationInputs[0].UserID).To(Equal(userID))
								Expect(svc.AuthClientImpl.CreateUserDeviceAuthorizationInputs[0].Create).To(Equal(create))
							})

							It("responds with status created", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusCreated}))
							})

							It("responds with the device authorization in the response body", func() {
								handlerFunc(res, req)
								Expect(res.WriteInputs).To(HaveLen(1))
								Expect(json.Marshal(deviceAuthorization)).To(MatchJSON(res.WriteInputs[0]))
							})

							It("responds with a device authorization token in the header", func() {
								handlerFunc(res, req)
								Expect(*res.HeaderOutput).To(HaveKey("X-Tidepool-Bearer-Token"))
								Expect((*res.HeaderOutput)["X-Tidepool-Bearer-Token"]).To(Equal([]string{deviceAuthorization.Token}))
							})
						})
					})
				})

				Context("UpdateDeviceAuthorization", func() {
					BeforeEach(func() {
						req.Method = http.MethodPost
						req.URL.Path = "/v1/device_authorizations"
					})

					Context("responds with JSON", func() {
						BeforeEach(func() {
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						AfterEach(func() {
							Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
						})

						When("update a device authorization with a valid body", func() {
							var update *auth.DeviceAuthorizationUpdate
							var deviceAuthorization *auth.DeviceAuthorization
							var token string

							BeforeEach(func() {
								update = test.RandomDeviceAuthorizationUpdate()
								deviceAuthorization = test.RandomDeviceAuthorization()
								deviceAuthorization.Status = auth.DeviceAuthorizationSuccessful
								token = deviceAuthorization.Token

								output := test.GetDeviceAuthorizationByTokenOutput{
									Authorization: deviceAuthorization,
									Err:           nil,
								}
								svc.AuthClientImpl.GetDeviceAuthorizationByTokenOutputs = []test.GetDeviceAuthorizationByTokenOutput{
									output,
								}

								updateOutput := test.UpdateDeviceAuthorizationOutput{
									Authorization: deviceAuthorization,
									Err:           nil,
								}
								svc.AuthClientImpl.UpdateDeviceAuthorizationOutputs = []test.UpdateDeviceAuthorizationOutput{
									updateOutput,
								}

								req.Body = ioutil.NopCloser(bytes.NewBuffer(test2.MarshalRequestBody(update)))
								req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
							})

							AfterEach(func() {
								Expect(svc.AuthClientImpl.GetDeviceAuthorizationByTokenInputs).To(Not(BeEmpty()))
								Expect(svc.AuthClientImpl.GetDeviceAuthorizationByTokenInputs[0].Token).To(Equal(token))
								Expect(svc.AuthClientImpl.UpdateDeviceAuthorizationInputs).To(Not(BeEmpty()))
								Expect(svc.AuthClientImpl.UpdateDeviceAuthorizationInputs[0].ID).To(Equal(deviceAuthorization.ID))
								Expect(svc.AuthClientImpl.UpdateDeviceAuthorizationInputs[0].Update).To(Equal(update))
							})

							It("responds with status ok and the updated device authorization", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(1))
								Expect(json.Marshal(deviceAuthorization)).To(MatchJSON(res.WriteInputs[0]))
							})
						})

						When("when a device authorization with the provided token doesn't exist", func() {
							var token string

							BeforeEach(func() {
								token = test.RandomDeviceAuthorizationToken()

								output := test.GetDeviceAuthorizationByTokenOutput{
									Authorization: nil,
									Err:           nil,
								}
								svc.AuthClientImpl.GetDeviceAuthorizationByTokenOutputs = []test.GetDeviceAuthorizationByTokenOutput{
									output,
								}

								req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
							})

							AfterEach(func() {
								Expect(svc.AuthClientImpl.GetDeviceAuthorizationByTokenInputs).To(Not(BeEmpty()))
								Expect(svc.AuthClientImpl.GetDeviceAuthorizationByTokenInputs[0].Token).To(Equal(token))
								Expect(svc.AuthClientImpl.UpdateDeviceAuthorizationInputs).To(BeEmpty())
							})

							It("responds with unauthorized error", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusUnauthorized}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
							})
						})

						When("the body has an empty (invalid) verification code", func() {
							var update *auth.DeviceAuthorizationUpdate
							var deviceAuthorization *auth.DeviceAuthorization
							var token string

							BeforeEach(func() {
								update = test.RandomDeviceAuthorizationUpdate()
								update.Status = auth.DeviceAuthorizationFailed
								update.VerificationCode = ""

								deviceAuthorization = test.RandomDeviceAuthorization()
								deviceAuthorization.Status = auth.DeviceAuthorizationFailed
								token = deviceAuthorization.Token

								output := test.GetDeviceAuthorizationByTokenOutput{
									Authorization: deviceAuthorization,
									Err:           nil,
								}
								svc.AuthClientImpl.GetDeviceAuthorizationByTokenOutputs = []test.GetDeviceAuthorizationByTokenOutput{
									output,
								}

								req.Body = ioutil.NopCloser(bytes.NewBuffer(test2.MarshalRequestBody(update)))
								req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
							})

							AfterEach(func() {
								Expect(svc.AuthClientImpl.GetDeviceAuthorizationByTokenInputs).To(Not(BeEmpty()))
								Expect(svc.AuthClientImpl.GetDeviceAuthorizationByTokenInputs[0].Token).To(Equal(token))
								Expect(svc.AuthClientImpl.UpdateDeviceAuthorizationInputs).To(BeEmpty())
							})

							It("responds with status ok and the updated device authorization", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(
									errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/verificationCode"),
									res.WriteInputs[0],
								)
							})
						})
					})
				})
			})
		})
	})
})
