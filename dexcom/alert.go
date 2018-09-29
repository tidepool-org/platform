package dexcom

import (
	"regexp"
	"sort"
	"strconv"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	AlertScheduleSettingsStartTimeDefault = "00:00"
	AlertScheduleSettingsEndTimeDefault   = "00:00"

	AlertScheduleSettingsDaySunday    = "sunday"
	AlertScheduleSettingsDayMonday    = "monday"
	AlertScheduleSettingsDayTuesday   = "tuesday"
	AlertScheduleSettingsDayWednesday = "wednesday"
	AlertScheduleSettingsDayThursday  = "thursday"
	AlertScheduleSettingsDayFriday    = "friday"
	AlertScheduleSettingsDaySaturday  = "saturday"

	AlertSettingAlertNameFall          = "fall"
	AlertSettingAlertNameHigh          = "high"
	AlertSettingAlertNameLow           = "low"
	AlertSettingAlertNameNoReadings    = "noReadings"
	AlertSettingAlertNameOutOfRange    = "outOfRange"
	AlertSettingAlertNameRise          = "rise"
	AlertSettingAlertNameUrgentLow     = "urgentLow"
	AlertSettingAlertNameUrgentLowSoon = "urgentLowSoon"

	AlertSettingUnitMinutes    = "minutes"
	AlertSettingUnitMgdL       = "mg/dL"
	AlertSettingUnitMgdLMinute = "mg/dL/min"
)

func AlertScheduleSettingsDays() []string {
	return []string{
		AlertScheduleSettingsDaySunday,
		AlertScheduleSettingsDayMonday,
		AlertScheduleSettingsDayTuesday,
		AlertScheduleSettingsDayWednesday,
		AlertScheduleSettingsDayThursday,
		AlertScheduleSettingsDayFriday,
		AlertScheduleSettingsDaySaturday,
	}
}

func AlertScheduleSettingsDayIndex(day string) int {
	switch day {
	case AlertScheduleSettingsDaySunday:
		return 1
	case AlertScheduleSettingsDayMonday:
		return 2
	case AlertScheduleSettingsDayTuesday:
		return 3
	case AlertScheduleSettingsDayWednesday:
		return 4
	case AlertScheduleSettingsDayThursday:
		return 5
	case AlertScheduleSettingsDayFriday:
		return 6
	case AlertScheduleSettingsDaySaturday:
		return 7
	default:
		return 0
	}
}

func AlertSettingAlertNames() []string {
	return []string{
		AlertSettingAlertNameFall,
		AlertSettingAlertNameHigh,
		AlertSettingAlertNameLow,
		AlertSettingAlertNameNoReadings,
		AlertSettingAlertNameOutOfRange,
		AlertSettingAlertNameRise,
		AlertSettingAlertNameUrgentLow,
		AlertSettingAlertNameUrgentLowSoon,
	}
}

func AlertSettingUnitFalls() []string {
	return []string{AlertSettingUnitMgdLMinute}
}

func AlertSettingValueFallMgdLMinutes() []float64 {
	return []float64{2, 3}
}

func AlertSettingSnoozeFalls() []int {
	return []int{0, 30}
}

func AlertSettingUnitHighs() []string {
	return []string{AlertSettingUnitMgdL}
}

func AlertSettingValueHighMgdLs() []float64 {
	return alertSettingValueHighMgdLs
}

var alertSettingValueHighMgdLs = generateFloatRange(120, 400, 10)

func AlertSettingSnoozeHighs() []int {
	return alertSettingSnoozeHighs
}

var alertSettingSnoozeHighs = append(append([]int{0}, generateIntegerRange(15, 240, 5)...), generateIntegerRange(255, 300, 15)...)

func AlertSettingUnitLows() []string {
	return []string{AlertSettingUnitMgdL}
}

func AlertSettingValueLowMgdLs() []float64 {
	return alertSettingValueLowMgdLs
}

var alertSettingValueLowMgdLs = generateFloatRange(60, 100, 5)

func AlertSettingSnoozeLows() []int {
	return alertSettingSnoozeLows
}

var alertSettingSnoozeLows = append(append([]int{0}, generateIntegerRange(15, 240, 5)...), generateIntegerRange(255, 300, 15)...)

func AlertSettingUnitNoReadings() []string {
	return []string{AlertSettingUnitMinutes}
}

