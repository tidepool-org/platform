package factory_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataDeduplicatorFactory "github.com/tidepool-org/platform/data/deduplicator/factory"
	dataDeduplicatorFactoryTest "github.com/tidepool-org/platform/data/deduplicator/factory/test"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
	dataTypesUploadTest "github.com/tidepool-org/platform/data/types/upload/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
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
		var dataSet *dataTypesUpload.Upload

		BeforeEach(func() {
			var err error
			factory, err = dataDeduplicatorFactory.New(deduplicators)
			Expect(err).ToNot(HaveOccurred())
			Expect(factory).ToNot(BeNil())
			dataSet = dataTypesUploadTest.RandomUpload()
			Expect(dataSet).ToNot(BeNil())
		})

		Context("New", func() {
			It("returns an error when the data set is missing", func() {
				deduplicator, err := factory.New(nil)
				Expect(err).To(MatchError("data set is missing"))
				Expect(deduplicator).To(BeNil())
			})

			It("returns an error when the data set is invalid", func() {
				dataSet.DeviceModel = pointer.FromString("")
				deduplicator, err := factory.New(dataSet)
				Expect(err).To(MatchError("data set is invalid; value is empty"))
				Expect(deduplicator).To(BeNil())
			})

			When("the data set has a deduplicator name", func() {
				BeforeEach(func() {
					dataSet.Deduplicator = &data.DeduplicatorDescriptor{
						Name: pointer.FromString(netTest.RandomReverseDomain()),
					}
				})

				AfterEach(func() {
					Expect(secondDeduplicator.GetInputs).To(Equal([]*dataTypesUpload.Upload{dataSet}))
					Expect(firstDeduplicator.GetInputs).To(Equal([]*dataTypesUpload.Upload{dataSet}))
				})

				It("returns an error when a deduplicator returns an error", func() {
					responseErr := errorsTest.RandomError()
					firstDeduplicator.GetOutputs = []dataDeduplicatorFactoryTest.GetOutput{{Found: false, Error: nil}}
					secondDeduplicator.GetOutputs = []dataDeduplicatorFactoryTest.GetOutput{{Found: false, Error: responseErr}}
					deduplicator, err := factory.New(dataSet)
					Expect(err).To(Equal(responseErr))
					Expect(deduplicator).To(BeNil())
				})

				It("returns successfully when a deduplicator returns successfully", func() {
					firstDeduplicator.GetOutputs = []dataDeduplicatorFactoryTest.GetOutput{{Found: false, Error: nil}}
					secondDeduplicator.GetOutputs = []dataDeduplicatorFactoryTest.GetOutput{{Found: true, Error: nil}}
					Expect(factory.New(dataSet)).To(Equal(secondDeduplicator))
				})

				It("returns an error when no deduplicator returns successfully", func() {
					firstDeduplicator.GetOutputs = []dataDeduplicatorFactoryTest.GetOutput{{Found: false, Error: nil}}
					secondDeduplicator.GetOutputs = []dataDeduplicatorFactoryTest.GetOutput{{Found: false, Error: nil}}
					deduplicator, err := factory.New(dataSet)
					Expect(err).To(MatchError("deduplicator not found"))
					Expect(deduplicator).To(BeNil())
				})
			})

			newAssertions := func() {
				AfterEach(func() {
					Expect(secondDeduplicator.NewInputs).To(Equal([]*dataTypesUpload.Upload{dataSet}))
					Expect(firstDeduplicator.NewInputs).To(Equal([]*dataTypesUpload.Upload{dataSet}))
				})

				It("returns an error when a deduplicator returns an error", func() {
					responseErr := errorsTest.RandomError()
					firstDeduplicator.NewOutputs = []dataDeduplicatorFactoryTest.NewOutput{{Found: false, Error: nil}}
					secondDeduplicator.NewOutputs = []dataDeduplicatorFactoryTest.NewOutput{{Found: false, Error: responseErr}}
					deduplicator, err := factory.New(dataSet)
					Expect(err).To(Equal(responseErr))
					Expect(deduplicator).To(BeNil())
				})

				It("returns successfully when a deduplicator returns successfully", func() {
					firstDeduplicator.NewOutputs = []dataDeduplicatorFactoryTest.NewOutput{{Found: false, Error: nil}}
					secondDeduplicator.NewOutputs = []dataDeduplicatorFactoryTest.NewOutput{{Found: true, Error: nil}}
					Expect(factory.New(dataSet)).To(Equal(secondDeduplicator))
				})

				It("returns an error when no deduplicator returns successfully", func() {
					firstDeduplicator.NewOutputs = []dataDeduplicatorFactoryTest.NewOutput{{Found: false, Error: nil}}
					secondDeduplicator.NewOutputs = []dataDeduplicatorFactoryTest.NewOutput{{Found: false, Error: nil}}
					deduplicator, err := factory.New(dataSet)
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
				deduplicator, err := factory.Get(nil)
				Expect(err).To(MatchError("data set is missing"))
				Expect(deduplicator).To(BeNil())
			})

			It("returns an error when the data set is invalid", func() {
				dataSet.DeviceModel = pointer.FromString("")
				deduplicator, err := factory.Get(dataSet)
				Expect(err).To(MatchError("data set is invalid; value is empty"))
				Expect(deduplicator).To(BeNil())
			})

			When("the data set has a deduplicator name", func() {
				AfterEach(func() {
					Expect(secondDeduplicator.GetInputs).To(Equal([]*dataTypesUpload.Upload{dataSet}))
					Expect(firstDeduplicator.GetInputs).To(Equal([]*dataTypesUpload.Upload{dataSet}))
				})

				It("returns an error when a deduplicator returns an error", func() {
					responseErr := errorsTest.RandomError()
					firstDeduplicator.GetOutputs = []dataDeduplicatorFactoryTest.GetOutput{{Found: false, Error: nil}}
					secondDeduplicator.GetOutputs = []dataDeduplicatorFactoryTest.GetOutput{{Found: false, Error: responseErr}}
					deduplicator, err := factory.Get(dataSet)
					Expect(err).To(Equal(responseErr))
					Expect(deduplicator).To(BeNil())
				})

				It("returns successfully when a deduplicator returns successfully", func() {
					firstDeduplicator.GetOutputs = []dataDeduplicatorFactoryTest.GetOutput{{Found: false, Error: nil}}
					secondDeduplicator.GetOutputs = []dataDeduplicatorFactoryTest.GetOutput{{Found: true, Error: nil}}
					Expect(factory.Get(dataSet)).To(Equal(secondDeduplicator))
				})

				It("returns an error when no deduplicator returns successfully", func() {
					firstDeduplicator.GetOutputs = []dataDeduplicatorFactoryTest.GetOutput{{Found: false, Error: nil}}
					secondDeduplicator.GetOutputs = []dataDeduplicatorFactoryTest.GetOutput{{Found: false, Error: nil}}
					deduplicator, err := factory.Get(dataSet)
					Expect(err).To(MatchError("deduplicator not found"))
					Expect(deduplicator).To(BeNil())
				})
			})

			It("returns successfully without a deduplicator when the data set does not have a deduplicator", func() {
				dataSet.Deduplicator = nil
				Expect(factory.Get(dataSet)).To(BeNil())
			})

			It("returns successfully without a deduplicator  when the data set does not have a deduplicator name", func() {
				dataSet.Deduplicator.Name = nil
				Expect(factory.Get(dataSet)).To(BeNil())
			})
		})
	})
})
