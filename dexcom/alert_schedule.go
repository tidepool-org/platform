package dexcom

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	dataTypesCommon "github.com/tidepool-org/platform/data/types/common"
	dataTypesSettingsCgm "github.com/tidepool-org/platform/data/types/settings/cgm"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	AlertScheduleSettingsStartTimeDefault          = "00:00"
	AlertScheduleSettingsStartTimeDefaultAlternate = "0:00"
	AlertScheduleSettingsEndTimeDefault            = "00:00"
	AlertScheduleSettingsEndTimeDefaultAlternate   = "0:00"

	AlertScheduleSettingsDaySunday    = dataTypesCommon.DaySunday
	AlertScheduleSettingsDayMonday    = dataTypesCommon.DayMonday
	AlertScheduleSettingsDayTuesday   = dataTypesCommon.DayTuesday
	AlertScheduleSettingsDayWednesday = dataTypesCommon.DayWednesday
	AlertScheduleSettingsDayThursday  = dataTypesCommon.DayThursday
	AlertScheduleSettingsDayFriday    = dataTypesCommon.DayFriday
	AlertScheduleSettingsDaySaturday  = dataTypesCommon.DaySaturday

	AlertScheduleSettingsOverrideModeUnknown = "unknown"
	AlertScheduleSettingsOverrideModeQuiet   = "quiet"
	AlertScheduleSettingsOverrideModeVibrate = "vibrate"

	AlertSettingAlertNameUnknown       = "unknown"
	AlertSettingAlertNameHigh          = "high"
	AlertSettingAlertNameLow           = "low"
	AlertSettingAlertNameRise          = "rise"
	AlertSettingAlertNameFall          = "fall"
	AlertSettingAlertNameOutOfRange    = "outOfRange"
	AlertSettingAlertNameUrgentLow     = "urgentLow"
	AlertSettingAlertNameUrgentLowSoon = "urgentLowSoon"
	AlertSettingAlertNameNoReadings    = "noReadings"
	AlertSettingAlertNameFixedLow      = "fixedLow"

	AlertSettingSnoozeMinutesMaximum = dataTypesSettingsCgm.SnoozeDurationMinutesMaximum
	AlertSettingSnoozeMinutesMinimum = dataTypesSettingsCgm.SnoozeDurationMinutesMinimum

	AlertSettingDelayMinimum = 0

	AlertSettingUnitUnknown     = "unknown"
	AlertSettingUnitMinutes     = "minutes"
	AlertSettingUnitMgdL        = "mg/dL"
	AlertSettingUnitMmolL       = "mmol/L"
	AlertSettingUnitMgdLMinute  = "mg/dL/min"
	AlertSettingUnitMmolLMinute = "mmol/L/min"

	AlertSettingValueHighMgdLMaximum  = dataTypesSettingsCgm.HighAlertLevelMgdLMaximum
	AlertSettingValueHighMgdLMinimum  = dataTypesSettingsCgm.HighAlertLevelMgdLMinimum
	AlertSettingValueHighMmolLMaximum = dataTypesSettingsCgm.HighAlertLevelMmolLMaximum
	AlertSettingValueHighMmolLMinimum = dataTypesSettingsCgm.HighAlertLevelMmolLMinimum

	AlertSettingValueLowMgdLMaximum  = dataTypesSettingsCgm.LowAlertLevelMgdLMaximum
	AlertSettingValueLowMgdLMinimum  = dataTypesSettingsCgm.LowAlertLevelMgdLMinimum
	AlertSettingValueLowMmolLMaximum = dataTypesSettingsCgm.LowAlertLevelMmolLMaximum
	AlertSettingValueLowMmolLMinimum = dataTypesSettingsCgm.LowAlertLevelMmolLMinimum

	AlertSettingValueRiseMgdLMinuteMaximum  = dataTypesSettingsCgm.RiseAlertRateMgdLMinuteMaximum
	AlertSettingValueRiseMgdLMinuteMinimum  = dataTypesSettingsCgm.RiseAlertRateMgdLMinuteMinimum
	AlertSettingValueRiseMmolLMinuteMaximum = dataTypesSettingsCgm.RiseAlertRateMmolLMinuteMaximum
	AlertSettingValueRiseMmolLMinuteMinimum = dataTypesSettingsCgm.RiseAlertRateMmolLMinuteMinimum

	AlertSettingValueFallMgdLMinuteMaximum  = dataTypesSettingsCgm.FallAlertRateMgdLMinuteMaximum
	AlertSettingValueFallMgdLMinuteMinimum  = dataTypesSettingsCgm.FallAlertRateMgdLMinuteMinimum
	AlertSettingValueFallMmolLMinuteMaximum = dataTypesSettingsCgm.FallAlertRateMmolLMinuteMaximum
	AlertSettingValueFallMmolLMinuteMinimum = dataTypesSettingsCgm.FallAlertRateMmolLMinuteMinimum

	AlertSettingValueOutOfRangeMinutesMaximum = dataTypesSettingsCgm.OutOfRangeAlertDurationMinutesMaximum
	AlertSettingValueOutOfRangeMinutesMinimum = dataTypesSettingsCgm.OutOfRangeAlertDurationMinutesMinimum

	AlertSettingValueUrgentLowMgdLMaximum  = dataTypesSettingsCgm.UrgentLowAlertLevelMgdLMaximum
	AlertSettingValueUrgentLowMgdLMinimum  = dataTypesSettingsCgm.UrgentLowAlertLevelMgdLMinimum
	AlertSettingValueUrgentLowMmolLMaximum = dataTypesSettingsCgm.UrgentLowAlertLevelMmolLMaximum
	AlertSettingValueUrgentLowMmolLMinimum = dataTypesSettingsCgm.UrgentLowAlertLevelMmolLMinimum

	AlertSettingValueUrgentLowSoonMgdLMaximum  = dataTypesSettingsCgm.UrgentLowAlertLevelMgdLMaximum
	AlertSettingValueUrgentLowSoonMgdLMinimum  = dataTypesSettingsCgm.UrgentLowAlertLevelMgdLMinimum
	AlertSettingValueUrgentLowSoonMmolLMaximum = dataTypesSettingsCgm.UrgentLowAlertLevelMmolLMaximum
	AlertSettingValueUrgentLowSoonMmolLMinimum = dataTypesSettingsCgm.UrgentLowAlertLevelMmolLMinimum

	AlertSettingValueNoReadingsMinutesMaximum = dataTypesSettingsCgm.NoDataAlertDurationMinutesMaximum
	AlertSettingValueNoReadingsMinutesMinimum = dataTypesSettingsCgm.NoDataAlertDurationMinutesMinimum

	AlertSettingValueFixedLowMinutesMaximum = dataTypesSettingsCgm.NoDataAlertDurationMinutesMaximum
	AlertSettingValueFixedLowMinutesMinimum = dataTypesSettingsCgm.NoDataAlertDurationMinutesMinimum

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

