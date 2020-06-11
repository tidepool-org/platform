package user_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	authTest "github.com/tidepool-org/platform/auth/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	requestTest "github.com/tidepool-org/platform/request/test"
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

	Context("Delete", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *user.Delete)) {
				datum := userTest.RandomDelete()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, userTest.NewObjectFromDelete(datum, test.ObjectFormatJSON))
			},
			Entry("succeeds",
				func(datum *user.Delete) {},
			),
			Entry("empty",
				func(datum *user.Delete) { *datum = user.Delete{} },
			),
		)

		Context("NewDelete", func() {
			It("returns successfully with default values", func() {
				Expect(user.NewDelete()).To(Equal(&user.Delete{}))
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *user.Delete), expectedErrors ...error) {
					expectedDatum := userTest.RandomDelete()
					object := userTest.NewObjectFromDelete(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &user.Delete{}
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *user.Delete) {},
				),
				Entry("password invalid type",
					func(object map[string]interface{}, expectedDatum *user.Delete) {
						object["password"] = true
						expectedDatum.Password = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/password"),
				),
				Entry("password valid",
					func(object map[string]interface{}, expectedDatum *user.Delete) {
						valid := userTest.RandomPassword()
						object["password"] = valid
						expectedDatum.Password = pointer.FromString(valid)
					},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *user.Delete) {
						object["password"] = true
						expectedDatum.Password = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/password"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *user.Delete), expectedErrors ...error) {
					datum := userTest.RandomDelete()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *user.Delete) {},
				),
				Entry("password missing",
					func(datum *user.Delete) { datum.Password = nil },
				),
				Entry("password empty",
					func(datum *user.Delete) { datum.Password = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/password"),
				),
				Entry("password valid",
					func(datum *user.Delete) { datum.Password = pointer.FromString(userTest.RandomPassword()) },
				),
				Entry("multiple errors",
					func(datum *user.Delete) {
						datum.Password = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/password"),
				),
			)
		})
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
			Entry("with modified time",
				func(datum *user.User) {
					datum.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.CreatedTime, time.Now()).Truncate(time.Second))
				},
			),
			Entry("with deleted time",
				func(datum *user.User) {
					datum.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.CreatedTime, time.Now()).Truncate(time.Second))
					datum.DeletedTime = pointer.CloneTime(datum.ModifiedTime)
				},
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
				Entry("authenticated invalid type",
					func(object map[string]interface{}, expectedDatum *user.User) {
						object["authenticated"] = "invalid"
						expectedDatum.Authenticated = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotBool("invalid"), "/authenticated"),
				),
				Entry("authenticated valid",
					func(object map[string]interface{}, expectedDatum *user.User) {
						valid := test.RandomBool()
						object["authenticated"] = valid
						expectedDatum.Authenticated = pointer.FromBool(valid)
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
				Entry("created time invalid type",
					func(object map[string]interface{}, expectedDatum *user.User) {
						object["createdTime"] = true
						expectedDatum.CreatedTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/createdTime"),
				),
				Entry("created time invalid",
					func(object map[string]interface{}, expectedDatum *user.User) {
						object["createdTime"] = "invalid"
						expectedDatum.CreatedTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/createdTime"),
				),
				Entry("created time valid",
					func(object map[string]interface{}, expectedDatum *user.User) {
						valid := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second)
						object["createdTime"] = valid.Format(time.RFC3339Nano)
						expectedDatum.CreatedTime = pointer.FromTime(valid)
					},
				),
				Entry("modified time invalid type",
					func(object map[string]interface{}, expectedDatum *user.User) {
						object["modifiedTime"] = true
						expectedDatum.ModifiedTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/modifiedTime"),
				),
				Entry("modified time invalid",
					func(object map[string]interface{}, expectedDatum *user.User) {
						object["modifiedTime"] = "invalid"
						expectedDatum.ModifiedTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/modifiedTime"),
				),
				Entry("modified time valid",
					func(object map[string]interface{}, expectedDatum *user.User) {
						valid := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second)
						object["modifiedTime"] = valid.Format(time.RFC3339Nano)
						expectedDatum.ModifiedTime = pointer.FromTime(valid)
					},
				),
				Entry("deleted time invalid type",
					func(object map[string]interface{}, expectedDatum *user.User) {
						object["deletedTime"] = true
						expectedDatum.DeletedTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/deletedTime"),
				),
				Entry("deleted time invalid",
					func(object map[string]interface{}, expectedDatum *user.User) {
						object["deletedTime"] = "invalid"
						expectedDatum.DeletedTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/deletedTime"),
				),
				Entry("deleted time valid",
					func(object map[string]interface{}, expectedDatum *user.User) {
						valid := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second)
						object["deletedTime"] = valid.Format(time.RFC3339Nano)
						expectedDatum.DeletedTime = pointer.FromTime(valid)
					},
				),
				Entry("revision invalid type",
					func(object map[string]interface{}, expectedDatum *user.User) {
						object["revision"] = true
						expectedDatum.Revision = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/revision"),
				),
				Entry("revision valid",
					func(object map[string]interface{}, expectedDatum *user.User) {
						valid := requestTest.RandomRevision()
						object["revision"] = valid
						expectedDatum.Revision = pointer.FromInt(valid)
					},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *user.User) {
						object["userid"] = true
						object["username"] = true
						object["authenticated"] = "invalid"
						object["termsAccepted"] = true
						object["roles"] = true
						object["createdTime"] = true
						object["modifiedTime"] = true
						object["deletedTime"] = true
						object["revision"] = true
						expectedDatum.UserID = nil
						expectedDatum.Username = nil
						expectedDatum.Authenticated = nil
						expectedDatum.TermsAccepted = nil
						expectedDatum.Roles = nil
						expectedDatum.CreatedTime = nil
						expectedDatum.ModifiedTime = nil
						expectedDatum.DeletedTime = nil
						expectedDatum.Revision = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/userid"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/username"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotBool("invalid"), "/authenticated"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/termsAccepted"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/roles"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/createdTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/modifiedTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/deletedTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/revision"),
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
				Entry("created time zero",
					func(datum *user.User) { datum.CreatedTime = pointer.FromTime(time.Time{}) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/createdTime"),
				),
				Entry("created time after now",
					func(datum *user.User) {
						datum.CreatedTime = pointer.FromTime(test.FutureFarTime())
						datum.ModifiedTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/createdTime"),
				),
				Entry("created time valid",
					func(datum *user.User) {
						datum.CreatedTime = pointer.FromTime(test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second))
						datum.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.CreatedTime, time.Now()).Truncate(time.Second))
					},
				),
				Entry("modified time after now",
					func(datum *user.User) { datum.ModifiedTime = pointer.FromTime(test.FutureFarTime()) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/modifiedTime"),
				),
				Entry("modified time valid",
					func(datum *user.User) {
						datum.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.CreatedTime, time.Now()).Truncate(time.Second))
					},
				),
				Entry("deleted time after now",
					func(datum *user.User) { datum.DeletedTime = pointer.FromTime(test.FutureFarTime()) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/deletedTime"),
				),
				Entry("deleted time valid",
					func(datum *user.User) {
						datum.DeletedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.CreatedTime, time.Now()).Truncate(time.Second))
					},
				),
				Entry("revision missing",
					func(datum *user.User) {
						datum.Revision = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/revision"),
				),
				Entry("revision out of range (lower)",
					func(datum *user.User) {
						datum.Revision = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/revision"),
				),
				Entry("revision in range (lower)",
					func(datum *user.User) {
						datum.Revision = pointer.FromInt(0)
					},
				),
				Entry("multiple errors",
					func(datum *user.User) {
						datum.UserID = nil
						datum.Username = nil
						datum.TermsAccepted = pointer.FromString("")
						datum.Roles = pointer.FromStringArray([]string{user.RoleClinic, "invalid"})
						datum.ModifiedTime = pointer.FromTime(test.FutureFarTime())
						datum.DeletedTime = pointer.FromTime(test.FutureFarTime())
						datum.Revision = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/userid"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/username"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("", time.RFC3339Nano), "/termsAccepted"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", user.Roles()), "/roles/1"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/modifiedTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/deletedTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/revision"),
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
					original.Authenticated = nil
					original.TermsAccepted = nil
					original.Roles = nil
					datum.Sanitize(nil)
					Expect(datum).To(Equal(original))
				})

				It("does sanitize renditions if details is not service and not matching user", func() {
					original.Username = nil
					original.Authenticated = nil
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
					original[index].Authenticated = nil
					original[index].TermsAccepted = nil
					original[index].Roles = nil
				}
				datum.Sanitize(nil)
				Expect(datum).To(Equal(original))
			})

			It("does sanitize renditions if details is not service", func() {
				for index := range original {
					original[index].Username = nil
					original[index].Authenticated = nil
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
