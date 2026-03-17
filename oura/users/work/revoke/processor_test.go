package revoke_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	ouraUsersWorkRevoke "github.com/tidepool-org/platform/oura/users/work/revoke"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("processor", func() {
	It("FailingRetryDuration is expected", func() {
		Expect(ouraUsersWorkRevoke.FailingRetryDuration).To(Equal(time.Minute))
	})

	It("FailingRetryDurationJitter is expected", func() {
		Expect(ouraUsersWorkRevoke.FailingRetryDurationJitter).To(Equal(5 * time.Second))
	})

	Context("with dependencies", func() {
		var ctx context.Context
		var mockController *gomock.Controller
		var mockWorkClient *workTest.MockClient
		var mockOuraClient *ouraTest.MockClient
		var dependencies ouraUsersWorkRevoke.Dependencies

		BeforeEach(func() {
			ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockWorkClient = workTest.NewMockClient(mockController)
			mockOuraClient = ouraTest.NewMockClient(mockController)
			dependencies = ouraUsersWorkRevoke.Dependencies{
				Dependencies: workBase.Dependencies{
					WorkClient: mockWorkClient,
				},
				OuraClient: mockOuraClient,
			}
		})

		Context("NewProcessor", func() {
			It("returns an error if dependencies is invalid", func() {
				dependencies.WorkClient = nil
				processor, err := ouraUsersWorkRevoke.NewProcessor(dependencies)
				Expect(err).To(MatchError("dependencies is invalid; work client is missing"))
				Expect(processor).To(BeNil())
			})

			It("returns successfully", func() {
				processor, err := ouraUsersWorkRevoke.NewProcessor(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(processor).ToNot(BeNil())
			})

			Context("with processor", func() {
				var providerSessionID string
				var oauthToken *auth.OAuthToken
				var wrk *work.Work
				var mockProcessingUpdater *workTest.MockProcessingUpdater
				var processor *ouraUsersWorkRevoke.Processor

				BeforeEach(func() {
					providerSessionID = authTest.RandomProviderSessionID()
					oauthToken = authTest.RandomToken()
					wrkCreate, err := ouraUsersWorkRevoke.NewWorkCreate(providerSessionID, oauthToken)
					Expect(err).ToNot(HaveOccurred())
					Expect(wrkCreate).ToNot(BeNil())
					wrk = workTest.NewWorkFromCreateWithState(wrkCreate, work.StateProcessing)
					mockProcessingUpdater = workTest.NewMockProcessingUpdater(mockController)
					processor, err = ouraUsersWorkRevoke.NewProcessor(dependencies)
					Expect(err).ToNot(HaveOccurred())
					Expect(processor).ToNot(BeNil())
				})

				Context("Process", func() {
					It("returns failing process result if unable to revoke oauth token", func() {
						testErr := errorsTest.RandomError()
						mockOuraClient.EXPECT().RevokeOAuthToken(gomock.Any(), oauthToken).Return(testErr)
						Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(errors.Wrap(testErr, "unable to revoke oauth token").Error())))
					})

					It("returns successful process result if able to revoke oauth token", func() {
						mockOuraClient.EXPECT().RevokeOAuthToken(gomock.Any(), oauthToken).Return(nil)
						Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchDeleteProcessResult())
					})
				})
			})
		})
	})
})
