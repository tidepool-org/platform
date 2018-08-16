package deduplicator_test

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("Base", func() {
	// 	var testName string
	// 	var testVersion string

	// 	BeforeEach(func() {
	// 		testName = id.New()
	// 		testVersion = "1.2.3"
	// 	})

	// 	Context("BaseFactory", func() {
	// 		Context("NewBaseFactory", func() {
	// 			It("returns an error if the name is missing", func() {
	// 				testFactory, err := deduplicator.NewBaseFactory("", testVersion)
	// 				Expect(err).To(MatchError("name is missing"))
	// 				Expect(testFactory).To(BeNil())
	// 			})

	// 			It("returns an error if the version is missing", func() {
	// 				testFactory, err := deduplicator.NewBaseFactory(testName, "")
	// 				Expect(err).To(MatchError("version is missing"))
	// 				Expect(testFactory).To(BeNil())
	// 			})

	// 			It("returns an error if the version is invalid", func() {
	// 				testFactory, err := deduplicator.NewBaseFactory(testName, "x.y.z")
	// 				Expect(err).To(MatchError("version is invalid"))
	// 				Expect(testFactory).To(BeNil())
	// 			})

	// 			It("returns a new factory", func() {
	// 				testFactory, err := deduplicator.NewBaseFactory(testName, testVersion)
	// 				Expect(err).ToNot(HaveOccurred())
	// 				Expect(testFactory).ToNot(BeNil())
	// 				Expect(testFactory.Factory).ToNot(BeNil())
	// 			})
	// 		})

	// 		Context("with a new factory", func() {
	// 			var testFactory *deduplicator.BaseFactory
	// 			var testDataSet *upload.Upload

	// 			BeforeEach(func() {
	// 				var err error
	// 				testFactory, err = deduplicator.NewBaseFactory(testName, testVersion)
	// 				Expect(err).ToNot(HaveOccurred())
	// 				Expect(testFactory).ToNot(BeNil())
	// 				testDataSet = upload.New()
	// 				Expect(testDataSet).ToNot(BeNil())
	// 				testDataSet.UserID = id.New()
	// 			})

	// 			Context("CanDeduplicateDataSet", func() {
	// 				It("returns an error if the data set is missing", func() {
	// 					can, err := testFactory.CanDeduplicateDataSet(nil)
	// 					Expect(err).To(MatchError("data set is missing"))
	// 					Expect(can).To(BeFalse())
	// 				})

	// 				It("returns false if the data set id is missing", func() {
	// 					testDataSet.UploadID = ""
	// 					Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeFalse())
	// 				})

	// 				It("returns false if the data set user id is missing", func() {
	// 					testDataSet.UserID = ""
	// 					Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeFalse())
	// 				})

	// 				It("returns true if successful", func() {
	// 					Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeTrue())
	// 				})
	// 			})

	// 			Context("with logger and data store session", func() {
	// 				var testLogger log.Logger
	// 				var testDataSession *testDataStoreDEPRECATED.DataSession

	// 				BeforeEach(func() {
	// 					testLogger = null.NewLogger()
	// 					Expect(testLogger).ToNot(BeNil())
	// 					testDataSession = testDataStoreDEPRECATED.NewDataSession()
	// 					Expect(testDataSession).ToNot(BeNil())
	// 				})

	// 				AfterEach(func() {
	// 					testDataSession.Expectations()
	// 				})

	// 				Context("NewDeduplicatorForDataSet", func() {
	// 					It("returns an error if the logger is missing", func() {
	// 						testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(nil, testDataSession, testDataSet)
	// 						Expect(err).To(MatchError("logger is missing"))
	// 						Expect(testDeduplicator).To(BeNil())
	// 					})

	// 					It("returns an error if the data store session is missing", func() {
	// 						testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, nil, testDataSet)
	// 						Expect(err).To(MatchError("data store session is missing"))
	// 						Expect(testDeduplicator).To(BeNil())
	// 					})

	// 					It("returns an error if the data set is missing", func() {
	// 						testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, nil)
	// 						Expect(err).To(MatchError("data set is missing"))
	// 						Expect(testDeduplicator).To(BeNil())
	// 					})

	// 					It("returns an error if the data set id is missing", func() {
	// 						testDataSet.UploadID = ""
	// 						testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
	// 						Expect(err).To(MatchError("data set id is missing"))
	// 						Expect(testDeduplicator).To(BeNil())
	// 					})

	// 					It("returns an error if the data set user id is missing", func() {
	// 						testDataSet.UserID = ""
	// 						testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
	// 						Expect(err).To(MatchError("data set user id is missing"))
	// 						Expect(testDeduplicator).To(BeNil())
	// 					})

	// 					It("returns a new deduplicator upon success", func() {
	// 						Expect(testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)).ToNot(BeNil())
	// 					})
	// 				})
	// 			})

	// 			Context("with registered data set", func() {
	// 				BeforeEach(func() {
	// 					testDataSet.Deduplicator = &data.DeduplicatorDescriptor{Name: testName, Version: testVersion}
	// 				})

	// 				Context("IsRegisteredWithDataSet", func() {
	// 					It("returns an error if the data set is missing", func() {
	// 						can, err := testFactory.IsRegisteredWithDataSet(nil)
	// 						Expect(err).To(MatchError("data set is missing"))
	// 						Expect(can).To(BeFalse())
	// 					})

	// 					It("returns false if the data set id is missing", func() {
	// 						testDataSet.UploadID = ""
	// 						Expect(testFactory.IsRegisteredWithDataSet(testDataSet)).To(BeFalse())
	// 					})

	// 					It("returns false if the data set user id is missing", func() {
	// 						testDataSet.UserID = ""
	// 						Expect(testFactory.IsRegisteredWithDataSet(testDataSet)).To(BeFalse())
	// 					})

	// 					It("returns false if there is no deduplicator descriptor", func() {
	// 						testDataSet.Deduplicator = nil
	// 						Expect(testFactory.IsRegisteredWithDataSet(testDataSet)).To(BeFalse())
	// 					})

	// 					It("returns false if the deduplicator descriptor name is missing", func() {
	// 						testDataSet.Deduplicator.Name = ""
	// 						Expect(testFactory.IsRegisteredWithDataSet(testDataSet)).To(BeFalse())
	// 					})

	// 					It("returns false if the deduplicator descriptor name does not match", func() {
	// 						testDataSet.Deduplicator.Name = id.New()
	// 						Expect(testFactory.IsRegisteredWithDataSet(testDataSet)).To(BeFalse())
	// 					})

	// 					It("returns true if successful", func() {
	// 						Expect(testFactory.IsRegisteredWithDataSet(testDataSet)).To(BeTrue())
	// 					})
	// 				})

	// 				Context("with logger and data store session", func() {
	// 					var testLogger log.Logger
	// 					var testDataSession *testDataStoreDEPRECATED.DataSession

	// 					BeforeEach(func() {
	// 						testLogger = null.NewLogger()
	// 						Expect(testLogger).ToNot(BeNil())
	// 						testDataSession = testDataStoreDEPRECATED.NewDataSession()
	// 						Expect(testDataSession).ToNot(BeNil())
	// 					})

	// 					AfterEach(func() {
	// 						testDataSession.Expectations()
	// 					})

	// 					Context("NewRegisteredDeduplicatorForDataSet", func() {
	// 						It("returns an error if the logger is missing", func() {
	// 							testDeduplicator, err := testFactory.NewRegisteredDeduplicatorForDataSet(nil, testDataSession, testDataSet)
	// 							Expect(err).To(MatchError("logger is missing"))
	// 							Expect(testDeduplicator).To(BeNil())
	// 						})

	// 						It("returns an error if the data store session is missing", func() {
	// 							testDeduplicator, err := testFactory.NewRegisteredDeduplicatorForDataSet(testLogger, nil, testDataSet)
	// 							Expect(err).To(MatchError("data store session is missing"))
	// 							Expect(testDeduplicator).To(BeNil())
	// 						})

	// 						It("returns an error if the data set is missing", func() {
	// 							testDeduplicator, err := testFactory.NewRegisteredDeduplicatorForDataSet(testLogger, testDataSession, nil)
	// 							Expect(err).To(MatchError("data set is missing"))
	// 							Expect(testDeduplicator).To(BeNil())
	// 						})

	// 						It("returns an error if the data set id is missing", func() {
	// 							testDataSet.UploadID = ""
	// 							testDeduplicator, err := testFactory.NewRegisteredDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
	// 							Expect(err).To(MatchError("data set id is missing"))
	// 							Expect(testDeduplicator).To(BeNil())
	// 						})

	// 						It("returns an error if the data set user id is missing", func() {
	// 							testDataSet.UserID = ""
	// 							testDeduplicator, err := testFactory.NewRegisteredDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
	// 							Expect(err).To(MatchError("data set user id is missing"))
	// 							Expect(testDeduplicator).To(BeNil())
	// 						})

	// 						It("returns an error if there is no deduplicator descriptor", func() {
	// 							testDataSet.Deduplicator = nil
	// 							testDeduplicator, err := testFactory.NewRegisteredDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
	// 							Expect(err).To(MatchError("data set deduplicator descriptor is missing"))
	// 							Expect(testDeduplicator).To(BeNil())
	// 						})

	// 						It("returns an error if the deduplicator descriptor name is missing", func() {
	// 							testDataSet.Deduplicator.Name = ""
	// 							testDeduplicator, err := testFactory.NewRegisteredDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
	// 							Expect(err).To(MatchError("data set deduplicator descriptor is not registered with expected deduplicator"))
	// 							Expect(testDeduplicator).To(BeNil())
	// 						})

	// 						It("returns an error if the deduplicator descriptor name does not match", func() {
	// 							testDataSet.Deduplicator.Name = id.New()
	// 							testDeduplicator, err := testFactory.NewRegisteredDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
	// 							Expect(err).To(MatchError("data set deduplicator descriptor is not registered with expected deduplicator"))
	// 							Expect(testDeduplicator).To(BeNil())
	// 						})

	// 						It("returns a new deduplicator upon success", func() {
	// 							Expect(testFactory.NewRegisteredDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)).ToNot(BeNil())
	// 						})
	// 					})
	// 				})
	// 			})
	// 		})
	// 	})

	// 	Context("BaseDeduplicator", func() {
	// 		var testLogger log.Logger
	// 		var testDataSession *testDataStoreDEPRECATED.DataSession
	// 		var testDataSet *upload.Upload

	// 		BeforeEach(func() {
	// 			testLogger = null.NewLogger()
	// 			Expect(testLogger).ToNot(BeNil())
	// 			testDataSession = testDataStoreDEPRECATED.NewDataSession()
	// 			Expect(testDataSession).ToNot(BeNil())
	// 			testDataSet = upload.New()
	// 			Expect(testDataSet).ToNot(BeNil())
	// 			testDataSet.UserID = id.New()
	// 		})

	// 		AfterEach(func() {
	// 			testDataSession.Expectations()
	// 		})

	// 		Context("NewBaseDeduplicator", func() {
	// 			It("returns an error if the name is missing", func() {
	// 				testDeduplicator, err := deduplicator.NewBaseDeduplicator("", testVersion, testLogger, testDataSession, testDataSet)
	// 				Expect(err).To(MatchError("name is missing"))
	// 				Expect(testDeduplicator).To(BeNil())
	// 			})

	// 			It("returns an error if the version is missing", func() {
	// 				testDeduplicator, err := deduplicator.NewBaseDeduplicator(testName, "", testLogger, testDataSession, testDataSet)
	// 				Expect(err).To(MatchError("version is missing"))
	// 				Expect(testDeduplicator).To(BeNil())
	// 			})

	// 			It("returns an error if the version is invalid", func() {
	// 				testDeduplicator, err := deduplicator.NewBaseDeduplicator(testName, "x.y.z", testLogger, testDataSession, testDataSet)
	// 				Expect(err).To(MatchError("version is invalid"))
	// 				Expect(testDeduplicator).To(BeNil())
	// 			})

	// 			It("returns an error if the logger is missing", func() {
	// 				testDeduplicator, err := deduplicator.NewBaseDeduplicator(testName, testVersion, nil, testDataSession, testDataSet)
	// 				Expect(err).To(MatchError("logger is missing"))
	// 				Expect(testDeduplicator).To(BeNil())
	// 			})

	// 			It("returns an error if the data store session is missing", func() {
	// 				testDeduplicator, err := deduplicator.NewBaseDeduplicator(testName, testVersion, testLogger, nil, testDataSet)
	// 				Expect(err).To(MatchError("data store session is missing"))
	// 				Expect(testDeduplicator).To(BeNil())
	// 			})

	// 			It("returns an error if the data set is missing", func() {
	// 				testDeduplicator, err := deduplicator.NewBaseDeduplicator(testName, testVersion, testLogger, testDataSession, nil)
	// 				Expect(err).To(MatchError("data set is missing"))
	// 				Expect(testDeduplicator).To(BeNil())
	// 			})

	// 			It("returns an error if the data set id is missing", func() {
	// 				testDataSet.UploadID = ""
	// 				testDeduplicator, err := deduplicator.NewBaseDeduplicator(testName, testVersion, testLogger, testDataSession, testDataSet)
	// 				Expect(err).To(MatchError("data set id is missing"))
	// 				Expect(testDeduplicator).To(BeNil())
	// 			})

	// 			It("returns an error if the data set user id is missing", func() {
	// 				testDataSet.UserID = ""
	// 				testDeduplicator, err := deduplicator.NewBaseDeduplicator(testName, testVersion, testLogger, testDataSession, testDataSet)
	// 				Expect(err).To(MatchError("data set user id is missing"))
	// 				Expect(testDeduplicator).To(BeNil())
	// 			})

	// 			It("successfully returns a new deduplicator", func() {
	// 				Expect(deduplicator.NewBaseDeduplicator(testName, testVersion, testLogger, testDataSession, testDataSet)).ToNot(BeNil())
	// 			})
	// 		})

	// 		Context("with a context and new deduplicator", func() {
	// 			var ctx context.Context
	// 			var testDeduplicator data.Deduplicator

	// 			BeforeEach(func() {
	// 				ctx = context.Background()
	// 				var err error
	// 				testDeduplicator, err = deduplicator.NewBaseDeduplicator(testName, testVersion, testLogger, testDataSession, testDataSet)
	// 				Expect(err).ToNot(HaveOccurred())
	// 				Expect(testDeduplicator).ToNot(BeNil())
	// 			})

	// 			Context("Name", func() {
	// 				It("returns the name", func() {
	// 					Expect(testDeduplicator.Name()).To(Equal(testName))
	// 				})
	// 			})

	// 			Context("Version", func() {
	// 				It("returns the version", func() {
	// 					Expect(testDeduplicator.Version()).To(Equal(testVersion))
	// 				})
	// 			})

	// 			Context("RegisterDataSet", func() {
	// 				It("returns an error if a deduplicator already registered data set", func() {
	// 					testDataSet.SetDeduplicatorDescriptor(&data.DeduplicatorDescriptor{Name: "test", Version: "0.0.0"})
	// 					err := testDeduplicator.RegisterDataSet(ctx)
	// 					Expect(err).To(MatchError(fmt.Sprintf(`already registered data set with id "%s"`, testDataSet.UploadID)))
	// 				})

	// 				Context("with updating data set", func() {
	// 					var hash string

	// 					BeforeEach(func() {
	// 						testDataSession.UpdateDataSetOutputs = []testDataStoreDEPRECATED.UpdateDataSetOutput{{DataSet: upload.New(), Error: nil}}
	// 					})

	// 					AfterEach(func() {
	// 						Expect(testDataSession.UpdateDataSetInputs).To(ConsistOf(testDataStoreDEPRECATED.UpdateDataSetInput{
	// 							Context: ctx,
	// 							ID:      testDataSet.UploadID,
	// 							Update: &data.DataSetUpdate{
	// 								Active: pointer.FromBool(false),
	// 								Deduplicator: &data.DeduplicatorDescriptor{
	// 									Name:    testName,
	// 									Version: testVersion,
	// 									Hash:    hash,
	// 								},
	// 							},
	// 						}))
	// 					})

	// 					It("returns an error if there is an error with UpdateDataSet", func() {
	// 						testDataSession.UpdateDataSetOutputs = []testDataStoreDEPRECATED.UpdateDataSetOutput{{DataSet: nil, Error: errors.New("test error")}}
	// 						err := testDeduplicator.RegisterDataSet(ctx)
	// 						Expect(err).To(MatchError(fmt.Sprintf(`unable to update data set with id "%s"; test error`, testDataSet.UploadID)))
	// 					})

	// 					It("returns successfully if there is no error", func() {
	// 						Expect(testDeduplicator.RegisterDataSet(ctx)).To(Succeed())
	// 						Expect(testDataSet.DeduplicatorDescriptor()).To(BeNil())
	// 					})

	// 					It("returns successfully even if there is a deduplicator description just without a name", func() {
	// 						hash = "test"
	// 						testDataSet.SetDeduplicatorDescriptor(&data.DeduplicatorDescriptor{Hash: hash})
	// 						Expect(testDeduplicator.RegisterDataSet(ctx)).To(Succeed())
	// 						Expect(testDataSet.DeduplicatorDescriptor()).To(Equal(&data.DeduplicatorDescriptor{Name: testName, Version: testVersion, Hash: hash}))
	// 					})
	// 				})
	// 			})

	// 			Context("AddDataSetData", func() {
	// 				var testDataData []*testData.Datum
	// 				var testDataSetData []data.Datum

	// 				BeforeEach(func() {
	// 					testDataData = []*testData.Datum{}
	// 					testDataSetData = []data.Datum{}
	// 					for i := 0; i < 3; i++ {
	// 						testDatum := testData.NewDatum()
	// 						testDataData = append(testDataData, testDatum)
	// 						testDataSetData = append(testDataSetData, testDatum)
	// 					}
	// 				})

	// 				AfterEach(func() {
	// 					for _, testDataDatum := range testDataData {
	// 						testDataDatum.Expectations()
	// 					}
	// 				})

	// 				It("returns successfully if the data is missing", func() {
	// 					Expect(testDeduplicator.AddDataSetData(ctx, nil)).To(Succeed())
	// 				})

	// 				It("returns successfully if the data is empty", func() {
	// 					Expect(testDeduplicator.AddDataSetData(ctx, []data.Datum{})).To(Succeed())
	// 				})

	// 				Context("with creating data set data", func() {
	// 					BeforeEach(func() {
	// 						testDataSession.CreateDataSetDataOutputs = []error{nil}
	// 					})

	// 					AfterEach(func() {
	// 						Expect(testDataSession.CreateDataSetDataInputs).To(ConsistOf(testDataStoreDEPRECATED.CreateDataSetDataInput{Context: ctx, DataSet: testDataSet, DataSetData: testDataSetData}))
	// 					})

	// 					It("returns an error if there is an error with CreateDataSetData", func() {
	// 						testDataSession.CreateDataSetDataOutputs = []error{errors.New("test error")}
	// 						err := testDeduplicator.AddDataSetData(ctx, testDataSetData)
	// 						Expect(err).To(MatchError(fmt.Sprintf(`unable to create data set data with id "%s"; test error`, testDataSet.UploadID)))
	// 					})

	// 					It("returns successfully if there is no error", func() {
	// 						Expect(testDeduplicator.AddDataSetData(ctx, testDataSetData)).To(Succeed())
	// 					})
	// 				})
	// 			})

	// 			Context("DeduplicateDataSet", func() {
	// 				Context("with activating data set data", func() {
	// 					BeforeEach(func() {
	// 						testDataSession.ActivateDataSetDataOutputs = []error{nil}
	// 					})

	// 					AfterEach(func() {
	// 						Expect(testDataSession.ActivateDataSetDataInputs).To(ConsistOf(testDataStoreDEPRECATED.ActivateDataSetDataInput{Context: ctx, DataSet: testDataSet}))
	// 					})

	// 					It("returns an error if there is an error with ActivateDataSetData", func() {
	// 						testDataSession.ActivateDataSetDataOutputs = []error{errors.New("test error")}
	// 						err := testDeduplicator.DeduplicateDataSet(ctx)
	// 						Expect(err).To(MatchError(fmt.Sprintf(`unable to activate data set data with id "%s"; test error`, testDataSet.UploadID)))
	// 					})

	// 					It("returns successfully if there is no error", func() {
	// 						Expect(testDeduplicator.DeduplicateDataSet(ctx)).To(Succeed())
	// 					})
	// 				})
	// 			})

	// 			Context("DeleteDataSet", func() {
	// 				Context("with deleting data set", func() {
	// 					BeforeEach(func() {
	// 						testDataSession.DeleteDataSetOutputs = []error{nil}
	// 					})

	// 					AfterEach(func() {
	// 						Expect(testDataSession.DeleteDataSetInputs).To(ConsistOf(testDataStoreDEPRECATED.DeleteDataSetInput{Context: ctx, DataSet: testDataSet}))
	// 					})

	// 					It("returns an error if there is an error with DeleteDataSet", func() {
	// 						testDataSession.DeleteDataSetOutputs = []error{errors.New("test error")}
	// 						err := testDeduplicator.DeleteDataSet(ctx)
	// 						Expect(err).To(MatchError(fmt.Sprintf(`unable to delete data set with id "%s"; test error`, testDataSet.UploadID)))
	// 					})

	// 					It("returns successfully if there is no error", func() {
	// 						Expect(testDeduplicator.DeleteDataSet(ctx)).To(Succeed())
	// 					})
	// 				})
	// 			})
	// 		})
	// 	})
})
