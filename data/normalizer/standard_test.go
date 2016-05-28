package normalizer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/service"
)

type TestDatum struct{}

func (t *TestDatum) Meta() interface{}                          { return nil }
func (t *TestDatum) Parse(parser data.ObjectParser) error       { return nil }
func (t *TestDatum) Validate(validator data.Validator) error    { return nil }
func (t *TestDatum) Normalize(normalizer data.Normalizer) error { return nil }
func (t *TestDatum) SetUserID(userID string)                    {}
func (t *TestDatum) SetGroupID(groupID string)                  {}
func (t *TestDatum) SetDatasetID(datasetID string)              {}

var _ = Describe("Standard", func() {
	It("New returns an error if context is nil", func() {
		standard, err := normalizer.NewStandard(nil)
		Expect(standard).To(BeNil())
		Expect(err).To(HaveOccurred())
	})

	Context("new normalizer", func() {
		var standardContext *context.Standard
		var standard *normalizer.Standard

		BeforeEach(func() {
			var err error
			standardContext = context.NewStandard()
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
			var firstDatum *TestDatum

			BeforeEach(func() {
				firstDatum = &TestDatum{}
				standard.AppendDatum(firstDatum)
			})

			It("has data", func() {
				Expect(standard.Data()).ToNot(BeEmpty())
			})

			It("has the datum", func() {
				Expect(standard.Data()).To(ConsistOf(firstDatum))
			})

			Context("and AppendDatum with a second data", func() {
				var secondDatum *TestDatum

				BeforeEach(func() {
					secondDatum = &TestDatum{}
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

		Context("NormalizeBloodGlucose", func() {
			var bloodGlucoseNormalizer data.BloodGlucoseNormalizer

			BeforeEach(func() {
				units := "mg/dl"
				bloodGlucoseNormalizer = standard.NormalizeBloodGlucose(&units)
			})

			It("exists", func() {
				Expect(bloodGlucoseNormalizer).ToNot(BeNil())
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

			Context("AppendDatum with a first error", func() {
				var firstDatum *TestDatum

				BeforeEach(func() {
					firstDatum = &TestDatum{}
					child.AppendDatum(firstDatum)
				})

				It("has data", func() {
					Expect(standard.Data()).ToNot(BeEmpty())
				})

				It("has the data", func() {
					Expect(standard.Data()).To(ConsistOf(firstDatum))
				})

				Context("and AppendDatum with a second error to the parent context", func() {
					var secondDatum *TestDatum

					BeforeEach(func() {
						secondDatum = &TestDatum{}
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
