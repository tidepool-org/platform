package user

import (
	"cmp"
	"encoding/json"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
)

type migrationStatus int

var (
	nonLetters = regexp.MustCompile(`[^A-Za-z]`)
)

const (
	MigrationUnmigrated migrationStatus = iota
	MigrationCompleted
	MigrationInProgress
	MigrationError

	MaxProfileFieldLen = 255
)

func IsMigrationCompleted(status migrationStatus) bool {
	return status == MigrationCompleted
}

const (
	DiabetesTypeType1       = "type1"
	DiabetesTypeType2       = "type2"
	DiabetesTypeGestational = "gestational"
	DiabetesTypeLada        = "lada"
	DiabetesTypeOther       = "other"
	DiabetesTypePrediabetes = "prediabetes"
	DiabetesTypeMody        = "mody"
)

var (
	DiabetesTypes = []string{
		DiabetesTypeType1,
		DiabetesTypeType2,
		DiabetesTypeGestational,
		DiabetesTypeLada,
		DiabetesTypeOther,
		DiabetesTypePrediabetes,
		DiabetesTypeMody,
	}
)

// Date is a string of type YYYY-mm-dd, the reason this isn't just a type definition
// of a time.Time is to ignore timezones when marshaling.
type Date string

// UserProfile represents the user modifiable attributes of a user. It is named
// somewhat redundantly as UserProfile instead of Profile because there already
// exists a type Profile in this package.
type UserProfile struct {
	FullName       string   `json:"fullName,omitempty"` // Name of the patient, fake child, or clinician
	Birthday       Date     `json:"birthday,omitempty"`
	DiagnosisDate  Date     `json:"diagnosisDate,omitempty"`
	DiagnosisType  string   `json:"diagnosisType,omitempty"`
	TargetDevices  []string `json:"targetDevices,omitempty"`
	TargetTimezone string   `json:"targetTimezone,omitempty"`
	About          string   `json:"about,omitempty"`
	MRN            string   `json:"mrn,omitempty"`
	BiologicalSex  string   `json:"biologicalSex,omitempty"`

	Custodian *Custodian `json:"custodian,omitempty"`
	// The PRESENCE of a clinic object in a profile is used by blip to determine which page to show so this needs to be returned in the response.
	// There are clinicians/legacy clinics with completely empty values within the clinic object but are still clinicians/clinics.
	Clinic *ClinicProfile `json:"clinic,omitempty"`
	Email  string         `json:"-"` // This is used when returning profiles in the legacy format. It is not stored in the profile, but is populated from the keycloak username and not returned in the new profiles route.
}

type ClinicProfile struct {
	Name      *string `json:"name,omitempty"` // Refers to the name of the clinic, not clinician
	Role      *string `json:"role,omitempty"`
	Telephone *string `json:"telephone,omitempty"`
	NPI       *string `json:"npi,omitempty"`
}

type Custodian struct {
	FullName string `json:"fullName"`
}

func HasPatientRole(roles []string) bool {
	return slices.Contains(roles, RolePatient)
}

func HasClinicOrClinicianRole(roles []string) bool {
	return slices.Contains(roles, RoleClinician) || slices.Contains(roles, RoleClinic)
}

// IsPatientProfile returns true if the profile is associated with a patient - note that this is not mutually exclusive w/ a clinician, as some users have both
func (up *UserProfile) IsPatientProfile(roles []string) bool {
	return HasPatientRole(roles) || up.hasPatientFields() || !HasClinicOrClinicianRole(roles)
}

func (up *UserProfile) hasPatientFields() bool {
	return up.DiagnosisDate != "" || up.DiagnosisType != "" || len(up.TargetDevices) > 0 || up.MRN != "" || up.About != "" || up.BiologicalSex != "" || up.Birthday != "" || up.Custodian != nil
}

// IsClinicianProfile returns true if the profile is associated with a clinician - note that this is not mutually exclusive w/ a patient, as some users have both
func (up *UserProfile) IsClinicianProfile(roles []string) bool {
	return up.Clinic != nil || HasClinicOrClinicianRole(roles)
}

