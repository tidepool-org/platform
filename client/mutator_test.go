package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http"

	"github.com/tidepool-org/platform/client"
	testHTTP "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Mutator", func() {
	Context("HeaderMutator", func() {
		var key string
		var value string

		BeforeEach(func() {
			key = testHTTP.NewHeaderKey()
			value = testHTTP.NewHeaderValue()
		})

		Context("NewHeaderMutator", func() {
			It("returns successfully", func() {
				Expect(client.NewHeaderMutator(key, value)).ToNot(BeNil())
			})
		})

		Context("with new header mutator", func() {
			var mutator *client.HeaderMutator

			BeforeEach(func() {
				mutator = client.NewHeaderMutator(key, value)
				Expect(mutator).ToNot(BeNil())
			})

			It("remembers the key", func() {
				Expect(mutator.Key).To(Equal(key))
			})

			It("remembers the value", func() {
				Expect(mutator.Value).To(Equal(value))
			})

			Context("Mutate", func() {
				var request *http.Request

				BeforeEach(func() {
					request = testHTTP.NewRequest()
				})

				It("returns an error if the request is missing", func() {
					Expect(mutator.Mutate(nil)).To(MatchError("request is missing"))
				})

				It("returns an error if the key is missing", func() {
					mutator.Key = ""
					Expect(mutator.Mutate(request)).To(MatchError("key is missing"))
				})

				It("adds the header", func() {
					Expect(mutator.Mutate(request)).To(Succeed())
					Expect(request.Header).To(HaveLen(1))
					Expect(request.Header).To(HaveKeyWithValue(key, []string{value}))
				})

				It("adds the header even if there are already headers", func() {
					existingKey := testHTTP.NewHeaderKey()
					existingValue := testHTTP.NewHeaderValue()
					request.Header.Add(existingKey, existingValue)
					Expect(mutator.Mutate(request)).To(Succeed())
					Expect(request.Header).To(HaveLen(2))
					Expect(request.Header).To(HaveKeyWithValue(existingKey, []string{existingValue}))
					Expect(request.Header).To(HaveKeyWithValue(key, []string{value}))
				})

				It("adds the header even if there are already headers with the same key", func() {
					existingValue := testHTTP.NewHeaderValue()
					request.Header.Add(key, existingValue)
					Expect(mutator.Mutate(request)).To(Succeed())
					Expect(request.Header).To(HaveLen(1))
					Expect(request.Header).To(HaveKeyWithValue(key, []string{existingValue, value}))
				})
			})
		})
	})

	Context("ParameterMutator", func() {
		var key string
		var value string

		BeforeEach(func() {
			key = testHTTP.NewParameterKey()
			value = testHTTP.NewParameterValue()
		})

		Context("NewParameterMutator", func() {
			It("returns successfully", func() {
				Expect(client.NewParameterMutator(key, value)).ToNot(BeNil())
			})
		})

		Context("with new parameter mutator", func() {
			var mutator *client.ParameterMutator

			BeforeEach(func() {
				mutator = client.NewParameterMutator(key, value)
				Expect(mutator).ToNot(BeNil())
			})

			It("remembers the key", func() {
				Expect(mutator.Key).To(Equal(key))
			})

			It("remembers the value", func() {
				Expect(mutator.Value).To(Equal(value))
			})

			Context("Mutate", func() {
				var request *http.Request

				BeforeEach(func() {
					request = testHTTP.NewRequest()

				})

				It("returns an error if the request is missing", func() {
					Expect(mutator.Mutate(nil)).To(MatchError("request is missing"))
				})

				It("returns an error if the key is missing", func() {
					mutator.Key = ""
					Expect(mutator.Mutate(request)).To(MatchError("key is missing"))
				})

				It("adds the parameter", func() {
					Expect(mutator.Mutate(request)).To(Succeed())
					Expect(request.URL.Query()).To(HaveLen(1))
					Expect(request.URL.Query()).To(HaveKeyWithValue(key, []string{value}))
				})

				It("adds the parameter even if there are already parameters", func() {
					existingKey := testHTTP.NewParameterKey()
					existingValue := testHTTP.NewParameterValue()
					query := request.URL.Query()
					query.Add(existingKey, existingValue)
					request.URL.RawQuery = query.Encode()
					Expect(mutator.Mutate(request)).To(Succeed())
					Expect(request.URL.Query()).To(HaveLen(2))
					Expect(request.URL.Query()).To(HaveKeyWithValue(existingKey, []string{existingValue}))
					Expect(request.URL.Query()).To(HaveKeyWithValue(key, []string{value}))
				})

				It("adds the parameter even if there are already parameters with the same key", func() {
					existingValue := testHTTP.NewParameterValue()
					query := request.URL.Query()
					query.Add(key, existingValue)
					request.URL.RawQuery = query.Encode()
					Expect(mutator.Mutate(request)).To(Succeed())
					Expect(request.URL.Query()).To(HaveLen(1))
					Expect(request.URL.Query()).To(HaveKeyWithValue(key, []string{existingValue, value}))
				})
			})
		})
	})
})
