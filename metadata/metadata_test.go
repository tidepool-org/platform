package metadata_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/metadata"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Metadata", func() {
	It("MetadataArrayLengthMaximum is expected", func() {
		Expect(metadata.MetadataArrayLengthMaximum).To(Equal(100))
	})

	It("MetadataSizeMaximum is expected", func() {
		Expect(metadata.MetadataSizeMaximum).To(Equal(6 * 1024))
	})

	Context("Metadata", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *metadata.Metadata)) {
				datum := metadataTest.RandomMetadata()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, metadataTest.NewObjectFromMetadata(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, metadataTest.NewObjectFromMetadata(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *metadata.Metadata) {},
			),
			Entry("empty",
				func(datum *metadata.Metadata) {
					*datum = *metadata.NewMetadata()
				},
			),
		)

		Context("ParseMetadata", func() {
			It("returns nil when the object is missing", func() {
				Expect(metadata.ParseMetadata(structureParser.NewObject(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := metadataTest.RandomMetadata()
				object := metadataTest.NewObjectFromMetadata(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(&object)
				Expect(metadata.ParseMetadata(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewMetadata", func() {
			It("returns successfully with default values", func() {
				Expect(metadata.NewMetadata()).To(Equal(&metadata.Metadata{}))
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *metadata.Metadata), expectedErrors ...error) {
					expectedDatum := metadataTest.RandomMetadata()
					object := metadataTest.NewObjectFromMetadata(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &metadata.Metadata{}
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *metadata.Metadata) {},
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *metadata.Metadata), expectedErrors ...error) {
					datum := metadataTest.RandomMetadata()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *metadata.Metadata) {},
				),
				Entry("empty",
					func(datum *metadata.Metadata) { *datum = *metadata.NewMetadata() },
					structureValidator.ErrorValueEmpty(),
				),
				Entry("single valid",
					func(datum *metadata.Metadata) {
						*datum = *metadata.NewMetadata()
						(*datum)[metadataTest.RandomMetadataKey()] = metadataTest.RandomMetadataValue()
					},
				),
				Entry("multiple valid",
					func(datum *metadata.Metadata) {
						*datum = *metadataTest.RandomMetadata()
					},
				),
				Entry("not serializable",
					func(datum *metadata.Metadata) {
						(*datum)[metadataTest.RandomMetadataKey()] = func() {}
					},
					structureValidator.ErrorValueNotSerializable(),
				),
				Entry("size in range (upper)",
					func(datum *metadata.Metadata) {
						*datum = *metadata.NewMetadata()
						(*datum)["size"] = test.RandomStringFromRangeAndCharset(4085, 4085, test.CharsetAlphaNumeric)
					},
				),
				Entry("size out of range (upper)",
					func(datum *metadata.Metadata) {
						*datum = *metadata.NewMetadata()
						(*datum)["size"] = test.RandomStringFromRangeAndCharset(6134, 6134, test.CharsetAlphaNumeric)
					},
					structureValidator.ErrorSizeNotLessThanOrEqualTo(6145, 6144),
				),
				Entry("multiple errors",
					func(datum *metadata.Metadata) {
						*datum = *metadata.NewMetadata()
					},
					structureValidator.ErrorValueEmpty(),
				),
			)
		})

		Context("with new metadata with data", func() {
			var key string
			var value interface{}
			var datum *metadata.Metadata

			BeforeEach(func() {
				key = metadataTest.RandomMetadataKey()
				value = metadataTest.RandomMetadataValue()
				datum = metadataTest.RandomMetadata()
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
					newValue := metadataTest.RandomMetadataValue()
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

	Context("MetadataArray", func() {
		Context("ParseMetadataArray", func() {
			It("returns nil when the object is missing", func() {
				Expect(metadata.ParseMetadataArray(structureParser.NewArray(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := metadataTest.RandomMetadataArray()
				array := metadataTest.NewArrayFromMetadataArray(datum, test.ObjectFormatJSON)
				parser := structureParser.NewArray(&array)
				Expect(metadata.ParseMetadataArray(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewMetadataArray", func() {
			It("returns successfully with default values", func() {
				Expect(metadata.NewMetadataArray()).To(Equal(&metadata.MetadataArray{}))
			})
		})

		Context("Parse", func() {
			It("successfully parses a nil array", func() {
				parser := structureParser.NewArray(nil)
				datum := metadata.NewMetadataArray()
				datum.Parse(parser)
				Expect(datum).To(Equal(metadata.NewMetadataArray()))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})

			It("successfully parses an empty array", func() {
				parser := structureParser.NewArray(&[]interface{}{})
				datum := metadata.NewMetadataArray()
				datum.Parse(parser)
				Expect(datum).To(Equal(metadata.NewMetadataArray()))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})

			It("successfully parses a non-empty array", func() {
				expectedDatum := metadataTest.RandomMetadataArray()
				array := metadataTest.NewArrayFromMetadataArray(expectedDatum, test.ObjectFormatJSON)
				parser := structureParser.NewArray(&array)
				datum := metadata.NewMetadataArray()
				datum.Parse(parser)
				Expect(datum).To(Equal(expectedDatum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *metadata.MetadataArray), expectedErrors ...error) {
					datum := metadataTest.RandomMetadataArray()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *metadata.MetadataArray) {},
				),
				Entry("empty",
					func(datum *metadata.MetadataArray) { *datum = *metadata.NewMetadataArray() },
					structureValidator.ErrorValueEmpty(),
				),
				Entry("nil",
					func(datum *metadata.MetadataArray) {
						*datum = append(*metadata.NewMetadataArray(), nil)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
				),
				Entry("single valid",
					func(datum *metadata.MetadataArray) {
						*datum = append(*metadata.NewMetadataArray(), metadataTest.RandomMetadata())
					},
				),
				Entry("multiple valid",
					func(datum *metadata.MetadataArray) {
						*datum = *metadataTest.RandomMetadataArray()
					},
				),
				Entry("multiple in range (upper)",
					func(datum *metadata.MetadataArray) {
						*datum = *metadata.NewMetadataArray()
						for count := 100; count > 0; count-- {
							*datum = append(*datum, metadataTest.RandomMetadata())
						}
					},
				),
				Entry("multiple out of range range (upper)",
					func(datum *metadata.MetadataArray) {
						*datum = *metadata.NewMetadataArray()
						for count := 101; count > 0; count-- {
							*datum = append(*datum, metadataTest.RandomMetadata())
						}
					},
					structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100),
				),
				Entry("multiple errors",
					func(datum *metadata.MetadataArray) {
						*datum = *metadata.NewMetadataArray()
					},
					structureValidator.ErrorValueEmpty(),
				),
			)
		})
	})
})
