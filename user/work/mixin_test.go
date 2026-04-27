package work_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/user"
	userTest "github.com/tidepool-org/platform/user/test"
	userWork "github.com/tidepool-org/platform/user/work"
	userWorkTest "github.com/tidepool-org/platform/user/work/test"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("mixin", func() {
	Context("Metadata", func() {
		Context("MetadataKeyUserID", func() {
			It("returns expected value", func() {
				Expect(userWork.MetadataKeyUserID).To(Equal("userId"))
			})
		})

		Context("Metadata", func() {
			DescribeTable("serializes the datum as expected",
				func(mutator func(datum *userWork.Metadata)) {
					datum := userWorkTest.RandomMetadata(test.AllowOptionals())
					mutator(datum)
					test.ExpectSerializedObjectJSON(datum, userWorkTest.NewObjectFromMetadata(datum, test.ObjectFormatJSON))
					test.ExpectSerializedObjectBSON(datum, userWorkTest.NewObjectFromMetadata(datum, test.ObjectFormatBSON))
				},
				Entry("succeeds",
					func(datum *userWork.Metadata) {},
				),
				Entry("empty",
					func(datum *userWork.Metadata) {
						*datum = userWork.Metadata{}
					},
				),
				Entry("all",
					func(datum *userWork.Metadata) {
						datum.UserID = pointer.From(userTest.RandomUserID())
					},
				),
			)

			Context("Parse", func() {
				DescribeTable("parses the datum",
					func(mutator func(object map[string]any, expectedDatum *userWork.Metadata), expectedErrors ...error) {
						expectedDatum := userWorkTest.RandomMetadata(test.AllowOptionals())
						object := userWorkTest.NewObjectFromMetadata(expectedDatum, test.ObjectFormatJSON)
						mutator(object, expectedDatum)
						result := &userWork.Metadata{}
						errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
						Expect(result).To(Equal(expectedDatum))
					},
					Entry("succeeds",
						func(object map[string]any, expectedDatum *userWork.Metadata) {},
					),
					Entry("empty",
						func(object map[string]any, expectedDatum *userWork.Metadata) {
							clear(object)
							*expectedDatum = userWork.Metadata{}
						},
					),
					Entry("multiple errors",
						func(object map[string]any, expectedDatum *userWork.Metadata) {
							object["userId"] = true
							expectedDatum.UserID = nil
						},
						errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/userId"),
					),
				)
			})

			Context("Validate", func() {
				DescribeTable("validates the datum",
					func(mutator func(datum *userWork.Metadata), expectedErrors ...error) {
						datum := userWorkTest.RandomMetadata(test.AllowOptionals())
						mutator(datum)
						errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
					},
					Entry("succeeds",
						func(datum *userWork.Metadata) {},
					),
					Entry("user id missing",
						func(datum *userWork.Metadata) {
							datum.UserID = nil
						},
					),
					Entry("user id empty",
						func(datum *userWork.Metadata) {
							datum.UserID = pointer.From("")
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/userId"),
					),
					Entry("user id invalid",
						func(datum *userWork.Metadata) {
							datum.UserID = pointer.From("invalid")
						},
						errorsTest.WithPointerSource(user.ErrorValueStringAsIDNotValid("invalid"), "/userId"),
					),
					Entry("user id valid",
						func(datum *userWork.Metadata) {
							datum.UserID = pointer.From(userTest.RandomUserID())
						},
					),
					Entry("multiple errors",
						func(datum *userWork.Metadata) {
							datum.UserID = pointer.From("")
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/userId"),
					),
				)
			})
		})
	})

	Context("with context and mocks", func() {
		var mockLogger *logTest.Logger
		var mockController *gomock.Controller
		var mockWorkProvider *workTest.Provider
		var mockUserClient *userTest.MockClient

		BeforeEach(func() {
			mockLogger = logTest.NewLogger()
			ctx := log.NewContextWithLogger(context.Background(), mockLogger)
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockWorkProvider = workTest.NewProvider(ctx)
			mockUserClient = userTest.NewMockClient(mockController)
		})

		Context("NewMixin", func() {
			It("returns an error when provider is missing", func() {
				mixin, err := userWork.NewMixin(nil, mockUserClient)
				Expect(err).To(MatchError("provider is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error when user client is missing", func() {
				mixin, err := userWork.NewMixin(mockWorkProvider, nil)
				Expect(err).To(MatchError("user client is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns successfully", func() {
				mixin, err := userWork.NewMixin(mockWorkProvider, mockUserClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})
		})

		Context("NewMixinFromWork", func() {
			var workMetadata *userWork.Metadata

			BeforeEach(func() {
				workMetadata = userWorkTest.RandomMetadata(test.AllowOptionals())
			})

			It("returns an error when provider is missing", func() {
				mixin, err := userWork.NewMixinFromWork(nil, mockUserClient, workMetadata)
				Expect(err).To(MatchError("provider is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error when user client is missing", func() {
				mixin, err := userWork.NewMixinFromWork(mockWorkProvider, nil, workMetadata)
				Expect(err).To(MatchError("user client is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error when work metadata is missing", func() {
				mixin, err := userWork.NewMixinFromWork(mockWorkProvider, mockUserClient, nil)
				Expect(err).To(MatchError("work metadata is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns successfully", func() {
				mixin, err := userWork.NewMixinFromWork(mockWorkProvider, mockUserClient, workMetadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})
		})

		Context("mixin", func() {
			var workMetadata *userWork.Metadata
			var mixin userWork.MixinFromWork

			BeforeEach(func() {
				var err error
				workMetadata = userWorkTest.RandomMetadata(test.AllowOptionals())
				mixin, err = userWork.NewMixinFromWork(mockWorkProvider, mockUserClient, workMetadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})

			Context("UserClient", func() {
				It("returns the user client", func() {
					Expect(mixin.UserClient()).To(Equal(mockUserClient))
				})
			})

			Context("HasUser", func() {
				It("returns false initially", func() {
					Expect(mixin.HasUser()).To(BeFalse())
				})

				It("returns true after SetUser is called with a user", func() {
					Expect(mixin.SetUser(userTest.RandomUser(test.AllowOptionals()))).To(BeNil())
					Expect(mixin.HasUser()).To(BeTrue())
				})

				It("returns false after SetUser is called with nil", func() {
					Expect(mixin.SetUser(userTest.RandomUser(test.AllowOptionals()))).To(BeNil())
					Expect(mixin.HasUser()).To(BeTrue())
					Expect(mixin.SetUser(nil)).To(BeNil())
					Expect(mixin.HasUser()).To(BeFalse())
				})
			})

			Context("User", func() {
				It("returns nil initially", func() {
					Expect(mixin.User()).To(BeNil())
				})

				It("returns the user after SetUser is called with a user", func() {
					usr := userTest.RandomUser(test.AllowOptionals())
					Expect(mixin.SetUser(usr)).To(BeNil())
					Expect(mixin.User()).To(Equal(usr))
				})

				It("returns nil after SetUser is called with nil", func() {
					Expect(mixin.SetUser(userTest.RandomUser(test.AllowOptionals()))).To(BeNil())
					Expect(mixin.SetUser(nil)).To(BeNil())
					Expect(mixin.User()).To(BeNil())
				})
			})

			Context("SetUser", func() {
				It("decodes metadata from user and returns nil", func() {
					usr := userTest.RandomUser(test.AllowOptionals())
					Expect(mixin.SetUser(usr)).To(BeNil())
					Expect(mixin.User()).To(Equal(usr))
				})

				It("clears metadata when user is nil and returns nil", func() {
					Expect(mixin.SetUser(userTest.RandomUser(test.AllowOptionals()))).To(BeNil())
					Expect(mixin.SetUser(nil)).To(BeNil())
					Expect(mixin.User()).To(BeNil())
				})
			})

			Context("FetchUser", func() {
				var usrID string

				BeforeEach(func() {
					usrID = userTest.RandomUserID()
				})

				It("returns failing process result when user client returns an error", func() {
					testErr := errorsTest.RandomError()
					mockUserClient.EXPECT().Get(gomock.Not(gomock.Nil()), usrID).Return(nil, testErr)
					Expect(mixin.FetchUser(usrID)).To(workTest.MatchFailingProcessResultError(MatchError(errors.Wrap(testErr, "unable to get user").Error())))
				})

				It("returns failed process result when user is nil", func() {
					mockUserClient.EXPECT().Get(gomock.Not(gomock.Nil()), usrID).Return(nil, nil)
					Expect(mixin.FetchUser(usrID)).To(workTest.MatchFailedProcessResultError(MatchError("user is missing")))
				})

				It("sets the user and returns nil on success", func() {
					usr := userTest.RandomUser(test.AllowOptionals())
					mockUserClient.EXPECT().Get(gomock.Not(gomock.Nil()), usrID).Return(usr, nil)
					Expect(mixin.FetchUser(usrID)).To(BeNil())
					Expect(mixin.User()).To(Equal(usr))
				})
			})

			Context("HasWorkMetadata", func() {
				It("returns false when work metadata is missing", func() {
					mixinWithoutMetadata, err := userWork.NewMixin(mockWorkProvider, mockUserClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(mixinWithoutMetadata).ToNot(BeNil())
					var ok bool
					mixin, ok = mixinWithoutMetadata.(userWork.MixinFromWork)
					Expect(ok).To(BeTrue())
					Expect(mixin.HasWorkMetadata()).To(BeFalse())
				})

				It("returns true when work metadata is present", func() {
					Expect(mixin.HasWorkMetadata()).To(BeTrue())
				})
			})

			Context("FetchUserFromWorkMetadata", func() {
				It("returns failed process result when work metadata is missing", func() {
					mixinWithoutMetadata, err := userWork.NewMixin(mockWorkProvider, mockUserClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(mixinWithoutMetadata).ToNot(BeNil())
					var ok bool
					mixin, ok = mixinWithoutMetadata.(userWork.MixinFromWork)
					Expect(ok).To(BeTrue())
					Expect(mixin.FetchUserFromWorkMetadata()).To(workTest.MatchFailedProcessResultError(MatchError("work metadata is missing")))
				})

				It("returns failed process result when work metadata user id is missing", func() {
					workMetadata.UserID = nil
					Expect(mixin.FetchUserFromWorkMetadata()).To(workTest.MatchFailedProcessResultError(MatchError("work metadata user id is missing")))
				})

				It("returns failing process result when the client returns an error", func() {
					workMetadata.UserID = pointer.From(userTest.RandomUserID())
					testErr := errorsTest.RandomError()
					mockUserClient.EXPECT().Get(gomock.Not(gomock.Nil()), *workMetadata.UserID).Return(nil, testErr)
					Expect(mixin.FetchUserFromWorkMetadata()).To(workTest.MatchFailingProcessResultError(MatchError(errors.Wrap(testErr, "unable to get user").Error())))
				})

				It("returns failed process result when the user is nil", func() {
					workMetadata.UserID = pointer.From(userTest.RandomUserID())
					mockUserClient.EXPECT().Get(gomock.Not(gomock.Nil()), *workMetadata.UserID).Return(nil, nil)
					Expect(mixin.FetchUserFromWorkMetadata()).To(workTest.MatchFailedProcessResultError(MatchError("user is missing")))
				})

				It("sets the user and returns nil on success", func() {
					workMetadata.UserID = pointer.From(userTest.RandomUserID())
					usr := userTest.RandomUser(test.AllowOptionals())
					mockUserClient.EXPECT().Get(gomock.Not(gomock.Nil()), *workMetadata.UserID).Return(usr, nil)
					Expect(mixin.FetchUserFromWorkMetadata()).To(BeNil())
					Expect(mixin.User()).To(Equal(usr))
				})
			})

			Context("UpdateWorkMetadataFromUser", func() {
				It("returns failed process result when user is missing", func() {
					Expect(mixin.UpdateWorkMetadataFromUser()).To(workTest.MatchFailedProcessResultError(MatchError("user is missing")))
				})

				It("returns failed process result when user id is missing", func() {
					usr := userTest.RandomUser(test.AllowOptionals())
					usr.UserID = nil
					Expect(mixin.SetUser(usr)).To(BeNil())
					Expect(mixin.UpdateWorkMetadataFromUser()).To(workTest.MatchFailedProcessResultError(MatchError("user id is missing")))
				})

				It("returns failed process result when work metadata is missing", func() {
					mixinWithoutMetadata, err := userWork.NewMixin(mockWorkProvider, mockUserClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(mixinWithoutMetadata).ToNot(BeNil())
					var ok bool
					mixin, ok = mixinWithoutMetadata.(userWork.MixinFromWork)
					Expect(ok).To(BeTrue())
					user := userTest.RandomUser(test.AllowOptionals())
					Expect(mixin.SetUser(user)).To(BeNil())
					Expect(mixin.UpdateWorkMetadataFromUser()).To(workTest.MatchFailedProcessResultError(MatchError("work metadata is missing")))
				})

				It("updates work metadata with the user id and returns nil", func() {
					usr := userTest.RandomUser(test.AllowOptionals())
					Expect(mixin.SetUser(usr)).To(BeNil())
					Expect(mixin.UpdateWorkMetadataFromUser()).To(BeNil())
					Expect(workMetadata.UserID).To(Equal(usr.UserID))
				})
			})

			Context("AddUserToContext", func() {
				It("adds nil fields to context", func() {
					mixin.AddUserToContext()
					Expect(mockWorkProvider.Fields).To(Equal(log.Fields{"user": log.Fields(nil)}))
				})

				It("adds non-nil fields to context", func() {
					usr := userTest.RandomUser(test.AllowOptionals())
					Expect(mixin.SetUser(usr)).To(BeNil())
					Expect(mockWorkProvider.Fields).To(Equal(log.Fields{
						"user": log.Fields{
							"id": usr.UserID,
						},
					}))
				})
			})
		})
	})
})
