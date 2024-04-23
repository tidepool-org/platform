package user

import (
	"slices"
	"time"

	"github.com/tidepool-org/platform/structure"
)

const (
	maxAboutLength = 256
	maxNameLength  = 256
)

// Date is a string of type YYYY-mm-dd, the reason this isn't just a type definition
// of a time.Time is to ignore timezones when marshaling.
type Date string

// UserProfile represents the user modifiable attributes of a user. It is named
// somewhat redundantly as UserProfile instead of Profile because there already
// exists a type Profile in this package.
type UserProfile struct {
	FullName       string     `json:"fullName"`
	Birthday       Date       `json:"birthday"`
	DiagnosisDate  Date       `json:"diagnosisDate"`
	DiagnosisType  string     `json:"diagnosisType"`
	TargetDevices  []string   `json:"targetDevices"`
	TargetTimezone string     `json:"targetTimezone"`
	About          string     `json:"about"`
	Custodian      *Custodian `json:"custodian,omitempty"`
}

type Custodian struct {
	FullName string `json:"fullName"`
}

func (up *UserProfile) ToLegacyProfile() *LegacyUserProfile {
	legacyProfile := &LegacyUserProfile{
		FullName: up.FullName,
		Patient: &PatientProfile{
			Birthday:       up.Birthday,
			DiagnosisDate:  up.DiagnosisDate,
			TargetDevices:  up.TargetDevices,
			TargetTimezone: up.TargetTimezone,
			About:          up.About,
		},
	}
	// only custodiaL fake child accounts have Patient.FullName set
	if up.Custodian != nil {
		legacyProfile.FullName = up.Custodian.FullName
		legacyProfile.Patient.FullName = up.FullName
		legacyProfile.Patient.IsOtherPerson = true
	}
	return legacyProfile
}

func (p *LegacyUserProfile) ToUserProfile() *UserProfile {
	up := &UserProfile{
		FullName: p.FullName,
	}
	if p.Patient != nil {
		up.FullName = p.Patient.FullName
		// Only users with isOtherPerson set has a patient.fullName field set so
		// they have a custodian.
		if p.Patient.FullName != "" || p.Patient.IsOtherPerson {
			up.Custodian = &Custodian{
				FullName: p.FullName,
			}
			if up.Custodian.FullName == "" {
				up.Custodian.FullName = p.FullName
			}
		}
		up.Birthday = p.Patient.Birthday
		up.DiagnosisDate = p.Patient.DiagnosisDate
		up.TargetDevices = p.Patient.TargetDevices
		up.TargetTimezone = p.Patient.TargetTimezone
		up.About = p.Patient.About
	}
	return up
}

// LegacyUserProfile represents the old seagull format for a profile.
type LegacyUserProfile struct {
	FullName string          `json:"fullName"`
	Patient  *PatientProfile `json:"patient,omitempty"`
	Clinic   *ClinicProfile  `json:"clinic,omitempty"`
}

type PatientProfile struct {
	FullName       string   `json:"fullName,omitempty"` // This is only non-empty if the user is also a fake child (has the patient.isOtherPerson field set)
	Birthday       Date     `json:"birthday"`
	DiagnosisDate  Date     `json:"diagnosisDate"`
	DiagnosisType  string   `json:"diagnosisType"`
	TargetDevices  []string `json:"targetDevices"`
	TargetTimezone string   `json:"targetTimezone"`
	About          string   `json:"about"`
	IsOtherPerson  bool     `json:"isOtherPerson,omitempty"`
}

type ClinicProfile struct {
	Name      string   `json:"diagnosisDate"`
	Role      []string `json:"role"`
	Telephone string   `json:"telephone"`
}

