package user

import (
	"slices"
)

type UserProfile struct {
	FullName string          `json:"fullName"`
	Patient  *PatientProfile `json:"patient,omitempty"`
	Clinic   *ClinicProfile  `json:"clinic,omitempty"`
}

type PatientProfile struct {
	Birthday       string   `json:"birthday"`
	DiagnosisDate  string   `json:"diagnosisDate"`
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
		addAttribute(attributes, "profile.fullName", u.FullName)
	}
	if u.Patient != nil {
		patient := u.Patient
		addAttribute(attributes, "profile.patient.birthday", patient.Birthday)
		addAttribute(attributes, "profile.patient.diagnosisDate", patient.DiagnosisDate)
		addAttribute(attributes, "profile.patient.diagnosisType", patient.DiagnosisType)
		addAttributes(attributes, "profile.patient.targetDevices", patient.TargetDevices...)
		addAttribute(attributes, "profile.patient.targetTimezone", patient.TargetTimezone)
		addAttribute(attributes, "profile.patient.about", patient.About)
	}

	if u.Clinic != nil {
		clinic := u.Clinic
		addAttribute(attributes, "profile.clinic.name", clinic.Name)
		addAttributes(attributes, "profile.clinic.role", clinic.Role...)
		addAttribute(attributes, "profile.clinic.telephone", clinic.Telephone)
	}

	return attributes
}

func profileFromAttributes(attributes map[string][]string) (profile *UserProfile, ok bool) {
	u := &UserProfile{}
	u.FullName = getAttribute(attributes, "profile.fullName")

	if containsAnyAttributeKeys(attributes, "profile.patient.birthday", "profile.patient.diagnosisDate", "profile.patient.diagnosisType", "profile.patient.targetDevices", "profile.patient.targetTimezone", "profile.patient.about") {
		patient := &PatientProfile{}
		patient.Birthday = getAttribute(attributes, "profile.patient.birthday")
		patient.DiagnosisDate = getAttribute(attributes, "profile.patient.diagnosisDate")
		patient.DiagnosisType = getAttribute(attributes, "profile.patient.diagnosisType")
		patient.TargetDevices = getAttributes(attributes, "profile.patient.targetDevices")
		patient.TargetTimezone = getAttribute(attributes, "profile.patient.targetTimezone")
		patient.About = getAttribute(attributes, "profile.patient.about")
		u.Patient = patient
	}

	if containsAnyAttributeKeys(attributes, "profile.clinic.name", "profile.clinic.role", "profile.clinic.telephone") {
		clinic := &ClinicProfile{}
		clinic.Name = getAttribute(attributes, "profile.clinic.name")
		clinic.Role = getAttributes(attributes, "profile.clinic.role")
		clinic.Telephone = getAttribute(attributes, "profile.clinic.telephone")
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
