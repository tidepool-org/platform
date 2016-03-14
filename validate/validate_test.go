package validate_test

import (
	. "github.com/tidepool-org/platform/validate"

	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
)

var _ = Describe("Validate", func() {

	Context("basic struct", func() {
		type testStruct struct {
			ID      string      `json:"id"  valid:"required"`
			Type    string      `json:"type"  valid:"required"`
			Offset  int         `json:"offset,omitempty"  valid:"omitempty,min=-10,max=10"`
			Payload interface{} `json:"payload,omitempty"  valid:"-"`
		}
		var (
			validator = PlatformValidator{}
		)
		Context("is valid", func() {
			It("all feilds set correctly", func() {
				all := testStruct{ID: "someid", Type: "test", Offset: -9, Payload: map[string]string{"some": "stuff"}}
				Expect(validator.ValidateStruct(all)).To(BeNil())
			})
			It("only core feilds set correctly", func() {
				core := testStruct{ID: "someid", Type: "test"}
				Expect(validator.ValidateStruct(core)).To(BeNil())
			})
		})
		Context("is invalid when", func() {
			It("expected feilds aren't set", func() {
				none := testStruct{}
				Expect(validator.ValidateStruct(none)).ToNot(BeNil())
			})
			It("expected feilds are incorrectly set", func() {
				invalid := testStruct{ID: "", Type: ""}
				Expect(validator.ValidateStruct(invalid)).ToNot(BeNil())
			})
			It("optional feild set are incorrectly", func() {
				invalid := testStruct{ID: "someid", Type: "a_type", Offset: -100}
				Expect(validator.ValidateStruct(invalid)).ToNot(BeNil())
			})
		})
	})

	Context("nested struct", func() {
		type nestedUser struct {
			Firstname string `json:"firstname"  valid:"required"`
			Lastname  string `json:"lastname,omitempty"  valid:"omitempty,required"`
		}
		type testStruct struct {
			ID     string `json:"id"  valid:"required"`
			Offset int    `json:"offset,omitempty"  valid:"omitempty,required"`
			User   nestedUser
		}
		var (
			validator = PlatformValidator{}
		)
		Context("is valid when", func() {
			It("all feilds set correctly", func() {
				all := testStruct{ID: "someid", Offset: 9, User: nestedUser{Firstname: "first", Lastname: "last"}}
				Expect(validator.ValidateStruct(all)).To(BeNil())
			})
			It("only core feilds set correctly", func() {
				core := testStruct{ID: "someid", User: nestedUser{Firstname: "first"}}
				Expect(validator.ValidateStruct(core)).To(BeNil())
			})
		})
		Context("is invalid when", func() {
			It("expected feilds aren't set", func() {
				none := testStruct{}
				Expect(validator.ValidateStruct(none)).ToNot(BeNil())
			})
			It("nested core feilds aren't set", func() {
				core := testStruct{ID: "someid", User: nestedUser{Firstname: ""}}
				Expect(validator.ValidateStruct(core)).ToNot(BeNil())
			})
			It("expected feilds are incorrectly set", func() {
				invalid := testStruct{ID: "", User: nestedUser{Firstname: ""}}
				Expect(validator.ValidateStruct(invalid)).ToNot(BeNil())
			})
		})
	})

})
