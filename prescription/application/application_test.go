package application_test

import (
	"net/http"
	"os"

	"github.com/tidepool-org/platform/prescription/application"

	authTest "github.com/tidepool-org/platform/auth/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"github.com/tidepool-org/platform/test"

	applicationTest "github.com/tidepool-org/platform/application/test"
	configTest "github.com/tidepool-org/platform/config/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Application", func() {
	Context("New", func() {
		It("returns successfully", func() {
			Expect(application.New()).ToNot(BeNil())
		})
	})

	Context("with started server, config reporter, and new service", func() {
		var provider *applicationTest.Provider
		var prescriptionStoreConfig map[string]interface{}
		var prescriptionServiceConfig map[string]interface{}
		var authClientConfig map[string]interface{}
		var serverSecret string
		var sessionToken string
		var server *Server
		var service *application.Application

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

			prescriptionStoreConfig = map[string]interface{}{
				"addresses": os.Getenv("TIDEPOOL_STORE_ADDRESSES"),
				"database":  test.RandomStringFromRangeAndCharset(4, 8, test.CharsetLowercase),
				"tls":       "false",
			}

			prescriptionServiceConfig = map[string]interface{}{
				"auth": map[string]interface{}{
					"client": authClientConfig,
				},
				"domain": "test.com",
				"secret": authTest.NewServiceSecret(),
				"prescription": map[string]interface{}{
					"store": prescriptionStoreConfig,
				},
				"server": map[string]interface{}{
					"address": testHttp.NewAddress(),
					"tls":     "false",
				},
				"user": map[string]interface{}{
					"client": map[string]interface{}{
						"address": server.URL(),
					},
				},
			}

			(*provider.ConfigReporterOutput).(*configTest.Reporter).Config = prescriptionServiceConfig

			service = application.New()
			Expect(service).ToNot(BeNil())
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		Context("Initialize", func() {
			It("returns an error when the provider is missing", func() {
				errorsTest.ExpectEqual(service.Initialize(nil), errors.New("provider is missing"))
			})

			It("returns an error when the prescription store config load returns an error", func() {
				prescriptionStoreConfig["timeout"] = "invalid"
				errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to load prescription store config"))
			})

			It("returns an error when the prescription store returns an error", func() {
				prescriptionStoreConfig["addresses"] = ""
				errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to create prescription store"))
			})

			It("returns successfully", func() {
				Expect(service.Initialize(provider)).To(Succeed())
				service.Terminate()
			})
		})

		Context("with being initialized", func() {
			BeforeEach(func() {
				Expect(service.Initialize(provider)).To(Succeed())
			})

			AfterEach(func() {
				service.Terminate()
			})

			Context("Terminate", func() {
				It("returns successfully", func() {
					service.Terminate()
				})
			})
		})
	})
})
