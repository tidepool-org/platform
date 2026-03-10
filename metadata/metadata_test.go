package metadata_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/metadata"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Metadata", func() {
	It("MetadataArrayLengthMaximum is expected", func() {
		Expect(metadata.MetadataArrayLengthMaximum).To(Equal(100))
	})

	It("MetadataSizeMaximum is expected", func() {
		Expect(metadata.MetadataSizeMaximum).To(Equal(4 * 1024))
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
				Expect(metadata.ParseMetadata(structureParser.NewObject(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := metadataTest.RandomMetadata()
				object := metadataTest.NewObjectFromMetadata(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(logTest.NewLogger(), &object)
				Expect(metadata.ParseMetadata(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewMetadata", func() {
			It("returns successfully with default values", func() {
				Expect(metadata.NewMetadata()).To(Equal(&metadata.Metadata{}))
			})
		})

		Context("MetadataFromMap", func() {
			It("returns nil when map is nil", func() {
				Expect(metadata.MetadataFromMap(nil)).To(BeNil())
			})

			It("returns successfully with default values", func() {
				datum := metadataTest.RandomMetadataMap()
				Expect(metadata.MetadataFromMap(datum)).To(PointTo(Equal(metadata.Metadata(datum))))
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *metadata.Metadata), expectedErrors ...error) {
					expectedDatum := metadataTest.RandomMetadata()
					object := metadataTest.NewObjectFromMetadata(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &metadata.Metadata{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
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
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
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
						(*datum)["size"] = test.RandomStringFromRangeAndCharset(4086, 4086, test.CharsetAlphaNumeric)
					},
					structureValidator.ErrorSizeNotLessThanOrEqualTo(4097, 4096),
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
				Expect(metadata.ParseMetadataArray(structureParser.NewArray(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := metadataTest.RandomMetadataArray()
				array := metadataTest.NewArrayFromMetadataArray(datum, test.ObjectFormatJSON)
				parser := structureParser.NewArray(logTest.NewLogger(), &array)
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
				parser := structureParser.NewArray(logTest.NewLogger(), nil)
				datum := metadata.NewMetadataArray()
				datum.Parse(parser)
				Expect(datum).To(Equal(metadata.NewMetadataArray()))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})

			It("successfully parses an empty array", func() {
				parser := structureParser.NewArray(logTest.NewLogger(), &[]interface{}{})
				datum := metadata.NewMetadataArray()
				datum.Parse(parser)
				Expect(datum).To(Equal(metadata.NewMetadataArray()))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})

			It("successfully parses a non-empty array", func() {
				expectedDatum := metadataTest.RandomMetadataArray()
				array := metadataTest.NewArrayFromMetadataArray(expectedDatum, test.ObjectFormatJSON)
				parser := structureParser.NewArray(logTest.NewLogger(), &array)
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
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
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

	Context("Decode", func() {
		var ctx context.Context

		BeforeEach(func() {
			ctx = context.Background()
		})

		Context("without decode options", func() {
			It("returns nil object if metadata is nil", func() {
				decodedObject, err := metadata.Decode[object](ctx, nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(decodedObject).To(BeNil())
			})

			It("returns copy of metadata if object is a pointer to metadata", func() {
				decodableMetadata := metadataTest.RandomMetadataMap()
				decodedObject, err := metadata.Decode[map[string]any](ctx, decodableMetadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(decodedObject).To(PointTo(Equal(decodableMetadata)))
			})

			It("returns an error if the metadata cannot be decoded to the object with all fields parsed", func() {
				decodableMetadata := metadataTest.RandomMetadataMap()
				decodedObject, err := metadata.Decode[object](ctx, decodableMetadata)
				Expect(err).To(MatchError(ContainSubstring("unable to decode metadata")))
				Expect(decodedObject).To(BeNil())
			})

			It("returns successfully with decoded object", func() {
				expectedObject := randomObject()
				decodableMetadata := objectToMetadata(expectedObject)
				decodedObject, err := metadata.Decode[object](ctx, decodableMetadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(decodedObject).To(Equal(expectedObject))
			})
		})

		Context("with decode option IgnoreNotParsed", func() {
			It("returns nil object if metadata is nil", func() {
				decodedObject, err := metadata.Decode[object](ctx, nil, request.IgnoreNotParsed())
				Expect(err).ToNot(HaveOccurred())
				Expect(decodedObject).To(BeNil())
			})

			It("returns copy of metadata if object is a pointer to metadata", func() {
				decodableMetadata := metadataTest.RandomMetadataMap()
				decodedObject, err := metadata.Decode[map[string]any](ctx, decodableMetadata, request.IgnoreNotParsed())
				Expect(err).ToNot(HaveOccurred())
				Expect(decodedObject).To(PointTo(Equal(decodableMetadata)))
			})

			It("returns successfully if the metadata can be decoded to the object without all fields parsed", func() {
				decodableMetadata := metadataTest.RandomMetadataMap()
				decodedObject, err := metadata.Decode[object](ctx, decodableMetadata, request.IgnoreNotParsed())
				Expect(err).ToNot(HaveOccurred())
				Expect(decodedObject).To(Equal(&object{}))
			})

			It("returns successfully with decoded object", func() {
				expectedObject := randomObject()
				decodableMetadata := objectToMetadata(expectedObject)
				decodedObject, err := metadata.Decode[object](ctx, decodableMetadata, request.IgnoreNotParsed())
				Expect(err).ToNot(HaveOccurred())
				Expect(decodedObject).To(Equal(expectedObject))
			})
		})
	})

	Context("Encode", func() {
		It("returns nil metadata if object is nil", func() {
			encodedMetadata, err := metadata.Encode[object](nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(encodedMetadata).To(BeNil())
		})

		It("returns copy of object if object is a pointer to metadata", func() {
			expectedMetadata := metadataTest.RandomMetadataMap()
			encodableObject := pointer.From(expectedMetadata)
			encodedMetadata, err := metadata.Encode(encodableObject)
			Expect(err).ToNot(HaveOccurred())
			Expect(encodedMetadata).To(Equal(expectedMetadata))
		})

		It("returns an error if the object cannot be encoded", func() {
			encodableObject := &object{Zulu: func() {}}
			encodedMetadata, err := metadata.Encode(encodableObject)
			Expect(err).To(MatchError(ContainSubstring("unable to encode object")))
			Expect(encodedMetadata).To(BeNil())
		})

		It("returns an error if the object cannot be decoded", func() {
			encodableObject := []string{"invalid"}
			encodedMetadata, err := metadata.Encode(&encodableObject)
			Expect(err).To(MatchError(ContainSubstring("unable to decode metadata")))
			Expect(encodedMetadata).To(BeNil())
		})

		It("returns successfully with encoded metadata", func() {
			encodableObject := randomObject()
			expectedMetadata := objectToMetadata(encodableObject)
			encodedMetadata, err := metadata.Encode(encodableObject)
			Expect(err).ToNot(HaveOccurred())
			Expect(encodedMetadata).To(Equal(expectedMetadata))
		})
	})
})

type object struct {
	Alpha *string `json:"alpha,omitempty" bson:"alpha,omitempty"`
	Bravo *bravo  `json:"bravo,omitempty" bson:"bravo,omitempty"`
	Zulu  any     `json:"any,omitempty" bson:"any,omitempty"` // Used to test encoding errors
}

func (o *object) Parse(parser structure.ObjectParser) {
	o.Alpha = parser.String("alpha")
	if bravoParser := parser.WithReferenceObjectParser("bravo"); bravoParser.Exists() {
		o.Bravo = &bravo{}
		o.Bravo.Parse(bravoParser)
	}
}

type bravo struct {
	Charlie *string `json:"charlie,omitempty" bson:"charlie,omitempty"`
}

func (b *bravo) Parse(parser structure.ObjectParser) {
	b.Charlie = parser.String("charlie")
}

func randomObject() *object {
	return &object{
		Alpha: test.RandomOptional(test.RandomString, test.AllowOptional()),
		Bravo: test.RandomOptionalPointer(randomBravo, test.AllowOptional()),
	}
}

func objectToMetadata(object *object) map[string]any {
	if object == nil {
		return nil
	}
	metadata := map[string]any{}
	if object.Alpha != nil {
		metadata["alpha"] = *object.Alpha
	}
	if object.Bravo != nil {
		metadata["bravo"] = bravoToMetadata(object.Bravo)
	}
	return metadata
}

func randomBravo() *bravo {
	return &bravo{
		Charlie: test.RandomOptional(test.RandomString, test.AllowOptional()),
	}
}

func bravoToMetadata(bravo *bravo) map[string]any {
	if bravo == nil {
		return nil
	}
	metadata := map[string]any{}
	if bravo.Charlie != nil {
		metadata["charlie"] = *bravo.Charlie
	}
	return metadata
}