func (up *UserProfile) ToAttributes() map[string][]string {
	attributes := map[string][]string{}

	if up.FullName != "" {
		addAttribute(attributes, "full_name", up.FullName)
	}
	if up.Custodian != nil && up.Custodian.FullName != "" {
		addAttribute(attributes, "custodian_full_name", up.Custodian.FullName)
		// The "has_custodian" attribute is only added so that filtering on users is simpler via the keycloak API - because
		// there is a way to filter by custom attribute values but not by the presence of one.
		addAttribute(attributes, "has_custodian", "true")
	}
	if string(up.Birthday) != "" {
		addAttribute(attributes, "birthday", string(up.Birthday))
	}
	if string(up.DiagnosisDate) != "" {
		addAttribute(attributes, "diagnosis_date", string(up.DiagnosisDate))
	}
	if up.DiagnosisType != "" {
		addAttribute(attributes, "diagnosis_type", up.DiagnosisType)
	}
	addAttributes(attributes, "target_devices", up.TargetDevices...)
	if up.TargetTimezone != "" {
		addAttribute(attributes, "target_timezone", up.TargetTimezone)
	}
	if up.About != "" {
		addAttribute(attributes, "about", up.About)
	}

	return attributes
}

func ProfileFromAttributes(attributes map[string][]string) (profile *UserProfile, ok bool) {
	up := &UserProfile{}
	if val := getAttribute(attributes, "full_name"); val != "" {
		up.FullName = val
		ok = true
	}
	if val := getAttribute(attributes, "custodian_full_name"); val != "" {
		up.Custodian = &Custodian{
			FullName: val,
		}
		ok = true
	}
	if val := getAttribute(attributes, "birthday"); val != "" {
		up.Birthday = Date(val)
		ok = true
	}
	if val := getAttribute(attributes, "diagnosis_date"); val != "" {
		up.DiagnosisDate = Date(val)
		ok = true
	}
	if val := getAttribute(attributes, "diagnosis_type"); val != "" {
		up.DiagnosisType = val
		ok = true
	}
	if vals := getAttributes(attributes, "target_devices"); len(vals) > 0 {
		up.TargetDevices = vals
		ok = true
	}
	if val := getAttribute(attributes, "target_timezone"); val != "" {
		up.TargetTimezone = val
		ok = true
	}
	if val := getAttribute(attributes, "about"); val != "" {
		up.About = val
		ok = true
	}

	return up, ok
}

func addAttribute(attributes map[string][]string, attribute, value string) (ok bool) {
	if !containsAttribute(attributes, attribute, value) {
		attributes[attribute] = append(attributes[attribute], value)
		return true
	}
	return false
}

func getAttribute(attributes map[string][]string, attribute string) string {
	if len(attributes[attribute]) > 0 {
		return attributes[attribute][0]
	}
	return ""
}

func getAttributes(attributes map[string][]string, attribute string) []string {
	return attributes[attribute]
}

func addAttributes(attributes map[string][]string, attribute string, values ...string) (ok bool) {
	for _, value := range values {
		if addAttribute(attributes, attribute, value) {
			ok = true
		}
	}
	return true
}

func containsAttribute(attributes map[string][]string, attribute, value string) bool {
	for key, vals := range attributes {
		if key == attribute && slices.Contains(vals, value) {
			return true
		}
	}
	return false
}

func containsAnyAttributeKeys(attributes map[string][]string, keys ...string) bool {
	for key, vals := range attributes {
		if len(vals) > 0 && slices.Contains(keys, key) {
			return true
		}
	}
	return false
}

func (d Date) Validate(v structure.Validator) {
	if d == "" {
		return
	}
	str := string(d)
	v.String("date", &str).AsTime(time.DateOnly)
}

func (up *UserProfile) Validate(v structure.Validator) {
	up.Birthday.Validate(v.WithReference("birthday"))
	up.DiagnosisDate.Validate(v.WithReference("diagnosisDate"))
	v.String("fullName", &up.FullName).LengthLessThanOrEqualTo(maxNameLength)
}

func (p *ClinicProfile) Validate(v structure.Validator) {
	v.String("name", &p.Name).NotEmpty().LengthLessThanOrEqualTo(maxNameLength)
}
