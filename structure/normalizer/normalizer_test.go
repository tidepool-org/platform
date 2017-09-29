package normalizer_test

// import (
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"

// 	"github.com/tidepool-org/platform/structure"
// 	structureNormalizer "github.com/tidepool-org/platform/structure/normalizer"
// 	testStructure "github.com/tidepool-org/platform/structure/test"
// 	"github.com/tidepool-org/platform/test"
// )

// var _ = Describe("Normalizer", func() {
// 	var base *testStructure.Base

// 	BeforeEach(func() {
// 		base = testStructure.NewBase()
// 	})

// 	AfterEach(func() {
// 		base.Expectations()
// 	})

// 	Context("New", func() {
// 		It("returns successfully", func() {
// 			Expect(structureNormalizer.New()).ToNot(BeNil())
// 		})
// 	})

// 	Context("NewNormalizer", func() {
// 		It("returns successfully", func() {
// 			Expect(structureNormalizer.NewNormalizer(base)).ToNot(BeNil())
// 		})
// 	})

// 	Context("with new normalizer", func() {
// 		var normalizer *structureNormalizer.Normalizer

// 		BeforeEach(func() {
// 			normalizer = structureNormalizer.NewNormalizer(base)
// 			Expect(normalizer).ToNot(BeNil())
// 		})

// 		Context("WithSource", func() {
// 			It("returns new normalizer", func() {
// 				base.WithSourceOutputs = []structure.Base{testStructure.NewBase()}
// 				withSource := testStructure.NewSource()
// 				result := normalizer.WithSource(withSource)
// 				Expect(result).ToNot(BeNil())
// 				Expect(result).ToNot(Equal(normalizer))
// 				Expect(base.WithSourceInputs).To(Equal([]structure.Source{withSource}))
// 			})
// 		})

// 		Context("WithMeta", func() {
// 			It("returns new normalizer", func() {
// 				base.WithMetaOutputs = []structure.Base{testStructure.NewBase()}
// 				withMeta := test.NewText(1, 128)
// 				result := normalizer.WithMeta(withMeta)
// 				Expect(result).ToNot(BeNil())
// 				Expect(result).ToNot(Equal(normalizer))
// 				Expect(base.WithMetaInputs).To(Equal([]interface{}{withMeta}))
// 			})
// 		})

// 		Context("WithReference", func() {
// 			It("returns new normalizer", func() {
// 				base.WithReferenceOutputs = []structure.Base{testStructure.NewBase()}
// 				withReference := testStructure.NewReference()
// 				result := normalizer.WithReference(withReference)
// 				Expect(result).ToNot(BeNil())
// 				Expect(result).ToNot(Equal(normalizer))
// 				Expect(base.WithReferenceInputs).To(Equal([]string{withReference}))
// 			})
// 		})
// 	})
// })
