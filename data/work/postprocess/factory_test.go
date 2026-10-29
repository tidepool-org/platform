package postprocess_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	clinicsTest "github.com/tidepool-org/platform/clinics/test"
	configTest "github.com/tidepool-org/platform/config/test"
	dataWorkPostprocess "github.com/tidepool-org/platform/data/work/postprocess"
	dataWorkPostprocessTest "github.com/tidepool-org/platform/data/work/postprocess/test"
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("Factory", func() {
	It("QuantityDefault is expected", func() {
		Expect(dataWorkPostprocess.QuantityDefault).To(Equal(60))
	})

	It("Frequency is expected", func() {
		Expect(dataWorkPostprocess.Frequency).To(Equal(5 * time.Second))
	})

	Context("Config", func() {
		var cfg *dataWorkPostprocess.Config

		BeforeEach(func() {
			cfg = dataWorkPostprocess.NewConfig()
		})

		It("NewConfig returns the defaults", func() {
			Expect(cfg).ToNot(BeNil())
			Expect(cfg.Quantity).To(Equal(dataWorkPostprocess.QuantityDefault))
		})

		Context("Load", func() {
			var configReporter *configTest.Reporter

			BeforeEach(func() {
				configReporter = configTest.NewReporter()
			})

			It("returns an error when the config reporter is missing", func() {
				Expect(cfg.Load(nil)).To(MatchError("config reporter is missing"))
			})

			It("returns an error when quantity is not parsable", func() {
				configReporter.Config["quantity"] = test.RandomStringFromCharset(test.CharsetAlpha)
				Expect(cfg.Load(configReporter)).To(MatchError("quantity is invalid"))
			})

			It("returns an error when quantity is not positive", func() {
				configReporter.Config["quantity"] = "0"
				Expect(cfg.Load(configReporter)).To(MatchError("quantity is invalid"))
			})

			It("keeps the default when quantity is not set", func() {
				Expect(cfg.Load(configReporter)).To(Succeed())
				Expect(cfg.Quantity).To(Equal(dataWorkPostprocess.QuantityDefault))
			})

			It("loads quantity", func() {
				configReporter.Config["quantity"] = "120"
				Expect(cfg.Load(configReporter)).To(Succeed())
				Expect(cfg.Quantity).To(Equal(120))
			})
		})

		Context("Validate", func() {
			It("succeeds with the defaults", func() {
				Expect(cfg.Validate()).To(Succeed())
			})

			It("returns an error when quantity is not positive", func() {
				cfg.Quantity = 0
				Expect(cfg.Validate()).To(MatchError("quantity is invalid"))
			})
		})
	})

	Context("with dependencies", func() {
		var controller *gomock.Controller
		var cfg *dataWorkPostprocess.Config
		var dependencies dataWorkPostprocess.Dependencies

		BeforeEach(func() {
			controller = gomock.NewController(GinkgoT())
			cfg = dataWorkPostprocess.NewConfig()
			cfg.Quantity = test.RandomIntFromRange(1, 100)
			dependencies = dataWorkPostprocess.Dependencies{
				Dependencies:  workBase.Dependencies{WorkClient: workTest.NewMockClient(controller)},
				Summarizers:   dataWorkPostprocessTest.NewMockSummarizers(controller),
				ClinicsClient: clinicsTest.NewMockClient(controller),
				UserClient:    userTest.NewMockClient(controller),
			}
		})

		AfterEach(func() {
			controller.Finish()
		})

		DescribeTable("Validate",
			func(mutator func(dependencies *dataWorkPostprocess.Dependencies), expected string) {
				mutator(&dependencies)
				if expected == "" {
					Expect(dependencies.Validate()).To(Succeed())
				} else {
					Expect(dependencies.Validate()).To(MatchError(expected))
				}
			},
			Entry("succeeds", func(dependencies *dataWorkPostprocess.Dependencies) {}, ""),
			Entry("reports the work client is missing", func(dependencies *dataWorkPostprocess.Dependencies) {
				dependencies.WorkClient = nil
			}, "work client is missing"),
			Entry("reports the summarizers are missing", func(dependencies *dataWorkPostprocess.Dependencies) {
				dependencies.Summarizers = nil
			}, "summarizers is missing"),
			Entry("reports the clinics client is missing", func(dependencies *dataWorkPostprocess.Dependencies) {
				dependencies.ClinicsClient = nil
			}, "clinics client is missing"),
			Entry("reports the user client is missing", func(dependencies *dataWorkPostprocess.Dependencies) {
				dependencies.UserClient = nil
			}, "user client is missing"),
		)

		It("NewProcessorFactory reports the type, configured quantity and frequency of the work", func() {
			processorFactory, err := dataWorkPostprocess.NewProcessorFactory(cfg, dependencies)
			Expect(err).ToNot(HaveOccurred())
			Expect(processorFactory.Type()).To(Equal(dataWorkPostprocess.Type))
			Expect(processorFactory.Quantity()).To(Equal(cfg.Quantity))
			Expect(processorFactory.Frequency()).To(Equal(dataWorkPostprocess.Frequency))

			processor, err := processorFactory.New()
			Expect(err).ToNot(HaveOccurred())
			Expect(processor).ToNot(BeNil())
		})

		It("NewProcessorFactory returns an error when the config is missing", func() {
			processorFactory, err := dataWorkPostprocess.NewProcessorFactory(nil, dependencies)
			Expect(err).To(MatchError("config is missing"))
			Expect(processorFactory).To(BeNil())
		})

		It("NewProcessorFactory returns an error when the config is invalid", func() {
			cfg.Quantity = 0
			processorFactory, err := dataWorkPostprocess.NewProcessorFactory(cfg, dependencies)
			Expect(err).To(MatchError("config is invalid; quantity is invalid"))
			Expect(processorFactory).To(BeNil())
		})

		It("NewProcessorFactory returns an error when the dependencies are invalid", func() {
			dependencies.ClinicsClient = nil
			processorFactory, err := dataWorkPostprocess.NewProcessorFactory(cfg, dependencies)
			Expect(err).To(MatchError(ContainSubstring("clinics client is missing")))
			Expect(processorFactory).To(BeNil())
		})
	})
})