func AlertSettingAlertNames() []string {
	return []string{
		AlertSettingAlertNameUnknown,
		AlertSettingAlertNameHigh,
		AlertSettingAlertNameLow,
		AlertSettingAlertNameRise,
		AlertSettingAlertNameFall,
		AlertSettingAlertNameOutOfRange,
		AlertSettingAlertNameUrgentLow,
		AlertSettingAlertNameUrgentLowSoon,
		AlertSettingAlertNameNoReadings,
		AlertSettingAlertNameFixedLow,
	}
}

func AlertSettingSoundThemes() []string {
	return []string{
		AlertSettingSoundThemeUnknown,
		AlertSettingSoundThemeModern,
		AlertSettingSoundThemeClassic,
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

func AlertSettingUnitUnknowns() []string {
	return []string{
		AlertSettingUnitUnknown,
	}
}

func AlertSettingUnitHighs() []string {
	return []string{
		AlertSettingUnitMgdL,
		AlertSettingUnitMmolL,
	}
}

func AlertSettingUnitLows() []string {
	return []string{
		AlertSettingUnitMgdL,
		AlertSettingUnitMmolL,
	}
}

func AlertSettingUnitRises() []string {
	return []string{
		AlertSettingUnitMgdLMinute,
		AlertSettingUnitMmolLMinute,
	}
}

func AlertSettingUnitFalls() []string {
	return []string{
		AlertSettingUnitMgdLMinute,
		AlertSettingUnitMmolLMinute,
	}
}

func AlertSettingUnitOutOfRanges() []string {
	return []string{
		AlertSettingUnitMinutes,
	}
}

func AlertSettingUnitUrgentLows() []string {
	return []string{
		AlertSettingUnitMgdL,
		AlertSettingUnitMmolL,
	}
}

func AlertSettingUnitUrgentLowSoons() []string {
	return []string{
		AlertSettingUnitMgdL,
		AlertSettingUnitMmolL,
	}
}

func AlertSettingUnitNoReadings() []string {
	return []string{
		AlertSettingUnitMinutes,
	}
}

func AlertSettingUnitFixedLows() []string {
	return []string{
		AlertSettingUnitMinutes,
	}
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
	sort.Stable(AlertSchedulesByName(*a))
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

type AlertSchedulesByName AlertSchedules

func (a AlertSchedulesByName) Len() int {
	return len(a)
}
func (a AlertSchedulesByName) Swap(i int, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a AlertSchedulesByName) Less(i int, j int) bool {
	left := a[i]
	right := a[j]
	if left == nil && right == nil {
		return false
	} else if left == nil {
		return true
	} else if right == nil {
		return false
	}

	leftName := left.Name()
	rightName := right.Name()
	if leftName == nil && rightName == nil {
		return false
	} else if leftName == nil {
		return true
	} else if rightName == nil {
		return false
	}

	return *leftName < *rightName
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
		return a.AlertScheduleSettings.AlertScheduleName
	}
	return nil
}

type AlertScheduleSettings struct {
	IsDefaultSchedule *bool     `json:"isDefaultSchedule,omitempty" yaml:"isDefaultSchedule,omitempty"`
	IsEnabled         *bool     `json:"isEnabled,omitempty" yaml:"isEnabled,omitempty"`
	IsActive          *bool     `json:"isActive,omitempty" yaml:"isActive,omitempty"`
	AlertScheduleName *string   `json:"alertScheduleName,omitempty" yaml:"alertScheduleName,omitempty"`
	StartTime         *string   `json:"startTime,omitempty" yaml:"startTime,omitempty"`
	EndTime           *string   `json:"endTime,omitempty" yaml:"endTime,omitempty"`
	DaysOfWeek        *[]string `json:"daysOfWeek,omitempty" yaml:"daysOfWeek,omitempty"`
	Override          *Override `json:"override,omitempty" yaml:"override,omitempty"`
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
	a.IsDefaultSchedule = parser.Bool("isDefaultSchedule")
	a.IsEnabled = parser.Bool("isEnabled")
	a.IsActive = parser.Bool("isActive")
	a.AlertScheduleName = parser.String("alertScheduleName")
	a.StartTime = ParseAlertScheduleSettingsTime(parser, "startTime")
	a.EndTime = ParseAlertScheduleSettingsTime(parser, "endTime")
	a.DaysOfWeek = parser.StringArray("daysOfWeek")
	a.Override = ParseOverride(parser.WithReferenceObjectParser("override"))
}

func (a *AlertScheduleSettings) Validate(validator structure.Validator) {
	// HACK: Dexcom - force default schedule to use expected startTime and endTime
	if a.IsDefaultSchedule != nil && *a.IsDefaultSchedule {
		if a.StartTime == nil || *a.StartTime != AlertScheduleSettingsStartTimeDefault {
			a.StartTime = pointer.FromString(AlertScheduleSettingsStartTimeDefault)
		}
		if a.EndTime == nil || *a.EndTime != AlertScheduleSettingsEndTimeDefault {
			a.EndTime = pointer.FromString(AlertScheduleSettingsEndTimeDefault)
		}
	}

	// HACK: Dexcom - remove empty strings from daysOfWeek
	if a.DaysOfWeek != nil {
		daysOfWeek := []string{}
		for _, dayOfWeek := range *a.DaysOfWeek {
			if dayOfWeek != "" {
				daysOfWeek = append(daysOfWeek, dayOfWeek)
			}
		}
		if len(daysOfWeek) != len(*a.DaysOfWeek) {
			validator.Logger().Warn("Empty string in daysOfWeek of alertScheduleSettings")
		}
		a.DaysOfWeek = pointer.FromStringArray(daysOfWeek)
	}

	validator.Bool("isDefaultSchedule", a.IsDefaultSchedule).Exists()
	if a.IsDefaultSchedule != nil && *a.IsDefaultSchedule {
		validator.Bool("isEnabled", a.IsEnabled).Exists()
		validator.Bool("isActive", a.IsActive) // Dexcom - May not exist
		validator.String("alertScheduleName", a.AlertScheduleName).Exists().Empty()
		validator.String("startTime", a.StartTime).Exists().EqualTo(AlertScheduleSettingsStartTimeDefault)
		validator.String("endTime", a.EndTime).Exists().EqualTo(AlertScheduleSettingsEndTimeDefault)
		validator.StringArray("daysOfWeek", a.DaysOfWeek).Exists().EachOneOf(AlertScheduleSettingsDays()...).EachUnique().LengthEqualTo(len(AlertScheduleSettingsDays()))
	} else {
		validator.Bool("isEnabled", a.IsEnabled).Exists()
		validator.Bool("isActive", a.IsActive) // Dexcom - May not exist
		validator.String("alertScheduleName", a.AlertScheduleName).Exists().NotEmpty()
		validator.String("startTime", a.StartTime).Exists().Using(AlertScheduleSettingsStartTimeValidator)
		validator.String("endTime", a.EndTime).Exists().Using(AlertScheduleSettingsEndTimeValidator)
		validator.StringArray("daysOfWeek", a.DaysOfWeek).Exists().EachOneOf(AlertScheduleSettingsDays()...).EachUnique()
	}

	if a.Override != nil {
		a.Override.Validate(validator.WithReference("override"))
	}
}

func (a *AlertScheduleSettings) Normalize(normalizer structure.Normalizer) {
	if a.DaysOfWeek != nil {
		sort.Sort(dataTypesCommon.DaysOfWeekByDayIndex(*a.DaysOfWeek))
	}
}

func (a *AlertScheduleSettings) IsDefault() bool {
	return a.IsDefaultSchedule != nil && *a.IsDefaultSchedule
}

type Override struct {
	IsOverrideEnabled *bool   `json:"isOverrideEnabled,omitempty" yaml:"isOverrideEnabled,omitempty"`
	Mode              *string `json:"mode,omitempty" yaml:"mode,omitempty"`
	EndTime           *Time   `json:"endTime,omitempty" yaml:"endTime,omitempty"`
}

func ParseOverride(parser structure.ObjectParser) *Override {
	if !parser.Exists() {
		return nil
	}
	datum := &Override{}
	parser.Parse(datum)
	return datum
}

func NewOverride() *Override {
	return &Override{}
}

func (o *Override) Parse(parser structure.ObjectParser) {
	o.IsOverrideEnabled = parser.Bool("isOverrideEnabled")
	o.Mode = parser.String("mode")
	o.EndTime = ParseTime(parser, "endTime")
}

func (o *Override) Validate(validator structure.Validator) {
	validator.Bool("isOverrideEnabled", o.IsOverrideEnabled).Exists()
	validator.String("mode", o.Mode).Exists().OneOf(AlertScheduleSettingsOverrideModes()...)
	validator.Time("endTime", o.EndTime.Raw()).Exists().NotZero()
}

type DaysOfWeekByDay []string

func (d DaysOfWeekByDay) Len() int {
	return len(d)
}
func (d DaysOfWeekByDay) Swap(i int, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d DaysOfWeekByDay) Less(i int, j int) bool {
	left := d[i]
	right := d[j]
	if left == right {
		return false
	}

	daysOfWeek := AlertScheduleSettingsDays()
	leftIndex := slices.Index(daysOfWeek, left)
	rightIndex := slices.Index(daysOfWeek, right)
	if leftIndex != -1 && rightIndex != -1 {
		return leftIndex < rightIndex
	} else if leftIndex != -1 {
		return true
	} else if rightIndex != -1 {
		return false
	}

	return left < right
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
	sort.Stable(AlertSettingsByName(*a))
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

type AlertSettingsByName AlertSettings

func (a AlertSettingsByName) Len() int {
	return len(a)
}
func (a AlertSettingsByName) Swap(i int, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a AlertSettingsByName) Less(i int, j int) bool {
	left := a[i]
	right := a[j]
	if left == nil && right == nil {
		return false
	} else if left == nil {
		return true
	} else if right == nil {
		return false
	}

	leftName := left.Name()
	rightName := right.Name()
	if leftName == nil && rightName == nil {
		return false
	} else if leftName == nil {
		return true
	} else if rightName == nil {
		return false
	}

	return *leftName < *rightName
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
	validator.Time("systemTime", a.SystemTime.Raw()).Exists().NotZero()
	validator.Time("displayTime", a.DisplayTime.Raw()).Exists().NotZero()
	validator.String("alertName", a.AlertName).Exists().OneOf(AlertSettingAlertNames()...)
	if a.AlertName != nil {
		switch *a.AlertName {
		case AlertSettingAlertNameUnknown:
			a.validateUnknown(validator)
		case AlertSettingAlertNameHigh:
			a.validateHigh(validator)
		case AlertSettingAlertNameLow:
			a.validateLow(validator)
		case AlertSettingAlertNameRise:
			a.validateRise(validator)
		case AlertSettingAlertNameFall:
			a.validateFall(validator)
		case AlertSettingAlertNameOutOfRange:
			a.validateOutOfRange(validator)
		case AlertSettingAlertNameUrgentLow:
			a.validateUrgentLow(validator)
		case AlertSettingAlertNameUrgentLowSoon:
			a.validateUrgentLowSoon(validator)
		case AlertSettingAlertNameNoReadings:
			a.validateNoReadings(validator)
		case AlertSettingAlertNameFixedLow:
			a.validateFixedLow(validator)
		}
	}
	validator.String("soundTheme", a.SoundTheme).OneOf(AlertSettingSoundThemes()...)                // Dexcom - May not exist
	validator.String("soundOutputMode", a.SoundOutputMode).OneOf(AlertSettingSoundOutputModes()...) // Dexcom - May not exist
}

func (a *AlertSetting) Name() *string {
	return a.AlertName
}

func (a *AlertSetting) IsNewerMatchThan(alertSetting *AlertSetting) bool {
	return a.AlertName != nil && alertSetting.AlertName != nil && *a.AlertName == *alertSetting.AlertName &&
		a.SystemTime != nil && alertSetting.SystemTime != nil && a.SystemTime.After(*alertSetting.SystemTime.Raw())
}

func (a *AlertSetting) validateUnknown(validator structure.Validator) {
	validator.String("unit", a.Unit).Exists().EqualTo(AlertSettingUnitUnknown)
	validator.Float64("value", a.Value).NotExists()
	validator.Int("snooze", a.Snooze).NotExists()
	validator.Bool("enabled", a.Enabled).Exists()
	validator.Int("delay", a.Delay).NotExists()
	validator.Int("secondaryTriggerCondition", a.SecondaryTriggerCondition).NotExists()
}

func (a *AlertSetting) validateHigh(validator structure.Validator) {
	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingUnitHighs()...)
	valueValidator := validator.Float64("value", a.Value)
	valueValidator.Exists()
	if a.Unit != nil {
		switch *a.Unit {
		case AlertSettingUnitMgdL:
			valueValidator.InRange(AlertSettingValueHighMgdLMinimum, AlertSettingValueHighMgdLMaximum)
		case AlertSettingUnitMmolL:
			valueValidator.InRange(AlertSettingValueHighMmolLMinimum, AlertSettingValueHighMmolLMaximum)
		}
	}
	validator.Int("snooze", a.Snooze).Exists().InRange(AlertSettingSnoozeMinutesMinimum, AlertSettingSnoozeMinutesMaximum)
	validator.Bool("enabled", a.Enabled).Exists()
	validator.Int("delay", a.Delay).GreaterThanOrEqualTo(AlertSettingDelayMinimum) // Dexcom - May not exist
	validator.Int("secondaryTriggerCondition", a.SecondaryTriggerCondition)        // Dexcom - May not exist
}

func (a *AlertSetting) validateLow(validator structure.Validator) {
	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingUnitLows()...)
	valueValidator := validator.Float64("value", a.Value)
	valueValidator.Exists()
	if a.Unit != nil {
		switch *a.Unit {
		case AlertSettingUnitMgdL:
			valueValidator.InRange(AlertSettingValueLowMgdLMinimum, AlertSettingValueLowMgdLMaximum)
		case AlertSettingUnitMmolL:
			valueValidator.InRange(AlertSettingValueLowMmolLMinimum, AlertSettingValueLowMmolLMaximum)
		}
	}
	validator.Int("snooze", a.Snooze).Exists().InRange(AlertSettingSnoozeMinutesMinimum, AlertSettingSnoozeMinutesMaximum)
	validator.Bool("enabled", a.Enabled).Exists()
	validator.Int("delay", a.Delay).NotExists()
	validator.Int("secondaryTriggerCondition", a.SecondaryTriggerCondition) // Dexcom - May not exist
}

func (a *AlertSetting) validateRise(validator structure.Validator) {
	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingUnitRises()...)
	valueValidator := validator.Float64("value", a.Value)
	valueValidator.Exists()
	if a.Unit != nil {
		switch *a.Unit {
		case AlertSettingUnitMgdLMinute:
			valueValidator.InRange(AlertSettingValueRiseMgdLMinuteMinimum, AlertSettingValueRiseMgdLMinuteMaximum)
		case AlertSettingUnitMmolLMinute:
			valueValidator.InRange(AlertSettingValueRiseMmolLMinuteMinimum, AlertSettingValueRiseMmolLMinuteMaximum)
		}
	}
	validator.Int("snooze", a.Snooze).NotExists()
	validator.Bool("enabled", a.Enabled).Exists()
	validator.Int("delay", a.Delay).NotExists()
	validator.Int("secondaryTriggerCondition", a.SecondaryTriggerCondition) // Dexcom - May not exist
}

