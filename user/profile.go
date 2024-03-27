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

type UserProfile struct {
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

func (u *UserProfile) ToAttributes() map[string][]string {
	attributes := map[string][]string{}

	if u.FullName != "" {
		addAttribute(attributes, "profile_full_name", u.FullName)
	}
	if u.Patient != nil {
		patient := u.Patient
		addAttribute(attributes, "profile_patient_birthday", string(patient.Birthday))
		addAttribute(attributes, "profile_patient_diagnosis_date", string(patient.DiagnosisDate))
		addAttribute(attributes, "profile_patient_diagnosis_type", patient.DiagnosisType)
		addAttributes(attributes, "profile_patient_target_devices", patient.TargetDevices...)
		addAttribute(attributes, "profile_patient_target_timezone", patient.TargetTimezone)
		addAttribute(attributes, "profile_patient_about", patient.About)
	}

	if u.Clinic != nil {
		clinic := u.Clinic
		addAttribute(attributes, "profile_clinic_name", clinic.Name)
		addAttributes(attributes, "profile_clinic_role", clinic.Role...)
		addAttribute(attributes, "profile_clinic_telephone", clinic.Telephone)
	}

	return attributes
}

func profileFromAttributes(attributes map[string][]string) (profile *UserProfile, ok bool) {
	u := &UserProfile{}
	u.FullName = getAttribute(attributes, "profile_full_name")

	if containsAnyAttributeKeys(attributes, "profile_patient_birthday", "profile_patient_diagnosis_date", "profile_patient_diagnosis_type", "profile_patient_target_devices", "profile_patient_target_timezone", "profile_patient_about") {
		patient := &PatientProfile{}
		patient.Birthday = Date(getAttribute(attributes, "profile_patient_birthday"))
		patient.DiagnosisDate = Date(getAttribute(attributes, "profile_patient_diagnosis_date"))
		patient.DiagnosisType = getAttribute(attributes, "profile_patient_diagnosis_type")
		patient.TargetDevices = getAttributes(attributes, "profile_patient_target_devices")
		patient.TargetTimezone = getAttribute(attributes, "profile_patient_target_timezone")
		patient.About = getAttribute(attributes, "profile_patient_about")
		u.Patient = patient
	}

	if containsAnyAttributeKeys(attributes, "profile_clinic_name", "profile_clinic_role", "profile_clinic_telephone") {
		clinic := &ClinicProfile{}
		clinic.Name = getAttribute(attributes, "profile_clinic_name")
		clinic.Role = getAttributes(attributes, "profile_clinic_role")
		clinic.Telephone = getAttribute(attributes, "profile_clinic_telephone")
		u.Clinic = clinic
	}

	if u.Clinic == nil && u.Patient == nil && u.FullName == "" {
		return nil, false
	}
	return u, true
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

func (p *PatientProfile) Validate(v structure.Validator) {
	p.Birthday.Validate(v.WithReference("birthday"))
	p.DiagnosisDate.Validate(v.WithReference("diagnosisDate"))
	v.String("about", &p.About).LengthLessThanOrEqualTo(maxAboutLength)
}

func (p *UserProfile) Validate(v structure.Validator) {
	if p.Patient != nil {
		p.Patient.Validate(v.WithReference("patient"))
	}
	if p.Clinic != nil {
		p.Clinic.Validate(v.WithReference("clinic"))
	}
	v.String("fullName", &p.FullName).LengthLessThanOrEqualTo(maxNameLength)
}

func (p *ClinicProfile) Validate(v structure.Validator) {
	v.String("name", &p.Name).NotEmpty().LengthLessThanOrEqualTo(maxNameLength)
}
