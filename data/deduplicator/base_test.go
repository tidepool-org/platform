package deduplicator_test

// import (
// 	"context"

// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"

// 	"errors"
// 	"fmt"

// 	"github.com/tidepool-org/platform/data"
// 	"github.com/tidepool-org/platform/data/deduplicator"
// 	testDataStoreDEPRECATED "github.com/tidepool-org/platform/data/storeDEPRECATED/test"
// 	testData "github.com/tidepool-org/platform/data/test"
// 	"github.com/tidepool-org/platform/data/types/upload"
// 	"github.com/tidepool-org/platform/id"
// 	"github.com/tidepool-org/platform/log"
// 	"github.com/tidepool-org/platform/log/null"
// 	"github.com/tidepool-org/platform/pointer"
// )

// var _ = Describe("Base", func() {
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
// 			var testDataset *upload.Upload

// 			BeforeEach(func() {
// 				var err error
// 				testFactory, err = deduplicator.NewBaseFactory(testName, testVersion)
// 				Expect(err).ToNot(HaveOccurred())
// 				Expect(testFactory).ToNot(BeNil())
// 				testDataset = upload.New()
// 				Expect(testDataset).ToNot(BeNil())
// 				testDataset.UserID = id.New()
// 			})

// 			Context("CanDeduplicateDataset", func() {
// 				It("returns an error if the dataset is missing", func() {
// 					can, err := testFactory.CanDeduplicateDataset(nil)
// 					Expect(err).To(MatchError("dataset is missing"))
// 					Expect(can).To(BeFalse())
// 				})

// 				It("returns false if the dataset id is missing", func() {
// 					testDataset.UploadID = ""
// 					Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
// 				})

// 				It("returns false if the dataset user id is missing", func() {
// 					testDataset.UserID = ""
// 					Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
// 				})

// 				It("returns true if successful", func() {
// 					Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeTrue())
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

// 				Context("NewDeduplicatorForDataset", func() {
// 					It("returns an error if the logger is missing", func() {
// 						testDeduplicator, err := testFactory.NewDeduplicatorForDataset(nil, testDataSession, testDataset)
// 						Expect(err).To(MatchError("logger is missing"))
// 						Expect(testDeduplicator).To(BeNil())
// 					})

// 					It("returns an error if the data store session is missing", func() {
// 						testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, nil, testDataset)
// 						Expect(err).To(MatchError("data store session is missing"))
// 						Expect(testDeduplicator).To(BeNil())
// 					})

// 					It("returns an error if the dataset is missing", func() {
// 						testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, nil)
// 						Expect(err).To(MatchError("dataset is missing"))
// 						Expect(testDeduplicator).To(BeNil())
// 					})

// 					It("returns an error if the dataset id is missing", func() {
// 						testDataset.UploadID = ""
// 						testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
// 						Expect(err).To(MatchError("dataset id is missing"))
// 						Expect(testDeduplicator).To(BeNil())
// 					})

// 					It("returns an error if the dataset user id is missing", func() {
// 						testDataset.UserID = ""
// 						testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
// 						Expect(err).To(MatchError("dataset user id is missing"))
// 						Expect(testDeduplicator).To(BeNil())
// 					})

// 					It("returns a new deduplicator upon success", func() {
// 						Expect(testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)).ToNot(BeNil())
// 					})
// 				})
// 			})

// 			Context("with registered dataset", func() {
// 				BeforeEach(func() {
// 					testDataset.Deduplicator = &data.DeduplicatorDescriptor{Name: testName, Version: testVersion}
// 				})

// 				Context("IsRegisteredWithDataset", func() {
// 					It("returns an error if the dataset is missing", func() {
// 						can, err := testFactory.IsRegisteredWithDataset(nil)
// 						Expect(err).To(MatchError("dataset is missing"))
// 						Expect(can).To(BeFalse())
// 					})

// 					It("returns false if the dataset id is missing", func() {
// 						testDataset.UploadID = ""
// 						Expect(testFactory.IsRegisteredWithDataset(testDataset)).To(BeFalse())
// 					})

// 					It("returns false if the dataset user id is missing", func() {
// 						testDataset.UserID = ""
// 						Expect(testFactory.IsRegisteredWithDataset(testDataset)).To(BeFalse())
// 					})

// 					It("returns false if there is no deduplicator descriptor", func() {
// 						testDataset.Deduplicator = nil
// 						Expect(testFactory.IsRegisteredWithDataset(testDataset)).To(BeFalse())
// 					})

// 					It("returns false if the deduplicator descriptor name is missing", func() {
// 						testDataset.Deduplicator.Name = ""
// 						Expect(testFactory.IsRegisteredWithDataset(testDataset)).To(BeFalse())
// 					})