func AlertSettingValueNoReadingsMinutes() []float64 {
	return []float64{0, 20}
}

func AlertSettingSnoozeNoReadings() []int {
	return []int{0, 20, 25, 30}
}

func AlertSettingUnitOutOfRanges() []string {
	return []string{AlertSettingUnitMinutes}
}

func AlertSettingValueOutOfRangeMinutes() []float64 {
	return alertSettingValueOutOfRangeMinutes
}

var alertSettingValueOutOfRangeMinutes = generateFloatRange(20, 240, 5)

func AlertSettingSnoozeOutOfRanges() []int {
	return []int{0, 20, 25, 30}
}

func AlertSettingUnitRises() []string {
	return []string{AlertSettingUnitMgdLMinute}
}

func AlertSettingValueRiseMgdLMinutes() []float64 {
	return []float64{2, 3}
}

func AlertSettingSnoozeRises() []int {
	return []int{0, 30}
}

func AlertSettingUnitUrgentLows() []string {
	return []string{AlertSettingUnitMgdL}
}

func AlertSettingValueUrgentLowMgdLs() []float64 {
	return []float64{55}
}

func AlertSettingSnoozeUrgentLows() []int {
	return []int{0, 30}
}

func AlertSettingUnitUrgentLowSoons() []string {
	return []string{AlertSettingUnitMgdL}
}

func AlertSettingValueUrgentLowSoonMgdLs() []float64 {
	return []float64{55}
}

func AlertSettingSnoozeUrgentLowSoons() []int {
	return []int{0, 30}
}

type AlertSchedules []*AlertSchedule

func ParseAlertSchedules(parser structure.ArrayParser) *AlertSchedules {
	if !parser.Exists() {
		return nil
	}
	datum := NewAlertSchedules()
	parser.Parse(datum)
	return datum
}

func NewAlertSchedules() *AlertSchedules {
	return &AlertSchedules{}
}

func (a *AlertSchedules) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*a = append(*a, ParseAlertSchedule(parser.WithReferenceObjectParser(reference)))
	}
}

