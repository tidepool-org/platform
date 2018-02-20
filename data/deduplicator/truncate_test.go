package deduplicator_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	"fmt"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	testDataStoreDEPRECATED "github.com/tidepool-org/platform/data/storeDEPRECATED/test"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
)

var _ = Describe("Truncate", func() {
	Context("NewTruncateFactory", func() {
		It("returns a new factory", func() {
			Expect(deduplicator.NewTruncateFactory()).ToNot(BeNil())
		})
	})

	Context("with a new factory", func() {
		var testFactory deduplicator.Factory
		var testUploadID string
		var testUserID string
		var testDataset *upload.Upload

		BeforeEach(func() {
			var err error
			testFactory, err = deduplicator.NewTruncateFactory()
			Expect(err).ToNot(HaveOccurred())
			Expect(testFactory).ToNot(BeNil())
			testUploadID = id.New()
			testUserID = id.New()
			testDataset = upload.Init()
			Expect(testDataset).ToNot(BeNil())
			testDataset.UploadID = &testUploadID
			testDataset.UserID = &testUserID
			testDataset.DeviceID = pointer.String(id.New())
			testDataset.DeviceManufacturers = pointer.StringArray([]string{"Animas"})
		})

		Context("CanDeduplicateDataset", func() {
			It("returns an error if the dataset is missing", func() {
				can, err := testFactory.CanDeduplicateDataset(nil)
				Expect(err).To(MatchError("dataset is missing"))
				Expect(can).To(BeFalse())
			})

			It("returns false if the dataset id is missing", func() {
				testDataset.UploadID = nil
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the dataset id is empty", func() {
				testDataset.UploadID = pointer.String("")
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the dataset user id is missing", func() {
				testDataset.UserID = nil
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the dataset user id is empty", func() {
				testDataset.UserID = pointer.String("")
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the device id is missing", func() {
				testDataset.DeviceID = nil
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the device id is empty", func() {
				testDataset.DeviceID = pointer.String("")
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the device manufacturers is missing", func() {
				testDataset.DeviceManufacturers = nil
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the device manufacturers is empty", func() {
				testDataset.DeviceManufacturers = pointer.StringArray([]string{})
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the device manufacturers does not contain expected device manufacturer", func() {
				testDataset.DeviceManufacturers = pointer.StringArray([]string{"Ant", "Zebra", "Cobra"})
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns true if the device id and expected device manufacturer are specified", func() {
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeTrue())
			})

			It("returns true if the device id and expected device manufacturer are specified with multiple device manufacturers", func() {
				testDataset.DeviceManufacturers = pointer.StringArray([]string{"Ant", "Zebra", "Animas", "Cobra"})
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeTrue())
			})
		})

		Context("with logger and data store session", func() {
			var testLogger log.Logger
			var testDataSession *testDataStoreDEPRECATED.DataSession

			BeforeEach(func() {
				testLogger = null.NewLogger()
				Expect(testLogger).ToNot(BeNil())
				testDataSession = testDataStoreDEPRECATED.NewDataSession()
				Expect(testDataSession).ToNot(BeNil())
			})

			AfterEach(func() {
				testDataSession.Expectations()
			})

			Context("NewDeduplicatorForDataset", func() {
				It("returns an error if the logger is missing", func() {
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(nil, testDataSession, testDataset)
					Expect(err).To(MatchError("logger is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data store session is missing", func() {
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, nil, testDataset)
					Expect(err).To(MatchError("data store session is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the dataset is missing", func() {
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, nil)
					Expect(err).To(MatchError("dataset is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the dataset id is missing", func() {
					testDataset.UploadID = nil
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("dataset id is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the dataset id is empty", func() {
					testDataset.UploadID = pointer.String("")
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("dataset id is empty"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the dataset user id is missing", func() {
					testDataset.UserID = nil
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("dataset user id is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the dataset user id is empty", func() {
					testDataset.UserID = pointer.String("")
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("dataset user id is empty"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the dataset device id is missing", func() {
					testDataset.DeviceID = nil
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("dataset device id is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the dataset device id is empty", func() {
					testDataset.DeviceID = pointer.String("")
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("dataset device id is empty"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the device manufacturers is missing", func() {
					testDataset.DeviceManufacturers = nil
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("dataset device manufacturers does not contain expected device manufacturers"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the device manufacturers is empty", func() {
					testDataset.DeviceManufacturers = pointer.StringArray([]string{})
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("dataset device manufacturers does not contain expected device manufacturers"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the device manufacturers does not contain expected device manufacturer", func() {
					testDataset.DeviceManufacturers = pointer.StringArray([]string{"Ant", "Zebra", "Cobra"})
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("dataset device manufacturers does not contain expected device manufacturers"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns a new deduplicator upon success", func() {
					Expect(testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)).ToNot(BeNil())
				})

				It("returns a new deduplicator upon success if the device id and expected device manufacturer are specified with multiple device manufacturers", func() {
					testDataset.DeviceManufacturers = pointer.StringArray([]string{"Ant", "Zebra", "Animas", "Cobra"})
					Expect(testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)).ToNot(BeNil())
				})
			})

			Context("with a context and new deduplicator", func() {
				var ctx context.Context
				var testDeduplicator data.Deduplicator

				BeforeEach(func() {
					ctx = context.Background()
					var err error
					testDeduplicator, err = testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).ToNot(HaveOccurred())
					Expect(testDeduplicator).ToNot(BeNil())
				})

				Context("DeduplicateDataset", func() {
					Context("with activating dataset data", func() {
						BeforeEach(func() {
							testDataSession.ActivateDatasetDataOutputs = []error{nil}
						})

						AfterEach(func() {
							Expect(testDataSession.ActivateDatasetDataInputs).To(ConsistOf(testDataStoreDEPRECATED.ActivateDatasetDataInput{Context: ctx, Dataset: testDataset}))
						})

						It("returns an error if there is an error with ActivateDatasetData", func() {
							testDataSession.ActivateDatasetDataOutputs = []error{errors.New("test error")}
							err := testDeduplicator.DeduplicateDataset(ctx)
							Expect(err).To(MatchError(fmt.Sprintf("unable to activate dataset data with id %q; test error", testUploadID)))
						})

						Context("with deleting other dataset data", func() {
							BeforeEach(func() {
								testDataSession.DeleteOtherDatasetDataOutputs = []error{nil}
							})

							AfterEach(func() {
								Expect(testDataSession.DeleteOtherDatasetDataInputs).To(ConsistOf(testDataStoreDEPRECATED.DeleteOtherDatasetDataInput{Context: ctx, Dataset: testDataset}))
							})

							It("returns an error if there is an error with DeleteOtherDatasetData", func() {
								testDataSession.DeleteOtherDatasetDataOutputs = []error{errors.New("test error")}
								err := testDeduplicator.DeduplicateDataset(ctx)
								Expect(err).To(MatchError(fmt.Sprintf("unable to remove all other data except dataset with id %q; test error", testUploadID)))
							})

							It("returns successfully if there is no error", func() {
								Expect(testDeduplicator.DeduplicateDataset(ctx)).To(Succeed())
							})
						})
					})
				})
			})
		})
	})
})
