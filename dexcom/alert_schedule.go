package dexcom

import (
	"sort"
	"strconv"
	"strings"
	"time"

	dataTypesSettingsCgm "github.com/tidepool-org/platform/data/types/settings/cgm"
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

	AlertScheduleSettingsOverrideModeUnknown = "unknown"
	AlertScheduleSettingsOverrideModeQuiet   = "quiet"
	AlertScheduleSettingsOverrideModeVibrate = "vibrate"

	AlertSettingAlertNameUnknown       = "unknown"
	AlertSettingAlertNameFall          = "fall"
	AlertSettingAlertNameHigh          = "high"
	AlertSettingAlertNameLow           = "low"
	AlertSettingAlertNameNoReadings    = "noReadings"
	AlertSettingAlertNameOutOfRange    = "outOfRange"
	AlertSettingAlertNameRise          = "rise"
	AlertSettingAlertNameUrgentLow     = "urgentLow"
	AlertSettingAlertNameUrgentLowSoon = "urgentLowSoon"
	AlertSettingAlertNameFixedLow      = "fixedLow"

	AlertSettingSnoozeMinutesMaximum = dataTypesSettingsCgm.SnoozeDurationMinutesMaximum
	AlertSettingSnoozeMinutesMinimum = dataTypesSettingsCgm.SnoozeDurationMinutesMinimum

	AlertSettingUnitUnknown    = "unknown"
	AlertSettingUnitMinutes    = "minutes"
	AlertSettingUnitMgdL       = "mg/dL"
	AlertSettingUnitMgdLMinute = "mg/dL/min"

	AlertSettingValueFallMgdLMinuteMaximum    = dataTypesSettingsCgm.FallAlertRateMgdLMinuteMaximum
	AlertSettingValueFallMgdLMinuteMinimum    = dataTypesSettingsCgm.FallAlertRateMgdLMinuteMinimum
	AlertSettingValueHighMgdLMaximum          = dataTypesSettingsCgm.HighAlertLevelMgdLMaximum
	AlertSettingValueHighMgdLMinimum          = dataTypesSettingsCgm.HighAlertLevelMgdLMinimum
	AlertSettingValueLowMgdLMaximum           = dataTypesSettingsCgm.LowAlertLevelMgdLMaximum
	AlertSettingValueLowMgdLMinimum           = dataTypesSettingsCgm.LowAlertLevelMgdLMinimum
	AlertSettingValueNoReadingsMgdLMaximum    = dataTypesSettingsCgm.NoDataAlertDurationMinutesMaximum
	AlertSettingValueNoReadingsMgdLMinimum    = dataTypesSettingsCgm.NoDataAlertDurationMinutesMinimum
	AlertSettingValueOutOfRangeMgdLMaximum    = dataTypesSettingsCgm.OutOfRangeAlertDurationMinutesMaximum
	AlertSettingValueOutOfRangeMgdLMinimum    = dataTypesSettingsCgm.OutOfRangeAlertDurationMinutesMinimum
	AlertSettingValueRiseMgdLMinuteMaximum    = dataTypesSettingsCgm.RiseAlertRateMgdLMinuteMaximum
	AlertSettingValueRiseMgdLMinuteMinimum    = dataTypesSettingsCgm.RiseAlertRateMgdLMinuteMinimum
	AlertSettingValueUrgentLowMgdLMaximum     = dataTypesSettingsCgm.UrgentLowAlertLevelMgdLMaximum
	AlertSettingValueUrgentLowMgdLMinimum     = dataTypesSettingsCgm.UrgentLowAlertLevelMgdLMinimum
	AlertSettingValueUrgentLowSoonMgdLMaximum = dataTypesSettingsCgm.UrgentLowAlertLevelMgdLMaximum
	AlertSettingValueUrgentLowSoonMgdLMinimum = dataTypesSettingsCgm.UrgentLowAlertLevelMgdLMinimum

	AlertSettingSoundThemeUnknown = "unknown"
	AlertSettingSoundThemeModern  = "modern"
	AlertSettingSoundThemeClassic = "classic"

	AlertSettingSoundOutputModeUnknown = "unknown"
	AlertSettingSoundOutputModeSound   = "sound"
	AlertSettingSoundOutputModeVibrate = "vibrate"
	AlertSettingSoundOutputModeMatch   = "match"
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

func AlertScheduleSettingsOverrideModes() []string {
	return []string{
		AlertScheduleSettingsOverrideModeUnknown,
		AlertScheduleSettingsOverrideModeQuiet,
		AlertScheduleSettingsOverrideModeVibrate,
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
		AlertSettingAlertNameUnknown,
		AlertSettingAlertNameFall,
		AlertSettingAlertNameHigh,
		AlertSettingAlertNameLow,
		AlertSettingAlertNameNoReadings,
		AlertSettingAlertNameOutOfRange,
		AlertSettingAlertNameRise,
		AlertSettingAlertNameUrgentLow,
		AlertSettingAlertNameUrgentLowSoon,
		AlertSettingAlertNameFixedLow,
	}
}

func AlertSettingSoundOutputModes() []string {
	return []string{
		AlertSettingSoundOutputModeUnknown,
		AlertSettingSoundOutputModeSound,
		AlertSettingSoundOutputModeVibrate,
		AlertSettingSoundOutputModeMatch,
	}
}

func AlertSettingSoundThemes() []string {
	return []string{
		AlertSettingSoundThemeUnknown,
		AlertSettingSoundThemeModern,
		AlertSettingSoundThemeClassic,
	}
}

func AlertSettingUnitFalls() []string {
	return []string{AlertSettingUnitMgdLMinute}
}

func AlertSettingUnitHighs() []string {
	return []string{AlertSettingUnitMgdL}
}

func AlertSettingUnitLows() []string {
	return []string{AlertSettingUnitMgdL}
}

func AlertSettingUnitNoReadings() []string {
	return []string{AlertSettingUnitMinutes}
}

func AlertSettingUnitOutOfRanges() []string {
	return []string{AlertSettingUnitMinutes}
}

func AlertSettingUnitRises() []string {
	return []string{AlertSettingUnitMgdLMinute}
}

func AlertSettingUnitUrgentLows() []string {
	return []string{AlertSettingUnitMgdL}
}

func AlertSettingUnitUrgentLowSoons() []string {
	return []string{AlertSettingUnitMgdL}
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
	// it is valid for a receiver to send an empty list
	if len(*a) != 0 {
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
	Name       *string          `json:"alertScheduleName,omitempty" yaml:"alertScheduleName,omitempty"`
	Enabled    *bool            `json:"isEnabled,omitempty" yaml:"isEnabled,omitempty"`
	Default    *bool            `json:"isDefaultSchedule,omitempty" yaml:"isDefaultSchedule,omitempty"`
	StartTime  *string          `json:"startTime,omitempty" yaml:"startTime,omitempty"`
	EndTime    *string          `json:"endTime,omitempty" yaml:"endTime,omitempty"`
	DaysOfWeek *[]string        `json:"daysOfWeek,omitempty" yaml:"daysOfWeek,omitempty"`
	Active     *bool            `json:"isActive,omitempty" yaml:"isActive,omitempty"`
	Override   *OverrideSetting `json:"override,omitempty" yaml:"override,omitempty"`
}

type OverrideSetting struct {
	Enabled *bool   `json:"isOverrideEnabled,omitempty" yaml:"isOverrideEnabled,omitempty"`
	Mode    *string `json:"mode,omitempty" yaml:"mode,omitempty"`
	EndTime *string `json:"endTime,omitempty" yaml:"endTime,omitempty"`
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
	a.Active = parser.Bool("isActive")
	a.Override = ParseOverrideSetting(parser.WithReferenceObjectParser("override"))
}

func ParseOverrideSetting(parser structure.ObjectParser) *OverrideSetting {
	if !parser.Exists() {
		return nil
	}
	datum := &OverrideSetting{}
	parser.Parse(datum)
	return datum
}

func (o *OverrideSetting) Parse(parser structure.ObjectParser) {
	o.Enabled = parser.Bool("isOverrideEnabled")
	o.Mode = parser.String("mode")
	o.EndTime = parser.String("endTime")
}

func (a *AlertScheduleSettings) Validate(validator structure.Validator) {
	// HACK: Dexcom - force default schedule to use expected startTime and endTime
	if a.Default != nil && *a.Default {
		a.StartTime = pointer.FromString(AlertScheduleSettingsStartTimeDefault)
		a.EndTime = pointer.FromString(AlertScheduleSettingsEndTimeDefault)
	}

	// HACK: Dexcom - remove empty strings from daysOfWeek
	if a.DaysOfWeek != nil {
		daysOfWeek := []string{}
		for _, dayOfWeek := range *a.DaysOfWeek {
			if dayOfWeek != "" {
				daysOfWeek = append(daysOfWeek, dayOfWeek)
			}
		}
		a.DaysOfWeek = pointer.FromStringArray(daysOfWeek)
	}

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
	SystemTime                *Time    `json:"systemTime,omitempty" yaml:"-"`
	DisplayTime               *Time    `json:"displayTime,omitempty" yaml:"-"`
	AlertName                 *string  `json:"alertName,omitempty" yaml:"alertName,omitempty"`
	Unit                      *string  `json:"unit,omitempty" yaml:"unit,omitempty"`
	Value                     *float64 `json:"value,omitempty" yaml:"value,omitempty"`
	Snooze                    *int     `json:"snooze,omitempty" yaml:"snooze,omitempty"`
	Enabled                   *bool    `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Delay                     *int     `json:"delay,omitempty" yaml:"delay,omitempty"`
	SecondaryTriggerCondition *int     `json:"secondaryTriggerCondition,omitempty" yaml:"secondaryTriggerCondition,omitempty"`
	SoundTheme                *string  `json:"soundTheme,omitempty" yaml:"soundTheme,omitempty"`
	SoundOutputMode           *string  `json:"soundOutputMode,omitempty" yaml:"soundOutputMode,omitempty"`
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
	a.SystemTime = ParseTime(parser, "systemTime")
	a.DisplayTime = ParseTime(parser, "displayTime")
	a.AlertName = parser.String("alertName")
	a.Unit = parser.String("unit")
	a.Value = parser.Float64("value")
	a.Snooze = parser.Int("snooze")
	a.Enabled = parser.Bool("enabled")
	a.Delay = parser.Int("delay")
	a.SecondaryTriggerCondition = parser.Int("secondaryTriggerCondition")
	a.SoundTheme = parser.String("soundTheme")
	a.SoundOutputMode = parser.String("soundOutputMode")
}

func (a *AlertSetting) Validate(validator structure.Validator) {
	validator = validator.WithMeta(a)
	if a.SystemTime != nil {
		validator.Time("systemTime", a.SystemTime.Raw()).Exists().NotZero().BeforeNow(SystemTimeNowThreshold)
	}
	if a.DisplayTime != nil {
		validator.Time("displayTime", a.DisplayTime.Raw()).Exists().NotZero()
	}
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
		case AlertSettingAlertNameFixedLow:
			a.validateFixedLow(validator)
		case AlertSettingAlertNameUnknown:
			a.validateUnknown(validator)
		}
	}

	validator.String("soundTheme", a.SoundTheme).OneOf(AlertSettingSoundThemes()...)
	validator.String("soundOutputMode", a.SoundOutputMode).OneOf(AlertSettingSoundOutputModes()...)
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
			validator.Float64("value", a.Value).Exists().InRange(AlertSettingValueFallMgdLMinuteMinimum, AlertSettingValueFallMgdLMinuteMaximum)
		}
	}
	validator.Int("snooze", a.Snooze).InRange(AlertSettingSnoozeMinutesMinimum, AlertSettingSnoozeMinutesMaximum)
	validator.Bool("enabled", a.Enabled).Exists()
}

func (a *AlertSetting) validateHigh(validator structure.Validator) {
	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingUnitHighs()...)
	if a.Unit != nil {
		switch *a.Unit {
		case AlertSettingUnitMgdL:
			validator.Float64("value", a.Value).Exists().InRange(AlertSettingValueHighMgdLMinimum, AlertSettingValueHighMgdLMaximum)
		}
	}
	validator.Int("snooze", a.Snooze).Exists().InRange(AlertSettingSnoozeMinutesMinimum, AlertSettingSnoozeMinutesMaximum)
	validator.Bool("enabled", a.Enabled).Exists()
}

func (a *AlertSetting) validateLow(validator structure.Validator) {
	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingUnitLows()...)
	if a.Unit != nil {
		switch *a.Unit {
		case AlertSettingUnitMgdL:
			validator.Float64("value", a.Value).Exists().InRange(AlertSettingValueLowMgdLMinimum, AlertSettingValueLowMgdLMaximum)
		}
	}
	validator.Int("snooze", a.Snooze).Exists().InRange(AlertSettingSnoozeMinutesMinimum, AlertSettingSnoozeMinutesMaximum)
	validator.Bool("enabled", a.Enabled).Exists()
}

func (a *AlertSetting) validateNoReadings(validator structure.Validator) {
	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingUnitNoReadings()...)
	if a.Unit != nil {
		switch *a.Unit {
		case AlertSettingUnitMinutes:
			validator.Float64("value", a.Value).Exists().InRange(AlertSettingValueNoReadingsMgdLMinimum, AlertSettingValueNoReadingsMgdLMaximum)
		}
	}
	validator.Int("snooze", a.Snooze).InRange(AlertSettingSnoozeMinutesMinimum, AlertSettingSnoozeMinutesMaximum)
	validator.Bool("enabled", a.Enabled).Exists()
}

func (a *AlertSetting) validateOutOfRange(validator structure.Validator) {
	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingUnitOutOfRanges()...)
	if a.Unit != nil {
		switch *a.Unit {
		case AlertSettingUnitMinutes:
			validator.Float64("value", a.Value).Exists().InRange(AlertSettingValueOutOfRangeMgdLMinimum, AlertSettingValueOutOfRangeMgdLMaximum)
		}
	}
	validator.Int("snooze", a.Snooze).InRange(AlertSettingSnoozeMinutesMinimum, AlertSettingSnoozeMinutesMaximum)
	validator.Bool("enabled", a.Enabled).Exists()
}

func (a *AlertSetting) validateRise(validator structure.Validator) {
	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingUnitRises()...)
	if a.Unit != nil {
		switch *a.Unit {
		case AlertSettingUnitMgdLMinute:
			validator.Float64("value", a.Value).Exists().InRange(AlertSettingValueRiseMgdLMinuteMinimum, AlertSettingValueRiseMgdLMinuteMaximum)
		}
	}
	validator.Int("snooze", a.Snooze).InRange(AlertSettingSnoozeMinutesMinimum, AlertSettingSnoozeMinutesMaximum)
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
			validator.Float64("value", a.Value).Exists().InRange(AlertSettingValueUrgentLowMgdLMinimum, AlertSettingValueUrgentLowMgdLMaximum)
		}
	}
	validator.Int("snooze", a.Snooze).Exists().InRange(AlertSettingSnoozeMinutesMinimum, AlertSettingSnoozeMinutesMaximum)
	validator.Bool("enabled", a.Enabled).Exists().True()
}

func (a *AlertSetting) validateUrgentLowSoon(validator structure.Validator) {
	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingUnitUrgentLowSoons()...)
	if a.Unit != nil {
		switch *a.Unit {
		case AlertSettingUnitMgdL:
			validator.Float64("value", a.Value).Exists().InRange(AlertSettingValueUrgentLowSoonMgdLMinimum, AlertSettingValueUrgentLowSoonMgdLMaximum)
		}
	}
	validator.Int("snooze", a.Snooze).Exists().InRange(AlertSettingSnoozeMinutesMinimum, AlertSettingSnoozeMinutesMaximum)
	validator.Bool("enabled", a.Enabled).Exists()
}

func (a *AlertSetting) validateUnknown(validator structure.Validator) {
	validator.String("unit", a.Unit).OneOf(AlertSettingUnitUnknown)
	validator.Bool("enabled", a.Enabled).Exists()
}

func (a *AlertSetting) validateFixedLow(validator structure.Validator) {
	validator.Bool("enabled", a.Enabled).Exists()
}

func ParseAlertScheduleSettingsTime(value string) (int, int, bool) {
	timeFormat := "15:04"
	value = strings.ToUpper(value)
	if strings.Contains(value, "AM") || strings.Contains(value, "PM") {
		timeFormat = "3:04PM"
		value = strings.ReplaceAll(value, " ", "")
	}
	t, err := time.Parse(timeFormat, value)
	if err != nil {
		return 0, 0, false
	}
	return t.Hour(), t.Minute(), true
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