func (a *AlertSchedules) Validate(validator structure.Validator) {
	if len(*a) == 0 {
		validator.ReportError(structureValidator.ErrorValueEmpty())
	}
	var hasDefault bool
	for index, alertSchedule := range *a {
		if alertScheduleValidator := validator.WithReference(strconv.Itoa(index)); alertSchedule != nil {
			alertSchedule.Validate(alertScheduleValidator)
			if alertSchedule.IsDefault() {
				if !hasDefault {
					hasDefault = true
				} else {
					alertScheduleValidator.ReportError(structureValidator.ErrorValueDuplicate())
				}
			}
		} else {
			alertScheduleValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (a *AlertSchedules) Normalize(normalizer structure.Normalizer) {
	sort.Sort(AlertSchedulesByAlertScheduleName(*a))
	for index, alertSchedule := range *a {
		alertSchedule.Normalize(normalizer.WithReference(strconv.Itoa(index)))
	}
}

func (a *AlertSchedules) Default() *AlertSchedule {
	for _, alertSchedule := range *a {
		if alertSchedule.IsDefault() {
			return alertSchedule
		}
	}
	return nil
}

type AlertSchedulesByAlertScheduleName AlertSchedules

func (a AlertSchedulesByAlertScheduleName) Len() int {
	return len(a)
}
func (a AlertSchedulesByAlertScheduleName) Swap(i int, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a AlertSchedulesByAlertScheduleName) Less(i int, j int) bool {
	if left := a[i]; left == nil {
		return true
	} else if right := a[j]; right == nil {
		return false
	} else if leftName := left.Name(); leftName == nil {
		return true
	} else if rightName := right.Name(); rightName == nil {
		return false
	} else {
		return *leftName < *rightName
	}
}

type AlertSchedule struct {
	AlertScheduleSettings *AlertScheduleSettings `json:"alertScheduleSettings,omitempty" yaml:"alertScheduleSettings,omitempty"`
	AlertSettings         *AlertSettings         `json:"alertSettings,omitempty" yaml:"alertSettings,omitempty"`
}

func ParseAlertSchedule(parser structure.ObjectParser) *AlertSchedule {
	if !parser.Exists() {
		return nil
	}
	datum := NewAlertSchedule()
	parser.Parse(datum)
	return datum
}

func NewAlertSchedule() *AlertSchedule {
	return &AlertSchedule{}
}

func (a *AlertSchedule) Parse(parser structure.ObjectParser) {
	a.AlertScheduleSettings = ParseAlertScheduleSettings(parser.WithReferenceObjectParser("alertScheduleSettings"))
	a.AlertSettings = ParseAlertSettings(parser.WithReferenceArrayParser("alertSettings"))
}

func (a *AlertSchedule) Validate(validator structure.Validator) {
	if alertScheduleSettingsValidator := validator.WithReference("alertScheduleSettings"); a.AlertScheduleSettings != nil {
		a.AlertScheduleSettings.Validate(alertScheduleSettingsValidator)
	} else {
		alertScheduleSettingsValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
	if alertSettingsValidator := validator.WithReference("alertSettings"); a.AlertSettings != nil {
		a.AlertSettings.Validate(alertSettingsValidator)
	} else {
		alertSettingsValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
}

func (a *AlertSchedule) Normalize(normalizer structure.Normalizer) {
	if a.AlertScheduleSettings != nil {
		a.AlertScheduleSettings.Normalize(normalizer.WithReference("alertScheduleSettings"))
	}
	if a.AlertSettings != nil {
		a.AlertSettings.Normalize(normalizer.WithReference("alertSettings"))
	}
}

func (a *AlertSchedule) IsDefault() bool {
	return a.AlertScheduleSettings != nil && a.AlertScheduleSettings.IsDefault()
}

func (a *AlertSchedule) Name() *string {
	if a.AlertScheduleSettings != nil {
		return a.AlertScheduleSettings.Name
	}
	return nil
}

type AlertScheduleSettings struct {
	Name       *string   `json:"alertScheduleName,omitempty" yaml:"alertScheduleName,omitempty"`
	Enabled    *bool     `json:"isEnabled,omitempty" yaml:"isEnabled,omitempty"`
	Default    *bool     `json:"isDefaultSchedule,omitempty" yaml:"isDefaultSchedule,omitempty"`
	StartTime  *string   `json:"startTime,omitempty" yaml:"startTime,omitempty"`
	EndTime    *string   `json:"endTime,omitempty" yaml:"endTime,omitempty"`
	DaysOfWeek *[]string `json:"daysOfWeek,omitempty" yaml:"daysOfWeek,omitempty"`
}

func ParseAlertScheduleSettings(parser structure.ObjectParser) *AlertScheduleSettings {
	if !parser.Exists() {
		return nil
	}
	datum := NewAlertScheduleSettings()
	parser.Parse(datum)
	return datum
}

func NewAlertScheduleSettings() *AlertScheduleSettings {
	return &AlertScheduleSettings{}
}

func (a *AlertScheduleSettings) Parse(parser structure.ObjectParser) {
	a.Name = parser.String("alertScheduleName")
	a.Enabled = parser.Bool("isEnabled")
	a.Default = parser.Bool("isDefaultSchedule")
	a.StartTime = parser.String("startTime")
	a.EndTime = parser.String("endTime")
	a.DaysOfWeek = parser.StringArray("daysOfWeek")
}

func (a *AlertScheduleSettings) Validate(validator structure.Validator) {
	validator = validator.WithMeta(a)
	validator.Bool("isDefaultSchedule", a.Default).Exists()
	if a.Default != nil && *a.Default {
		validator.String("alertScheduleName", a.Name).Exists().Empty()
		validator.Bool("isEnabled", a.Enabled).Exists()
		validator.String("startTime", a.StartTime).Exists().EqualTo(AlertScheduleSettingsStartTimeDefault)
		validator.String("endTime", a.EndTime).Exists().EqualTo(AlertScheduleSettingsEndTimeDefault)
		validator.StringArray("daysOfWeek", a.DaysOfWeek).Exists().EachOneOf(AlertScheduleSettingsDays()...).EachUnique().LengthEqualTo(len(AlertScheduleSettingsDays()))
	} else {
		validator.String("alertScheduleName", a.Name).Exists().NotEmpty()
		validator.Bool("isEnabled", a.Enabled).Exists()
		validator.String("startTime", a.StartTime).Exists().Using(AlertScheduleSettingsTimeValidator)
		validator.String("endTime", a.EndTime).Exists().Using(AlertScheduleSettingsTimeValidator)
		validator.StringArray("daysOfWeek", a.DaysOfWeek).Exists().EachOneOf(AlertScheduleSettingsDays()...).EachUnique()
	}
}

func (a *AlertScheduleSettings) Normalize(normalizer structure.Normalizer) {
	if a.DaysOfWeek != nil {
		sort.Sort(DaysOfWeekByAlertScheduleSettingsDayIndex(*a.DaysOfWeek))
	}
}

func (a *AlertScheduleSettings) IsDefault() bool {
	return a.Default != nil && *a.Default
}

type DaysOfWeekByAlertScheduleSettingsDayIndex []string

func (d DaysOfWeekByAlertScheduleSettingsDayIndex) Len() int {
	return len(d)
}
func (d DaysOfWeekByAlertScheduleSettingsDayIndex) Swap(i int, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d DaysOfWeekByAlertScheduleSettingsDayIndex) Less(i int, j int) bool {
	return AlertScheduleSettingsDayIndex(d[i]) < AlertScheduleSettingsDayIndex(d[j])
}

type AlertSettings []*AlertSetting

func ParseAlertSettings(parser structure.ArrayParser) *AlertSettings {
	if !parser.Exists() {
		return nil
	}
	datum := NewAlertSettings()
	parser.Parse(datum)
	return datum
}

func NewAlertSettings() *AlertSettings {
	return &AlertSettings{}
}

func (a *AlertSettings) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*a = append(*a, ParseAlertSetting(parser.WithReferenceObjectParser(reference)))
	}
}

func (a *AlertSettings) Validate(validator structure.Validator) {
	if len(*a) == 0 {
		validator.ReportError(structureValidator.ErrorValueEmpty())
	}
	for index, alertSetting := range *a {
		if alertSettingValidator := validator.WithReference(strconv.Itoa(index)); alertSetting != nil {
			alertSetting.Validate(alertSettingValidator)
		} else {
			alertSettingValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}

	if !validator.HasError() {
		a.Deduplicate() // HACK: Dexcom - duplicates not allowed; delete older
	}
}

func (a *AlertSettings) Normalize(normalizer structure.Normalizer) {
	sort.Sort(AlertSettingsByAlertSettingAlertName(*a))
	for index, alertSetting := range *a {
		alertSetting.Normalize(normalizer.WithReference(strconv.Itoa(index)))
	}
}

func (a *AlertSettings) Deduplicate() {
	alertSettings := []*AlertSetting{}
	alertNameMap := map[string]bool{}
	for outerIndex, alertSetting := range *a {
		if alertSetting.AlertName == nil {
			alertSettings = append(alertSettings, alertSetting)
		} else if !alertNameMap[*alertSetting.AlertName] {
			for innerIndex := outerIndex + 1; innerIndex < len(*a); innerIndex++ {
				if (*a)[innerIndex].IsNewerMatchThan(alertSetting) {
					alertSetting = nil
					break
				}
			}
			if alertSetting != nil {
				alertSettings = append(alertSettings, alertSetting)
				alertNameMap[*alertSetting.AlertName] = true
			}
		}
	}
	*a = alertSettings
}

type AlertSettingsByAlertSettingAlertName AlertSettings

func (a AlertSettingsByAlertSettingAlertName) Len() int {
	return len(a)
}
func (a AlertSettingsByAlertSettingAlertName) Swap(i int, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a AlertSettingsByAlertSettingAlertName) Less(i int, j int) bool {
	if left := a[i]; left == nil {
		return true
	} else if right := a[j]; right == nil {
		return false
	} else if leftAlertName := left.AlertName; leftAlertName == nil {
		return true
	} else if rightAlertName := right.AlertName; rightAlertName == nil {
		return false
	} else {
		return *leftAlertName < *rightAlertName
	}
}

type AlertSetting struct {
	SystemTime  *Time    `json:"systemTime,omitempty" yaml:"-"`
	DisplayTime *Time    `json:"displayTime,omitempty" yaml:"-"`
	AlertName   *string  `json:"alertName,omitempty" yaml:"alertName,omitempty"`
	Unit        *string  `json:"unit,omitempty" yaml:"unit,omitempty"`
	Value       *float64 `json:"value,omitempty" yaml:"value,omitempty"`
	Snooze      *int     `json:"snooze,omitempty" yaml:"snooze,omitempty"`
	Enabled     *bool    `json:"enabled,omitempty" yaml:"enabled,omitempty"`
}

func ParseAlertSetting(parser structure.ObjectParser) *AlertSetting {
	if !parser.Exists() {
		return nil
	}
	datum := NewAlertSetting()
	parser.Parse(datum)
	return datum
}

func NewAlertSetting() *AlertSetting {
	return &AlertSetting{}
}

func (a *AlertSetting) Parse(parser structure.ObjectParser) {
	a.SystemTime = TimeFromRaw(parser.Time("systemTime", TimeFormat))
	a.DisplayTime = TimeFromRaw(parser.Time("displayTime", TimeFormat))
	a.AlertName = parser.String("alertName")
	a.Unit = parser.String("unit")
	a.Value = parser.Float64("value")
	a.Snooze = parser.Int("snooze")
	a.Enabled = parser.Bool("enabled")
}

func (a *AlertSetting) Validate(validator structure.Validator) {
	validator = validator.WithMeta(a)
	validator.Time("systemTime", a.SystemTime.Raw()).Exists().NotZero().BeforeNow(SystemTimeNowThreshold)
	validator.Time("displayTime", a.DisplayTime.Raw()).Exists().NotZero()
	validator.String("alertName", a.AlertName).Exists().OneOf(AlertSettingAlertNames()...)

	if a.AlertName != nil {
		switch *a.AlertName {
		case AlertSettingAlertNameFall:
			a.validateFall(validator)
		case AlertSettingAlertNameHigh:
			a.validateHigh(validator)
		case AlertSettingAlertNameLow:
			a.validateLow(validator)
		case AlertSettingAlertNameNoReadings:
			a.validateNoReadings(validator)
		case AlertSettingAlertNameOutOfRange:
			a.validateOutOfRange(validator)
		case AlertSettingAlertNameRise:
			a.validateRise(validator)
		case AlertSettingAlertNameUrgentLow:
			a.validateUrgentLow(validator)
		case AlertSettingAlertNameUrgentLowSoon:
			a.validateUrgentLowSoon(validator)
		}
	}
}

func (a *AlertSetting) Normalize(normalizer structure.Normalizer) {}

func (a *AlertSetting) IsNewerMatchThan(alertSetting *AlertSetting) bool {
	return a.AlertName != nil && alertSetting.AlertName != nil && *a.AlertName == *alertSetting.AlertName &&
		a.SystemTime != nil && alertSetting.SystemTime != nil && a.SystemTime.After(*alertSetting.SystemTime.Raw())
}

func (a *AlertSetting) validateFall(validator structure.Validator) {
	// HACK: Dexcom - negative value is invalid; use positive value instead (per Dexcom)
	if a.Value != nil && *a.Value < 0 {
		a.Value = pointer.FromFloat64(-*a.Value)
	}

	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingUnitFalls()...)
	if a.Unit != nil {
		switch *a.Unit {
		case AlertSettingUnitMgdLMinute:
			validator.Float64("value", a.Value).Exists().OneOf(AlertSettingValueFallMgdLMinutes()...)
		}
	}
	validator.Int("snooze", a.Snooze).OneOf(AlertSettingSnoozeFalls()...)
	validator.Bool("enabled", a.Enabled).Exists()
}

func (a *AlertSetting) validateHigh(validator structure.Validator) {
	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingUnitHighs()...)
	if a.Unit != nil {
		switch *a.Unit {
		case AlertSettingUnitMgdL:
			validator.Float64("value", a.Value).Exists().OneOf(AlertSettingValueHighMgdLs()...)
		}
	}
	validator.Int("snooze", a.Snooze).Exists().OneOf(AlertSettingSnoozeHighs()...)
	validator.Bool("enabled", a.Enabled).Exists()
}

func (a *AlertSetting) validateLow(validator structure.Validator) {
	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingUnitLows()...)
	if a.Unit != nil {
		switch *a.Unit {
		case AlertSettingUnitMgdL:
			validator.Float64("value", a.Value).Exists().OneOf(AlertSettingValueLowMgdLs()...)
		}
	}
	validator.Int("snooze", a.Snooze).Exists().OneOf(AlertSettingSnoozeLows()...)
	validator.Bool("enabled", a.Enabled).Exists()
}

func (a *AlertSetting) validateNoReadings(validator structure.Validator) {
	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingUnitNoReadings()...)
	if a.Unit != nil {
		switch *a.Unit {
		case AlertSettingUnitMinutes:
			validator.Float64("value", a.Value).Exists().OneOf(AlertSettingValueNoReadingsMinutes()...)
		}
	}
	validator.Int("snooze", a.Snooze).OneOf(AlertSettingSnoozeNoReadings()...)
	validator.Bool("enabled", a.Enabled).Exists()
}

