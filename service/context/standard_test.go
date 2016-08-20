package context_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"encoding/json"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/context"
)

type Stringer struct {
	str string
}

func (s *Stringer) String() string { return s.str }

var _ = Describe("Standard", func() {
	Context("JSONResponse", func() {
		Context("encoded as JSON", func() {
			It("is an empty object if no fields are specified", func() {
				jsonResponse := &context.JSONResponse{}
				Expect(json.Marshal(jsonResponse)).To(MatchJSON(`{}`))
			})

			It("is a populated object if fields are specified", func() {
				jsonResponse := &context.JSONResponse{
					Errors: []*service.Error{
						{
							Code:   "test-code",
							Detail: "test-detail",
							Status: 400,
							Title:  "test-title",
						},
					},
					Meta: &context.Meta{
						Trace: &context.Trace{
							Request: "test-request",
							Session: "test-session",
						},
					},
				}
				Expect(json.Marshal(jsonResponse)).To(MatchJSON(`{
					"errors": [
						{
							"code": "test-code",
							"detail": "test-detail",
							"status": "400",
							"title": "test-title"
						}
					],
					"meta": {
						"trace": {
							"request": "test-request",
							"session": "test-session"
						}
					}
				}`))
			})
		})
	})

	Context("Meta", func() {
		Context("encoded as JSON", func() {
			It("is an empty object if no fields are specified", func() {
				meta := &context.Meta{}
				Expect(json.Marshal(meta)).To(MatchJSON(`{}`))
			})

			It("is a populated object if fields are specified", func() {
				meta := &context.Meta{
					Trace: &context.Trace{
						Request: "test-request",
						Session: "test-session",
					},
				}
				Expect(json.Marshal(meta)).To(MatchJSON(`{
					"trace": {
						"request": "test-request",
						"session": "test-session"
					}
				}`))
			})
		})
	})

	Context("Trace", func() {
		Context("encoded as JSON", func() {
			It("is an empty object if no fields are specified", func() {
				trace := &context.Trace{}
				Expect(json.Marshal(trace)).To(MatchJSON(`{}`))
			})

			It("is a populated object if fields are specified", func() {
				trace := &context.Trace{
					Request: "test-request",
					Session: "test-session",
				}
				Expect(json.Marshal(trace)).To(MatchJSON(`{
					"request": "test-request",
					"session": "test-session"
				}`))
			})
		})
	})

	Context("with request and response", func() {
		var request *rest.Request
		var response *TestResponseWriter

		BeforeEach(func() {
			request = NewTestRequest()
			request.Env["TRACE-REQUEST"] = "request-revenant"
			request.Env["TRACE-SESSION"] = "session-spectre"
			response = NewTestResponseWriter()
		})

		Context("NewStandard", func() {
			It("is successful", func() {
				Expect(context.NewStandard(response, request)).ToNot(BeNil())
			})

			It("returns an error if the response is missing", func() {
				standardContext, err := context.NewStandard(nil, request)
				Expect(err).To(MatchError("context: response is missing"))
				Expect(standardContext).To(BeNil())
			})

			It("returns an error if the request is missing", func() {
				standardContext, err := context.NewStandard(response, nil)
				Expect(err).To(MatchError("context: request is missing"))
				Expect(standardContext).To(BeNil())
			})
		})

		Context("with standard", func() {
			var standardContext *context.Standard

			BeforeEach(func() {
				var err error
				standardContext, err = context.NewStandard(response, request)
				Expect(err).ToNot(HaveOccurred())
				Expect(standardContext).ToNot(BeNil())
			})

			Context("Logger", func() {
				It("returns a logger", func() {
					Expect(standardContext.Logger()).ToNot(BeNil())
				})
			})

			Context("Request", func() {
				It("returns the request", func() {
					Expect(standardContext.Request()).To(Equal(request))
				})
			})

			Context("Response", func() {
				It("returns the response", func() {
					Expect(standardContext.Response()).To(Equal(response))
				})
			})

			Context("with errors", func() {
				var testErrors []*service.Error
				var internalServerFailureErrors []*service.Error

				BeforeEach(func() {
					response.WriteJSONOutputs = []error{nil}
					testErrors = []*service.Error{
						{
							Code:   "test-error-code-1",
							Status: 400,
							Title:  "test-error-title-1",
							Detail: "test-error-detail-1",
						},
						{
							Code:   "test-error-code-2",
							Status: 400,
							Title:  "test-error-title-2",
							Detail: "test-error-detail-2",
						},
					}
					internalServerFailureErrors = []*service.Error{
						service.ErrorInternalServerFailure(),
					}
				})

				Context("RespondWithError", func() {
					It("responds with internal server failure if error is missing", func() {
						standardContext.RespondWithError(nil)
						Expect(request.Env["ERRORS"]).To(Equal(internalServerFailureErrors))
						Expect(response.WriteHeaderInputs).To(ConsistOf(500))
						Expect(response.WriteJSONInputs).To(ConsistOf(&context.JSONResponse{
							Errors: internalServerFailureErrors,
							Meta: &context.Meta{
								Trace: &context.Trace{
									Request: "request-revenant",
									Session: "session-spectre",
								},
							},
						}))
					})

					It("responds with internal server failure if the status is less than zero", func() {
						testErrors[0].Status = -1
						standardContext.RespondWithError(testErrors[0])
						Expect(request.Env["ERRORS"]).To(Equal(internalServerFailureErrors))
						Expect(response.WriteHeaderInputs).To(ConsistOf(500))
						Expect(response.WriteJSONInputs).To(ConsistOf(&context.JSONResponse{
							Errors: internalServerFailureErrors,
							Meta: &context.Meta{
								Trace: &context.Trace{
									Request: "request-revenant",
									Session: "session-spectre",
								},
							},
						}))
					})

					It("responds with internal server failure if the status is zero", func() {
						testErrors[0].Status = 0
						standardContext.RespondWithError(testErrors[0])
						Expect(request.Env["ERRORS"]).To(Equal(internalServerFailureErrors))
						Expect(response.WriteHeaderInputs).To(ConsistOf(500))
						Expect(response.WriteJSONInputs).To(ConsistOf(&context.JSONResponse{
							Errors: internalServerFailureErrors,
							Meta: &context.Meta{
								Trace: &context.Trace{
									Request: "request-revenant",
									Session: "session-spectre",
								},
							},
						}))
					})

					It("responds with valid errors", func() {
						standardContext.RespondWithError(testErrors[1])
						Expect(request.Env["ERRORS"]).To(ConsistOf(testErrors[1]))
						Expect(response.WriteHeaderInputs).To(ConsistOf(400))
						Expect(response.WriteJSONInputs).To(ConsistOf(&context.JSONResponse{
							Errors: []*service.Error{testErrors[1]},
							Meta: &context.Meta{
								Trace: &context.Trace{
									Request: "request-revenant",
									Session: "session-spectre",
								},
							},
						}))
					})
				})

				Context("RespondWithInternalServerFailure", func() {
					It("is successful", func() {
						standardContext.RespondWithInternalServerFailure("test")
						Expect(request.Env["ERRORS"]).To(Equal(internalServerFailureErrors))
						Expect(response.WriteHeaderInputs).To(ConsistOf(500))
						Expect(response.WriteJSONInputs).To(ConsistOf(&context.JSONResponse{
							Errors: internalServerFailureErrors,
							Meta: &context.Meta{
								Trace: &context.Trace{
									Request: "request-revenant",
									Session: "session-spectre",
								},
							},
						}))
					})

					It("is successful with optional arguments", func() {
						standardContext.RespondWithInternalServerFailure("test",
							"string-optional", &Stringer{"stringer-optional"},
							errors.New("error-optional"), []string{"string-array-optional"})
						Expect(request.Env["ERRORS"]).To(Equal(internalServerFailureErrors))
						Expect(response.WriteHeaderInputs).To(ConsistOf(500))
						Expect(response.WriteJSONInputs).To(ConsistOf(&context.JSONResponse{
							Errors: internalServerFailureErrors,
							Meta: &context.Meta{
								Trace: &context.Trace{
									Request: "request-revenant",
									Session: "session-spectre",
								},
							},
						}))
					})
				})

				Context("RespondWithStatusAndErrors", func() {
					It("is successful", func() {
						standardContext.RespondWithStatusAndErrors(401, testErrors)
						Expect(request.Env["ERRORS"]).To(Equal(testErrors))
						Expect(response.WriteHeaderInputs).To(ConsistOf(401))
						Expect(response.WriteJSONInputs).To(ConsistOf(&context.JSONResponse{
							Errors: testErrors,
							Meta: &context.Meta{
								Trace: &context.Trace{
									Request: "request-revenant",
									Session: "session-spectre",
								},
							},
						}))
					})

					It("does not add trace request if not present", func() {
						request.Env["TRACE-REQUEST"] = nil
						standardContext.RespondWithStatusAndErrors(402, testErrors)
						Expect(request.Env["ERRORS"]).To(Equal(testErrors))
						Expect(response.WriteHeaderInputs).To(ConsistOf(402))
						Expect(response.WriteJSONInputs).To(ConsistOf(&context.JSONResponse{
							Errors: testErrors,
							Meta: &context.Meta{
								Trace: &context.Trace{
									Session: "session-spectre",
								},
							},
						}))
					})

					It("does not add trace session if not present", func() {
						request.Env["TRACE-SESSION"] = nil
						standardContext.RespondWithStatusAndErrors(403, testErrors)
						Expect(request.Env["ERRORS"]).To(Equal(testErrors))
						Expect(response.WriteHeaderInputs).To(ConsistOf(403))
						Expect(response.WriteJSONInputs).To(ConsistOf(&context.JSONResponse{
							Errors: testErrors,
							Meta: &context.Meta{
								Trace: &context.Trace{
									Request: "request-revenant",
								},
							},
						}))
					})
				})
			})

			Context("with data", func() {
				var testData interface{}

				BeforeEach(func() {
					response.WriteJSONOutputs = []error{nil}
					testData = []map[string]string{
						{
							"data-1": "value-1",
							"data-2": "value-2",
						},
						{
							"data-3": "value-3",
							"data-4": "value-4",
						},
					}
				})

				Context("RespondWithStatusAndData", func() {
					It("is successful", func() {
						standardContext.RespondWithStatusAndData(200, testData)
						Expect(response.WriteHeaderInputs).To(ConsistOf(200))
						Expect(response.WriteJSONInputs).To(ConsistOf(&context.JSONResponse{
							Data: testData,
							Meta: &context.Meta{
								Trace: &context.Trace{
									Request: "request-revenant",
									Session: "session-spectre",
								},
							},
						}))
					})

					It("does not add trace request if not present", func() {
						request.Env["TRACE-REQUEST"] = nil
						standardContext.RespondWithStatusAndData(201, testData)
						Expect(response.WriteHeaderInputs).To(ConsistOf(201))
						Expect(response.WriteJSONInputs).To(ConsistOf(&context.JSONResponse{
							Data: testData,
							Meta: &context.Meta{
								Trace: &context.Trace{
									Session: "session-spectre",
								},
							},
						}))
					})

					It("does not add trace session if not present", func() {
						request.Env["TRACE-SESSION"] = nil
						standardContext.RespondWithStatusAndData(202, testData)
						Expect(response.WriteHeaderInputs).To(ConsistOf(202))
						Expect(response.WriteJSONInputs).To(ConsistOf(&context.JSONResponse{
							Data: testData,
							Meta: &context.Meta{
								Trace: &context.Trace{
									Request: "request-revenant",
								},
							},
						}))
					})
				})
			})
		})
	})
})
