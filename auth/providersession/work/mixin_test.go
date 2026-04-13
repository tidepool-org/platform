package work_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/auth"
	providerSessionTest "github.com/tidepool-org/platform/auth/providersession/test"
	providerSessionWork "github.com/tidepool-org/platform/auth/providersession/work"
	providerSessionWorkTest "github.com/tidepool-org/platform/auth/providersession/work/test"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("mixin", func() {
	Context("Metadata", func() {
		Context("MetadataKeyProviderSessionID", func() {
			It("returns expected value", func() {
				Expect(providerSessionWork.MetadataKeyProviderSessionID).To(Equal("providerSessionId"))
			})
		})

		Context("Metadata", func() {
			DescribeTable("serializes the datum as expected",
				func(mutator func(datum *providerSessionWork.Metadata)) {
					datum := providerSessionWorkTest.RandomMetadata(test.AllowOptional())
					mutator(datum)
					test.ExpectSerializedObjectJSON(datum, providerSessionWorkTest.NewObjectFromMetadata(datum, test.ObjectFormatJSON))
					test.ExpectSerializedObjectBSON(datum, providerSessionWorkTest.NewObjectFromMetadata(datum, test.ObjectFormatBSON))
				},
				Entry("succeeds",
					func(datum *providerSessionWork.Metadata) {},
				),
				Entry("empty",
					func(datum *providerSessionWork.Metadata) {
						*datum = providerSessionWork.Metadata{}
					},
				),
				Entry("all",
					func(datum *providerSessionWork.Metadata) {
						datum.ProviderSessionID = pointer.From(authTest.RandomProviderSessionID())
					},
				),
			)

			Context("Parse", func() {
				DescribeTable("parses the datum",
					func(mutator func(object map[string]any, expectedDatum *providerSessionWork.Metadata), expectedErrors ...error) {
						expectedDatum := providerSessionWorkTest.RandomMetadata(test.AllowOptional())
						object := providerSessionWorkTest.NewObjectFromMetadata(expectedDatum, test.ObjectFormatJSON)
						mutator(object, expectedDatum)
						result := &providerSessionWork.Metadata{}
						errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
						Expect(result).To(Equal(expectedDatum))
					},
					Entry("succeeds",
						func(object map[string]any, expectedDatum *providerSessionWork.Metadata) {},
					),
					Entry("empty",
						func(object map[string]any, expectedDatum *providerSessionWork.Metadata) {
							clear(object)
							*expectedDatum = providerSessionWork.Metadata{}
						},
					),
					Entry("multiple errors",
						func(object map[string]any, expectedDatum *providerSessionWork.Metadata) {
							object["providerSessionId"] = true
							expectedDatum.ProviderSessionID = nil
						},
						errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerSessionId"),
					),
				)
			})

			Context("Validate", func() {
				DescribeTable("validates the datum",
					func(mutator func(datum *providerSessionWork.Metadata), expectedErrors ...error) {
						datum := providerSessionWorkTest.RandomMetadata(test.AllowOptional())
						mutator(datum)
						errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
					},
					Entry("succeeds",
						func(datum *providerSessionWork.Metadata) {},
					),
					Entry("provider session id missing",
						func(datum *providerSessionWork.Metadata) {
							datum.ProviderSessionID = nil
						},
					),
					Entry("provider session id empty",
						func(datum *providerSessionWork.Metadata) {
							datum.ProviderSessionID = pointer.From("")
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerSessionId"),
					),
					Entry("provider session id invalid",
						func(datum *providerSessionWork.Metadata) {
							datum.ProviderSessionID = pointer.From("invalid")
						},
						errorsTest.WithPointerSource(auth.ErrorValueStringAsProviderSessionIDNotValid("invalid"), "/providerSessionId"),
					),
					Entry("provider session id valid",
						func(datum *providerSessionWork.Metadata) {
							datum.ProviderSessionID = pointer.From(authTest.RandomProviderSessionID())
						},
					),
					Entry("multiple errors",
						func(datum *providerSessionWork.Metadata) {
							datum.ProviderSessionID = pointer.From("")
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerSessionId"),
					),
				)
			})
		})
	})

	Context("with context and mocks", func() {
		var mockLogger *logTest.Logger
		var mockController *gomock.Controller
		var mockWorkProvider *workTest.Provider
		var mockProviderSessionClient *providerSessionTest.MockClient

		BeforeEach(func() {
			mockLogger = logTest.NewLogger()
			ctx := log.NewContextWithLogger(context.Background(), mockLogger)
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockWorkProvider = workTest.NewProvider(ctx)
			mockProviderSessionClient = providerSessionTest.NewMockClient(mockController)
		})

		Context("NewMixin", func() {
			It("returns an error when provider is missing", func() {
				mixin, err := providerSessionWork.NewMixin(nil, mockProviderSessionClient)
				Expect(err).To(MatchError("provider is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error when provider session client is missing", func() {
				mixin, err := providerSessionWork.NewMixin(mockWorkProvider, nil)
				Expect(err).To(MatchError("provider session client is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns successfully", func() {
				mixin, err := providerSessionWork.NewMixin(mockWorkProvider, mockProviderSessionClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})
		})

		Context("NewMixinFromWork", func() {
			var workMetadata *providerSessionWork.Metadata

			BeforeEach(func() {
				workMetadata = providerSessionWorkTest.RandomMetadata(test.AllowOptional())
			})

			It("returns an error when provider is missing", func() {
				mixin, err := providerSessionWork.NewMixinFromWork(nil, mockProviderSessionClient, workMetadata)
				Expect(err).To(MatchError("provider is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error when provider session client is missing", func() {
				mixin, err := providerSessionWork.NewMixinFromWork(mockWorkProvider, nil, workMetadata)
				Expect(err).To(MatchError("provider session client is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error when work metadata is missing", func() {
				mixin, err := providerSessionWork.NewMixinFromWork(mockWorkProvider, mockProviderSessionClient, nil)
				Expect(err).To(MatchError("work metadata is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns successfully", func() {
				mixin, err := providerSessionWork.NewMixinFromWork(mockWorkProvider, mockProviderSessionClient, workMetadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})
		})

		Context("mixin", func() {
			var workMetadata *providerSessionWork.Metadata
			var mixin providerSessionWork.MixinFromWork

			BeforeEach(func() {
				var err error
				workMetadata = providerSessionWorkTest.RandomMetadata(test.AllowOptional())
				mixin, err = providerSessionWork.NewMixinFromWork(mockWorkProvider, mockProviderSessionClient, workMetadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})

			Context("ProviderSessionClient", func() {
				It("returns the provider session client", func() {
					Expect(mixin.ProviderSessionClient()).To(Equal(mockProviderSessionClient))
				})
			})

			Context("HasProviderSession", func() {
				It("returns false initially", func() {
					Expect(mixin.HasProviderSession()).To(BeFalse())
				})

				It("returns true after SetProviderSession is called with a provider session", func() {
					Expect(mixin.SetProviderSession(authTest.RandomProviderSession(test.AllowOptional()))).To(BeNil())
					Expect(mixin.HasProviderSession()).To(BeTrue())
				})

				It("returns false after SetProviderSession is called with nil", func() {
					Expect(mixin.SetProviderSession(authTest.RandomProviderSession(test.AllowOptional()))).To(BeNil())
					Expect(mixin.HasProviderSession()).To(BeTrue())
					Expect(mixin.SetProviderSession(nil)).To(BeNil())
					Expect(mixin.HasProviderSession()).To(BeFalse())
				})
			})

			Context("ProviderSession", func() {
				It("returns nil initially", func() {
					Expect(mixin.ProviderSession()).To(BeNil())
				})

				It("returns the provider session after SetProviderSession is called with a provider session", func() {
					providerSession := authTest.RandomProviderSession(test.AllowOptional())
					Expect(mixin.SetProviderSession(providerSession)).To(BeNil())
					Expect(mixin.ProviderSession()).To(Equal(providerSession))
				})

				It("returns nil after SetProviderSession is called with nil", func() {
					Expect(mixin.SetProviderSession(authTest.RandomProviderSession(test.AllowOptional()))).To(BeNil())
					Expect(mixin.SetProviderSession(nil)).To(BeNil())
					Expect(mixin.ProviderSession()).To(BeNil())
				})
			})

			Context("SetProviderSession", func() {
				It("decodes metadata from provider session and returns nil", func() {
					providerSession := authTest.RandomProviderSession(test.AllowOptional())
					Expect(mixin.SetProviderSession(providerSession)).To(BeNil())
					Expect(mixin.ProviderSession()).To(Equal(providerSession))
				})

				It("clears metadata when provider session is nil and returns nil", func() {
					Expect(mixin.SetProviderSession(authTest.RandomProviderSession(test.AllowOptional()))).To(BeNil())
					Expect(mixin.SetProviderSession(nil)).To(BeNil())
					Expect(mixin.ProviderSession()).To(BeNil())
				})
			})

			Context("FetchProviderSession", func() {
				var providerSessionID string

				BeforeEach(func() {
					providerSessionID = authTest.RandomProviderSessionID()
				})

				It("returns failing process result when provider session client returns an error", func() {
					testErr := errorsTest.RandomError()
					mockProviderSessionClient.EXPECT().GetProviderSession(gomock.Not(gomock.Nil()), providerSessionID).Return(nil, testErr)
					Expect(mixin.FetchProviderSession(providerSessionID)).To(workTest.MatchFailingProcessResultError(MatchError(errors.Wrap(testErr, "unable to get provider session").Error())))
				})

				It("returns failed process result when provider session is nil", func() {
					mockProviderSessionClient.EXPECT().GetProviderSession(gomock.Not(gomock.Nil()), providerSessionID).Return(nil, nil)
					Expect(mixin.FetchProviderSession(providerSessionID)).To(workTest.MatchFailedProcessResultError(MatchError("provider session is missing")))
				})

				It("sets the provider session and returns nil on success", func() {
					providerSession := authTest.RandomProviderSession(test.AllowOptional())
					mockProviderSessionClient.EXPECT().GetProviderSession(gomock.Not(gomock.Nil()), providerSessionID).Return(providerSession, nil)
					Expect(mixin.FetchProviderSession(providerSessionID)).To(BeNil())
					Expect(mixin.ProviderSession()).To(Equal(providerSession))
				})
			})

			Context("UpdateProviderSession", func() {
				It("returns failed process result when provider session is missing", func() {
					Expect(mixin.UpdateProviderSession(&auth.ProviderSessionUpdate{})).To(workTest.MatchFailedProcessResultError(MatchError("provider session is missing")))
				})

				Context("with an existing provider session", func() {
					var providerSession *auth.ProviderSession
					var providerSessionUpdate *auth.ProviderSessionUpdate

					BeforeEach(func() {
						providerSession = authTest.RandomProviderSession(test.AllowOptional())
						Expect(mixin.SetProviderSession(providerSession)).To(BeNil())
						providerSessionUpdate = authTest.RandomProviderSessionUpdate(test.AllowOptional())
					})

					It("returns failing process result when the client returns an error", func() {
						testErr := errorsTest.RandomError()
						mockProviderSessionClient.EXPECT().UpdateProviderSession(gomock.Not(gomock.Nil()), providerSession.ID, providerSessionUpdate).Return(nil, testErr)
						Expect(mixin.UpdateProviderSession(providerSessionUpdate)).To(workTest.MatchFailingProcessResultError(MatchError(errors.Wrap(testErr, "unable to update provider session").Error())))
					})

					It("returns failed process result when the client returns a nil provider session", func() {
						mockProviderSessionClient.EXPECT().UpdateProviderSession(gomock.Not(gomock.Nil()), providerSession.ID, providerSessionUpdate).Return(nil, nil)
						Expect(mixin.UpdateProviderSession(providerSessionUpdate)).To(workTest.MatchFailedProcessResultError(MatchError("provider session is missing")))
					})

					It("updates the provider session and returns nil on success", func() {
						updatedDataSrc := authTest.RandomProviderSession(test.AllowOptional())
						mockProviderSessionClient.EXPECT().UpdateProviderSession(gomock.Not(gomock.Nil()), providerSession.ID, providerSessionUpdate).Return(updatedDataSrc, nil)
						Expect(mixin.UpdateProviderSession(providerSessionUpdate)).To(BeNil())
						Expect(mixin.ProviderSession()).To(Equal(updatedDataSrc))
					})
				})
			})

			Context("HasWorkMetadata", func() {
				It("returns false when work metadata is missing", func() {
					mixinWithoutMetadata, err := providerSessionWork.NewMixin(mockWorkProvider, mockProviderSessionClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(mixinWithoutMetadata).ToNot(BeNil())
					mixin = mixinWithoutMetadata.(providerSessionWork.MixinFromWork)
					Expect(mixin.HasWorkMetadata()).To(BeFalse())
				})

				It("returns true when work metadata is present", func() {
					Expect(mixin.HasWorkMetadata()).To(BeTrue())
				})
			})

			Context("FetchProviderSessionFromWorkMetadata", func() {
				It("returns failed process result when work metadata is missing", func() {
					mixinWithoutMetadata, err := providerSessionWork.NewMixin(mockWorkProvider, mockProviderSessionClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(mixinWithoutMetadata).ToNot(BeNil())
					mixin = mixinWithoutMetadata.(providerSessionWork.MixinFromWork)
					Expect(mixin.FetchProviderSessionFromWorkMetadata()).To(workTest.MatchFailedProcessResultError(MatchError("work metadata is missing")))
				})

				It("returns failed process result when work metadata provider session id is missing", func() {
					workMetadata.ProviderSessionID = nil
					Expect(mixin.FetchProviderSessionFromWorkMetadata()).To(workTest.MatchFailedProcessResultError(MatchError("work metadata provider session id is missing")))
				})

				It("returns failing process result when the client returns an error", func() {
					workMetadata.ProviderSessionID = pointer.From(authTest.RandomProviderSessionID())
					testErr := errorsTest.RandomError()
					mockProviderSessionClient.EXPECT().GetProviderSession(gomock.Not(gomock.Nil()), *workMetadata.ProviderSessionID).Return(nil, testErr)
					Expect(mixin.FetchProviderSessionFromWorkMetadata()).To(workTest.MatchFailingProcessResultError(MatchError(errors.Wrap(testErr, "unable to get provider session").Error())))
				})

				It("returns failed process result when the provider session is nil", func() {
					workMetadata.ProviderSessionID = pointer.From(authTest.RandomProviderSessionID())
					mockProviderSessionClient.EXPECT().GetProviderSession(gomock.Not(gomock.Nil()), *workMetadata.ProviderSessionID).Return(nil, nil)
					Expect(mixin.FetchProviderSessionFromWorkMetadata()).To(workTest.MatchFailedProcessResultError(MatchError("provider session is missing")))
				})

				It("sets the provider session and returns nil on success", func() {
					workMetadata.ProviderSessionID = pointer.From(authTest.RandomProviderSessionID())
					providerSession := authTest.RandomProviderSession(test.AllowOptional())
					mockProviderSessionClient.EXPECT().GetProviderSession(gomock.Not(gomock.Nil()), *workMetadata.ProviderSessionID).Return(providerSession, nil)
					Expect(mixin.FetchProviderSessionFromWorkMetadata()).To(BeNil())
					Expect(mixin.ProviderSession()).To(Equal(providerSession))
				})
			})

			Context("UpdateWorkMetadataFromProviderSession", func() {
				It("returns failed process result when provider session is missing", func() {
					Expect(mixin.UpdateWorkMetadataFromProviderSession()).To(workTest.MatchFailedProcessResultError(MatchError("provider session is missing")))
				})

				It("returns failed process result when work metadata is missing", func() {
					mixinWithoutMetadata, err := providerSessionWork.NewMixin(mockWorkProvider, mockProviderSessionClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(mixinWithoutMetadata).ToNot(BeNil())
					mixin = mixinWithoutMetadata.(providerSessionWork.MixinFromWork)
					providerSession := authTest.RandomProviderSession(test.AllowOptional())
					Expect(mixin.SetProviderSession(providerSession)).To(BeNil())
					Expect(mixin.UpdateWorkMetadataFromProviderSession()).To(workTest.MatchFailedProcessResultError(MatchError("work metadata is missing")))
				})

				It("updates work metadata with the provider session id and returns nil", func() {
					providerSession := authTest.RandomProviderSession(test.AllowOptional())
					Expect(mixin.SetProviderSession(providerSession)).To(BeNil())
					Expect(mixin.UpdateWorkMetadataFromProviderSession()).To(BeNil())
					Expect(workMetadata.ProviderSessionID).To(Equal(&providerSession.ID))
				})
			})

			Context("AddProviderSessionToContext", func() {
				It("adds nil fields to context", func() {
					mixin.AddProviderSessionToContext()
					Expect(mockWorkProvider.Fields).To(Equal(log.Fields{"providerSession": log.Fields(nil)}))
				})

				It("adds non-nil fields to context", func() {
					providerSession := authTest.RandomProviderSession(test.AllowOptional())
					Expect(mixin.SetProviderSession(providerSession)).To(BeNil())
					Expect(mockWorkProvider.Fields).To(Equal(log.Fields{
						"providerSession": log.Fields{
							"id":         providerSession.ID,
							"userId":     providerSession.UserID,
							"type":       providerSession.Type,
							"name":       providerSession.Name,
							"externalId": providerSession.ExternalID,
						},
					}))
				})
			})
		})
	})
})
