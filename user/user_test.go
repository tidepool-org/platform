package user_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	authTest "github.com/tidepool-org/platform/auth/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureTest "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/user"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("User", func() {
	It("RoleClinic is expected", func() {
		Expect(user.RoleClinic).To(Equal("clinic"))
	})

	It("Roles returns expected", func() {
		Expect(user.Roles()).To(Equal([]string{"clinic"}))
	})

	Context("User", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *user.User)) {
				datum := userTest.RandomUser()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, userTest.NewObjectFromUser(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, userTest.NewObjectFromUser(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *user.User) {},
			),
			Entry("empty",
				func(datum *user.User) { *datum = user.User{} },
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *user.User), expectedErrors ...error) {
					expectedDatum := userTest.RandomUser()
					object := userTest.NewObjectFromUser(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &user.User{}
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(userTest.MatchUser(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *user.User) {},
				),
				Entry("user id invalid type",
					func(object map[string]interface{}, expectedDatum *user.User) {
						object["userid"] = true
						expectedDatum.UserID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/userid"),
				),
				Entry("user id valid",
					func(object map[string]interface{}, expectedDatum *user.User) {
						valid := userTest.RandomID()
						object["userid"] = valid
						expectedDatum.UserID = pointer.FromString(valid)
					},
				),
				Entry("username invalid type",
					func(object map[string]interface{}, expectedDatum *user.User) {
						object["username"] = true
						expectedDatum.Username = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/username"),
				),
				Entry("username valid",
					func(object map[string]interface{}, expectedDatum *user.User) {
						valid := userTest.RandomUsername()
						object["username"] = valid
						expectedDatum.Username = pointer.FromString(valid)
					},
				),
				Entry("emailVerified invalid type",
					func(object map[string]interface{}, expectedDatum *user.User) {
						object["emailVerified"] = "invalid"
						expectedDatum.EmailVerified = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotBool("invalid"), "/emailVerified"),
				),
				Entry("emailVerified valid",
					func(object map[string]interface{}, expectedDatum *user.User) {
						valid := test.RandomBool()
						object["emailVerified"] = valid
						expectedDatum.EmailVerified = pointer.FromBool(valid)
					},
				),
				Entry("terms accepted invalid type",
					func(object map[string]interface{}, expectedDatum *user.User) {
						object["termsAccepted"] = true
						expectedDatum.TermsAccepted = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/termsAccepted"),
				),
				Entry("terms accepted valid",
					func(object map[string]interface{}, expectedDatum *user.User) {
						valid := test.RandomTime().Format(time.RFC3339Nano)
						object["termsAccepted"] = valid
						expectedDatum.TermsAccepted = pointer.FromString(valid)
					},
				),
				Entry("roles invalid type",
					func(object map[string]interface{}, expectedDatum *user.User) {
						object["roles"] = true
						expectedDatum.Roles = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/roles"),
				),
				Entry("roles valid",
					func(object map[string]interface{}, expectedDatum *user.User) {
						valid := user.Roles()
						object["roles"] = valid
						expectedDatum.Roles = pointer.FromStringArray(valid)
					},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *user.User) {
						object["userid"] = true
						object["username"] = true
						object["emailVerified"] = "invalid"
						object["termsAccepted"] = true
						object["roles"] = true
						expectedDatum.UserID = nil
						expectedDatum.Username = nil
						expectedDatum.EmailVerified = nil
						expectedDatum.TermsAccepted = nil
						expectedDatum.Roles = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/userid"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/username"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotBool("invalid"), "/emailVerified"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/termsAccepted"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/roles"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *user.User), expectedErrors ...error) {
					datum := userTest.RandomUser()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *user.User) {},
				),
				Entry("user id missing",
					func(datum *user.User) { datum.UserID = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/userid"),
				),
				Entry("user id empty",
					func(datum *user.User) { datum.UserID = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/userid"),
				),
				Entry("user id invalid",
					func(datum *user.User) { datum.UserID = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(user.ErrorValueStringAsIDNotValid("invalid"), "/userid"),
				),
				Entry("user id valid",
					func(datum *user.User) { datum.UserID = pointer.FromString(userTest.RandomID()) },
				),
				Entry("username missing",
					func(datum *user.User) { datum.Username = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/username"),
				),
				Entry("username empty",
					func(datum *user.User) { datum.Username = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/username"),
				),
				Entry("username valid",
					func(datum *user.User) { datum.Username = pointer.FromString(userTest.RandomUsername()) },
				),
				Entry("terms accepted missing",
					func(datum *user.User) { datum.TermsAccepted = nil },
				),
				Entry("terms accepted invalid",
					func(datum *user.User) { datum.TermsAccepted = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("", time.RFC3339Nano), "/termsAccepted"),
				),
				Entry("terms accepted invalid",
					func(datum *user.User) { datum.TermsAccepted = pointer.FromString(time.Time{}.Format(time.RFC3339Nano)) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/termsAccepted"),
				),
				Entry("terms accepted valid",
					func(datum *user.User) {
						datum.TermsAccepted = pointer.FromString(test.RandomTime().Format(time.RFC3339Nano))
					},
				),
				Entry("roles missing",
					func(datum *user.User) { datum.Roles = nil },
				),
				Entry("roles empty",
					func(datum *user.User) { datum.Roles = pointer.FromStringArray([]string{}) },
				),
				Entry("roles invalid",
					func(datum *user.User) { datum.Roles = pointer.FromStringArray([]string{user.RoleClinic, "invalid"}) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", user.Roles()), "/roles/1"),
				),
				Entry("roles duplicate",
					func(datum *user.User) {
						datum.Roles = pointer.FromStringArray([]string{user.RoleClinic, user.RoleClinic})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/roles/1"),
				),
				Entry("roles valid",
					func(datum *user.User) { datum.Roles = pointer.FromStringArray(user.Roles()) },
				),
				Entry("multiple errors",
					func(datum *user.User) {
						datum.UserID = nil
						datum.Username = nil
						datum.TermsAccepted = pointer.FromString("")
						datum.Roles = pointer.FromStringArray([]string{user.RoleClinic, "invalid"})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/userid"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/username"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("", time.RFC3339Nano), "/termsAccepted"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", user.Roles()), "/roles/1"),
				),
			)
		})

		Context("with new user", func() {
			var datum *user.User

			BeforeEach(func() {
				datum = userTest.RandomUser()
			})

			Context("HasRole", func() {
				It("returns false if the user roles is missing", func() {
					datum.Roles = nil
					Expect(datum.HasRole(user.RoleClinic)).To(BeFalse())
				})

				It("returns false if the user roles is empty", func() {
					datum.Roles = pointer.FromStringArray([]string{})
					Expect(datum.HasRole(user.RoleClinic)).To(BeFalse())
				})

				It("returns false if the user does not have the role", func() {
					datum.Roles = pointer.FromStringArray([]string{"invalid"})
					Expect(datum.HasRole(user.RoleClinic)).To(BeFalse())
				})

				It("returns true if the user has the role", func() {
					datum.Roles = pointer.FromStringArray([]string{user.RoleClinic})
					Expect(datum.HasRole(user.RoleClinic)).To(BeTrue())
				})
			})

			Context("Sanitize", func() {
				var original *user.User

				BeforeEach(func() {
					original = userTest.CloneUser(datum)
				})

				It("does sanitize renditions if details is missing", func() {
					original.Username = nil
					original.EmailVerified = nil
					original.TermsAccepted = nil
					original.Roles = nil
					datum.Sanitize(nil)
					Expect(datum).To(Equal(original))
				})

				It("does sanitize renditions if details is not service and not matching user", func() {
					original.Username = nil
					original.EmailVerified = nil
					original.TermsAccepted = nil
					original.Roles = nil
					datum.Sanitize(request.NewDetails(request.MethodSessionToken, userTest.RandomID(), authTest.NewSessionToken()))
					Expect(datum).To(Equal(original))
				})

				It("does not sanitize renditions if details is service", func() {
					datum.Sanitize(request.NewDetails(request.MethodSessionToken, *original.UserID, authTest.NewSessionToken()))
					Expect(datum).To(Equal(original))
				})

				It("does not sanitize renditions if details is service", func() {
					datum.Sanitize(request.NewDetails(request.MethodServiceSecret, "", authTest.NewServiceSecret()))
					Expect(datum).To(Equal(original))
				})
			})
		})
	})

	Context("UserArray", func() {
		Context("Sanitize", func() {
			var datum user.UserArray
			var original user.UserArray

			BeforeEach(func() {
				datum = userTest.RandomUserArray(0, 2)
				original = userTest.CloneUserArray(datum)
			})

			It("does sanitize renditions if details is missing", func() {
				for index := range original {
					original[index].Username = nil
					original[index].EmailVerified = nil
					original[index].TermsAccepted = nil
					original[index].Roles = nil
				}
				datum.Sanitize(nil)
				Expect(datum).To(Equal(original))
			})

			It("does sanitize renditions if details is not service", func() {
				for index := range original {
					original[index].Username = nil
					original[index].EmailVerified = nil
					original[index].TermsAccepted = nil
					original[index].Roles = nil
				}
				datum.Sanitize(request.NewDetails(request.MethodSessionToken, userTest.RandomID(), authTest.NewSessionToken()))
				Expect(datum).To(Equal(original))
			})

			It("does not sanitize renditions if details is service", func() {
				datum.Sanitize(request.NewDetails(request.MethodServiceSecret, "", authTest.NewServiceSecret()))
				Expect(datum).To(Equal(original))
			})
		})
	})

	Context("ID", func() {
		Context("NewID", func() {
			It("returns a string of 10 lowercase hexidecimal characters", func() {
				Expect(user.NewID()).To(MatchRegexp("^[0-9a-f]{10}$"))
			})

			It("returns different IDs for each invocation", func() {
				Expect(user.NewID()).ToNot(Equal(user.NewID()))
			})
		})

		Context("IsValidID, IDValidator, and ValidateID", func() {
			DescribeTable("return the expected results when the input",
				func(value string, expectedErrors ...error) {
					Expect(user.IsValidID(value)).To(Equal(len(expectedErrors) == 0))
					errorReporter := structureTest.NewErrorReporter()
					user.IDValidator(value, errorReporter)
					errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
					errorsTest.ExpectEqual(user.ValidateID(value), expectedErrors...)
				},
				Entry("is an empty", "", structureValidator.ErrorValueEmpty()),
				Entry("has string length out of range (lower)", "01234abcd", user.ErrorValueStringAsIDNotValid("01234abcd")),
				Entry("has string length in range", test.RandomStringFromRangeAndCharset(10, 10, test.CharsetHexidecimalLowercase)),
				Entry("has string length out of range (upper)", "01234abcdef", user.ErrorValueStringAsIDNotValid("01234abcdef")),
				Entry("has uppercase characters", "01234ABCDE", user.ErrorValueStringAsIDNotValid("01234ABCDE")),
				Entry("has symbols", "012$%^&cde", user.ErrorValueStringAsIDNotValid("012$%^&cde")),
				Entry("has whitespace", "012    cde", user.ErrorValueStringAsIDNotValid("012    cde")),
			)
		})
	})
})
