package association_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/association"
	associationTest "github.com/tidepool-org/platform/association/test"
	"github.com/tidepool-org/platform/data"
	dataTest "github.com/tidepool-org/platform/data/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/net"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Association", func() {
	It("AssociationArrayLengthMaximum is expected", func() {
		Expect(association.AssociationArrayLengthMaximum).To(Equal(100))
	})

	It("ReasonLengthMaximum is expected", func() {
		Expect(association.ReasonLengthMaximum).To(Equal(1000))
	})

	It("TypeBlob is expected", func() {
		Expect(association.TypeBlob).To(Equal("blob"))
	})

	It("TypeDatum is expected", func() {
		Expect(association.TypeDatum).To(Equal("datum"))
	})

	It("TypeImage is expected", func() {
		Expect(association.TypeImage).To(Equal("image"))
	})

	It("TypeURL is expected", func() {
		Expect(association.TypeURL).To(Equal("url"))
	})

	It("Types returns expected", func() {
		Expect(association.Types()).To(Equal([]string{"blob", "datum", "image", "url"}))
	})

	Context("Association", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *association.Association)) {
				datum := associationTest.RandomAssociation()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, associationTest.NewObjectFromAssociation(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, associationTest.NewObjectFromAssociation(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *association.Association) {},
			),
			Entry("empty",
				func(datum *association.Association) {
					*datum = *association.NewAssociation()
				},
			),
			Entry("all",
				func(datum *association.Association) {
					datum.ID = pointer.FromString(dataTest.RandomDatumID())
					datum.Reason = pointer.FromString(associationTest.RandomReason())
					datum.Type = pointer.FromString(associationTest.RandomType())
					datum.URL = pointer.FromString(testHttp.NewURLString())
				},
			),
		)

		Context("ParseAssociation", func() {
			It("returns nil when the object is missing", func() {
				Expect(association.ParseAssociation(structureParser.NewObject(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := associationTest.RandomAssociation()
				object := associationTest.NewObjectFromAssociation(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(logTest.NewLogger(), &object)
				Expect(association.ParseAssociation(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewAssociation", func() {
			It("returns successfully with default values", func() {
				Expect(association.NewAssociation()).To(Equal(&association.Association{}))
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *association.Association), expectedErrors ...error) {
					expectedDatum := associationTest.RandomAssociation()
					object := associationTest.NewObjectFromAssociation(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &association.Association{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *association.Association) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *association.Association) {
						object["id"] = true
						object["reason"] = true
						object["type"] = true
						object["url"] = true
						expectedDatum.ID = nil
						expectedDatum.Reason = nil
						expectedDatum.Type = nil
						expectedDatum.URL = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/id"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/reason"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/type"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/url"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *association.Association), expectedErrors ...error) {
					datum := associationTest.RandomAssociation()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *association.Association) {},
				),
				Entry("type missing; id missing",
					func(datum *association.Association) {
						datum.ID = nil
						datum.Type = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/type"),
				),
				Entry("type missing; id empty",
					func(datum *association.Association) {
						datum.ID = pointer.FromString("")
						datum.Type = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/type"),
				),
				Entry("type missing; id invalid",
					func(datum *association.Association) {
						datum.ID = pointer.FromString("invalid")
						datum.Type = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/type"),
				),
				Entry("type missing; id valid",
					func(datum *association.Association) {
						datum.ID = pointer.FromString(dataTest.RandomDatumID())
						datum.Type = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/type"),
				),
				Entry("type blob; id missing",
					func(datum *association.Association) {
						datum.ID = nil
						datum.Type = pointer.FromString("blob")
						datum.URL = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/id"),
				),
				Entry("type blob; id empty",
					func(datum *association.Association) {
						datum.ID = pointer.FromString("")
						datum.Type = pointer.FromString("blob")
						datum.URL = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/id"),
				),
				Entry("type blob; id invalid",
					func(datum *association.Association) {
						datum.ID = pointer.FromString("invalid")
						datum.Type = pointer.FromString("blob")
						datum.URL = nil
					},
					errorsTest.WithPointerSource(data.ErrorValueStringAsIDNotValid("invalid"), "/id"),
				),
				Entry("type blob; id valid",
					func(datum *association.Association) {
						datum.ID = pointer.FromString(dataTest.RandomDatumID())
						datum.Type = pointer.FromString("blob")
						datum.URL = nil
					},
				),
				Entry("type datum; id missing",
					func(datum *association.Association) {
						datum.ID = nil
						datum.Type = pointer.FromString("datum")
						datum.URL = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/id"),
				),
				Entry("type datum; id empty",
					func(datum *association.Association) {
						datum.ID = pointer.FromString("")
						datum.Type = pointer.FromString("datum")
						datum.URL = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/id"),
				),
				Entry("type datum; id invalid",
					func(datum *association.Association) {
						datum.ID = pointer.FromString("invalid")
						datum.Type = pointer.FromString("datum")
						datum.URL = nil
					},
					errorsTest.WithPointerSource(data.ErrorValueStringAsIDNotValid("invalid"), "/id"),
				),
				Entry("type datum; id valid",
					func(datum *association.Association) {
						datum.ID = pointer.FromString(dataTest.RandomDatumID())
						datum.Type = pointer.FromString("datum")
						datum.URL = nil
					},
				),
				Entry("type image; id missing",
					func(datum *association.Association) {
						datum.ID = nil
						datum.Type = pointer.FromString("image")
						datum.URL = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/id"),
				),
				Entry("type image; id empty",
					func(datum *association.Association) {
						datum.ID = pointer.FromString("")
						datum.Type = pointer.FromString("image")
						datum.URL = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/id"),
				),
				Entry("type image; id invalid",
					func(datum *association.Association) {
						datum.ID = pointer.FromString("invalid")
						datum.Type = pointer.FromString("image")
						datum.URL = nil
					},
					errorsTest.WithPointerSource(data.ErrorValueStringAsIDNotValid("invalid"), "/id"),
				),
				Entry("type image; id valid",
					func(datum *association.Association) {
						datum.ID = pointer.FromString(dataTest.RandomDatumID())
						datum.Type = pointer.FromString("image")
						datum.URL = nil
					},
				),
				Entry("type url; id missing",
					func(datum *association.Association) {
						datum.ID = nil
						datum.Type = pointer.FromString("url")
						datum.URL = pointer.FromString(testHttp.NewURLString())
					},
				),
				Entry("type url; id empty",
					func(datum *association.Association) {
						datum.ID = pointer.FromString("")
						datum.Type = pointer.FromString("url")
						datum.URL = pointer.FromString(testHttp.NewURLString())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/id"),
				),
				Entry("type url; id invalid",
					func(datum *association.Association) {
						datum.ID = pointer.FromString("invalid")
						datum.Type = pointer.FromString("url")
						datum.URL = pointer.FromString(testHttp.NewURLString())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/id"),
				),
				Entry("type url; id valid",
					func(datum *association.Association) {
						datum.ID = pointer.FromString(dataTest.RandomDatumID())
						datum.Type = pointer.FromString("url")
						datum.URL = pointer.FromString(testHttp.NewURLString())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/id"),
				),
				Entry("reason missing",
					func(datum *association.Association) { datum.Reason = nil },
				),
				Entry("reason empty",
					func(datum *association.Association) { datum.Reason = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/reason"),
				),
				Entry("reason length; in range (upper)",
					func(datum *association.Association) {
						datum.Reason = pointer.FromString(test.RandomStringFromRange(1000, 1000))
					},
				),
				Entry("reason length; out of range (upper)",
					func(datum *association.Association) {
						datum.Reason = pointer.FromString(test.RandomStringFromRange(1001, 1001))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(1001, 1000), "/reason"),
				),
				Entry("type missing",
					func(datum *association.Association) { datum.Type = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/type"),
				),
				Entry("type invalid",
					func(datum *association.Association) { datum.Type = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"blob", "datum", "image", "url"}), "/type"),
				),
				Entry("type blob",
					func(datum *association.Association) {
						datum.ID = pointer.FromString(dataTest.RandomDatumID())
						datum.Type = pointer.FromString("blob")
						datum.URL = nil
					},
				),
				Entry("type datum",
					func(datum *association.Association) {
						datum.ID = pointer.FromString(dataTest.RandomDatumID())
						datum.Type = pointer.FromString("datum")
						datum.URL = nil
					},
				),
				Entry("type image",
					func(datum *association.Association) {
						datum.ID = pointer.FromString(dataTest.RandomDatumID())
						datum.Type = pointer.FromString("image")
						datum.URL = nil
					},
				),
				Entry("type url",
					func(datum *association.Association) {
						datum.ID = nil
						datum.Type = pointer.FromString("url")
						datum.URL = pointer.FromString(testHttp.NewURLString())
					},
				),
				Entry("type missing; url missing",
					func(datum *association.Association) {
						datum.Type = nil
						datum.URL = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/type"),
				),
				Entry("type missing; url empty",
					func(datum *association.Association) {
						datum.Type = nil
						datum.URL = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/type"),
				),
				Entry("type missing; url length in range (upper)",
					func(datum *association.Association) {
						datum.Type = nil
						datum.URL = pointer.FromString(strings.Repeat("x", 2047))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/type"),
				),
				Entry("type missing; url length out of range (upper)",
					func(datum *association.Association) {
						datum.Type = nil
						datum.URL = pointer.FromString(strings.Repeat("x", 2048))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/type"),
				),
				Entry("type blob; url missing",
					func(datum *association.Association) {
						datum.ID = pointer.FromString(dataTest.RandomDatumID())
						datum.Type = pointer.FromString("blob")
						datum.URL = nil
					},
				),
				Entry("type blob; url empty",
					func(datum *association.Association) {
						datum.ID = pointer.FromString(dataTest.RandomDatumID())
						datum.Type = pointer.FromString("blob")
						datum.URL = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/url"),
				),
				Entry("type blob; url length in range (upper)",
					func(datum *association.Association) {
						datum.ID = pointer.FromString(dataTest.RandomDatumID())
						datum.Type = pointer.FromString("blob")
						datum.URL = pointer.FromString(strings.Repeat("x", 2047))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/url"),
				),
				Entry("type blob; url length out of range (upper)",
					func(datum *association.Association) {
						datum.ID = pointer.FromString(dataTest.RandomDatumID())
						datum.Type = pointer.FromString("blob")
						datum.URL = pointer.FromString(strings.Repeat("x", 2048))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/url"),
				),
				Entry("type datum; url missing",
					func(datum *association.Association) {
						datum.ID = pointer.FromString(dataTest.RandomDatumID())
						datum.Type = pointer.FromString("datum")
						datum.URL = nil
					},
				),
				Entry("type datum; url empty",
					func(datum *association.Association) {
						datum.ID = pointer.FromString(dataTest.RandomDatumID())
						datum.Type = pointer.FromString("datum")
						datum.URL = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/url"),
				),
				Entry("type datum; url length in range (upper)",
					func(datum *association.Association) {
						datum.ID = pointer.FromString(dataTest.RandomDatumID())
						datum.Type = pointer.FromString("datum")
						datum.URL = pointer.FromString(strings.Repeat("x", 2047))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/url"),
				),
				Entry("type datum; url length out of range (upper)",
					func(datum *association.Association) {
						datum.ID = pointer.FromString(dataTest.RandomDatumID())
						datum.Type = pointer.FromString("datum")
						datum.URL = pointer.FromString(strings.Repeat("x", 2048))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/url"),
				),
				Entry("type image; url missing",
					func(datum *association.Association) {
						datum.ID = pointer.FromString(dataTest.RandomDatumID())
						datum.Type = pointer.FromString("image")
						datum.URL = nil
					},
				),
				Entry("type image; url empty",
					func(datum *association.Association) {
						datum.ID = pointer.FromString(dataTest.RandomDatumID())
						datum.Type = pointer.FromString("image")
						datum.URL = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/url"),
				),
				Entry("type image; url length in range (upper)",
					func(datum *association.Association) {
						datum.ID = pointer.FromString(dataTest.RandomDatumID())
						datum.Type = pointer.FromString("image")
						datum.URL = pointer.FromString(strings.Repeat("x", 2047))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/url"),
				),
				Entry("type image; url length out of range (upper)",
					func(datum *association.Association) {
						datum.ID = pointer.FromString(dataTest.RandomDatumID())
						datum.Type = pointer.FromString("image")
						datum.URL = pointer.FromString(strings.Repeat("x", 2048))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/url"),
				),
				Entry("type url; url missing",
					func(datum *association.Association) {
						datum.ID = nil
						datum.Type = pointer.FromString("url")
						datum.URL = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/url"),
				),
				Entry("type url; url empty",
					func(datum *association.Association) {
						datum.ID = nil
						datum.Type = pointer.FromString("url")
						datum.URL = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/url"),
				),
				Entry("type url; url invalid",
					func(datum *association.Association) {
						datum.ID = nil
						datum.Type = pointer.FromString("url")
						datum.URL = pointer.FromString("http:::")
					},
					errorsTest.WithPointerSource(net.ErrorValueStringAsURLNotValid("http:::"), "/url"),
				),
				Entry("type url; url valid",
					func(datum *association.Association) {
						datum.ID = nil
						datum.Type = pointer.FromString("url")
						datum.URL = pointer.FromString(testHttp.NewURLString())
					},
				),
				Entry("type url; url valid; length in range (upper)",
					func(datum *association.Association) {
						datum.ID = nil
						datum.Type = pointer.FromString("url")
						datum.URL = pointer.FromString("http://" + strings.Repeat("x", 2040))
					},
				),
				Entry("type url; url valid; length out of range (upper)",
					func(datum *association.Association) {
						datum.ID = nil
						datum.Type = pointer.FromString("url")
						datum.URL = pointer.FromString("http://" + strings.Repeat("x", 2041))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(2048, 2047), "/url"),
				),
				Entry("multiple errors",
					func(datum *association.Association) {
						datum.ID = pointer.FromString("")
						datum.Reason = pointer.FromString("")
						datum.Type = pointer.FromString(association.TypeDatum)
						datum.URL = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/id"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/reason"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/url"),
				),
			)
		})
	})

	Context("AssociationArray", func() {
		Context("ParseAssociationArray", func() {
			It("returns nil when the object is missing", func() {
				Expect(association.ParseAssociationArray(structureParser.NewArray(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := associationTest.RandomAssociationArray()
				array := associationTest.NewArrayFromAssociationArray(datum, test.ObjectFormatJSON)
				parser := structureParser.NewArray(logTest.NewLogger(), &array)
				Expect(association.ParseAssociationArray(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewAssociationArray", func() {
			It("returns successfully with default values", func() {
				Expect(association.NewAssociationArray()).To(Equal(&association.AssociationArray{}))
			})
		})

		Context("Parse", func() {
			It("successfully parses a nil array", func() {
				parser := structureParser.NewArray(logTest.NewLogger(), nil)
				datum := association.NewAssociationArray()
				datum.Parse(parser)
				Expect(datum).To(Equal(association.NewAssociationArray()))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})

			It("successfully parses an empty array", func() {
				parser := structureParser.NewArray(logTest.NewLogger(), &[]interface{}{})
				datum := association.NewAssociationArray()
				datum.Parse(parser)
				Expect(datum).To(Equal(association.NewAssociationArray()))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})

			It("successfully parses a non-empty array", func() {
				expectedDatum := associationTest.RandomAssociationArray()
				array := associationTest.NewArrayFromAssociationArray(expectedDatum, test.ObjectFormatJSON)
				parser := structureParser.NewArray(logTest.NewLogger(), &array)
				datum := association.NewAssociationArray()
				datum.Parse(parser)
				Expect(datum).To(Equal(expectedDatum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *association.AssociationArray), expectedErrors ...error) {
					datum := associationTest.RandomAssociationArray()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *association.AssociationArray) {},
				),
				Entry("empty",
					func(datum *association.AssociationArray) { *datum = *association.NewAssociationArray() },
					structureValidator.ErrorValueEmpty(),
				),
				Entry("nil",
					func(datum *association.AssociationArray) {
						*datum = append(*association.NewAssociationArray(), nil)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
				),
				Entry("single invalid",
					func(datum *association.AssociationArray) {
						invalid := associationTest.RandomAssociation()
						invalid.Type = nil
						*datum = append(*association.NewAssociationArray(), invalid)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0/type"),
				),
				Entry("single valid",
					func(datum *association.AssociationArray) {
						*datum = append(*association.NewAssociationArray(), associationTest.RandomAssociation())
					},
				),
				Entry("multiple invalid",
					func(datum *association.AssociationArray) {
						invalid := associationTest.RandomAssociation()
						invalid.Type = nil
						*datum = append(*association.NewAssociationArray(), associationTest.RandomAssociation(), invalid, associationTest.RandomAssociation())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/type"),
				),
				Entry("multiple valid",
					func(datum *association.AssociationArray) {
						*datum = *associationTest.RandomAssociationArray()
					},
				),
				Entry("multiple in range (upper)",
					func(datum *association.AssociationArray) {
						*datum = *association.NewAssociationArray()
						for count := 100; count > 0; count-- {
							*datum = append(*datum, associationTest.RandomAssociation())
						}
					},
				),
				Entry("multiple out of range range (upper)",
					func(datum *association.AssociationArray) {
						*datum = *association.NewAssociationArray()
						for count := 101; count > 0; count-- {
							*datum = append(*datum, associationTest.RandomAssociation())
						}
					},
					structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100),
				),
				Entry("multiple errors",
					func(datum *association.AssociationArray) {
						invalid := associationTest.RandomAssociation()
						invalid.Type = nil
						*datum = append(*association.NewAssociationArray(), nil, invalid, associationTest.RandomAssociation())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/type"),
				),
			)
		})
	})
})
