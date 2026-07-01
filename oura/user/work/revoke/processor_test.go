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
	oauthWorkTest "github.com/tidepool-org/platform/oauth/work/test"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	ouraUserWorkRevoke "github.com/tidepool-org/platform/oura/user/work/revoke"
	ouraUserWorkRevokeTest "github.com/tidepool-org/platform/oura/user/work/revoke/test"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("processor", func() {
	It("FailingRetryDuration is expected", func() {
		Expect(ouraUserWorkRevoke.FailingRetryDuration).To(Equal(1 * time.Minute))
	})

	It("FailingRetryDurationJitter is expected", func() {
		Expect(ouraUserWorkRevoke.FailingRetryDurationJitter).To(Equal(5 * time.Second))
	})

	It("FailingRetryDurationMaximum is expected", func() {
		Expect(ouraUserWorkRevoke.FailingRetryDurationMaximum).To(Equal(24 * time.Hour))
	})

	Context("Metadata", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *ouraUserWorkRevoke.Metadata)) {
				datum := ouraUserWorkRevokeTest.RandomMetadata(test.AllowOptionals())
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, ouraUserWorkRevokeTest.NewObjectFromMetadata(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, ouraUserWorkRevokeTest.NewObjectFromMetadata(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *ouraUserWorkRevoke.Metadata) {},
			),
			Entry("empty",
				func(datum *ouraUserWorkRevoke.Metadata) {
					*datum = ouraUserWorkRevoke.Metadata{}
				},
			),
			Entry("all",
				func(datum *ouraUserWorkRevoke.Metadata) {
					datum.TokenMetadata = *oauthWorkTest.RandomTokenMetadata()
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *ouraUserWorkRevoke.Metadata), expectedErrors ...error) {
					expectedDatum := ouraUserWorkRevokeTest.RandomMetadata(test.AllowOptionals())
					object := ouraUserWorkRevokeTest.NewObjectFromMetadata(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					result := &ouraUserWorkRevoke.Metadata{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
					Expect(result).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *ouraUserWorkRevoke.Metadata) {},
				),
				Entry("empty",
					func(object map[string]any, expectedDatum *ouraUserWorkRevoke.Metadata) {
						clear(object)
						*expectedDatum = ouraUserWorkRevoke.Metadata{}
					},
				),
				Entry("multiple errors",
					func(object map[string]any, expectedDatum *ouraUserWorkRevoke.Metadata) {
						object["oauthToken"] = true
						expectedDatum.OAuthToken = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/oauthToken"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *ouraUserWorkRevoke.Metadata), expectedErrors ...error) {
					datum := ouraUserWorkRevokeTest.RandomMetadata(test.AllowOptionals())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *ouraUserWorkRevoke.Metadata) {},
				),
				Entry("oauth token missing",
					func(datum *ouraUserWorkRevoke.Metadata) {
						datum.OAuthToken = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/oauthToken"),
				),
				Entry("oauth token invalid",
					func(datum *ouraUserWorkRevoke.Metadata) {
						datum.OAuthToken.AccessToken = ""
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/oauthToken/accessToken"),
				),
				Entry("multiple errors",
					func(datum *ouraUserWorkRevoke.Metadata) {
						datum.OAuthToken = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/oauthToken"),
				),
			)
		})
	})

	Context("with dependencies", func() {
		var ctx context.Context
		var mockController *gomock.Controller
		var mockWorkClient *workTest.MockClient
		var mockOuraClient *ouraTest.MockClient
		var dependencies ouraUserWorkRevoke.Dependencies

		BeforeEach(func() {
			ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockWorkClient = workTest.NewMockClient(mockController)
			mockOuraClient = ouraTest.NewMockClient(mockController)
			dependencies = ouraUserWorkRevoke.Dependencies{
				Dependencies: workBase.Dependencies{
					WorkClient: mockWorkClient,
				},
				OuraClient: mockOuraClient,
			}
		})

		Context("NewProcessor", func() {
			It("returns an error if dependencies is invalid", func() {
				dependencies.WorkClient = nil
				processor, err := ouraUserWorkRevoke.NewProcessor(dependencies)
				Expect(err).To(MatchError("dependencies is invalid; work client is missing"))
				Expect(processor).To(BeNil())
			})

			It("returns successfully", func() {
				processor, err := ouraUserWorkRevoke.NewProcessor(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(processor).ToNot(BeNil())
			})

			Context("with processor", func() {
				var providerSessionID string
				var oauthToken *auth.OAuthToken
				var wrk *work.Work
				var mockProcessingUpdater *workTest.MockProcessingUpdater
				var processor *ouraUserWorkRevoke.Processor

				BeforeEach(func() {
					providerSessionID = authTest.RandomProviderSessionID()
					oauthToken = authTest.RandomToken()
					wrkCreate, err := ouraUserWorkRevoke.NewWorkCreate(providerSessionID, oauthToken)
					Expect(err).ToNot(HaveOccurred())
					Expect(wrkCreate).ToNot(BeNil())
					wrk = workTest.NewWorkFromCreateWithState(wrkCreate, work.StateProcessing)
					mockProcessingUpdater = workTest.NewMockProcessingUpdater(mockController)
					processor, err = ouraUserWorkRevoke.NewProcessor(dependencies)
					Expect(err).ToNot(HaveOccurred())
					Expect(processor).ToNot(BeNil())
				})

				Context("Process", func() {
					It("returns failing process result if unable to revoke oauth token", func() {
						testErr := errorsTest.RandomError()
						mockOuraClient.EXPECT().RevokeOAuthToken(gomock.Not(gomock.Nil()), oauthToken).Return(testErr)
						Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(errors.Wrap(testErr, "unable to revoke oauth token").Error())))
					})

					It("returns successful process result if able to revoke oauth token", func() {
						mockOuraClient.EXPECT().RevokeOAuthToken(gomock.Not(gomock.Nil()), oauthToken).Return(nil)
						Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchDeleteProcessResult())
					})
				})
			})
		})
	})
})
