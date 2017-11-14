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
		validator.Validating(strconv.Itoa(index), device).Exists().Validate()
	}
}

type Device struct {
	Model             string          `json:"model,omitempty"`
	LastUploadDate    time.Time       `json:"lastUploadDate,omitempty"`
	AlertSettings     []*AlertSetting `json:"alertSettings,omitempty"`
	UDI               *string         `json:"udi,omitempty"`
	SerialNumber      *string         `json:"serialNumber,omitempty"`
	TransmitterID     *string         `json:"transmitterId,omitempty"`
	SoftwareVersion   *string         `json:"softwareVersion,omitempty"`
	SoftwareNumber    *string         `json:"softwareNumber,omitempty"`
	Language          *string         `json:"language,omitempty"`
	IsMmolDisplayMode *bool           `json:"isMmolDisplayMode,omitempty"`
	IsBlindedMode     *bool           `json:"isBlindedMode,omitempty"`
	Is24HourMode      *bool           `json:"is24HourMode,omitempty"`
	DisplayTimeOffset *int            `json:"displayTimeOffset,omitempty"`
	SystemTimeOffset  *int            `json:"systemTimeOffset,omitempty"`
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
	validator.String("model", &d.Model).OneOf(ModelG5MobileApp, ModelG5Receiver, ModelG4WithShareReceiver, ModelG4Receiver)
	validator.Time("lastUploadDate", &d.LastUploadDate).NotZero()
	existingAlertNames := &[]string{}
	validator = validator.WithReference("alertSettings")
	for index, alertSetting := range d.AlertSettings {
		validator.Validating(strconv.Itoa(index), structureValidator.NewValidatableWithStringArrayAdapter(alertSetting, existingAlertNames)).Exists().Validate() // TODO: Exists broken!!!
	}
	validator.String("udi", d.UDI).NotEmpty()
	validator.String("serialNumber", d.SerialNumber).NotEmpty()
	validator.String("transmitterId", d.TransmitterID).Matches(TransmitterIDExpression)
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

func (a *AlertSetting) Validate(validator structure.Validator, existingAlertNames *[]string) {
	validator.Time("systemTime", &a.SystemTime).BeforeNow(NowThreshold)
	validator.Time("displayTime", &a.DisplayTime).NotZero()
	validator.String("alertName", &a.AlertName).OneOf(AlertFixedLow, AlertLow, AlertHigh, AlertRise, AlertFall, AlertOutOfRange).NotOneOf(*existingAlertNames...)

	switch a.AlertName {
	case AlertFixedLow:
		a.validateFixedLow(validator)
	case AlertLow:
		a.validateLow(validator)
	case AlertHigh:
		a.validateHigh(validator)
	case AlertRise:
		a.validateRise(validator)
	case AlertFall:
		a.validateFall(validator)
	case AlertOutOfRange:
		a.validateOutOfRange(validator)
	}

	*existingAlertNames = append(*existingAlertNames, a.AlertName)
}

func (a *AlertSetting) validateFixedLow(validator structure.Validator) {
	validator.String("unit", &a.Unit).OneOf(UnitMgdL) // TODO: Add UnitMmolL
	switch a.Unit {
	case UnitMgdL:
		validator.Float64("value", &a.Value).EqualTo(55)
	case UnitMmolL:
		// TODO: Add value validation
	}
	validator.Int("delay", &a.Delay).EqualTo(0)
	validator.Int("snooze", &a.Snooze).OneOf(fixedLowSnoozes...)
	validator.Bool("enabled", &a.Enabled).True()
}

func (a *AlertSetting) validateLow(validator structure.Validator) {
	validator.String("unit", &a.Unit).OneOf(UnitMgdL) // TODO: Add UnitMmolL
	switch a.Unit {
	case UnitMgdL:
		validator.Float64("value", &a.Value).OneOf(lowValues...)
	case UnitMmolL:
		// TODO: Add value validation
	}
	validator.Int("delay", &a.Delay).EqualTo(0)
	validator.Int("snooze", &a.Snooze).OneOf(lowSnoozes...)
}

func (a *AlertSetting) validateHigh(validator structure.Validator) {
	validator.String("unit", &a.Unit).OneOf(UnitMgdL) // TODO: Add UnitMmolL
	switch a.Unit {
	case UnitMgdL:
		validator.Float64("value", &a.Value).OneOf(highValues...)
	case UnitMmolL:
		// TODO: Add value validation
	}
	validator.Int("delay", &a.Delay).EqualTo(0)
	validator.Int("snooze", &a.Snooze).OneOf(highSnoozes...)
}

func (a *AlertSetting) validateRise(validator structure.Validator) {
	// HACK: Dexcom - use "mg/dL/min" rather than incorrect "minutes"
	if a.Unit == UnitMinutes {
		a.Unit = UnitMgdLMin
	}

	validator.String("unit", &a.Unit).OneOf(UnitMgdLMin) // TODO: Add UnitMmolLMin
	switch a.Unit {
	case UnitMgdLMin:
		validator.Float64("value", &a.Value).OneOf(2, 3)
	case UnitMmolLMin:
		// TODO: Add value validation
	}
	validator.Int("delay", &a.Delay).EqualTo(0)
	validator.Int("snooze", &a.Snooze).OneOf(riseSnoozes...)
}

func (a *AlertSetting) validateFall(validator structure.Validator) {
	// HACK: Dexcom - use "mg/dL/min" rather than incorrect "minutes"
	if a.Unit == UnitMinutes {
		a.Unit = UnitMgdLMin
	}
	// HACK: Dexcom - use positive value rather than incorrect negative value
	if a.Value < 0 {
		a.Value = -a.Value
	}

	validator.String("unit", &a.Unit).OneOf(UnitMgdLMin) // TODO: UnitMmolLMin
	switch a.Unit {
	case UnitMgdLMin:
		validator.Float64("value", &a.Value).OneOf(2, 3)
	case UnitMmolLMin:
		// TODO: Add value validation
	}
	validator.Int("delay", &a.Delay).EqualTo(0)
	validator.Int("snooze", &a.Snooze).OneOf(fallSnoozes...)
}

func (a *AlertSetting) validateOutOfRange(validator structure.Validator) {

	validator.String("unit", &a.Unit).EqualTo(UnitMinutes)
	validator.Float64("value", &a.Value).OneOf(outOfRangeValues...)
	validator.Int("delay", &a.Delay).EqualTo(0)
	validator.Int("snooze", &a.Snooze).OneOf(outOfRangeSnoozes...)
}

var fixedLowSnoozes = []int{0, 30}
var lowValues = generateFloatRange(60, 100, 5)
var lowSnoozes = append(append([]int{0}, generateIntegerRange(15, 240, 5)...), generateIntegerRange(255, 300, 15)...)
var highValues = generateFloatRange(120, 400, 10)
var highSnoozes = append(append([]int{0}, generateIntegerRange(15, 240, 5)...), generateIntegerRange(255, 300, 15)...)
var riseSnoozes = []int{0, 30}
var fallSnoozes = []int{0, 30}
var outOfRangeValues = generateFloatRange(20, 240, 5)
var outOfRangeSnoozes = []int{0, 20, 30}

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
