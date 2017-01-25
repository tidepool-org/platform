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
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/log"
)

var _ = Describe("AfterDeactivateOld", func() {
	Context("NewAfterDeactivateOldFactory", func() {
		It("returns a new factory", func() {
			Expect(deduplicator.NewAfterDeactivateOldFactory()).ToNot(BeNil())
		})
	})

	Context("with a new factory", func() {
		var testFactory deduplicator.Factory
		var testDataset *upload.Upload

		BeforeEach(func() {
			var err error
			testFactory, err = deduplicator.NewAfterDeactivateOldFactory()
			Expect(err).ToNot(HaveOccurred())
			Expect(testFactory).ToNot(BeNil())
			testDataset = upload.Init()
			Expect(testDataset).ToNot(BeNil())
			testDataset.UserID = app.NewID()
			testDataset.GroupID = app.NewID()
			testDataset.DeviceID = app.StringAsPointer(app.NewID())
			testDataset.DeviceManufacturers = app.StringArrayAsPointer([]string{"UNUSED"})
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

			It("returns false if the device id is missing", func() {
				testDataset.DeviceID = nil
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the device id is empty", func() {
				testDataset.DeviceID = app.StringAsPointer("")
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the device manufacturers is missing", func() {
				testDataset.DeviceManufacturers = nil
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the device manufacturers is empty", func() {
				testDataset.DeviceManufacturers = app.StringArrayAsPointer([]string{})
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the device manufacturers does not contain expected device manufacturer", func() {
				testDataset.DeviceManufacturers = app.StringArrayAsPointer([]string{"Ant", "Zebra", "Cobra"})
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns true if the device id and expected device manufacturer are specified", func() {
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeTrue())
			})

			It("returns true if the device id and expected device manufacturer are specified with multiple device manufacturers", func() {
				testDataset.DeviceManufacturers = app.StringArrayAsPointer([]string{"Ant", "Zebra", "UNUSED", "Cobra"})
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

				It("returns an error if the dataset device id is missing", func() {
					testDataset.DeviceID = nil
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)
					Expect(err).To(MatchError("deduplicator: dataset device id is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the dataset device id is empty", func() {
					testDataset.DeviceID = app.StringAsPointer("")
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)
					Expect(err).To(MatchError("deduplicator: dataset device id is empty"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the device manufacturers is missing", func() {
					testDataset.DeviceManufacturers = nil
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)
					Expect(err).To(MatchError("deduplicator: dataset device manufacturers is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the device manufacturers is empty", func() {
					testDataset.DeviceManufacturers = app.StringArrayAsPointer([]string{})
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)
					Expect(err).To(MatchError("deduplicator: dataset device manufacturers does not contain expected device manufacturers"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the device manufacturers does not contain expected device manufacturer", func() {
					testDataset.DeviceManufacturers = app.StringArrayAsPointer([]string{"Ant", "Zebra", "Cobra"})
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)
					Expect(err).To(MatchError("deduplicator: dataset device manufacturers does not contain expected device manufacturers"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns a new deduplicator upon success", func() {
					Expect(testFactory.NewDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)).ToNot(BeNil())
				})

				It("returns a new deduplicator upon success if the device id and expected device manufacturer are specified with multiple device manufacturers", func() {
					testDataset.DeviceManufacturers = app.StringArrayAsPointer([]string{"Ant", "Zebra", "UNUSED", "Cobra"})
					Expect(testFactory.NewDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)).ToNot(BeNil())
				})
			})

			Context("with a new deduplicator", func() {
				var testDeduplicator data.Deduplicator

				BeforeEach(func() {
					var err error
					testDeduplicator, err = testFactory.NewDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)
					Expect(err).ToNot(HaveOccurred())
					Expect(testDeduplicator).ToNot(BeNil())
				})

				Context("DeduplicateDataset", func() {
					Context("with finding earliest dataset data time", func() {
						BeforeEach(func() {
							testDataStoreSession.FindEarliestDatasetDataTimeOutputs = []testDataStore.FindEarliestDatasetDataTimeOutput{{Time: "test-after-time", Error: nil}}
						})

						AfterEach(func() {
							Expect(testDataStoreSession.FindEarliestDatasetDataTimeInputs).To(ConsistOf(testDataset))
						})

						It("returns an error if there is an error with FindEarliestDatasetDataTime", func() {
							testDataStoreSession.FindEarliestDatasetDataTimeOutputs = []testDataStore.FindEarliestDatasetDataTimeOutput{{Time: "", Error: errors.New("test error")}}
							err := testDeduplicator.DeduplicateDataset()
							Expect(err).To(MatchError(fmt.Sprintf(`deduplicator: unable to get earliest data in dataset with id "%s"; test error`, testDataset.UploadID)))
						})

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

							Context("with deactivating other dataset data after time", func() {
								BeforeEach(func() {
									testDataStoreSession.DeactivateOtherDatasetDataAfterTimeOutputs = []error{nil}
								})

								AfterEach(func() {
									Expect(testDataStoreSession.DeactivateOtherDatasetDataAfterTimeInputs).To(ConsistOf([]testDataStore.DeactivateOtherDatasetDataAfterTimeInput{{Dataset: testDataset, Time: "test-after-time"}}))
								})

								It("returns an error if there is an error with DeactivateOtherDatasetDataAfterTime", func() {
									testDataStoreSession.DeactivateOtherDatasetDataAfterTimeOutputs = []error{errors.New("test error")}
									err := testDeduplicator.DeduplicateDataset()
									Expect(err).To(MatchError(fmt.Sprintf(`deduplicator: unable to remove all other data except dataset with id "%s"; test error`, testDataset.UploadID)))
								})

								It("returns successfully if there is no error", func() {
									Expect(testDeduplicator.DeduplicateDataset()).To(Succeed())
								})
							})

							Context("without deactivating other dataset data after time", func() {
								BeforeEach(func() {
									testDataStoreSession.FindEarliestDatasetDataTimeOutputs = []testDataStore.FindEarliestDatasetDataTimeOutput{{Time: "", Error: nil}}
								})

								It("returns successfully if there is no error", func() {
									Expect(testDeduplicator.DeduplicateDataset()).To(Succeed())
								})
							})
						})
					})
				})
			})
		})
	})
})