func (a *AlertSetting) validateOutOfRange(validator structure.Validator) {
	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingUnitOutOfRanges()...)
	if a.Unit != nil {
		switch *a.Unit {
		case AlertSettingUnitMinutes:
			validator.Float64("value", a.Value).Exists().OneOf(AlertSettingValueOutOfRangeMinutes()...)
		}
	}
	validator.Int("snooze", a.Snooze).OneOf(AlertSettingSnoozeOutOfRanges()...)
	validator.Bool("enabled", a.Enabled).Exists()
}

func (a *AlertSetting) validateRise(validator structure.Validator) {
	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingUnitRises()...)
	if a.Unit != nil {
		switch *a.Unit {
		case AlertSettingUnitMgdLMinute:
			validator.Float64("value", a.Value).Exists().OneOf(AlertSettingValueRiseMgdLMinutes()...)
		}
	}
	validator.Int("snooze", a.Snooze).OneOf(AlertSettingSnoozeRises()...)
	validator.Bool("enabled", a.Enabled).Exists()
}

func (a *AlertSetting) validateUrgentLow(validator structure.Validator) {
	// HACK: Dexcom - snooze of 28 is invalid; use snooze of 30 instead (per Dexcom); exists in v2 (20180914)
	if a.Snooze != nil && *a.Snooze == 28 {
		a.Snooze = pointer.FromInt(30)
	}

	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingUnitUrgentLows()...)
	if a.Unit != nil {
		switch *a.Unit {
		case AlertSettingUnitMgdL:
			validator.Float64("value", a.Value).Exists().OneOf(AlertSettingValueUrgentLowMgdLs()...)
		}
	}
	validator.Int("snooze", a.Snooze).Exists().OneOf(AlertSettingSnoozeUrgentLows()...)
	validator.Bool("enabled", a.Enabled).Exists().True()
}