func (up *UserProfile) ToLegacyProfile(roles []string) *LegacyUserProfile {
	legacyProfile := &LegacyUserProfile{
		FullName:        up.FullName,
		MigrationStatus: MigrationCompleted, // If we have a non legacy UserProfile, then that means the legacy version has been migrated from seagull (or it never existed which is equivalent for the new user profile purposes)
	}

	if up.IsClinicianProfile(roles) {
		legacyProfile.Clinic = up.Clinic
		// Frontend uses the PRESENCE of a clinic object in some of its logic to
		// determine what pages to show so if this is a clinician so if there are
		// no actual clinician fields in the profile (No clinician role (such as
		// clinic_manager, endocrinologist, etc), npi, telephone etc), make an
		// empty, non-nil object.
		if legacyProfile.Clinic == nil {
			legacyProfile.Clinic = &ClinicProfile{}
		}
	}

	if up.IsPatientProfile(roles) {
		legacyProfile.Patient = &LegacyPatientProfile{
			Birthday:       up.Birthday,
			DiagnosisDate:  up.DiagnosisDate,
			DiagnosisType:  up.DiagnosisType,
			TargetDevices:  up.TargetDevices,
			TargetTimezone: up.TargetTimezone,
			About:          up.About,
			MRN:            up.MRN,
			BiologicalSex:  up.BiologicalSex,
		}
		if up.Email != "" && !IsUnclaimedCustodialEmail(up.Email) {
			legacyProfile.Patient.Email = up.Email
			legacyProfile.Patient.Emails = []string{up.Email}
			legacyProfile.Email = up.Email
			legacyProfile.Emails = []string{up.Email}
		}
	}
	// only custodiaL fake child accounts have Patient.FullName set
	if up.Custodian != nil {
		legacyProfile.Patient.IsOtherPerson = true
		// Handle case where Custodian user (contains fake child) and one of the FullName's is empty.
		legacyProfile.FullName = cmp.Or(up.Custodian.FullName, up.FullName)
		legacyProfile.Patient.FullName = pointer.FromString(cmp.Or(up.FullName, up.Custodian.FullName))
	}
	return legacyProfile
}

// ClearPatientInfo makes a copy of up, clearing out certain patient information - this is called usually due to lack of permissions to the patient information
func (up *UserProfile) ClearPatientInfo() *UserProfile {
	// explicitly specifying the type to make sure it's a value instead of pointer
	var newProfile UserProfile = *up
	newProfile.Birthday = ""
	newProfile.DiagnosisDate = ""
	newProfile.TargetDevices = nil
	newProfile.TargetTimezone = ""
	newProfile.About = ""
	newProfile.MRN = ""
	newProfile.BiologicalSex = ""
	newProfile.Custodian = nil
	newProfile.Clinic = nil
	return &newProfile
}

func (p *LegacyUserProfile) ToUserProfile() *UserProfile {
	up := &UserProfile{
		FullName: p.FullName,
		Clinic:   p.Clinic,
	}

	if p.Patient != nil {
		// The new profiles FullName refer to the true "owner" of the profile - which
		// may be the "fake child" so set it to the FullName within the Patient Object if it exists.
		up.FullName = cmp.Or(pointer.ToString(p.Patient.FullName), p.FullName)
		// Only users with isOtherPerson set has a patient.fullName field set so these users
		// also have a custodian
		if p.Patient.IsOtherPerson {
			// Handle the few cases where one of either the fake child fullName or the profile fullName is empty (neither are both empty)
			// The custodian's name would be the the profile.fullName field in the legacy
			// format. But there are few cases where it's empty so set it to profile.patient.fullName if it exists
			up.Custodian = &Custodian{
				FullName: cmp.Or(p.FullName, pointer.ToString(p.Patient.FullName)),
			}
		}
		up.Birthday = p.Patient.Birthday
		up.DiagnosisDate = p.Patient.DiagnosisDate
		up.DiagnosisType = p.Patient.DiagnosisType
		up.TargetDevices = p.Patient.TargetDevices
		up.TargetTimezone = p.Patient.TargetTimezone
		up.About = p.Patient.About
		up.MRN = p.Patient.MRN
		up.BiologicalSex = p.Patient.BiologicalSex
	}
	if p.Clinic != nil {
		up.Clinic = &ClinicProfile{
			Name:      pointer.CloneString(p.Clinic.Name),
			Role:      pointer.CloneString(p.Clinic.Role),
			Telephone: pointer.CloneString(p.Clinic.Telephone),
			NPI:       pointer.CloneString(p.Clinic.NPI),
		}
	}
	return up
}

// LegacyUserProfile represents the old seagull format for a profile.
type LegacyUserProfile struct {
	FullName        string                `json:"fullName,omitempty"` // string pointer because some old profiles have empty string as full name
	Patient         *LegacyPatientProfile `json:"patient,omitempty"`
	Clinic          *ClinicProfile        `json:"clinic,omitempty"`
	MigrationStatus migrationStatus       `json:"-"`
	// The Email and Emails fields are legacy properties that will be populated from the keycloak user if the profile is finished migrating, otherwise from the seagull collection
	Email  string   `json:"email,omitempty"`
	Emails []string `json:"emails,omitempty"`
}

