package structure_test

// import (
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"

// 	"github.com/tidepool-org/platform/structure"
// 	testStructure "github.com/tidepool-org/platform/structure/test"
// )

// var _ = Describe("Structure", func() {
// 	Context("NewParameterSource", func() {
// 		It("returns successfully", func() {
// 			Expect(structure.NewParameterSource()).ToNot(BeNil())
// 		})
// 	})

// 	Context("with new parameter source", func() {
// 		var parameterSource *structure.ParameterSource

// 		BeforeEach(func() {
// 			parameterSource = structure.NewParameterSource()
// 			Expect(parameterSource).ToNot(BeNil())
// 		})

// 		Context("Source", func() {
// 			It("returns nil if no parameter", func() {
// 				Expect(parameterSource.Source()).To(BeNil())
// 			})

// 			It("returns successfully with parameter", func() {
// 				reference := testStructure.NewReference()
// 				source := parameterSource.WithReference(reference).Source()
// 				Expect(source).ToNot(BeNil())
// 				Expect(source.Parameter).To(Equal(reference))
// 				Expect(source.Pointer).To(BeEmpty())
// 			})
// 		})

// 		Context("WithReference", func() {
// 			It("returns successfully with parameter", func() {
// 				reference := testStructure.NewReference()
// 				Expect(parameterSource.WithReference(reference)).ToNot(BeNil())
// 			})

// 			It("returns nil if already parameter", func() {
// 				reference1 := testStructure.NewReference()
// 				reference2 := testStructure.NewReference()
// 				Expect(parameterSource.WithReference(reference1).WithReference(reference2)).To(BeNil())
// 			})
// 		})
// 	})

// 	Context("NewPointerSource", func() {
// 		It("returns successfully", func() {
// 			Expect(structure.NewPointerSource()).ToNot(BeNil())
// 		})
// 	})

// 	Context("with new pointer source", func() {
// 		var pointerSource *structure.PointerSource

// 		BeforeEach(func() {
// 			pointerSource = structure.NewPointerSource()
// 			Expect(pointerSource).ToNot(BeNil())
// 		})

// 		Context("Source", func() {
// 			It("returns empty if no pointer", func() {
// 				source := pointerSource.Source()
// 				Expect(source).ToNot(BeNil())
// 				Expect(source.Parameter).To(BeEmpty())
// 				Expect(source.Pointer).To(BeEmpty())
// 			})

// 			It("returns successfully with pointer", func() {
// 				reference := testStructure.NewReference()
// 				source := pointerSource.WithReference(reference).Source()
// 				Expect(source).ToNot(BeNil())
// 				Expect(source.Parameter).To(BeEmpty())
// 				Expect(source.Pointer).To(Equal("/" + reference))
// 			})

// 			It("returns successfully with multiple pointers", func() {
// 				reference1 := testStructure.NewReference()
// 				reference2 := testStructure.NewReference()
// 				source := pointerSource.WithReference(reference1).WithReference(reference2).Source()
// 				Expect(source).ToNot(BeNil())
// 				Expect(source.Parameter).To(BeEmpty())
// 				Expect(source.Pointer).To(Equal("/" + reference1 + "/" + reference2))
// 			})

// 			It("returns successfully with multiple pointers that require encoding", func() {
// 				source := pointerSource.WithReference("ab~cd~ef").WithReference("12/34/56").Source()
// 				Expect(source).ToNot(BeNil())
// 				Expect(source.Parameter).To(BeEmpty())
// 				Expect(source.Pointer).To(Equal("/ab~0cd~0ef/12~134~156"))
// 			})
// 		})

// 		Context("WithReference", func() {
// 			It("returns successfully with pointer", func() {
// 				reference := testStructure.NewReference()
// 				Expect(pointerSource.WithReference(reference)).ToNot(BeNil())
// 			})

// 			It("returns successfully with multiple pointers", func() {
// 				reference1 := testStructure.NewReference()
// 				reference2 := testStructure.NewReference()
// 				Expect(pointerSource.WithReference(reference1).WithReference(reference2)).ToNot(BeNil())
// 			})
// 		})
// 	})

// 	Context("EncodePointerReference", func() {
// 		It("returns same string that does not require encoding", func() {
// 			reference := testStructure.NewReference()
// 			Expect(structure.EncodePointerReference(reference)).To(Equal(reference))
// 		})

// 		It("returns encoded string that does require encoding", func() {
// 			Expect(structure.EncodePointerReference("ab~cd/ef")).To(Equal("ab~0cd~1ef"))
// 		})
// 	})
// })