func (a *AlertSetting) validateUrgentLowSoon(validator structure.Validator) {
	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingUnitUrgentLowSoons()...)
	if a.Unit != nil {
		switch *a.Unit {
		case AlertSettingUnitMgdL:
			validator.Float64("value", a.Value).Exists().OneOf(AlertSettingValueUrgentLowSoonMgdLs()...)
		}
	}
	validator.Int("snooze", a.Snooze).Exists().OneOf(AlertSettingSnoozeUrgentLowSoons()...)
	validator.Bool("enabled", a.Enabled).Exists()
}

func ParseAlertScheduleSettingsTime(value string) (int, int, bool) {
	if submatch := alertScheduleSettingsTimeExpression.FindStringSubmatch(value); len(submatch) != 3 {
		return 0, 0, false
	} else if hour, hourErr := strconv.Atoi(submatch[1]); hourErr != nil || hour < 0 || hour > 23 {
		return 0, 0, false
	} else if minute, minuteErr := strconv.Atoi(submatch[2]); minuteErr != nil || minute < 0 || minute > 59 {
		return 0, 0, false
	} else {
		return hour, minute, true
	}
}

func IsValidAlertScheduleSettingsTime(value string) bool {
	return ValidateAlertScheduleSettingsTime(value) == nil
}

func AlertScheduleSettingsTimeValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateAlertScheduleSettingsTime(value))
}

func ValidateAlertScheduleSettingsTime(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if _, _, ok := ParseAlertScheduleSettingsTime(value); !ok {
		return ErrorValueStringAsAlertScheduleSettingsTimeNotValid(value)
	}
	return nil
}

func ErrorValueStringAsAlertScheduleSettingsTimeNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as alert schedule settings time", value)
}

var alertScheduleSettingsTimeExpression = regexp.MustCompile("^([0-9][0-9]):([0-9][0-9])$")

func generateFloatRange(min float64, max float64, step float64) []float64 {
	r := []float64{}
	for v := min; v <= max; v += step {
		r = append(r, v)
	}
	return r
}

func generateIntegerRange(min int, max int, step int) []int {
	r := []int{}
	for v := min; v <= max; v += step {
		r = append(r, v)
	}
	return r
}