func (a *AlertSetting) validateFall(validator structure.Validator) {
	// HACK: Dexcom - negative value is invalid; use positive value instead (per Dexcom)
	if a.Value != nil && *a.Value < 0 {
		a.Value = pointer.FromFloat64(-*a.Value)
		validator.Logger().Warn("Negative value for value of fall alert setting")
	}

	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingUnitFalls()...)
	valueValidator := validator.Float64("value", a.Value)
	valueValidator.Exists()
	if a.Unit != nil {
		switch *a.Unit {
		case AlertSettingUnitMgdLMinute:
			valueValidator.InRange(AlertSettingValueFallMgdLMinuteMinimum, AlertSettingValueFallMgdLMinuteMaximum)
		case AlertSettingUnitMmolLMinute:
			valueValidator.InRange(AlertSettingValueFallMmolLMinuteMinimum, AlertSettingValueFallMmolLMinuteMaximum)
		}
	}
	validator.Int("snooze", a.Snooze).InRange(AlertSettingSnoozeMinutesMinimum, AlertSettingSnoozeMinutesMaximum) // Dexcom - May not exist
	validator.Bool("enabled", a.Enabled).Exists()
	validator.Int("delay", a.Delay).NotExists()
	validator.Int("secondaryTriggerCondition", a.SecondaryTriggerCondition) // Dexcom - May not exist
}

