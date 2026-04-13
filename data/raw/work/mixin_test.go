package work_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	dataRaw "github.com/tidepool-org/platform/data/raw"
	dataRawTest "github.com/tidepool-org/platform/data/raw/test"
	dataRawWork "github.com/tidepool-org/platform/data/raw/work"
	dataRawWorkTest "github.com/tidepool-org/platform/data/raw/work/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("mixin", func() {
	Context("Metadata", func() {
		Context("MetadataKeyDataRawID", func() {
			It("returns expected value", func() {
				Expect(dataRawWork.MetadataKeyDataRawID).To(Equal("dataRawId"))
			})
		})

		Context("Metadata", func() {
			DescribeTable("serializes the datum as expected",
				func(mutator func(datum *dataRawWork.Metadata)) {
					datum := dataRawWorkTest.RandomMetadata(test.AllowOptional())
					mutator(datum)
					test.ExpectSerializedObjectJSON(datum, dataRawWorkTest.NewObjectFromMetadata(datum, test.ObjectFormatJSON))
					test.ExpectSerializedObjectBSON(datum, dataRawWorkTest.NewObjectFromMetadata(datum, test.ObjectFormatBSON))
				},
				Entry("succeeds",
					func(datum *dataRawWork.Metadata) {},
				),
				Entry("empty",
					func(datum *dataRawWork.Metadata) {
						*datum = dataRawWork.Metadata{}
					},
				),
				Entry("all",
					func(datum *dataRawWork.Metadata) {
						datum.DataRawID = pointer.From(dataRawTest.RandomDataRawID())
					},
				),
			)

			Context("Parse", func() {
				DescribeTable("parses the datum",
					func(mutator func(object map[string]any, expectedDatum *dataRawWork.Metadata), expectedErrors ...error) {
						expectedDatum := dataRawWorkTest.RandomMetadata(test.AllowOptional())
						object := dataRawWorkTest.NewObjectFromMetadata(expectedDatum, test.ObjectFormatJSON)
						mutator(object, expectedDatum)
						result := &dataRawWork.Metadata{}
						errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
						Expect(result).To(Equal(expectedDatum))
					},
					Entry("succeeds",
						func(object map[string]any, expectedDatum *dataRawWork.Metadata) {},
					),
					Entry("empty",
						func(object map[string]any, expectedDatum *dataRawWork.Metadata) {
							clear(object)
							*expectedDatum = dataRawWork.Metadata{}
						},
					),
					Entry("multiple errors",
						func(object map[string]any, expectedDatum *dataRawWork.Metadata) {
							object["dataRawId"] = true
							expectedDatum.DataRawID = nil
						},
						errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/dataRawId"),
					),
				)
			})

			Context("Validate", func() {
				DescribeTable("validates the datum",
					func(mutator func(datum *dataRawWork.Metadata), expectedErrors ...error) {
						datum := dataRawWorkTest.RandomMetadata(test.AllowOptional())
						mutator(datum)
						errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
					},
					Entry("succeeds",
						func(datum *dataRawWork.Metadata) {},
					),
					Entry("data raw id missing",
						func(datum *dataRawWork.Metadata) {
							datum.DataRawID = nil
						},
					),
					Entry("data raw id empty",
						func(datum *dataRawWork.Metadata) {
							datum.DataRawID = pointer.From("")
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/dataRawId"),
					),
					Entry("data raw id invalid",
						func(datum *dataRawWork.Metadata) {
							datum.DataRawID = pointer.From("invalid")
						},
						errorsTest.WithPointerSource(dataRaw.ErrorValueStringAsDataRawIDNotValid("invalid"), "/dataRawId"),
					),
					Entry("data raw id valid",
						func(datum *dataRawWork.Metadata) {
							datum.DataRawID = pointer.From(dataRawTest.RandomDataRawID())
						},
					),
					Entry("multiple errors",
						func(datum *dataRawWork.Metadata) {
							datum.DataRawID = pointer.From("")
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/dataRawId"),
					),
				)
			})
		})
	})

	Context("with context and mocks", func() {
		var mockLogger *logTest.Logger
		var mockController *gomock.Controller
		var mockWorkProvider *workTest.Provider
		var mockDataRawClient *dataRawTest.MockClient

		BeforeEach(func() {
			mockLogger = logTest.NewLogger()
			ctx := log.NewContextWithLogger(context.Background(), mockLogger)
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockWorkProvider = workTest.NewProvider(ctx)
			mockDataRawClient = dataRawTest.NewMockClient(mockController)
		})

		Context("NewMixin", func() {
			It("returns an error when provider is missing", func() {
				mixin, err := dataRawWork.NewMixin(nil, mockDataRawClient)
				Expect(err).To(MatchError("provider is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error when data raw client is missing", func() {
				mixin, err := dataRawWork.NewMixin(mockWorkProvider, nil)
				Expect(err).To(MatchError("data raw client is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns successfully", func() {
				mixin, err := dataRawWork.NewMixin(mockWorkProvider, mockDataRawClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})
		})

		Context("NewMixinWithParsedMetadata", func() {
			It("returns an error when provider is missing", func() {
				mixin, err := dataRawWork.NewMixinWithParsedMetadata[workTest.MockMetadata](nil, mockDataRawClient)
				Expect(err).To(MatchError("provider is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error when data raw client is missing", func() {
				mixin, err := dataRawWork.NewMixinWithParsedMetadata[workTest.MockMetadata](mockWorkProvider, nil)
				Expect(err).To(MatchError("data raw client is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns successfully", func() {
				mixin, err := dataRawWork.NewMixinWithParsedMetadata[workTest.MockMetadata](mockWorkProvider, mockDataRawClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})
		})

		Context("NewMixinFromWork", func() {
			var workMetadata *dataRawWork.Metadata

			BeforeEach(func() {
				workMetadata = dataRawWorkTest.RandomMetadata(test.AllowOptional())
			})

			It("returns an error when provider is missing", func() {
				mixin, err := dataRawWork.NewMixinFromWork(nil, mockDataRawClient, workMetadata)
				Expect(err).To(MatchError("provider is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error when data raw client is missing", func() {
				mixin, err := dataRawWork.NewMixinFromWork(mockWorkProvider, nil, workMetadata)
				Expect(err).To(MatchError("data raw client is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error when work metadata is missing", func() {
				mixin, err := dataRawWork.NewMixinFromWork(mockWorkProvider, mockDataRawClient, nil)
				Expect(err).To(MatchError("work metadata is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns successfully", func() {
				mixin, err := dataRawWork.NewMixinFromWork(mockWorkProvider, mockDataRawClient, workMetadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})
		})

		Context("NewMixinFromWorkWithParsedMetadata", func() {
			var workMetadata *dataRawWork.Metadata

			BeforeEach(func() {
				workMetadata = dataRawWorkTest.RandomMetadata(test.AllowOptional())
			})

			It("returns an error when provider is missing", func() {
				mixin, err := dataRawWork.NewMixinFromWorkWithParsedMetadata[workTest.MockMetadata](nil, mockDataRawClient, workMetadata)
				Expect(err).To(MatchError("provider is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error when data raw client is missing", func() {
				mixin, err := dataRawWork.NewMixinFromWorkWithParsedMetadata[workTest.MockMetadata](mockWorkProvider, nil, workMetadata)
				Expect(err).To(MatchError("data raw client is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error when work metadata is missing", func() {
				mixin, err := dataRawWork.NewMixinFromWorkWithParsedMetadata[workTest.MockMetadata](mockWorkProvider, mockDataRawClient, nil)
				Expect(err).To(MatchError("work metadata is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns successfully", func() {
				mixin, err := dataRawWork.NewMixinFromWorkWithParsedMetadata[workTest.MockMetadata](mockWorkProvider, mockDataRawClient, workMetadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})
		})

		Context("mixin", func() {
			var workMetadata *dataRawWork.Metadata
			var mixin dataRawWork.MixinFromWorkWithParsedMetadata[workTest.MockMetadata]

			BeforeEach(func() {
				var err error
				workMetadata = dataRawWorkTest.RandomMetadata(test.AllowOptional())
				mixin, err = dataRawWork.NewMixinFromWorkWithParsedMetadata[workTest.MockMetadata](mockWorkProvider, mockDataRawClient, workMetadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})

			Context("DataRawClient", func() {
				It("returns the data raw client", func() {
					Expect(mixin.DataRawClient()).To(Equal(mockDataRawClient))
				})
			})

			Context("HasDataRaw", func() {
				It("returns false initially", func() {
					Expect(mixin.HasDataRaw()).To(BeFalse())
				})

				It("returns true after SetDataRaw is called with a data raw", func() {
					Expect(mixin.SetDataRaw(randomDataRawWithMockMetadata())).To(BeNil())
					Expect(mixin.HasDataRaw()).To(BeTrue())
				})

				It("returns false after SetDataRaw is called with nil", func() {
					Expect(mixin.SetDataRaw(randomDataRawWithMockMetadata())).To(BeNil())
					Expect(mixin.HasDataRaw()).To(BeTrue())
					Expect(mixin.SetDataRaw(nil)).To(BeNil())
					Expect(mixin.HasDataRaw()).To(BeFalse())
				})
			})

			Context("DataRaw", func() {
				It("returns nil initially", func() {
					Expect(mixin.DataRaw()).To(BeNil())
				})

				It("returns the data raw after SetDataRaw is called with a data raw", func() {
					dataRw := randomDataRawWithMockMetadata()
					Expect(mixin.SetDataRaw(dataRw)).To(BeNil())
					Expect(mixin.DataRaw()).To(Equal(dataRw))
				})

				It("returns nil after SetDataRaw is called with nil", func() {
					Expect(mixin.SetDataRaw(randomDataRawWithMockMetadata())).To(BeNil())
					Expect(mixin.SetDataRaw(nil)).To(BeNil())
					Expect(mixin.DataRaw()).To(BeNil())
				})
			})

			Context("SetDataRaw", func() {
				It("returns failing process result when unable to decode metadata", func() {
					dataRw := randomDataRawWithMockMetadata()
					dataRw.Metadata["mock"] = true
					Expect(mixin.SetDataRaw(dataRw)).To(workTest.MatchFailingProcessResultError(MatchError("unable to decode data raw metadata; unable to decode metadata; type is not string, but bool")))
				})

				It("decodes metadata from data raw and returns nil", func() {
					dataRwMetadata := workTest.RandomMockMetadata()
					dataRw := randomDataRawWithMockMetadata()
					dataRw.Metadata["mock"] = *dataRwMetadata.Mock
					Expect(mixin.SetDataRaw(dataRw)).To(BeNil())
					Expect(mixin.DataRaw()).To(Equal(dataRw))
					Expect(mixin.DataRawMetadata()).To(Equal(dataRwMetadata))
				})

				It("clears metadata when data raw is nil and returns nil", func() {
					Expect(mixin.SetDataRaw(randomDataRawWithMockMetadata())).To(BeNil())
					Expect(mixin.SetDataRaw(nil)).To(BeNil())
					Expect(mixin.DataRaw()).To(BeNil())
					Expect(mixin.DataRawMetadata()).To(BeNil())
				})
			})

			Context("HasDataRawMetadata", func() {
				It("returns false initially", func() {
					Expect(mixin.HasDataRawMetadata()).To(BeFalse())
				})

				It("returns true after SetDataRawMetadata is called with non-nil value", func() {
					dataRwMetadata := workTest.RandomMockMetadata(test.AllowOptional())
					Expect(mixin.SetDataRawMetadata(dataRwMetadata)).To(BeNil())
					Expect(mixin.HasDataRawMetadata()).To(BeTrue())
				})

				It("returns false after SetDataRawMetadata is called with nil", func() {
					dataRwMetadata := workTest.RandomMockMetadata(test.AllowOptional())
					Expect(mixin.SetDataRawMetadata(dataRwMetadata)).To(BeNil())
					Expect(mixin.HasDataRawMetadata()).To(BeTrue())
					Expect(mixin.SetDataRawMetadata(nil)).To(BeNil())
					Expect(mixin.HasDataRawMetadata()).To(BeFalse())
				})
			})

			Context("DataRawMetadata", func() {
				It("returns nil initially", func() {
					Expect(mixin.DataRawMetadata()).To(BeNil())
				})

				It("returns the metadata after SetDataRawMetadata is called", func() {
					dataRwMetadata := workTest.RandomMockMetadata(test.AllowOptional())
					Expect(mixin.SetDataRawMetadata(dataRwMetadata)).To(BeNil())
					Expect(mixin.DataRawMetadata()).To(Equal(dataRwMetadata))
				})
			})

			Context("SetDataRawMetadata", func() {
				It("sets the metadata and returns nil", func() {
					dataRwMetadata := workTest.RandomMockMetadata(test.AllowOptional())
					Expect(mixin.SetDataRawMetadata(dataRwMetadata)).To(BeNil())
					Expect(mixin.DataRawMetadata()).To(Equal(dataRwMetadata))
				})

				It("sets the metadata to nil and returns nil", func() {
					dataRwMetadata := workTest.RandomMockMetadata(test.AllowOptional())
					Expect(mixin.SetDataRawMetadata(dataRwMetadata)).To(BeNil())
					Expect(mixin.SetDataRawMetadata(nil)).To(BeNil())
					Expect(mixin.DataRawMetadata()).To(BeNil())
				})
			})

			Context("FetchDataRaw", func() {
				var dataRwID string

				BeforeEach(func() {
					dataRwID = dataRawTest.RandomDataRawID()
				})

				It("returns failing process result when data raw client returns an error", func() {
					testErr := errorsTest.RandomError()
					mockDataRawClient.EXPECT().Get(gomock.Not(gomock.Nil()), dataRwID, nil).Return(nil, testErr)
					Expect(mixin.FetchDataRaw(dataRwID)).To(workTest.MatchFailingProcessResultError(MatchError(errors.Wrap(testErr, "unable to get data raw").Error())))
				})

				It("returns failed process result when data raw is nil", func() {
					mockDataRawClient.EXPECT().Get(gomock.Not(gomock.Nil()), dataRwID, nil).Return(nil, nil)
					Expect(mixin.FetchDataRaw(dataRwID)).To(workTest.MatchFailedProcessResultError(MatchError("data raw is missing")))
				})

				It("sets the data raw and returns nil on success", func() {
					dataRw := randomDataRawWithMockMetadata()
					mockDataRawClient.EXPECT().Get(gomock.Not(gomock.Nil()), dataRwID, nil).Return(dataRw, nil)
					Expect(mixin.FetchDataRaw(dataRwID)).To(BeNil())
					Expect(mixin.DataRaw()).To(Equal(dataRw))
				})
			})

			Context("CreateDataRaw", func() {
				var userID string
				var dataSetID string
				var dataRawCreate *dataRaw.Create
				var reader *test.Reader

				BeforeEach(func() {
					userID = userTest.RandomUserID()
					dataSetID = dataTest.RandomDataSetID()
					dataRawCreate = dataRawTest.RandomCreate(test.AllowOptional())
					reader = test.NewReader()
				})

				It("returns failed process result when data raw already exists", func() {
					Expect(mixin.SetDataRaw(randomDataRawWithMockMetadata())).To(BeNil())
					Expect(mixin.CreateDataRaw(userID, dataSetID, dataRawCreate, reader)).To(workTest.MatchFailedProcessResultError(MatchError("data raw already exists")))
				})

				Context("with an existing data raw", func() {
					It("returns failing process result when the client returns an error", func() {
						testErr := errorsTest.RandomError()
						mockDataRawClient.EXPECT().Create(gomock.Not(gomock.Nil()), userID, dataSetID, dataRawCreate, reader).Return(nil, testErr)
						Expect(mixin.CreateDataRaw(userID, dataSetID, dataRawCreate, reader)).To(workTest.MatchFailingProcessResultError(MatchError(errors.Wrap(testErr, "unable to create data raw").Error())))
					})

					It("returns failed process result when the client returns a nil data raw", func() {
						mockDataRawClient.EXPECT().Create(gomock.Not(gomock.Nil()), userID, dataSetID, dataRawCreate, reader).Return(nil, nil)
						Expect(mixin.CreateDataRaw(userID, dataSetID, dataRawCreate, reader)).To(workTest.MatchFailedProcessResultError(MatchError("data raw is missing")))
					})

					It("returns failed process result when the client returns a data raw with undecodable metadata", func() {
						createdDataRaw := randomDataRawWithMockMetadata()
						createdDataRaw.Metadata["invalid"] = true
						mockDataRawClient.EXPECT().Create(gomock.Not(gomock.Nil()), userID, dataSetID, dataRawCreate, reader).Return(createdDataRaw, nil)
						Expect(mixin.CreateDataRaw(userID, dataSetID, dataRawCreate, reader)).To(workTest.MatchFailingProcessResultError(MatchError(ContainSubstring("unable to decode data raw metadata"))))
					})

					It("updates the data raw and returns nil on success", func() {
						createdDataRaw := randomDataRawWithMockMetadata()
						mockDataRawClient.EXPECT().Create(gomock.Not(gomock.Nil()), userID, dataSetID, dataRawCreate, reader).Return(createdDataRaw, nil)
						Expect(mixin.CreateDataRaw(userID, dataSetID, dataRawCreate, reader)).To(BeNil())
						Expect(mixin.DataRaw()).To(Equal(createdDataRaw))
					})
				})
			})

			Context("UpdateDataRaw", func() {
				It("returns failed process result when data raw is missing", func() {
					Expect(mixin.UpdateDataRaw(&dataRaw.Update{})).To(workTest.MatchFailedProcessResultError(MatchError("data raw is missing")))
				})

				Context("with an existing data raw", func() {
					var dataRw *dataRaw.Raw
					var dataRwUpdate *dataRaw.Update
					var expectedDataRwUpdate *dataRaw.Update

					BeforeEach(func() {
						testString := test.RandomString()
						dataRw = randomDataRawWithMockMetadata()
						Expect(mixin.SetDataRaw(dataRw)).To(BeNil())
						mixin.DataRawMetadata().Mock = pointer.From(testString)
						dataRwUpdate = dataRawTest.RandomUpdate(test.AllowOptional())
						dataRwUpdate.Metadata = nil
						expectedDataRwUpdate = dataRawTest.CloneUpdate(dataRwUpdate)
						expectedDataRwUpdate.Metadata = &map[string]any{"mock": testString}
					})

					It("returns failing process result when the update metadata cannot be decoded", func() {
						mixin.DataRawMetadata().Any = func() {}
						Expect(mixin.UpdateDataRaw(dataRwUpdate)).To(workTest.MatchFailingProcessResultError(MatchError("unable to encode data raw metadata; unable to encode object; json: unsupported type: func()")))
					})

					It("returns failing process result when the client returns an error", func() {
						testErr := errorsTest.RandomError()
						mockDataRawClient.EXPECT().Update(gomock.Not(gomock.Nil()), dataRw.ID, nil, expectedDataRwUpdate).Return(nil, testErr)
						Expect(mixin.UpdateDataRaw(dataRwUpdate)).To(workTest.MatchFailingProcessResultError(MatchError(errors.Wrap(testErr, "unable to update data raw").Error())))
					})

					It("returns failed process result when the client returns a nil data raw", func() {
						mockDataRawClient.EXPECT().Update(gomock.Not(gomock.Nil()), dataRw.ID, nil, expectedDataRwUpdate).Return(nil, nil)
						Expect(mixin.UpdateDataRaw(dataRwUpdate)).To(workTest.MatchFailedProcessResultError(MatchError("data raw is missing")))
					})

					It("updates the data raw and returns nil on success", func() {
						updatedDataRw := randomDataRawWithMockMetadata()
						mockDataRawClient.EXPECT().Update(gomock.Not(gomock.Nil()), dataRw.ID, nil, expectedDataRwUpdate).Return(updatedDataRw, nil)
						Expect(mixin.UpdateDataRaw(dataRwUpdate)).To(BeNil())
						Expect(mixin.DataRaw()).To(Equal(updatedDataRw))
					})
				})
			})

			Context("UpdateDataRawMetadata", func() {
				It("returns failed process result when data raw is missing", func() {
					Expect(mixin.UpdateDataRawMetadata()).To(workTest.MatchFailedProcessResultError(MatchError("data raw is missing")))
				})

				It("updates the data raw and returns nil on success", func() {
					testString := test.RandomString()
					dataRw := randomDataRawWithMockMetadata()
					Expect(mixin.SetDataRaw(dataRw)).To(BeNil())
					mixin.DataRawMetadata().Mock = pointer.From(testString)
					expectedDataRwUpdate := &dataRaw.Update{}
					expectedDataRwUpdate.Metadata = &map[string]any{"mock": testString}
					updatedDataRw := randomDataRawWithMockMetadata()
					mockDataRawClient.EXPECT().Update(gomock.Not(gomock.Nil()), dataRw.ID, nil, expectedDataRwUpdate).Return(updatedDataRw, nil)
					Expect(mixin.UpdateDataRawMetadata()).To(BeNil())
					Expect(mixin.DataRaw()).To(Equal(updatedDataRw))
				})
			})

			Context("HasDataRawContent", func() {
				It("returns false initially", func() {
					Expect(mixin.HasDataRawContent()).To(BeFalse())
				})

				Context("with data raw content", func() {
					var dataRw *dataRaw.Raw
					var dataRwContent *dataRaw.Content

					BeforeEach(func() {
						dataRw = randomDataRawWithMockMetadata()
						dataRwContent = dataRawTest.RandomContent()
						mockDataRawClient.EXPECT().GetContent(gomock.Not(gomock.Nil()), dataRw.ID, nil).Return(dataRwContent, nil)
						Expect(mixin.SetDataRaw(dataRw)).To(BeNil())
						Expect(mixin.FetchDataRawContentFromDataRaw()).To(BeNil())
					})

					It("returns true after SetDataRaw is called with a data raw", func() {
						Expect(mixin.HasDataRawContent()).To(BeTrue())
					})

					It("returns false after SetDataRaw is called with nil", func() {
						Expect(mixin.HasDataRawContent()).To(BeTrue())
						Expect(mixin.SetDataRaw(nil)).To(BeNil())
						Expect(mixin.HasDataRawContent()).To(BeFalse())
					})
				})
			})

			Context("DataRawContent", func() {
				It("returns nil initially", func() {
					Expect(mixin.DataRawContent()).To(BeNil())
				})

				Context("with data raw content", func() {
					var dataRw *dataRaw.Raw
					var dataRwContent *dataRaw.Content

					BeforeEach(func() {
						dataRw = randomDataRawWithMockMetadata()
						dataRwContent = dataRawTest.RandomContent()
						mockDataRawClient.EXPECT().GetContent(gomock.Not(gomock.Nil()), dataRw.ID, nil).Return(dataRwContent, nil)
						Expect(mixin.SetDataRaw(dataRw)).To(BeNil())
						Expect(mixin.FetchDataRawContentFromDataRaw()).To(BeNil())
					})

					It("returns the data raw after SetDataRaw is called with a data raw", func() {
						Expect(mixin.DataRawContent()).To(Equal(dataRwContent))
					})

					It("returns false after SetDataRaw is called with nil", func() {
						Expect(mixin.DataRawContent()).To(Equal(dataRwContent))
						Expect(mixin.SetDataRaw(nil)).To(BeNil())
						Expect(mixin.DataRawContent()).To(BeNil())
					})
				})
			})

			Context("FetchDataRawContent", func() {
				var dataRwID string

				BeforeEach(func() {
					dataRwID = dataRawTest.RandomDataRawID()
				})

				It("returns failing process result when data raw client returns an error", func() {
					testErr := errorsTest.RandomError()
					mockDataRawClient.EXPECT().GetContent(gomock.Not(gomock.Nil()), dataRwID, nil).Return(nil, testErr)
					Expect(mixin.FetchDataRawContent(dataRwID)).To(workTest.MatchFailingProcessResultError(MatchError(errors.Wrap(testErr, "unable to get data raw content").Error())))
				})

				It("returns failed process result when data raw content is nil", func() {
					mockDataRawClient.EXPECT().GetContent(gomock.Not(gomock.Nil()), dataRwID, nil).Return(nil, nil)
					Expect(mixin.FetchDataRawContent(dataRwID)).To(workTest.MatchFailedProcessResultError(MatchError("data raw content is missing")))
				})

				It("sets the data raw content and returns nil on success", func() {
					dataRwContent := dataRawTest.RandomContent()
					mockDataRawClient.EXPECT().GetContent(gomock.Not(gomock.Nil()), dataRwID, nil).Return(dataRwContent, nil)
					Expect(mixin.FetchDataRawContent(dataRwID)).To(BeNil())
					Expect(mixin.DataRawContent()).To(Equal(dataRwContent))
				})
			})

			Context("FetchDataRawContentFromDataRaw", func() {
				It("returns failed process result when data raw is missing", func() {
					Expect(mixin.FetchDataRawContentFromDataRaw()).To(workTest.MatchFailedProcessResultError(MatchError("data raw is missing")))
				})

				Context("with data raw", func() {
					var dataRw *dataRaw.Raw

					BeforeEach(func() {
						dataRw = randomDataRawWithMockMetadata()
						Expect(mixin.SetDataRaw(dataRw)).To(BeNil())
					})

					It("returns failing process result when data raw client returns an error", func() {
						testErr := errorsTest.RandomError()
						mockDataRawClient.EXPECT().GetContent(gomock.Not(gomock.Nil()), dataRw.ID, nil).Return(nil, testErr)
						Expect(mixin.FetchDataRawContentFromDataRaw()).To(workTest.MatchFailingProcessResultError(MatchError(errors.Wrap(testErr, "unable to get data raw content").Error())))
					})

					It("returns failed process result when data raw content is nil", func() {
						mockDataRawClient.EXPECT().GetContent(gomock.Not(gomock.Nil()), dataRw.ID, nil).Return(nil, nil)
						Expect(mixin.FetchDataRawContentFromDataRaw()).To(workTest.MatchFailedProcessResultError(MatchError("data raw content is missing")))
					})

					It("sets the data raw content and returns nil on success", func() {
						dataRwContent := dataRawTest.RandomContent()
						mockDataRawClient.EXPECT().GetContent(gomock.Not(gomock.Nil()), dataRw.ID, nil).Return(dataRwContent, nil)
						Expect(mixin.FetchDataRawContentFromDataRaw()).To(BeNil())
						Expect(mixin.DataRawContent()).To(Equal(dataRwContent))
					})
				})
			})

			Context("CloseDataRawContent", func() {
				It("returns nil if there is no data raw content", func() {
					Expect(mixin.CloseDataRawContent()).To(BeNil())
				})

				Context("with data raw content", func() {
					var dataRw *dataRaw.Raw
					var dataRwContent *dataRaw.Content

					BeforeEach(func() {
						dataRw = randomDataRawWithMockMetadata()
						dataRwContent = dataRawTest.RandomContent()
						mockDataRawClient.EXPECT().GetContent(gomock.Not(gomock.Nil()), dataRw.ID, nil).Return(dataRwContent, nil)
						Expect(mixin.SetDataRaw(dataRw)).To(BeNil())
						Expect(mixin.FetchDataRawContentFromDataRaw()).To(BeNil())
					})

					It("returns nil if there is no data raw content read closer", func() {
						dataRwContent.ReadCloser = nil
						Expect(mixin.CloseDataRawContent()).To(BeNil())
					})

					It("logs a warning when the data raw content read closer returns an error", func() {
						testErr := errorsTest.RandomError()
						dataRwContent.ReadCloser = test.ErrorReadCloser(dataRwContent.ReadCloser, testErr)
						Expect(mixin.CloseDataRawContent()).To(BeNil())
						mockLogger.AssertWarn("unable to close data raw content", log.Fields{"error": errors.NewSerializable(testErr)})
					})

					It("returns nil on success", func() {
						Expect(mixin.CloseDataRawContent()).To(BeNil())
					})
				})
			})

			Context("HasWorkMetadata", func() {
				It("returns false when work metadata is missing", func() {
					mixinWithoutMetadata, err := dataRawWork.NewMixinWithParsedMetadata[workTest.MockMetadata](mockWorkProvider, mockDataRawClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(mixinWithoutMetadata).ToNot(BeNil())
					mixin, _ = mixinWithoutMetadata.(dataRawWork.MixinFromWorkWithParsedMetadata[workTest.MockMetadata])
					Expect(mixin.HasWorkMetadata()).To(BeFalse())
				})

				It("returns true when work metadata is present", func() {
					Expect(mixin.HasWorkMetadata()).To(BeTrue())
				})
			})

			Context("FetchDataRawFromWorkMetadata", func() {
				It("returns failed process result when work metadata is missing", func() {
					mixinWithoutMetadata, err := dataRawWork.NewMixinWithParsedMetadata[workTest.MockMetadata](mockWorkProvider, mockDataRawClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(mixinWithoutMetadata).ToNot(BeNil())
					mixin, _ = mixinWithoutMetadata.(dataRawWork.MixinFromWorkWithParsedMetadata[workTest.MockMetadata])
					Expect(mixin.FetchDataRawFromWorkMetadata()).To(workTest.MatchFailedProcessResultError(MatchError("work metadata is missing")))
				})

				It("returns failed process result when work metadata data raw id is missing", func() {
					workMetadata.DataRawID = nil
					Expect(mixin.FetchDataRawFromWorkMetadata()).To(workTest.MatchFailedProcessResultError(MatchError("work metadata data raw id is missing")))
				})

				It("returns failing process result when the client returns an error", func() {
					workMetadata.DataRawID = pointer.From(dataRawTest.RandomDataRawID())
					testErr := errorsTest.RandomError()
					mockDataRawClient.EXPECT().Get(gomock.Not(gomock.Nil()), *workMetadata.DataRawID, nil).Return(nil, testErr)
					Expect(mixin.FetchDataRawFromWorkMetadata()).To(workTest.MatchFailingProcessResultError(MatchError(errors.Wrap(testErr, "unable to get data raw").Error())))
				})

				It("returns failed process result when the data raw is nil", func() {
					workMetadata.DataRawID = pointer.From(dataRawTest.RandomDataRawID())
					mockDataRawClient.EXPECT().Get(gomock.Not(gomock.Nil()), *workMetadata.DataRawID, nil).Return(nil, nil)
					Expect(mixin.FetchDataRawFromWorkMetadata()).To(workTest.MatchFailedProcessResultError(MatchError("data raw is missing")))
				})

				It("sets the data raw and returns nil on success", func() {
					workMetadata.DataRawID = pointer.From(dataRawTest.RandomDataRawID())
					dataRw := randomDataRawWithMockMetadata()
					mockDataRawClient.EXPECT().Get(gomock.Not(gomock.Nil()), *workMetadata.DataRawID, nil).Return(dataRw, nil)
					Expect(mixin.FetchDataRawFromWorkMetadata()).To(BeNil())
					Expect(mixin.DataRaw()).To(Equal(dataRw))
				})
			})

			Context("UpdateWorkMetadataFromDataRaw", func() {
				It("returns failed process result when data raw is missing", func() {
					Expect(mixin.UpdateWorkMetadataFromDataRaw()).To(workTest.MatchFailedProcessResultError(MatchError("data raw is missing")))
				})

				It("returns failed process result when work metadata is missing", func() {
					mixinWithoutMetadata, err := dataRawWork.NewMixinWithParsedMetadata[workTest.MockMetadata](mockWorkProvider, mockDataRawClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(mixinWithoutMetadata).ToNot(BeNil())
					mixin, _ = mixinWithoutMetadata.(dataRawWork.MixinFromWorkWithParsedMetadata[workTest.MockMetadata])
					dataRw := randomDataRawWithMockMetadata()
					Expect(mixin.SetDataRaw(dataRw)).To(BeNil())
					Expect(mixin.UpdateWorkMetadataFromDataRaw()).To(workTest.MatchFailedProcessResultError(MatchError("work metadata is missing")))
				})

				It("updates work metadata with the data raw id and returns nil", func() {
					dataRw := randomDataRawWithMockMetadata()
					Expect(mixin.SetDataRaw(dataRw)).To(BeNil())
					Expect(mixin.UpdateWorkMetadataFromDataRaw()).To(BeNil())
					Expect(workMetadata.DataRawID).To(Equal(&dataRw.ID))
				})
			})

			Context("AddDataRawToContext", func() {
				It("adds nil fields to context", func() {
					mixin.AddDataRawToContext()
					Expect(mockWorkProvider.Fields).To(Equal(log.Fields{"dataRaw": log.Fields(nil), "dataRawMetadata": (*workTest.MockMetadata)(nil)}))
				})

				It("adds non-nil fields to context", func() {
					dataRwMetadata := workTest.RandomMockMetadata(test.AllowOptional())
					dataRw := randomDataRawWithMockMetadata()
					dataRw.Metadata = dataRwMetadata.AsObject()
					Expect(mixin.SetDataRaw(dataRw)).To(BeNil())
					Expect(mockWorkProvider.Fields).To(Equal(log.Fields{
						"dataRaw": log.Fields{
							"id":            dataRw.ID,
							"userId":        dataRw.UserID,
							"dataSetId":     dataRw.DataSetID,
							"metadata":      dataRw.Metadata,
							"mediaType":     dataRw.MediaType,
							"size":          dataRw.Size,
							"processedTime": dataRw.ProcessedTime,
						},
						"dataRawMetadata": dataRwMetadata,
					}))
				})
			})
		})
	})
})

func randomDataRawWithMockMetadata() *dataRaw.Raw {
	dataRw := dataRawTest.RandomRaw(test.AllowOptional())
	dataRw.Metadata = workTest.RandomMockMetadata(test.AllowOptional()).AsObject()
	return dataRw
}
