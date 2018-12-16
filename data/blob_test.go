package data_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTest "github.com/tidepool-org/platform/data/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Blob", func() {
	Context("ParseBlob", func() {
		// TODO
	})

	Context("NewBlob", func() {
		It("is successful", func() {
			Expect(data.NewBlob()).To(Equal(&data.Blob{}))
		})
	})

	Context("Blob", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(datum *data.Blob, expectedErrors ...error) {
					validator := structureValidator.New()
					Expect(validator).ToNot(BeNil())
					datum.Validate(validator)
					errorsTest.ExpectEqual(validator.Error(), expectedErrors...)
				},
				Entry("succeeds",
					dataTest.NewBlob(),
				),
			)
		})

		Context("Normalize", func() {
			It("does not modified the datum", func() {
				datum := dataTest.NewBlob()
				expectedDatum := dataTest.CloneBlob(datum)
				normalizer := dataNormalizer.New()
				Expect(normalizer).ToNot(BeNil())
				datum.Normalize(normalizer)
				Expect(normalizer.Error()).To(BeNil())
				Expect(normalizer.Data()).To(BeEmpty())
				Expect(datum).To(Equal(expectedDatum))
			})
		})

		Context("with new blob with data", func() {
			var key string
			var value string
			var datum *data.Blob

			BeforeEach(func() {
				key = test.NewVariableString(1, 8, test.CharsetAlphaNumeric)
				value = test.NewText(0, 32)
				datum = dataTest.NewBlob()
				(*datum)[key] = value
			})

			Context("Get", func() {
				It("returns nil if value does not exist for key", func() {
					delete(*datum, key)
					Expect(datum.Get(key)).To(BeNil())
				})

				It("returns value if it exists for key", func() {
					Expect(datum.Get(key)).To(Equal(value))
				})
			})

			Context("Set", func() {
				It("sets nil value for the key", func() {
					datum.Set(key, nil)
					Expect((*datum)[key]).To(BeNil())
				})

				It("sets empty value for the key", func() {
					datum.Set(key, "")
					Expect((*datum)[key]).To(BeEmpty())
				})

				It("sets new value for the key", func() {
					newValue := test.NewText(0, 32)
					datum.Set(key, newValue)
					Expect((*datum)[key]).To(Equal(newValue))
				})
			})

			Context("Delete", func() {
				It("deletes value for the key", func() {
					datum.Delete(key)
					_, exists := (*datum)[key]
					Expect(exists).To(BeFalse())
				})
			})
		})
	})

	Context("ParseBlobArray", func() {
		// TODO
	})

	Context("NewBlobArray", func() {
		It("is successful", func() {
			Expect(data.NewBlobArray()).To(Equal(&data.BlobArray{}))
		})
	})

	Context("BlobArray", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(datum *data.BlobArray, expectedErrors ...error) {
					validator := structureValidator.New()
					Expect(validator).ToNot(BeNil())
					datum.Validate(validator)
					errorsTest.ExpectEqual(validator.Error(), expectedErrors...)
				},
				Entry("succeeds",
					dataTest.NewBlobArray(),
				),
			)
		})

		Context("Normalize", func() {
			It("does not modified the datum", func() {
				datum := dataTest.NewBlobArray()
				expectedDatum := dataTest.CloneBlobArray(datum)
				normalizer := dataNormalizer.New()
				Expect(normalizer).ToNot(BeNil())
				datum.Normalize(normalizer)
				Expect(normalizer.Error()).To(BeNil())
				Expect(normalizer.Data()).To(BeEmpty())
				Expect(datum).To(Equal(expectedDatum))
			})
		})
	})
})