func (a *AlertSetting) validateOutOfRange(validator structure.Validator) {
	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingUnitOutOfRanges()...)
	valueValidator := validator.Float64("value", a.Value)
	valueValidator.Exists()
	if a.Unit != nil {
		switch *a.Unit {
		case AlertSettingUnitMinutes:
			valueValidator.InRange(AlertSettingValueOutOfRangeMinutesMinimum, AlertSettingValueOutOfRangeMinutesMaximum)
		}
	}
	validator.Int("snooze", a.Snooze).NotExists()
	validator.Bool("enabled", a.Enabled).Exists()
	validator.Int("delay", a.Delay).NotExists()
	validator.Int("secondaryTriggerCondition", a.SecondaryTriggerCondition) // Dexcom - May not exist
}

func (a *AlertSetting) validateUrgentLow(validator structure.Validator) {
	// HACK: Dexcom - snooze of 28 is invalid; use snooze of 30 instead (per Dexcom); exists in v2 (20180914)
	if a.Snooze != nil && *a.Snooze == 28 {
		// Do not warn as there are too many of these
		// validator.Logger().Warn("Invalid value for snooze of urgent low alert setting")
		a.Snooze = pointer.FromInt(30)
	}

	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingUnitUrgentLows()...)
	valueValidator := validator.Float64("value", a.Value)
	valueValidator.Exists()
	if a.Unit != nil {
		switch *a.Unit {
		case AlertSettingUnitMgdL:
			valueValidator.InRange(AlertSettingValueUrgentLowMgdLMinimum, AlertSettingValueUrgentLowMgdLMaximum)
		case AlertSettingUnitMmolL:
			valueValidator.InRange(AlertSettingValueUrgentLowMmolLMinimum, AlertSettingValueUrgentLowMmolLMaximum)
		}
	}
	validator.Int("snooze", a.Snooze).Exists().InRange(AlertSettingSnoozeMinutesMinimum, AlertSettingSnoozeMinutesMaximum)
	validator.Bool("enabled", a.Enabled).Exists().True()
	validator.Int("delay", a.Delay).NotExists()
	validator.Int("secondaryTriggerCondition", a.SecondaryTriggerCondition) // Dexcom - May not exist
}

