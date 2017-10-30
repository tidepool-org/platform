package platform_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http"

	"github.com/tidepool-org/platform/auth"
	testAuth "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
	testHTTP "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Mutator", func() {
	Context("SessionTokenHeaderMutator", func() {
		var sessionToken string

		BeforeEach(func() {
			sessionToken = testAuth.NewSessionToken()
		})

		Context("NewSessionTokenHeaderMutator", func() {
			It("returns successfully", func() {
				Expect(platform.NewSessionTokenHeaderMutator(sessionToken)).ToNot(BeNil())
			})
		})

		Context("with new session token header mutator", func() {
			var mutator *platform.SessionTokenHeaderMutator

			BeforeEach(func() {
				mutator = platform.NewSessionTokenHeaderMutator(sessionToken)
				Expect(mutator).ToNot(BeNil())
			})

			It("remembers the session token header key", func() {
				Expect(mutator.Key).To(Equal(auth.TidepoolSessionTokenHeaderKey))
			})

			It("remembers the session token header value", func() {
				Expect(mutator.Value).To(Equal(sessionToken))
			})

			Context("Mutate", func() {
				var request *http.Request

				BeforeEach(func() {
					request = testHTTP.NewRequest()
				})

				It("returns an error if the session token header value is missing", func() {
					mutator.Value = ""
					Expect(mutator.Mutate(request)).To(MatchError("session token is missing"))
				})

				It("adds the header", func() {
					Expect(mutator.Mutate(request)).To(Succeed())
					Expect(request.Header).To(HaveLen(1))
					Expect(request.Header).To(HaveKeyWithValue(auth.TidepoolSessionTokenHeaderKey, []string{sessionToken}))
				})
			})
		})
	})

	Context("RestrictedTokenParameterMutator", func() {
		var restrictedToken string

		BeforeEach(func() {
			restrictedToken = testAuth.NewRestrictedToken()
		})

		Context("NewRestrictedTokenParameterMutator", func() {
			It("returns successfully", func() {
				Expect(platform.NewRestrictedTokenParameterMutator(restrictedToken)).ToNot(BeNil())
			})
		})

		Context("with new restricted token parameter mutator", func() {
			var mutator *platform.RestrictedTokenParameterMutator

			BeforeEach(func() {
				mutator = platform.NewRestrictedTokenParameterMutator(restrictedToken)
				Expect(mutator).ToNot(BeNil())
			})

			It("remembers the restricted token parameter key", func() {
				Expect(mutator.Key).To(Equal(auth.TidepoolRestrictedTokenParameterKey))
			})

			It("remembers the restricted token parameter value", func() {
				Expect(mutator.Value).To(Equal(restrictedToken))
			})

			Context("Mutate", func() {
				var request *http.Request

				BeforeEach(func() {
					request = testHTTP.NewRequest()
				})

				It("returns an error if the restricted token parameter value is missing", func() {
					mutator.Value = ""
					Expect(mutator.Mutate(request)).To(MatchError("restricted token is missing"))
				})

				It("adds the header", func() {
					Expect(mutator.Mutate(request)).To(Succeed())
					Expect(request.URL.Query()).To(HaveLen(1))
					Expect(request.URL.Query()).To(HaveKeyWithValue(auth.TidepoolRestrictedTokenParameterKey, []string{restrictedToken}))
				})
			})
		})
	})

	Context("ServiceSecretHeaderMutator", func() {
		var serviceSecret string

		BeforeEach(func() {
			serviceSecret = testAuth.NewServiceSecret()
		})

		Context("NewServiceSecretHeaderMutator", func() {
			It("returns successfully", func() {
				Expect(platform.NewServiceSecretHeaderMutator(serviceSecret)).ToNot(BeNil())
			})
		})

		Context("with new service secret header mutator", func() {
			var mutator *platform.ServiceSecretHeaderMutator

			BeforeEach(func() {
				mutator = platform.NewServiceSecretHeaderMutator(serviceSecret)
				Expect(mutator).ToNot(BeNil())
			})

			It("remembers the service secret header key", func() {
				Expect(mutator.Key).To(Equal(auth.TidepoolServiceSecretHeaderKey))
			})

			It("remembers the service secret header value", func() {
				Expect(mutator.Value).To(Equal(serviceSecret))
			})

			Context("Mutate", func() {
				var request *http.Request

				BeforeEach(func() {
					request = testHTTP.NewRequest()
				})

				It("returns an error if the service secret header value is missing", func() {
					mutator.Value = ""
					Expect(mutator.Mutate(request)).To(MatchError("service secret is missing"))
				})

				It("adds the header", func() {
					Expect(mutator.Mutate(request)).To(Succeed())
					Expect(request.Header).To(HaveLen(1))
					Expect(request.Header).To(HaveKeyWithValue(auth.TidepoolServiceSecretHeaderKey, []string{serviceSecret}))
				})
			})
		})
	})

	Context("TraceMutator", func() {
		var traceRequest string
		var traceSession string
		var ctx context.Context

		BeforeEach(func() {
			traceRequest = test.NewString(32, test.CharsetAlphaNumeric)
			traceSession = test.NewString(32, test.CharsetAlphaNumeric)
			ctx = context.Background()
			ctx = request.NewContextWithTraceRequest(ctx, traceRequest)
			ctx = request.NewContextWithTraceSession(ctx, traceSession)
		})

		Context("NewTraceMutator", func() {
			It("returns successfully", func() {
				Expect(platform.NewTraceMutator(ctx)).ToNot(BeNil())
			})
		})

		Context("with new trace mutator", func() {
			var mutator *platform.TraceMutator

			BeforeEach(func() {
				mutator = platform.NewTraceMutator(ctx)
				Expect(mutator).ToNot(BeNil())
			})

			It("remembers the context", func() {
				Expect(mutator.Context).To(Equal(ctx))
			})

			Context("Mutate", func() {
				var req *http.Request

				BeforeEach(func() {
					req = testHTTP.NewRequest()
				})

				It("returns an error if the request is missing", func() {
					Expect(mutator.Mutate(nil)).To(MatchError("request is missing"))
				})

				It("adds the header", func() {
					Expect(mutator.Mutate(req)).To(Succeed())
					Expect(req.Header).To(HaveLen(2))
					Expect(req.Header).To(HaveKeyWithValue(request.HTTPHeaderTraceRequest, []string{traceRequest}))
					Expect(req.Header).To(HaveKeyWithValue(request.HTTPHeaderTraceSession, []string{traceSession}))
				})
			})
		})
	})
})
