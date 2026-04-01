package auth_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	structureTest "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Scope", func() {
	DescribeTable("ParseScope",
		func(scope string, expectedScope []string, expectedError error) {
			actualScope, actualErr := auth.ParseScope(scope)
			if expectedError != nil {
				Expect(actualErr).To(MatchError(expectedError.Error()))
			} else {
				Expect(actualErr).ToNot(HaveOccurred())
			}
			Expect(actualScope).To(Equal(expectedScope))
		},
		Entry("returns error for invalid scope token", "read write\"", nil, errors.New("scope token is invalid")),
		Entry("returns empty slice for empty string", "", nil, nil),
		Entry("returns empty slice for whitespace-only string", "   ", nil, nil),
		Entry("parses a single scope token", "read", []string{"read"}, nil),
		Entry("parses multiple scope tokens", "read write", []string{"read", "write"}, nil),
		Entry("trims whitespace from scope tokens", "  read    write  ", []string{"read", "write"}, nil),
		Entry("sorts scope tokens alphabetically", "write read delete", []string{"delete", "read", "write"}, nil),
		Entry("handles scope tokens with special characters", "user:read user:write", []string{"user:read", "user:write"}, nil),
		Entry("handles scope tokens with dots", "data.read data.write", []string{"data.read", "data.write"}, nil),
		Entry("handles scope tokens with hyphens", "read-all write-all", []string{"read-all", "write-all"}, nil),
		Entry("handles duplicate scope tokens", "read write read", []string{"read", "write"}, nil),
		Entry("handles mixed case scope tokens", "Read WRITE delete", []string{"Read", "WRITE", "delete"}, nil),
	)

	DescribeTable("JoinScope",
		func(scope []string, expectedScope string) {
			Expect(auth.JoinScope(scope)).To(Equal(expectedScope))
		},
		Entry("joins empty slice", []string{}, ""),
		Entry("joins single scope token", []string{"read"}, "read"),
		Entry("joins multiple scope tokens", []string{"read", "write"}, "read write"),
		Entry("joins scope tokens with special characters", []string{"user:read", "user:write"}, "user:read user:write"),
		Entry("joins scope tokens with dots", []string{"data.read", "data.write"}, "data.read data.write"),
		Entry("joins scope tokens with hyphens", []string{"read-all", "write-all"}, "read-all write-all"),
	)

	DescribeTable("IsValidScopeToken",
		func(scopeToken string, expectedValid bool) {
			Expect(auth.IsValidScopeToken(scopeToken)).To(Equal(expectedValid))
		},
		Entry("returns false for empty string", "", false),
		Entry("returns false for whitespace-only string", "   ", false),
		Entry("returns true for valid scope tokens", "read", true),
		Entry("returns true for valid scope tokens with special characters", "user:read", true),
		Entry("returns true for valid scope tokens with dots", "data.read", true),
		Entry("returns true for valid scope tokens with hyphens", "read-all", true),
		Entry("returns false for scope tokens with spaces", "read write", false),
		Entry("returns false for scope tokens with control characters", "write\n", false),
	)

	DescribeTable("ScopeTokenValidator",
		func(scopeToken string, expectedError error) {
			errorReporter := structureTest.NewErrorReporter()
			auth.ScopeTokenValidator(scopeToken, errorReporter)
			actualError := errorReporter.Error()
			if expectedError != nil {
				Expect(actualError).To(MatchError(expectedError.Error()))
			} else {
				Expect(actualError).ToNot(HaveOccurred())
			}
		},
		Entry("returns error for empty string", "", structureValidator.ErrorValueEmpty()),
		Entry("returns error for whitespace-only string", "   ", auth.ErrorValueStringAsScopeTokenNotValid("   ")),
		Entry("returns nil for valid scope tokens", "read", nil),
		Entry("returns nil for valid scope tokens with special characters", "user:read", nil),
		Entry("returns nil for valid scope tokens with dots", "data.read", nil),
		Entry("returns nil for valid scope tokens with hyphens", "read-all", nil),
		Entry("returns error for scope tokens with spaces", "read write", auth.ErrorValueStringAsScopeTokenNotValid("read write")),
		Entry("returns error for scope tokens with control characters", "write\n", auth.ErrorValueStringAsScopeTokenNotValid("write\n")),
	)

	DescribeTable("ValidateScopeToken",
		func(scopeToken string, expectedError error) {
			actualError := auth.ValidateScopeToken(scopeToken)
			if expectedError != nil {
				Expect(actualError).To(MatchError(expectedError.Error()))
			} else {
				Expect(actualError).ToNot(HaveOccurred())
			}
		},
		Entry("returns error for empty string", "", structureValidator.ErrorValueEmpty()),
		Entry("returns error for whitespace-only string", "   ", auth.ErrorValueStringAsScopeTokenNotValid("   ")),
		Entry("returns nil for valid scope tokens", "read", nil),
		Entry("returns nil for valid scope tokens with special characters", "user:read", nil),
		Entry("returns nil for valid scope tokens with dots", "data.read", nil),
		Entry("returns nil for valid scope tokens with hyphens", "read-all", nil),
		Entry("returns error for scope tokens with spaces", "read write", auth.ErrorValueStringAsScopeTokenNotValid("read write")),
		Entry("returns error for scope tokens with control characters", "write\n", auth.ErrorValueStringAsScopeTokenNotValid("write\n")),
	)
})