func (a *AlertSetting) validateUrgentLowSoon(validator structure.Validator) {
	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingUnitUrgentLowSoons()...)
	valueValidator := validator.Float64("value", a.Value)
	valueValidator.Exists()
	if a.Unit != nil {
		switch *a.Unit {
		case AlertSettingUnitMgdL:
			valueValidator.InRange(AlertSettingValueUrgentLowSoonMgdLMinimum, AlertSettingValueUrgentLowSoonMgdLMaximum)
		case AlertSettingUnitMmolL:
			valueValidator.InRange(AlertSettingValueUrgentLowSoonMmolLMinimum, AlertSettingValueUrgentLowSoonMmolLMaximum)
		}
	}
	validator.Int("snooze", a.Snooze).Exists().InRange(AlertSettingSnoozeMinutesMinimum, AlertSettingSnoozeMinutesMaximum)
	validator.Bool("enabled", a.Enabled).Exists()
	validator.Int("delay", a.Delay).NotExists()
	validator.Int("secondaryTriggerCondition", a.SecondaryTriggerCondition) // Dexcom - May not exist
}

func (a *AlertSetting) validateNoReadings(validator structure.Validator) {
	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingUnitNoReadings()...)
	valueValidator := validator.Float64("value", a.Value)
	valueValidator.Exists()
	if a.Unit != nil {
		switch *a.Unit {
		case AlertSettingUnitMinutes:
			valueValidator.InRange(AlertSettingValueNoReadingsMinutesMinimum, AlertSettingValueNoReadingsMinutesMaximum)
		}
	}
	validator.Int("snooze", a.Snooze).InRange(AlertSettingSnoozeMinutesMinimum, AlertSettingSnoozeMinutesMaximum) // Dexcom - May not exist
	validator.Bool("enabled", a.Enabled).Exists()
	validator.Int("delay", a.Delay).NotExists()
	validator.Int("secondaryTriggerCondition", a.SecondaryTriggerCondition) // Dexcom - May not exist
}

