package client_test

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("Client", func() {
	// 	var cfg *client.Config

	// 	BeforeEach(func() {
	// 		cfg = client.NewConfig()
	// 		Expect(cfg).ToNot(BeNil())
	// 	})

	// 	Context("New", func() {
	// 		BeforeEach(func() {
	// 			cfg.Address = "http://localhost:1234"
	// 		})

	// 		It("returns an error if unsuccessful", func() {
	// 			clnt, err := taskClient.New(nil)
	// 			Expect(err).To(HaveOccurred())
	// 			Expect(clnt).To(BeNil())
	// 		})

	// 		It("returns success", func() {
	// 			clnt, err := taskClient.New(cfg)
	// 			Expect(err).ToNot(HaveOccurred())
	// 			Expect(clnt).ToNot(BeNil())
	// 		})
	// 	})

	// 	Context("with server and new client", func() {
	// 		var svr *Server
	// 		var clnt task.Client
	// 		var ctx *testAuth.Context

	// 		BeforeEach(func() {
	// 			svr = NewServer()
	// 			Expect(svr).ToNot(BeNil())
	// 			cfg.Address = svr.URL()
	// 			var err error
	// 			clnt, err = taskClient.New(cfg)
	// 			Expect(err).ToNot(HaveOccurred())
	// 			Expect(clnt).ToNot(BeNil())
	// 			ctx = testAuth.NewContext()
	// 			Expect(ctx).ToNot(BeNil())
	// 		})

	// 		AfterEach(func() {
	// 			if svr != nil {
	// 				svr.Close()
	// 			}
	// 			Expect(ctx.UnusedOutputsCount()).To(Equal(0))
	// 		})
	// 	})
})
