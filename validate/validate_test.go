package validate_test

import (
	"time"

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
			validator = NewPlatformValidator()
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
			validator = NewPlatformValidator()
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

	Context("datetime", func() {
		type testStruct struct {
			GivenDate string `json:"givenDate"  valid:"datetime"`
		}
		var (
			validator = NewPlatformValidator()
		)
		Context("as a string", func() {
			Context("is invalid when", func() {
				It("there is no date", func() {
					nodate := testStruct{GivenDate: ""}
					Expect(validator.ValidateStruct(nodate)).ToNot(BeNil())
				})
				It("the date is not the right spec", func() {
					wrongspec := testStruct{GivenDate: "Monday, 02 Jan 2016"}
					Expect(validator.ValidateStruct(wrongspec)).ToNot(BeNil())
				})
				It("the date does not include hours and mins", func() {
					notime := testStruct{GivenDate: "2016-02-05"}
					Expect(validator.ValidateStruct(notime)).ToNot(BeNil())
				})
				It("the date does not include mins", func() {
					notime := testStruct{GivenDate: "2016-02-05T20"}
					Expect(validator.ValidateStruct(notime)).ToNot(BeNil())
				})
			})
			Context("is valid when", func() {
				It("the date is RFC3339 formated - e.g. 1", func() {
					validdate := testStruct{GivenDate: "2016-03-14T20:22:21+13:00"}
					Expect(validator.ValidateStruct(validdate)).To(BeNil())
				})
				It("the date is RFC3339 formated - e.g. 2", func() {
					validdate := testStruct{GivenDate: "2016-02-05T15:53:00"}
					Expect(validator.ValidateStruct(validdate)).To(BeNil())
				})
				It("the date is RFC3339 formated - e.g. 3", func() {
					validdate := testStruct{GivenDate: "2016-02-05T15:53:00.000Z"}
					Expect(validator.ValidateStruct(validdate)).To(BeNil())
				})
			})
		})
		Context("as time", func() {
			type testStruct struct {
				GivenDate time.Time `json:"givenDate"  valid:"datetime"`
			}
			Context("is invalid when", func() {

				It("in the future", func() {
					furturedate := testStruct{GivenDate: time.Now().Add(time.Hour * 36)}
					Expect(validator.ValidateStruct(furturedate)).ToNot(BeNil())
				})
				It("zero", func() {
					zerodate := testStruct{}
					Expect(validator.ValidateStruct(zerodate)).ToNot(BeNil())
				})
			})
			Context("is valid when", func() {

				It("set now", func() {
					nowdate := testStruct{GivenDate: time.Now()}
					Expect(validator.ValidateStruct(nowdate)).To(BeNil())
				})
				It("set in the past", func() {
					pastdate := testStruct{GivenDate: time.Now().AddDate(0, -2, 0)}
					Expect(validator.ValidateStruct(pastdate)).To(BeNil())
				})
			})
		})

	})

})
