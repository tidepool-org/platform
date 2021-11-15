package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/mdblp/go-json-rest/rest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	requestTest "github.com/tidepool-org/platform/request/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	testRest "github.com/tidepool-org/platform/test/rest"
	"github.com/tidepool-org/platform/user"
	userServiceApiV1 "github.com/tidepool-org/platform/user/service/api/v1"
	userServiceApiV1Test "github.com/tidepool-org/platform/user/service/api/v1/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("V1", func() {
	var provider *userServiceApiV1Test.Provider

	BeforeEach(func() {
		provider = userServiceApiV1Test.NewProvider()
	})

	AfterEach(func() {
		provider.AssertOutputsEmpty()
	})

	Context("NewRouter", func() {
		It("returns an error when the provider is missing", func() {
			router, err := userServiceApiV1.NewRouter(nil)
			errorsTest.ExpectEqual(err, errors.New("provider is missing"))
			Expect(router).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(userServiceApiV1.NewRouter(provider)).ToNot(BeNil())
		})
	})

	Context("with new router", func() {
		var router *userServiceApiV1.Router

		BeforeEach(func() {
			var err error
			router, err = userServiceApiV1.NewRouter(provider)
			Expect(err).ToNot(HaveOccurred())
			Expect(router).ToNot(BeNil())
		})

		Context("Routes", func() {
			It("returns the expected routes", func() {
				Expect(router.Routes()).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/users/:id")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodDelete), "PathExp": Equal("/v1/users/:id")})),
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

			Context("with id", func() {
				var id string

				BeforeEach(func() {
					id = userTest.RandomID()
				})

				Context("Get", func() {
					BeforeEach(func() {
						req.Method = http.MethodGet
						req.URL.Path = fmt.Sprintf("/v1/users/%s", id)
					})

					It("panics when the response is missing", func() {
						Expect(func() { router.Get(nil, req) }).To(Panic())
					})

					It("panics when the request is missing", func() {
						Expect(func() { router.Get(res, nil) }).To(Panic())
					})

					Context("responds with JSON", func() {
						BeforeEach(func() {
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						AfterEach(func() {
							Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
						})

						When("the path contains an invalid id", func() {
							BeforeEach(func() {
								req.URL.Path = "/v1/users/invalid"
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterInvalid("id"), res.WriteInputs[0])
							})
						})

						Context("with client", func() {
							var client *userTest.Client

							BeforeEach(func() {
								client = userTest.NewClient()
								provider.UserClientOutputs = []user.Client{client}
							})

							AfterEach(func() {
								Expect(client.GetInputs).To(Equal([]string{id}))
								client.AssertOutputsEmpty()
							})

							It("responds with an unauthorized error when the client returns an unauthorized error", func() {
								client.GetOutputs = []userTest.GetOutput{{User: nil, Error: request.ErrorUnauthorized()}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
							})

							It("responds with an internal server error when the client returns an unknown error", func() {
								client.GetOutputs = []userTest.GetOutput{{User: nil, Error: errorsTest.RandomError()}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
							})

							It("responds with not found error when the client does not return an user", func() {
								client.GetOutputs = []userTest.GetOutput{{User: nil, Error: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusNotFound}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorResourceNotFoundWithID(id), res.WriteInputs[0])
							})

							It("responds successfully", func() {
								responseResult := userTest.RandomUser()
								client.GetOutputs = []userTest.GetOutput{{User: responseResult, Error: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(1))
								Expect(json.Marshal(responseResult)).To(MatchJSON(res.WriteInputs[0]))
							})
						})
					})
				})

				Context("Delete", func() {
					var revision *int

					BeforeEach(func() {
						revision = pointer.FromInt(requestTest.RandomRevision())
						req.Method = http.MethodDelete
						req.URL.Path = fmt.Sprintf("/v1/users/%s", id)
					})

					JustBeforeEach(func() {
						if revision != nil {
							query := url.Values{"revision": []string{strconv.Itoa(*revision)}}
							req.URL.RawQuery = query.Encode()
						}
					})

					It("panics when the response is missing", func() {
						Expect(func() { router.Delete(nil, req) }).To(Panic())
					})

					It("panics when the request is missing", func() {
						Expect(func() { router.Delete(res, nil) }).To(Panic())
					})

					When("responds", func() {
						BeforeEach(func() {
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						When("the path contains an invalid id", func() {
							BeforeEach(func() {
								req.URL.Path = "/v1/users/invalid"
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterInvalid("id"), res.WriteInputs[0])
							})
						})

						When("the query contains an invalid revision", func() {
							BeforeEach(func() {
								revision = pointer.FromInt(-1)
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(errorsTest.WithParameterSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "revision"), res.WriteInputs[0])
							})
						})

						Context("with delete", func() {
							var deleet *user.Delete

							BeforeEach(func() {
								deleet = userTest.RandomDelete()
							})

							When("the body contains an invalid delete", func() {
								BeforeEach(func() {
									deleet.Password = pointer.FromString("")
									req.Body = ioutil.NopCloser(bytes.NewBuffer(test.MarshalRequestBody(deleet)))
								})

								It("responds with an bad request error", func() {
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/password"), res.WriteInputs[0])
								})
							})

							Context("with client", func() {
								var client *userTest.Client

								BeforeEach(func() {
									client = userTest.NewClient()
									provider.UserClientOutputs = []user.Client{client}
								})

								AfterEach(func() {
									client.AssertOutputsEmpty()
								})

								revisionAssertions := func() {
									deleteAssertions := func() {
										Context("deletes", func() {
											AfterEach(func() {
												Expect(client.DeleteInputs).To(Equal([]userTest.DeleteInput{{ID: id, Delete: deleet, Condition: &request.Condition{Revision: revision}}}))
											})

											It("responds with an unauthorized error when the client returns an unauthorized error", func() {
												client.DeleteOutputs = []userTest.DeleteOutput{{Deleted: false, Error: request.ErrorUnauthorized()}}
												handlerFunc(res, req)
												Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
												Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
												Expect(res.WriteInputs).To(HaveLen(1))
												errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
											})

											It("responds with an internal server error when the client returns an unknown error", func() {
												client.DeleteOutputs = []userTest.DeleteOutput{{Deleted: false, Error: errorsTest.RandomError()}}
												handlerFunc(res, req)
												Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
												Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
												Expect(res.WriteInputs).To(HaveLen(1))
												errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
											})

											It("responds with not found error when the client does not return an user", func() {
												client.DeleteOutputs = []userTest.DeleteOutput{{Deleted: false, Error: nil}}
												handlerFunc(res, req)
												Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusNotFound}))
												Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
												Expect(res.WriteInputs).To(HaveLen(1))
												errorsTest.ExpectErrorJSON(request.ErrorResourceNotFoundWithIDAndOptionalRevision(id, revision), res.WriteInputs[0])
											})

											It("responds successfully", func() {
												res.WriteOutputs = nil
												client.DeleteOutputs = []userTest.DeleteOutput{{Deleted: true, Error: nil}}
												handlerFunc(res, req)
												Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusNoContent}))
												Expect(res.HeaderOutput).To(Equal(&http.Header{}))
											})
										})
									}

									When("delete is missing", func() {
										BeforeEach(func() {
											deleet = nil
										})

										deleteAssertions()
									})

									When("delete password is missing", func() {
										BeforeEach(func() {
											deleet.Password = nil
										})

										JustBeforeEach(func() {
											req.Body = ioutil.NopCloser(bytes.NewBuffer(test.MarshalRequestBody(deleet)))
										})

										deleteAssertions()
									})

									When("delete password is present", func() {
										BeforeEach(func() {
											deleet.Password = nil
										})

										JustBeforeEach(func() {
											req.Body = ioutil.NopCloser(bytes.NewBuffer(test.MarshalRequestBody(deleet)))
										})

										deleteAssertions()
									})
								}

								When("revision is missing", func() {
									BeforeEach(func() {
										revision = nil
									})

									revisionAssertions()
								})

								When("revision is present", func() {
									revisionAssertions()
								})
							})
						})
					})
				})
			})
		})
	})
})
