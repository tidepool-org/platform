package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"github.com/tidepool-org/platform/notification"
	notificationClient "github.com/tidepool-org/platform/notification/client"
	"github.com/tidepool-org/platform/platform"
)

var _ = Describe("Client", func() {
	var cfg *platform.Config

	BeforeEach(func() {
		cfg = platform.NewConfig()
		Expect(cfg).ToNot(BeNil())
	})

	Context("New", func() {
		BeforeEach(func() {
			cfg.Address = "http://localhost:1234"
		})

		It("returns an error if unsuccessful", func() {
			clnt, err := notificationClient.New(nil)
			Expect(err).To(HaveOccurred())
			Expect(clnt).To(BeNil())
		})

		It("returns success", func() {
			clnt, err := notificationClient.New(cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
		})
	})

	Context("with server and new client", func() {
		var svr *Server
		var clnt notification.Client

		BeforeEach(func() {
			svr = NewServer()
			Expect(svr).ToNot(BeNil())
			cfg.Address = svr.URL()
			var err error
			clnt, err = notificationClient.New(cfg)
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