type LegacyPatientProfile struct {
	FullName       *string  `json:"fullName,omitempty"` // This is only non-empty if the user is also a fake child (has the patient.isOtherPerson field set - there are cases where it is an empty string but the field exists)
	Birthday       Date     `json:"birthday,omitempty"`
	DiagnosisDate  Date     `json:"diagnosisDate,omitempty"`
	DiagnosisType  string   `json:"diagnosisType,omitempty"`
	TargetDevices  []string `json:"targetDevices,omitempty"`
	TargetTimezone string   `json:"targetTimezone,omitempty"`
	About          string   `json:"about,omitempty"`
	IsOtherPerson  jsonBool `json:"isOtherPerson,omitempty"`
	MRN            string   `json:"mrn,omitempty"`
	BiologicalSex  string   `json:"biologicalSex,omitempty"`
	// The Email and Emails fields are legacy properties that will be populated from the keycloak user if the profile is finished migrating, otherwise from the seagull collection
	Email  string   `json:"email,omitempty"`
	Emails []string `json:"emails,omitempty"`
}

func (l *LegacyPatientProfile) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	// Handle some old seagull fields that contained an empty string for the patient field, return an empty object in that case
	dataStr := string(data)
	if dataStr == `""` {
		return nil
	}

	// Create a new type definition w/ same underlying type as
	// LegacyPatientProfile so we can use the "default" UnmarshalJSON of
	// LegacyPatientProfile as if it didn't implement json.Unmarshaler (to
	// prevent an infinite loop)
	type tempType LegacyPatientProfile
	return json.Unmarshal(data, (*tempType)(l))
}

// jsonBool is a bool type that can be marshaled from string fields - this is only in support of legacy seagull profiles.
// Once all seagull profiles have been migrated over, LegacyProfile along w/ jsonBool will be removed
type jsonBool bool

func (b *jsonBool) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	dataStr := string(data)
	boolStr := strings.ToLower(nonLetters.ReplaceAllString(dataStr, ""))
	if boolStr == "true" {
		*b = true
	} else {
		*b = false
	}
	return nil
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
	if up.MRN != "" {
		addAttribute(attributes, "mrn", up.MRN)
	}
	if up.BiologicalSex != "" {
		addAttribute(attributes, "biological_sex", up.BiologicalSex)
	}

	if up.Clinic != nil {
		if val := pointer.ToString(up.Clinic.Name); val != "" {
			addAttribute(attributes, "clinic_name", val)
		}
		if val := pointer.ToString(up.Clinic.Role); val != "" {
			addAttribute(attributes, "clinic_role", val)
		}
		if val := pointer.ToString(up.Clinic.Telephone); val != "" {
			addAttribute(attributes, "clinic_telephone", val)
		}
		if val := pointer.ToString(up.Clinic.NPI); val != "" {
			addAttribute(attributes, "clinic_npi", val)
		}
	}

	return attributes
}

