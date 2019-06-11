package api_test

import (
	awsSdkGoAwsSession "github.com/aws/aws-sdk-go/aws/session"
	awsSdkGoServiceS3 "github.com/aws/aws-sdk-go/service/s3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/aws"
	awsApi "github.com/tidepool-org/platform/aws/api"
	awsTest "github.com/tidepool-org/platform/aws/test"
	"github.com/tidepool-org/platform/pointer"
)

var _ = Describe("API", func() {
	var session *awsSdkGoAwsSession.Session

	BeforeEach(func() {
		var err error
		session, err = awsSdkGoAwsSession.NewSession()
		Expect(err).ToNot(HaveOccurred())
		Expect(session).ToNot(BeNil())
	})

	Context("New", func() {
		It("returns an error if the aws session is missing", func() {
			api, err := awsApi.New(nil)
			Expect(err).To(MatchError("aws session is missing"))
			Expect(api).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(awsApi.New(session)).ToNot(BeNil())
		})
	})

	Context("with new api", func() {
		var api *awsApi.API

		BeforeEach(func() {
			var err error
			api, err = awsApi.New(session)
			Expect(err).ToNot(HaveOccurred())
			Expect(api).ToNot(BeNil())
		})

		Context("S3", func() {
			It("returns successfully", func() {
				Expect(api.S3()).ToNot(BeNil())
			})
		})

		Context("S3Manager", func() {
			It("returns successfully", func() {
				Expect(api.S3Manager()).ToNot(BeNil())
			})
		})

		Context("with s3 manager", func() {
			var s3Manager aws.S3Manager

			BeforeEach(func() {
				s3Manager = api.S3Manager()
				Expect(s3Manager).ToNot(BeNil())
			})

			Context("Downloader", func() {
				It("returns successfully", func() {
					Expect(s3Manager.Downloader()).ToNot(BeNil())
				})
			})

			Context("Uploader", func() {
				It("returns successfully", func() {
					Expect(s3Manager.Uploader()).ToNot(BeNil())
				})
			})

			Context("NewBatchDeleteWithClient", func() {
				It("returns successfully", func() {
					Expect(s3Manager.NewBatchDeleteWithClient()).ToNot(BeNil())
				})
			})

			Context("NewDeleteListIterator", func() {
				It("returns successfully", func() {
					listObjectsInput := &awsSdkGoServiceS3.ListObjectsInput{
						Bucket: pointer.FromString(awsTest.RandomBucket()),
					}
					Expect(s3Manager.NewDeleteListIterator(listObjectsInput)).ToNot(BeNil())
				})
			})
		})
	})
})
