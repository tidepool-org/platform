package service_test

import (
	"net/http"
	"os"

	eventsTest "github.com/tidepool-org/platform/events/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	applicationTest "github.com/tidepool-org/platform/application/test"
	authServiceService "github.com/tidepool-org/platform/auth/service/service"
	authTest "github.com/tidepool-org/platform/auth/test"
	configTest "github.com/tidepool-org/platform/config/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Service", func() {
	Context("New", func() {
		It("returns successfully", func() {
			Expect(authServiceService.New()).ToNot(BeNil())
		})
	})

	Context("with started server, config reporter, and new service", func() {
		var provider *applicationTest.Provider
		var serverSecret string
		var sessionToken string
		var server *Server
		var authClientConfig map[string]interface{}
		var authStoreConfig map[string]interface{}
		var dataClientConfig map[string]interface{}
		var dataSourceClientConfig map[string]interface{}
		var taskClientConfig map[string]interface{}
		var authServiceConfig map[string]interface{}
		var service *authServiceService.Service
		var oldKafkaConfig map[string]string

		BeforeEach(func() {
			provider = applicationTest.NewProviderWithDefaults()

			serverSecret = authTest.NewServiceSecret()
			sessionToken = authTest.NewSessionToken()
			server = NewServer()
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest("POST", "/serverlogin"), // by default the path prefix is empty to the auth service unless set in the env var TIDEPOOL_AUTH_CLIENT_EXTERNAL_PATH_PREFIX
					VerifyHeaderKV("X-Tidepool-Server-Name", *provider.NameOutput),
					VerifyHeaderKV("X-Tidepool-Server-Secret", serverSecret),
					VerifyBody(nil),
					RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{sessionToken}})),
			)

			authClientConfig = map[string]interface{}{
				"external": map[string]interface{}{
					"address":                     server.URL(),
					"server_session_token_secret": serverSecret,
				},
			}
			authStoreConfig = map[string]interface{}{
				"addresses": os.Getenv("TIDEPOOL_STORE_ADDRESSES"),
				"database":  test.RandomStringFromRangeAndCharset(4, 8, test.CharsetLowercase),
				"tls":       "false",
			}
			dataClientConfig = map[string]interface{}{
				"address":             server.URL(),
				"server_token_secret": authTest.NewServiceSecret(),
			}
			dataSourceClientConfig = map[string]interface{}{
				"address":             server.URL(),
				"server_token_secret": authTest.NewServiceSecret(),
			}
			taskClientConfig = map[string]interface{}{
				"address":             server.URL(),
				"server_token_secret": authTest.NewServiceSecret(),
			}

			authServiceConfig = map[string]interface{}{
				"auth": map[string]interface{}{
					"client": authClientConfig,
					"store":  authStoreConfig,
				},
				"data": map[string]interface{}{
					"client": dataClientConfig,
				},
				"data_source": map[string]interface{}{
					"client": dataSourceClientConfig,
				},
				"domain": "test.com",
				"secret": authTest.NewServiceSecret(),
				"server": map[string]interface{}{
					"address": testHttp.NewAddress(),
					"tls":     "false",
				},
				"task": map[string]interface{}{
					"client": taskClientConfig,
				},
			}

			(*provider.ConfigReporterOutput).(*configTest.Reporter).Config = authServiceConfig
			oldKafkaConfig = eventsTest.SetTestEnvironmentVariables()

			service = authServiceService.New()
			Expect(service).ToNot(BeNil())
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
			eventsTest.RestoreOldEnvironmentVariables(oldKafkaConfig)
		})

		Context("Initialize", func() {
			It("returns an error when the provider is missing", func() {
				errorsTest.ExpectEqual(service.Initialize(nil), errors.New("provider is missing"))
			})

			It("returns an error when the underlying service returns an error", func() {
				dataSourceClientConfig["address"] = ""
				errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to create data source client"))
			})

			It("returns an error when the auth store config load returns an error", func() {
				timeout, timeoutSet := os.LookupEnv("TIDEPOOL_STORE_TIMEOUT")
				os.Setenv("TIDEPOOL_STORE_TIMEOUT", "invalid")
				errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to load auth store config"))
				if timeoutSet {
					os.Setenv("TIDEPOOL_STORE_TIMEOUT", timeout)
				} else {
					os.Unsetenv("TIDEPOOL_STORE_TIMEOUT")
				}
			})

			It("returns an error when the auth store returns an error", func() {
				addresses := os.Getenv("TIDEPOOL_STORE_ADDRESSES")
				os.Setenv("TIDEPOOL_STORE_ADDRESSES", "")
				errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to create auth store"))
				os.Setenv("TIDEPOOL_STORE_ADDRESSES", addresses)
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

			Context("AuthStore", func() {
				It("returns successfully", func() {
					Expect(service.AuthStore()).ToNot(BeNil())
				})
			})
		})
	})
})
