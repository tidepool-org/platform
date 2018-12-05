package api_test

import (
	"github.com/aws/aws-sdk-go/aws/session"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/aws/api"
)

var _ = Describe("API", func() {
	var awsSession *session.Session

	BeforeEach(func() {
		var err error
		awsSession, err = session.NewSession()
		Expect(err).ToNot(HaveOccurred())
		Expect(awsSession).ToNot(BeNil())
	})

	Context("New", func() {
		It("returns an error if the aws session is missing", func() {
			ehpi, err := api.New(nil)
			Expect(err).To(MatchError("aws session is missing"))
			Expect(ehpi).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(api.New(awsSession)).ToNot(BeNil())
		})
	})

	Context("with new api", func() {
		var ehpi *api.API

		BeforeEach(func() {
			var err error
			ehpi, err = api.New(awsSession)
			Expect(err).ToNot(HaveOccurred())
			Expect(ehpi).ToNot(BeNil())
		})

		Context("S3", func() {
			It("returns successfully", func() {
				Expect(ehpi.S3()).ToNot(BeNil())
			})
		})

		Context("S3ManagerDownloader", func() {
			It("returns successfully", func() {
				Expect(ehpi.S3ManagerDownloader()).ToNot(BeNil())
			})
		})

		Context("S3ManagerUploader", func() {
			It("returns successfully", func() {
				Expect(ehpi.S3ManagerUploader()).ToNot(BeNil())
			})
		})
	})
})