func (a *AlertSetting) validateFixedLow(validator structure.Validator) {
	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingUnitFixedLows()...)
	valueValidator := validator.Float64("value", a.Value)
	valueValidator.Exists()
	if a.Unit != nil {
		switch *a.Unit {
		case AlertSettingUnitMinutes:
			valueValidator.InRange(AlertSettingValueFixedLowMinutesMinimum, AlertSettingValueFixedLowMinutesMaximum)
		}
	}
	validator.Int("snooze", a.Snooze).NotExists()
	validator.Bool("enabled", a.Enabled).Exists().True()
	validator.Int("delay", a.Delay).NotExists()
	validator.Int("secondaryTriggerCondition", a.SecondaryTriggerCondition).NotExists()
}

func ParseAlertScheduleSettingsTime(parser structure.ObjectParser, reference string) *string {
	serialized := parser.String(reference)
	if serialized == nil {
		return nil
	}

	hours, minutes, err := ParseAlertScheduleSettingsTimeHoursAndMinutes(*serialized)
	if err != nil {
		parser.WithReferenceErrorReporter(reference).ReportError(structureParser.ErrorValueTimeNotParsable(*serialized, "15:04"))
		return nil
	}

	return pointer.FromString(fmt.Sprintf("%02d:%02d", hours, minutes))
}

