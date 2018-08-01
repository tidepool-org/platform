package deduplicator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	testDataDeduplicator "github.com/tidepool-org/platform/data/deduplicator/test"
	testDataStoreDEPRECATED "github.com/tidepool-org/platform/data/storeDEPRECATED/test"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Delegate", func() {
	Context("NewDelegateFactory", func() {
		It("returns an error if factories is nil", func() {
			testFactory, err := deduplicator.NewDelegateFactory(nil)
			Expect(err).To(MatchError("factories is missing"))
			Expect(testFactory).To(BeNil())
		})

		It("returns an error if there are no factories", func() {
			testFactory, err := deduplicator.NewDelegateFactory([]deduplicator.Factory{})
			Expect(err).To(MatchError("factories is missing"))
			Expect(testFactory).To(BeNil())
		})

		It("returns success with one factory", func() {
			Expect(deduplicator.NewDelegateFactory([]deduplicator.Factory{testDataDeduplicator.NewFactory()})).ToNot(BeNil())
		})

		It("returns success with multiple factories", func() {
			Expect(deduplicator.NewDelegateFactory([]deduplicator.Factory{testDataDeduplicator.NewFactory(), testDataDeduplicator.NewFactory(), testDataDeduplicator.NewFactory(), testDataDeduplicator.NewFactory()})).ToNot(BeNil())
		})
	})

	Context("with a new delegate factory", func() {
		var testFirstFactory *testDataDeduplicator.Factory
		var testSecondFactory *testDataDeduplicator.Factory
		var testDelegateFactory deduplicator.Factory
		var testDataSet *upload.Upload

		BeforeEach(func() {
			var err error
			testFirstFactory = testDataDeduplicator.NewFactory()
			testSecondFactory = testDataDeduplicator.NewFactory()
			testDelegateFactory, err = deduplicator.NewDelegateFactory([]deduplicator.Factory{testFirstFactory, testSecondFactory})
			Expect(err).ToNot(HaveOccurred())
			Expect(testDelegateFactory).ToNot(BeNil())
			testDataSet = upload.New()
			Expect(testDataSet).ToNot(BeNil())
		})

		AfterEach(func() {
			Expect(testSecondFactory.UnusedOutputsCount()).To(Equal(0))
			Expect(testFirstFactory.UnusedOutputsCount()).To(Equal(0))
		})

		Context("with unregistered data set", func() {
			BeforeEach(func() {
				testFirstFactory.CanDeduplicateDataSetOutputs = []testDataDeduplicator.CanDeduplicateDataSetOutput{{Can: false, Error: nil}}
				testSecondFactory.CanDeduplicateDataSetOutputs = []testDataDeduplicator.CanDeduplicateDataSetOutput{{Can: false, Error: nil}}
			})

			Context("CanDeduplicateDataSet", func() {
				It("returns an error if the data set is missing", func() {
					testFirstFactory.CanDeduplicateDataSetOutputs = []testDataDeduplicator.CanDeduplicateDataSetOutput{}
					testSecondFactory.CanDeduplicateDataSetOutputs = []testDataDeduplicator.CanDeduplicateDataSetOutput{}
					can, err := testDelegateFactory.CanDeduplicateDataSet(nil)
					Expect(err).To(MatchError("data set is missing"))
					Expect(can).To(BeFalse())
				})

				It("returns an error if any contained factory returns an error", func() {
					testSecondFactory.CanDeduplicateDataSetOutputs = []testDataDeduplicator.CanDeduplicateDataSetOutput{{Can: false, Error: errors.New("test error")}}
					can, err := testDelegateFactory.CanDeduplicateDataSet(testDataSet)
					Expect(err).To(MatchError("test error"))
					Expect(can).To(BeFalse())
					Expect(testFirstFactory.CanDeduplicateDataSetInputs).To(ConsistOf(testDataSet))
					Expect(testSecondFactory.CanDeduplicateDataSetInputs).To(ConsistOf(testDataSet))
				})

				It("return false if no factory can deduplicate the data set", func() {
					Expect(testDelegateFactory.CanDeduplicateDataSet(testDataSet)).To(BeFalse())
					Expect(testFirstFactory.CanDeduplicateDataSetInputs).To(ConsistOf(testDataSet))
					Expect(testSecondFactory.CanDeduplicateDataSetInputs).To(ConsistOf(testDataSet))
				})

				It("returns true if any contained factory can deduplicate the data set", func() {
					testSecondFactory.CanDeduplicateDataSetOutputs = []testDataDeduplicator.CanDeduplicateDataSetOutput{{Can: true, Error: nil}}
					Expect(testDelegateFactory.CanDeduplicateDataSet(testDataSet)).To(BeTrue())
					Expect(testFirstFactory.CanDeduplicateDataSetInputs).To(ConsistOf(testDataSet))
					Expect(testSecondFactory.CanDeduplicateDataSetInputs).To(ConsistOf(testDataSet))
				})

				It("returns true if any contained factory can deduplicate the data set even if a later factory returns an error", func() {
					testFirstFactory.CanDeduplicateDataSetOutputs = []testDataDeduplicator.CanDeduplicateDataSetOutput{{Can: true, Error: nil}}
					testSecondFactory.CanDeduplicateDataSetOutputs = []testDataDeduplicator.CanDeduplicateDataSetOutput{}
					Expect(testDelegateFactory.CanDeduplicateDataSet(testDataSet)).To(BeTrue())
					Expect(testFirstFactory.CanDeduplicateDataSetInputs).To(ConsistOf(testDataSet))
					Expect(testSecondFactory.CanDeduplicateDataSetInputs).To(BeEmpty())
				})
			})

			Context("NewDeduplicatorForDataSet", func() {
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

				It("returns an error if the logger is missing", func() {
					testFirstFactory.CanDeduplicateDataSetOutputs = []testDataDeduplicator.CanDeduplicateDataSetOutput{}
					testSecondFactory.CanDeduplicateDataSetOutputs = []testDataDeduplicator.CanDeduplicateDataSetOutput{}
					testDeduplicator, err := testDelegateFactory.NewDeduplicatorForDataSet(nil, testDataSession, testDataSet)
					Expect(err).To(MatchError("logger is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data store session is missing", func() {
					testFirstFactory.CanDeduplicateDataSetOutputs = []testDataDeduplicator.CanDeduplicateDataSetOutput{}
					testSecondFactory.CanDeduplicateDataSetOutputs = []testDataDeduplicator.CanDeduplicateDataSetOutput{}
					testDeduplicator, err := testDelegateFactory.NewDeduplicatorForDataSet(testLogger, nil, testDataSet)
					Expect(err).To(MatchError("data store session is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data set is missing", func() {
					testFirstFactory.CanDeduplicateDataSetOutputs = []testDataDeduplicator.CanDeduplicateDataSetOutput{}
					testSecondFactory.CanDeduplicateDataSetOutputs = []testDataDeduplicator.CanDeduplicateDataSetOutput{}
					testDeduplicator, err := testDelegateFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, nil)
					Expect(err).To(MatchError("data set is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				When("the deduplicator name is specified in the data set", func() {
					BeforeEach(func() {
						testDataSet.Deduplicator = &data.DeduplicatorDescriptor{
							Name: test.RandomString(),
						}
						testFirstFactory.CanDeduplicateDataSetOutputs = []testDataDeduplicator.CanDeduplicateDataSetOutput{}
						testSecondFactory.CanDeduplicateDataSetOutputs = []testDataDeduplicator.CanDeduplicateDataSetOutput{}
					})

					It("returns an error if the name is missing", func() {
						testDataSet.Deduplicator.Name = ""
						testDeduplicator, err := testDelegateFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
						Expect(err).To(MatchError("data set deduplicator name is missing"))
						Expect(testDeduplicator).To(BeNil())
					})

					It("returns an error if any contained factory does not match the deduplicator", func() {
						testFirstFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{{Is: false, Error: nil}}
						testSecondFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{{Is: false, Error: nil}}
						testDeduplicator, err := testDelegateFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
						Expect(err).To(MatchError("data set deduplicator name is unknown"))
						Expect(testDeduplicator).To(BeNil())
						Expect(testFirstFactory.IsRegisteredWithDataSetInputs).To(ConsistOf(testDataSet))
						Expect(testSecondFactory.IsRegisteredWithDataSetInputs).To(ConsistOf(testDataSet))
					})

					It("returns an error if any contained factory returns an error from IsRegisteredWithDataSet", func() {
						testFirstFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{{Is: false, Error: nil}}
						testSecondFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{{Is: false, Error: errors.New("test error")}}
						testDeduplicator, err := testDelegateFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
						Expect(err).To(MatchError("test error"))
						Expect(testDeduplicator).To(BeNil())
						Expect(testFirstFactory.IsRegisteredWithDataSetInputs).To(ConsistOf(testDataSet))
						Expect(testSecondFactory.IsRegisteredWithDataSetInputs).To(ConsistOf(testDataSet))
					})

					It("returns an error if any contained factory returns an error from NewDeduplicatorForDataSet", func() {
						testFirstFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{{Is: false, Error: nil}}
						testSecondFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{{Is: true, Error: nil}}
						testSecondFactory.NewDeduplicatorForDataSetOutputs = []testDataDeduplicator.NewDeduplicatorForDataSetOutput{{Deduplicator: nil, Error: errors.New("test error")}}
						testDeduplicator, err := testDelegateFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
						Expect(err).To(MatchError("test error"))
						Expect(testDeduplicator).To(BeNil())
						Expect(testFirstFactory.IsRegisteredWithDataSetInputs).To(ConsistOf(testDataSet))
						Expect(testSecondFactory.IsRegisteredWithDataSetInputs).To(ConsistOf(testDataSet))
						Expect(testSecondFactory.NewDeduplicatorForDataSetInputs).To(ConsistOf(testDataDeduplicator.NewDeduplicatorForDataSetInput{Logger: testLogger, DataSession: testDataSession, DataSet: testDataSet}))
					})

					It("returns the deduplicator from a matching factory", func() {
						secondDeduplicator := testData.NewDeduplicator()
						testFirstFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{{Is: false, Error: nil}}
						testSecondFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{{Is: true, Error: nil}}
						testSecondFactory.NewDeduplicatorForDataSetOutputs = []testDataDeduplicator.NewDeduplicatorForDataSetOutput{{Deduplicator: secondDeduplicator, Error: nil}}
						testDeduplicator, err := testDelegateFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
						Expect(err).ToNot(HaveOccurred())
						Expect(testDeduplicator).To(Equal(secondDeduplicator))
						Expect(testFirstFactory.IsRegisteredWithDataSetInputs).To(ConsistOf(testDataSet))
						Expect(testSecondFactory.IsRegisteredWithDataSetInputs).To(ConsistOf(testDataSet))
						Expect(testSecondFactory.NewDeduplicatorForDataSetInputs).To(ConsistOf(testDataDeduplicator.NewDeduplicatorForDataSetInput{Logger: testLogger, DataSession: testDataSession, DataSet: testDataSet}))
					})
				})

				When("the deduplicator is not specified in the data set", func() {
					It("returns an error if any contained factory returns an error", func() {
						testSecondFactory.CanDeduplicateDataSetOutputs = []testDataDeduplicator.CanDeduplicateDataSetOutput{{Can: false, Error: errors.New("test error")}}
						testDeduplicator, err := testDelegateFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
						Expect(err).To(MatchError("test error"))
						Expect(testDeduplicator).To(BeNil())
						Expect(testFirstFactory.CanDeduplicateDataSetInputs).To(ConsistOf(testDataSet))
						Expect(testSecondFactory.CanDeduplicateDataSetInputs).To(ConsistOf(testDataSet))
					})

					It("returns an error if no factory can deduplicate the data set", func() {
						testDeduplicator, err := testDelegateFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
						Expect(err).To(MatchError("deduplicator not found"))
						Expect(testDeduplicator).To(BeNil())
						Expect(testFirstFactory.CanDeduplicateDataSetInputs).To(ConsistOf(testDataSet))
						Expect(testSecondFactory.CanDeduplicateDataSetInputs).To(ConsistOf(testDataSet))
					})

					It("returns a deduplicator if any contained factory can deduplicate the data set", func() {
						secondDeduplicator := testData.NewDeduplicator()
						testSecondFactory.CanDeduplicateDataSetOutputs = []testDataDeduplicator.CanDeduplicateDataSetOutput{{Can: true, Error: nil}}
						testSecondFactory.NewDeduplicatorForDataSetOutputs = []testDataDeduplicator.NewDeduplicatorForDataSetOutput{{Deduplicator: secondDeduplicator, Error: nil}}
						Expect(testDelegateFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)).To(Equal(secondDeduplicator))
						Expect(testFirstFactory.CanDeduplicateDataSetInputs).To(ConsistOf(testDataSet))
						Expect(testSecondFactory.CanDeduplicateDataSetInputs).To(ConsistOf(testDataSet))
					})

					It("returns a deduplicator if any contained factory can deduplicate the data set even if a later factory returns an error", func() {
						firstDeduplicator := testData.NewDeduplicator()
						testFirstFactory.CanDeduplicateDataSetOutputs = []testDataDeduplicator.CanDeduplicateDataSetOutput{{Can: true, Error: nil}}
						testFirstFactory.NewDeduplicatorForDataSetOutputs = []testDataDeduplicator.NewDeduplicatorForDataSetOutput{{Deduplicator: firstDeduplicator, Error: nil}}
						testSecondFactory.CanDeduplicateDataSetOutputs = []testDataDeduplicator.CanDeduplicateDataSetOutput{}
						Expect(testDelegateFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)).To(Equal(firstDeduplicator))
						Expect(testFirstFactory.CanDeduplicateDataSetInputs).To(ConsistOf(testDataSet))
						Expect(testFirstFactory.NewDeduplicatorForDataSetInputs).To(ConsistOf(testDataDeduplicator.NewDeduplicatorForDataSetInput{Logger: testLogger, DataSession: testDataSession, DataSet: testDataSet}))
					})

					It("returns an error if any contained factory can deduplicate the data set, but returns an error when creating", func() {
						testSecondFactory.CanDeduplicateDataSetOutputs = []testDataDeduplicator.CanDeduplicateDataSetOutput{{Can: true, Error: nil}}
						testSecondFactory.NewDeduplicatorForDataSetOutputs = []testDataDeduplicator.NewDeduplicatorForDataSetOutput{{Deduplicator: nil, Error: errors.New("test error")}}
						testDeduplicator, err := testDelegateFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
						Expect(err).To(MatchError("test error"))
						Expect(testDeduplicator).To(BeNil())
						Expect(testFirstFactory.CanDeduplicateDataSetInputs).To(ConsistOf(testDataSet))
						Expect(testSecondFactory.CanDeduplicateDataSetInputs).To(ConsistOf(testDataSet))
						Expect(testSecondFactory.NewDeduplicatorForDataSetInputs).To(ConsistOf(testDataDeduplicator.NewDeduplicatorForDataSetInput{Logger: testLogger, DataSession: testDataSession, DataSet: testDataSet}))
					})
				})
			})
		})

		Context("with registered data set", func() {
			BeforeEach(func() {
				testFirstFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{{Is: false, Error: nil}}
				testSecondFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{{Is: false, Error: nil}}
				testDataSet.SetDeduplicatorDescriptor(&data.DeduplicatorDescriptor{Name: "test"})
			})

			Context("IsRegisteredWithDataSet", func() {
				It("returns an error if the data set is missing", func() {
					testFirstFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
					testSecondFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
					can, err := testDelegateFactory.IsRegisteredWithDataSet(nil)
					Expect(err).To(MatchError("data set is missing"))
					Expect(can).To(BeFalse())
				})

				It("returns an error if any contained factory returns an error", func() {
					testSecondFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{{Is: false, Error: errors.New("test error")}}
					can, err := testDelegateFactory.IsRegisteredWithDataSet(testDataSet)
					Expect(err).To(MatchError("test error"))
					Expect(can).To(BeFalse())
					Expect(testFirstFactory.IsRegisteredWithDataSetInputs).To(ConsistOf(testDataSet))
					Expect(testSecondFactory.IsRegisteredWithDataSetInputs).To(ConsistOf(testDataSet))
				})

				It("return false if no factory is registered with the data set", func() {
					Expect(testDelegateFactory.IsRegisteredWithDataSet(testDataSet)).To(BeFalse())
					Expect(testFirstFactory.IsRegisteredWithDataSetInputs).To(ConsistOf(testDataSet))
					Expect(testSecondFactory.IsRegisteredWithDataSetInputs).To(ConsistOf(testDataSet))
				})

				It("returns true if any contained factory is registered with the data set", func() {
					testSecondFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{{Is: true, Error: nil}}
					Expect(testDelegateFactory.IsRegisteredWithDataSet(testDataSet)).To(BeTrue())
					Expect(testFirstFactory.IsRegisteredWithDataSetInputs).To(ConsistOf(testDataSet))
					Expect(testSecondFactory.IsRegisteredWithDataSetInputs).To(ConsistOf(testDataSet))
				})

				It("returns true if any contained factory is registered with the data set even if a later factory returns an error", func() {
					testFirstFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{{Is: true, Error: nil}}
					testSecondFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
					Expect(testDelegateFactory.IsRegisteredWithDataSet(testDataSet)).To(BeTrue())
					Expect(testFirstFactory.IsRegisteredWithDataSetInputs).To(ConsistOf(testDataSet))
					Expect(testSecondFactory.IsRegisteredWithDataSetInputs).To(BeEmpty())
				})
			})

			Context("NewRegisteredDeduplicatorForDataSet", func() {
				var testLogger log.Logger
				var testDataSession *testDataStoreDEPRECATED.DataSession

				BeforeEach(func() {
					testLogger = null.NewLogger()
					testDataSession = testDataStoreDEPRECATED.NewDataSession()
					Expect(testDataSession).ToNot(BeNil())
				})

				AfterEach(func() {
					testDataSession.Expectations()
				})

				It("returns an error if the logger is missing", func() {
					testFirstFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
					testSecondFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
					testDeduplicator, err := testDelegateFactory.NewRegisteredDeduplicatorForDataSet(nil, testDataSession, testDataSet)
					Expect(err).To(MatchError("logger is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data store session is missing", func() {
					testFirstFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
					testSecondFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
					testDeduplicator, err := testDelegateFactory.NewRegisteredDeduplicatorForDataSet(testLogger, nil, testDataSet)
					Expect(err).To(MatchError("data store session is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data set is missing", func() {
					testFirstFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
					testSecondFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
					testDeduplicator, err := testDelegateFactory.NewRegisteredDeduplicatorForDataSet(testLogger, testDataSession, nil)
					Expect(err).To(MatchError("data set is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data set is not registered with a deduplicator", func() {
					testFirstFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
					testSecondFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
					testDataSet.SetDeduplicatorDescriptor(nil)
					testDeduplicator, err := testDelegateFactory.NewRegisteredDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("data set not registered with deduplicator"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data set is not registered with a deduplicator with a name", func() {
					testFirstFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
					testSecondFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
					testDataSet.SetDeduplicatorDescriptor(&data.DeduplicatorDescriptor{})
					testDeduplicator, err := testDelegateFactory.NewRegisteredDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("data set not registered with deduplicator"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if any contained factory returns an error", func() {
					testSecondFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{{Is: false, Error: errors.New("test error")}}
					testDeduplicator, err := testDelegateFactory.NewRegisteredDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("test error"))
					Expect(testDeduplicator).To(BeNil())
					Expect(testFirstFactory.IsRegisteredWithDataSetInputs).To(ConsistOf(testDataSet))
					Expect(testSecondFactory.IsRegisteredWithDataSetInputs).To(ConsistOf(testDataSet))
				})

				It("returns an error if no factory is registered with the data set", func() {
					testDeduplicator, err := testDelegateFactory.NewRegisteredDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("deduplicator not found"))
					Expect(testDeduplicator).To(BeNil())
					Expect(testFirstFactory.IsRegisteredWithDataSetInputs).To(ConsistOf(testDataSet))
					Expect(testSecondFactory.IsRegisteredWithDataSetInputs).To(ConsistOf(testDataSet))
				})

				It("returns a deduplicator if any contained factory is registered with the data set", func() {
					secondDeduplicator := testData.NewDeduplicator()
					testSecondFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{{Is: true, Error: nil}}
					testSecondFactory.NewRegisteredDeduplicatorForDataSetOutputs = []testDataDeduplicator.NewRegisteredDeduplicatorForDataSetOutput{{Deduplicator: secondDeduplicator, Error: nil}}
					Expect(testDelegateFactory.NewRegisteredDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)).To(Equal(secondDeduplicator))
					Expect(testFirstFactory.IsRegisteredWithDataSetInputs).To(ConsistOf(testDataSet))
					Expect(testSecondFactory.IsRegisteredWithDataSetInputs).To(ConsistOf(testDataSet))
				})

				It("returns a deduplicator if any contained factory is registered with the data set even if a later factory returns an error", func() {
					firstDeduplicator := testData.NewDeduplicator()
					testFirstFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{{Is: true, Error: nil}}
					testFirstFactory.NewRegisteredDeduplicatorForDataSetOutputs = []testDataDeduplicator.NewRegisteredDeduplicatorForDataSetOutput{{Deduplicator: firstDeduplicator, Error: nil}}
					testSecondFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
					Expect(testDelegateFactory.NewRegisteredDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)).To(Equal(firstDeduplicator))
					Expect(testFirstFactory.IsRegisteredWithDataSetInputs).To(ConsistOf(testDataSet))
					Expect(testFirstFactory.NewRegisteredDeduplicatorForDataSetInputs).To(ConsistOf(testDataDeduplicator.NewRegisteredDeduplicatorForDataSetInput{Logger: testLogger, DataSession: testDataSession, DataSet: testDataSet}))
				})

				It("returns an error if any contained factory is registered with the data set, but returns an error when creating", func() {
					testSecondFactory.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{{Is: true, Error: nil}}
					testSecondFactory.NewRegisteredDeduplicatorForDataSetOutputs = []testDataDeduplicator.NewRegisteredDeduplicatorForDataSetOutput{{Deduplicator: nil, Error: errors.New("test error")}}
					testDeduplicator, err := testDelegateFactory.NewRegisteredDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("test error"))
					Expect(testDeduplicator).To(BeNil())
					Expect(testFirstFactory.IsRegisteredWithDataSetInputs).To(ConsistOf(testDataSet))
					Expect(testSecondFactory.IsRegisteredWithDataSetInputs).To(ConsistOf(testDataSet))
					Expect(testSecondFactory.NewRegisteredDeduplicatorForDataSetInputs).To(ConsistOf(testDataDeduplicator.NewRegisteredDeduplicatorForDataSetInput{Logger: testLogger, DataSession: testDataSession, DataSet: testDataSet}))
				})
			})
		})
	})
})
