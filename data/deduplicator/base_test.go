package deduplicator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	"fmt"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	testDataStore "github.com/tidepool-org/platform/data/store/test"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/log"
)

var _ = Describe("Base", func() {
	var testName string

	BeforeEach(func() {
		testName = app.NewID()
	})

	Context("BaseFactory", func() {
		Context("NewBaseFactory", func() {
			It("returns an error if the name is missing", func() {
				testFactory, err := deduplicator.NewBaseFactory("")
				Expect(err).To(MatchError("deduplicator: name is missing"))
				Expect(testFactory).To(BeNil())
			})

			It("returns a new factory", func() {
				testFactory, err := deduplicator.NewBaseFactory(testName)
				Expect(err).ToNot(HaveOccurred())
				Expect(testFactory).ToNot(BeNil())
				Expect(testFactory.Factory).ToNot(BeNil())
			})
		})

		Context("with a new factory", func() {
			var testFactory *deduplicator.BaseFactory
			var testDataset *upload.Upload

			BeforeEach(func() {
				var err error
				testFactory, err = deduplicator.NewBaseFactory(testName)
				Expect(err).ToNot(HaveOccurred())
				Expect(testFactory).ToNot(BeNil())
				testDataset = upload.Init()
				Expect(testDataset).ToNot(BeNil())
				testDataset.UserID = app.NewID()
				testDataset.GroupID = app.NewID()
			})

			Context("CanDeduplicateDataset", func() {
				It("returns an error if the dataset is missing", func() {
					can, err := testFactory.CanDeduplicateDataset(nil)
					Expect(err).To(MatchError("deduplicator: dataset is missing"))
					Expect(can).To(BeFalse())
				})

				It("returns false if the dataset id is missing", func() {
					testDataset.UploadID = ""
					Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
				})

				It("returns false if the dataset user id is missing", func() {
					testDataset.UserID = ""
					Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
				})

				It("returns false if the dataset group id is missing", func() {
					testDataset.GroupID = ""
					Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
				})

				It("returns true if successful", func() {
					Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeTrue())
				})
			})

			Context("with logger and data store session", func() {
				var testLogger log.Logger
				var testDataStoreSession *testDataStore.Session

				BeforeEach(func() {
					testLogger = log.NewNull()
					Expect(testLogger).ToNot(BeNil())
					testDataStoreSession = testDataStore.NewSession()
					Expect(testDataStoreSession).ToNot(BeNil())
				})

				AfterEach(func() {
					Expect(testDataStoreSession.UnusedOutputsCount()).To(Equal(0))
				})

				Context("NewDeduplicatorForDataset", func() {
					It("returns an error if the logger is missing", func() {
						testDeduplicator, err := testFactory.NewDeduplicatorForDataset(nil, testDataStoreSession, testDataset)
						Expect(err).To(MatchError("deduplicator: logger is missing"))
						Expect(testDeduplicator).To(BeNil())
					})

					It("returns an error if the data store session is missing", func() {
						testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, nil, testDataset)
						Expect(err).To(MatchError("deduplicator: data store session is missing"))
						Expect(testDeduplicator).To(BeNil())
					})

					It("returns an error if the dataset is missing", func() {
						testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataStoreSession, nil)
						Expect(err).To(MatchError("deduplicator: dataset is missing"))
						Expect(testDeduplicator).To(BeNil())
					})

					It("returns an error if the dataset id is missing", func() {
						testDataset.UploadID = ""
						testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)
						Expect(err).To(MatchError("deduplicator: dataset id is missing"))
						Expect(testDeduplicator).To(BeNil())
					})

					It("returns an error if the dataset user id is missing", func() {
						testDataset.UserID = ""
						testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)
						Expect(err).To(MatchError("deduplicator: dataset user id is missing"))
						Expect(testDeduplicator).To(BeNil())
					})

					It("returns an error if the dataset group id is missing", func() {
						testDataset.GroupID = ""
						testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)
						Expect(err).To(MatchError("deduplicator: dataset group id is missing"))
						Expect(testDeduplicator).To(BeNil())
					})

					It("returns a new deduplicator upon success", func() {
						Expect(testFactory.NewDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)).ToNot(BeNil())
					})
				})
			})

			Context("with registered dataset", func() {
				BeforeEach(func() {
					testDataset.Deduplicator = data.NewDeduplicatorDescriptor()
					testDataset.Deduplicator.RegisterWithNamedDeduplicator(testName)
				})

				Context("IsRegisteredWithDataset", func() {
					It("returns an error if the dataset is missing", func() {
						can, err := testFactory.IsRegisteredWithDataset(nil)
						Expect(err).To(MatchError("deduplicator: dataset is missing"))
						Expect(can).To(BeFalse())
					})

					It("returns false if the dataset id is missing", func() {
						testDataset.UploadID = ""
						Expect(testFactory.IsRegisteredWithDataset(testDataset)).To(BeFalse())
					})

					It("returns false if the dataset user id is missing", func() {
						testDataset.UserID = ""
						Expect(testFactory.IsRegisteredWithDataset(testDataset)).To(BeFalse())
					})

					It("returns false if the dataset group id is missing", func() {
						testDataset.GroupID = ""
						Expect(testFactory.IsRegisteredWithDataset(testDataset)).To(BeFalse())
					})

					It("returns false if there is no deduplicator descriptor", func() {
						testDataset.Deduplicator = nil
						Expect(testFactory.IsRegisteredWithDataset(testDataset)).To(BeFalse())
					})

					It("returns false if the deduplicator descriptor name is missing", func() {
						testDataset.Deduplicator.Name = ""
						Expect(testFactory.IsRegisteredWithDataset(testDataset)).To(BeFalse())
					})

					It("returns false if the deduplicator descriptor name does not match", func() {
						testDataset.Deduplicator.Name = app.NewID()
						Expect(testFactory.IsRegisteredWithDataset(testDataset)).To(BeFalse())
					})

					It("returns true if successful", func() {
						Expect(testFactory.IsRegisteredWithDataset(testDataset)).To(BeTrue())
					})
				})

				Context("with logger and data store session", func() {
					var testLogger log.Logger
					var testDataStoreSession *testDataStore.Session

					BeforeEach(func() {
						testLogger = log.NewNull()
						Expect(testLogger).ToNot(BeNil())
						testDataStoreSession = testDataStore.NewSession()
						Expect(testDataStoreSession).ToNot(BeNil())
					})

					AfterEach(func() {
						Expect(testDataStoreSession.UnusedOutputsCount()).To(Equal(0))
					})

					Context("NewRegisteredDeduplicatorForDataset", func() {
						It("returns an error if the logger is missing", func() {
							testDeduplicator, err := testFactory.NewRegisteredDeduplicatorForDataset(nil, testDataStoreSession, testDataset)
							Expect(err).To(MatchError("deduplicator: logger is missing"))
							Expect(testDeduplicator).To(BeNil())
						})

						It("returns an error if the data store session is missing", func() {
							testDeduplicator, err := testFactory.NewRegisteredDeduplicatorForDataset(testLogger, nil, testDataset)
							Expect(err).To(MatchError("deduplicator: data store session is missing"))
							Expect(testDeduplicator).To(BeNil())
						})

						It("returns an error if the dataset is missing", func() {
							testDeduplicator, err := testFactory.NewRegisteredDeduplicatorForDataset(testLogger, testDataStoreSession, nil)
							Expect(err).To(MatchError("deduplicator: dataset is missing"))
							Expect(testDeduplicator).To(BeNil())
						})

						It("returns an error if the dataset id is missing", func() {
							testDataset.UploadID = ""
							testDeduplicator, err := testFactory.NewRegisteredDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)
							Expect(err).To(MatchError("deduplicator: dataset id is missing"))
							Expect(testDeduplicator).To(BeNil())
						})

						It("returns an error if the dataset user id is missing", func() {
							testDataset.UserID = ""
							testDeduplicator, err := testFactory.NewRegisteredDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)
							Expect(err).To(MatchError("deduplicator: dataset user id is missing"))
							Expect(testDeduplicator).To(BeNil())
						})

						It("returns an error if the dataset group id is missing", func() {
							testDataset.GroupID = ""
							testDeduplicator, err := testFactory.NewRegisteredDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)
							Expect(err).To(MatchError("deduplicator: dataset group id is missing"))
							Expect(testDeduplicator).To(BeNil())
						})

						It("returns an error if there is no deduplicator descriptor", func() {
							testDataset.Deduplicator = nil
							testDeduplicator, err := testFactory.NewRegisteredDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)
							Expect(err).To(MatchError("deduplicator: dataset deduplicator descriptor is missing"))
							Expect(testDeduplicator).To(BeNil())
						})

						It("returns an error if the deduplicator descriptor name is missing", func() {
							testDataset.Deduplicator.Name = ""
							testDeduplicator, err := testFactory.NewRegisteredDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)
							Expect(err).To(MatchError("deduplicator: dataset deduplicator descriptor is not registered with expected deduplicator"))
							Expect(testDeduplicator).To(BeNil())
						})

						It("returns an error if the deduplicator descriptor name does not match", func() {
							testDataset.Deduplicator.Name = app.NewID()
							testDeduplicator, err := testFactory.NewRegisteredDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)
							Expect(err).To(MatchError("deduplicator: dataset deduplicator descriptor is not registered with expected deduplicator"))
							Expect(testDeduplicator).To(BeNil())
						})

						It("returns a new deduplicator upon success", func() {
							Expect(testFactory.NewRegisteredDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)).ToNot(BeNil())
						})
					})
				})
			})
		})
	})

	Context("BaseDeduplicator", func() {
		var testLogger log.Logger
		var testDataStoreSession *testDataStore.Session
		var testDataset *upload.Upload

		BeforeEach(func() {
			testLogger = log.NewNull()
			Expect(testLogger).ToNot(BeNil())
			testDataStoreSession = testDataStore.NewSession()
			Expect(testDataStoreSession).ToNot(BeNil())
			testDataset = upload.Init()
			Expect(testDataset).ToNot(BeNil())
			testDataset.UserID = app.NewID()
			testDataset.GroupID = app.NewID()
		})

		AfterEach(func() {
			Expect(testDataStoreSession.UnusedOutputsCount()).To(Equal(0))
		})

		Context("NewBaseDeduplicator", func() {
			It("returns an error if the name is missing", func() {
				testDeduplicator, err := deduplicator.NewBaseDeduplicator("", testLogger, testDataStoreSession, testDataset)
				Expect(err).To(MatchError("deduplicator: name is missing"))
				Expect(testDeduplicator).To(BeNil())
			})

			It("returns an error if the logger is missing", func() {
				testDeduplicator, err := deduplicator.NewBaseDeduplicator(testName, nil, testDataStoreSession, testDataset)
				Expect(err).To(MatchError("deduplicator: logger is missing"))
				Expect(testDeduplicator).To(BeNil())
			})

			It("returns an error if the data store session is missing", func() {
				testDeduplicator, err := deduplicator.NewBaseDeduplicator(testName, testLogger, nil, testDataset)
				Expect(err).To(MatchError("deduplicator: data store session is missing"))
				Expect(testDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset is missing", func() {
				testDeduplicator, err := deduplicator.NewBaseDeduplicator(testName, testLogger, testDataStoreSession, nil)
				Expect(err).To(MatchError("deduplicator: dataset is missing"))
				Expect(testDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset id is missing", func() {
				testDataset.UploadID = ""
				testDeduplicator, err := deduplicator.NewBaseDeduplicator(testName, testLogger, testDataStoreSession, testDataset)
				Expect(err).To(MatchError("deduplicator: dataset id is missing"))
				Expect(testDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset user id is missing", func() {
				testDataset.UserID = ""
				testDeduplicator, err := deduplicator.NewBaseDeduplicator(testName, testLogger, testDataStoreSession, testDataset)
				Expect(err).To(MatchError("deduplicator: dataset user id is missing"))
				Expect(testDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset group id is missing", func() {
				testDataset.GroupID = ""
				testDeduplicator, err := deduplicator.NewBaseDeduplicator(testName, testLogger, testDataStoreSession, testDataset)
				Expect(err).To(MatchError("deduplicator: dataset group id is missing"))
				Expect(testDeduplicator).To(BeNil())
			})

			It("successfully returns a new deduplicator", func() {
				Expect(deduplicator.NewBaseDeduplicator(testName, testLogger, testDataStoreSession, testDataset)).ToNot(BeNil())
			})
		})

		Context("with a new deduplicator", func() {
			var testDeduplicator data.Deduplicator

			BeforeEach(func() {
				var err error
				testDeduplicator, err = deduplicator.NewBaseDeduplicator(testName, testLogger, testDataStoreSession, testDataset)
				Expect(err).ToNot(HaveOccurred())
				Expect(testDeduplicator).ToNot(BeNil())
			})

			Context("Name", func() {
				It("returns the name", func() {
					Expect(testDeduplicator.Name()).To(Equal(testName))
				})
			})

			Context("RegisterDataset", func() {
				It("returns an error if a deduplicator already registered dataset", func() {
					testDataset.SetDeduplicatorDescriptor(&data.DeduplicatorDescriptor{Name: "test"})
					err := testDeduplicator.RegisterDataset()
					Expect(err).To(MatchError(fmt.Sprintf(`deduplicator: already registered dataset with id "%s"`, testDataset.UploadID)))
				})

				Context("with updating dataset", func() {
					BeforeEach(func() {
						testDataStoreSession.UpdateDatasetOutputs = []error{nil}
					})

					AfterEach(func() {
						Expect(testDataStoreSession.UpdateDatasetInputs).To(ConsistOf(testDataset))
					})

					It("returns an error if there is an error with UpdateDataset", func() {
						testDataStoreSession.UpdateDatasetOutputs = []error{errors.New("test error")}
						err := testDeduplicator.RegisterDataset()
						Expect(err).To(MatchError(fmt.Sprintf(`deduplicator: unable to update dataset with id "%s"; test error`, testDataset.UploadID)))
					})

					It("returns successfully if there is no error", func() {
						Expect(testDeduplicator.RegisterDataset()).To(Succeed())
						Expect(testDataset.DeduplicatorDescriptor()).To(Equal(&data.DeduplicatorDescriptor{Name: testName}))
					})

					It("returns successfully even if there is a deduplicator description just without a name", func() {
						testDataset.SetDeduplicatorDescriptor(&data.DeduplicatorDescriptor{Hash: "test"})
						Expect(testDeduplicator.RegisterDataset()).To(Succeed())
						Expect(testDataset.DeduplicatorDescriptor()).To(Equal(&data.DeduplicatorDescriptor{Name: testName, Hash: "test"}))
					})
				})
			})

			Context("AddDatasetData", func() {
				var testDataData []*testData.Datum
				var testDatasetData []data.Datum

				BeforeEach(func() {
					testDataData = []*testData.Datum{}
					testDatasetData = []data.Datum{}
					for i := 0; i < 3; i++ {
						testDatum := testData.NewDatum()
						testDataData = append(testDataData, testDatum)
						testDatasetData = append(testDatasetData, testDatum)
					}
				})

				AfterEach(func() {
					for _, testDataDatum := range testDataData {
						Expect(testDataDatum.UnusedOutputsCount()).To(Equal(0))
					}
				})

				It("returns successfully if the data is missing", func() {
					Expect(testDeduplicator.AddDatasetData(nil)).To(Succeed())
				})

				It("returns successfully if the data is empty", func() {
					Expect(testDeduplicator.AddDatasetData([]data.Datum{})).To(Succeed())
				})

				Context("with creating dataset data", func() {
					BeforeEach(func() {
						testDataStoreSession.CreateDatasetDataOutputs = []error{nil}
					})

					AfterEach(func() {
						Expect(testDataStoreSession.CreateDatasetDataInputs).To(ConsistOf(testDataStore.CreateDatasetDataInput{Dataset: testDataset, DatasetData: testDatasetData}))
					})

					It("returns an error if there is an error with CreateDatasetData", func() {
						testDataStoreSession.CreateDatasetDataOutputs = []error{errors.New("test error")}
						err := testDeduplicator.AddDatasetData(testDatasetData)
						Expect(err).To(MatchError(fmt.Sprintf(`deduplicator: unable to create dataset data with id "%s"; test error`, testDataset.UploadID)))
					})

					It("returns successfully if there is no error", func() {
						Expect(testDeduplicator.AddDatasetData(testDatasetData)).To(Succeed())
					})
				})
			})

			Context("DeduplicateDataset", func() {
				Context("with activating dataset data", func() {
					BeforeEach(func() {
						testDataStoreSession.ActivateDatasetDataOutputs = []error{nil}
					})

					AfterEach(func() {
						Expect(testDataStoreSession.ActivateDatasetDataInputs).To(ConsistOf(testDataset))
					})

					It("returns an error if there is an error with ActivateDatasetData", func() {
						testDataStoreSession.ActivateDatasetDataOutputs = []error{errors.New("test error")}
						err := testDeduplicator.DeduplicateDataset()
						Expect(err).To(MatchError(fmt.Sprintf(`deduplicator: unable to activate dataset data with id "%s"; test error`, testDataset.UploadID)))
					})

					It("returns successfully if there is no error", func() {
						Expect(testDeduplicator.DeduplicateDataset()).To(Succeed())
					})
				})
			})

			Context("DeleteDataset", func() {
				Context("with deleting dataset", func() {
					BeforeEach(func() {
						testDataStoreSession.DeleteDatasetOutputs = []error{nil}
					})

					AfterEach(func() {
						Expect(testDataStoreSession.DeleteDatasetInputs).To(ConsistOf(testDataset))
					})

					It("returns an error if there is an error with DeleteDataset", func() {
						testDataStoreSession.DeleteDatasetOutputs = []error{errors.New("test error")}
						err := testDeduplicator.DeleteDataset()
						Expect(err).To(MatchError(fmt.Sprintf(`deduplicator: unable to delete dataset with id "%s"; test error`, testDataset.UploadID)))
					})

					It("returns successfully if there is no error", func() {
						Expect(testDeduplicator.DeleteDataset()).To(Succeed())
					})
				})
			})
		})
	})
})