// Parse HH:MM where HH and MM can be any two digits, potentially parsing AM/PM suffix to convert
// standard time.
func ParseAlertScheduleSettingsTimeHoursAndMinutes(value string) (int, int, error) {
	var hour *int
	var minute *int

	// Parse a copy to retain original, upper case with spaces removed
	parsable := strings.ReplaceAll(strings.ToUpper(value), " ", "")

	// Determine if there is a meridiem, if so only use the non-meridiem portion
	meridiemMatches := meridiemRegexp.FindStringSubmatch(parsable)
	if meridiemMatches != nil {
		parsable = meridiemMatches[1]
	}

	// Parse out all of the fields, be *very* lenient
	pipeline := parsePipeline{
		digitsParser(1, 2, &hour),
		characterParser(":"),
		digitsParser(1, 2, &minute),
	}
	remaining, err := pipeline.Parse(parsable)

	// If we had an error or there are unparsed characters left, then bail
	if err != nil || remaining != "" {
		return 0, 0, errors.New("alert schedule settings time is not parsable")
	}

	// If meridiem exists, then translate to 24-hour time
	if meridiemMatches != nil {
		if hour == nil {
			return 0, 0, errors.New("alert schedule settings time is not parsable")
		}
		if *hour == 12 {
			*hour = *hour - 12
		}
		if meridiem := meridiemMatches[2]; meridiem == "P" || meridiem == "PM" {
			*hour = *hour + 12
		}
	}

	return intOrDefault(hour, 0), intOrDefault(minute, 0), nil
}

