package test

import (
	"math/rand"

	"github.com/onsi/gomega"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewClockDriftOffset() int {
	return -86400000 + rand.Intn(86400000+86400000)
}

func NewConversionOffset() int {
	return -9999999999 + rand.Intn(9999999999+9999999999)
}

func NewNote(minimumLength int, maximumLength int) string {
	return test.RandomStringFromRange(minimumLength, maximumLength)
}

func NewNotes(minimumLength int, maximumLength int) []string {
	result := make([]string, minimumLength+rand.Intn(maximumLength-minimumLength+1))
	for index := range result {
		result[index] = NewNote(1, 1000)
	}
	return result
}

func NewTag(minimumLength int, maximumLength int) string {
	return test.RandomStringFromRange(minimumLength, maximumLength)
}

func NewTags(minimumLength int, maximumLength int) []string {
	var result []string
	exists := map[string]bool{}
	for length := minimumLength + rand.Intn(maximumLength-minimumLength+1); len(result) < length; {
		if tag := NewTag(1, 100); !exists[tag] {
			result = append(result, tag)
			exists[tag] = true
		}
	}
	return result
}

func NewTimeZoneOffset() int {
	return -4440 + rand.Intn(4440+6960)
}

func NewType() string {
	return test.RandomStringFromRangeAndCharset(1, 32, test.CharsetAlphaNumeric+"/")
}

func NewVersionInternal() int {
	return rand.Intn(10)
}

func ValidateWithOrigin(validatable structure.Validatable, origin structure.Origin, expectedErrors ...error) {
	validator := structureValidator.New(logTest.NewLogger())
	gomega.Expect(validator).ToNot(gomega.BeNil())
	validatable.Validate(validator.WithOrigin(origin))
	errorsTest.ExpectEqual(validator.Error(), expectedErrors...)
}

func ValidateWithExpectedOrigins(validatable structure.Validatable, expectedOrigins []structure.Origin, expectedErrors ...error) {
	for _, origin := range structure.Origins() {
		var expected bool
		for _, expectedOrigin := range expectedOrigins {
			if expected = (expectedOrigin == origin); expected {
				break
			}
		}
		if expected {
			ValidateWithOrigin(validatable, origin, expectedErrors...)
		} else {
			ValidateWithOrigin(validatable, origin)
		}
	}
}
