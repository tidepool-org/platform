package service_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"net/http"
	"os"

	applicationTest "github.com/tidepool-org/platform/application/test"
	authTest "github.com/tidepool-org/platform/auth/test"
	blobService "github.com/tidepool-org/platform/blob/service"
	configTest "github.com/tidepool-org/platform/config/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Service", func() {
	Context("New", func() {
		It("returns successfully", func() {
			Expect(blobService.New()).ToNot(BeNil())
		})
	})

	Context("with started server, config reporter, and new service", func() {
		var provider *applicationTest.Provider
		var serverSecret string
		var sessionToken string
		var server *Server
		var authClientConfig map[string]interface{}
		var blobStructuredStoreConfig map[string]interface{}
		var blobUnstructuredStoreConfig map[string]interface{}
		var userClientConfig map[string]interface{}
		var blobServiceConfig map[string]interface{}
		var service *blobService.Service

		BeforeEach(func() {
			provider = applicationTest.NewProviderWithDefaults()

			serverSecret = authTest.NewServiceSecret()
			sessionToken = authTest.NewSessionToken()
			server = NewServer()
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest("POST", "/auth/serverlogin"),
					VerifyHeaderKV("X-Tidepool-Server-Name", *provider.NameOutput),
					VerifyHeaderKV("X-Tidepool-Server-Secret", serverSecret),
					VerifyBody(nil),
					RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{sessionToken}})),
			)

			authClientConfig = map[string]interface{}{
				"address":             server.URL(),
				"server_token_secret": authTest.NewServiceSecret(),
				"external": map[string]interface{}{
					"address":                     server.URL(),
					"server_session_token_secret": serverSecret,
				},
			}
			blobStructuredStoreConfig = map[string]interface{}{
				"addresses": os.Getenv("TIDEPOOL_STORE_ADDRESSES"),
				"database":  test.RandomStringFromRangeAndCharset(4, 8, test.CharsetLowercase),
				"tls":       "false",
			}
			blobUnstructuredStoreConfig = map[string]interface{}{
				"type": "s3",
				"s3": map[string]interface{}{
					"bucket": test.RandomStringFromRangeAndCharset(4, 8, test.CharsetLowercase),
					"prefix": test.RandomStringFromRangeAndCharset(4, 8, test.CharsetLowercase),
				},
			}
			userClientConfig = map[string]interface{}{
				"address": server.URL(),
			}
			blobServiceConfig = map[string]interface{}{
				"auth": map[string]interface{}{
					"client": authClientConfig,
				},
				"blob": map[string]interface{}{
					"structured": map[string]interface{}{
						"store": blobStructuredStoreConfig,
					},
					"unstructured": map[string]interface{}{
						"store": blobUnstructuredStoreConfig,
					},
				},
				"secret": authTest.NewServiceSecret(),
				"server": map[string]interface{}{
					"address": testHttp.NewAddress(),
					"tls":     "false",
				},
				"user": map[string]interface{}{
					"client": userClientConfig,
				},
			}
			(*provider.ConfigReporterOutput).(*configTest.Reporter).Config = blobServiceConfig

			service = blobService.New()
			Expect(service).ToNot(BeNil())
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
			provider.AssertOutputsEmpty()
		})

		Context("with Terminate after", func() {
			AfterEach(func() {
				service.Terminate()
			})

			Context("Initialize", func() {
				It("returns an error when the provider is missing", func() {
					errorsTest.ExpectEqual(service.Initialize(nil), errors.New("provider is missing"))
				})

				It("returns an error when the underlying service returns an error", func() {
					authClientConfig["address"] = ""
					errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to create auth client"))
				})

				It("returns an error when the blob structured store config load returns an error", func() {
					blobStructuredStoreConfig["timeout"] = "invalid"
					errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to load blob structured store config"))
				})

				It("returns an error when the blob structured store returns an error", func() {
					blobStructuredStoreConfig["addresses"] = ""
					errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to create blob structured store"))
				})

				It("returns an error when the blob unstructured store returns an error", func() {
					blobUnstructuredStoreConfig["type"] = ""
					errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to create unstructured store"))
				})

				It("returns an error when the user client returns an error", func() {
					userClientConfig["address"] = ""
					errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to create user client"))
				})

				It("returns successfully", func() {
					Expect(service.Initialize(provider)).To(Succeed())
				})
			})

			Context("with Initialize before", func() {
				BeforeEach(func() {
					Expect(service.Initialize(provider)).To(Succeed())
				})

				Context("Terminate", func() {
					It("returns successfully", func() {
						service.Terminate()
					})
				})

				Context("Status", func() {
					It("returns successfully", func() {
						Expect(service.Status()).ToNot(BeNil())
					})
				})

				Context("BlobStructuredStore", func() {
					It("returns successfully", func() {
						Expect(service.BlobStructuredStore()).ToNot(BeNil())
					})
				})

				Context("BlobUnstructuredStore", func() {
					It("returns successfully", func() {
						Expect(service.BlobUnstructuredStore()).ToNot(BeNil())
					})
				})

				Context("UserClient", func() {
					It("returns successfully", func() {
						Expect(service.UserClient()).ToNot(BeNil())
					})
				})

				Context("BlobClient", func() {
					It("returns successfully", func() {
						Expect(service.BlobClient()).ToNot(BeNil())
					})
				})
			})
		})
	})
})
