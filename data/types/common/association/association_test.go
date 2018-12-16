package association_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/common/association"
	dataTypesCommonAssociationTest "github.com/tidepool-org/platform/data/types/common/association/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/net"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
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

	It("TypeDatum is expected", func() {
		Expect(association.TypeDatum).To(Equal("datum"))
	})

	It("TypeURL is expected", func() {
		Expect(association.TypeURL).To(Equal("url"))
	})

	It("Types returns expected", func() {
		Expect(association.Types()).To(Equal([]string{"datum", "url"}))
	})

	Context("ParseAssociation", func() {
		// TODO
	})

	Context("NewAssociation", func() {
		It("is successful", func() {
			Expect(association.NewAssociation()).To(Equal(&association.Association{}))
		})
	})

	Context("Association", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *association.Association), expectedErrors ...error) {
					datum := dataTypesCommonAssociationTest.NewAssociation()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
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
						datum.ID = pointer.FromString(dataTest.RandomID())
						datum.Type = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/type"),
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
						datum.ID = pointer.FromString(dataTest.RandomID())
						datum.Type = pointer.FromString("datum")
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
						datum.ID = pointer.FromString(dataTest.RandomID())
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
					func(datum *association.Association) { datum.Reason = pointer.FromString(test.NewText(1000, 1000)) },
				),
				Entry("reason length; out of range (upper)",
					func(datum *association.Association) { datum.Reason = pointer.FromString(test.NewText(1001, 1001)) },
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(1001, 1000), "/reason"),
				),
				Entry("type missing",
					func(datum *association.Association) { datum.Type = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/type"),
				),
				Entry("type invalid",
					func(datum *association.Association) { datum.Type = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"datum", "url"}), "/type"),
				),
				Entry("type datum",
					func(datum *association.Association) {
						datum.ID = pointer.FromString(dataTest.RandomID())
						datum.Type = pointer.FromString("datum")
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
				Entry("type datum; url missing",
					func(datum *association.Association) {
						datum.ID = pointer.FromString(dataTest.RandomID())
						datum.Type = pointer.FromString("datum")
						datum.URL = nil
					},
				),
				Entry("type datum; url empty",
					func(datum *association.Association) {
						datum.ID = pointer.FromString(dataTest.RandomID())
						datum.Type = pointer.FromString("datum")
						datum.URL = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/url"),
				),
				Entry("type datum; url length in range (upper)",
					func(datum *association.Association) {
						datum.ID = pointer.FromString(dataTest.RandomID())
						datum.Type = pointer.FromString("datum")
						datum.URL = pointer.FromString(strings.Repeat("x", 2047))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/url"),
				),
				Entry("type datum; url length out of range (upper)",
					func(datum *association.Association) {
						datum.ID = pointer.FromString(dataTest.RandomID())
						datum.Type = pointer.FromString("datum")
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

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *association.Association)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesCommonAssociationTest.NewAssociation()
						mutator(datum)
						expectedDatum := dataTypesCommonAssociationTest.CloneAssociation(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *association.Association) {},
				),
				Entry("does not modify the datum; id missing",
					func(datum *association.Association) { datum.ID = nil },
				),
				Entry("does not modify the datum; reason missing",
					func(datum *association.Association) { datum.Reason = nil },
				),
				Entry("does not modify the datum; type missing",
					func(datum *association.Association) { datum.Type = nil },
				),
				Entry("does not modify the datum; url missing",
					func(datum *association.Association) { datum.URL = nil },
				),
				Entry("does not modify the datum; all missing",
					func(datum *association.Association) { *datum = association.Association{} },
				),
			)
		})
	})

	Context("ParseAssociationArray", func() {
		// TODO
	})

	Context("NewAssociationArray", func() {
		It("is successful", func() {
			Expect(association.NewAssociationArray()).To(Equal(&association.AssociationArray{}))
		})
	})

	Context("AssociationArray", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *association.AssociationArray), expectedErrors ...error) {
					datum := dataTypesCommonAssociationTest.NewAssociationArray()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *association.AssociationArray) {},
				),
				Entry("empty",
					func(datum *association.AssociationArray) { *datum = *association.NewAssociationArray() },
					structureValidator.ErrorValueEmpty(),
				),
				Entry("nil",
					func(datum *association.AssociationArray) { *datum = append(*association.NewAssociationArray(), nil) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
				),
				Entry("single invalid",
					func(datum *association.AssociationArray) {
						invalid := dataTypesCommonAssociationTest.NewAssociation()
						invalid.Type = nil
						*datum = append(*association.NewAssociationArray(), invalid)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0/type"),
				),
				Entry("single valid",
					func(datum *association.AssociationArray) {
						*datum = append(*association.NewAssociationArray(), dataTypesCommonAssociationTest.NewAssociation())
					},
				),
				Entry("multiple invalid",
					func(datum *association.AssociationArray) {
						invalid := dataTypesCommonAssociationTest.NewAssociation()
						invalid.Type = nil
						*datum = append(*association.NewAssociationArray(), dataTypesCommonAssociationTest.NewAssociation(), invalid, dataTypesCommonAssociationTest.NewAssociation())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/type"),
				),
				Entry("multiple valid",
					func(datum *association.AssociationArray) {
						*datum = *dataTypesCommonAssociationTest.NewAssociationArray()
					},
				),
				Entry("multiple errors",
					func(datum *association.AssociationArray) {
						invalid := dataTypesCommonAssociationTest.NewAssociation()
						invalid.Type = nil
						*datum = append(*association.NewAssociationArray(), nil, invalid, dataTypesCommonAssociationTest.NewAssociation())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/type"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *association.AssociationArray)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesCommonAssociationTest.NewAssociationArray()
						mutator(datum)
						expectedDatum := dataTypesCommonAssociationTest.CloneAssociationArray(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *association.AssociationArray) {},
				),
				Entry("does not modify the datum; id missing",
					func(datum *association.AssociationArray) { (*datum)[0].ID = nil },
				),
				Entry("does not modify the datum; reason missing",
					func(datum *association.AssociationArray) { (*datum)[0].Reason = nil },
				),
				Entry("does not modify the datum; type missing",
					func(datum *association.AssociationArray) { (*datum)[0].Type = nil },
				),
				Entry("does not modify the datum; url missing",
					func(datum *association.AssociationArray) { (*datum)[0].URL = nil },
				),
			)
		})
	})
})