func IsValidAlertScheduleSettingsStartTime(value string) bool {
	return ValidateAlertScheduleSettingsStartTime(value) == nil
}

func AlertScheduleSettingsStartTimeValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateAlertScheduleSettingsStartTime(value))
}

func ValidateAlertScheduleSettingsStartTime(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if hour, minute, err := ParseAlertScheduleSettingsTimeHoursAndMinutes(value); err != nil || hour > 23 || minute > 59 {
		return ErrorValueStringAsAlertScheduleSettingsTimeNotValid(value)
	}
	return nil
}

func IsValidAlertScheduleSettingsEndTime(value string) bool {
	return ValidateAlertScheduleSettingsEndTime(value) == nil
}

func AlertScheduleSettingsEndTimeValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateAlertScheduleSettingsEndTime(value))
}

func ValidateAlertScheduleSettingsEndTime(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if hour, minute, err := ParseAlertScheduleSettingsTimeHoursAndMinutes(value); err != nil || hour > 47 || minute > 59 {
		return ErrorValueStringAsAlertScheduleSettingsTimeNotValid(value)
	}
	return nil
}

func ErrorValueStringAsAlertScheduleSettingsTimeNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as alert schedule settings time", value)
}

var meridiemRegexp = regexp.MustCompile(`^(.*)(A|P|AM|PM)$`) // Assumes upper case
