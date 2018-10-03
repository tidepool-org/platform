package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"github.com/tidepool-org/platform/notification"
	notificationClient "github.com/tidepool-org/platform/notification/client"
	"github.com/tidepool-org/platform/platform"
	testHTTP "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Client", func() {
	var cfg *platform.Config

	BeforeEach(func() {
		cfg = platform.NewConfig()
		Expect(cfg).ToNot(BeNil())
	})

	Context("New", func() {
		BeforeEach(func() {
			cfg.Address = testHTTP.NewAddress()
			cfg.UserAgent = testHTTP.NewUserAgent()
		})

		It("returns an error if unsuccessful", func() {
			clnt, err := notificationClient.New(nil, platform.AuthorizeAsService)
			Expect(err).To(HaveOccurred())
			Expect(clnt).To(BeNil())
		})

		It("returns success", func() {
			clnt, err := notificationClient.New(cfg, platform.AuthorizeAsService)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
		})
	})

	Context("with server and new client", func() {
		var svr *Server
		var userAgent string
		var clnt notification.Client

		BeforeEach(func() {
			svr = NewServer()
			userAgent = testHTTP.NewUserAgent()
			Expect(svr).ToNot(BeNil())
			cfg.Address = svr.URL()
			cfg.UserAgent = userAgent
			var err error
			clnt, err = notificationClient.New(cfg, platform.AuthorizeAsService)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
		})

		AfterEach(func() {
			if svr != nil {
				svr.Close()
			}
		})
	})
})
