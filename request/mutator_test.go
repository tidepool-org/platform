package request_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Mutator", func() {
	Context("HeaderMutator", func() {
		var key string
		var value string

		BeforeEach(func() {
			key = testHttp.NewHeaderKey()
			value = testHttp.NewHeaderValue()
		})

		Context("NewHeaderMutator", func() {
			It("returns successfully", func() {
				Expect(request.NewHeaderMutator(key, value)).ToNot(BeNil())
			})
		})

		Context("with new header mutator", func() {
			var mutator *request.HeaderMutator

			BeforeEach(func() {
				mutator = request.NewHeaderMutator(key, value)
				Expect(mutator).ToNot(BeNil())
			})

			It("remembers the key", func() {
				Expect(mutator.Key).To(Equal(key))
			})

			It("remembers the value", func() {
				Expect(mutator.Value).To(Equal(value))
			})

			Context("MutateRequest", func() {
				var request *http.Request

				BeforeEach(func() {
					request = testHttp.NewRequest()
				})

				It("returns an error if the request is missing", func() {
					Expect(mutator.MutateRequest(nil)).To(MatchError("request is missing"))
				})

				It("returns an error if the key is missing", func() {
					mutator.Key = ""
					Expect(mutator.MutateRequest(request)).To(MatchError("key is missing"))
				})

				It("adds the header", func() {
					Expect(mutator.MutateRequest(request)).To(Succeed())
					Expect(request.Header).To(HaveLen(1))
					Expect(request.Header).To(HaveKeyWithValue(key, []string{value}))
				})

				It("adds the header even if there are already headers", func() {
					existingKey := generateNewUniqueKey(testHttp.NewHeaderKey, map[string]any{key: struct{}{}})
					existingValue := testHttp.NewHeaderValue()
					request.Header.Add(existingKey, existingValue)
					Expect(mutator.MutateRequest(request)).To(Succeed())
					Expect(request.Header).To(HaveLen(2))
					Expect(request.Header).To(HaveKeyWithValue(existingKey, []string{existingValue}))
					Expect(request.Header).To(HaveKeyWithValue(key, []string{value}))
				})

				It("adds the header even if there are already headers with the same key", func() {
					existingValue := testHttp.NewHeaderValue()
					request.Header.Add(key, existingValue)
					Expect(mutator.MutateRequest(request)).To(Succeed())
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
			key = testHttp.NewParameterKey()
			value = testHttp.NewParameterValue()
		})

		Context("NewParameterMutator", func() {
			It("returns successfully", func() {
				Expect(request.NewParameterMutator(key, value)).ToNot(BeNil())
			})
		})

		Context("with new parameter mutator", func() {
			var mutator *request.ParameterMutator

			BeforeEach(func() {
				mutator = request.NewParameterMutator(key, value)
				Expect(mutator).ToNot(BeNil())
			})

			It("remembers the key", func() {
				Expect(mutator.Key).To(Equal(key))
			})

			It("remembers the value", func() {
				Expect(mutator.Value).To(Equal(value))
			})

			Context("MutateRequest", func() {
				var request *http.Request

				BeforeEach(func() {
					request = testHttp.NewRequest()
				})

				It("returns an error if the request is missing", func() {
					Expect(mutator.MutateRequest(nil)).To(MatchError("request is missing"))
				})

				It("returns an error if the key is missing", func() {
					mutator.Key = ""
					Expect(mutator.MutateRequest(request)).To(MatchError("key is missing"))
				})

				It("adds the parameter", func() {
					Expect(mutator.MutateRequest(request)).To(Succeed())
					Expect(request.URL.Query()).To(HaveLen(1))
					Expect(request.URL.Query()).To(HaveKeyWithValue(key, []string{value}))
				})

				It("adds the parameter even if there are already parameters", func() {
					existingKey := generateNewUniqueKey(testHttp.NewParameterKey, map[string]any{key: struct{}{}})
					existingValue := testHttp.NewParameterValue()
					query := request.URL.Query()
					query.Add(existingKey, existingValue)
					request.URL.RawQuery = query.Encode()
					Expect(mutator.MutateRequest(request)).To(Succeed())
					Expect(request.URL.Query()).To(HaveLen(2))
					Expect(request.URL.Query()).To(HaveKeyWithValue(existingKey, []string{existingValue}))
					Expect(request.URL.Query()).To(HaveKeyWithValue(key, []string{value}))
				})

				It("adds the parameter even if there are already parameters with the same key", func() {
					existingValue := testHttp.NewParameterValue()
					query := request.URL.Query()
					query.Add(key, existingValue)
					request.URL.RawQuery = query.Encode()
					Expect(mutator.MutateRequest(request)).To(Succeed())
					Expect(request.URL.Query()).To(HaveLen(1))
					Expect(request.URL.Query()).To(HaveKeyWithValue(key, []string{existingValue, value}))
				})
			})
		})
	})

	Context("ParametersMutator", func() {
		var parameters map[string]string

		BeforeEach(func() {
			parameters = map[string]string{}
			for index := test.RandomIntFromRange(0, 3); index >= 0; index-- {
				parameters[testHttp.NewParameterKey()] = testHttp.NewParameterValue()
			}
		})

		Context("NewParametersMutator", func() {
			It("returns successfully", func() {
				Expect(request.NewParametersMutator(parameters)).ToNot(BeNil())
			})
		})

		Context("with new parameters mutator", func() {
			var mutator *request.ParametersMutator

			BeforeEach(func() {
				mutator = request.NewParametersMutator(parameters)
				Expect(mutator).ToNot(BeNil())
			})

			It("remembers the parameters", func() {
				Expect(mutator.Parameters).To(Equal(parameters))
			})

			Context("MutateRequest", func() {
				var request *http.Request

				BeforeEach(func() {
					request = testHttp.NewRequest()
				})

				It("returns an error if the request is missing", func() {
					Expect(mutator.MutateRequest(nil)).To(MatchError("request is missing"))
				})

				It("returns an error if a key is missing", func() {
					mutator.Parameters[""] = testHttp.NewParameterValue()
					Expect(mutator.MutateRequest(request)).To(MatchError("key is missing"))
				})

				It("adds the parameters", func() {
					Expect(mutator.MutateRequest(request)).To(Succeed())
					Expect(request.URL.Query()).To(HaveLen(len(parameters)))
					for key, value := range parameters {
						Expect(request.URL.Query()).To(HaveKeyWithValue(key, []string{value}))
					}
				})

				It("adds the parameters even if there are already parameters", func() {
					existingKey := generateNewUniqueKey(testHttp.NewParameterKey, parameters)
					existingValue := testHttp.NewParameterValue()
					query := request.URL.Query()
					query.Add(existingKey, existingValue)
					request.URL.RawQuery = query.Encode()
					Expect(mutator.MutateRequest(request)).To(Succeed())
					Expect(request.URL.Query()).To(HaveLen(1 + len(parameters)))
					Expect(request.URL.Query()).To(HaveKeyWithValue(existingKey, []string{existingValue}))
					for key, value := range parameters {
						Expect(request.URL.Query()).To(HaveKeyWithValue(key, []string{value}))
					}
				})

				It("adds the parameters even if there are already parameters with the same key", func() {
					var existingKey string
					for existingKey = range parameters {
						break
					}
					existingValue := testHttp.NewParameterValue()
					query := request.URL.Query()
					query.Add(existingKey, existingValue)
					request.URL.RawQuery = query.Encode()
					Expect(mutator.MutateRequest(request)).To(Succeed())
					Expect(request.URL.Query()).To(HaveLen(len(parameters)))
					for key, value := range parameters {
						if key == existingKey {
							Expect(request.URL.Query()).To(HaveKeyWithValue(key, []string{existingValue, value}))
						} else {
							Expect(request.URL.Query()).To(HaveKeyWithValue(key, []string{value}))
						}
					}
				})
			})
		})
	})

	Context("ArrayParametersMutator", func() {
		var parameters map[string][]string

		BeforeEach(func() {
			parameters = map[string][]string{}
			for index := test.RandomIntFromRange(0, 3); index >= 0; index-- {
				parameters[testHttp.NewParameterKey()] = test.RandomStringArrayFromRangeAndGeneratorWithDuplicates(1, 3, testHttp.NewParameterValue)
			}
		})

		Context("NewArrayParametersMutator", func() {
			It("returns successfully", func() {
				Expect(request.NewArrayParametersMutator(parameters)).ToNot(BeNil())
			})
		})

		Context("with new parameters mutator", func() {
			var mutator *request.ArrayParametersMutator

			BeforeEach(func() {
				mutator = request.NewArrayParametersMutator(parameters)
				Expect(mutator).ToNot(BeNil())
			})

			It("remembers the parameters", func() {
				Expect(mutator.Parameters).To(Equal(parameters))
			})

			Context("MutateRequest", func() {
				var request *http.Request

				BeforeEach(func() {
					request = testHttp.NewRequest()
				})

				It("returns an error if the request is missing", func() {
					Expect(mutator.MutateRequest(nil)).To(MatchError("request is missing"))
				})

				It("returns an error if a key is missing", func() {
					mutator.Parameters[""] = test.RandomStringArrayFromRangeAndGeneratorWithDuplicates(1, 3, testHttp.NewParameterValue)
					Expect(mutator.MutateRequest(request)).To(MatchError("key is missing"))
				})

				It("adds the parameters", func() {
					Expect(mutator.MutateRequest(request)).To(Succeed())
					Expect(request.URL.Query()).To(HaveLen(len(parameters)))
					for key, value := range parameters {
						Expect(request.URL.Query()).To(HaveKeyWithValue(key, value))
					}
				})

				It("adds the parameters even if there are already parameters", func() {
					existingKey := generateNewUniqueKey(testHttp.NewParameterKey, parameters)
					existingValue := test.RandomStringArrayFromRangeAndGeneratorWithDuplicates(1, 3, testHttp.NewParameterValue)
					query := request.URL.Query()
					query[existingKey] = existingValue
					request.URL.RawQuery = query.Encode()
					Expect(mutator.MutateRequest(request)).To(Succeed())
					Expect(request.URL.Query()).To(HaveLen(1 + len(parameters)))
					Expect(request.URL.Query()).To(HaveKeyWithValue(existingKey, existingValue))
					for key, value := range parameters {
						Expect(request.URL.Query()).To(HaveKeyWithValue(key, value))
					}
				})

				It("adds the parameters even if there are already parameters with the same key", func() {
					var existingKey string
					for existingKey = range parameters {
						break
					}
					existingValue := test.RandomStringArrayFromRangeAndGeneratorWithDuplicates(1, 3, testHttp.NewParameterValue)
					query := request.URL.Query()
					query[existingKey] = existingValue
					request.URL.RawQuery = query.Encode()
					Expect(mutator.MutateRequest(request)).To(Succeed())
					Expect(request.URL.Query()).To(HaveLen(len(parameters)))
					for key, value := range parameters {
						if key == existingKey {
							Expect(request.URL.Query()).To(HaveKeyWithValue(key, append(existingValue, value...)))
						} else {
							Expect(request.URL.Query()).To(HaveKeyWithValue(key, value))
						}
					}
				})
			})
		})
	})
})
