package apple_test

import (
	"bytes"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/apple"
	"github.com/tidepool-org/platform/apple/test"
)

var _ = Describe("DeviceChecker", func() {
	Context("NewDeviceCheck", func() {
		It("returns successfully", func() {
			cfg := &apple.DeviceCheckConfig{
				PrivateKey:                test.PrivateKey,
				Issuer:                    test.Issuer,
				KeyID:                     test.Kid,
				UseDevelopmentEnvironment: true,
			}
			httpClient := &http.Client{
				Timeout: 2,
			}
			Expect(apple.NewDeviceCheck(cfg, httpClient)).ToNot(BeNil())
		})
	})

	Context("IsValidDeviceToken", func() {
		cfg := &apple.DeviceCheckConfig{
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
			Body:       ioutil.NopCloser(bytes.NewBufferString("the device token is missing or badly formatted")),
			Header:     make(http.Header),
		}

		unauthorizedDeviceTokenResponse := &http.Response{
			StatusCode: 401,
			Body:       ioutil.NopCloser(bytes.NewBufferString("the authentication token can't be verified")),
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
			deviceChecker := apple.NewDeviceCheck(cfg, httpClient)
			_, _ = deviceChecker.IsTokenValid("device-token")
		})

		It("returns true on successful response", func() {
			httpClient := test.NewTestHTTPClient(func(req *http.Request) *http.Response {
				return successfulResponse
			})
			deviceChecker := apple.NewDeviceCheck(cfg, httpClient)
			result, err := deviceChecker.IsTokenValid("device-token")
			Expect(err).To(Not(HaveOccurred()))
			Expect(result).To(BeTrue())
		})

		It("returns false on bad device token response", func() {
			httpClient := test.NewTestHTTPClient(func(req *http.Request) *http.Response {
				return badDeviceTokenResponse
			})
			deviceChecker := apple.NewDeviceCheck(cfg, httpClient)
			result, err := deviceChecker.IsTokenValid("device-token")
			Expect(err).To(Not(HaveOccurred()))
			Expect(result).To(BeFalse())
		})

		It("returns false on unauthorized token response", func() {
			httpClient := test.NewTestHTTPClient(func(req *http.Request) *http.Response {
				return unauthorizedDeviceTokenResponse
			})
			deviceChecker := apple.NewDeviceCheck(cfg, httpClient)
			result, err := deviceChecker.IsTokenValid("device-token")
			Expect(err).To(Not(HaveOccurred()))
			Expect(result).To(BeFalse())
		})

		It("returns an error on other error response", func() {
			httpClient := test.NewTestHTTPClient(func(req *http.Request) *http.Response {
				return failedResponse
			})
			deviceChecker := apple.NewDeviceCheck(cfg, httpClient)
			_, err := deviceChecker.IsTokenValid("device-token")
			Expect(err).To(HaveOccurred())
		})
	})
})
