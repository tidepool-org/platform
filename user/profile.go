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
	return &LegacyUserProfile{
		FullName: up.FullName,
		Patient: &PatientProfile{
			Birthday:       up.Birthday,
			DiagnosisDate:  up.DiagnosisDate,
			TargetDevices:  up.TargetDevices,
			TargetTimezone: up.TargetTimezone,
			About:          up.About,
		},
	}
}

func (p *LegacyUserProfile) ToUserProfile() *UserProfile {
	return &UserProfile{
		FullName:       p.FullName,
		Birthday:       p.Patient.Birthday,
		DiagnosisDate:  p.Patient.DiagnosisDate,
		TargetDevices:  p.Patient.TargetDevices,
		TargetTimezone: p.Patient.TargetTimezone,
		About:          p.Patient.About,
	}
}

type LegacyUserProfile struct {
	FullName string          `json:"fullName"`
	Patient  *PatientProfile `json:"patient,omitempty"`
	Clinic   *ClinicProfile  `json:"clinic,omitempty"`
}

type PatientProfile struct {
	Birthday       Date     `json:"birthday"`
	DiagnosisDate  Date     `json:"diagnosisDate"`
	DiagnosisType  string   `json:"diagnosisType"`
	TargetDevices  []string `json:"targetDevices"`
	TargetTimezone string   `json:"targetTimezone"`
	About          string   `json:"about"`
}

type ClinicProfile struct {
	Name      string   `json:"diagnosisDate"`
	Role      []string `json:"role"`
	Telephone string   `json:"telephone"`
}

func (up *UserProfile) ToAttributes() map[string][]string {
	attributes := map[string][]string{}

	if up.FullName != "" {
		addAttribute(attributes, "profile_full_name", up.FullName)
	}
	if up.Custodian != nil && up.Custodian.FullName != "" {
		addAttribute(attributes, "profile_custodian_full_name", up.Custodian.FullName)
	}
	if string(up.Birthday) != "" {
		addAttribute(attributes, "profile_birthday", string(up.Birthday))
	}
	if string(up.DiagnosisDate) != "" {
		addAttribute(attributes, "profile_diagnosis_date", string(up.DiagnosisDate))
	}
	if up.DiagnosisType != "" {
		addAttribute(attributes, "profile_diagnosis_type", up.DiagnosisType)
	}
	addAttributes(attributes, "profile_target_devices", up.TargetDevices...)
	if up.TargetTimezone != "" {
		addAttribute(attributes, "profile_target_timezone", up.TargetTimezone)
	}
	if up.About != "" {
		addAttribute(attributes, "profile_about", up.About)
	}

	return attributes
}

func ProfileFromAttributes(attributes map[string][]string) (profile *UserProfile, ok bool) {
	up := &UserProfile{}
	if val := getAttribute(attributes, "profile_full_name"); val != "" {
		up.FullName = val
		ok = true
	}
	if val := getAttribute(attributes, "profile_custodian_full_name"); val != "" {
		up.Custodian = &Custodian{
			FullName: val,
		}
		ok = true
	}
	if val := getAttribute(attributes, "profile_birthday"); val != "" {
		up.Birthday = Date(val)
		ok = true
	}
	if val := getAttribute(attributes, "profile_diagnosis_date"); val != "" {
		up.DiagnosisDate = Date(val)
		ok = true
	}
	if val := getAttribute(attributes, "profile_diagnosis_type"); val != "" {
		up.DiagnosisType = val
		ok = true
	}
	if vals := getAttributes(attributes, "profile_target_devices"); len(vals) > 0 {
		up.TargetDevices = vals
		ok = true
	}
	if val := getAttribute(attributes, "profile_target_timezone"); val != "" {
		up.TargetTimezone = val
		ok = true
	}
	if val := getAttribute(attributes, "profile_about"); val != "" {
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
