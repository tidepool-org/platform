package device_check_test

import (
	"bytes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidepool-org/platform/apple/device_check"
	"github.com/tidepool-org/platform/apple/test"
	"io/ioutil"
	"net/http"
)

var _ = Describe("DeviceChecker", func() {
	Context("New", func() {
		It("returns successfully", func() {
			cfg := &device_check.DeviceCheckConfig{
				test.PrivateKey,
				test.Issuer,
				test.Kid,
				true,
			}
			httpClient := &http.Client{
				Timeout: 2,
			}
			Expect(device_check.New(cfg, httpClient)).ToNot(BeNil())
		})
	})

	Context("IsValidDeviceToken", func() {
		cfg := &device_check.DeviceCheckConfig{
			test.PrivateKey,
			test.Issuer,
			test.Kid,
			false,
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
			httpClient := test.NewTestHttpClient(func(req *http.Request) *http.Response {
				Expect(req.URL.String()).To(Equal("https://api.devicecheck.apple.com/v1/validate_device_token"))
				Expect(req.Method).To(Equal(http.MethodPost))
				Expect(req.Header).To(HaveKey("Authorization"))
				return successfulResponse
			})
			deviceChecker := device_check.New(cfg, httpClient)
			deviceChecker.IsValidDeviceToken("device-token")
		})

		It("returns true on successful response", func() {
			httpClient := test.NewTestHttpClient(func(req *http.Request) *http.Response {
				return successfulResponse
			})
			deviceChecker := device_check.New(cfg, httpClient)
			result, err := deviceChecker.IsValidDeviceToken("device-token")
			Expect(err).To(Not(HaveOccurred()))
			Expect(result).To(BeTrue())
		})

		It("returns false on bad device token response", func() {
			httpClient := test.NewTestHttpClient(func(req *http.Request) *http.Response {
				return badDeviceTokenResponse
			})
			deviceChecker := device_check.New(cfg, httpClient)
			result, err := deviceChecker.IsValidDeviceToken("device-token")
			Expect(err).To(Not(HaveOccurred()))
			Expect(result).To(BeFalse())
		})

		It("returns an error on other error response", func() {
			httpClient := test.NewTestHttpClient(func(req *http.Request) *http.Response {
				return failedResponse
			})
			deviceChecker := device_check.New(cfg, httpClient)
			_, err := deviceChecker.IsValidDeviceToken("device-token")
			Expect(err).To(HaveOccurred())
		})
	})
})
