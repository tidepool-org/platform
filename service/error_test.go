package service_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Error", func() {
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
					Status: 400,
					Title:  "test-title",
				}
				Expect(json.Marshal(err)).To(MatchJSON(`{
					"code": "test-code",
					"detail": "test-detail",
					"meta": "test-meta",
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
					Meta:   "test-meta",
				}
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
