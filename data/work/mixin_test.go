package work_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	providerSessionTest "github.com/tidepool-org/platform/auth/providersession/test"
	providerSessionWorkTest "github.com/tidepool-org/platform/auth/providersession/work/test"
	authTest "github.com/tidepool-org/platform/auth/test"
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
					mockProviderSessionMixin.EXPECT().HasProviderSession().Return(false)
					Expect(mixin.FetchDataSourceFromProviderSession()).To(workTest.MatchFailedProcessResultError(MatchError("provider session is missing")))
				})

				It("returns successfully", func() {
					expectedResult := workTest.RandomSuccessProcessResult()
					providerSession := authTest.RandomProviderSession()
					mockProviderSessionMixin.EXPECT().HasProviderSession().Return(true)
					mockProviderSessionMixin.EXPECT().ProviderSession().Return(providerSession)
					mockDataSourceMixin.EXPECT().FetchDataSourceFromProviderSessionID(providerSession.ID).Return(expectedResult)
					Expect(mixin.FetchDataSourceFromProviderSession()).To(Equal(expectedResult))
				})
			})

			Describe("FetchProviderSessionFromDataSource", func() {
				It("returns failed result if data source is missing", func() {
					mockDataSourceMixin.EXPECT().HasDataSource().Return(false)
					Expect(mixin.FetchProviderSessionFromDataSource()).To(workTest.MatchFailedProcessResultError(MatchError("data source is missing")))
				})

				It("returns failed result if data source provider session id is missing", func() {
					dataSrc := dataSourceTest.RandomSource()
					dataSrc.ProviderSessionID = nil
					mockDataSourceMixin.EXPECT().HasDataSource().Return(true)
					mockDataSourceMixin.EXPECT().DataSource().Return(dataSrc)
					Expect(mixin.FetchProviderSessionFromDataSource()).To(workTest.MatchFailedProcessResultError(MatchError("data source provider session id is missing")))
				})

				It("returns successfully", func() {
					expectedResult := workTest.RandomSuccessProcessResult()
					dataSrc := dataSourceTest.RandomSource()
					dataSrc.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
					mockDataSourceMixin.EXPECT().HasDataSource().Return(true)
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
					mockDataSourceMixin.EXPECT().HasDataSource().Return(false)
					Expect(mixin.FetchDataSetFromDataSource()).To(workTest.MatchFailedProcessResultError(MatchError("data source is missing")))
				})

				It("returns failed result if data source data set id is missing", func() {
					dataSrc := dataSourceTest.RandomSource()
					dataSrc.DataSetID = nil
					mockDataSourceMixin.EXPECT().HasDataSource().Return(true)
					mockDataSourceMixin.EXPECT().DataSource().Return(dataSrc)
					Expect(mixin.FetchDataSetFromDataSource()).To(workTest.MatchFailedProcessResultError(MatchError("data source data set id is missing")))
				})

				It("returns successfully", func() {
					expectedResult := workTest.RandomSuccessProcessResult()
					dataSrc := dataSourceTest.RandomSource()
					dataSrc.DataSetID = pointer.FromString(dataTest.RandomDataSetID())
					mockDataSourceMixin.EXPECT().HasDataSource().Return(true)
					mockDataSourceMixin.EXPECT().DataSource().Return(dataSrc)
					mockDataSetMixin.EXPECT().FetchDataSet(*dataSrc.DataSetID).Return(expectedResult)
					Expect(mixin.FetchDataSetFromDataSource()).To(Equal(expectedResult))
				})
			})
		})
	})

	Context("DataSourceReplacerMixin", func() {
		var mockLogger *logTest.Logger
		var mockController *gomock.Controller
		var mockWorkProvider *workTest.Provider
		var mockDataSourceMixin *dataSourceWorkTest.MockMixinFromWork
		var mockProviderSessionClient *providerSessionTest.MockClient
		var mockDataSourceClient *dataSourceTest.MockClient

		BeforeEach(func() {
			var ctx context.Context
			mockLogger = logTest.NewLogger()
			mockController, ctx = gomock.WithContext(log.NewContextWithLogger(context.Background(), mockLogger), GinkgoT())
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
					replacementDataSource = dataSourceTest.RandomSource()
					replacementDataSource.State = dataSource.StateDisconnected
				})

				It("returns an error if replacement data source is missing", func() {
					replacementDataSource = nil
					Expect(mixin.ReplaceDataSource(replacementDataSource)).To(workTest.MatchFailedProcessResultError(MatchError("replacement data source is missing")))
				})

				assertWithWorkMetadata := func(inner func()) {
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

							inner()
						})
					})
				}

				assertWithOriginalDataSource := func() {
					Context("with original data source", func() {
						var originalDataSource *dataSource.Source

						BeforeEach(func() {
							originalDataSource = dataSourceTest.RandomSource()
							mockDataSourceMixin.EXPECT().DataSource().Return(originalDataSource)
							replacementDataSourceUpdate = &dataSource.Update{
								ProviderSessionID: originalDataSource.ProviderSessionID,
								State:             pointer.FromString(originalDataSource.State),
							}
						})

						It("returns result if unable to set replacement data source", func() {
							expectedResult := workTest.RandomFailedProcessResult()
							mockDataSourceMixin.EXPECT().SetDataSource(replacementDataSource).Return(expectedResult)
							Expect(mixin.ReplaceDataSource(replacementDataSource)).To(Equal(expectedResult))
						})

						Context("replacement data source set successfully", func() {
							BeforeEach(func() {
								mockDataSourceMixin.EXPECT().SetDataSource(replacementDataSource).Return(nil)
							})

							It("returns result if unable to update replacement data source", func() {
								expectedResult := workTest.RandomFailedProcessResult()
								mockDataSourceMixin.EXPECT().UpdateDataSource(replacementDataSourceUpdate).Return(expectedResult)
								Expect(mixin.ReplaceDataSource(replacementDataSource)).To(Equal(expectedResult))
							})

							Context("replacement data source updated successfully", func() {
								BeforeEach(func() {
									mockDataSourceMixin.EXPECT().UpdateDataSource(replacementDataSourceUpdate).Return(nil)
								})

								assertWithWorkMetadata(func() {
									It("logs warning if original data source cannot be deleted", func() {
										testErr := errorsTest.RandomError()
										mockDataSourceClient.EXPECT().Delete(gomock.Any(), originalDataSource.ID, nil).Return(false, testErr)
										Expect(mixin.ReplaceDataSource(replacementDataSource)).To(BeNil())
										mockLogger.AssertWarn("unable to delete existing data source", log.Fields{"dataSourceId": originalDataSource.ID, "error": errors.NewSerializable(testErr)})
									})

									Context("original data source deleted successfully", func() {
										BeforeEach(func() {
											mockDataSourceClient.EXPECT().Delete(gomock.Any(), originalDataSource.ID, nil).Return(true, nil)
										})

										It("returns successfully", func() {
											Expect(mixin.ReplaceDataSource(replacementDataSource)).To(BeNil())
										})
									})
								})
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
							replacementDataSource.ProviderSessionID = pointer.FromString(replacementProviderSessionID)
						})

						It("returns failed result if unable to delete provider session", func() {
							testErr := errorsTest.RandomError()
							mockProviderSessionClient.EXPECT().DeleteProviderSession(gomock.Any(), replacementProviderSessionID).Return(testErr)
							Expect(mixin.ReplaceDataSource(replacementDataSource)).To(workTest.MatchFailingProcessResultError(MatchError("unable to delete replacement data source provider session; " + testErr.Error())))
						})

						Context("replacement data source provider session deleted successfully", func() {
							BeforeEach(func() {
								mockProviderSessionClient.EXPECT().DeleteProviderSession(gomock.Any(), replacementProviderSessionID).Return(nil)
							})

							It("returns failed result if unable to get replacement data source after deleting provider session", func() {
								testErr := errorsTest.RandomError()
								mockDataSourceClient.EXPECT().Get(gomock.Any(), replacementDataSource.ID).Return(nil, testErr)
								Expect(mixin.ReplaceDataSource(replacementDataSource)).To(workTest.MatchFailingProcessResultError(MatchError("unable to get replacement data source after deleting provider session; " + testErr.Error())))
							})

							Context("replacement data source get successfully", func() {
								BeforeEach(func() {
									mockDataSourceClient.EXPECT().Get(gomock.Any(), replacementDataSource.ID).Return(replacementDataSource, nil)
								})

								Context("without original data source", func() {
									BeforeEach(func() {
										mockDataSourceMixin.EXPECT().DataSource().Return(nil)
									})

									It("returns result if unable to set replacement data source", func() {
										expectedResult := workTest.RandomFailedProcessResult()
										mockDataSourceMixin.EXPECT().SetDataSource(replacementDataSource).Return(expectedResult)
										Expect(mixin.ReplaceDataSource(replacementDataSource)).To(Equal(expectedResult))
									})

									Context("replacement data source set successfully", func() {
										BeforeEach(func() {
											mockDataSourceMixin.EXPECT().SetDataSource(replacementDataSource).Return(nil)
										})

										assertWithWorkMetadata(func() {
											It("returns successfully", func() {
												Expect(mixin.ReplaceDataSource(replacementDataSource)).To(BeNil())
											})
										})
									})
								})

								assertWithOriginalDataSource()
							})
						})
					})

					Context("without replacement data source provider session id", func() {
						BeforeEach(func() {
							replacementDataSource.ProviderSessionID = nil
							replacementDataSourceUpdate = &dataSource.Update{
								State: pointer.FromString(dataSource.StateDisconnected),
							}
						})

						AfterEach(func() {
							mockLogger.AssertWarn("replacement data source not disconnected and without provider session id", log.Fields{"replacementDataSourceId": replacementDataSource.ID})
						})

						Context("without original data source", func() {
							BeforeEach(func() {
								mockDataSourceMixin.EXPECT().DataSource().Return(nil)
							})

							It("returns result if unable to set replacement data source", func() {
								expectedResult := workTest.RandomFailedProcessResult()
								mockDataSourceMixin.EXPECT().SetDataSource(replacementDataSource).Return(expectedResult)
								Expect(mixin.ReplaceDataSource(replacementDataSource)).To(Equal(expectedResult))
							})

							Context("replacement data source set successfully", func() {
								BeforeEach(func() {
									mockDataSourceMixin.EXPECT().SetDataSource(replacementDataSource).Return(nil)
								})

								It("returns result if unable to update replacement data source", func() {
									expectedResult := workTest.RandomFailedProcessResult()
									mockDataSourceMixin.EXPECT().UpdateDataSource(replacementDataSourceUpdate).Return(expectedResult)
									Expect(mixin.ReplaceDataSource(replacementDataSource)).To(Equal(expectedResult))
								})

								Context("replacement data source updated successfully", func() {
									BeforeEach(func() {
										mockDataSourceMixin.EXPECT().UpdateDataSource(replacementDataSourceUpdate).Return(nil)
									})

									assertWithWorkMetadata(func() {
										It("returns successfully", func() {
											Expect(mixin.ReplaceDataSource(replacementDataSource)).To(BeNil())
										})
									})
								})
							})
						})

						assertWithOriginalDataSource()
					})
				})

				Context("with replacement data source state disconnected", func() {
					BeforeEach(func() {
						replacementDataSource.State = dataSource.StateDisconnected
						replacementDataSource.ProviderSessionID = nil
					})

					Context("without original data source", func() {
						BeforeEach(func() {
							mockDataSourceMixin.EXPECT().DataSource().Return(nil)
						})

						It("returns result if unable to set replacement data source", func() {
							expectedResult := workTest.RandomFailedProcessResult()
							mockDataSourceMixin.EXPECT().SetDataSource(replacementDataSource).Return(expectedResult)
							Expect(mixin.ReplaceDataSource(replacementDataSource)).To(Equal(expectedResult))
						})

						Context("replacement data source set successfully", func() {
							BeforeEach(func() {
								mockDataSourceMixin.EXPECT().SetDataSource(replacementDataSource).Return(nil)
							})

							assertWithWorkMetadata(func() {
								It("returns successfully", func() {
									Expect(mixin.ReplaceDataSource(replacementDataSource)).To(BeNil())
								})
							})
						})
					})

					assertWithOriginalDataSource()
				})
			})
		})
	})
})