func ProfileFromAttributes(username string, attributes map[string][]string, roles []string) (profile *UserProfile, ok bool) {
	up := &UserProfile{
		Email: username,
	}
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
	if val := getAttribute(attributes, "mrn"); val != "" {
		up.MRN = val
		ok = true
	}
	if val := getAttribute(attributes, "biological_sex"); val != "" {
		up.BiologicalSex = val
		ok = true
	}

	var clinicProfile ClinicProfile
	// A clinic may have all empty fields but still needs a clinic object
	// returned so check both the presence of the clinic / clinician role and
	// individual clinic properties - It may be enough to just check the roles
	hasClinicProfile := HasClinicOrClinicianRole(roles)
	if val := getAttribute(attributes, "clinic_name"); val != "" {
		clinicProfile.Name = pointer.FromString(val)
		hasClinicProfile = true
	}
	if val := getAttribute(attributes, "clinic_role"); val != "" {
		clinicProfile.Role = pointer.FromString(val)
		hasClinicProfile = true
	}
	if val := getAttribute(attributes, "clinic_telephone"); val != "" {
		clinicProfile.Telephone = pointer.FromString(val)
		hasClinicProfile = true
	}
	if val := getAttribute(attributes, "clinic_npi"); val != "" {
		clinicProfile.NPI = pointer.FromString(val)
		hasClinicProfile = true
	}
	if hasClinicProfile {
		up.Clinic = &clinicProfile
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

func (d *Date) Validate(v structure.Validator) {
	if d == nil || *d == "" {
		return
	}
	str := string(*d)
	v.String("date", &str).AsTime(time.DateOnly)
}

func (d *Date) Normalize(normalizer structure.Normalizer) {
	if d == nil || *d == "" {
		return
	}
	*d = Date(strings.TrimSpace(string(*d)))
}

func (up *UserProfile) Validate(v structure.Validator) {
	v.String("fullName", &up.FullName).LengthLessThanOrEqualTo(MaxProfileFieldLen)
	v.String("diagnosisType", &up.DiagnosisType).LengthLessThanOrEqualTo(MaxProfileFieldLen)
	v.String("targetTimezone", &up.TargetTimezone).LengthLessThanOrEqualTo(MaxProfileFieldLen)
	v.String("about", &up.About).LengthLessThanOrEqualTo(MaxProfileFieldLen)
	v.String("mrn", &up.MRN).LengthLessThanOrEqualTo(MaxProfileFieldLen)
	v.String("biologicalSex", &up.BiologicalSex).LengthLessThanOrEqualTo(MaxProfileFieldLen)

	up.Birthday.Validate(v.WithReference("birthday"))
	up.DiagnosisDate.Validate(v.WithReference("diagnosisDate"))
	if up.DiagnosisType != "" {
		v.String("diagnosisType", &up.DiagnosisType).OneOf(DiabetesTypes...)
	}
}

func (up *UserProfile) Normalize(normalizer structure.Normalizer) {
	up.FullName = strings.TrimSpace(up.FullName)
	up.DiagnosisType = strings.TrimSpace(up.DiagnosisType)
	up.TargetTimezone = strings.TrimSpace(up.TargetTimezone)
	up.About = strings.TrimSpace(up.About)
	up.MRN = strings.TrimSpace(up.MRN)
	up.BiologicalSex = strings.TrimSpace(up.BiologicalSex)

	up.Birthday.Normalize(normalizer.WithReference("birthday"))
	up.DiagnosisDate.Normalize(normalizer.WithReference("diagnosisDate"))
	if up.Clinic != nil {
		up.Clinic.Normalize(normalizer.WithReference("clinic"))
	}
}

func (p *ClinicProfile) Normalize(normalizer structure.Normalizer) {
	if p.Name != nil {
		*p.Name = strings.TrimSpace(*p.Name)
	}
	if p.Role != nil {
		*p.Role = strings.TrimSpace(*p.Role)
	}
	if p.Telephone != nil {
		*p.Telephone = strings.TrimSpace(*p.Telephone)
	}
	if p.NPI != nil {
		*p.NPI = strings.TrimSpace(*p.NPI)
	}
}

func (up *LegacyUserProfile) Validate(v structure.Validator) {
	if up.Patient != nil {
		up.Patient.Validate(v.WithReference("patient"))
	}
	v.String("fullName", &up.FullName).LengthLessThanOrEqualTo(MaxProfileFieldLen)
}

func (up *LegacyUserProfile) Normalize(normalizer structure.Normalizer) {
	up.FullName = strings.TrimSpace(up.FullName)
	if up.Patient != nil {
		up.Patient.Normalize(normalizer.WithReference("patient"))
	}
	if up.Clinic != nil {
		up.Clinic.Normalize(normalizer.WithReference("clinic"))
	}
	// Email and Emails are read-only so they are ignored in normalizing / validation
}

func (pp *LegacyPatientProfile) Validate(v structure.Validator) {
	pp.Birthday.Validate(v.WithReference("birthday"))
	pp.DiagnosisDate.Validate(v.WithReference("diagnosisDate"))

	v.String("fullName", pp.FullName).LengthLessThanOrEqualTo(MaxProfileFieldLen)
	v.String("targetTimezone", &pp.TargetTimezone).LengthLessThanOrEqualTo(MaxProfileFieldLen)
	v.String("about", &pp.About).LengthLessThanOrEqualTo(MaxProfileFieldLen)
	v.String("mrn", &pp.MRN).LengthLessThanOrEqualTo(MaxProfileFieldLen)

	if pp.DiagnosisType != "" {
		v.String("diagnosisType", &pp.DiagnosisType).OneOf(DiabetesTypes...)
	}
}

func (pp *LegacyPatientProfile) Normalize(normalizer structure.Normalizer) {
	pp.Birthday.Normalize(normalizer.WithReference("birthday"))
	pp.DiagnosisDate.Normalize(normalizer.WithReference("diagnosisDate"))

	if pp.FullName != nil {
		pp.FullName = pointer.FromString(strings.TrimSpace(pointer.ToString(pp.FullName)))
	}
	pp.DiagnosisType = strings.TrimSpace(pp.DiagnosisType)
	if pp.TargetTimezone != "" {
		pp.TargetTimezone = strings.TrimSpace(pp.TargetTimezone)
	}
	pp.About = strings.TrimSpace(pp.About)
	pp.MRN = strings.TrimSpace(pp.MRN)
	pp.BiologicalSex = strings.TrimSpace(pp.BiologicalSex)
}
