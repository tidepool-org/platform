package deduplicator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"context"
	"errors"
	"fmt"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	testDataStoreDEPRECATED "github.com/tidepool-org/platform/data/storeDEPRECATED/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	userTest "github.com/tidepool-org/platform/user/test"
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
		var testDataSet *upload.Upload

		BeforeEach(func() {
			var err error
			testFactory, err = deduplicator.NewTruncateFactory()
			Expect(err).ToNot(HaveOccurred())
			Expect(testFactory).ToNot(BeNil())
			testUploadID = dataTest.RandomSetID()
			testUserID = userTest.RandomID()
			testDataSet = upload.New()
			Expect(testDataSet).ToNot(BeNil())
			testDataSet.UploadID = &testUploadID
			testDataSet.UserID = &testUserID
			testDataSet.DeviceID = pointer.FromString(dataTest.NewDeviceID())
			testDataSet.DeviceManufacturers = pointer.FromStringArray([]string{"Animas"})
		})

		Context("CanDeduplicateDataSet", func() {
			It("returns an error if the data set is missing", func() {
				can, err := testFactory.CanDeduplicateDataSet(nil)
				Expect(err).To(MatchError("data set is missing"))
				Expect(can).To(BeFalse())
			})

			It("returns false if the data set id is missing", func() {
				testDataSet.UploadID = nil
				Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeFalse())
			})

			It("returns false if the data set id is empty", func() {
				testDataSet.UploadID = pointer.FromString("")
				Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeFalse())
			})

			It("returns false if the data set user id is missing", func() {
				testDataSet.UserID = nil
				Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeFalse())
			})

			It("returns false if the data set user id is empty", func() {
				testDataSet.UserID = pointer.FromString("")
				Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeFalse())
			})

			It("returns false if the device id is missing", func() {
				testDataSet.DeviceID = nil
				Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeFalse())
			})

			It("returns false if the device id is empty", func() {
				testDataSet.DeviceID = pointer.FromString("")
				Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeFalse())
			})

			It("returns false if the device manufacturers is missing", func() {
				testDataSet.DeviceManufacturers = nil
				Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeFalse())
			})

			It("returns false if the device manufacturers is empty", func() {
				testDataSet.DeviceManufacturers = pointer.FromStringArray([]string{})
				Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeFalse())
			})

			It("returns false if the device manufacturers does not contain expected device manufacturer", func() {
				testDataSet.DeviceManufacturers = pointer.FromStringArray([]string{"Ant", "Zebra", "Cobra"})
				Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeFalse())
			})

			It("returns true if the device id and expected device manufacturer are specified", func() {
				Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeTrue())
			})

			It("returns true if the device id and expected device manufacturer are specified with multiple device manufacturers", func() {
				testDataSet.DeviceManufacturers = pointer.FromStringArray([]string{"Ant", "Zebra", "Animas", "Cobra"})
				Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeTrue())
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

			Context("NewDeduplicatorForDataSet", func() {
				It("returns an error if the logger is missing", func() {
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(nil, testDataSession, testDataSet)
					Expect(err).To(MatchError("logger is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data store session is missing", func() {
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, nil, testDataSet)
					Expect(err).To(MatchError("data store session is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data set is missing", func() {
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, nil)
					Expect(err).To(MatchError("data set is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data set id is missing", func() {
					testDataSet.UploadID = nil
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("data set id is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data set id is empty", func() {
					testDataSet.UploadID = pointer.FromString("")
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("data set id is empty"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data set user id is missing", func() {
					testDataSet.UserID = nil
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("data set user id is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data set user id is empty", func() {
					testDataSet.UserID = pointer.FromString("")
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("data set user id is empty"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data set device id is missing", func() {
					testDataSet.DeviceID = nil
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("data set device id is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data set device id is empty", func() {
					testDataSet.DeviceID = pointer.FromString("")
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("data set device id is empty"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the device manufacturers is missing", func() {
					testDataSet.DeviceManufacturers = nil
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("data set device manufacturers does not contain expected device manufacturers"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the device manufacturers is empty", func() {
					testDataSet.DeviceManufacturers = pointer.FromStringArray([]string{})
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("data set device manufacturers does not contain expected device manufacturers"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the device manufacturers does not contain expected device manufacturer", func() {
					testDataSet.DeviceManufacturers = pointer.FromStringArray([]string{"Ant", "Zebra", "Cobra"})
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("data set device manufacturers does not contain expected device manufacturers"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns a new deduplicator upon success", func() {
					Expect(testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)).ToNot(BeNil())
				})

				It("returns a new deduplicator upon success if the device id and expected device manufacturer are specified with multiple device manufacturers", func() {
					testDataSet.DeviceManufacturers = pointer.FromStringArray([]string{"Ant", "Zebra", "Animas", "Cobra"})
					Expect(testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)).ToNot(BeNil())
				})
			})

			Context("with a context and new deduplicator", func() {
				var ctx context.Context
				var testDeduplicator data.Deduplicator

				BeforeEach(func() {
					ctx = context.Background()
					var err error
					testDeduplicator, err = testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).ToNot(HaveOccurred())
					Expect(testDeduplicator).ToNot(BeNil())
				})

				Context("DeduplicateDataSet", func() {
					Context("with activating data set data", func() {
						BeforeEach(func() {
							testDataSession.ActivateDataSetDataOutputs = []error{nil}
						})

						AfterEach(func() {
							Expect(testDataSession.ActivateDataSetDataInputs).To(ConsistOf(testDataStoreDEPRECATED.ActivateDataSetDataInput{Context: ctx, DataSet: testDataSet}))
						})

						It("returns an error if there is an error with ActivateDataSetData", func() {
							testDataSession.ActivateDataSetDataOutputs = []error{errors.New("test error")}
							err := testDeduplicator.DeduplicateDataSet(ctx)
							Expect(err).To(MatchError(fmt.Sprintf("unable to activate data set data with id %q; test error", testUploadID)))
						})

						Context("with deleting other data set data", func() {
							BeforeEach(func() {
								testDataSession.DeleteOtherDataSetDataOutputs = []error{nil}
							})

							AfterEach(func() {
								Expect(testDataSession.DeleteOtherDataSetDataInputs).To(ConsistOf(testDataStoreDEPRECATED.DeleteOtherDataSetDataInput{Context: ctx, DataSet: testDataSet}))
							})

							It("returns an error if there is an error with DeleteOtherDataSetData", func() {
								testDataSession.DeleteOtherDataSetDataOutputs = []error{errors.New("test error")}
								err := testDeduplicator.DeduplicateDataSet(ctx)
								Expect(err).To(MatchError(fmt.Sprintf("unable to remove all other data except data set with id %q; test error", testUploadID)))
							})

							It("returns successfully if there is no error", func() {
								Expect(testDeduplicator.DeduplicateDataSet(ctx)).To(Succeed())
							})
						})
					})
				})
			})
		})
	})
})
