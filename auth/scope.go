package auth

import (
	"regexp"
	"slices"
	"strings"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

// See https://datatracker.ietf.org/doc/html/rfc6749#section-3.3

func ParseScope(scope string) ([]string, error) {
	var parsedScope []string
	for scopeToken := range strings.SplitSeq(scope, scopeTokenSeparator) {
		if scopeToken = strings.TrimSpace(scopeToken); scopeToken == "" {
			continue
		} else if !IsValidScopeToken(scopeToken) {
			return nil, errors.New("scope token is invalid")
		} else {
			parsedScope = append(parsedScope, scopeToken)
		}
	}
	slices.Sort(parsedScope)
	return slices.Compact(parsedScope), nil
}

func JoinScope(scope []string) string {
	return strings.Join(scope, scopeTokenSeparator)
}

func IsValidScopeToken(value string) bool {
	return ValidateScopeToken(value) == nil
}

func ScopeTokenValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateScopeToken(value))
}

func ValidateScopeToken(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if !scopeTokenExpression.MatchString(value) {
		return ErrorValueStringAsScopeTokenNotValid(value)
	}
	return nil
}

func ErrorValueStringAsScopeTokenNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as scope token", value)
}

const scopeTokenSeparator = " "

var scopeTokenExpression = regexp.MustCompile(`^[\x21\x23-\x5B\x5D-\x7E]+$`)
