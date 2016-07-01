package service_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"encoding/json"

	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Error", func() {
	Context("Source", func() {
		Context("encoded as JSON", func() {
			It("is an empty object if no fields are specified", func() {
				source := &service.Source{}
				Expect(json.Marshal(source)).To(MatchJSON(`{}`))
			})

			It("is a populated object if fields are specified", func() {
				source := &service.Source{
					Parameter: "test-parameter",
					Pointer:   "test-pointer",
				}
				Expect(json.Marshal(source)).To(MatchJSON(`{
					"parameter": "test-parameter",
					"pointer": "test-pointer"
				}`))
			})
		})
	})

	Context("Error", func() {
		Context("encoded as JSON", func() {
			It("is an empty object if no fields are specified", func() {
				err := &service.Error{}
				Expect(json.Marshal(err)).To(MatchJSON(`{}`))
			})

			It("is a populated object if fields are specified", func() {
				err := &service.Error{
					Code:   "test-code",
					Detail: "test-detail",
					Meta:   "test-meta",
					Source: &service.Source{
						Parameter: "test-parameter",
						Pointer:   "test-pointer",
					},
					Status: 400,
					Title:  "test-title",
				}
				Expect(json.Marshal(err)).To(MatchJSON(`{
					"code": "test-code",
					"detail": "test-detail",
					"meta": "test-meta",
					"source": {
						"parameter": "test-parameter",
						"pointer": "test-pointer"
					},
					"status": "400",
					"title": "test-title"
				}`))
			})
		})

		Context("with an error", func() {
			var err *service.Error

			BeforeEach(func() {
				err = &service.Error{
					Code:   "test-error",
					Title:  "test error",
					Detail: "Test error",
					Source: &service.Source{
						Parameter: "test-parameter",
						Pointer:   "test-pointer",
					},
					Meta: "test-meta",
				}
			})

			Context("WithSourceParameter", func() {
				It("sets the parameter", func() {
					err.WithSourceParameter("new-test-parameter")
					Expect(err.Source).ToNot(BeNil())
					Expect(err.Source.Parameter).To(Equal("new-test-parameter"))
				})

				It("sets an empty parameter", func() {
					err.WithSourceParameter("")
					Expect(err.Source).ToNot(BeNil())
					Expect(err.Source.Parameter).To(Equal(""))
				})

				It("sets the parameter even if parameter is initially missing", func() {
					err.Source.Parameter = ""
					err.WithSourceParameter("new-test-parameter")
					Expect(err.Source).ToNot(BeNil())
					Expect(err.Source.Parameter).To(Equal("new-test-parameter"))
				})

				It("sets the parameter even if source is initially missing", func() {
					err.Source = nil
					err.WithSourceParameter("new-test-parameter")
					Expect(err.Source).ToNot(BeNil())
					Expect(err.Source.Parameter).To(Equal("new-test-parameter"))
				})
			})

			Context("WithSourcePointer", func() {
				It("sets the pointer", func() {
					err.WithSourcePointer("new-test-pointer")
					Expect(err.Source).ToNot(BeNil())
					Expect(err.Source.Pointer).To(Equal("new-test-pointer"))
				})

				It("sets an empty pointer", func() {
					err.WithSourcePointer("")
					Expect(err.Source).ToNot(BeNil())
					Expect(err.Source.Pointer).To(Equal(""))
				})

				It("sets the pointer even if pointer is initially missing", func() {
					err.Source.Pointer = ""
					err.WithSourcePointer("new-test-pointer")
					Expect(err.Source).ToNot(BeNil())
					Expect(err.Source.Pointer).To(Equal("new-test-pointer"))
				})

				It("sets the pointer even if source is initially missing", func() {
					err.Source = nil
					err.WithSourcePointer("new-test-pointer")
					Expect(err.Source).ToNot(BeNil())
					Expect(err.Source.Pointer).To(Equal("new-test-pointer"))
				})
			})

			Context("WithMeta", func() {
				It("sets the meta", func() {
					err.WithMeta("new-test-meta")
					Expect(err.Meta).To(Equal("new-test-meta"))
				})

				It("sets a missing meta", func() {
					err.WithMeta(nil)
					Expect(err.Meta).To(BeNil())
				})

				It("sets the meta even if meta is initially missing", func() {
					err.Meta = nil
					err.WithMeta("new-test-meta")
					Expect(err.Meta).To(Equal("new-test-meta"))
				})
			})
		})
	})
})
