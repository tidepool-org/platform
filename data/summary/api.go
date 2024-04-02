package summary

import (
	"time"

	"github.com/tidepool-org/platform/structure"
)

type RealtimePatientsResponse struct {
	Config  RealtimePatientConfigResponse `json:"config"`
	Results []RealtimePatientResponse     `json:"results"`
}

type RealtimePatientConfigResponse struct {
	Code      string    `json:"code"`
	ClinicId  string    `json:"clinicId"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

type RealtimePatientResponse struct {
	Id                string    `json:"id"`
	FullName          string    `json:"fullName"`
	BirthDate         time.Time `json:"birthDate"`
	MRN               *string   `json:"mrn"`
	RealtimeDays      int       `json:"realtimeDays"`
	HasSufficientData bool      `json:"hasSufficientData"`
}

type RealtimePatientsFilter struct {
	StartTime *time.Time
	EndTime   *time.Time
}

func NewRealtimePatientsFilter() *RealtimePatientsFilter {
	return &RealtimePatientsFilter{}
}

func (d *RealtimePatientsFilter) Parse(parser structure.ObjectParser) {
	d.StartTime = parser.Time("startDate", time.RFC3339)
	d.EndTime = parser.Time("endDate", time.RFC3339)
}

func (d *RealtimePatientsFilter) Validate(validator structure.Validator) {
	validator.Time("startDate", d.StartTime).NotZero()
	validator.Time("endDate", d.EndTime).NotZero()
}