// 					It("returns false if the deduplicator descriptor name does not match", func() {
// 						testDataset.Deduplicator.Name = id.New()
// 						Expect(testFactory.IsRegisteredWithDataset(testDataset)).To(BeFalse())
// 					})

// 					It("returns true if successful", func() {
// 						Expect(testFactory.IsRegisteredWithDataset(testDataset)).To(BeTrue())
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

// 					Context("NewRegisteredDeduplicatorForDataset", func() {
// 						It("returns an error if the logger is missing", func() {
// 							testDeduplicator, err := testFactory.NewRegisteredDeduplicatorForDataset(nil, testDataSession, testDataset)
// 							Expect(err).To(MatchError("logger is missing"))
// 							Expect(testDeduplicator).To(BeNil())
// 						})

// 						It("returns an error if the data store session is missing", func() {
// 							testDeduplicator, err := testFactory.NewRegisteredDeduplicatorForDataset(testLogger, nil, testDataset)
// 							Expect(err).To(MatchError("data store session is missing"))
// 							Expect(testDeduplicator).To(BeNil())
// 						})

// 						It("returns an error if the dataset is missing", func() {
// 							testDeduplicator, err := testFactory.NewRegisteredDeduplicatorForDataset(testLogger, testDataSession, nil)
// 							Expect(err).To(MatchError("dataset is missing"))
// 							Expect(testDeduplicator).To(BeNil())
// 						})

// 						It("returns an error if the dataset id is missing", func() {
// 							testDataset.UploadID = ""
// 							testDeduplicator, err := testFactory.NewRegisteredDeduplicatorForDataset(testLogger, testDataSession, testDataset)
// 							Expect(err).To(MatchError("dataset id is missing"))
// 							Expect(testDeduplicator).To(BeNil())
// 						})

// 						It("returns an error if the dataset user id is missing", func() {
// 							testDataset.UserID = ""
// 							testDeduplicator, err := testFactory.NewRegisteredDeduplicatorForDataset(testLogger, testDataSession, testDataset)
// 							Expect(err).To(MatchError("dataset user id is missing"))
// 							Expect(testDeduplicator).To(BeNil())
// 						})

// 						It("returns an error if there is no deduplicator descriptor", func() {
// 							testDataset.Deduplicator = nil
// 							testDeduplicator, err := testFactory.NewRegisteredDeduplicatorForDataset(testLogger, testDataSession, testDataset)
// 							Expect(err).To(MatchError("dataset deduplicator descriptor is missing"))
// 							Expect(testDeduplicator).To(BeNil())
// 						})

// 						It("returns an error if the deduplicator descriptor name is missing", func() {
// 							testDataset.Deduplicator.Name = ""
// 							testDeduplicator, err := testFactory.NewRegisteredDeduplicatorForDataset(testLogger, testDataSession, testDataset)
// 							Expect(err).To(MatchError("dataset deduplicator descriptor is not registered with expected deduplicator"))
// 							Expect(testDeduplicator).To(BeNil())
// 						})

// 						It("returns an error if the deduplicator descriptor name does not match", func() {
// 							testDataset.Deduplicator.Name = id.New()
// 							testDeduplicator, err := testFactory.NewRegisteredDeduplicatorForDataset(testLogger, testDataSession, testDataset)
// 							Expect(err).To(MatchError("dataset deduplicator descriptor is not registered with expected deduplicator"))
// 							Expect(testDeduplicator).To(BeNil())
// 						})

// 						It("returns a new deduplicator upon success", func() {
// 							Expect(testFactory.NewRegisteredDeduplicatorForDataset(testLogger, testDataSession, testDataset)).ToNot(BeNil())
// 						})
// 					})
// 				})
// 			})
// 		})
// 	})

// 	Context("BaseDeduplicator", func() {
// 		var testLogger log.Logger
// 		var testDataSession *testDataStoreDEPRECATED.DataSession
// 		var testDataset *upload.Upload

// 		BeforeEach(func() {
// 			testLogger = null.NewLogger()
// 			Expect(testLogger).ToNot(BeNil())
// 			testDataSession = testDataStoreDEPRECATED.NewDataSession()
// 			Expect(testDataSession).ToNot(BeNil())
// 			testDataset = upload.New()
// 			Expect(testDataset).ToNot(BeNil())
// 			testDataset.UserID = id.New()
// 		})

// 		AfterEach(func() {
// 			testDataSession.Expectations()
// 		})

// 		Context("NewBaseDeduplicator", func() {
// 			It("returns an error if the name is missing", func() {
// 				testDeduplicator, err := deduplicator.NewBaseDeduplicator("", testVersion, testLogger, testDataSession, testDataset)
// 				Expect(err).To(MatchError("name is missing"))
// 				Expect(testDeduplicator).To(BeNil())
// 			})

