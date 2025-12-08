package jotform

import (
	"time"

	"github.com/tidepool-org/platform/structure/validator"
)

type OuraEligibilitySurvey struct {
	DateOfBirth string
	Name        string
}

func (o *OuraEligibilitySurvey) Validate(v *validator.Validator) {
	eighteenYearsAgo := time.Now().AddDate(-18, 0, 0)
	v.String("dateOfBirth", &o.DateOfBirth).NotEmpty().AsTime(time.DateTime).Before(eighteenYearsAgo)
	v.String("name", &o.Name).NotEmpty()
}
