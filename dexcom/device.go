package dexcom

import (
	"strconv"
	"time"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type DevicesResponse struct {
	Devices []*Device `json:"devices,omitempty"`
}

func NewDevicesResponse() *DevicesResponse {
	return &DevicesResponse{}
}

func (d *DevicesResponse) Parse(parser structure.ObjectParser) {
	if devicesParser := parser.WithReferenceArrayParser("devices"); devicesParser.Exists() {
		for _, reference := range devicesParser.References() {
			if deviceParser := devicesParser.WithReferenceObjectParser(reference); deviceParser.Exists() {
				device := NewDevice()
				device.Parse(deviceParser)
				deviceParser.NotParsed()
				d.Devices = append(d.Devices, device)
			}
		}
		devicesParser.NotParsed()
	}
}

func (d *DevicesResponse) Validate(validator structure.Validator) {
	validator = validator.WithReference("devices")
	for index, device := range d.Devices {
		if deviceValidator := validator.WithReference(strconv.Itoa(index)); device != nil {
			device.Validate(deviceValidator)
		} else {
			deviceValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

type Device struct {
	Model             string        `json:"model,omitempty"`
	LastUploadDate    time.Time     `json:"lastUploadDate,omitempty"`
	AlertSettings     AlertSettings `json:"alertSettings,omitempty"`
	UDI               *string       `json:"udi,omitempty"`
	SerialNumber      *string       `json:"serialNumber,omitempty"`
	TransmitterID     *string       `json:"transmitterId,omitempty"`
	SoftwareVersion   *string       `json:"softwareVersion,omitempty"`
	SoftwareNumber    *string       `json:"softwareNumber,omitempty"`
	Language          *string       `json:"language,omitempty"`
	IsMmolDisplayMode *bool         `json:"isMmolDisplayMode,omitempty"`
	IsBlindedMode     *bool         `json:"isBlindedMode,omitempty"`
	Is24HourMode      *bool         `json:"is24HourMode,omitempty"`
	DisplayTimeOffset *int          `json:"displayTimeOffset,omitempty"`
	SystemTimeOffset  *int          `json:"systemTimeOffset,omitempty"`
}

func NewDevice() *Device {
	return &Device{}
}

func (d *Device) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("model"); ptr != nil {
		d.Model = *ptr
	}
	if ptr := parser.Time("lastUploadDate", DateTimeFormat); ptr != nil {
		d.LastUploadDate = *ptr
	}
	if alertSettingsParser := parser.WithReferenceArrayParser("alertSettings"); alertSettingsParser.Exists() {
		for _, reference := range alertSettingsParser.References() {
			if alertSettingParser := alertSettingsParser.WithReferenceObjectParser(reference); alertSettingParser.Exists() {
				alertSetting := NewAlertSetting()
				alertSetting.Parse(alertSettingParser)
				alertSettingParser.NotParsed()
				d.AlertSettings = append(d.AlertSettings, alertSetting)
			}
		}
		alertSettingsParser.NotParsed()
		d.AlertSettings = d.AlertSettings.Deduplicate() // HACK: Dexcom - duplicates not allowed; delete older
	}
	d.UDI = parser.String("udi")
	d.SerialNumber = parser.String("serialNumber")
	d.TransmitterID = parser.String("transmitterId")
	d.SoftwareVersion = parser.String("softwareVersion")
	d.SoftwareNumber = parser.String("softwareNumber")
	d.Language = parser.String("language")
	d.IsMmolDisplayMode = parser.Bool("isMmolDisplayMode")
	d.IsBlindedMode = parser.Bool("isBlindedMode")
	d.Is24HourMode = parser.Bool("is24HourMode")
	d.DisplayTimeOffset = parser.Int("displayTimeOffset")
	d.SystemTimeOffset = parser.Int("systemTimeOffset")
}

func (d *Device) Validate(validator structure.Validator) {
	validator.String("model", &d.Model).OneOf(ModelG5MobileApp, ModelG5Receiver, ModelG4WithShareReceiver, ModelG4Receiver, ModelUnknown)
	validator.Time("lastUploadDate", &d.LastUploadDate).NotZero()
	validator = validator.WithReference("alertSettings")
	for index, alertSetting := range d.AlertSettings {
		if alertSettingValidator := validator.WithReference(strconv.Itoa(index)); alertSetting != nil {
			alertSetting.Validate(alertSettingValidator)
		} else {
			alertSettingValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
	validator.String("udi", d.UDI).NotEmpty()
	validator.String("serialNumber", d.SerialNumber).NotEmpty()
	validator.String("transmitterId", d.TransmitterID).Matches(transmitterIDExpression)
	validator.String("softwareVersion", d.SoftwareVersion).NotEmpty()
	validator.String("softwareNumber", d.SoftwareNumber).NotEmpty()
	validator.String("language", d.Language).NotEmpty()
}

type AlertSetting struct {
	SystemTime  time.Time `json:"systemTime,omitempty"`
	DisplayTime time.Time `json:"displayTime,omitempty"`
	AlertName   string    `json:"alertName,omitempty"`
	Unit        string    `json:"unit,omitempty"`
	Value       float64   `json:"value,omitempty"`
	Delay       int       `json:"delay,omitempty"`
	Snooze      int       `json:"snooze,omitempty"`
	Enabled     bool      `json:"enabled,omitempty"`
}

func NewAlertSetting() *AlertSetting {
	return &AlertSetting{}
}

func (a *AlertSetting) Parse(parser structure.ObjectParser) {
	if ptr := parser.Time("systemTime", DateTimeFormat); ptr != nil {
		a.SystemTime = *ptr
	}
	if ptr := parser.Time("displayTime", DateTimeFormat); ptr != nil {
		a.DisplayTime = *ptr
	}
	if ptr := parser.String("alertName"); ptr != nil {
		a.AlertName = *ptr
	}
	if ptr := parser.String("unit"); ptr != nil {
		a.Unit = *ptr
	}
	if ptr := parser.Float64("value"); ptr != nil {
		a.Value = *ptr
	}
	if ptr := parser.Int("delay"); ptr != nil {
		a.Delay = *ptr
	}
	if ptr := parser.Int("snooze"); ptr != nil {
		a.Snooze = *ptr
	}
	if ptr := parser.Bool("enabled"); ptr != nil {
		a.Enabled = *ptr
	}
}

func (a *AlertSetting) Validate(validator structure.Validator) {
	validator.Time("systemTime", &a.SystemTime).NotZero().BeforeNow(NowThreshold)
	validator.Time("displayTime", &a.DisplayTime).NotZero()
	validator.String("alertName", &a.AlertName).OneOf(AlertNames()...)

	switch a.AlertName {
	case AlertNameFixedLow:
		a.validateFixedLow(validator)
	case AlertNameLow:
		a.validateLow(validator)
	case AlertNameHigh:
		a.validateHigh(validator)
	case AlertNameRise:
		a.validateRise(validator)
	case AlertNameFall:
		a.validateFall(validator)
	case AlertNameOutOfRange:
		a.validateOutOfRange(validator)
	}
}

func (a *AlertSetting) IsNewerMatchThan(alertSetting *AlertSetting) bool {
	return a.AlertName == alertSetting.AlertName && a.SystemTime.After(alertSetting.SystemTime)
}

func (a *AlertSetting) validateFixedLow(validator structure.Validator) {
	// HACK: Dexcom - snooze of 28 is invalid; use snooze of 30 instead (per Dexcom)
	if a.Snooze == 28 {
		a.Snooze = 30
	}

	validator.String("unit", &a.Unit).OneOf(AlertSettingFixedLowUnits()...)
	if values := AlertSettingFixedLowValuesForUnits(a.Unit); values != nil {
		validator.Float64("value", &a.Value).OneOf(values...)
	}
	validator.Int("delay", &a.Delay).OneOf(0)
	validator.Int("snooze", &a.Snooze).OneOf(AlertSettingFixedLowSnoozes()...)
	validator.Bool("enabled", &a.Enabled).True()
}

func (a *AlertSetting) validateLow(validator structure.Validator) {
	validator.String("unit", &a.Unit).OneOf(AlertSettingLowUnits()...)
	if values := AlertSettingLowValuesForUnits(a.Unit); values != nil {
		validator.Float64("value", &a.Value).OneOf(values...)
	}
	validator.Int("delay", &a.Delay).OneOf(0)
	validator.Int("snooze", &a.Snooze).OneOf(AlertSettingLowSnoozes()...)
}

func (a *AlertSetting) validateHigh(validator structure.Validator) {
	validator.String("unit", &a.Unit).OneOf(AlertSettingHighUnits()...)
	if values := AlertSettingHighValuesForUnits(a.Unit); values != nil {
		validator.Float64("value", &a.Value).OneOf(values...)
	}
	validator.Int("delay", &a.Delay).OneOf(0)
	validator.Int("snooze", &a.Snooze).OneOf(AlertSettingHighSnoozes()...)
}

func (a *AlertSetting) validateRise(validator structure.Validator) {
	validator.String("unit", &a.Unit).OneOf(AlertSettingRiseUnits()...)
	if values := AlertSettingRiseValuesForUnits(a.Unit); values != nil {
		validator.Float64("value", &a.Value).OneOf(values...)
	}
	validator.Int("delay", &a.Delay).OneOf(0)
	validator.Int("snooze", &a.Snooze).OneOf(AlertSettingRiseSnoozes()...)
}

func (a *AlertSetting) validateFall(validator structure.Validator) {
	// HACK: Dexcom - negative value is invalid; use positive value instead (per Dexcom)
	if a.Value < 0 {
		a.Value = -a.Value
	}

	validator.String("unit", &a.Unit).OneOf(AlertSettingFallUnits()...)
	if values := AlertSettingFallValuesForUnits(a.Unit); values != nil {
		validator.Float64("value", &a.Value).OneOf(values...)
	}
	validator.Int("delay", &a.Delay).OneOf(0)
	validator.Int("snooze", &a.Snooze).OneOf(AlertSettingFallSnoozes()...)
}

func (a *AlertSetting) validateOutOfRange(validator structure.Validator) {
	validator.String("unit", &a.Unit).OneOf(AlertSettingOutOfRangeUnits()...)
	if values := AlertSettingOutOfRangeValuesForUnits(a.Unit); values != nil {
		validator.Float64("value", &a.Value).OneOf(values...)
	}
	validator.Int("delay", &a.Delay).OneOf(0)
	validator.Int("snooze", &a.Snooze).OneOf(AlertSettingOutOfRangeSnoozes()...)
}

type AlertSettings []*AlertSetting

func (a AlertSettings) ContainsNewerMatch(alertSetting *AlertSetting) bool {
	for _, testAlertSetting := range a {
		if testAlertSetting.IsNewerMatchThan(alertSetting) {
			return true
		}
	}
	return false
}

func (a AlertSettings) Deduplicate() AlertSettings {
	alertSettings := AlertSettings{}
	alertNameMap := map[string]bool{}
	for index, alertSetting := range a {
		if !alertNameMap[alertSetting.AlertName] && !a[index+1:].ContainsNewerMatch(alertSetting) {
			alertSettings = append(alertSettings, alertSetting)
			alertNameMap[alertSetting.AlertName] = true
		}
	}
	return alertSettings
}

func AlertSettingFixedLowUnits() []string {
	return []string{UnitMgdL} // TODO: Add UnitMmolL
}

func AlertSettingFixedLowValuesForUnits(units string) []float64 {
	switch units {
	case UnitMgdL:
		return []float64{55}
	case UnitMmolL:
		return nil // TODO: Add values
	}
	return nil
}

func AlertSettingFixedLowSnoozes() []int {
	return []int{0, 30}
}

func AlertSettingLowUnits() []string {
	return []string{UnitMgdL} // TODO: Add UnitMmolL
}

func AlertSettingLowValuesForUnits(units string) []float64 {
	switch units {
	case UnitMgdL:
		return alertSettingLowValuesMgdL
	case UnitMmolL:
		return alertSettingLowValuesMmolL
	}
	return nil
}

var alertSettingLowValuesMgdL = generateFloatRange(60, 100, 5)
var alertSettingLowValuesMmolL []float64 // TODO: Add values

func AlertSettingLowSnoozes() []int {
	return alertSettingLowSnoozes
}

var alertSettingLowSnoozes = append(append([]int{0}, generateIntegerRange(15, 240, 5)...), generateIntegerRange(255, 300, 15)...)

func AlertSettingHighUnits() []string {
	return []string{UnitMgdL} // TODO: Add UnitMmolL
}

func AlertSettingHighValuesForUnits(units string) []float64 {
	switch units {
	case UnitMgdL:
		return alertSettingHighValuesMgdL
	case UnitMmolL:
		return alertSettingHighValuesMmolL
	}
	return nil
}

var alertSettingHighValuesMgdL = generateFloatRange(120, 400, 10)
var alertSettingHighValuesMmolL []float64 // TODO: Add values

func AlertSettingHighSnoozes() []int {
	return alertSettingsHighSnoozes
}

var alertSettingsHighSnoozes = append(append([]int{0}, generateIntegerRange(15, 240, 5)...), generateIntegerRange(255, 300, 15)...)

func AlertSettingRiseUnits() []string {
	return []string{UnitMgdLMin} // TODO: UnitMmolLMin
}

func AlertSettingRiseValuesForUnits(units string) []float64 {
	switch units {
	case UnitMgdLMin:
		return []float64{2, 3}
	case UnitMmolLMin:
		return nil // TODO: Add values
	}
	return nil
}

func AlertSettingRiseSnoozes() []int {
	return []int{0, 30}
}

func AlertSettingFallUnits() []string {
	return []string{UnitMgdLMin} // TODO: UnitMmolLMin
}

func AlertSettingFallValuesForUnits(units string) []float64 {
	switch units {
	case UnitMgdLMin:
		return []float64{2, 3}
	case UnitMmolLMin:
		return nil // TODO: Add values
	}
	return nil
}

func AlertSettingFallSnoozes() []int {
	return []int{0, 30}
}

func AlertSettingOutOfRangeUnits() []string {
	return []string{UnitMinutes}
}

func AlertSettingOutOfRangeValuesForUnits(units string) []float64 {
	switch units {
	case UnitMinutes:
		return alertSettingOutOfRangeValuesMinutes
	}
	return nil
}

var alertSettingOutOfRangeValuesMinutes = generateFloatRange(20, 240, 5)

func AlertSettingOutOfRangeSnoozes() []int {
	return []int{0, 20, 25, 30}
}

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