// 			It("returns an error if the version is missing", func() {
// 				testDeduplicator, err := deduplicator.NewBaseDeduplicator(testName, "", testLogger, testDataSession, testDataset)
// 				Expect(err).To(MatchError("version is missing"))
// 				Expect(testDeduplicator).To(BeNil())
// 			})

// 			It("returns an error if the version is invalid", func() {
// 				testDeduplicator, err := deduplicator.NewBaseDeduplicator(testName, "x.y.z", testLogger, testDataSession, testDataset)
// 				Expect(err).To(MatchError("version is invalid"))
// 				Expect(testDeduplicator).To(BeNil())
// 			})

// 			It("returns an error if the logger is missing", func() {
// 				testDeduplicator, err := deduplicator.NewBaseDeduplicator(testName, testVersion, nil, testDataSession, testDataset)
// 				Expect(err).To(MatchError("logger is missing"))
// 				Expect(testDeduplicator).To(BeNil())
// 			})

// 			It("returns an error if the data store session is missing", func() {
// 				testDeduplicator, err := deduplicator.NewBaseDeduplicator(testName, testVersion, testLogger, nil, testDataset)
// 				Expect(err).To(MatchError("data store session is missing"))
// 				Expect(testDeduplicator).To(BeNil())
// 			})

// 			It("returns an error if the dataset is missing", func() {
// 				testDeduplicator, err := deduplicator.NewBaseDeduplicator(testName, testVersion, testLogger, testDataSession, nil)
// 				Expect(err).To(MatchError("dataset is missing"))
// 				Expect(testDeduplicator).To(BeNil())
// 			})

// 			It("returns an error if the dataset id is missing", func() {
// 				testDataset.UploadID = ""
// 				testDeduplicator, err := deduplicator.NewBaseDeduplicator(testName, testVersion, testLogger, testDataSession, testDataset)
// 				Expect(err).To(MatchError("dataset id is missing"))
// 				Expect(testDeduplicator).To(BeNil())
// 			})

// 			It("returns an error if the dataset user id is missing", func() {
// 				testDataset.UserID = ""
// 				testDeduplicator, err := deduplicator.NewBaseDeduplicator(testName, testVersion, testLogger, testDataSession, testDataset)
// 				Expect(err).To(MatchError("dataset user id is missing"))
// 				Expect(testDeduplicator).To(BeNil())
// 			})

// 			It("successfully returns a new deduplicator", func() {
// 				Expect(deduplicator.NewBaseDeduplicator(testName, testVersion, testLogger, testDataSession, testDataset)).ToNot(BeNil())
// 			})
// 		})

// 		Context("with a context and new deduplicator", func() {
// 			var ctx context.Context
// 			var testDeduplicator data.Deduplicator

// 			BeforeEach(func() {
// 				ctx = context.Background()
// 				var err error
// 				testDeduplicator, err = deduplicator.NewBaseDeduplicator(testName, testVersion, testLogger, testDataSession, testDataset)
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

// 			Context("RegisterDataset", func() {
// 				It("returns an error if a deduplicator already registered dataset", func() {
// 					testDataset.SetDeduplicatorDescriptor(&data.DeduplicatorDescriptor{Name: "test", Version: "0.0.0"})
// 					err := testDeduplicator.RegisterDataset(ctx)
// 					Expect(err).To(MatchError(fmt.Sprintf(`already registered dataset with id "%s"`, testDataset.UploadID)))
// 				})

// 				Context("with updating dataset", func() {
// 					var hash string

// 					BeforeEach(func() {
// 						testDataSession.UpdateDataSetOutputs = []testDataStoreDEPRECATED.UpdateDataSetOutput{{DataSet: upload.New(), Error: nil}}
// 					})

// 					AfterEach(func() {
// 						Expect(testDataSession.UpdateDataSetInputs).To(ConsistOf(testDataStoreDEPRECATED.UpdateDataSetInput{
// 							Context: ctx,
// 							ID:      testDataset.UploadID,
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

// 					It("returns an error if there is an error with UpdateDataset", func() {
// 						testDataSession.UpdateDataSetOutputs = []testDataStoreDEPRECATED.UpdateDataSetOutput{{DataSet: nil, Error: errors.New("test error")}}
// 						err := testDeduplicator.RegisterDataset(ctx)
// 						Expect(err).To(MatchError(fmt.Sprintf(`unable to update dataset with id "%s"; test error`, testDataset.UploadID)))
// 					})

// 					It("returns successfully if there is no error", func() {
// 						Expect(testDeduplicator.RegisterDataset(ctx)).To(Succeed())
// 						Expect(testDataset.DeduplicatorDescriptor()).To(BeNil())
// 					})

