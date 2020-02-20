package service_test

import (
	"net/http"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	applicationTest "github.com/tidepool-org/platform/application/test"
	authServiceService "github.com/tidepool-org/platform/auth/service/service"
	authTest "github.com/tidepool-org/platform/auth/test"
	appleTest "github.com/tidepool-org/platform/apple/test"
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
		var dataSourceClientConfig map[string]interface{}
		var taskClientConfig map[string]interface{}
		var authServiceConfig map[string]interface{}
		var appleDeviceCheckerConfig map[string]interface{}
		var service *authServiceService.Service

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
			dataSourceClientConfig = map[string]interface{}{
				"address":             server.URL(),
				"server_token_secret": authTest.NewServiceSecret(),
			}
			taskClientConfig = map[string]interface{}{
				"address":             server.URL(),
				"server_token_secret": authTest.NewServiceSecret(),
			}
			appleDeviceCheckerConfig = map[string]interface{}{
				"key_id":                      appleTest.Kid,
				"issuer":                      appleTest.Issuer,
				"private_key":                 appleTest.PrivateKey,
				"use_development_environment": "true",
			}
			authServiceConfig = map[string]interface{}{
				"apple_device_checker": appleDeviceCheckerConfig,
				"auth": map[string]interface{}{
					"client": authClientConfig,
					"store":  authStoreConfig,
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

			service = authServiceService.New()
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

			It("returns an error when the underlying service returns an error", func() {
				dataSourceClientConfig["address"] = ""
				errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to create data source client"))
			})

			It("returns an error when the auth store config load returns an error", func() {
				authStoreConfig["timeout"] = "invalid"
				errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to load auth store config"))
			})

			It("returns an error when the auth store returns an error", func() {
				authStoreConfig["addresses"] = ""
				errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to create auth store"))
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
