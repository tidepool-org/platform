package normalizer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/normalizer"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Standard", func() {
	It("NewStandard returns an error if context is nil", func() {
		standard, err := normalizer.NewStandard(nil)
		Expect(err).To(MatchError("context is missing"))
		Expect(standard).To(BeNil())
	})

	Context("new normalizer", func() {
		var standardContext *context.Standard
		var standard *normalizer.Standard

		BeforeEach(func() {
			var err error
			standardContext, err = context.NewStandard(null.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(standardContext).ToNot(BeNil())
			standard, err = normalizer.NewStandard(standardContext)
			Expect(err).ToNot(HaveOccurred())
		})

		It("exists", func() {
			Expect(standard).ToNot(BeNil())
		})

		It("has a contained Data that is empty", func() {
			Expect(standard.Data()).To(BeEmpty())
		})

		It("ignores sending a nil datum to AppendDatum", func() {
			standard.AppendDatum(nil)
			Expect(standard.Data()).To(BeEmpty())
		})

		It("Logger returns a logger", func() {
			Expect(standard.Logger()).ToNot(BeNil())
		})

		Context("SetMeta", func() {
			It("sets the meta on the context", func() {
				meta := "metametameta"
				standard.SetMeta(meta)
				Expect(standardContext.Meta()).To(BeIdenticalTo(meta))
			})
		})

		Context("AppendError", func() {
			It("appends an error on the context", func() {
				standard.AppendError("append-error", &service.Error{})
				Expect(standardContext.Errors()).To(HaveLen(1))
			})
		})

		Context("AppendDatum with a first datum", func() {
			var firstDatum *testData.Datum

			BeforeEach(func() {
				firstDatum = testData.NewDatum()
				standard.AppendDatum(firstDatum)
			})

			It("has data", func() {
				Expect(standard.Data()).ToNot(BeEmpty())
			})

			It("has the datum", func() {
				Expect(standard.Data()).To(ConsistOf(firstDatum))
			})

			Context("and AppendDatum with a second data", func() {
				var secondDatum *testData.Datum

				BeforeEach(func() {
					secondDatum = testData.NewDatum()
					standard.AppendDatum(secondDatum)
				})

				It("has data", func() {
					Expect(standard.Data()).ToNot(BeEmpty())
				})

				It("has both data", func() {
					Expect(standard.Data()).To(ConsistOf(firstDatum, secondDatum))
				})
			})
		})

		Context("NewChildNormalizer", func() {
			var child data.Normalizer

			BeforeEach(func() {
				child = standard.NewChildNormalizer("child")
			})

			It("exists", func() {
				Expect(child).ToNot(BeNil())
			})

			It("Logger returns a logger", func() {
				Expect(child.Logger()).ToNot(BeNil())
			})

			Context("AppendDatum with a first error", func() {
				var firstDatum *testData.Datum

				BeforeEach(func() {
					firstDatum = testData.NewDatum()
					child.AppendDatum(firstDatum)
				})

				It("has data", func() {
					Expect(standard.Data()).ToNot(BeEmpty())
				})

				It("has the data", func() {
					Expect(standard.Data()).To(ConsistOf(firstDatum))
				})

				Context("and AppendDatum with a second error to the parent context", func() {
					var secondDatum *testData.Datum

					BeforeEach(func() {
						secondDatum = testData.NewDatum()
						standard.AppendDatum(secondDatum)
					})

					It("has data", func() {
						Expect(standard.Data()).ToNot(BeEmpty())
					})

					It("has both data", func() {
						Expect(standard.Data()).To(ConsistOf(firstDatum, secondDatum))
					})
				})
			})
		})
	})
})