// 					It("returns successfully even if there is a deduplicator description just without a name", func() {
// 						hash = "test"
// 						testDataset.SetDeduplicatorDescriptor(&data.DeduplicatorDescriptor{Hash: hash})
// 						Expect(testDeduplicator.RegisterDataset(ctx)).To(Succeed())
// 						Expect(testDataset.DeduplicatorDescriptor()).To(Equal(&data.DeduplicatorDescriptor{Name: testName, Version: testVersion, Hash: hash}))
// 					})
// 				})
// 			})

// 			Context("AddDatasetData", func() {
// 				var testDataData []*testData.Datum
// 				var testDatasetData []data.Datum

// 				BeforeEach(func() {
// 					testDataData = []*testData.Datum{}
// 					testDatasetData = []data.Datum{}
// 					for i := 0; i < 3; i++ {
// 						testDatum := testData.NewDatum()
// 						testDataData = append(testDataData, testDatum)
// 						testDatasetData = append(testDatasetData, testDatum)
// 					}
// 				})

// 				AfterEach(func() {
// 					for _, testDataDatum := range testDataData {
// 						testDataDatum.Expectations()
// 					}
// 				})

// 				It("returns successfully if the data is missing", func() {
// 					Expect(testDeduplicator.AddDatasetData(ctx, nil)).To(Succeed())
// 				})

// 				It("returns successfully if the data is empty", func() {
// 					Expect(testDeduplicator.AddDatasetData(ctx, []data.Datum{})).To(Succeed())
// 				})

// 				Context("with creating dataset data", func() {
// 					BeforeEach(func() {
// 						testDataSession.CreateDatasetDataOutputs = []error{nil}
// 					})

// 					AfterEach(func() {
// 						Expect(testDataSession.CreateDatasetDataInputs).To(ConsistOf(testDataStoreDEPRECATED.CreateDatasetDataInput{Context: ctx, Dataset: testDataset, DatasetData: testDatasetData}))
// 					})

// 					It("returns an error if there is an error with CreateDatasetData", func() {
// 						testDataSession.CreateDatasetDataOutputs = []error{errors.New("test error")}
// 						err := testDeduplicator.AddDatasetData(ctx, testDatasetData)
// 						Expect(err).To(MatchError(fmt.Sprintf(`unable to create dataset data with id "%s"; test error`, testDataset.UploadID)))
// 					})

// 					It("returns successfully if there is no error", func() {
// 						Expect(testDeduplicator.AddDatasetData(ctx, testDatasetData)).To(Succeed())
// 					})
// 				})
// 			})

// 			Context("DeduplicateDataset", func() {
// 				Context("with activating dataset data", func() {
// 					BeforeEach(func() {
// 						testDataSession.ActivateDatasetDataOutputs = []error{nil}
// 					})

// 					AfterEach(func() {
// 						Expect(testDataSession.ActivateDatasetDataInputs).To(ConsistOf(testDataStoreDEPRECATED.ActivateDatasetDataInput{Context: ctx, Dataset: testDataset}))
// 					})

// 					It("returns an error if there is an error with ActivateDatasetData", func() {
// 						testDataSession.ActivateDatasetDataOutputs = []error{errors.New("test error")}
// 						err := testDeduplicator.DeduplicateDataset(ctx)
// 						Expect(err).To(MatchError(fmt.Sprintf(`unable to activate dataset data with id "%s"; test error`, testDataset.UploadID)))
// 					})

// 					It("returns successfully if there is no error", func() {
// 						Expect(testDeduplicator.DeduplicateDataset(ctx)).To(Succeed())
// 					})
// 				})
// 			})

// 			Context("DeleteDataset", func() {
// 				Context("with deleting dataset", func() {
// 					BeforeEach(func() {
// 						testDataSession.DeleteDatasetOutputs = []error{nil}
// 					})

// 					AfterEach(func() {
// 						Expect(testDataSession.DeleteDatasetInputs).To(ConsistOf(testDataStoreDEPRECATED.DeleteDatasetInput{Context: ctx, Dataset: testDataset}))
// 					})

// 					It("returns an error if there is an error with DeleteDataset", func() {
// 						testDataSession.DeleteDatasetOutputs = []error{errors.New("test error")}
// 						err := testDeduplicator.DeleteDataset(ctx)
// 						Expect(err).To(MatchError(fmt.Sprintf(`unable to delete dataset with id "%s"; test error`, testDataset.UploadID)))
// 					})

// 					It("returns successfully if there is no error", func() {
// 						Expect(testDeduplicator.DeleteDataset(ctx)).To(Succeed())
// 					})
// 				})
// 			})
// 		})
// 	})
// })
