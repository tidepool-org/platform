package context_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"encoding/json"
	"errors"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/context"
	testRest "github.com/tidepool-org/platform/test/rest"
)

type Stringer struct {
	str string
}

func (s *Stringer) String() string { return s.str }

var _ = Describe("Responder", func() {
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
		var response *testRest.ResponseWriter

		BeforeEach(func() {
			request = testRest.NewRequest()
			request.Env["TRACE-REQUEST"] = "request-revenant"
			request.Env["TRACE-SESSION"] = "session-spectre"
			response = testRest.NewResponseWriter()
		})

		Context("NewResponder", func() {
			It("returns an error if the response is missing", func() {
				responder, err := context.NewResponder(nil, request)
				Expect(err).To(MatchError("response is missing"))
				Expect(responder).To(BeNil())
			})

			It("returns an error if the request is missing", func() {
				responder, err := context.NewResponder(response, nil)
				Expect(err).To(MatchError("request is missing"))
				Expect(responder).To(BeNil())
			})

			It("is successful", func() {
				Expect(context.NewResponder(response, request)).ToNot(BeNil())
			})
		})

		Context("with new responder", func() {
			var responder *context.Responder

			BeforeEach(func() {
				var err error
				responder, err = context.NewResponder(response, request)
				Expect(err).ToNot(HaveOccurred())
				Expect(responder).ToNot(BeNil())
			})

			Context("Response", func() {
				It("returns the response", func() {
					Expect(responder.Request()).To(Equal(request))
				})
			})

			Context("Request", func() {
				It("returns the request", func() {
					Expect(responder.Request()).To(Equal(request))
				})
			})

			Context("Logger", func() {
				It("returns a logger", func() {
					Expect(responder.Logger()).ToNot(BeNil())
				})
			})

			Context("with errors", func() {
				var testErrors []*service.Error
				var internalServerFailureErrors []*service.Error

				BeforeEach(func() {
					response.WriteJsonOutputs = []error{nil}
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
						responder.RespondWithError(nil)
						Expect(request.Env["ERRORS"]).To(Equal(internalServerFailureErrors))
						Expect(response.WriteHeaderInputs).To(ConsistOf(500))
						Expect(response.WriteJsonInputs).To(ConsistOf(&context.JSONResponse{
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
						responder.RespondWithError(testErrors[0])
						Expect(request.Env["ERRORS"]).To(Equal(internalServerFailureErrors))
						Expect(response.WriteHeaderInputs).To(ConsistOf(500))
						Expect(response.WriteJsonInputs).To(ConsistOf(&context.JSONResponse{
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
						responder.RespondWithError(testErrors[0])
						Expect(request.Env["ERRORS"]).To(Equal(internalServerFailureErrors))
						Expect(response.WriteHeaderInputs).To(ConsistOf(500))
						Expect(response.WriteJsonInputs).To(ConsistOf(&context.JSONResponse{
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
						responder.RespondWithError(testErrors[1])
						Expect(request.Env["ERRORS"]).To(ConsistOf(testErrors[1]))
						Expect(response.WriteHeaderInputs).To(ConsistOf(400))
						Expect(response.WriteJsonInputs).To(ConsistOf(&context.JSONResponse{
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
						responder.RespondWithInternalServerFailure("test")
						Expect(request.Env["ERRORS"]).To(Equal(internalServerFailureErrors))
						Expect(response.WriteHeaderInputs).To(ConsistOf(500))
						Expect(response.WriteJsonInputs).To(ConsistOf(&context.JSONResponse{
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
						responder.RespondWithInternalServerFailure("test",
							"string-optional", &Stringer{"stringer-optional"},
							errors.New("error-optional"), []string{"string-array-optional"})
						Expect(request.Env["ERRORS"]).To(Equal(internalServerFailureErrors))
						Expect(response.WriteHeaderInputs).To(ConsistOf(500))
						Expect(response.WriteJsonInputs).To(ConsistOf(&context.JSONResponse{
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
						responder.RespondWithStatusAndErrors(401, testErrors)
						Expect(request.Env["ERRORS"]).To(Equal(testErrors))
						Expect(response.WriteHeaderInputs).To(ConsistOf(401))
						Expect(response.WriteJsonInputs).To(ConsistOf(&context.JSONResponse{
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
						delete(request.Env, "TRACE-REQUEST")
						responder.RespondWithStatusAndErrors(402, testErrors)
						Expect(request.Env["ERRORS"]).To(Equal(testErrors))
						Expect(response.WriteHeaderInputs).To(ConsistOf(402))
						Expect(response.WriteJsonInputs).To(ConsistOf(&context.JSONResponse{
							Errors: testErrors,
							Meta: &context.Meta{
								Trace: &context.Trace{
									Session: "session-spectre",
								},
							},
						}))
					})

					It("does not add trace session if not present", func() {
						delete(request.Env, "TRACE-SESSION")
						responder.RespondWithStatusAndErrors(403, testErrors)
						Expect(request.Env["ERRORS"]).To(Equal(testErrors))
						Expect(response.WriteHeaderInputs).To(ConsistOf(403))
						Expect(response.WriteJsonInputs).To(ConsistOf(&context.JSONResponse{
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
					response.WriteJsonOutputs = []error{nil}
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
						responder.RespondWithStatusAndData(200, testData)
						Expect(response.WriteHeaderInputs).To(ConsistOf(200))
						Expect(response.WriteJsonInputs).To(ConsistOf(&context.JSONResponse{
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
						delete(request.Env, "TRACE-REQUEST")
						responder.RespondWithStatusAndData(201, testData)
						Expect(response.WriteHeaderInputs).To(ConsistOf(201))
						Expect(response.WriteJsonInputs).To(ConsistOf(&context.JSONResponse{
							Data: testData,
							Meta: &context.Meta{
								Trace: &context.Trace{
									Session: "session-spectre",
								},
							},
						}))
					})

					It("does not add trace session if not present", func() {
						delete(request.Env, "TRACE-SESSION")
						responder.RespondWithStatusAndData(202, testData)
						Expect(response.WriteHeaderInputs).To(ConsistOf(202))
						Expect(response.WriteJsonInputs).To(ConsistOf(&context.JSONResponse{
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
