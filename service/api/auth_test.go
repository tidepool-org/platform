package api_test

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	authTest "github.com/tidepool-org/platform/auth/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service/api"
	serviceTest "github.com/tidepool-org/platform/service/test"
	testRest "github.com/tidepool-org/platform/test/rest"
)

var _ = Describe("Auth", func() {
	var res *testRest.ResponseWriter
	var req *rest.Request
	var handlerFunc rest.HandlerFunc
	var details request.Details

	BeforeEach(func() {
		res = testRest.NewResponseWriter()
		res.HeaderOutput = &http.Header{}
		req = testRest.NewRequest()
		req.Request = req.WithContext(log.NewContextWithLogger(req.Context(), logNull.NewLogger()))
		handlerFunc = func(res rest.ResponseWriter, req *rest.Request) {
			Expect(res).ToNot(BeNil())
			Expect(req).ToNot(BeNil())
			res.WriteHeader(0)
		}
		details = nil
	})

	JustBeforeEach(func() {
		req.Request = req.WithContext(request.NewContextWithDetails(req.Context(), details))
	})

	AfterEach(func() {
		res.AssertOutputsEmpty()
	})

	Context("Require", func() {
		It("does nothing if handlerFunc is nil", func() {
			requireFunc := api.Require(nil)
			Expect(requireFunc).ToNot(BeNil())
			requireFunc(res, req)
			Expect(res.WriteHeaderInputs).To(BeEmpty())
			Expect(res.WriteInputs).To(BeEmpty())
		})

		Context("with handlerFunc func", func() {
			var requireFunc rest.HandlerFunc

			BeforeEach(func() {
				requireFunc = api.Require(handlerFunc)
				Expect(requireFunc).ToNot(BeNil())
			})

			It("does nothing if response is nil", func() {
				requireFunc(nil, req)
				Expect(res.WriteHeaderInputs).To(BeEmpty())
				Expect(res.WriteInputs).To(BeEmpty())
			})

			It("does nothing if request is nil", func() {
				requireFunc(res, nil)
				Expect(res.WriteHeaderInputs).To(BeEmpty())
				Expect(res.WriteInputs).To(BeEmpty())
			})

			It("responds with unauthenticated error if details are missing", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				requireFunc(res, req)
				Expect(res.WriteHeaderInputs).To(Equal([]int{401}))
				Expect(res.WriteInputs).To(HaveLen(1))
				errorsTest.ExpectErrorJSON(request.ErrorUnauthenticated(), res.WriteInputs[0])
			})

			Context("with server details", func() {
				BeforeEach(func() {
					details = request.NewDetails(request.MethodSessionToken, "", authTest.NewSessionToken())
				})

				It("responds successfully", func() {
					requireFunc(res, req)
					Expect(res.WriteHeaderInputs).To(Equal([]int{0}))
					Expect(res.WriteInputs).To(BeEmpty())
				})
			})

			Context("with user details", func() {
				BeforeEach(func() {
					details = request.NewDetails(request.MethodSessionToken, serviceTest.NewUserID(), authTest.NewSessionToken())
				})

				It("responds successfully", func() {
					requireFunc(res, req)
					Expect(res.WriteHeaderInputs).To(Equal([]int{0}))
					Expect(res.WriteInputs).To(BeEmpty())
				})
			})
		})
	})

	Context("RequireServer", func() {
		It("does nothing if handlerFunc is nil", func() {
			requireFunc := api.RequireServer(nil)
			Expect(requireFunc).ToNot(BeNil())
			requireFunc(res, req)
			Expect(res.WriteHeaderInputs).To(BeEmpty())
			Expect(res.WriteInputs).To(BeEmpty())
		})

		Context("with handlerFunc func", func() {
			var requireFunc rest.HandlerFunc

			BeforeEach(func() {
				requireFunc = api.RequireServer(handlerFunc)
				Expect(requireFunc).ToNot(BeNil())
			})

			It("does nothing if response is nil", func() {
				requireFunc(nil, req)
				Expect(res.WriteHeaderInputs).To(BeEmpty())
				Expect(res.WriteInputs).To(BeEmpty())
			})

			It("does nothing if request is nil", func() {
				requireFunc(res, nil)
				Expect(res.WriteHeaderInputs).To(BeEmpty())
				Expect(res.WriteInputs).To(BeEmpty())
			})

			It("responds with unauthenticated error if details are missing", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				requireFunc(res, req)
				Expect(res.WriteHeaderInputs).To(Equal([]int{401}))
				Expect(res.WriteInputs).To(HaveLen(1))
				errorsTest.ExpectErrorJSON(request.ErrorUnauthenticated(), res.WriteInputs[0])
			})

			Context("with server details", func() {
				BeforeEach(func() {
					details = request.NewDetails(request.MethodSessionToken, "", authTest.NewSessionToken())
				})

				It("responds successfully", func() {
					requireFunc(res, req)
					Expect(res.WriteHeaderInputs).To(Equal([]int{0}))
					Expect(res.WriteInputs).To(BeEmpty())
				})
			})

			Context("with user details", func() {
				BeforeEach(func() {
					details = request.NewDetails(request.MethodSessionToken, serviceTest.NewUserID(), authTest.NewSessionToken())
				})

				It("responds with unauthorized error", func() {
					res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
					requireFunc(res, req)
					Expect(res.WriteHeaderInputs).To(Equal([]int{403}))
					Expect(res.WriteInputs).To(HaveLen(1))
					errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
				})
			})
		})
	})

	Context("RequireUser", func() {
		It("does nothing if handlerFunc is nil", func() {
			requireFunc := api.RequireUser(nil)
			Expect(requireFunc).ToNot(BeNil())
			requireFunc(res, req)
			Expect(res.WriteHeaderInputs).To(BeEmpty())
			Expect(res.WriteInputs).To(BeEmpty())
		})

		Context("with handlerFunc func", func() {
			var requireFunc rest.HandlerFunc

			BeforeEach(func() {
				requireFunc = api.RequireUser(handlerFunc)
				Expect(requireFunc).ToNot(BeNil())
			})

			It("does nothing if response is nil", func() {
				requireFunc(nil, req)
				Expect(res.WriteHeaderInputs).To(BeEmpty())
				Expect(res.WriteInputs).To(BeEmpty())
			})

			It("does nothing if request is nil", func() {
				requireFunc(res, nil)
				Expect(res.WriteHeaderInputs).To(BeEmpty())
				Expect(res.WriteInputs).To(BeEmpty())
			})

			It("responds with unauthenticated error if details are missing", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				requireFunc(res, req)
				Expect(res.WriteHeaderInputs).To(Equal([]int{401}))
				Expect(res.WriteInputs).To(HaveLen(1))
				errorsTest.ExpectErrorJSON(request.ErrorUnauthenticated(), res.WriteInputs[0])
			})

			Context("with server details", func() {
				BeforeEach(func() {
					details = request.NewDetails(request.MethodSessionToken, "", authTest.NewSessionToken())
				})

				It("responds with unauthorized error", func() {
					res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
					requireFunc(res, req)
					Expect(res.WriteHeaderInputs).To(Equal([]int{403}))
					Expect(res.WriteInputs).To(HaveLen(1))
					errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
				})
			})

			Context("with user details", func() {
				BeforeEach(func() {
					details = request.NewDetails(request.MethodSessionToken, serviceTest.NewUserID(), authTest.NewSessionToken())
				})

				It("responds successfully", func() {
					requireFunc(res, req)
					Expect(res.WriteHeaderInputs).To(Equal([]int{0}))
					Expect(res.WriteInputs).To(BeEmpty())
				})
			})
		})
	})
})
