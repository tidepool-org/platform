package test

import (
	"math/rand"

	"github.com/onsi/gomega"

	testErrors "github.com/tidepool-org/platform/errors/test"
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

func NewTimezoneOffset() int {
	return -4440 + rand.Intn(4440+6960)
}

func NewType() string {
	return test.NewVariableString(1, 32, test.CharsetAlphaNumeric+"/")
}

func NewVersion() int {
	return rand.Intn(10)
}

func ValidateWithOrigin(validatable structure.Validatable, origin structure.Origin, expectedErrors ...error) {
	validator := structureValidator.New()
	gomega.Expect(validator).ToNot(gomega.BeNil())
	validatable.Validate(validator.WithOrigin(origin))
	testErrors.ExpectEqual(validator.Error(), expectedErrors...)
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
