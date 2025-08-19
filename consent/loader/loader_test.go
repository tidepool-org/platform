package loader_test

import (
	"context"
	"errors"
	"io/fs"
	"testing/fstest"

	"github.com/tidepool-org/platform/log/test"

	"github.com/tidepool-org/platform/log"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/consent"
	"github.com/tidepool-org/platform/consent/loader"
	serviceTest "github.com/tidepool-org/platform/consent/service/test"
)

var _ = Describe("SeedConsents", func() {
	var (
		ctx         context.Context
		mockCtrl    *gomock.Controller
		mockService *serviceTest.MockService
		logger      log.Logger
	)

	BeforeEach(func() {
		ctx = context.Background()
		mockCtrl = gomock.NewController(GinkgoT())
		mockService = serviceTest.NewMockService(mockCtrl)
		logger = test.NewLogger()
	})

	AfterEach(func() {
		mockCtrl.Finish()
		loader.ResetContentFS()
	})

	Context("with embedded fs", func() {
		It("should successfully seed all valid consent files", func() {
			// Set up expectations for each valid file
			mockService.EXPECT().EnsureConsent(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, cons *consent.Consent) error {
				// Verify the consent object is properly constructed
				Expect(cons.ContentType).To(Equal(consent.ContentTypeMarkdown))
				Expect(len(cons.Type)).To(BeNumerically(">", 0))
				Expect(cons.Version).To(BeNumerically("==", 1))
				Expect(cons.Content).ToNot(BeEmpty())
				return nil
			}).Times(1)

			err := loader.SeedConsents(ctx, logger, mockService)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should create consents with correct properties", func() {
			var capturedConsents []*consent.Consent

			mockService.EXPECT().EnsureConsent(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, consent *consent.Consent) error {
				// Capture the consent for detailed verification
				capturedConsents = append(capturedConsents, consent)
				return nil
			}).Times(1)

			err := loader.SeedConsents(ctx, logger, mockService)
			Expect(err).ToNot(HaveOccurred())

			// Verify we captured
			Expect(capturedConsents).To(HaveLen(1))

			// Verify big_data_donation_project consent
			tbddp := findConsentByType(capturedConsents, "big_data_donation_project")
			Expect(tbddp).ToNot(BeNil())
			Expect(string(tbddp.Type)).To(Equal("big_data_donation_project"))
			Expect(tbddp.Version).To(Equal(1))
			Expect(tbddp.ContentType).To(Equal(consent.ContentTypeMarkdown))
			Expect(tbddp.Content).To(ContainSubstring("Tidepool Big Data Donation Project"))
		})
	})

	Context("with valid markdown files", func() {
		BeforeEach(func() {
			// Mock the embedded file system with valid consent files
			mockFS := fstest.MapFS{
				"privacy.v1.md": &fstest.MapFile{
					Data: []byte("# Privacy Policy\n\nThis is the privacy policy content."),
				},
				"terms.v2.md": &fstest.MapFile{
					Data: []byte("# Terms of Service\n\nThis is the terms of service content."),
				},
				"data-use.v10.md": &fstest.MapFile{
					Data: []byte("# Data Use Agreement\n\nThis is the data use agreement content."),
				},
			}

			loader.SetContentFS(mockFS, ".")
		})

		It("should successfully seed all valid consent files", func() {
			// Set up expectations for each valid file
			mockService.EXPECT().EnsureConsent(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, cons *consent.Consent) error {
				// Verify the consent object is properly constructed
				Expect(cons.ContentType).To(Equal(consent.ContentTypeMarkdown))
				Expect(cons.Type).ToNot(BeNil())
				Expect(cons.Version).To(BeNumerically(">", 0))
				Expect(cons.Content).ToNot(BeEmpty())
				return nil
			}).Times(3)

			err := loader.SeedConsents(ctx, logger, mockService)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should create consents with correct properties", func() {
			var capturedConsents []*consent.Consent

			mockService.EXPECT().EnsureConsent(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, consent *consent.Consent) error {
				// Capture the consent for detailed verification
				capturedConsents = append(capturedConsents, consent)
				return nil
			}).Times(3)

			err := loader.SeedConsents(ctx, logger, mockService)
			Expect(err).ToNot(HaveOccurred())

			// Verify we captured all 3 consents
			Expect(capturedConsents).To(HaveLen(3))

			// Verify privacy consent
			privacyConsent := findConsentByType(capturedConsents, "privacy")
			Expect(privacyConsent).ToNot(BeNil())
			Expect(privacyConsent.Version).To(Equal(1))
			Expect(privacyConsent.ContentType).To(Equal(consent.ContentTypeMarkdown))
			Expect(privacyConsent.Content).To(ContainSubstring("Privacy Policy"))

			// Verify terms consent
			termsConsent := findConsentByType(capturedConsents, "terms")
			Expect(termsConsent).ToNot(BeNil())
			Expect(termsConsent.Version).To(Equal(2))
			Expect(termsConsent.Content).To(ContainSubstring("Terms of Service"))

			// Verify data-use consent
			dataUseConsent := findConsentByType(capturedConsents, "data-use")
			Expect(dataUseConsent).ToNot(BeNil())
			Expect(dataUseConsent.Version).To(Equal(10))
			Expect(dataUseConsent.Content).To(ContainSubstring("Data Use Agreement"))
		})
	})

	Context("with mixed valid and invalid files", func() {
		BeforeEach(func() {
			mockFS := fstest.MapFS{
				"privacy.v1.md": &fstest.MapFile{
					Data: []byte("# Privacy Policy"),
				},
				"invalid-file.txt": &fstest.MapFile{
					Data: []byte("This should be ignored"),
				},
				"no-version.md": &fstest.MapFile{
					Data: []byte("This should be ignored"),
				},
				"subdirectory": &fstest.MapFile{
					Mode: fs.ModeDir,
				},
				"terms.v3.md": &fstest.MapFile{
					Data: []byte("# Terms of Service"),
				},
			}
			loader.SetContentFS(mockFS, ".")
		})

		It("should only process valid markdown files and ignore others", func() {
			// Expect exactly 2 calls for the valid files
			mockService.EXPECT().EnsureConsent(ctx, gomock.Any()).Return(nil).Times(2)

			err := loader.SeedConsents(ctx, logger, mockService)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("when service EnsureConsent fails", func() {
		BeforeEach(func() {
			mockFS := fstest.MapFS{
				"privacy.v1.md": &fstest.MapFile{
					Data: []byte("# Privacy Policy"),
				},
			}
			loader.SetContentFS(mockFS, ".")
		})

		It("should return an error", func() {
			serviceErr := errors.New("service error")
			mockService.EXPECT().EnsureConsent(ctx, gomock.Any()).Return(serviceErr).Times(1)

			err := loader.SeedConsents(ctx, logger, mockService)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to ensure consent privacy.v1.md exists"))
		})
	})

	Context("regex edge cases", func() {
		DescribeTable("filename matching",
			func(filename string, shouldMatch bool) {
				mockFS := fstest.MapFS{}
				if shouldMatch {
					mockFS[filename] = &fstest.MapFile{
						Data: []byte("test content"),
					}
				}
				loader.SetContentFS(mockFS, ".")

				if shouldMatch {
					mockService.EXPECT().EnsureConsent(ctx, gomock.Any()).Return(nil).Times(1)
				}
				// If shouldMatch is false, no expectations are set

				err := loader.SeedConsents(ctx, logger, mockService)
				Expect(err).ToNot(HaveOccurred())
			},
			Entry("valid simple name", "test.v1.md", true),
			Entry("valid name with dashes", "test-name.v1.md", true),
			Entry("valid name with underscores", "test_name.v1.md", true),
			Entry("valid name with numbers", "test123.v1.md", true),
			Entry("valid high version", "test.v999.md", true),
			Entry("invalid - no version", "test.md", false),
			Entry("invalid - wrong extension", "test.v1.txt", false),
			Entry("invalid - no extension", "test.v1", false),
			Entry("invalid - special characters in name", "test@name.v1.md", false),
			Entry("invalid - spaces in name", "test name.v1.md", false),
			Entry("valid - version zero", "test.v0.md", true),
			Entry("invalid - negative version", "test.v-1.md", false),
		)
	})
})

// Helper function to find consent by type
func findConsentByType(consents []*consent.Consent, typeName string) *consent.Consent {
	for _, c := range consents {
		if string(c.Type) == typeName {
			return c
		}
	}
	return nil
}
