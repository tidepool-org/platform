package service_test

import (
	"context"
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
	notificationServiceService "github.com/tidepool-org/platform/notification/service/service"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Service", func() {
	Context("New", func() {
		It("returns successfully", func() {
			Expect(notificationServiceService.New()).ToNot(BeNil())
		})
	})

	Context("with started server, config reporter, and new service", func() {
		var provider *applicationTest.Provider
		var serverSecret string
		var sessionToken string
		var server *Server
		var authClientConfig map[string]interface{}
		var notificationStoreConfig map[string]interface{}
		var notificationServiceConfig map[string]interface{}
		var service *notificationServiceService.Service

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
			notificationStoreConfig = map[string]interface{}{
				"addresses": os.Getenv("TIDEPOOL_STORE_ADDRESSES"),
				"database":  test.RandomStringFromRangeAndCharset(4, 8, test.CharsetLowercase),
				"tls":       "false",
			}

			notificationServiceConfig = map[string]interface{}{
				"auth": map[string]interface{}{
					"client": authClientConfig,
				},
				"notification": map[string]interface{}{
					"store": notificationStoreConfig,
				},
				"secret": authTest.NewServiceSecret(),
				"server": map[string]interface{}{
					"address": testHttp.NewAddress(),
					"tls":     "false",
				},
			}

			(*provider.ConfigReporterOutput).(*configTest.Reporter).Config = notificationServiceConfig

			service = notificationServiceService.New()
			Expect(service).ToNot(BeNil())
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
			provider.AssertOutputsEmpty()
		})

		Context("Initialize", func() {
			It("returns an error when the provider is missing", func() {
				errorsTest.ExpectEqual(service.Initialize(nil), errors.New("provider is missing"))
			})

			It("returns an error when the underlying service returns an error", func() {
				authClientConfig["address"] = ""
				errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to create auth client"))
			})

			It("returns an error when the notification store config load returns an error", func() {
				timeout, timeoutSet := os.LookupEnv("TIDEPOOL_STORE_TIMEOUT")
				os.Setenv("TIDEPOOL_STORE_TIMEOUT", "invalid")
				errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to load notification store config"))
				if timeoutSet {
					os.Setenv("TIDEPOOL_STORE_TIMEOUT", timeout)
				} else {
					os.Unsetenv("TIDEPOOL_STORE_TIMEOUT")
				}
			})

			It("returns an error when the notification store returns an error", func() {
				addresses := os.Getenv("TIDEPOOL_STORE_ADDRESSES")
				os.Setenv("TIDEPOOL_STORE_ADDRESSES", "")
				errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to create notification store"))
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

			Context("NotificationStore", func() {
				It("returns successfully", func() {
					Expect(service.NotificationStore()).ToNot(BeNil())
				})
			})

			Context("Status", func() {
				It("returns successfully", func() {
					Expect(service.Status(context.Background())).ToNot(BeNil())
				})
			})
		})
	})
})
