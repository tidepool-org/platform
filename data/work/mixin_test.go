package work_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	providerSessionTest "github.com/tidepool-org/platform/auth/providersession/test"
	providerSessionWorkTest "github.com/tidepool-org/platform/auth/providersession/work/test"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/data"
	dataRaw "github.com/tidepool-org/platform/data/raw"
	dataRawTest "github.com/tidepool-org/platform/data/raw/test"
	dataRawWorkTest "github.com/tidepool-org/platform/data/raw/work/test"
	dataSetWorkTest "github.com/tidepool-org/platform/data/set/work/test"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	dataSourceWorkTest "github.com/tidepool-org/platform/data/source/work/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/work"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("mixin", func() {
	Context("ProviderSessionDataSourceMixin", func() {
		var mockController *gomock.Controller
		var mockWorkProvider *workTest.Provider
		var mockProviderSessionMixin *providerSessionWorkTest.MockMixin
		var mockDataSourceMixin *dataSourceWorkTest.MockMixin

		BeforeEach(func() {
			ctx := log.NewContextWithLogger(context.Background(), logTest.NewLogger())
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockWorkProvider = workTest.NewProvider(ctx)
			mockProviderSessionMixin = providerSessionWorkTest.NewMockMixin(mockController)
			mockDataSourceMixin = dataSourceWorkTest.NewMockMixin(mockController)
		})

		Context("NewProviderSessionDataSourceMixin", func() {
			It("returns an error if provider is missing", func() {
				mixin, err := work.NewProviderSessionDataSourceMixin(nil, mockProviderSessionMixin, mockDataSourceMixin)
				Expect(err).To(MatchError("provider is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error if provider session mixin is missing", func() {
				mixin, err := work.NewProviderSessionDataSourceMixin(mockWorkProvider, nil, mockDataSourceMixin)
				Expect(err).To(MatchError("provider session mixin is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error if data source mixin is missing", func() {
				mixin, err := work.NewProviderSessionDataSourceMixin(mockWorkProvider, mockProviderSessionMixin, nil)
				Expect(err).To(MatchError("data source mixin is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns successfully", func() {
				mixin, err := work.NewProviderSessionDataSourceMixin(mockWorkProvider, mockProviderSessionMixin, mockDataSourceMixin)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})
		})

		Context("with a new mixin", func() {
			var mixin work.ProviderSessionDataSourceMixin

			BeforeEach(func() {
				var err error
				mixin, err = work.NewProviderSessionDataSourceMixin(mockWorkProvider, mockProviderSessionMixin, mockDataSourceMixin)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})

			Describe("FetchDataSourceFromProviderSession", func() {
				It("returns failed result if provider session is missing", func() {
					mockProviderSessionMixin.EXPECT().ProviderSession().Return(nil)
					Expect(mixin.FetchDataSourceFromProviderSession()).To(workTest.MatchFailedProcessResultError(MatchError("provider session is missing")))
				})

				It("returns successfully", func() {
					expectedResult := workTest.RandomSuccessProcessResult()
					providerSession := authTest.RandomProviderSession(test.AllowOptional())
					mockProviderSessionMixin.EXPECT().ProviderSession().Return(providerSession)
					mockDataSourceMixin.EXPECT().FetchDataSourceFromProviderSessionID(providerSession.ID).Return(expectedResult)
					Expect(mixin.FetchDataSourceFromProviderSession()).To(Equal(expectedResult))
				})
			})

			Describe("FetchProviderSessionFromDataSource", func() {
				It("returns failed result if data source is missing", func() {
					mockDataSourceMixin.EXPECT().DataSource().Return(nil)
					Expect(mixin.FetchProviderSessionFromDataSource()).To(workTest.MatchFailedProcessResultError(MatchError("data source is missing")))
				})

				It("returns failed result if data source provider session id is missing", func() {
					dataSrc := dataSourceTest.RandomSource(test.AllowOptional())
					dataSrc.ProviderSessionID = nil
					mockDataSourceMixin.EXPECT().DataSource().Return(dataSrc)
					Expect(mixin.FetchProviderSessionFromDataSource()).To(workTest.MatchFailedProcessResultError(MatchError("data source provider session id is missing")))
				})

				It("returns successfully", func() {
					expectedResult := workTest.RandomSuccessProcessResult()
					dataSrc := dataSourceTest.RandomSource(test.AllowOptional())
					dataSrc.ProviderSessionID = pointer.From(authTest.RandomProviderSessionID())
					mockDataSourceMixin.EXPECT().DataSource().Return(dataSrc)
					mockProviderSessionMixin.EXPECT().FetchProviderSession(*dataSrc.ProviderSessionID).Return(expectedResult)
					Expect(mixin.FetchProviderSessionFromDataSource()).To(Equal(expectedResult))
				})
			})
		})
	})

	Context("DataSourceDataSetMixin", func() {
		var mockController *gomock.Controller
		var mockWorkProvider *workTest.Provider
		var mockDataSourceMixin *dataSourceWorkTest.MockMixin
		var mockDataSetMixin *dataSetWorkTest.MockMixin

		BeforeEach(func() {
			ctx := log.NewContextWithLogger(context.Background(), logTest.NewLogger())
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockWorkProvider = workTest.NewProvider(ctx)
			mockDataSourceMixin = dataSourceWorkTest.NewMockMixin(mockController)
			mockDataSetMixin = dataSetWorkTest.NewMockMixin(mockController)
		})

		Context("NewDataSourceDataSetMixin", func() {
			It("returns an error if provider is missing", func() {
				mixin, err := work.NewDataSourceDataSetMixin(nil, mockDataSourceMixin, mockDataSetMixin)
				Expect(err).To(MatchError("provider is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error if data source mixin is missing", func() {
				mixin, err := work.NewDataSourceDataSetMixin(mockWorkProvider, nil, mockDataSetMixin)
				Expect(err).To(MatchError("data source mixin is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error if data set mixin is missing", func() {
				mixin, err := work.NewDataSourceDataSetMixin(mockWorkProvider, mockDataSourceMixin, nil)
				Expect(err).To(MatchError("data set mixin is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns successfully", func() {
				mixin, err := work.NewDataSourceDataSetMixin(mockWorkProvider, mockDataSourceMixin, mockDataSetMixin)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})
		})

		Context("with a new mixin", func() {
			var mixin work.DataSourceDataSetMixin

			BeforeEach(func() {
				var err error
				mixin, err = work.NewDataSourceDataSetMixin(mockWorkProvider, mockDataSourceMixin, mockDataSetMixin)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})

			Describe("FetchDataSetFromDataSource", func() {
				It("returns failed result if data source is missing", func() {
					mockDataSourceMixin.EXPECT().DataSource().Return(nil)
					Expect(mixin.FetchDataSetFromDataSource()).To(workTest.MatchFailedProcessResultError(MatchError("data source is missing")))
				})

				It("returns failed result if data source data set id is missing", func() {
					dataSrc := dataSourceTest.RandomSource(test.AllowOptional())
					dataSrc.DataSetID = nil
					mockDataSourceMixin.EXPECT().DataSource().Return(dataSrc)
					Expect(mixin.FetchDataSetFromDataSource()).To(workTest.MatchFailedProcessResultError(MatchError("data source data set id is missing")))
				})

				It("returns successfully", func() {
					expectedResult := workTest.RandomSuccessProcessResult()
					dataSrc := dataSourceTest.RandomSource(test.AllowOptional())
					dataSrc.DataSetID = pointer.From(dataTest.RandomDataSetID())
					mockDataSourceMixin.EXPECT().DataSource().Return(dataSrc)
					mockDataSetMixin.EXPECT().FetchDataSet(*dataSrc.DataSetID).Return(expectedResult)
					Expect(mixin.FetchDataSetFromDataSource()).To(Equal(expectedResult))
				})
			})

			Describe("CreateDataSetForDataSource", func() {
				var dataSetCreate *data.DataSetCreate

				BeforeEach(func() {
					dataSetCreate = dataTest.RandomDataSetCreate(test.AllowOptional())
				})

				It("returns failed result if data source is missing", func() {
					mockDataSourceMixin.EXPECT().DataSource().Return(nil)
					Expect(mixin.CreateDataSetForDataSource(dataSetCreate)).To(workTest.MatchFailedProcessResultError(MatchError("data source is missing")))
				})

				Context("with data source", func() {
					var dataSrc *dataSource.Source

					BeforeEach(func() {
						dataSrc = dataSourceTest.RandomSource(test.AllowOptional())
						dataSrc.DataSetID = nil
						mockDataSourceMixin.EXPECT().DataSource().Return(dataSrc)
					})

					It("returns failed result if data source data set id already exists", func() {
						dataSrc.DataSetID = pointer.From(dataTest.RandomDataSetID())
						Expect(mixin.CreateDataSetForDataSource(dataSetCreate)).To(workTest.MatchFailedProcessResultError(MatchError("data source data set id already exists")))
					})

					It("returns result if create data set returns a result", func() {
						expectedResult := workTest.RandomFailingProcessResult()
						mockDataSetMixin.EXPECT().CreateDataSet(dataSrc.UserID, dataSetCreate).Return(expectedResult)
						Expect(mixin.CreateDataSetForDataSource(dataSetCreate)).To(Equal(expectedResult))
					})

					It("returns successfully", func() {
						dataSt := dataTest.RandomDataSet(test.AllowOptional())
						expectedResult := workTest.RandomSuccessProcessResult()
						mockDataSetMixin.EXPECT().CreateDataSet(dataSrc.UserID, dataSetCreate).Return(nil)
						mockDataSetMixin.EXPECT().DataSet().Return(dataSt)
						mockDataSourceMixin.EXPECT().UpdateDataSource(&dataSource.Update{DataSetID: dataSt.ID}).Return(expectedResult)
						Expect(mixin.CreateDataSetForDataSource(dataSetCreate)).To(Equal(expectedResult))
					})
				})
			})
		})
	})

	Context("DataSourceDataRawMixin", func() {
		var mockController *gomock.Controller
		var mockWorkProvider *workTest.Provider
		var mockDataSourceMixin *dataSourceWorkTest.MockMixin
		var mockDataRawMixin *dataRawWorkTest.MockMixin

		BeforeEach(func() {
			ctx := log.NewContextWithLogger(context.Background(), logTest.NewLogger())
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockWorkProvider = workTest.NewProvider(ctx)
			mockDataSourceMixin = dataSourceWorkTest.NewMockMixin(mockController)
			mockDataRawMixin = dataRawWorkTest.NewMockMixin(mockController)
		})

		Context("NewDataSourceDataRawMixin", func() {
			It("returns an error if provider is missing", func() {
				mixin, err := work.NewDataSourceDataRawMixin(nil, mockDataSourceMixin, mockDataRawMixin)
				Expect(err).To(MatchError("provider is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error if data source mixin is missing", func() {
				mixin, err := work.NewDataSourceDataRawMixin(mockWorkProvider, nil, mockDataRawMixin)
				Expect(err).To(MatchError("data source mixin is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error if data raw mixin is missing", func() {
				mixin, err := work.NewDataSourceDataRawMixin(mockWorkProvider, mockDataSourceMixin, nil)
				Expect(err).To(MatchError("data raw mixin is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns successfully", func() {
				mixin, err := work.NewDataSourceDataRawMixin(mockWorkProvider, mockDataSourceMixin, mockDataRawMixin)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})
		})

		Context("with a new mixin", func() {
			var mixin work.DataSourceDataRawMixin

			BeforeEach(func() {
				var err error
				mixin, err = work.NewDataSourceDataRawMixin(mockWorkProvider, mockDataSourceMixin, mockDataRawMixin)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})

			Describe("CreateDataRawForDataSource", func() {
				var dataRawCreate *dataRaw.Create
				var reader *test.Reader

				BeforeEach(func() {
					dataRawCreate = dataRawTest.RandomCreate(test.AllowOptional())
					reader = test.NewReader()
				})

				It("returns failed result if data source is missing", func() {
					mockDataSourceMixin.EXPECT().DataSource().Return(nil)
					Expect(mixin.CreateDataRawForDataSource(dataRawCreate, reader)).To(workTest.MatchFailedProcessResultError(MatchError("data source is missing")))
				})

				Context("with data source", func() {
					var dataSrc *dataSource.Source

					BeforeEach(func() {
						dataSrc = dataSourceTest.RandomSource(test.AllowOptional())
						dataSrc.DataSetID = pointer.From(dataTest.RandomDataSetID())
						mockDataSourceMixin.EXPECT().DataSource().Return(dataSrc)
					})

					It("returns failed result if data source data set id is missing", func() {
						dataSrc.DataSetID = nil
						Expect(mixin.CreateDataRawForDataSource(dataRawCreate, reader)).To(workTest.MatchFailedProcessResultError(MatchError("data source data set id is missing")))
					})

					It("returns result if create data set returns a result", func() {
						expectedResult := workTest.RandomFailingProcessResult()
						mockDataRawMixin.EXPECT().CreateDataRaw(dataSrc.UserID, *dataSrc.DataSetID, dataRawCreate, reader).Return(expectedResult)
						Expect(mixin.CreateDataRawForDataSource(dataRawCreate, reader)).To(Equal(expectedResult))
					})

					It("returns successfully", func() {
						dataRw := dataRawTest.RandomRaw(test.AllowOptional())
						expectedResult := workTest.RandomSuccessProcessResult()
						mockDataRawMixin.EXPECT().CreateDataRaw(dataSrc.UserID, *dataSrc.DataSetID, dataRawCreate, reader).Return(nil)
						mockDataRawMixin.EXPECT().DataRaw().Return(dataRw)
						mockDataSourceMixin.EXPECT().UpdateDataSource(&dataSource.Update{LastImportTime: pointer.From(dataRw.CreatedTime)}).Return(expectedResult)
						Expect(mixin.CreateDataRawForDataSource(dataRawCreate, reader)).To(Equal(expectedResult))
					})
				})
			})
		})
	})

	Context("DataSourceReplacerMixin", func() {
		var ctx context.Context
		var mockLogger *logTest.Logger
		var mockController *gomock.Controller
		var mockWorkProvider *workTest.Provider
		var mockDataSourceMixin *dataSourceWorkTest.MockMixinFromWork
		var mockProviderSessionClient *providerSessionTest.MockClient
		var mockDataSourceClient *dataSourceTest.MockClient

		BeforeEach(func() {
			mockLogger = logTest.NewLogger()
			ctx = log.NewContextWithLogger(context.Background(), mockLogger)
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockWorkProvider = workTest.NewProvider(ctx)
			mockDataSourceMixin = dataSourceWorkTest.NewMockMixinFromWork(mockController)
			mockProviderSessionClient = providerSessionTest.NewMockClient(mockController)
			mockDataSourceClient = dataSourceTest.NewMockClient(mockController)
			mockDataSourceMixin.EXPECT().DataSourceClient().Return(mockDataSourceClient).AnyTimes()
		})

		Context("NewDataSourceReplacerMixin", func() {
			It("returns an error if provider is missing", func() {
				mixin, err := work.NewDataSourceReplacerMixin(nil, mockDataSourceMixin, mockProviderSessionClient)
				Expect(err).To(MatchError("provider is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error if data source mixin is missing", func() {
				mixin, err := work.NewDataSourceReplacerMixin(mockWorkProvider, nil, mockProviderSessionClient)
				Expect(err).To(MatchError("data source mixin is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error if provider session client is missing", func() {
				mixin, err := work.NewDataSourceReplacerMixin(mockWorkProvider, mockDataSourceMixin, nil)
				Expect(err).To(MatchError("provider session client is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns successfully", func() {
				mixin, err := work.NewDataSourceReplacerMixin(mockWorkProvider, mockDataSourceMixin, mockProviderSessionClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})
		})

		Context("with a new mixin", func() {
			var mixin work.DataSourceReplacerMixin

			JustBeforeEach(func() {
				var err error
				mixin, err = work.NewDataSourceReplacerMixin(mockWorkProvider, mockDataSourceMixin, mockProviderSessionClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})

			Describe("ReplaceDataSource", func() {
				var replacementDataSource *dataSource.Source
				var replacementDataSourceUpdate *dataSource.Update

				BeforeEach(func() {
					replacementDataSource = dataSourceTest.RandomSource(test.AllowOptional())
					replacementDataSource.State = dataSource.StateDisconnected
				})

				It("returns an error if replacement data source is missing", func() {
					replacementDataSource = nil
					Expect(mixin.ReplaceDataSource(replacementDataSource)).To(workTest.MatchFailedProcessResultError(MatchError("replacement data source is missing")))
				})

				assertWorkMetadata := func() {
					Context("without work metadata", func() {
						BeforeEach(func() {
							mockDataSourceMixin.EXPECT().HasWorkMetadata().Return(false)
						})

						It("returns successfully", func() {
							Expect(mixin.ReplaceDataSource(replacementDataSource)).To(BeNil())
						})
					})

					Context("with work metadata", func() {
						BeforeEach(func() {
							mockDataSourceMixin.EXPECT().HasWorkMetadata().Return(true)
						})

						It("returns result if unable to update work metadata", func() {
							expectedResult := workTest.RandomFailedProcessResult()
							mockDataSourceMixin.EXPECT().UpdateWorkMetadataFromDataSource().Return(expectedResult)
							Expect(mixin.ReplaceDataSource(replacementDataSource)).To(Equal(expectedResult))
						})

						Context("work metadata updated successfully", func() {
							BeforeEach(func() {
								mockDataSourceMixin.EXPECT().UpdateWorkMetadataFromDataSource().Return(nil)
							})

							It("returns successfully", func() {
								Expect(mixin.ReplaceDataSource(replacementDataSource)).To(BeNil())
							})
						})
					})
				}

				assertDataSourceUpdate := func() {
					Context("with data source update", func() {
						It("returns result if unable to update replacement data source", func() {
							expectedResult := workTest.RandomFailedProcessResult()
							mockDataSourceMixin.EXPECT().UpdateDataSource(replacementDataSourceUpdate).Return(expectedResult)
							Expect(mixin.ReplaceDataSource(replacementDataSource)).To(Equal(expectedResult))
						})

						Context("replacement data source updated successfully", func() {
							BeforeEach(func() {
								mockDataSourceMixin.EXPECT().UpdateDataSource(replacementDataSourceUpdate).Return(nil)
							})

							assertWorkMetadata()
						})
					})
				}

				assertDataSourceSet := func(inner func()) {
					Context("with replacement data source", func() {
						It("returns result if unable to set replacement data source", func() {
							expectedResult := workTest.RandomFailedProcessResult()
							mockDataSourceMixin.EXPECT().SetDataSource(replacementDataSource).Return(expectedResult)
							Expect(mixin.ReplaceDataSource(replacementDataSource)).To(Equal(expectedResult))
						})

						Context("replacement data source set successfully", func() {
							BeforeEach(func() {
								mockDataSourceMixin.EXPECT().SetDataSource(replacementDataSource).Return(nil)
							})

							inner()
						})
					})
				}

				assertWithoutOriginalDataSource := func(inner func()) {
					Context("without original data source", func() {
						BeforeEach(func() {
							mockDataSourceMixin.EXPECT().DataSource().Return(nil)
						})

						assertDataSourceSet(inner)
					})
				}

				assertWithOriginalDataSource := func() {
					Context("with original data source", func() {
						var originalDataSource *dataSource.Source

						BeforeEach(func() {
							originalDataSource = dataSourceTest.RandomSource(test.AllowOptional())
							mockDataSourceMixin.EXPECT().DataSource().Return(originalDataSource)
							replacementDataSourceUpdate = &dataSource.Update{
								ProviderSessionID: originalDataSource.ProviderSessionID,
								State:             pointer.From(originalDataSource.State),
							}
						})

						assertDataSourceSet(func() {
							Context("original data source delete returns error", func() {
								var testErr error

								BeforeEach(func() {
									testErr = errorsTest.RandomError()
									mockDataSourceClient.EXPECT().Delete(gomock.Not(gomock.Nil()), originalDataSource.ID, nil).Return(false, testErr)
								})

								AfterEach(func() {
									mockLogger.AssertWarn("unable to delete existing data source", log.Fields{"dataSourceId": originalDataSource.ID, "error": errors.NewSerializable(testErr)})
								})

								assertDataSourceUpdate()
							})

							Context("original data source deleted successfully", func() {
								BeforeEach(func() {
									mockDataSourceClient.EXPECT().Delete(gomock.Not(gomock.Nil()), originalDataSource.ID, nil).Return(true, nil)
								})

								assertDataSourceUpdate()
							})
						})
					})
				}

				Context("with replacement data source state connected or error", func() {
					BeforeEach(func() {
						replacementDataSource.State = test.RandomStringFromArray([]string{dataSource.StateConnected, dataSource.StateError})
					})

					Context("with replacement data source provider session id", func() {
						var replacementProviderSessionID string

						BeforeEach(func() {
							replacementProviderSessionID = authTest.RandomProviderSessionID()
							replacementDataSource.ProviderSessionID = pointer.From(replacementProviderSessionID)
						})

						It("returns failed result if unable to delete provider session", func() {
							testErr := errorsTest.RandomError()
							mockProviderSessionClient.EXPECT().DeleteProviderSession(gomock.Not(gomock.Nil()), replacementProviderSessionID).Return(testErr)
							Expect(mixin.ReplaceDataSource(replacementDataSource)).To(workTest.MatchFailingProcessResultError(MatchError("unable to delete replacement data source provider session; " + testErr.Error())))
						})

						Context("replacement data source provider session deleted successfully", func() {
							BeforeEach(func() {
								mockProviderSessionClient.EXPECT().DeleteProviderSession(gomock.Not(gomock.Nil()), replacementProviderSessionID).Return(nil)
							})

							It("returns failed result if unable to get replacement data source after deleting provider session", func() {
								testErr := errorsTest.RandomError()
								mockDataSourceClient.EXPECT().Get(gomock.Not(gomock.Nil()), replacementDataSource.ID).Return(nil, testErr)
								Expect(mixin.ReplaceDataSource(replacementDataSource)).To(workTest.MatchFailingProcessResultError(MatchError("unable to get replacement data source after deleting provider session; " + testErr.Error())))
							})

							Context("replacement data source get successfully", func() {
								BeforeEach(func() {
									mockDataSourceClient.EXPECT().Get(gomock.Not(gomock.Nil()), replacementDataSource.ID).Return(replacementDataSource, nil)
								})

								assertWithoutOriginalDataSource(assertWorkMetadata)
								assertWithOriginalDataSource()
							})
						})
					})

					Context("without replacement data source provider session id", func() {
						BeforeEach(func() {
							replacementDataSource.ProviderSessionID = nil
							replacementDataSourceUpdate = &dataSource.Update{
								State: pointer.From(dataSource.StateDisconnected),
							}
						})

						AfterEach(func() {
							mockLogger.AssertWarn("replacement data source not disconnected and without provider session id", log.Fields{"replacementDataSourceId": replacementDataSource.ID})
						})

						assertWithoutOriginalDataSource(assertDataSourceUpdate)
						assertWithOriginalDataSource()
					})
				})

				Context("with replacement data source state disconnected", func() {
					BeforeEach(func() {
						replacementDataSource.State = dataSource.StateDisconnected
						replacementDataSource.ProviderSessionID = nil
					})

					assertWithoutOriginalDataSource(assertWorkMetadata)
					assertWithOriginalDataSource()
				})
			})
		})
	})
})
