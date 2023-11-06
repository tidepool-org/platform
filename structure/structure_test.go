package structure_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/structure"
	structureTest "github.com/tidepool-org/platform/structure/test"
)

var _ = Describe("Structure", func() {
	It("Origins returns expected", func() {
		Expect(structure.Origins()).To(Equal([]structure.Origin{0, 1, 2}))
	})

	Context("NewParameterSource", func() {
		It("returns successfully", func() {
			Expect(structure.NewParameterSource()).ToNot(BeNil())
		})
	})

	Context("with new parameter source", func() {
		var src structure.Source

		BeforeEach(func() {
			src = structure.NewParameterSource()
			Expect(src).ToNot(BeNil())
		})

		Context("with no references", func() {
			Context("Parameter", func() {
				It("returns empty string", func() {
					Expect(src.Parameter()).To(BeEmpty())
				})
			})

			Context("Pointer", func() {
				It("returns empty string", func() {
					Expect(src.Pointer()).To(BeEmpty())
				})
			})

			Context("WithReference", func() {
				It("returns a source", func() {
					Expect(src.WithReference(structureTest.NewReference())).ToNot(BeNil())
				})
			})
		})

		Context("with one reference", func() {
			var reference string

			BeforeEach(func() {
				reference = structureTest.NewReference()
				src = src.WithReference(reference)
			})

			Context("Parameter", func() {
				It("returns reference", func() {
					Expect(src.Parameter()).To(Equal(reference))
				})
			})

			Context("Pointer", func() {
				It("returns empty string", func() {
					Expect(src.Pointer()).To(BeEmpty())
				})
			})

			Context("WithReference", func() {
				It("returns the same source", func() {
					Expect(src.WithReference(structureTest.NewReference())).To(BeIdenticalTo(src))
				})
			})
		})

		Context("with one reference containing slash", func() {
			var referenceLeft string
			var referenceRight string
			var reference string

			BeforeEach(func() {
				referenceLeft = structureTest.NewReference()
				referenceRight = structureTest.NewReference()
				reference = fmt.Sprintf("%s/%s", referenceLeft, referenceRight)
				src = src.WithReference(reference)
			})

			Context("Parameter", func() {
				It("returns reference", func() {
					Expect(src.Parameter()).To(Equal(reference))
				})
			})

			Context("Pointer", func() {
				It("returns empty string", func() {
					Expect(src.Pointer()).To(BeEmpty())
				})
			})

			Context("WithReference", func() {
				It("returns the same source", func() {
					Expect(src.WithReference(structureTest.NewReference())).To(BeIdenticalTo(src))
				})
			})
		})
	})

	Context("NewPointerSource", func() {
		It("returns successfully", func() {
			Expect(structure.NewPointerSource()).ToNot(BeNil())
		})
	})

	Context("with new pointer source", func() {
		var src structure.Source

		BeforeEach(func() {
			src = structure.NewPointerSource()
			Expect(src).ToNot(BeNil())
		})

		Context("with no references", func() {
			Context("Parameter", func() {
				It("returns empty string", func() {
					Expect(src.Parameter()).To(BeEmpty())
				})
			})

			Context("Pointer", func() {
				It("returns empty string", func() {
					Expect(src.Pointer()).To(BeEmpty())
				})
			})

			Context("WithReference", func() {
				It("returns a source", func() {
					Expect(src.WithReference(structureTest.NewReference())).ToNot(BeNil())
				})
			})
		})

		Context("with one reference", func() {
			var reference1 string

			BeforeEach(func() {
				reference1 = structureTest.NewReference()
				src = src.WithReference(reference1)
			})

			Context("Parameter", func() {
				It("returns empty string", func() {
					Expect(src.Parameter()).To(BeEmpty())
				})
			})

			Context("Pointer", func() {
				It("returns path reference", func() {
					Expect(src.Pointer()).To(Equal(fmt.Sprintf("/%s", reference1)))
				})
			})

			Context("WithReference", func() {
				It("returns a source", func() {
					Expect(src.WithReference(structureTest.NewReference())).ToNot(BeNil())
				})
			})

			Context("with a second references", func() {
				var reference2 string

				BeforeEach(func() {
					reference2 = structureTest.NewReference()
					src = src.WithReference(reference2)
				})

				Context("Parameter", func() {
					It("returns empty string", func() {
						Expect(src.Parameter()).To(BeEmpty())
					})
				})

				Context("Pointer", func() {
					It("returns path reference", func() {
						Expect(src.Pointer()).To(Equal(fmt.Sprintf("/%s/%s", reference1, reference2)))
					})
				})

				Context("WithReference", func() {
					It("returns a source", func() {
						Expect(src.WithReference(structureTest.NewReference())).ToNot(BeNil())
					})
				})

				Context("with a third references", func() {
					var reference3 string

					BeforeEach(func() {
						reference3 = structureTest.NewReference()
						src = src.WithReference(reference3)
					})

					Context("Parameter", func() {
						It("returns empty string", func() {
							Expect(src.Parameter()).To(BeEmpty())
						})
					})

					Context("Pointer", func() {
						It("returns path reference", func() {
							Expect(src.Pointer()).To(Equal(fmt.Sprintf("/%s/%s/%s", reference1, reference2, reference3)))
						})
					})

					Context("WithReference", func() {
						It("returns a source", func() {
							Expect(src.WithReference(structureTest.NewReference())).ToNot(BeNil())
						})
					})
				})
			})
		})

		Context("with one reference containing slash", func() {
			var referenceLeft string
			var referenceRight string
			var reference string

			BeforeEach(func() {
				referenceLeft = structureTest.NewReference()
				referenceRight = structureTest.NewReference()
				reference = fmt.Sprintf("%s/%s", referenceLeft, referenceRight)
				src = src.WithReference(reference)
			})

			Context("Parameter", func() {
				It("returns empty string", func() {
					Expect(src.Parameter()).To(BeEmpty())
				})
			})

			Context("Pointer", func() {
				It("returns encoded reference", func() {
					Expect(src.Pointer()).To(Equal(fmt.Sprintf("/%s~1%s", referenceLeft, referenceRight)))
				})
			})

			Context("WithReference", func() {
				It("returns a source", func() {
					Expect(src.WithReference(structureTest.NewReference())).ToNot(BeNil())
				})
			})
		})
	})

	Context("EncodePointerReference", func() {
		It("returns same string that does not require encoding", func() {
			reference := structureTest.NewReference()
			Expect(structure.EncodePointerReference(reference)).To(Equal(reference))
		})

		It("returns encoded string that does require encoding", func() {
			Expect(structure.EncodePointerReference("ab~/cd/~ef")).To(Equal("ab~0~1cd~1~0ef"))
		})
	})
})
