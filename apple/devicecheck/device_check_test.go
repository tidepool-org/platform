package devicecheck_test

import (
	"bytes"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/apple/devicecheck"
	"github.com/tidepool-org/platform/apple/test"
)

var _ = Describe("DeviceChecker", func() {
	Context("New", func() {
		It("returns successfully", func() {
			cfg := &devicecheck.Config{
				PrivateKey:                test.PrivateKey,
				Issuer:                    test.Issuer,
				KeyID:                     test.Kid,
				UseDevelopmentEnvironment: true,
			}
			httpClient := &http.Client{
				Timeout: 2,
			}
			Expect(devicecheck.New(cfg, httpClient)).ToNot(BeNil())
		})
	})

	Context("IsValidDeviceToken", func() {
		cfg := &devicecheck.Config{
			PrivateKey:                test.PrivateKey,
			Issuer:                    test.Issuer,
			KeyID:                     test.Kid,
			UseDevelopmentEnvironment: false,
		}

		successfulResponse := &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString("")),
			Header:     make(http.Header),
		}

		badDeviceTokenResponse := &http.Response{
			StatusCode: 400,
			Body:       ioutil.NopCloser(bytes.NewBufferString("Missing or incorrectly formatted device token payload")),
			Header:     make(http.Header),
		}

		failedResponse := &http.Response{
			StatusCode: 500,
			Body:       ioutil.NopCloser(bytes.NewBufferString("Server error")),
			Header:     make(http.Header),
		}

		It("calls the apple production api", func() {
			httpClient := test.NewTestHTTPClient(func(req *http.Request) *http.Response {
				Expect(req.URL.String()).To(Equal("https://api.devicecheck.apple.com/v1/validate_device_token"))
				Expect(req.Method).To(Equal(http.MethodPost))
				Expect(req.Header).To(HaveKey("Authorization"))
				return successfulResponse
			})
			deviceChecker := devicecheck.New(cfg, httpClient)
			deviceChecker.IsValidDeviceToken("device-token")
		})

		It("returns true on successful response", func() {
			httpClient := test.NewTestHTTPClient(func(req *http.Request) *http.Response {
				return successfulResponse
			})
			deviceChecker := devicecheck.New(cfg, httpClient)
			result, err := deviceChecker.IsValidDeviceToken("device-token")
			Expect(err).To(Not(HaveOccurred()))
			Expect(result).To(BeTrue())
		})

		It("returns false on bad device token response", func() {
			httpClient := test.NewTestHTTPClient(func(req *http.Request) *http.Response {
				return badDeviceTokenResponse
			})
			deviceChecker := devicecheck.New(cfg, httpClient)
			result, err := deviceChecker.IsValidDeviceToken("device-token")
			Expect(err).To(Not(HaveOccurred()))
			Expect(result).To(BeFalse())
		})

		It("returns an error on other error response", func() {
			httpClient := test.NewTestHTTPClient(func(req *http.Request) *http.Response {
				return failedResponse
			})
			deviceChecker := devicecheck.New(cfg, httpClient)
			_, err := deviceChecker.IsValidDeviceToken("device-token")
			Expect(err).To(HaveOccurred())
		})
	})
})
