package user_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/user"
)

var _ = Describe("User", func() {
	DescribeTable("HasRole",
		func(roles []string, role string, expectedResult bool) {
			testUser := &user.User{
				Roles: roles,
			}
			Expect(testUser.HasRole(role)).To(Equal(expectedResult))
		},
		Entry("roles is nil, role is empty", nil, "", false),
		Entry("roles is nil, role is specified", nil, user.ClinicRole, false),
		Entry("roles is empty, role is empty", []string{}, "", false),
		Entry("roles is empty, role is specified", []string{}, user.ClinicRole, false),
		Entry("roles has one, role is empty", []string{user.ClinicRole}, "", false),
		Entry("roles has one, role is specified, not in roles", []string{user.ClinicRole}, "unknown", false),
		Entry("roles has one, role is specified, in roles", []string{user.ClinicRole}, user.ClinicRole, true),
		Entry("roles has many, role is empty", []string{"administrator", user.ClinicRole, "manager"}, "", false),
		Entry("roles has many, role is specified, not in roles", []string{"administrator", user.ClinicRole, "manager"}, "unknown", false),
		Entry("roles has many, role is specified, in roles", []string{"administrator", user.ClinicRole, "manager"}, user.ClinicRole, true),
	)
})
