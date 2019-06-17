package service_test

import (
	"net/http"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	applicationTest "github.com/tidepool-org/platform/application/test"
	authTest "github.com/tidepool-org/platform/auth/test"
	configTest "github.com/tidepool-org/platform/config/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	imageService "github.com/tidepool-org/platform/image/service"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Service", func() {
	Context("New", func() {
		It("returns successfully", func() {
			Expect(imageService.New()).ToNot(BeNil())
		})
	})

	Context("with started server, config reporter, and new service", func() {
		var provider *applicationTest.Provider
		var serverSecret string
		var sessionToken string
		var server *Server
		var authClientConfig map[string]interface{}
		var imageStructuredStoreConfig map[string]interface{}
		var imageUnstructuredStoreConfig map[string]interface{}
		var imageServiceConfig map[string]interface{}
		var service *imageService.Service

		BeforeEach(func() {
			provider = applicationTest.NewProviderWithDefaults()

			serverSecret = authTest.NewServiceSecret()
			sessionToken = authTest.NewSessionToken()
			server = NewServer()
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/auth/serverlogin"),
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
			imageStructuredStoreConfig = map[string]interface{}{
				"addresses": os.Getenv("TIDEPOOL_STORE_ADDRESSES"),
				"database":  test.RandomStringFromRangeAndCharset(4, 8, test.CharsetLowercase),
				"tls":       "false",
			}
			imageUnstructuredStoreConfig = map[string]interface{}{
				"type": "s3",
				"s3": map[string]interface{}{
					"bucket": test.RandomStringFromRangeAndCharset(4, 8, test.CharsetLowercase),
					"prefix": test.RandomStringFromRangeAndCharset(4, 8, test.CharsetLowercase),
				},
			}
			imageServiceConfig = map[string]interface{}{
				"auth": map[string]interface{}{
					"client": authClientConfig,
				},
				"structured": map[string]interface{}{
					"store": imageStructuredStoreConfig,
				},
				"unstructured": map[string]interface{}{
					"store": imageUnstructuredStoreConfig,
				},
				"secret": authTest.NewServiceSecret(),
				"server": map[string]interface{}{
					"address": testHttp.NewAddress(),
					"tls":     "false",
				},
			}
			(*provider.ConfigReporterOutput).(*configTest.Reporter).Config = imageServiceConfig

			service = imageService.New()
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

				It("returns an error when the image structured store config load returns an error", func() {
					imageStructuredStoreConfig["timeout"] = "invalid"
					errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to load image structured store config"))
				})

				It("returns an error when the image structured store returns an error", func() {
					imageStructuredStoreConfig["addresses"] = ""
					errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to create image structured store"))
				})

				It("returns an error when the image unstructured store returns an error", func() {
					imageUnstructuredStoreConfig["type"] = ""
					errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to create unstructured store"))
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

				Context("ImageStructuredStore", func() {
					It("returns successfully", func() {
						Expect(service.ImageStructuredStore()).ToNot(BeNil())
					})
				})

				Context("ImageUnstructuredStore", func() {
					It("returns successfully", func() {
						Expect(service.ImageUnstructuredStore()).ToNot(BeNil())
					})
				})

				Context("ImageTransformer", func() {
					It("returns successfully", func() {
						Expect(service.ImageTransformer()).ToNot(BeNil())
					})
				})

				Context("ImageMultipartFormDecoder", func() {
					It("returns successfully", func() {
						Expect(service.ImageMultipartFormDecoder()).ToNot(BeNil())
					})
				})

				Context("ImageClient", func() {
					It("returns successfully", func() {
						Expect(service.ImageClient()).ToNot(BeNil())
					})
				})
			})
		})
	})
})
