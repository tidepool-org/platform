package revoke_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
	oauthWork "github.com/tidepool-org/platform/oauth/work"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	ouraUserWorkRevoke "github.com/tidepool-org/platform/oura/user/work/revoke"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("factory", func() {
	It("Type is expected", func() {
		Expect(ouraUserWorkRevoke.Type).To(Equal("org.tidepool.oura.user.revoke"))
	})

	It("Quantity is expected", func() {
		Expect(ouraUserWorkRevoke.Quantity).To(Equal(1))
	})

	It("Frequency is expected", func() {
		Expect(ouraUserWorkRevoke.Frequency).To(Equal(5 * time.Second))
	})

	It("ProcessingTimeout is expected", func() {
		Expect(ouraUserWorkRevoke.ProcessingTimeout).To(Equal(3 * time.Minute))
	})

	Context("with dependencies", func() {
		var mockController *gomock.Controller
		var mockWorkClient *workTest.MockClient
		var mockOuraClient *ouraTest.MockClient
		var dependencies ouraUserWorkRevoke.Dependencies

		BeforeEach(func() {
			mockController = gomock.NewController(GinkgoT())
			mockWorkClient = workTest.NewMockClient(mockController)
			mockOuraClient = ouraTest.NewMockClient(mockController)
			dependencies = ouraUserWorkRevoke.Dependencies{
				Dependencies: workBase.Dependencies{
					WorkClient: mockWorkClient,
				},
				OuraClient: mockOuraClient,
			}
		})

		Context("Dependencies", func() {
			Context("Validate", func() {
				It("returns an error if work client is missing", func() {
					dependencies.WorkClient = nil
					Expect(dependencies.Validate()).To(MatchError("work client is missing"))
				})

				It("returns an error if oura client is missing", func() {
					dependencies.OuraClient = nil
					Expect(dependencies.Validate()).To(MatchError("oura client is missing"))
				})

				It("returns successfully", func() {
					Expect(dependencies.Validate()).To(Succeed())
				})
			})
		})

		Context("NewProcessorFactory", func() {
			It("returns an error if dependencies is invalid", func() {
				dependencies.WorkClient = nil
				processorFactory, err := ouraUserWorkRevoke.NewProcessorFactory(dependencies)
				Expect(err).To(MatchError("dependencies is invalid; work client is missing"))
				Expect(processorFactory).To(BeNil())
			})

			It("returns successfully", func() {
				processorFactory, err := ouraUserWorkRevoke.NewProcessorFactory(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(processorFactory).ToNot(BeNil())
			})

			Context("with processor factory", func() {
				var processorFactory *workBase.ProcessorFactory

				BeforeEach(func() {
					var err error
					processorFactory, err = ouraUserWorkRevoke.NewProcessorFactory(dependencies)
					Expect(err).ToNot(HaveOccurred())
					Expect(processorFactory).ToNot(BeNil())
				})

				Context("Type", func() {
					It("returns the expected type", func() {
						Expect(processorFactory.Type()).To(Equal(ouraUserWorkRevoke.Type))
					})
				})

				Context("Quantity", func() {
					It("returns the expected quantity", func() {
						Expect(processorFactory.Quantity()).To(Equal(ouraUserWorkRevoke.Quantity))
					})
				})

				Context("Frequency", func() {
					It("returns the expected frequency", func() {
						Expect(processorFactory.Frequency()).To(Equal(ouraUserWorkRevoke.Frequency))
					})
				})

				Context("New", func() {
					It("returns a new processor", func() {
						processor, err := processorFactory.New()
						Expect(err).ToNot(HaveOccurred())
						Expect(processor).ToNot(BeNil())
					})
				})
			})
		})
	})

	Context("NewWorkCreate", func() {
		var providerSessionID string
		var oauthToken *auth.OAuthToken

		BeforeEach(func() {
			providerSessionID = authTest.RandomProviderSessionID()
			oauthToken = authTest.RandomToken()
		})

		It("returns an error if provider session id is missing", func() {
			workCreate, err := ouraUserWorkRevoke.NewWorkCreate("", oauthToken)
			Expect(err).To(MatchError("provider session id is missing"))
			Expect(workCreate).To(BeNil())
		})

		It("returns an error if oauth token is missing", func() {
			workCreate, err := ouraUserWorkRevoke.NewWorkCreate(providerSessionID, nil)
			Expect(err).To(MatchError("oauth token is missing"))
			Expect(workCreate).To(BeNil())
		})

		It("returns successfully", func() {
			workCreate, err := ouraUserWorkRevoke.NewWorkCreate(providerSessionID, oauthToken)
			Expect(err).ToNot(HaveOccurred())
			Expect(workCreate).To(Equal(&work.Create{
				Type:              ouraUserWorkRevoke.Type,
				DeduplicationID:   pointer.From(providerSessionID),
				ProcessingTimeout: 180,
				Metadata: map[string]any{
					oauthWork.MetadataKeyOAuthToken: authTest.NewObjectFromToken(oauthToken, test.ObjectFormatJSON),
				},
			}))
		})
	})
})
