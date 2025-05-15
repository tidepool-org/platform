package factory_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataDeduplicatorFactory "github.com/tidepool-org/platform/data/deduplicator/factory"
	dataDeduplicatorFactoryTest "github.com/tidepool-org/platform/data/deduplicator/factory/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/pointer"
)

var _ = Describe("Factory", func() {
	var firstDeduplicator *dataDeduplicatorFactoryTest.Deduplicator
	var secondDeduplicator *dataDeduplicatorFactoryTest.Deduplicator
	var deduplicators []dataDeduplicatorFactory.Deduplicator

	BeforeEach(func() {
		firstDeduplicator = dataDeduplicatorFactoryTest.NewDeduplicator()
		secondDeduplicator = dataDeduplicatorFactoryTest.NewDeduplicator()
		deduplicators = []dataDeduplicatorFactory.Deduplicator{
			firstDeduplicator,
			secondDeduplicator,
		}
	})

	AfterEach(func() {
		secondDeduplicator.AssertOutputsEmpty()
		firstDeduplicator.AssertOutputsEmpty()
	})

	Context("New", func() {
		It("returns an error when the deduplicators is missing", func() {
			factory, err := dataDeduplicatorFactory.New(nil)
			Expect(err).To(MatchError("deduplicators is missing"))
			Expect(factory).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(dataDeduplicatorFactory.New(deduplicators)).ToNot(BeNil())
		})
	})

	Context("with a new factory", func() {
		var factory *dataDeduplicatorFactory.Factory
		var dataSet *data.DataSet
		var ctx context.Context

		BeforeEach(func() {
			var err error
			factory, err = dataDeduplicatorFactory.New(deduplicators)
			Expect(err).ToNot(HaveOccurred())
			Expect(factory).ToNot(BeNil())
			dataSet = dataTest.RandomDataSet()
			Expect(dataSet).ToNot(BeNil())
			ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
		})

		Context("New", func() {
			It("returns an error when the data set is missing", func() {
				deduplicator, err := factory.New(ctx, nil)
				Expect(err).To(MatchError("data set is missing"))
				Expect(deduplicator).To(BeNil())
			})

			When("the data set has a deduplicator name", func() {

				BeforeEach(func() {
					dataSet.Deduplicator = &data.DeduplicatorDescriptor{
						Name: pointer.FromString(netTest.RandomReverseDomain()),
					}
				})

				AfterEach(func() {
					Expect(secondDeduplicator.GetInputs).To(Equal([]dataDeduplicatorFactoryTest.GetInput{{Context: ctx, DataSet: dataSet}}))
					Expect(firstDeduplicator.GetInputs).To(Equal([]dataDeduplicatorFactoryTest.GetInput{{Context: ctx, DataSet: dataSet}}))
				})

				It("returns an error when a deduplicator returns an error", func() {
					responseErr := errorsTest.RandomError()
					firstDeduplicator.GetOutputs = []dataDeduplicatorFactoryTest.GetOutput{{Found: false, Error: nil}}
					secondDeduplicator.GetOutputs = []dataDeduplicatorFactoryTest.GetOutput{{Found: false, Error: responseErr}}
					deduplicator, err := factory.New(ctx, dataSet)
					Expect(err).To(Equal(responseErr))
					Expect(deduplicator).To(BeNil())
				})

				It("returns successfully when a deduplicator returns successfully", func() {
					firstDeduplicator.GetOutputs = []dataDeduplicatorFactoryTest.GetOutput{{Found: false, Error: nil}}
					secondDeduplicator.GetOutputs = []dataDeduplicatorFactoryTest.GetOutput{{Found: true, Error: nil}}
					Expect(factory.New(ctx, dataSet)).To(Equal(secondDeduplicator))
				})

				It("returns an error when no deduplicator returns successfully", func() {
					firstDeduplicator.GetOutputs = []dataDeduplicatorFactoryTest.GetOutput{{Found: false, Error: nil}}
					secondDeduplicator.GetOutputs = []dataDeduplicatorFactoryTest.GetOutput{{Found: false, Error: nil}}
					deduplicator, err := factory.New(ctx, dataSet)
					Expect(err).To(MatchError("deduplicator not found"))
					Expect(deduplicator).To(BeNil())
				})
			})

			newAssertions := func() {
				AfterEach(func() {
					Expect(secondDeduplicator.NewInputs).To(Equal([]dataDeduplicatorFactoryTest.NewInput{{Context: ctx, DataSet: dataSet}}))
					Expect(firstDeduplicator.NewInputs).To(Equal([]dataDeduplicatorFactoryTest.NewInput{{Context: ctx, DataSet: dataSet}}))
				})

				It("returns an error when a deduplicator returns an error", func() {
					responseErr := errorsTest.RandomError()
					firstDeduplicator.NewOutputs = []dataDeduplicatorFactoryTest.NewOutput{{Found: false, Error: nil}}
					secondDeduplicator.NewOutputs = []dataDeduplicatorFactoryTest.NewOutput{{Found: false, Error: responseErr}}
					deduplicator, err := factory.New(ctx, dataSet)
					Expect(err).To(Equal(responseErr))
					Expect(deduplicator).To(BeNil())
				})

				It("returns successfully when a deduplicator returns successfully", func() {
					firstDeduplicator.NewOutputs = []dataDeduplicatorFactoryTest.NewOutput{{Found: false, Error: nil}}
					secondDeduplicator.NewOutputs = []dataDeduplicatorFactoryTest.NewOutput{{Found: true, Error: nil}}
					Expect(factory.New(ctx, dataSet)).To(Equal(secondDeduplicator))
				})

				It("returns an error when no deduplicator returns successfully", func() {
					firstDeduplicator.NewOutputs = []dataDeduplicatorFactoryTest.NewOutput{{Found: false, Error: nil}}
					secondDeduplicator.NewOutputs = []dataDeduplicatorFactoryTest.NewOutput{{Found: false, Error: nil}}
					deduplicator, err := factory.New(ctx, dataSet)
					Expect(err).To(MatchError("deduplicator not found"))
					Expect(deduplicator).To(BeNil())
				})
			}

			When("the data set does not have a deduplicator", func() {
				BeforeEach(func() {
					dataSet.Deduplicator = nil
				})

				newAssertions()
			})

			When("the data set does not have a deduplicator name", func() {
				BeforeEach(func() {
					dataSet.Deduplicator.Name = nil
				})

				newAssertions()
			})
		})

		Context("Get", func() {
			It("returns an error when the data set is missing", func() {
				deduplicator, err := factory.Get(ctx, nil)
				Expect(err).To(MatchError("data set is missing"))
				Expect(deduplicator).To(BeNil())
			})

			When("the data set has a deduplicator name", func() {
				AfterEach(func() {
					Expect(secondDeduplicator.GetInputs).To(Equal([]dataDeduplicatorFactoryTest.GetInput{{Context: ctx, DataSet: dataSet}}))
					Expect(firstDeduplicator.GetInputs).To(Equal([]dataDeduplicatorFactoryTest.GetInput{{Context: ctx, DataSet: dataSet}}))
				})

				It("returns an error when a deduplicator returns an error", func() {
					responseErr := errorsTest.RandomError()
					firstDeduplicator.GetOutputs = []dataDeduplicatorFactoryTest.GetOutput{{Found: false, Error: nil}}
					secondDeduplicator.GetOutputs = []dataDeduplicatorFactoryTest.GetOutput{{Found: false, Error: responseErr}}
					deduplicator, err := factory.Get(ctx, dataSet)
					Expect(err).To(Equal(responseErr))
					Expect(deduplicator).To(BeNil())
				})

				It("returns successfully when a deduplicator returns successfully", func() {
					firstDeduplicator.GetOutputs = []dataDeduplicatorFactoryTest.GetOutput{{Found: false, Error: nil}}
					secondDeduplicator.GetOutputs = []dataDeduplicatorFactoryTest.GetOutput{{Found: true, Error: nil}}
					Expect(factory.Get(ctx, dataSet)).To(Equal(secondDeduplicator))
				})

				It("returns an error when no deduplicator returns successfully", func() {
					firstDeduplicator.GetOutputs = []dataDeduplicatorFactoryTest.GetOutput{{Found: false, Error: nil}}
					secondDeduplicator.GetOutputs = []dataDeduplicatorFactoryTest.GetOutput{{Found: false, Error: nil}}
					deduplicator, err := factory.Get(ctx, dataSet)
					Expect(err).To(MatchError("deduplicator not found"))
					Expect(deduplicator).To(BeNil())
				})
			})

			It("returns successfully without a deduplicator when the data set does not have a deduplicator", func() {
				dataSet.Deduplicator = nil
				Expect(factory.Get(ctx, dataSet)).To(BeNil())
			})

			It("returns successfully without a deduplicator  when the data set does not have a deduplicator name", func() {
				dataSet.Deduplicator.Name = nil
				Expect(factory.Get(ctx, dataSet)).To(BeNil())
			})
		})
	})
})
