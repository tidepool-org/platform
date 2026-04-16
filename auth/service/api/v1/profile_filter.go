package v1

import (
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/pointer"
	userLib "github.com/tidepool-org/platform/user"
)

var (
	ANY   = []string{"any"}
	NONE  = []string{"none"}
	TRUES = []string{"true", "yes", "y", "1"}
)

type parsedQueryPermissions []string

type usersProfileFilter struct {
	TrustorPermissions parsedQueryPermissions
	TrusteePermissions parsedQueryPermissions
	Email              *regexp.Regexp
	EmailVerified      *bool
	TermsAccepted      *regexp.Regexp
	Name               *regexp.Regexp
	Birthday           *regexp.Regexp
	DiagnosisDate      *regexp.Regexp
}

// parsePermissions replicates the functionality of the seagull node.js
// parsePermissions function which parses a query string value to a permissions
// slice
func parsePermissions(queryValuesOfKey string) parsedQueryPermissions {
	queryValuesOfKey = strings.TrimSpace(queryValuesOfKey)
	if queryValuesOfKey == "" {
		return nil
	}
	var vals []string
	for _, val := range strings.Split(queryValuesOfKey, ",") {
		val = strings.TrimSpace(val)
		// Remove falsey values that are strings
		if val == "0" || val == "false" || val == "" {
			continue
		}
		vals = append(vals, val)
	}
	if slices.Compare(vals, ANY) == 0 {
		return slices.Clone(ANY)
	}
	if slices.Compare(vals, NONE) == 0 {
		return slices.Clone(NONE)
	}
	for _, val := range vals {
		nonEmpty := val != ""
		if !nonEmpty {
			return vals
		}
	}
	return nil
}

// This is the logic I'm most unsure about.
// It's seems to be saying if a parsed query contains either of ANY or NONE that it is not a valid permissions query.
// Yet later in arePermissionsValid it allows those values.
func arePermissionsValid(perms parsedQueryPermissions) bool {
	if len(perms) > 1 {
		union := append(ANY, NONE...)
		for _, perm := range perms {
			// quadratic time complexity but very few elements so don't care
			if slices.Contains(union, perm) {
				return false
			}
		}
	}
	return true
}

func arePermissionsSatisfied(queryPermissions parsedQueryPermissions, userPermissions permission.Permission) bool {
	if slices.Compare(queryPermissions, ANY) == 0 {
		nonEmpty := len(userPermissions) > 0
		return nonEmpty
	}
	if slices.Compare(queryPermissions, NONE) == 0 {
		empty := len(userPermissions) == 0
		return empty
	}
	// Todo: really test this part
	for _, queryPerm := range queryPermissions {
		if _, ok := userPermissions[queryPerm]; !ok {
			return false
		}
	}
	return true
}

func stringToBoolean(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	return slices.Contains(TRUES, value)
}

func parseUsersQuery(query url.Values) *usersProfileFilter {
	var filter usersProfileFilter
	trustorPermissions := parsePermissions(query.Get("trustorPermissions"))
	// The original seagull code checks for nil so I don't know if just checking
	// for a zero length array is safe or not but I believe so - todo add test for this
	if len(trustorPermissions) > 0 {
		filter.TrustorPermissions = trustorPermissions
	}

	trusteePermissions := parsePermissions(query.Get("trusteePermissions"))
	if len(trusteePermissions) > 0 {
		filter.TrusteePermissions = trusteePermissions
	}

	if email := strings.TrimSpace(query.Get("email")); email != "" {
		if regex, err := regexp.Compile("(?i)" + email); err == nil {
			filter.Email = regex
		}
	}

	if emailVerified := strings.TrimSpace(query.Get("emailVerified")); emailVerified != "" {
		filter.EmailVerified = pointer.FromBool(stringToBoolean(emailVerified))
	}

	if termsAccepted := strings.TrimSpace(query.Get("termsAccepted")); termsAccepted != "" {
		if regex, err := regexp.Compile("(?i)" + termsAccepted); err == nil {
			filter.TermsAccepted = regex
		}
	}

	if name := strings.TrimSpace(query.Get("name")); name != "" {
		if regex, err := regexp.Compile("(?i)" + name); err == nil {
			filter.Name = regex
		}
	}

	if birthday := strings.TrimSpace(query.Get("birthday")); birthday != "" {
		if regex, err := regexp.Compile("(?i)" + birthday); err == nil {
			filter.Birthday = regex
		}
	}

	if diagnosisDate := strings.TrimSpace(query.Get("diagnosisDate")); diagnosisDate != "" {
		if regex, err := regexp.Compile("(?i)" + diagnosisDate); err == nil {
			filter.DiagnosisDate = regex
		}
	}

	return &filter
}

func isUsersQueryValid(filter *usersProfileFilter) bool {
	if filter == nil {
		return true
	}
	if len(filter.TrustorPermissions) > 0 && !arePermissionsValid(filter.TrustorPermissions) {
		return false
	}
	if len(filter.TrusteePermissions) > 0 && !arePermissionsValid(filter.TrusteePermissions) {
		return false
	}
	return true
}
func userMatchesQueryOnPermissions(trustPerms permission.TrustPermissions, filter *usersProfileFilter) bool {
	if filter == nil {
		return true
	}
	if len(filter.TrustorPermissions) > 0 && (trustPerms.TrustorPermissions == nil || !arePermissionsSatisfied(filter.TrustorPermissions, *trustPerms.TrustorPermissions)) {
		return false
	}
	if len(filter.TrusteePermissions) > 0 && (trustPerms.TrusteePermissions == nil || !arePermissionsSatisfied(filter.TrusteePermissions, *trustPerms.TrusteePermissions)) {
		return false
	}

	return true
}
func userMatchesQueryOnUser(user *userLib.User, filter *usersProfileFilter) bool {
	if filter == nil {
		return true
	}
	if filter.Email != nil && filter.Email.FindStringIndex(user.Email()) == nil {
		return false
	}
	if filter.EmailVerified != nil && (user.EmailVerified == nil || *filter.EmailVerified != *user.EmailVerified) {
		return false
	}
	if filter.TermsAccepted != nil && (user.TermsAccepted == nil || filter.TermsAccepted.FindStringIndex(*user.TermsAccepted) == nil) {
		return false
	}
	return true
}

func userMatchesQueryOnProfile(user *userLib.User, filter *usersProfileFilter) bool {
	if filter == nil {
		return true
	}
	profile := user.Profile
	if filter.Name != nil && (profile == nil || filter.Name.FindStringIndex(profile.FullName) == nil) {
		return false
	}
	if filter.Birthday != nil && (profile == nil || filter.Birthday.FindStringIndex(string(profile.Birthday)) == nil) {
		return false
	}
	if filter.DiagnosisDate != nil && (profile == nil || filter.DiagnosisDate.FindStringIndex(string(profile.DiagnosisDate)) == nil) {
		return false
	}
	return true
}

func userMatchingQuery(user *userLib.User, filter *usersProfileFilter) *userLib.User {
	if filter == nil {
		return user
	}
	trustPerms := permission.TrustPermissions{
		TrustorPermissions: user.TrustorPermissions,
		TrusteePermissions: user.TrusteePermissions,
	}
	if !userMatchesQueryOnPermissions(trustPerms, filter) ||
		!userMatchesQueryOnUser(user, filter) ||
		!userMatchesQueryOnProfile(user, filter) {
		return nil
	}
	return user
}
