package work_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	authTest "github.com/tidepool-org/platform/auth/test"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	dataSourceWork "github.com/tidepool-org/platform/data/source/work"
	dataSourceWorkTest "github.com/tidepool-org/platform/data/source/work/test"
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
		Context("MetadataKeyDataSourceID", func() {
			It("returns expected value", func() {
				Expect(dataSourceWork.MetadataKeyDataSourceID).To(Equal("dataSourceId"))
			})
		})

		Context("Metadata", func() {
			DescribeTable("serializes the datum as expected",
				func(mutator func(datum *dataSourceWork.Metadata)) {
					datum := dataSourceWorkTest.RandomMetadata(test.AllowOptional())
					mutator(datum)
					test.ExpectSerializedObjectJSON(datum, dataSourceWorkTest.NewObjectFromMetadata(datum, test.ObjectFormatJSON))
					test.ExpectSerializedObjectBSON(datum, dataSourceWorkTest.NewObjectFromMetadata(datum, test.ObjectFormatBSON))
				},
				Entry("succeeds",
					func(datum *dataSourceWork.Metadata) {},
				),
				Entry("empty",
					func(datum *dataSourceWork.Metadata) {
						*datum = dataSourceWork.Metadata{}
					},
				),
				Entry("all",
					func(datum *dataSourceWork.Metadata) {
						datum.DataSourceID = pointer.From(dataSourceTest.RandomDataSourceID())
					},
				),
			)

			Context("Parse", func() {
				DescribeTable("parses the datum",
					func(mutator func(object map[string]any, expectedDatum *dataSourceWork.Metadata), expectedErrors ...error) {
						expectedDatum := dataSourceWorkTest.RandomMetadata(test.AllowOptional())
						object := dataSourceWorkTest.NewObjectFromMetadata(expectedDatum, test.ObjectFormatJSON)
						mutator(object, expectedDatum)
						result := &dataSourceWork.Metadata{}
						errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
						Expect(result).To(Equal(expectedDatum))
					},
					Entry("succeeds",
						func(object map[string]any, expectedDatum *dataSourceWork.Metadata) {},
					),
					Entry("empty",
						func(object map[string]any, expectedDatum *dataSourceWork.Metadata) {
							clear(object)
							*expectedDatum = dataSourceWork.Metadata{}
						},
					),
					Entry("multiple errors",
						func(object map[string]any, expectedDatum *dataSourceWork.Metadata) {
							object["dataSourceId"] = true
							expectedDatum.DataSourceID = nil
						},
						errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/dataSourceId"),
					),
				)
			})

			Context("Validate", func() {
				DescribeTable("validates the datum",
					func(mutator func(datum *dataSourceWork.Metadata), expectedErrors ...error) {
						datum := dataSourceWorkTest.RandomMetadata(test.AllowOptional())
						mutator(datum)
						errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
					},
					Entry("succeeds",
						func(datum *dataSourceWork.Metadata) {},
					),
					Entry("data source id missing",
						func(datum *dataSourceWork.Metadata) {
							datum.DataSourceID = nil
						},
					),
					Entry("data source id empty",
						func(datum *dataSourceWork.Metadata) {
							datum.DataSourceID = pointer.From("")
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/dataSourceId"),
					),
					Entry("data source id invalid",
						func(datum *dataSourceWork.Metadata) {
							datum.DataSourceID = pointer.From("invalid")
						},
						errorsTest.WithPointerSource(dataSource.ErrorValueStringAsIDNotValid("invalid"), "/dataSourceId"),
					),
					Entry("data source id valid",
						func(datum *dataSourceWork.Metadata) {
							datum.DataSourceID = pointer.From(dataSourceTest.RandomDataSourceID())
						},
					),
					Entry("multiple errors",
						func(datum *dataSourceWork.Metadata) {
							datum.DataSourceID = pointer.From("")
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/dataSourceId"),
					),
				)
			})
		})
	})

	Context("with context and mocks", func() {
		var mockLogger *logTest.Logger
		var mockController *gomock.Controller
		var mockWorkProvider *workTest.Provider
		var mockDataSourceClient *dataSourceTest.MockClient

		BeforeEach(func() {
			mockLogger = logTest.NewLogger()
			ctx := log.NewContextWithLogger(context.Background(), mockLogger)
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockWorkProvider = workTest.NewProvider(ctx)
			mockDataSourceClient = dataSourceTest.NewMockClient(mockController)
		})

		Context("NewMixin", func() {
			It("returns an error when provider is missing", func() {
				mixin, err := dataSourceWork.NewMixin(nil, mockDataSourceClient)
				Expect(err).To(MatchError("provider is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error when data source client is missing", func() {
				mixin, err := dataSourceWork.NewMixin(mockWorkProvider, nil)
				Expect(err).To(MatchError("data source client is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns successfully", func() {
				mixin, err := dataSourceWork.NewMixin(mockWorkProvider, mockDataSourceClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})
		})

		Context("NewMixinWithParsedMetadata", func() {
			It("returns an error when provider is missing", func() {
				mixin, err := dataSourceWork.NewMixinWithParsedMetadata[workTest.MockMetadata](nil, mockDataSourceClient)
				Expect(err).To(MatchError("provider is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error when data source client is missing", func() {
				mixin, err := dataSourceWork.NewMixinWithParsedMetadata[workTest.MockMetadata](mockWorkProvider, nil)
				Expect(err).To(MatchError("data source client is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns successfully", func() {
				mixin, err := dataSourceWork.NewMixinWithParsedMetadata[workTest.MockMetadata](mockWorkProvider, mockDataSourceClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})
		})

		Context("NewMixinFromWork", func() {
			var workMetadata *dataSourceWork.Metadata

			BeforeEach(func() {
				workMetadata = dataSourceWorkTest.RandomMetadata(test.AllowOptional())
			})

			It("returns an error when provider is missing", func() {
				mixin, err := dataSourceWork.NewMixinFromWork(nil, mockDataSourceClient, workMetadata)
				Expect(err).To(MatchError("provider is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error when data source client is missing", func() {
				mixin, err := dataSourceWork.NewMixinFromWork(mockWorkProvider, nil, workMetadata)
				Expect(err).To(MatchError("data source client is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error when work metadata is missing", func() {
				mixin, err := dataSourceWork.NewMixinFromWork(mockWorkProvider, mockDataSourceClient, nil)
				Expect(err).To(MatchError("work metadata is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns successfully", func() {
				mixin, err := dataSourceWork.NewMixinFromWork(mockWorkProvider, mockDataSourceClient, workMetadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})
		})

		Context("NewMixinFromWorkWithParsedMetadata", func() {
			var workMetadata *dataSourceWork.Metadata

			BeforeEach(func() {
				workMetadata = dataSourceWorkTest.RandomMetadata(test.AllowOptional())
			})

			It("returns an error when provider is missing", func() {
				mixin, err := dataSourceWork.NewMixinFromWorkWithParsedMetadata[workTest.MockMetadata](nil, mockDataSourceClient, workMetadata)
				Expect(err).To(MatchError("provider is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error when data source client is missing", func() {
				mixin, err := dataSourceWork.NewMixinFromWorkWithParsedMetadata[workTest.MockMetadata](mockWorkProvider, nil, workMetadata)
				Expect(err).To(MatchError("data source client is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error when work metadata is missing", func() {
				mixin, err := dataSourceWork.NewMixinFromWorkWithParsedMetadata[workTest.MockMetadata](mockWorkProvider, mockDataSourceClient, nil)
				Expect(err).To(MatchError("work metadata is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns successfully", func() {
				mixin, err := dataSourceWork.NewMixinFromWorkWithParsedMetadata[workTest.MockMetadata](mockWorkProvider, mockDataSourceClient, workMetadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})
		})

		Context("mixin", func() {
			var workMetadata *dataSourceWork.Metadata
			var mixin dataSourceWork.MixinFromWorkWithParsedMetadata[workTest.MockMetadata]

			BeforeEach(func() {
				var err error
				workMetadata = dataSourceWorkTest.RandomMetadata(test.AllowOptional())
				mixin, err = dataSourceWork.NewMixinFromWorkWithParsedMetadata[workTest.MockMetadata](mockWorkProvider, mockDataSourceClient, workMetadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})

			Context("DataSourceClient", func() {
				It("returns the data source client", func() {
					Expect(mixin.DataSourceClient()).To(Equal(mockDataSourceClient))
				})
			})

			Context("HasDataSource", func() {
				It("returns false initially", func() {
					Expect(mixin.HasDataSource()).To(BeFalse())
				})

				It("returns true after SetDataSource is called with a data source", func() {
					Expect(mixin.SetDataSource(randomDataSourceWithMockMetadata())).To(BeNil())
					Expect(mixin.HasDataSource()).To(BeTrue())
				})

				It("returns false after SetDataSource is called with nil", func() {
					Expect(mixin.SetDataSource(randomDataSourceWithMockMetadata())).To(BeNil())
					Expect(mixin.HasDataSource()).To(BeTrue())
					Expect(mixin.SetDataSource(nil)).To(BeNil())
					Expect(mixin.HasDataSource()).To(BeFalse())
				})
			})

			Context("DataSource", func() {
				It("returns nil initially", func() {
					Expect(mixin.DataSource()).To(BeNil())
				})

				It("returns the data source after SetDataSource is called with a data source", func() {
					dataSrc := randomDataSourceWithMockMetadata()
					Expect(mixin.SetDataSource(dataSrc)).To(BeNil())
					Expect(mixin.DataSource()).To(Equal(dataSrc))
				})

				It("returns nil after SetDataSource is called with nil", func() {
					Expect(mixin.SetDataSource(randomDataSourceWithMockMetadata())).To(BeNil())
					Expect(mixin.SetDataSource(nil)).To(BeNil())
					Expect(mixin.DataSource()).To(BeNil())
				})
			})

			Context("SetDataSource", func() {
				It("returns failing process result when unable to decode metadata", func() {
					dataSrc := randomDataSourceWithMockMetadata()
					dataSrc.Metadata["mock"] = true
					Expect(mixin.SetDataSource(dataSrc)).To(workTest.MatchFailingProcessResultError(MatchError("unable to decode data source metadata; unable to decode metadata; type is not string, but bool")))
				})

				It("decodes metadata from data source and returns nil", func() {
					dataSrcMetadata := workTest.RandomMockMetadata()
					dataSrc := randomDataSourceWithMockMetadata()
					dataSrc.Metadata["mock"] = *dataSrcMetadata.Mock
					Expect(mixin.SetDataSource(dataSrc)).To(BeNil())
					Expect(mixin.DataSource()).To(Equal(dataSrc))
					Expect(mixin.DataSourceMetadata()).To(Equal(dataSrcMetadata))
				})

				It("clears metadata when data source is nil and returns nil", func() {
					Expect(mixin.SetDataSource(randomDataSourceWithMockMetadata())).To(BeNil())
					Expect(mixin.SetDataSource(nil)).To(BeNil())
					Expect(mixin.DataSource()).To(BeNil())
					Expect(mixin.DataSourceMetadata()).To(BeNil())
				})
			})

			Context("HasDataSourceMetadata", func() {
				It("returns false initially", func() {
					Expect(mixin.HasDataSourceMetadata()).To(BeFalse())
				})

				It("returns true after SetDataSourceMetadata is called with non-nil value", func() {
					dataSrcMetadata := workTest.RandomMockMetadata(test.AllowOptional())
					Expect(mixin.SetDataSourceMetadata(dataSrcMetadata)).To(BeNil())
					Expect(mixin.HasDataSourceMetadata()).To(BeTrue())
				})

				It("returns false after SetDataSourceMetadata is called with nil", func() {
					dataSrcMetadata := workTest.RandomMockMetadata(test.AllowOptional())
					Expect(mixin.SetDataSourceMetadata(dataSrcMetadata)).To(BeNil())
					Expect(mixin.HasDataSourceMetadata()).To(BeTrue())
					Expect(mixin.SetDataSourceMetadata(nil)).To(BeNil())
					Expect(mixin.HasDataSourceMetadata()).To(BeFalse())
				})
			})

			Context("DataSourceMetadata", func() {
				It("returns nil initially", func() {
					Expect(mixin.DataSourceMetadata()).To(BeNil())
				})

				It("returns the metadata after SetDataSourceMetadata is called", func() {
					dataSrcMetadata := workTest.RandomMockMetadata(test.AllowOptional())
					Expect(mixin.SetDataSourceMetadata(dataSrcMetadata)).To(BeNil())
					Expect(mixin.DataSourceMetadata()).To(Equal(dataSrcMetadata))
				})
			})

			Context("SetDataSourceMetadata", func() {
				It("sets the metadata and returns nil", func() {
					dataSrcMetadata := workTest.RandomMockMetadata(test.AllowOptional())
					Expect(mixin.SetDataSourceMetadata(dataSrcMetadata)).To(BeNil())
					Expect(mixin.DataSourceMetadata()).To(Equal(dataSrcMetadata))
				})

				It("sets the metadata to nil and returns nil", func() {
					dataSrcMetadata := workTest.RandomMockMetadata(test.AllowOptional())
					Expect(mixin.SetDataSourceMetadata(dataSrcMetadata)).To(BeNil())
					Expect(mixin.SetDataSourceMetadata(nil)).To(BeNil())
					Expect(mixin.DataSourceMetadata()).To(BeNil())
				})
			})

			Context("FetchDataSource", func() {
				var dataSrcID string

				BeforeEach(func() {
					dataSrcID = dataSourceTest.RandomDataSourceID()
				})

				It("returns failing process result when data source client returns an error", func() {
					testErr := errorsTest.RandomError()
					mockDataSourceClient.EXPECT().Get(gomock.Not(gomock.Nil()), dataSrcID).Return(nil, testErr)
					Expect(mixin.FetchDataSource(dataSrcID)).To(workTest.MatchFailingProcessResultError(MatchError(errors.Wrap(testErr, "unable to get data source").Error())))
				})

				It("returns failed process result when data source is nil", func() {
					mockDataSourceClient.EXPECT().Get(gomock.Not(gomock.Nil()), dataSrcID).Return(nil, nil)
					Expect(mixin.FetchDataSource(dataSrcID)).To(workTest.MatchFailedProcessResultError(MatchError("data source is missing")))
				})

				It("sets the data source and returns nil on success", func() {
					dataSrc := randomDataSourceWithMockMetadata()
					mockDataSourceClient.EXPECT().Get(gomock.Not(gomock.Nil()), dataSrcID).Return(dataSrc, nil)
					Expect(mixin.FetchDataSource(dataSrcID)).To(BeNil())
					Expect(mixin.DataSource()).To(Equal(dataSrc))
				})
			})

			Context("FetchDataSourceFromProviderSessionID", func() {
				var providerSessionID string

				BeforeEach(func() {
					providerSessionID = authTest.RandomProviderSessionID()
				})

				It("returns failing process result when data source client returns an error", func() {
					testErr := errorsTest.RandomError()
					mockDataSourceClient.EXPECT().GetFromProviderSession(gomock.Not(gomock.Nil()), providerSessionID).Return(nil, testErr)
					Expect(mixin.FetchDataSourceFromProviderSessionID(providerSessionID)).To(workTest.MatchFailingProcessResultError(MatchError(errors.Wrap(testErr, "unable to get data source from provider session").Error())))
				})

				It("returns failed process result when data source is nil", func() {
					mockDataSourceClient.EXPECT().GetFromProviderSession(gomock.Not(gomock.Nil()), providerSessionID).Return(nil, nil)
					Expect(mixin.FetchDataSourceFromProviderSessionID(providerSessionID)).To(workTest.MatchFailedProcessResultError(MatchError("data source is missing")))
				})

				It("sets the data source and returns nil on success", func() {
					dataSrc := randomDataSourceWithMockMetadata()
					mockDataSourceClient.EXPECT().GetFromProviderSession(gomock.Not(gomock.Nil()), providerSessionID).Return(dataSrc, nil)
					Expect(mixin.FetchDataSourceFromProviderSessionID(providerSessionID)).To(BeNil())
					Expect(mixin.DataSource()).To(Equal(dataSrc))
				})
			})

			Context("UpdateDataSource", func() {
				It("returns failed process result when data source is missing", func() {
					Expect(mixin.UpdateDataSource(&dataSource.Update{})).To(workTest.MatchFailedProcessResultError(MatchError("data source is missing")))
				})

				Context("with an existing data source", func() {
					var dataSrc *dataSource.Source
					var dataSrcUpdate *dataSource.Update
					var expectedDataSrcUpdate *dataSource.Update

					BeforeEach(func() {
						testString := test.RandomString()
						dataSrc = randomDataSourceWithMockMetadata()
						Expect(mixin.SetDataSource(dataSrc)).To(BeNil())
						mixin.DataSourceMetadata().Mock = pointer.From(testString)
						dataSrcUpdate = dataSourceTest.RandomUpdate(test.AllowOptional())
						dataSrcUpdate.Metadata = nil
						expectedDataSrcUpdate = dataSourceTest.CloneUpdate(dataSrcUpdate)
						expectedDataSrcUpdate.Metadata = &map[string]any{"mock": testString}
					})

					It("returns failing process result when the update metadata cannot be decoded", func() {
						mixin.DataSourceMetadata().Any = func() {}
						Expect(mixin.UpdateDataSource(dataSrcUpdate)).To(workTest.MatchFailingProcessResultError(MatchError("unable to encode data source metadata; unable to encode object; json: unsupported type: func()")))
					})

					It("returns failing process result when the client returns an error", func() {
						testErr := errorsTest.RandomError()
						mockDataSourceClient.EXPECT().Update(gomock.Not(gomock.Nil()), dataSrc.ID, nil, expectedDataSrcUpdate).Return(nil, testErr)
						Expect(mixin.UpdateDataSource(dataSrcUpdate)).To(workTest.MatchFailingProcessResultError(MatchError(errors.Wrap(testErr, "unable to update data source").Error())))
					})

					It("returns failed process result when the client returns a nil data source", func() {
						mockDataSourceClient.EXPECT().Update(gomock.Not(gomock.Nil()), dataSrc.ID, nil, expectedDataSrcUpdate).Return(nil, nil)
						Expect(mixin.UpdateDataSource(dataSrcUpdate)).To(workTest.MatchFailedProcessResultError(MatchError("data source is missing")))
					})

					It("updates the data source and returns nil on success", func() {
						updatedDataSrc := randomDataSourceWithMockMetadata()
						mockDataSourceClient.EXPECT().Update(gomock.Not(gomock.Nil()), dataSrc.ID, nil, expectedDataSrcUpdate).Return(updatedDataSrc, nil)
						Expect(mixin.UpdateDataSource(dataSrcUpdate)).To(BeNil())
						Expect(mixin.DataSource()).To(Equal(updatedDataSrc))
					})
				})
			})

			Context("UpdateDataSourceMetadata", func() {
				It("returns failed process result when data source is missing", func() {
					Expect(mixin.UpdateDataSourceMetadata()).To(workTest.MatchFailedProcessResultError(MatchError("data source is missing")))
				})

				It("updates the data source and returns nil on success", func() {
					testString := test.RandomString()
					dataSrc := randomDataSourceWithMockMetadata()
					Expect(mixin.SetDataSource(dataSrc)).To(BeNil())
					mixin.DataSourceMetadata().Mock = pointer.From(testString)
					expectedDataSrcUpdate := &dataSource.Update{}
					expectedDataSrcUpdate.Metadata = &map[string]any{"mock": testString}
					updatedDataSrc := randomDataSourceWithMockMetadata()
					mockDataSourceClient.EXPECT().Update(gomock.Not(gomock.Nil()), dataSrc.ID, nil, expectedDataSrcUpdate).Return(updatedDataSrc, nil)
					Expect(mixin.UpdateDataSourceMetadata()).To(BeNil())
					Expect(mixin.DataSource()).To(Equal(updatedDataSrc))
				})
			})

			Context("HasWorkMetadata", func() {
				It("returns false when work metadata is missing", func() {
					mixinWithoutMetadata, err := dataSourceWork.NewMixinWithParsedMetadata[workTest.MockMetadata](mockWorkProvider, mockDataSourceClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(mixinWithoutMetadata).ToNot(BeNil())
					mixin = mixinWithoutMetadata.(dataSourceWork.MixinFromWorkWithParsedMetadata[workTest.MockMetadata])
					Expect(mixin.HasWorkMetadata()).To(BeFalse())
				})

				It("returns true when work metadata is present", func() {
					Expect(mixin.HasWorkMetadata()).To(BeTrue())
				})
			})

			Context("FetchDataSourceFromWorkMetadata", func() {
				It("returns failed process result when work metadata is missing", func() {
					mixinWithoutMetadata, err := dataSourceWork.NewMixinWithParsedMetadata[workTest.MockMetadata](mockWorkProvider, mockDataSourceClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(mixinWithoutMetadata).ToNot(BeNil())
					mixin = mixinWithoutMetadata.(dataSourceWork.MixinFromWorkWithParsedMetadata[workTest.MockMetadata])
					Expect(mixin.FetchDataSourceFromWorkMetadata()).To(workTest.MatchFailedProcessResultError(MatchError("work metadata is missing")))
				})

				It("returns failed process result when work metadata data source id is missing", func() {
					workMetadata.DataSourceID = nil
					Expect(mixin.FetchDataSourceFromWorkMetadata()).To(workTest.MatchFailedProcessResultError(MatchError("work metadata data source id is missing")))
				})

				It("returns failing process result when the client returns an error", func() {
					workMetadata.DataSourceID = pointer.From(dataSourceTest.RandomDataSourceID())
					testErr := errorsTest.RandomError()
					mockDataSourceClient.EXPECT().Get(gomock.Not(gomock.Nil()), *workMetadata.DataSourceID).Return(nil, testErr)
					Expect(mixin.FetchDataSourceFromWorkMetadata()).To(workTest.MatchFailingProcessResultError(MatchError(errors.Wrap(testErr, "unable to get data source").Error())))
				})

				It("returns failed process result when the data source is nil", func() {
					workMetadata.DataSourceID = pointer.From(dataSourceTest.RandomDataSourceID())
					mockDataSourceClient.EXPECT().Get(gomock.Not(gomock.Nil()), *workMetadata.DataSourceID).Return(nil, nil)
					Expect(mixin.FetchDataSourceFromWorkMetadata()).To(workTest.MatchFailedProcessResultError(MatchError("data source is missing")))
				})

				It("sets the data source and returns nil on success", func() {
					workMetadata.DataSourceID = pointer.From(dataSourceTest.RandomDataSourceID())
					dataSrc := randomDataSourceWithMockMetadata()
					mockDataSourceClient.EXPECT().Get(gomock.Not(gomock.Nil()), *workMetadata.DataSourceID).Return(dataSrc, nil)
					Expect(mixin.FetchDataSourceFromWorkMetadata()).To(BeNil())
					Expect(mixin.DataSource()).To(Equal(dataSrc))
				})
			})

			Context("UpdateWorkMetadataFromDataSource", func() {
				It("returns failed process result when data source is missing", func() {
					Expect(mixin.UpdateWorkMetadataFromDataSource()).To(workTest.MatchFailedProcessResultError(MatchError("data source is missing")))
				})

				It("returns failed process result when work metadata is missing", func() {
					mixinWithoutMetadata, err := dataSourceWork.NewMixinWithParsedMetadata[workTest.MockMetadata](mockWorkProvider, mockDataSourceClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(mixinWithoutMetadata).ToNot(BeNil())
					mixin = mixinWithoutMetadata.(dataSourceWork.MixinFromWorkWithParsedMetadata[workTest.MockMetadata])
					dataRw := randomDataSourceWithMockMetadata()
					Expect(mixin.SetDataSource(dataRw)).To(BeNil())
					Expect(mixin.UpdateWorkMetadataFromDataSource()).To(workTest.MatchFailedProcessResultError(MatchError("work metadata is missing")))
				})

				It("updates work metadata with the data source id and returns nil", func() {
					dataSrc := randomDataSourceWithMockMetadata()
					Expect(mixin.SetDataSource(dataSrc)).To(BeNil())
					Expect(mixin.UpdateWorkMetadataFromDataSource()).To(BeNil())
					Expect(workMetadata.DataSourceID).To(Equal(&dataSrc.ID))
				})
			})

			Context("AddDataSourceToContext", func() {
				It("adds nil fields to context", func() {
					mixin.AddDataSourceToContext()
					Expect(mockWorkProvider.Fields).To(Equal(log.Fields{"dataSource": log.Fields(nil), "dataSourceMetadata": (*workTest.MockMetadata)(nil)}))
				})

				It("adds non-nil fields to context", func() {
					dataSrcMetadata := workTest.RandomMockMetadata(test.AllowOptional())
					dataSrc := randomDataSourceWithMockMetadata()
					dataSrc.Metadata = dataSrcMetadata.AsObject()
					Expect(mixin.SetDataSource(dataSrc)).To(BeNil())
					Expect(mockWorkProvider.Fields).To(Equal(log.Fields{
						"dataSource": log.Fields{
							"id":                 dataSrc.ID,
							"userId":             dataSrc.UserID,
							"providerType":       dataSrc.ProviderType,
							"providerName":       dataSrc.ProviderName,
							"providerExternalId": dataSrc.ProviderExternalID,
							"providerSessionId":  dataSrc.ProviderSessionID,
							"state":              dataSrc.State,
							"metadata":           dataSrc.Metadata,
							"dataSetId":          dataSrc.DataSetID,
						},
						"dataSourceMetadata": dataSrcMetadata,
					}))
				})
			})
		})
	})
})

func randomDataSourceWithMockMetadata() *dataSource.Source {
	dataSrc := dataSourceTest.RandomSource(test.AllowOptional())
	dataSrc.Metadata = workTest.RandomMockMetadata(test.AllowOptional()).AsObject()
	return dataSrc
}
